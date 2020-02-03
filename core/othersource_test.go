package core

import "testing"

var domain = "yahoo.com"

func TestOtherSources(t *testing.T) {
	urls := OtherSources(domain)
	t.Log(len(urls))
	t.Log(urls)
}

func TestGetCommonCrawlURLs(t *testing.T) {
	urls, err := getCommonCrawlURLs(domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(urls))
	t.Log(urls)
}

func TestGetVirusTotalURLs(t *testing.T) {
	urls, err := getVirusTotalURLs(domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(urls))
	t.Log(urls)
}

func TestGetWaybackURLs(t *testing.T) {
	urls, err := getWaybackURLs(domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(urls))
	t.Log(urls)
}
