package webscraper

import (
	"apex_tournament_noti/internal/linebot/webhook"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

type KeyStanding struct {
	key      string
	standing int
}

type Notification interface {
	createMessage() string
	pushMessage(msg string)
}

type LiveUpdateNotification struct {
	Title     string
	CurrRound int
	FirstExec bool
	Channel   chan<- string
	OldMapPtr *map[int]map[string]MatchData
	NewMapPtr *map[int]map[string]MatchData
}

type UserUpdateNotification struct {
	Title           string
	Token           string
	CurrRound       int
	Channel         chan<- webhook.WebhookPayload
	MatchDataMapPtr *map[int]map[string]MatchData
}

func PushNotificationMessage(n Notification) {
	fmt.Println("Sending notification...")
	msg := n.createMessage()
	fmt.Println(msg)
	n.pushMessage(msg)
}

func (s *UserUpdateNotification) createMessage() string {
	roundMap := (*s.MatchDataMapPtr)[s.CurrRound]
	var msg strings.Builder

	ksArr := make([]KeyStanding, 0, len(roundMap))
	for k, _ := range roundMap {
		ks := KeyStanding{key: k, standing: roundMap[k].standing}
		ksArr = append(ksArr, ks)
	}

	// Sort kps by the descending order of total points
	sort.Slice(ksArr, func(i, j int) bool {
		return ksArr[i].standing < ksArr[j].standing
	})

	// Iterate through the map by the descending order of total points
	for i, kp := range ksArr {
		k := kp.key
		v := roundMap[k] // MatchData struct

		msg.WriteString(fmt.Sprintf(" %d. %s: %d |", i+1, k, v.totalPoints))
	}

	return msg.String()
}

func (s *UserUpdateNotification) pushMessage(msg string) {
	msgPayload := webhook.WebhookPayload{
		Token: s.Token,
		Text:  msg,
	}
	s.Channel <- msgPayload
}

func (s *LiveUpdateNotification) createMessage() string {
	// Check if the team data has been updated
	updated := !reflect.DeepEqual(*s.OldMapPtr, *s.NewMapPtr)
	oldRoundMap := (*s.OldMapPtr)[s.CurrRound]
	roundMap := (*s.NewMapPtr)[s.CurrRound]
	fmt.Println("roundMap: ", roundMap)

	// Get current game number
	// Find the minimum number of the non-filled entries across all teams
	priviousMinEmptyGame := 100.0 // Set initial value to a large number
	minEmptyGame := 100.0         // Set initial value to a large number
	totalGameNum := 0
	priviousGame := 0
	currentGame := 0
	allFilled := true
	for _, v := range oldRoundMap {
		totalGameNum = len(v.gamePlacements)
		for i := 0; i < len(v.gamePlacements); i++ {
			if v.gamePlacements[i] == -1 {
				priviousMinEmptyGame = math.Min(priviousMinEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.gameKills); i++ {
			if v.gameKills[i] == -1 {
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
	for _, v := range roundMap {
		totalGameNum = len(v.gamePlacements)
		for i := 0; i < len(v.gamePlacements); i++ {
			if v.gamePlacements[i] == -1 {
				minEmptyGame = math.Min(minEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.gameKills); i++ {
			if v.gameKills[i] == -1 {
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

	if updated && !s.FirstExec {
		fmt.Println("Updated!")
		msg.WriteString(fmt.Sprintf("%s: ", s.Title))

		gameFinished := true
		for _, v := range roundMap {
			if v.gamePlacements[currentGame-1] == -1 {
				gameFinished = false
				break
			}
		}

		ksArr := make([]KeyStanding, 0, len(roundMap))
		for k, _ := range roundMap {
			ks := KeyStanding{key: k, standing: roundMap[k].standing}
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
			v := roundMap[k] // MatchData struct

			if gameFinished {
				msg.WriteString(fmt.Sprintf(" %d. %s: %d |", i+1, k, v.totalPoints))
			} else {
				// If the gamePlacements are different between old and new MatchData, the team is eleiminated
				if v.gamePlacements[currentGame-1] != (*s.OldMapPtr)[s.CurrRound][k].gamePlacements[currentGame-1] {
					msg.WriteString(fmt.Sprintf(" %s's placement: %d |", k, v.gamePlacements[currentGame-1]))
				}
			}
		}

	} else {
		fmt.Println("Not updated!")
		// msg.WriteString("Not updated!")
	}

	return msg.String()
}

func (s *LiveUpdateNotification) pushMessage(msg string) {
	if len(msg) > 0 {
		s.Channel <- msg
	}
}

// TODO: update check should be implemented in another function.
/* func SendNotificationMessage(title string, round int, firstExec bool, noti chan<- string, oldMapPtr *map[int]map[string]MatchData, newMapPtr *map[int]map[string]MatchData) {
	fmt.Printf("--round--: %d\n", round)
	// Check if the team data has been updated
	updated := !reflect.DeepEqual(*oldMapPtr, *newMapPtr)
	currentRound := round
	oldRoundMap := (*oldMapPtr)[currentRound]
	roundMap := (*newMapPtr)[currentRound]
	fmt.Println("roundMap: ", roundMap)

	// Get current game number
	// Find the minimum number of the non-filled entries across all teams
	priviousMinEmptyGame := 100.0 // Set initial value to a large number
	minEmptyGame := 100.0         // Set initial value to a large number
	totalGameNum := 0
	priviousGame := 0
	currentGame := 0
	allFilled := true
	for _, v := range oldRoundMap {
		totalGameNum = len(v.gamePlacements)
		for i := 0; i < len(v.gamePlacements); i++ {
			if v.gamePlacements[i] == -1 {
				priviousMinEmptyGame = math.Min(priviousMinEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.gameKills); i++ {
			if v.gameKills[i] == -1 {
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
	for _, v := range roundMap {
		totalGameNum = len(v.gamePlacements)
		for i := 0; i < len(v.gamePlacements); i++ {
			if v.gamePlacements[i] == -1 {
				minEmptyGame = math.Min(minEmptyGame, float64(i+1))
				allFilled = false
			}
		}
		for i := 0; i < len(v.gameKills); i++ {
			if v.gameKills[i] == -1 {
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

	if updated && !firstExec {
		fmt.Println("Updated!")
		msg.WriteString(fmt.Sprintf("%s: ", title))

		gameFinished := true
		for _, v := range roundMap {
			if v.gamePlacements[currentGame-1] == -1 {
				gameFinished = false
				break
			}
		}

		ksArr := make([]KeyStanding, 0, len(roundMap))
		for k, _ := range roundMap {
			ks := KeyStanding{key: k, standing: roundMap[k].standing}
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
			v := roundMap[k] // MatchData struct

			if gameFinished {
				msg.WriteString(fmt.Sprintf(" %d. %s: %d |", i+1, k, v.totalPoints))
			} else {
				// If the gamePlacements are different between old and new MatchData, the team is eleiminated
				if v.gamePlacements[currentGame-1] != (*oldMapPtr)[currentRound][k].gamePlacements[currentGame-1] {
					msg.WriteString(fmt.Sprintf(" %s's placement: %d |", k, v.gamePlacements[currentGame-1]))
				}
			}
		}

		noti <- msg.String()

	} else {
		fmt.Println("Not updated!")
	}

}

func GetNotification(noti <-chan string) string {
	return <-noti
}
*/
