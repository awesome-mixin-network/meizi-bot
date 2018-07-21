package clients

import (
	"time"
)

type Item struct {
	Id      string
	Name    string
	Desc    string
	Ref     string
	Created time.Time
	Urls    []string
}
