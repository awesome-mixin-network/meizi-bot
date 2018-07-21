package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/fox-one/foxone-mixin-bot/config"
	"github.com/fox-one/foxone-mixin-bot/session"

	bot "github.com/lyricat/bot-api-go-client"
)

func StartBlaze(db *sql.DB) error {
	log.Println("start blaze")
	ctx, cancel := newBlazeContext(db)
	defer cancel()

	for {
		if err := bot.Loop(ctx, ResponseMessage{}, config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey); err != nil {
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
