package webscraper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// Team data struct
type MatchData struct {
	teamName            string
	standing            int
	totalPoints         int
	round               int
	gamePlacements      []int
	gamePlacementPoints []int
	gameKills           []int
	gameKillPoints      []int
}

type TeamStanding struct {
	TeamName    string
	Standing    int
	TotalPoints int
}

type GroupStageStandings struct {
	Standings []*TeamStanding
}

func NewTeamStanding(teamName string, standing int, totalPoints int) *TeamStanding {
	ts := TeamStanding{
		TeamName:    teamName,
		Standing:    standing,
		TotalPoints: totalPoints,
	}
	return &ts
}

func (g *GroupStageStandings) AddTeamStanding(ts *TeamStanding) {
	g.Standings = append(g.Standings, ts)
}

func (g *GroupStageStandings) SortByStanding() {
	// Sort kps by the descending order of total points
	sort.Slice(g.Standings, func(i, j int) bool {
		return g.Standings[i].Standing < g.Standings[j].Standing
	})
}

func (g *GroupStageStandings) GetData(url string) {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	c.OnHTML("table.wikitable.wikitable-bordered.wikitable-striped [data-toggle-area-content='6']", func(e *colly.HTMLElement) {
		// Get the text of the <th> tag and convert it to int
		standing, _ := strconv.Atoi(strings.Trim(e.ChildText("th"), "."))
		//fmt.Printf("Standing: %v\n", standing)

		// Get the value of "data-highlightingkey" attribute of the first <td> tag
		teamName := e.ChildText("td:nth-child(2)")
		//fmt.Printf("Team Name: %v\n", teamName)

		// Get the text of the second <td> tag
		totalPoints, _ := strconv.Atoi(e.ChildText("td:nth-child(3) > b"))
		//fmt.Printf("Total Points: %v\n", totalPoints)

		// Create a TeamStanding struct
		ts := NewTeamStanding(teamName, standing, totalPoints)
		g.AddTeamStanding(ts)
	})

	c.Visit(url)

	g.SortByStanding()
	//fmt.Printf("Standings: %v\n", g.Standings)
}

// MatchData constructor
func NewMatchData(teamName string) *MatchData {
	td := MatchData{teamName: teamName}
	return &td
}

func GetMatchData(url string, MatchDataMapPtr *map[int]map[string]MatchData) int {
	c := colly.NewCollector()
	tableNum := 0
	currRound := 0

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
		fmt.Println("Table No: ", tableNum)
		//fmt.Println(e.Text)

		// Create a map of the current table to store each team's data
		(*MatchDataMapPtr)[tableNum] = make(map[string]MatchData)

		rowCounter := 0
		e.ForEach("tr", func(_ int, eTr *colly.HTMLElement) {
			//fmt.Printf("Row%d: %s\n", rowCounter, eTr.Text)
			//fmt.Println("Row No: ", rowCounter)
			colCounter := 0
			td := *NewMatchData("")
			td.round = tableNum
			eTr.ForEach("td > span", func(_ int, eTrTdSpan *colly.HTMLElement) {
				if eTrTdSpan.Attr("class") == "team-template-team-short" {
					//fmt.Println("Team: ", eTrTdSpan.Text)
					td.teamName = strings.Trim(eTrTdSpan.Text, " ")
					td.standing = rowCounter - 2 // -2 for the headers
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
						td.gamePlacements = append(td.gamePlacements, placement)
					} else if colCounter%4 == 1 {
						//fmt.Println("Placement Point: ", el.Text)
						placementPoint, _ := strconv.Atoi(eTrTdSpan.Text)
						td.gamePlacementPoints = append(td.gamePlacementPoints, placementPoint)
					} else if colCounter%4 == 2 {
						//fmt.Println("Kills: ", el.Text)
						kills := 0
						if eTrTdSpan.Text == "-" {
							kills = -1
						} else {
							kills, _ = strconv.Atoi(eTrTdSpan.Text)
							currRound = tableNum
						}
						td.gameKills = append(td.gameKills, kills)
					} else if colCounter%4 == 3 {
						//fmt.Println("Kill Points: ", el.Text)
						killPoint, _ := strconv.Atoi(eTrTdSpan.Text)
						td.gameKillPoints = append(td.gameKillPoints, killPoint)
					}
					colCounter++
				}
			})

			if len(td.teamName) > 0 && len(td.gamePlacements) > 0 && td.teamName != "TBD" {
				var totalPoints string
				if strings.Contains(url, "Finals") {
					totalPoints = eTr.ChildText("td > abbr > b")[:len(eTr.ChildText("td > abbr > b"))-1]
				} else {
					totalPoints = eTr.ChildText("td > abbr > b")[:len(eTr.ChildText("td > abbr > b"))]
				}
				td.totalPoints, _ = strconv.Atoi(totalPoints)
				//fmt.Println(td)

				// Add team data to the map
				(*MatchDataMapPtr)[tableNum][td.teamName] = td
			}
			rowCounter++
		})
		tableNum++
	})

	/* c.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	}) */

	c.Visit(url)

	fmt.Printf("tableNum: %v\n", tableNum)
	fmt.Printf("currRound: %v\n", currRound)
	//fmt.Printf("MatchDataMap: %v\n", *MatchDataMapPtr)

	return currRound
}
