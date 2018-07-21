package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fox-one/foxone-mixin-bot/session"
)

type Hub struct {
	context  context.Context
	services map[string]Service
}

func NewHub(db *sql.DB) *Hub {
	hub := &Hub{services: make(map[string]Service)}
	hub.context = session.WithDatabase(context.Background(), db)
	hub.registerServices()
	return hub
}

func (hub *Hub) StartService(name string) error {
	service := hub.services[name]
	if service == nil {
		return fmt.Errorf("no service found: %s", name)
	}
	fmt.Printf("start service: %s\n", name)
	ctx := hub.context

	return service.Run(ctx)
}

func (hub *Hub) registerServices() {
	hub.services["twitter"] = &TwitterService{}
	hub.services["jandan"] = &JandanService{}
}
