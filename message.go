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

const InstructionHint = "发送 start 订阅\n发送 stop 取消订阅\n发送 random 随便看看\n发送 donate 请我喝牛奶"

type ResponseMessage struct {
}

type Button struct {
	Label  string `json:"label"`
	Color  string `json:"color"`
	Action string `json:"action"`
	Asset  string `json:-`
	Amount string `json:-`
}

type transferMessageData struct {
	Amount        string `json: "amount"`
	AssetID       string `json: "asset_id"`
	CounterUserID string `json: "counter_user_id"`
	CreatedAt     string `json: "created_at"`
	Memo          string `json: "memo"`
	OpponentID    string `json: "opponent_id"`
	SnapshotID    string `json: "snapshot_id"`
	TraceID       string `json: "trace_id"`
	Type          string `json: "type"`
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func crackAccountSnapshotMessage(ctx context.Context, mc *bot.MessageContext, msg bot.MessageView) error {
	/* for transfer message
	{
	"conversation_id": "65e03fac-3cf9-3768-b35f-00ca70574a21",
	"user_id": "dfa655ef-55db-4e18-bdd7-29a7c576a223",
	"message_id": "ceb6a852-6975-49fa-b353-e2189d795df8",
	"category": "SYSTEM_ACCOUNT_SNAPSHOT",
	"data": "eyJhbW91bnQiOiI1IiwiYXNzZXRfaWQiOiIzZWRiNzM0Yy02ZDZmLTMyZmYtYWIwMy00ZWI0MzY0MGM3NTgiLCJjb3VudGVyX3VzZXJfaWQiOiJkZmE2NTVlZi01NWRiLTRlMTgtYmRkNy0yOWE3YzU3NmEyMjMiLCJjcmVhdGVkX2F0IjoiMjAxOC0wOC0yMFQxMzo1NTozNi4yNDExODA4MTNaIiwibWVtbyI6IkEgY3VwIG9mIG1pbGsiLCJvcHBvbmVudF9pZCI6ImRmYTY1NWVmLTU1ZGItNGUxOC1iZGQ3LTI5YTdjNTc2YTIyMyIsInNuYXBzaG90X2lkIjoiOTY0MjYxNzItNjEzOS00MmE4LTgzOTQtNTczMmYwYzYyNzNiIiwidHJhY2VfaWQiOiI0YWMxYWNjMy0wMDJjLTRjZDgtNTllMS1hMmU2NjRmYjYzZDkiLCJ0eXBlIjoidHJhbnNmZXIifQ==",
	data: {
		"amount": "5",
		"asset_id": "3edb734c-6d6f-32ff-ab03-4eb43640c758",
		"counter_user_id": "dfa655ef-55db-4e18-bdd7-29a7c576a223",
		"created_at": "2018-08-20T13:55:36.241180813Z",
		"memo": "A cup of milk",
		"opponent_id": "dfa655ef-55db-4e18-bdd7-29a7c576a223",
		"snapshot_id": "96426172-6139-42a8-8394-5732f0c6273b",
		"trace_id": "4ac1acc3-002c-4cd8-59e1-a2e664fb63d9",
		"type": "transfer"
	}
	"status": "SENT",
	"source": "CREATE_MESSAGE",
	"created_at": "2018-08-20T13:55:36.241282355Z",
	"updated_at": "2018-08-20T13:55:36.241282355Z"
	}
	*/
	var err error
	var rawData []byte
	data := &transferMessageData{}
	rawData, err = base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	log.Printf("Transfer: %s, %d, %s\n", data.SnapshotID, data.Amount, data.Memo)
	return nil
}

func (r ResponseMessage) OnMessage(ctx context.Context, mc *bot.MessageContext, msg bot.MessageView) error {
	res2B, _ := json.Marshal(msg)
	log.Printf("%s   %s\n", string(res2B), msg.Category)
	if msg.Category == "SYSTEM_ACCOUNT_SNAPSHOT" {
		crackAccountSnapshotMessage(ctx, mc, msg)
	} else if msg.Category != bot.MessageCategorySystemAccountSnapshot && msg.Category != bot.MessageCategorySystemConversation && msg.ConversationId == bot.UniqueConversationId(config.Global.MixinClientId, msg.UserId) {
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
					filepath, _ := services.DownloadMedia(ctx, url, "twitter-heart-disturbed")
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
				config := config.Global.Spiders[0]
				service.Run(ctx, config)
			} else if "debug" == string(data) {
				var user *models.Subscriber
				user, err = models.FindSubscriberById(ctx, msg.UserId)
				if err == nil {
					if err := bot.SendPlainText(ctx, mc, msg, "你正在订阅本机器人。ID："+user.UserId); err != nil {
						return bot.BlazeServerError(ctx, err)
					}
				} else {
					bot.SendPlainText(ctx, mc, msg, "你没有在订阅本机器人。ID："+msg.UserId)
				}
			} else if "donate" == string(data) {
				recipient := "56e28252-1aef-48a6-a0c7-2213652570d7" // "dfa655ef-55db-4e18-bdd7-29a7c576a223" // my account
				if err != nil {
					log.Println("error:", err)
					return bot.BlazeServerError(ctx, err)
				}
				buttonInfos := []Button{
					Button{
						Label:  "0.1 PRS",
						Color:  "#5555CC",
						Action: "",
						Amount: "0.1",
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
