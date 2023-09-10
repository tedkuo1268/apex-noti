package notification

import (
	"apex_tournament_noti/internal/linebot/webhook"
	"apex_tournament_noti/internal/webscraper"
	"fmt"
	"sort"
	"strings"
)

type UserCommandNow struct {
	Title          string
	Token          string
	Channel        chan<- webhook.WebhookPayload
	StageMatchData *webscraper.StageMatchData
}

func (u *UserCommandNow) createMessage() string {
	MatchDataSlice := (*u).StageMatchData.Data
	MatchDataMap := (*MatchDataSlice)[(*u).StageMatchData.CurrRound]
	var msg strings.Builder

	ksArr := make([]KeyStanding, 0, len(MatchDataMap))
	for k, _ := range MatchDataMap {
		ks := KeyStanding{key: k, standing: MatchDataMap[k].Standing}
		ksArr = append(ksArr, ks)
	}

	// Sort kps by the descending order of total points
	sort.Slice(ksArr, func(i, j int) bool {
		return ksArr[i].standing < ksArr[j].standing
	})

	// Iterate through the map by the descending order of total points
	for i, kp := range ksArr {
		k := kp.key
		v := MatchDataMap[k] // MatchData struct

		msg.WriteString(fmt.Sprintf(" %d. %s: %d |", i+1, k, v.TotalPoints))
	}

	return msg.String()
}

func (s *UserCommandNow) pushMessage(msg string) {
	msgPayload := webhook.WebhookPayload{
		Token: s.Token,
		Text:  msg,
	}
	s.Channel <- msgPayload
}
