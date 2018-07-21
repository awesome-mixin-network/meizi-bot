package services

import (
	"context"
	clients "github.com/fox-one/foxone-mixin-bot/clients"
	"github.com/fox-one/foxone-mixin-bot/config"
	"github.com/fox-one/foxone-mixin-bot/models"
	"github.com/jasonlvhit/gocron"
	"log"
)

type JandanService struct{}

func GetJandanItems(ctx context.Context) []*clients.Item {
	jandan := clients.NewJandan(
		"jandan-meizi",
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

func (service *JandanService) SendPicturesToChannel(ctx context.Context, stats *ServiceStats) {
	items := GetJandanItems(ctx)
	if len(items) == 0 {
		return
	}
	newItems := stats.filterDuplicate(ctx, items, "jandan")
	// save records to db
	for _, item := range newItems {
		models.CreateItem(ctx, item.Id, item.Name, item.Desc, item.Ref, item.Urls, "jandan")
	}
	SendPicturesToChannel(ctx, newItems, "jandan")
}

func (service *JandanService) Run(ctx context.Context) error {
	stats := &ServiceStats{}
	gocron.Every(config.JandanInterval).Minutes().Do(service.SendPicturesToChannel, ctx, stats)
	<-gocron.Start()
	return nil
}
