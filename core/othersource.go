package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

func OtherSources(domain string, includeSubs bool) []string {
	noSubs := true
	if includeSubs {
		noSubs = false
	}
	var urls []string

	fetchFns := []fetchFn{
		getWaybackURLs,
		getCommonCrawlURLs,
		getVirusTotalURLs,
		getOtxUrls,
	}

	var wg sync.WaitGroup

	for _, fn := range fetchFns {
		wUrlChan := make(chan wurl)
		wg.Add(1)
		fetch := fn
		go func() {
			defer wg.Done()
			resp, err := fetch(domain, noSubs)
			if err != nil {
				return
			}
			for _, r := range resp {
				wUrlChan <- r
			}
		}()

		go func() {
			wg.Wait()
			close(wUrlChan)
		}()

		for w := range wUrlChan {
			urls = append(urls, w.url)
		}
	}
	return Unique(urls)
}

type wurl struct {
	date string
	url  string
}

type fetchFn func(string, bool) ([]wurl, error)

func getWaybackURLs(domain string, noSubs bool) ([]wurl, error) {
	subsWildcard := "*."
	if noSubs {
		subsWildcard = ""
	}
	res, err := http.Get(
		fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey", subsWildcard, domain),
	)
	if err != nil {
		return []wurl{}, err
	}

	raw, err := ioutil.ReadAll(res.Body)

	res.Body.Close()
	if err != nil {
		return []wurl{}, err
	}

	var wrapper [][]string
	err = json.Unmarshal(raw, &wrapper)

	out := make([]wurl, 0, len(wrapper))

	skip := true
	for _, urls := range wrapper {
		// The first item is always just the string "original",
		// so we should skip the first item
		if skip {
			skip = false
			continue
		}
		out = append(out, wurl{date: urls[1], url: urls[2]})
	}

	return out, nil

}

func getCommonCrawlURLs(domain string, noSubs bool) ([]wurl, error) {
	subsWildcard := "*."
	if noSubs {
		subsWildcard = ""
	}
	res, err := http.Get(
		fmt.Sprintf("http://index.commoncrawl.org/CC-MAIN-2019-51-index?url=%s%s/*&output=json", subsWildcard, domain),
	)
	if err != nil {
		return []wurl{}, err
	}

	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)

	out := make([]wurl, 0)

	for sc.Scan() {
		wrapper := struct {
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		}{}
		err = json.Unmarshal([]byte(sc.Text()), &wrapper)

		if err != nil {
			continue
		}

		out = append(out, wurl{date: wrapper.Timestamp, url: wrapper.URL})
	}

	return out, nil

}

func getVirusTotalURLs(domain string, noSubs bool) ([]wurl, error) {
	out := make([]wurl, 0)

	apiKey := os.Getenv("VT_API_KEY")
	if apiKey == "" {
		Logger.Warnf("You are not set VirusTotal API Key yet.")
		return out, nil
	}

	fetchURL := fmt.Sprintf(
		"https://www.virustotal.com/vtapi/v2/domain/report?apikey=%s&domain=%s",
		apiKey,
		domain,
	)

	resp, err := http.Get(fetchURL)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	wrapper := struct {
		URLs []struct {
			URL string `json:"url"`
		} `json:"detected_urls"`
	}{}

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&wrapper)

	for _, u := range wrapper.URLs {
		out = append(out, wurl{url: u.URL})
	}

	return out, nil
}

func getOtxUrls(domain string, noSubs bool) ([]wurl, error) {
	var urls []wurl
	page := 0
	for {
		r, err := http.Get(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/hostname/%s/url_list?limit=50&page=%d", domain, page))
		if err != nil {
			return []wurl{}, err
		}
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return []wurl{}, err
		}
		r.Body.Close()

		wrapper := struct {
			HasNext    bool `json:"has_next"`
			ActualSize int  `json:"actual_size"`
			URLList    []struct {
				Domain   string `json:"domain"`
				URL      string `json:"url"`
				Hostname string `json:"hostname"`
				Httpcode int    `json:"httpcode"`
				PageNum  int    `json:"page_num"`
				FullSize int    `json:"full_size"`
				Paged    bool   `json:"paged"`
			} `json:"url_list"`
		}{}
		err = json.Unmarshal(bytes, &wrapper)
		if err != nil {
			return []wurl{}, err
		}
		for _, url := range wrapper.URLList {
			urls = append(urls, wurl{url: url.URL})
		}
		if !wrapper.HasNext {
			break
		}
		page++
	}
	return urls, nil
}
