package core

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"
	"net/url"
	"sync"
)

func ParseSiteMap(site *url.URL, output *Output, c *colly.Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	sitemapUrls := []string{"/sitemap.xml", "/sitemap_news.xml", "/sitemap_index.xml", "/sitemap-index.xml", "/sitemapindex.xml",
		"/sitemap-news.xml", "/post-sitemap.xml", "/page-sitemap.xml", "/portfolio-sitemap.xml", "/home_slider-sitemap.xml", "/category-sitemap.xml",
		"/author-sitemap.xml"}

	for _, path := range sitemapUrls {
		// Ignore error when that not valid sitemap.xml path
		Logger.Infof("Trying to find %s", site.String()+path)
		_ = sitemap.ParseFromSite(site.String()+path, func(entry sitemap.Entry) error {
			outputFormat := fmt.Sprintf("[sitemap] - %s", entry.GetLocation())
			fmt.Println(outputFormat)
			if output != nil {
				output.WriteToFile(outputFormat)
			}
			_ = c.Visit(entry.GetLocation())
			return nil
		})
	}

}
