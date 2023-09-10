package notification

import (
	"apex_tournament_noti/internal/webscraper"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

type LiveUpdateNotification struct {
	Title              string
	FirstExec          bool
	Channel            chan<- string
	PrevStageMatchData webscraper.StageMatchData
	CurrStageMatchData webscraper.StageMatchData
}

func (l *LiveUpdateNotification) createMessage() string {
	prevMatchDataSlice := (*l).PrevStageMatchData.Data
	currMatchDataSlice := (*l).CurrStageMatchData.Data

	// Check if the team data has been updated
	//fmt.Printf("PrevStageMatchData: %v\n", len(*prevMatchDataSlice))
	//fmt.Printf("CurrStageMatchData: %v\n", len(*currMatchDataSlice))
	updated := !reflect.DeepEqual(*prevMatchDataSlice, *currMatchDataSlice)

	// First execution
	if len(*prevMatchDataSlice) == 0 {
		return ""
	}

	prevMatchDataMap := (*prevMatchDataSlice)[(*l).PrevStageMatchData.CurrRound]
	currMatchDataMap := (*currMatchDataSlice)[(*l).CurrStageMatchData.CurrRound]
	//fmt.Println("roundMap: ", roundMap)

	// Get current game number
	// Find the minimum number of the non-filled entries across all teams
	priviousMinEmptyGame := 100.0 // Set initial value to a large number
	minEmptyGame := 100.0         // Set initial value to a large number
	totalGameNum := 0
	priviousGame := 0
	currentGame := 0
	allFilled := true
	for _, v := range prevMatchDataMap {
		totalGameNum = len(v.GamePlacements)
		for i := 0; i < len(v.GamePlacements); i++ {
			if v.GamePlacements[i] == -1 {
				priviousMinEmptyGame = math.Min(priviousMinEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.GameKills); i++ {
			if v.GameKills[i] == -1 {
				priviousMinEmptyGame = math.Min(priviousMinEmptyGame, float64(i+1))
				allFilled = false
			}
		}
	}
	if allFilled {
		priviousGame = totalGameNum
	} else {
		priviousGame = int(priviousMinEmptyGame)
	}

	allFilled = true
	for _, v := range currMatchDataMap {
		totalGameNum = len(v.GamePlacements)
		for i := 0; i < len(v.GamePlacements); i++ {
			if v.GamePlacements[i] == -1 {
				minEmptyGame = math.Min(minEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.GameKills); i++ {
			if v.GameKills[i] == -1 {
				minEmptyGame = math.Min(minEmptyGame, float64(i+1))
				allFilled = false
			}
		}
	}
	if allFilled {
		currentGame = totalGameNum
	} else {
		currentGame = int(minEmptyGame)
	}
	// The last update of the previous game
	if currentGame == (priviousGame+1) && currentGame > 1 {
		currentGame = priviousGame
	}
	fmt.Println("Current game: ", currentGame)

	var msg strings.Builder

	if updated && !l.FirstExec {
		fmt.Println("Updated!")
		msg.WriteString(fmt.Sprintf("%s: ", l.Title))

		gameFinished := true
		for _, v := range currMatchDataMap {
			if v.GamePlacements[currentGame-1] == -1 {
				gameFinished = false
				break
			}
		}

		ksArr := make([]KeyStanding, 0, len(currMatchDataMap))
		for k, _ := range currMatchDataMap {
			ks := KeyStanding{key: k, standing: currMatchDataMap[k].Standing}
			ksArr = append(ksArr, ks)
		}

		// Sort kps by the descending order of total points
		sort.Slice(ksArr, func(i, j int) bool {
			return ksArr[i].standing < ksArr[j].standing
		})

		//if gameFinished {
		//	msg.WriteString("Updated standings:\n")
		//}

		// Iterate through the map by the descending order of total points
		for i, kp := range ksArr {
			k := kp.key
			v := currMatchDataMap[k] // MatchData struct

			if gameFinished {
				msg.WriteString(fmt.Sprintf(" %d. %s: %d |", i+1, k, v.TotalPoints))
			} else {
				// If the GamePlacements are different between old and new MatchData, the team is eleiminated
				if v.GamePlacements[currentGame-1] != (*prevMatchDataSlice)[(*l).PrevStageMatchData.CurrRound][k].GamePlacements[currentGame-1] {
					msg.WriteString(fmt.Sprintf(" %s's placement: %d |", k, v.GamePlacements[currentGame-1]))
				}
			}
		}

	} else {
		fmt.Println("Not updated!")
		// msg.WriteString("Not updated!")
	}

	return msg.String()
}

func (l *LiveUpdateNotification) pushMessage(msg string) {
	if len(msg) > 0 {
		l.Channel <- msg
	}
}
