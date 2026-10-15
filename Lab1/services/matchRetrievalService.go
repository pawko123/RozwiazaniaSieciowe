package services

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"Lab1/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type MatchRetrievalService struct {
	url string
}

func NewMatchRetrievalService(url string) *MatchRetrievalService {
	return &MatchRetrievalService{url: url}
}

var seasonYearRegex = regexp.MustCompile(`(\d{4})/(\d{4})`)

// FetchSchedule downloads the page stream, parses XPath, and builds the team map
func (s *MatchRetrievalService) FetchSchedule() (map[string]*models.Team, error) {
	resp, err := http.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Parse directly from io.Reader stream
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("xpath/dom parse error: %w", err)
	}

	teams := make(map[string]*models.Team)
	getOrCreateTeam := func(name string) *models.Team {
		if team, ok := teams[name]; ok {
			return team
		}
		team := models.NewTeam(name)
		teams[name] = team
		return team
	}

	seasonNode := htmlquery.FindOne(doc, "//h2")
	matches := seasonYearRegex.FindStringSubmatch(htmlquery.InnerText(seasonNode))

	startYear, _ := strconv.Atoi(matches[1])
	endYear, _ := strconv.Atoi(matches[2])

	rows := htmlquery.Find(doc, "//table//tbody/tr")
	for _, row := range rows {
		hostNode := htmlquery.FindOne(row, "./td[1]//a/span")
		guestNode := htmlquery.FindOne(row, "./td[2]//a/span")
		dateNode := htmlquery.FindOne(row, "./td[3]//span[1]")
		timeNode := htmlquery.FindOne(row, "./td[3]//span[2]")

		if hostNode == nil || guestNode == nil || dateNode == nil || timeNode == nil {
			continue
		}

		hostName := strings.TrimSpace(htmlquery.InnerText(hostNode))
		guestName := strings.TrimSpace(htmlquery.InnerText(guestNode))
		dateStr := strings.TrimSpace(htmlquery.InnerText(dateNode))
		timeStr := strings.TrimSpace(htmlquery.InnerText(timeNode))

		matchDate := s.parseDate(dateStr, timeStr, startYear, endYear)
		hostPts, guestPts := s.parseScore(row)

		hTeam := getOrCreateTeam(hostName)
		gTeam := getOrCreateTeam(guestName)

		match := &models.Match{
			Host:        hTeam,
			Guest:       gTeam,
			HostPoints:  hostPts,
			GuestPoints: guestPts,
			Date:        matchDate,
		}

		hTeam.HomeMatches = append(hTeam.HomeMatches, match)
		gTeam.AwayMatches = append(gTeam.AwayMatches, match)
	}

	return teams, nil
}

func (s *MatchRetrievalService) parseDate(datePart, timePart string, startYear, endYear int) time.Time {
	var day, month, hour, minute int
	fmt.Sscanf(datePart, "%d.%d", &day, &month)
	fmt.Sscanf(timePart, "%d:%d", &hour, &minute)

	year := startYear
	if month < 8 {
		year = endYear
	}

	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
}

func (s *MatchRetrievalService) parseScore(row *html.Node) (int, int) {
	scoreNode := htmlquery.FindOne(row, "./td[5]")
	if scoreNode == nil {
		return 0, 0
	}
	text := strings.TrimSpace(htmlquery.InnerText(scoreNode))
	parts := strings.Split(text, ":")
	if len(parts) != 2 {
		return 0, 0
	}
	hostPoints, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	guestPoints, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return hostPoints, guestPoints
}
