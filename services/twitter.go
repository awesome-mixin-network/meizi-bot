package services

import (
	"context"
	clients "github.com/fox-one/foxone-mixin-bot/clients"

	"github.com/fox-one/foxone-mixin-bot/config"
	"github.com/fox-one/foxone-mixin-bot/models"
	"github.com/jasonlvhit/gocron"
	"log"
)

type TwitterService struct{}

func sendTopPicturesToChannel(ctx context.Context, stats *ServiceStats) {
	items := GetTwitterItems(ctx)
	if len(items) == 0 {
		return
	}
	newItems := stats.filterDuplicate(ctx, items, "twitter")
	// save records to db
	for _, item := range newItems {
		models.CreateItem(ctx, item.Id, item.Name, item.Desc, item.Ref, item.Urls, "twitter")
	}
	SendPicturesToChannel(ctx, newItems, "twitter")
}

func GetTwitterItems(ctx context.Context) []*clients.Item {
	twit := clients.NewTwitter(
		"meizi",
		config.TwitterConsumerKey,
		config.TwitterConsumerSecret,
		config.TwitterToken,
		config.TwitterTokenSecret,
		config.TwitterQuery,
		config.TwitterUsername,
		true,
		0,
		0,
		1)
	items, err := twit.Fetch()
	if err != nil {
		log.Fatalln("error", err)
	}
	return items
}

func (service *TwitterService) Run(ctx context.Context) error {
	// items := GetTwitterItems(ctx)
	stats := &ServiceStats{}
	// stats.filterDuplicate(ctx, items)
	gocron.Every(config.TwitterInterval).Minutes().Do(sendTopPicturesToChannel, ctx, stats)
	<-gocron.Start()
	return nil
}
