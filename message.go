package main

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"bytes"
	"log"

	bot "github.com/lyricat/bot-api-go-client"
	"github.com/lyricat/meizi-bot/config"
	"github.com/lyricat/meizi-bot/models"
	"github.com/lyricat/meizi-bot/services"
	uuid "github.com/nu7hatch/gouuid"
)

const InstructionHint = "发送 start 订阅妹子图\n发送 stop 取消订阅妹子图\n发送 random 随机看妹子\n发送 donate 请我喝牛奶"

type ResponseMessage struct {
}

type Button struct {
	Label  string `json:"label"`
	Color  string `json:"color"`
	Action string `json:"action"`
	Asset  string `json:-`
	Amount string `json:-`
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func (r ResponseMessage) OnMessage(ctx context.Context, mc *bot.MessageContext, msg bot.MessageView) error {
	if msg.Category != bot.MessageCategorySystemAccountSnapshot && msg.Category != bot.MessageCategorySystemConversation && msg.ConversationId == bot.UniqueConversationId(config.Global.MixinClientId, msg.UserId) {
		if msg.Category == "PLAIN_TEXT" {
			data, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				return bot.BlazeServerError(ctx, err)
			}
			if "start" == string(data) {
				_, err = models.CreateSubscriber(ctx, msg.UserId)
				if err == nil {
					if err := bot.SendPlainText(ctx, mc, msg, "订阅成功"); err != nil {
						return bot.BlazeServerError(ctx, err)
					}
				}
			} else if "stop" == string(data) {
				err = models.RemoveSubscriber(ctx, msg.UserId)
				if err == nil {
					if err := bot.SendPlainText(ctx, mc, msg, "已取消订阅"); err != nil {
						return bot.BlazeServerError(ctx, err)
					}
				}
			} else if "random" == string(data) {
				url := services.GetRandomPicture(ctx)
				if url != "" {
					filepath, _ := services.DownloadMedia(ctx, url, "twitter")
					imageData, _ := services.CreateAttachment(ctx, filepath)
					msgType := services.GetAttachmentType(filepath)
					if err := bot.PostMessage(ctx, msg.ConversationId, msg.UserId, bot.NewV4().String(), msgType, imageData, config.Global.MixinClientId, config.Global.MixinSessionId, config.Global.MixinPrivateKey); err != nil {
						log.Println("err", err)
					}
				} else {
					bot.SendPlainText(ctx, mc, msg, "No media in database")
				}
			} else if "fetch" == string(data) {
				bot.SendPlainText(ctx, mc, msg, "I know, I know, be patient.")
				service := &services.TwitterService{}
				config := config.Global.Spiders[2]
				service.Run(ctx, config)
			} else if "debug" == string(data) {
				var user *models.Subscriber
				user, err = models.FindSubscriberById(ctx, msg.UserId)
				if err == nil {
					if err := bot.SendPlainText(ctx, mc, msg, "你正在订阅妹子图。ID："+user.UserId); err != nil {
						return bot.BlazeServerError(ctx, err)
					}
				} else {
					bot.SendPlainText(ctx, mc, msg, "你没有在订阅妹子图。ID："+msg.UserId)
				}
			} else if "donate" == string(data) {
				recipient := "983cb9e0-4812-4870-a358-aa9f5bb969a9" // bot
				if err != nil {
					log.Println("error:", err)
					return bot.BlazeServerError(ctx, err)
				}
				buttonInfos := []Button{
					Button{
						Label:  "5 PRS",
						Color:  "#5555CC",
						Action: "",
						Amount: "5",
						Asset:  "3edb734c-6d6f-32ff-ab03-4eb43640c758",
					},
					Button{
						Label:  "0.005 XIN",
						Color:  "#3333FF",
						Action: "",
						Amount: "0.005",
						Asset:  "c94ac88f-4671-3976-b60a-09064f1811e8",
					},
				}
				for i := 0; i < 2; i += 1 {
					traceID, _ := uuid.NewV4()
					buttonInfos[i].Action = "mixin://pay?recipient=" + recipient + "&asset=" + buttonInfos[i].Asset + "&amount=" + buttonInfos[i].Amount + "&trace=" + traceID.String() + "&memo=A%20cup%20of%20milk"
				}
				biJson, _ := JSONMarshal(buttonInfos)
				log.Printf("%s\n", string(biJson))
				if err := bot.SendGeneralMessage(ctx, mc, msg, "APP_BUTTON_GROUP", string(biJson)); err != nil {
					log.Printf("err: %v\n", err)
					return bot.BlazeServerError(ctx, err)
				}
			} else {
				if err := bot.SendPlainText(ctx, mc, msg, InstructionHint); err != nil {
					return bot.BlazeServerError(ctx, err)
				}
			}
		} else {
			if err := bot.SendPlainText(ctx, mc, msg, InstructionHint); err != nil {
				return bot.BlazeServerError(ctx, err)
			}
		}
	}
	return nil
}
