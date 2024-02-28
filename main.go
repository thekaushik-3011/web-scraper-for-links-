package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/PuerkitoBio/goquery"
	// "time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X_10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X_10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X_10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko)  Version/11.0 Safari/604.1.38",
}

func randomUserAgent() string {
	// rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func discoverLinks(response *http.Response) []string {
	if response != nil {
		body := response.Body
		doc, _ := goquery.NewDocumentFromReader(body)
		foundURLs := []string{}
		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection){
				res, _ := s.Attr("href")
				foundURLs = append(foundURLs, res)
			})
		}
		return foundURLs
	} else {
		return []string{}
	}
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil{
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())

	res, err := client.Do(req)
	if err != nil{
		return nil, err
	} else {
		return res, nil
	}
}

func checkRelative(href string, baseURL string) string{
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseURL, href)
	} else {
		return ""
	}
}

func resolveRelativeLinks(href string, baseURL string) (bool, string) {
	resultHref := checkRelative(href, baseURL)
	baseParse, _ := url.Parse(baseURL)
	resultParse, _ := url.Parse(resultHref)
	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		} else {
			return false, ""
		}
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)

func Crawl(targetURL string, baseURL string) []string {
	fmt.Println(targetURL)
	tokens <- struct{}{}
	resp, _ := getRequest(targetURL)
	<-tokens

	links := discoverLinks(resp)
	foundURLs := []string{}

	for _, link := range links{
		ok, correctLink := resolveRelativeLinks(link, baseURL)
		if ok {
			if correctLink != "" {
				foundURLs = append(foundURLs, correctLink)
			}
		}
	}
	// ParseHTML(res)
	return foundURLs
}

// func ParseHTML(response *http.Response) {
// 	//fill this
// }

func main() {
	workList := make(chan []string)
	var n int
	n++

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the URL: ")
	baseDomain, _ := reader.ReadString('\n')
	baseDomain = strings.TrimSpace(baseDomain)

	// baseDomain := "http://www.theguardian.com"
	go func(){workList <- []string{baseDomain} }()

	seen := make(map[string]bool)

	for; n>0; n--{
		list := <- workList

		for _, link := range list{
			if !seen[link]{
				seen[link] = true
				n++

				go func(link string, baseURL string) {
					foundLinks := Crawl(link, baseDomain)
					if foundLinks != nil{
						workList <- foundLinks
					}
				}(link, baseDomain)
			}
		}
	}
}