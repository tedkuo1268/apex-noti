package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"apex_tournament_noti/internal/linebot/routes"
	"apex_tournament_noti/internal/linebot/webhook"
	"apex_tournament_noti/internal/notification"
	"apex_tournament_noti/internal/webscraper"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

var (
	LineBotChannelSecret = os.Getenv("LINE_BOT_CHANNEL_SECRET")
	LineBotChannelToken  = os.Getenv("LINE_BOT_CHANNEL_TOKEN")
)

type Payload struct {
	IsSuccess bool `json:"isSuccess"`
}

func main() {
	// Create a channel for notification
	broadcastChannel := make(chan string)
	webhookChannel := make(chan webhook.WebhookPayload, 100)
	responseChannel := make(chan webhook.WebhookPayload, 100)

	bot, err := linebot.New(
		LineBotChannelSecret,
		LineBotChannelToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		firstExec := true
		url := "https://liquipedia.net/apexlegends/Apex_Legends_Global_Series/2023/Championship/Group_Stage"

		currStageMatchData := webscraper.StageMatchData{Data: &[]map[string]webscraper.MatchData{}, CurrRound: 0}
		fmt.Printf("currStageMatchData: %v\n", currStageMatchData)

		for {
			title := "Championship Group Stage"

			prevStageMatchData := webscraper.StageMatchData{Data: &[]map[string]webscraper.MatchData{}, CurrRound: 0}
			if len(*currStageMatchData.Data) != 0 {
				prevStageMatchData.CurrRound = currStageMatchData.CurrRound
				for _, v := range *currStageMatchData.Data {
					prevStageMatchData.AddMatchDataMap(v)
				}
			}

			webscraper.Scrape(&currStageMatchData, url)
			liveNoti := notification.LiveUpdateNotification{Title: title, FirstExec: firstExec, Channel: broadcastChannel, PrevStageMatchData: prevStageMatchData, CurrStageMatchData: currStageMatchData}
			notification.PushNotificationMessage(&liveNoti)

			time.Sleep(10 * time.Second)

			if firstExec {
				firstExec = false
			}
		}
	}()

	go func() {
		for {
			msg := <-broadcastChannel
			fmt.Println("msg:")
			fmt.Println(msg)

			message := linebot.NewTextMessage(msg)
			if _, err := bot.BroadcastMessage(message).Do(); err != nil {
				log.Printf("Error: %s", err)
			}
		}
	}()

	go func() {
		for {
			msgPayload := <-webhookChannel
			fmt.Printf("Webhook msgPayload Token: %v\n", msgPayload.Token)
			fmt.Printf("Webhook msgPayload Text: %v\n", msgPayload.Text)

			command := strings.ToLower(strings.Trim(msgPayload.Text, " "))
			fmt.Printf("Command: %v\n", command)
			switch command {
			case "/help":
				go func() {
					text := "Commands:\n" +
						"!now: Show the standings and scores of current series\n" +
						"!gs: Show the group stage team scores and standings\n"
					responsePayload := webhook.WebhookPayload{
						Token: msgPayload.Token,
						Text:  text,
					}
					responseChannel <- responsePayload
				}()
			case "!now":
				go func() {
					url := "https://liquipedia.net/apexlegends/Apex_Legends_Global_Series/2023/Championship/Group_Stage"
					title := "Championship Group Stage"
					currStageMatchData := webscraper.StageMatchData{}
					webscraper.Scrape(&currStageMatchData, url)
					userNoti := notification.UserCommandNow{Title: title, Token: msgPayload.Token, Channel: responseChannel, StageMatchData: &currStageMatchData}
					notification.PushNotificationMessage(&userNoti)
				}()
			case "!gs":
				go func() {
					url := "https://liquipedia.net/apexlegends/Apex_Legends_Global_Series/2023/Championship/Group_Stage"
					title := "Championship Group Stage"
					groupStageStandings := webscraper.GroupStageStandings{}
					webscraper.Scrape(&groupStageStandings, url)
					userNoti := notification.UserCommandGroupStanding{Title: title, Token: msgPayload.Token, Channel: responseChannel, GroupStanding: &groupStageStandings}
					notification.PushNotificationMessage(&userNoti)
				}()
			}
		}
	}()

	go func() {
		for {
			msgPayload := <-responseChannel
			fmt.Printf("Response msgPayload Token: %v\n", msgPayload.Token)
			fmt.Printf("Response msgPayload Text: %v\n", msgPayload.Text)

			if _, err = bot.ReplyMessage(msgPayload.Token, linebot.NewTextMessage(msgPayload.Text)).Do(); err != nil {
				log.Print(err)
			}
		}
	}()

	r := gin.Default()
	routes.SetupRoutes(r, webhookChannel)
	//r.Use(signatureMiddleware)
	r.Run(":8080") // listen and serve on port 8080
}
