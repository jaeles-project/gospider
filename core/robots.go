package core

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "regexp"
    "strings"
    "sync"

    "github.com/gocolly/colly/v2"
)

func ParseRobots(site *url.URL, quiet bool, output *Output, c *colly.Collector, wg *sync.WaitGroup) {
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
                if !quiet {
                    fmt.Println(outputFormat)
                }

                if output != nil {
                    output.WriteToFile(outputFormat)
                }
                _ = c.Visit(url)
            }
        }
    }

}
