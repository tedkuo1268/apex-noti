package webscraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// Team data struct
type MatchData struct {
	TeamName            string
	Standing            int
	TotalPoints         int
	Round               int
	GamePlacements      []int
	GamePlacementPoints []int
	GameKills           []int
	GameKillPoints      []int
}

type StageMatchData struct {
	Data      *[]map[string]MatchData
	CurrRound int
}

// MatchData constructor
func NewMatchData(teamName string) *MatchData {
	td := MatchData{TeamName: teamName}
	return &td
}

func (s *StageMatchData) AddMatchDataMap(td map[string]MatchData) {
	*s.Data = append(*s.Data, td)
}

func (s *StageMatchData) GetData(url string) {
	c := colly.NewCollector()
	tableNum := 0
	currRound := 0

	// Set s.Data slice to nil
	*s.Data = nil

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	/* c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited: ", r.Request.URL)
	}) */

	c.OnHTML("table.table-battleroyale-results", func(e *colly.HTMLElement) {
		//fmt.Println("Table No: ", tableNum)
		//fmt.Println(e.Text)

		// Create a map of the current table to store each team's data
		matchDataMap := make(map[string]MatchData)

		rowCounter := 0
		e.ForEach("tr", func(_ int, eTr *colly.HTMLElement) {
			//fmt.Printf("Row%d: %s\n", rowCounter, eTr.Text)
			//fmt.Println("Row No: ", rowCounter)
			colCounter := 0
			md := *NewMatchData("")
			md.Round = tableNum
			eTr.ForEach("td > span", func(_ int, eTrTdSpan *colly.HTMLElement) {
				if eTrTdSpan.Attr("class") == "team-template-team-short" {
					//fmt.Println("Team: ", eTrTdSpan.Text)
					md.TeamName = strings.Trim(eTrTdSpan.Text, " ")
					md.Standing = rowCounter - 2 // -2 for the headers
				} else {
					// fmt.Println("colCounter: ", colCounter)
					if colCounter%4 == 0 {
						//fmt.Println("Placement: ", eTrTdSpan.Text)
						placement := 0
						if eTrTdSpan.Text == "-" {
							placement = -1
						} else {
							placement, _ = strconv.Atoi(eTrTdSpan.Text)
							currRound = tableNum
						}
						md.GamePlacements = append(md.GamePlacements, placement)
					} else if colCounter%4 == 1 {
						//fmt.Println("Placement Point: ", el.Text)
						placementPoint, _ := strconv.Atoi(eTrTdSpan.Text)
						md.GamePlacementPoints = append(md.GamePlacementPoints, placementPoint)
					} else if colCounter%4 == 2 {
						//fmt.Println("Kills: ", el.Text)
						kills := 0
						if eTrTdSpan.Text == "-" {
							kills = -1
						} else {
							kills, _ = strconv.Atoi(eTrTdSpan.Text)
							currRound = tableNum
						}
						md.GameKills = append(md.GameKills, kills)
					} else if colCounter%4 == 3 {
						//fmt.Println("Kill Points: ", el.Text)
						killPoint, _ := strconv.Atoi(eTrTdSpan.Text)
						md.GameKillPoints = append(md.GameKillPoints, killPoint)
					}
					colCounter++
				}
			})

			if len(md.TeamName) > 0 && len(md.GamePlacements) > 0 && md.TeamName != "TBD" {
				var totalPoints string
				if strings.Contains(url, "Finals") {
					totalPoints = eTr.ChildText("td > abbr > b")[:len(eTr.ChildText("td > abbr > b"))-1]
				} else {
					totalPoints = eTr.ChildText("td > abbr > b")[:len(eTr.ChildText("td > abbr > b"))]
				}
				md.TotalPoints, _ = strconv.Atoi(totalPoints)
				//fmt.Println(td)

				matchDataMap[md.TeamName] = md
			}
			rowCounter++
		})
		s.AddMatchDataMap(matchDataMap)
		tableNum++
	})

	/* c.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	}) */

	c.Visit(url)

	//fmt.Printf("tableNum: %v\n", tableNum)
	//fmt.Printf("currRound: %v\n", currRound)
	//fmt.Printf("MatchDataMap: %v\n", *MatchDataMapPtr)

	s.CurrRound = currRound
}
