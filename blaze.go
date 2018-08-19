package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/lyricat/meizi-bot/config"
	"github.com/lyricat/meizi-bot/session"

	bot "github.com/lyricat/bot-api-go-client"
)

func StartBlaze(db *sql.DB) error {
	log.Println("start blaze")
	ctx, cancel := newBlazeContext(db)
	defer cancel()

	for {
		if err := bot.Loop(ctx, ResponseMessage{}, config.Global.MixinClientId, config.Global.MixinSessionId, config.Global.MixinPrivateKey); err != nil {
			log.Println(err)
		}
		log.Println("connection loop end")
		time.Sleep(time.Second)
	}
	return nil
}

func newBlazeContext(db *sql.DB) (context.Context, context.CancelFunc) {
	ctx := session.WithDatabase(context.Background(), db)
	return context.WithCancel(ctx)
}
