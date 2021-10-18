package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

var (
	URL        = "http://stats.sportsadmin.dk/schedule.aspx?tournamentID=1783"
	timeformat = "02-01-2006 15:04"
)

type Week struct {
	Round   int
	Matches []MatchRow
}

type MatchRow struct {
	Date     time.Time
	HomeTeam Team
	AwayTeam Team
	Result   string
	Winner   Team
}

type Team struct {
	Name string
	Logo string
}

func main() {
	schedule := getFullSchedule()
	// fmt.Println(len(schedule))
	rounds := weekSplitter(schedule)
	_ = rounds
	var counter int
	for _, v := range rounds {
		for _, v2 := range v.Matches {
			fmt.Println(v2.Date)
			counter += 1
		}
	}
	fmt.Println(counter)
}

func getFullSchedule() []MatchRow {
	var headings, row []string
	var rows [][]string
	var matches []MatchRow

	res, err := http.Get(URL)
	if err != nil {
		log.Error().Err(err).Msg("could not contact sportsadmin")
		return nil
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("could not get document from response")
		return nil
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				headings = append(headings, tableheading.Text())
			})
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			rows = append(rows, row)
			row = nil
		})
	})

	for i := 1; i < len(rows); i++ {
		date, err := time.Parse(timeformat, rows[i][0]+" "+rows[i][1])
		if err != nil {
			log.Error().Err(err).Msg("could not parse time")
			continue
		}

		hometeam := Team{rows[i][3], ""}
		awayteam := Team{rows[i][4], ""}
		result := rows[i][5]
		match := MatchRow{
			Date:     date,
			HomeTeam: hometeam,
			AwayTeam: awayteam,
			Result:   result,
		}

		matches = append(matches, match)

	}

	return matches
}

func checkWinner(hometeam, awayteam Team, result string) Team {
	values := strings.Split(result, "-")
	fmt.Println(values)
	if values[0] > values[1] {
		return hometeam
	} else if values[1] > values[0] {
		return awayteam
	} else {
		return Team{}
	}
}

func weekSplitter(matches []MatchRow) []Week {
	var roundmatches []MatchRow
	var weeks []Week
	starttime := matches[0].Date.Add(-time.Duration(matches[0].Date.Hour()) * time.Hour)
	endtime := starttime.Add(8 * 24 * time.Hour)

	var round = 0
	for _, v := range matches {
		fmt.Println(starttime, endtime, v.Date)
		if v.Date.After(starttime) && v.Date.Before(endtime) || v.Date.Day() == endtime.Day() || v.Date.Day() == starttime.Day() {
			fmt.Println(v.HomeTeam, v.AwayTeam, "between")
			roundmatches = append(roundmatches, v)
		} else {
			weeks = append(weeks, Week{round, roundmatches})
			roundmatches = nil
			roundmatches = append(roundmatches, v)
			round += 1
			starttime = starttime.Add(7 * 24 * time.Hour)
			endtime = endtime.Add(7 * 24 * time.Hour)
			fmt.Println("New timerange: ", starttime, endtime)
		}
	}

	weeks = append(weeks, Week{round, roundmatches})
	return weeks
}
