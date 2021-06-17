package core

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/net/publicsuffix"
)

var nameStripRE = regexp.MustCompile("(?i)^((20)|(25)|(2b)|(2f)|(3d)|(3a)|(40))+")

func GetRawCookie(cookies []*http.Cookie) string {
	var rawCookies []string
	for _, c := range cookies {
		e := fmt.Sprintf("%s=%s", c.Name, c.Value)
		rawCookies = append(rawCookies, e)
	}
	return strings.Join(rawCookies, "; ")
}

func GetDomain(site *url.URL) string {
	domain, err := publicsuffix.EffectiveTLDPlusOne(site.Hostname())
	if err != nil {
		return ""
	}
	return domain
}

// func FixUrl(site *url.URL, nextLoc string) string {
//     var newUrl string
//     if strings.HasPrefix(nextLoc, "//") {
//         // //google.com/example.php
//         newUrl = site.Scheme + ":" + nextLoc
//
//     } else if strings.HasPrefix(nextLoc, "http") {
//         // http://google.com || https://google.com
//         newUrl = nextLoc
//
//     } else if !strings.HasPrefix(nextLoc, "//") {
//         // if strings.HasPrefix(nextLoc, "/") {
//         //     // Ex: /?thread=10
//         //     newUrl = site.Scheme + "://" + site.Host + nextLoc
//         //
//         // } else {
//         //     if strings.HasPrefix(nextLoc, ".") {
//         //         if strings.HasPrefix(nextLoc, "..") {
//         //             newUrl = site.Scheme + "://" + site.Host + nextLoc[2:]
//         //         } else {
//         //             newUrl = site.Scheme + "://" + site.Host + nextLoc[1:]
//         //         }
//         //     } else {
//         //         // "console/test.php"
//         //         newUrl = site.Scheme + "://" + site.Host + "/" + nextLoc
//         //     }
//         // }
//         nextLocUrl, err := url.Parse(nextLoc)
//         if err != nil {
//             return ""
//         }
//         newUrl = site.ResolveReference(nextLocUrl).String()
//     }
//     return newUrl
// }

func FixUrl(mainSite *url.URL, nextLoc string) string {
	nextLocUrl, err := url.Parse(nextLoc)
	if err != nil {
		return ""
	}
	return mainSite.ResolveReference(nextLocUrl).String()
}

func Unique(intSlice []string) []string {
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

func CleanSubdomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "*.")
	// s = strings.Trim("u00","")
	s = cleanName(s)
	return s
}

// Clean up the names scraped from the web.
// Get from Amass
func cleanName(name string) string {
	for {
		if i := nameStripRE.FindStringIndex(name); i != nil {
			name = name[i[1]:]
		} else {
			break
		}
	}

	name = strings.Trim(name, "-")
	// Remove dots at the beginning of names
	if len(name) > 1 && name[0] == '.' {
		name = name[1:]
	}
	return name
}

func FilterNewLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

func DecodeChars(s string) string {
	source, err := url.QueryUnescape(s)
	if err == nil {
		s = source
	}

	// In case json encoded chars
	replacer := strings.NewReplacer(
		`\u002f`, "/",
		`\u0026`, "&",
	)
	s = replacer.Replace(s)
	return s
}

func InScope(u *url.URL, regexps []*regexp.Regexp) bool {
    for _, r := range regexps {
        if r.MatchString(u.String()) {
            return true
        }
    }
    return false
}

// NormalizePath the path
func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~") {
		path, _ = homedir.Expand(path)
	}
	return path
}

// ReadingLines Reading file and return content as []string
func ReadingLines(filename string) []string {
	var result []string
	if strings.HasPrefix(filename, "~") {
		filename, _ = homedir.Expand(filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			continue
		}
		result = append(result, val)
	}

	if err := scanner.Err(); err != nil {
		return result
	}
	return result
}

func contains(i []int,j int) bool {
    for _, value := range i {
        if value == j {
            return true
        }
    }
    return false
}