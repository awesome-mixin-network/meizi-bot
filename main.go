package main

import (
	"flag"
	"log"

	"github.com/lyricat/meizi-bot/config"

	"context"

	"github.com/lyricat/meizi-bot/durable"
	"github.com/lyricat/meizi-bot/services"
)

func main() {
	config.LoadConfig()
	service := flag.String("service", "blaze", "run a service")
	flag.Parse()
	db := durable.OpenDatabaseClient(context.Background())
	defer db.Close()

	switch *service {
	case "blaze":
		err := StartBlaze(db)
		if err != nil {
			log.Println(err)
		}
	default:
		hub := services.NewHub(db)
		err := hub.StartService(*service)
		if err != nil {
			log.Println(err)
		}
	}
}
