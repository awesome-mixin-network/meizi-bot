package services

import (
	"bytes"
	"context"
	"encoding/base64"
	bot "github.com/lyricat/bot-api-go-client"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"time"

	simplejson "github.com/bitly/go-simplejson"

	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fox-one/foxone-mixin-bot/clients"
	"github.com/fox-one/foxone-mixin-bot/config"
	"github.com/fox-one/foxone-mixin-bot/models"
)

type ServiceStats struct {
	exists map[string]interface{}
}

func (self *ServiceStats) filterDuplicate(ctx context.Context, items []*clients.Item, serviceName string) []*clients.Item {
	// 去掉重复的，只发最新的，用 新的+重复 的替换现有的
	result := []*clients.Item{}
	for _, item := range items {
		if !models.IsExistedItemByItemId(ctx, item.Id, serviceName) {
			result = append(result, item)
		}
	}
	return result
}

func GetRandomPicture(ctx context.Context) string {
	item, _ := models.GetRandomItem(ctx)
	var pickedUrl string
	if len(item.Urls) != 0 {
		pickedUrl = item.Urls[rand.Intn(len(item.Urls))]
	}
	return pickedUrl
}

func CreateAttachment(ctx context.Context, filepath string) (string, error) {
	var err error
	var attachmentId, uploadUrl string
	// request upload info
	attachmentId, uploadUrl, _, err = requestUploadAttachmentInfo(ctx)
	if err != nil {
		log.Println("requestUploadAttachmentInfo failed", err)
		return "", err
	}
	// prepare http client and post
	return PostAttachementFile(attachmentId, uploadUrl, filepath)
}

func PostAttachementFile(attachmentId, uploadUrl, filepath string) (string, error) {
	chunkSize := 1024
	//open file and retrieve info
	file, _ := os.Open(filepath)
	fi, err := file.Stat()
	if err != nil {
		log.Println("file.Stat failed", filepath, err)
		return "", err
	}
	totalSize := fi.Size()
	defer file.Close()
	// get meta info
	attachmentWidth, attachmentHeight := getImageDimension(filepath)
	attachmentContentType := getImageContentType(filepath)

	//buffer for storing multipart data
	byteBuf := &bytes.Buffer{}

	//part: parameters
	mpWriter := multipart.NewWriter(byteBuf)

	//part: file
	mpWriter.CreateFormFile("file", fi.Name())
	nmulti := byteBuf.Len()
	multi := make([]byte, nmulti)
	_, _ = byteBuf.Read(multi)
	mpWriter.Close()

	//calculate content length

	//use pipe to pass request
	rd, wr := io.Pipe()
	defer rd.Close()

	// just write file, no boundary needed
	go func() {
		defer wr.Close()
		buf := make([]byte, chunkSize)
		for {
			n, err := file.Read(buf)
			if err != nil {
				break
			}
			_, _ = wr.Write(buf[:n])
		}
	}()

	// send request
	client := &http.Client{}
	req, err := http.NewRequest("PUT", uploadUrl, rd)
	if err != nil {
		log.Println("NewRequest Failed", err)
		return "", err
	}
	req.Header.Add("x-amz-acl", "public-read")
	req.Header.Add("Content-Type", "application/octet-stream")
	req.ContentLength = totalSize

	// do request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Put Error", err, filepath)
		if resp != nil && resp.Body != nil {
			message, _ := ioutil.ReadAll(resp.Body)
			log.Println("Put Error Message", message)
		}
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	// return meta info
	js := simplejson.New()
	js.Set("attachment_id", attachmentId)
	js.Set("mine_type", attachmentContentType)
	js.Set("mime_type", attachmentContentType)
	js.Set("width", attachmentWidth)
	js.Set("height", attachmentHeight)
	js.Set("size", totalSize)
	js.Set("thumbnail", "/9j/4AAQSkZJRgABAQAASABIAAD/4QDIRXhpZgAATU0AKgAAAAgABwESAAMAAAABAAEAAAEaAAUAAAABAAAAYgEbAAUAAAABAAAAagEoAAMAAAABAAIAAAExAAIAAAAPAAAAcgEyAAIAAAAUAAAAgodpAAQAAAABAAAAlgAAAAAAAABIAAAAAQAAAEgAAAABUGl4ZWxtYXRvciAzLjcAADIwMTg6MDU6MzEgMTk6MDU6NjMAAAOgAQADAAAAAQABAACgAgAEAAAAAQAAAECgAwAEAAAAAQAAAEAAAAAA/+EJ9Gh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8APD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iWE1QIENvcmUgNS40LjAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1wOk1vZGlmeURhdGU9IjIwMTgtMDUtMzFUMTk6MDU6NjMiIHhtcDpDcmVhdG9yVG9vbD0iUGl4ZWxtYXRvciAzLjciPiA8ZGM6c3ViamVjdD4gPHJkZjpCYWcvPiA8L2RjOnN1YmplY3Q+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIDw/eHBhY2tldCBlbmQ9InciPz4A/+0AOFBob3Rvc2hvcCAzLjAAOEJJTQQEAAAAAAAAOEJJTQQlAAAAAAAQ1B2M2Y8AsgTpgAmY7PhCfv/AABEIAEAAQAMBIgACEQEDEQH/xAAfAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBwgJCgv/xAC1EAACAQMDAgQDBQUEBAAAAX0BAgMABBEFEiExQQYTUWEHInEUMoGRoQgjQrHBFVLR8CQzYnKCCQoWFxgZGiUmJygpKjQ1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4eLj5OXm5+jp6vHy8/T19vf4+fr/xAAfAQADAQEBAQEBAQEBAAAAAAAAAQIDBAUGBwgJCgv/xAC1EQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEyIygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2dri4+Tl5ufo6ery8/T19vf4+fr/2wBDAAEBAQEBAQIBAQIDAgICAwQDAwMDBAYEBAQEBAYHBgYGBgYGBwcHBwcHBwcICAgICAgJCQkJCQsLCwsLCwsLCwv/2wBDAQICAgMDAwUDAwULCAYICwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwsLCwv/3QAEAAT/2gAMAwEAAhEDEQA/AP6SLNQ2K6+y04yr0rmtJhZ3Ar2jw9pfmYyOtfkmEwntEfpOJxHJqcoNAdxwppT4bk7Ka+gLDw0sijitxPB6tztrteTNnA81sfMg8OPjlTQ3h+RR0NfTzeDwB92su78Loin5aqOTNdBf2rc+XbvSmjHTpXKXdsysRivovWdAABwtecXugPu4FY4jLnFbHRSxikf/0P6i9G0RlYEivaNAtUh25FWYPDogGdtWdjWxwBXx+X4Dktc+srYn2rPTtHaIgCvQrO1jkQYrxLSL1w4FexaPcF4wa99UEkeXiINampPZxovIxXFaoIkyK7PUJiIya8m1u9dXIFV7FGdGLkzGvrWKdiAKxW8OpKcha0racySAV2+nwK4yRXLWwkZHTKo6Z//R/tQurNFGK5O7tAzcV0Wpaii5zXLnU42frXkqpGLO2lWkmaWl2JWQEV6xpMWyMV5xpl7ESDmu/sr1NoxXRGqmXWrtmzfJvjOK8x1jTzIxNehSXiMtYN0Y3OKvmIpV+U4C209kkyRXa2CGMCqwRAauxOB0rNzTHVrcx//S/rv1rXgM/NXFprzGXg8Vx+p6y0meaxra6aSUDNfnNTMm5aM9mlhtD6C0TV2cjmvUtPv3ZRXgvhoM+MV7ro9sWjFe9gKzmtTkxNOxum5cDOapyXR5rVNk2Kyrq3KivTnKyuccVqUnvdvNCakAeuKwb12jzXMz6myHrXk1cdys6fZ3R//Z")

	imageData, _ := js.MarshalJSON()
	fmt.Println("imageData", attachmentId, attachmentContentType, attachmentWidth, attachmentHeight, totalSize)
	encoded := base64.StdEncoding.EncodeToString(imageData)
	return encoded, nil
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

func getImageContentType(imagePath string) string {
	filename := strings.ToLower(path.Base(imagePath))
	if strings.HasSuffix(filename, ".png") {
		return "image/png"
	} else if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(filename, ".gif") {
		return "image/gif"
	}
	return "image/jpeg"
}

func DownloadImage(ctx context.Context, url, serviceName string) (string, error) {
	filename := path.Base(url)
	currentwd, _ := os.Getwd()
	filepath := path.Join(currentwd, path.Dir("./storage/"+serviceName+"/"), filename)
	if _, err := os.Stat(filepath); err == nil {
		// already exists
		return filepath, nil
	}
	// download it.
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()
	//open file for writing
	file, err := os.Create(path.Join(currentwd, path.Dir("./storage/"+serviceName+"/"), filename))
	if err != nil {
		log.Fatal(err)
		return filepath, err
	}
	// copy the huge content to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return filepath, err
		log.Fatal(err)
	}
	file.Close()
	return filepath, nil
}

