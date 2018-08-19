package services

import (
	"context"

	"github.com/lyricat/meizi-bot/config"
)

type Service interface {
	Run(context.Context, *config.SpiderConf) error
}
