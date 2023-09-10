package notification

import (
	"apex_tournament_noti/internal/linebot/webhook"
	"apex_tournament_noti/internal/webscraper"
	"fmt"
	"strings"
)

type UserCommandGroupStanding struct {
	Title         string
	Token         string
	Channel       chan<- webhook.WebhookPayload
	GroupStanding *webscraper.GroupStageStandings
}

func (u *UserCommandGroupStanding) createMessage() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf(" %s: ", u.Title))

	for _, ts := range u.GroupStanding.Standings {
		msg.WriteString(fmt.Sprintf(" %d. %s: %d |", ts.Standing, ts.TeamName, ts.TotalPoints))
	}

	return msg.String()
}

func (u *UserCommandGroupStanding) pushMessage(msg string) {
	msgPayload := webhook.WebhookPayload{
		Token: u.Token,
		Text:  msg,
	}
	u.Channel <- msgPayload
}