func requestUploadAttachmentInfo(ctx context.Context) (attachmentId, uploadUrl, viewUrl string, err error) {
	accessToken, err := bot.SignAuthenticationToken(config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey, "POST", "/attachments", "")
	if err != nil {
		log.Println("SignAuthenticationToken Failed", err)
	}
	body, err := bot.Request(ctx, "POST", "/attachments", []byte(""), accessToken)
	if err != nil {
		log.Println("Request Failed", err)
		return
	}
	js, err := simplejson.NewJson(body)
	if err != nil {
		log.Println("Failed to create Json object", err)
		return
	}
	attachmentId, err = js.GetPath("data", "attachment_id").String()
	uploadUrl, err = js.GetPath("data", "upload_url").String()
	viewUrl, err = js.GetPath("data", "view_url").String()
	// log.Println("attachmentId", attachmentId)
	// log.Println("uploadUrl", uploadUrl)
	// log.Println("viewUrl", viewUrl)
	return
}

func SendPicturesToChannel(ctx context.Context, items []*clients.Item, serviceName string) {
	log.Printf("send %d pictures (%s) to channel...", len(items), serviceName)
	if len(items) != 0 {
		for _, item := range items {
			for _, url := range item.Urls {
				filepath, err1 := DownloadImage(ctx, url, serviceName)
				imageData, err2 := CreateAttachment(ctx, filepath)
				if err1 == nil && err2 == nil {
					sendMessageToSubscribers(ctx, "PLAIN_IMAGE", imageData)
				} else {
					log.Println("Error", err1, err2)
				}
			}
		}
	}
}

func sendMessageToSubscribers(ctx context.Context, messageType, message string) {
	subscribers, _ := models.FindSubscribers(ctx)
	for _, subscriber := range subscribers {
		conversationId := bot.UniqueConversationId(config.MixinClientId, subscriber.UserId)
		bot.PostMessage(ctx, conversationId, subscriber.UserId, bot.NewV4().String(), messageType, message, config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey)
		time.Sleep(300 * time.Millisecond)
	}
}
