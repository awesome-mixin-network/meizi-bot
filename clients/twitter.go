package clients

import (
	"fmt"
	"log"
	"strings"
	"time"
	// "encoding/json"

	. "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Twitter struct {
	client *Client

	Name      string
	Interval  int
	MediaOnly bool

	Query    string
	Username string

	MinFavCount int
	MinRetCount int
}

// when query and username both exist, will fetch by username and filter by query
func NewTwitter(name,
	consumerKey, consumerSecret, token, tokenSecret,
	query, username string, mediaOnly bool,
	minFavCount, minRetCount int,
	interval int) *Twitter {

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	oauthToken := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, oauthToken)

	// Twitter client
	client := NewClient(httpClient)

	return &Twitter{
		client:      client,
		Name:        name,
		Interval:    interval,
		MediaOnly:   mediaOnly,
		Query:       query,
		Username:    username,
		MinFavCount: minFavCount,
		MinRetCount: minRetCount,
	}
}

func (c *Twitter) GetName() string {
	return "Twitter-" + c.Name
}

func (c *Twitter) GetInterval() time.Duration {
	return time.Minute * time.Duration(c.Interval)
}

func (c *Twitter) Fetch() (results []*Item, err error) {
	var tweets []Tweet

	if c.Query != "" {
		params := &SearchTweetParams{
			Query: c.Query,
			Count: 20,
		}

		search, _, err1 := c.client.Search.Tweets(params)
		if err1 != nil {
			log.Println("error", err1)
			return
		}

		tweets = search.Statuses
		log.Printf("[%s] fetched %d with query: %s\n", c.GetName(), len(tweets), c.Query)
	} else if c.Username != "" {
		params := &UserTimelineParams{ScreenName: c.Username}
		tweets, _, err = c.client.Timelines.UserTimeline(params)
		log.Printf("[%s] fetched %d with username: %s\n", c.GetName(), len(tweets), c.Username)

		if err != nil {
			log.Println("error", err)
			return
		}
	}
	// res2B, _ := json.Marshal(tweets)
	// fmt.Println(string(res2B))

	for _, t := range tweets {
		if c.Username != "" {
			if t.User.ScreenName != c.Username {
				continue
			}
		}
		if c.Query != "" {
			if !strings.Contains(t.Text, c.Query) {
				continue
			}
		}

		item := c.toItem(t)
		if item != nil {
			// log.Printf("[%s] %s %s\n", c.GetName(), t.Text, item.Ref)
			results = append(results, item)
		}
	}
	return
}

func (c *Twitter) toItem(t Tweet) *Item {
	if t.FavoriteCount < c.MinFavCount {
		log.Println("fav count:", t.FavoriteCount)
		return nil
	}

	if t.RetweetCount < c.MinRetCount {
		log.Println("ret count:", t.RetweetCount)
		return nil
	}

	animated := false

	var media []string = nil

	if t.ExtendedEntities != nil {
		for _, m := range t.ExtendedEntities.Media {
			if m.Type == "animated_gif" {
				animated = true
				media = append(media, m.VideoInfo.Variants[0].URL)
			} else if m.MediaURLHttps != "" {
				media = append(media, m.MediaURLHttps)
			}
		}
	}
	// res2B, _ := json.Marshal(t)
	// fmt.Println(string(res2B))
	if len(media) == 0 && t.Entities != nil && !animated {
		for _, m := range t.Entities.Media {
			if m.MediaURLHttps != "" {
				media = append(media, m.MediaURLHttps)
			}
		}
	} else {
		log.Println("no media found")
	}
	log.Printf("media: %s\n", media)

	if c.MediaOnly && len(media) == 0 {
		return nil
	}

	ref := fmt.Sprintf("https://twitter.com/%s/status/%d", t.User.ScreenName, t.ID)

	desc := ""
	created := time.Time{}
	desc = fmt.Sprintf("%s %s", c.Username, c.Query)

	return &Item{
		Id:      fmt.Sprintf("%d", t.ID),
		Name:    c.GetName(),
		Desc:    desc,
		Ref:     ref,
		Created: created,
		Urls:    media,
	}
}
