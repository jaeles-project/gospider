# gospider
**gospider** is a simple spider written in Go.
 
## Installation
```
go get -u github.com/theblackturtle/gospider
```

## Features
* Fast web crawling
* Brute force and parse sitemap.xml
* Parse robots.txt
* Parse and verify link from JavaScript files
* Find AWS-S3 from response source
* Find Subdomain from response source
* Get URLs from Wayback Machine, Common Crawl, Virus Total
* Format output easy to Grep
* Support Burp input
* Crawl multiple sites in parallel
* Random mobile/web User-Agent

## Usage
```
A Simple Web Spider by @theblackturtle

Usage:
  gospider [flags]

Flags:
  -s, --site string          Site to crawl
  -S, --sites string         Site list to crawl
  -p, --proxy string         Proxy (Ex: http://127.0.0.1:8080)
  -o, --output string        Output folder
  -u, --user-agent string    User Agent to use
                                web: random web user-agent
                                mobi: random mobile user-agent
                                or you can set your special user-agent (default "web")
      --cookie string        Cookie to use (testA=a; testB=b)
  -H, --header stringArray   Header to use (Use multiple flag to set multiple header)
      --burp string          Load headers and cookie from burp raw http request
      --blacklist string     Blacklist URL Regex
  -t, --threads int          Number of threads (Run sites in parallel) (default 1)
  -c, --concurrent int       The number of the maximum allowed concurrent requests of the matching domains (default 5)
  -d, --depth int            MaxDepth limits the recursion depth of visited URLs. (Set it to 0 for infinite recursion) (default 1)
  -k, --delay int            Delay is the duration to wait before creating a new request to the matching domains (second)
  -K, --random-delay int     RandomDelay is the extra randomized duration to wait added to Delay before creating a new request (second)
      --sitemap              Try to crawl sitemap.xml
      --robots               Try to crawl robots.txt (default true)
      --other-source         Find url from another sources
      --debug                Turn on debug mode
      --no-redirect          Disable redirect
  -h, --help                 help for gospider
```

## Example commands
#### Run with single site
```
webcrawler -s "https://google.com/" -o output -c 10 -d 1 --other-source
```

#### Run with site list
```
webcrawler -S sites.txt -o output -c 10 -d 1 -t 5 --other-source
```

#### Use custom header/cookies
```
webcrawler -s "https://google.com/" -o output -c 10 -d 1 --other-source -H "Accept: */*" --cookie "testA=a; testB=b"

webcrawler -s "https://google.com/" -o output -c 10 -d 1 --other-source --burp burp_req.txt
```

#### Blacklist url/file extension
```
webcrawler -s "https://google.com/" -o output -c 10 -d 1 --blacklist ".(woff|pdf)"
   ```
