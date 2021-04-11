package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"

	"github.com/gocolly/colly/v2"
)

func ParseRobots(site *url.URL, crawler *Crawler, c *colly.Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	robotsURL := site.String() + "/robots.txt"

	resp, err := http.Get(robotsURL)
	if err != nil {
		return
	}
	if resp.StatusCode == 200 {
		Logger.Infof("Found robots.txt: %s", robotsURL)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		lines := strings.Split(string(body), "\n")

		var re = regexp.MustCompile(".*llow: ")
		for _, line := range lines {
			if strings.Contains(line, "llow: ") {
				url := re.ReplaceAllString(line, "")
				url = FixUrl(site, url)
				if url == "" {
					continue
				}
				outputFormat := fmt.Sprintf("[robots] - %s", url)

				if crawler.JsonOutput {
					sout := SpiderOutput{
						Input:      crawler.Input,
						Source:     "robots",
						OutputType: "url",
						Output:     url,
					}
					if data, err := jsoniter.MarshalToString(sout); err == nil {
						outputFormat = data
					}
				} else if crawler.Quiet {
					outputFormat = url
				}
				fmt.Println(outputFormat)
				if crawler.Output != nil {
					crawler.Output.WriteToFile(outputFormat)
				}
				_ = c.Visit(url)
			}
		}
	}

}
