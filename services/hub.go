package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lyricat/meizi-bot/config"
	"github.com/lyricat/meizi-bot/session"
)

type Hub struct {
	context        context.Context
	serviceConfigs map[string]*config.SpiderConf
	services       map[string]Service
}

func NewHub(db *sql.DB) *Hub {
	hub := &Hub{services: make(map[string]Service), serviceConfigs: make(map[string]*config.SpiderConf)}
	hub.context = session.WithDatabase(context.Background(), db)
	hub.registerServices()
	return hub
}

func (hub *Hub) StartService(name string) error {
	service := hub.services[name]
	config := hub.serviceConfigs[name]
	if service == nil {
		return fmt.Errorf("no service found: %s", name)
	}
	fmt.Printf("start service: %s\n", name)
	ctx := hub.context

	return service.Run(ctx, config)
}

func (hub *Hub) registerServices() {
	for _, spider := range config.Global.Spiders {
		log.Printf("[i] Register service: %s\n", spider.Name)
		if spider.Type == "twitter" {
			hub.services[spider.Name] = &TwitterService{}
		} else if spider.Type == "jandan" {
			hub.services[spider.Name] = &JandanService{}
		}
		hub.serviceConfigs[spider.Name] = spider
	}
}
