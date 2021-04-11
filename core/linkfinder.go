package core

import (
	"regexp"
	"strings"
)

var linkFinderRegex = regexp.MustCompile(`(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{3,}(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:[\?|#][^"|']{0,}|)))(?:"|')`)

func LinkFinder(source string) ([]string, error) {
	var links []string
	// source = strings.ToLower(source)
	if len(source) > 1000000 {
		source = strings.ReplaceAll(source, ";", ";\r\n")
		source = strings.ReplaceAll(source, ",", ",\r\n")
	}
	source = DecodeChars(source)

	match := linkFinderRegex.FindAllStringSubmatch(source, -1)
	for _, m := range match {
		matchGroup1 := FilterNewLines(m[1])
		if matchGroup1 == "" {
			continue
		}
		links = append(links, matchGroup1)
	}
	links = Unique(links)
	return links, nil
}
