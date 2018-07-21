package session

import (
	"context"
	"database/sql"
)

type contextValueKey int

const (
	keyRequest  contextValueKey = 0
	keyDatabase contextValueKey = 1
	keyLogger   contextValueKey = 2
)

func Database(ctx context.Context) *sql.DB {
	v, _ := ctx.Value(keyDatabase).(*sql.DB)
	return v
}

func WithDatabase(ctx context.Context, database *sql.DB) context.Context {
	return context.WithValue(ctx, keyDatabase, database)
}
