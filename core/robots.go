package core

import (
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

func ParseRobots(site *url.URL, output *Output, c *colly.Collector, wg *sync.WaitGroup) {
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
				url = FixUrl(url, site)
				outputFormat := fmt.Sprintf("[robots] - %s", url)
				fmt.Println(outputFormat)
				if output != nil {
					output.WriteToFile(outputFormat)
				}
				_ = c.Visit(url)
			}
		}
	}

}
