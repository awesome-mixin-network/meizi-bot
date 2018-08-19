package services

import (
	"context"

	clients "github.com/lyricat/meizi-bot/clients"

	"log"

	"github.com/jasonlvhit/gocron"
	"github.com/lyricat/meizi-bot/config"
	"github.com/lyricat/meizi-bot/models"
)

type TwitterService struct{}

func sendTopPicturesToChannel(ctx context.Context, stats *ServiceStats, conf *config.SpiderConf) {
	items := GetTwitterItems(ctx, conf)
	if len(items) == 0 {
		return
	}
	newItems := stats.filterDuplicate(ctx, items, conf.Name)
	// save records to db
	for _, item := range newItems {
		models.CreateItem(ctx, item.Id, item.Name, item.Desc, item.Ref, item.Urls, conf.Name, "twitter")
	}
	SendPicturesToChannel(ctx, newItems, conf.Name)
}

func GetTwitterItems(ctx context.Context, conf *config.SpiderConf) []*clients.Item {
	twit := clients.NewTwitter(
		conf.Name,
		config.Global.TwitterConsumerKey,
		config.Global.TwitterConsumerSecret,
		config.Global.TwitterToken,
		config.Global.TwitterTokenSecret,
		conf.Query,
		conf.Username,
		true,
		0,
		0,
		1)
	items, err := twit.Fetch()
	log.Printf("%+v\n", items)
	if err != nil {
		log.Fatalln("error", err)
	}
	return items
}

func (service *TwitterService) Run(ctx context.Context, conf *config.SpiderConf) error {
	// items := GetTwitterItems(ctx)
	stats := &ServiceStats{}
	// stats.filterDuplicate(ctx, items)
	gocron.Every(conf.Interval).Minutes().Do(sendTopPicturesToChannel, ctx, stats, conf)
	<-gocron.Start()
	return nil
}
