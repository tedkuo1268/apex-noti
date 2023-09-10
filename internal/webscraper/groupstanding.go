package webscraper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

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

func (g *GroupStageStandings) SortByPoints() {
	sort.Slice(g.Standings, func(i, j int) bool {
		return g.Standings[i].TotalPoints > g.Standings[j].TotalPoints
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
