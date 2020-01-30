package core

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/spf13/cobra"
	"github.com/theblackturtle/gospider/stringset"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Crawler struct {
	cmd      *cobra.Command
	C        *colly.Collector
	Output   *Output
	domainRe *regexp.Regexp

	site   string
	domain string
}

func NewCrawler(site string, cmd *cobra.Command) *Crawler {
	domain := GetDomain(site)
	if domain == "" {
		Logger.Error("Failed to parse domain")
		os.Exit(1)
	}
	Logger.Infof("Crawling site: %s", site)

	maxDepth, _ := cmd.Flags().GetInt("depth")
	concurrent, _ := cmd.Flags().GetInt("concurrent")
	delay, _ := cmd.Flags().GetInt("delay")
	randomDelay, _ := cmd.Flags().GetInt("random-delay")

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(maxDepth),
		colly.IgnoreRobotsTxt(),
	)

	// Setup http client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Set proxy
	proxy, _ := cmd.Flags().GetString("proxy")
	if proxy != "" {
		Logger.Debugf("Proxy: %s", proxy)
		pU, err := url.Parse(proxy)
		if err != nil {
			Logger.Error("Failed to set proxy")
		} else {
			tr.Proxy = http.ProxyURL(pU)
		}
	}
	c.WithTransport(tr)

	// Get headers here to overwrite if "burp" flag used
	burpFile, _ := cmd.Flags().GetString("burp")
	if burpFile != "" {
		bF, err := os.Open(burpFile)
		if err != nil {
			Logger.Errorf("Failed to open Burp File: %s", err)
		} else {
			rd := bufio.NewReader(bF)
			req, err := http.ReadRequest(rd)
			if err != nil {
				Logger.Errorf("Failed to Parse Raw Request in %s: %s", burpFile, err)
			} else {
				// Set cookie
				c.OnRequest(func(r *colly.Request) {
					r.Headers.Set("Cookie", GetRawCookie(req.Cookies()))
				})

				// Set headers
				c.OnRequest(func(r *colly.Request) {
					for k, v := range req.Header {
						r.Headers.Set(strings.TrimSpace(k), strings.TrimSpace(v[0]))
					}
				})

			}
		}
	}

	// Set cookies
	cookie, _ := cmd.Flags().GetString("cookie")
	if cookie != "" && burpFile == "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", cookie)
		})
	}

	// Set headers
	headers, _ := cmd.Flags().GetStringArray("header")
	if burpFile == "" {
		for _, h := range headers {
			headerArgs := strings.SplitN(h, ":", 2)
			headerKey := strings.TrimSpace(headerArgs[0])
			headerValue := strings.TrimSpace(headerArgs[1])
			c.OnRequest(func(r *colly.Request) {
				r.Headers.Set(headerKey, headerValue)
			})
		}
	}

	// Set User-Agent
	randomUA, _ := cmd.Flags().GetString("user-agent")
	switch ua := strings.ToLower(randomUA); {
	case ua == "mobi":
		extensions.RandomMobileUserAgent(c)
	case ua == "web":
		extensions.RandomUserAgent(c)
	default:
		c.UserAgent = ua
	}

	// Set referer
	extensions.Referer(c)

	// Disable redirect
	noRedirect, _ := cmd.Flags().GetBool("no-redirect")
	if noRedirect {
		c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		})
	}

	// Init Output
	var output *Output
	outputFolder, _ := cmd.Flags().GetString("output")
	if outputFolder != "" {
		filename := strings.ReplaceAll(domain, ".", "_")
		output = NewOutput(outputFolder, filename)
	}

	// Set url whitelist regex
	domainRegex := "^(https?|mms|mssx|mmsh|rtsp|pnm)://([^/]+[.])?(?i:" + strings.ReplaceAll(domain, ".", "[.]") + ")(/.*)?$"
	domainRe := regexp.MustCompile(domainRegex)
	c.URLFilters = append(c.URLFilters, domainRe)

	// Set Limit Rule
	err := c.Limit(&colly.LimitRule{
		DomainRegexp: domainRegex,
		Parallelism:  concurrent,
		Delay:        time.Duration(delay) * time.Second,
		RandomDelay:  time.Duration(randomDelay) * time.Second,
	})
	if err != nil {
		Logger.Errorf("Failed to set Limit Rule: %s", err)
		os.Exit(1)
	}

	// Set blacklist url regex
	disallowedRegex := `.(jpg|jpeg|gif|css|tif|tiff|png|ttf|woff|ico)(?:\?|#|$)`
	c.DisallowedURLFilters = append(c.DisallowedURLFilters, regexp.MustCompile(disallowedRegex))

	// Set optional blacklist url regex
	blacklists, _ := cmd.Flags().GetString("blacklist")
	if blacklists != "" {
		c.DisallowedURLFilters = append(c.DisallowedURLFilters, regexp.MustCompile(blacklists))
	}

	return &Crawler{
		cmd:      cmd,
		C:        c,
		site:     site,
		domain:   domain,
		Output:   output,
		domainRe: domainRe,
	}
}

