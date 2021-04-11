package core

import (
	"fmt"
	"net/url"
	"sync"

	jsoniter "github.com/json-iterator/go"

	"github.com/gocolly/colly/v2"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

func ParseSiteMap(site *url.URL, crawler *Crawler, c *colly.Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	sitemapUrls := []string{"/sitemap.xml", "/sitemap_news.xml", "/sitemap_index.xml", "/sitemap-index.xml", "/sitemapindex.xml",
		"/sitemap-news.xml", "/post-sitemap.xml", "/page-sitemap.xml", "/portfolio-sitemap.xml", "/home_slider-sitemap.xml", "/category-sitemap.xml",
		"/author-sitemap.xml"}

	for _, path := range sitemapUrls {
		// Ignore error when that not valid sitemap.xml path
		Logger.Infof("Trying to find %s", site.String()+path)
		_ = sitemap.ParseFromSite(site.String()+path, func(entry sitemap.Entry) error {
			outputFormat := fmt.Sprintf("[sitemap] - %s", entry.GetLocation())

			if crawler.JsonOutput {
				sout := SpiderOutput{
					Input:      crawler.Input,
					Source:     "sitemap",
					OutputType: "url",
					Output:     entry.GetLocation(),
				}
				if data, err := jsoniter.MarshalToString(sout); err == nil {
					outputFormat = data
				}
			} else if crawler.Quiet {
				outputFormat = entry.GetLocation()
			}
			fmt.Println(outputFormat)
			if crawler.Output != nil {
				crawler.Output.WriteToFile(outputFormat)
			}
			_ = c.Visit(entry.GetLocation())
			return nil
		})
	}

}
