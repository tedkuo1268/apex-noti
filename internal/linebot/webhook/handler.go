package webhook

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

var (
	LineBotChannelSecret = os.Getenv("LINE_BOT_CHANNEL_SECRET")
	LineBotChannelToken  = os.Getenv("LINE_BOT_CHANNEL_TOKEN")
)

type WebhookPayload struct {
	Token string
	Text  string
}

type JSONResponse struct {
	Message string `json:"message"`
}

func HandleWebhook(ch chan WebhookPayload) gin.HandlerFunc {
	return func(c *gin.Context) {
		bot, err := linebot.New(
			LineBotChannelSecret,
			LineBotChannelToken,
		)
		if err != nil {
			log.Fatal(err)
		}

		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Print(err)
				c.AbortWithStatus(http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid signature",
				})
			} else {
				log.Print(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal server error",
				})
			}
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					fmt.Printf("UserId: %s\n", event.Source.UserID)

					msgPayload := WebhookPayload{
						Token: event.ReplyToken,
						Text:  message.Text,
					}

					ch <- msgPayload
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
		})
	}
}

func TestHandleWebhook(ch chan WebhookPayload) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p WebhookPayload
		if err := c.ShouldBindJSON(&p); err != nil {
			log.Printf("Error: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ch <- p

		c.JSON(http.StatusOK, gin.H{
			"isSuccess": "true",
		})
	}

}