func (crawler *Crawler) Start() {
	// Handle url
	urlSet := stringset.NewStringFilter()
	crawler.C.OnHTML("[href]", func(e *colly.HTMLElement) {
		urlString := e.Request.AbsoluteURL(e.Attr("href"))
		urlString = FixUrl(urlString, crawler.site)
		if urlString == "" {
			return
		}
		if !urlSet.Duplicate(urlString) {
			_ = e.Request.Visit(urlString)
		}
	})

	// Handle form
	formSet := stringset.NewStringFilter()
	crawler.C.OnHTML("form[action]", func(e *colly.HTMLElement) {
		formUrl := e.Request.AbsoluteURL(e.Attr("action"))
		formUrl = FixUrl(formUrl, crawler.site)
		if formUrl == "" {
			return
		}
		// Just print
		if !formSet.Duplicate(formUrl) {
			if crawler.domainRe.MatchString(formUrl) {
				outputFormat := fmt.Sprintf("[form] - %s", formUrl)
				Logger.Info(outputFormat + "\n")
				if crawler.Output != nil {
					crawler.Output.WriteToFile(outputFormat)
				}
			}
		}
	})

	// Handle js files
	jsFileSet := stringset.NewStringFilter()
	crawler.C.OnHTML("[src]", func(e *colly.HTMLElement) {
		jsFileUrl := e.Request.AbsoluteURL(e.Attr("src"))
		jsFileUrl = FixUrl(jsFileUrl, crawler.site)
		if jsFileUrl == "" {
			return
		}
		if !jsFileSet.Duplicate(jsFileUrl) {
			outputFormat := fmt.Sprintf("[javascript] - %s", jsFileUrl)
			Logger.Info(outputFormat + "\n")
			if crawler.Output != nil {
				crawler.Output.WriteToFile(outputFormat)
			}
			// Request and Get JS link
			crawler.linkFinder(crawler.site, jsFileUrl)
		}
	})

	subSet := stringset.NewStringFilter()
	awsSet := stringset.NewStringFilter()
	crawler.C.OnResponse(func(response *colly.Response) {
		respStr := string(response.Body)

		//// Use regex to get links in all response code
		//urls, err := ParseJSSource(respStr, site)
		//if err != nil {
		//	Logger.Error(err)
		//	return
		//}
		//// request with depth = 0
		//for _, u := range urls {
		//	Logger.Debugf("Regex url: %s", u)
		//	_ = c.Visit(u)
		//}

		// Parse subdomains from source
		subs := GetSubdomains(respStr, crawler.domain)
		for _, sub := range subs {
			if !subSet.Duplicate(sub) {
				outputFormat := fmt.Sprintf("[subdomains] - %s", sub)
				Logger.Info(outputFormat + "\n")
				if crawler.Output != nil {
					crawler.Output.WriteToFile(outputFormat)
				}
			}
		}

		// Grep AWS S3 from source
		aws := GetAWSS3(respStr)
		for _, e := range aws {
			if !awsSet.Duplicate(e) {
				outputFormat := fmt.Sprintf("[aws-s3] - %s", e)
				Logger.Info(outputFormat + "\n")
				if crawler.Output != nil {
					crawler.Output.WriteToFile(outputFormat)
				}
			}
		}

		// We will pass 404 Not Found and 429 status code
		if response.StatusCode == 404 || response.StatusCode == 429 {
			return
		}

		// Verify which links are working
		u := response.Request.URL.String()
		if crawler.domainRe.MatchString(u) {
			outputFormat := fmt.Sprintf("[url] - %d - %s", response.StatusCode, u)
			Logger.Info(outputFormat + "\n")
			if crawler.Output != nil {
				crawler.Output.WriteToFile(outputFormat)
			}
		}
	})
	_ = crawler.C.Visit(crawler.site)
	crawler.C.Wait()
}

// This function will request and parse external javascript
// and pass to main collector with scope setup
func (crawler *Crawler) linkFinder(site string, jsUrl string) {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(jsUrl)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	// if the js file exists
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
	links, err := ParseJSSource(string(res))
	if err != nil {
		Logger.Error(err)
		return
	}

	for _, link := range links {
		// If link is url, check with regex to make sure it in scope
		_, err := url.Parse(link)
		if err == nil {
			if !crawler.domainRe.MatchString(link) {
				continue
			}
		}

		if strings.HasPrefix(link, "//") {
			newJsPath := "https:" + link
			if !crawler.domainRe.MatchString(newJsPath) {
				continue
			}
		}

		// JS Regex Result
		outputFormat := fmt.Sprintf("[linkfinder] - %s - %s", link, jsUrl)
		Logger.Info(outputFormat + "\n")
		if crawler.Output != nil {
			crawler.Output.WriteToFile(outputFormat)
		}
		// Try to request JS path
		_ = crawler.C.Visit(FixUrl(link, site))
	}

}
