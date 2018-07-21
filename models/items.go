package models

import (
	"context"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/fox-one/foxone-mixin-bot/session"
	"log"
	"time"
)

type Item struct {
	Id          int
	ItemId      string
	Name        string
	Desc        string
	Ref         string
	Urls        []string
	ServiceName string
	CreatedAt   time.Time
}

func CreateItem(ctx context.Context, itemId string, name string, desc string, ref string, urls []string, serviceName string) (*Item, error) {
	existedItem, err := findExistedItemByItemId(ctx, itemId, serviceName)
	if existedItem == nil {
		urlsJs := json.New()
		urlsJs.Set("urls", urls)
		urlsBytes, _ := urlsJs.EncodePretty()

		item := &Item{
			ItemId:      string(itemId),
			Name:        name,
			Desc:        desc,
			Ref:         ref,
			Urls:        urls,
			ServiceName: serviceName,
			CreatedAt:   time.Now().UTC(),
		}

		_, err = session.Database(ctx).Exec("Insert into items(item_id, name, desc, ref, urls, service_name, created_at) values($1, $2, $3, $4, $5, $6, $7)",
			item.ItemId, item.Name, item.Desc, item.Ref, string(urlsBytes), item.ServiceName, fmt.Sprintf("%v", item.CreatedAt))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return item, nil

	}
	return existedItem, nil
}

func RemoveItem(ctx context.Context, itemId string, serviceName string) error {
	existedItem, err := findExistedItemByItemId(ctx, itemId, serviceName)
	if existedItem == nil {
		return nil
	}
	_, err = session.Database(ctx).Exec("Delete from subscribers where item_id = $1", itemId)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetRandomItem(ctx context.Context) (*Item, error) {
	row := session.Database(ctx).QueryRow("select * from items order by RANDOM() limit 1")
	var tempUrls []byte
	var item Item
	if err := row.Scan(&item.Id, &item.ItemId, &item.Name, &item.Desc, &item.Ref, &tempUrls, &item.ServiceName, &item.CreatedAt); err != nil {
		log.Println(err)
		return nil, err
	}
	js, _ := json.NewJson(tempUrls)
	for _, v := range js.Get("urls").MustArray() {
		item.Urls = append(item.Urls, v.(string))
	}
	return &item, nil
}

func IsExistedItemByItemId(ctx context.Context, itemId string, serviceName string) bool {
	row := session.Database(ctx).QueryRow("select id from items where item_id = $1 and service_name = $2", itemId, serviceName)
	var i int64
	if err := row.Scan(&i); err != nil {
		return false
	}
	return true
}

func findExistedItemByItemId(ctx context.Context, itemId string, serviceName string) (*Item, error) {
	row := session.Database(ctx).QueryRow("select * from items where item_id = $1 and service_name = $2", itemId, serviceName)
	var item Item
	var tempUrls []byte
	if err := row.Scan(&item.Id, &item.ItemId, &item.Name, &item.Desc, &item.Ref, &tempUrls, &item.ServiceName, &item.CreatedAt); err != nil {
		log.Println(err)
		return nil, err
	}
	js, _ := json.NewJson(tempUrls)
	for _, v := range js.Get("urls").MustArray() {
		item.Urls = append(item.Urls, v.(string))
	}
	return &item, nil
}
