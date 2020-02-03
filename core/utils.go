package core

import (
	"fmt"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func GetRawCookie(cookies []*http.Cookie) string {
	var rawCookies []string
	for _, c := range cookies {
		e := fmt.Sprintf("%s=%s", c.Name, c.Value)
		rawCookies = append(rawCookies, e)
	}
	return strings.Join(rawCookies, "; ")
}

func GetDomain(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return ""
	}
	return domain
}

func FixUrl(url, site string) string {
	var newUrl string

	if strings.HasPrefix(url, "http") {
		newUrl = url
	} else if strings.HasPrefix(url, "//") {
		newUrl = "https:" + url
	} else if !strings.HasPrefix(url, "http") && len(url) > 0 {
		if url[:1] == "/" { // Ex: /?thread=10
			newUrl = site + url
		} else { // Ex: ?thread=10
			newUrl = site + "/" + url
		}
	}
	return newUrl
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func LoadCookies(rawCookie string) []*http.Cookie {
	httpCookies := []*http.Cookie{}
	cookies := strings.Split(rawCookie, ";")
	for _, cookie := range cookies {
		cookieArgs := strings.SplitN(cookie, "=", 2)
		if len(cookieArgs) > 2 {
			continue
		}

		ck := &http.Cookie{Name: strings.TrimSpace(cookieArgs[0]), Value: strings.TrimSpace(cookieArgs[1])}
		httpCookies = append(httpCookies, ck)
	}
	return httpCookies
}

func GetExtType(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return ""
	}
	return path.Ext(u.Path)
}
