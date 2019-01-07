package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type SpiderConf struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Interval uint64 `json:"interval"`
	Query    string `json:"query"`
	Username string `json:"username"`
}

type Conf struct {
	DatabasePath      string `json:"database_path"`
	MixinClientId     string `json:"mixin_client_id"`
	MixinClientSecret string `json:"mixin_client_secret"`
	MixinPin          string `json:"mixin_pin"`
	MixinSessionId    string `json:"mixin_session_id"`
	MixinPinToken     string `json:"mixin_pin_token"`
	MixinPrivateKey   string `json:"mixin_private_key"`
	// for twitter
	TwitterConsumerKey    string `json:"twitter_consumer_key"`
	TwitterConsumerSecret string `json:"twitter_consumer_secret"`
	TwitterToken          string `json:"twitter_token_key"`
	TwitterTokenSecret    string `json:"twitter_token_secret"`
	// spider settings
	Spiders []*SpiderConf `json:"spiders"`
}

var Global *Conf

func LoadConfig() {
	Global = &Conf{}
	Global.Spiders = make([]*SpiderConf, 0)
	j, _ := ioutil.ReadFile("deploy/config.json")
	json.Unmarshal(j, Global)
	PrintConfig()
}

func PrintConfig() {
	log.Println("[i]Load Config")
	log.Printf("- DatabasePath: %s\n", Global.DatabasePath)
	log.Printf("- MixinClientId: %s\n", Global.MixinClientId)
	log.Printf("- MixinClientSecret: %s\n", Global.MixinClientSecret)
	log.Printf("- MixinPin: %s\n", Global.MixinPin)
	log.Printf("- MixinSessionId: %s\n", Global.MixinSessionId)
	log.Printf("- MixinPinToken: %s\n", Global.MixinPinToken)
	log.Printf("- MixinPrivateKey: %s\n", Global.MixinPrivateKey)
	log.Printf("- TwitterConsumerKey: %s\n", Global.TwitterConsumerKey)
	log.Printf("- TwitterConsumerSecret: %s\n", Global.TwitterConsumerSecret)
	log.Printf("- TwitterToken: %s\n", Global.TwitterToken)
	log.Printf("- TwitterTokenSecret: %s\n", Global.TwitterTokenSecret)
}
