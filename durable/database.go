package durable

import (
	"context"
	"database/sql"
	"log"

	"github.com/lyricat/meizi-bot/config"
	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabaseClient(ctx context.Context) *sql.DB {
	log.Printf("[i] Open Database: %s\n", config.Global.DatabasePath)
	db, err := sql.Open("sqlite3", config.Global.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
