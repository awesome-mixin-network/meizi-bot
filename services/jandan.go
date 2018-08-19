package services

import (
	"context"
	"log"

	"github.com/jasonlvhit/gocron"
	clients "github.com/lyricat/meizi-bot/clients"
	"github.com/lyricat/meizi-bot/config"
	"github.com/lyricat/meizi-bot/models"
)

type JandanService struct{}

func GetJandanItems(ctx context.Context, conf *config.SpiderConf) []*clients.Item {
	jandan := clients.NewJandan(
		conf.Name,
		10,
		10,
		1,
	)
	items, err := jandan.Fetch()
	if err != nil {
		log.Fatalln("error", err)
	}
	return items
}

func (service *JandanService) SendPicturesToChannel(ctx context.Context, stats *ServiceStats, conf *config.SpiderConf) {
	items := GetJandanItems(ctx, conf)
	if len(items) == 0 {
		return
	}
	newItems := stats.filterDuplicate(ctx, items, conf.Name)
	// save records to db
	for _, item := range newItems {
		models.CreateItem(ctx, item.Id, item.Name, item.Desc, item.Ref, item.Urls, conf.Name, "jandan")
	}
	SendPicturesToChannel(ctx, newItems, conf.Name)
}

func (service *JandanService) Run(ctx context.Context, conf *config.SpiderConf) error {
	stats := &ServiceStats{}
	gocron.Every(conf.Interval).Minutes().Do(service.SendPicturesToChannel, ctx, stats, conf)
	<-gocron.Start()
	return nil
}
