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

func OtherSources(domain string) []string {
	var urls []string

	fetchFns := []fetchFn{
		getWaybackURLs,
		getCommonCrawlURLs,
		getVirusTotalURLs,
	}

	var wg sync.WaitGroup

	for _, fn := range fetchFns {
		wUrlChan := make(chan wurl)
		wg.Add(1)
		fetch := fn
		go func() {
			defer wg.Done()
			resp, err := fetch(domain)
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
	return unique(urls)
}

type wurl struct {
	date string
	url  string
}

type fetchFn func(string) ([]wurl, error)

func getWaybackURLs(domain string) ([]wurl, error) {
	res, err := http.Get(
		fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&Output=json&collapse=urlkey", domain),
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

func getCommonCrawlURLs(domain string) ([]wurl, error) {
	res, err := http.Get(
		fmt.Sprintf("http://index.commoncrawl.org/CC-MAIN-2019-51-index?url=*.%s/*&Output=json", domain),
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

func getVirusTotalURLs(domain string) ([]wurl, error) {
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
