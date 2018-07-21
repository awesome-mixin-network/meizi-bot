package clients

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Jandan struct {
	Name          string
	MinFavCount   int
	MaxUnfavCount int
	Interval      int
}

func NewJandan(name string, like, unlike, interval int) *Jandan {
	return &Jandan{
		Name:          name,
		Interval:      interval,
		MinFavCount:   like,
		MaxUnfavCount: unlike,
	}
}

func findWatchingPageUrls() []string {
	// Request the HTML page.
	res, err := http.Get("http://jandan.net/ooxx/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the link of second page
	nextPageLink := doc.Find(".previous-comment-page").Eq(0)
	nextPageHref, exists := nextPageLink.Attr("href")
	if exists {
		fmt.Println("nextPageHref", nextPageHref)
	}
	if !strings.HasPrefix(nextPageHref, "http") {
		nextPageHref = "http:" + nextPageHref
	}
	var ret []string
	ret = append(ret, nextPageHref)
	return ret
}

func (c *Jandan) toItem(id, desc, ref string, like, unlike int, url string) *Item {
	if like < c.MinFavCount {
		return nil
	}

	if unlike > c.MaxUnfavCount {
		return nil
	}

	if strings.HasSuffix(strings.ToLower(url), ".gif") {
		return nil
	}

	if !strings.HasPrefix(url, "http") {
		url = "http:" + url
	}

	if !strings.HasPrefix(ref, "http") {
		ref = "http:" + ref
	}

	return &Item{
		Id:      id,
		Name:    c.Name,
		Desc:    desc,
		Ref:     ref,
		Created: time.Time{},
		Urls:    []string{url},
	}
}

func (c *Jandan) Fetch() ([]*Item, error) {
	urls := findWatchingPageUrls()
	var ret []*Item
	for _, url := range urls {
		// visit this page
		res, err := http.Get(url)
		defer res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		doc.Find(".commentlist li").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			id, _ := s.Attr("id")
			desc := s.Find(".author").Text()
			ref, _ := s.Find(".righttext a").Attr("href")
			imageHash := s.Find(".img-hash").Text()
			if len(imageHash) != 0 {
				imageUrl, _ := base64.StdEncoding.DecodeString(imageHash)
				if len(imageUrl) != 0 && strings.HasPrefix(string(imageUrl), "//") {
					like, _ := strconv.Atoi(s.Find(".tucao-like-container span").Text())
					unlike, _ := strconv.Atoi(s.Find(".tucao-unlike-container span").Text())
					if like != 0 && unlike != 0 {
						// fmt.Printf("Picture (%d, %d): %s\n", like, unlike, imageUrl)
						newItem := c.toItem(id, desc, ref, like, unlike, string(imageUrl))
						if newItem != nil {
							ret = append(ret, newItem)
						}
					}
				}
			}
		})
	}
	return ret, nil
}
