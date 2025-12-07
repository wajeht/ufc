package ufc

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	BaseURL   = "https://www.ufc.com"
	EventsURL = BaseURL + "/events"
)

var whitespaceRe = regexp.MustCompile(`\s+`)

type Scraper struct {
	client *http.Client
}

func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Scraper) fetch(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; mma-cal/1.0)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	return doc, nil
}

func (s *Scraper) GetEvents() ([]Event, error) {
	doc, err := s.fetch(EventsURL)
	if err != nil {
		return nil, err
	}

	var events []Event

	doc.Find("article.c-card-event--result").Each(func(_ int, sel *goquery.Selection) {
		event := s.parseEventCard(sel)
		if event.URL != "" {
			events = append(events, event)
		}
	})

	sort.Slice(events, func(i, j int) bool {
		return events[i].ParsedDate.Before(events[j].ParsedDate)
	})

	today := time.Now().Truncate(24 * time.Hour)
	var upcoming []Event
	for _, e := range events {
		if !e.ParsedDate.Before(today) {
			upcoming = append(upcoming, e)
		}
	}

	return upcoming, nil
}

func (s *Scraper) parseEventCard(sel *goquery.Selection) Event {
	event := Event{
		URL:      sel.Find("a").First().AttrOr("href", ""),
		Name:     cleanText(sel.Find(".c-card-event--result__logo img").AttrOr("alt", "")),
		Headline: cleanText(sel.Find(".c-card-event--result__headline").Text()),
		Date:     cleanText(sel.Find(".c-card-event--result__date").Text()),
		Venue:    cleanText(sel.Find(".field--name-taxonomy-term-title").Text()),
		Location: cleanText(sel.Find(".address").Text()),
	}

	if event.Name == "" && event.URL != "" {
		event.Name = parseEventNameFromURL(event.URL)
	}

	event.ParsedDate = parseEventDate(event.Date)
	return event
}

func (s *Scraper) GetEventDetails(event Event) (*EventDetails, error) {
	doc, err := s.fetch(BaseURL + event.URL)
	if err != nil {
		return nil, err
	}

	details := &EventDetails{
		Event: event,
	}

	if details.Headline == "" {
		details.Headline = cleanText(doc.Find(".c-hero__headline").Text())
	}

	doc.Find(".l-listing__item").Each(func(_ int, sel *goquery.Selection) {
		if fight := s.parseFight(sel); fight != nil {
			details.Fights = append(details.Fights, *fight)
		}
	})

	return details, nil
}

func (s *Scraper) parseFight(sel *goquery.Selection) *Fight {
	fighters := sel.Find(".c-listing-fight__corner-name")
	if fighters.Length() < 2 {
		return nil
	}

	fighter1 := cleanText(fighters.Eq(0).Text())
	fighter2 := cleanText(fighters.Eq(1).Text())

	if fighter1 == "" || fighter2 == "" {
		return nil
	}

	weightClass := cleanText(sel.Find(".c-listing-fight__class-text").Text())
	if idx := strings.Index(weightClass, "Bout"); idx != -1 {
		weightClass = weightClass[:idx+4]
	}

	return &Fight{
		Fighter1:    fighter1,
		Fighter2:    fighter2,
		WeightClass: weightClass,
	}
}

func cleanText(s string) string {
	return strings.TrimSpace(whitespaceRe.ReplaceAllString(s, " "))
}

func parseEventNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	slug := parts[len(parts)-1]

	if strings.HasPrefix(slug, "ufc-fight-night") {
		return "UFC Fight Night"
	}
	if after, ok :=strings.CutPrefix(slug, "ufc-"); ok  {
		num := after
		if _, err := fmt.Sscanf(num, "%d", new(int)); err == nil {
			return "UFC " + strings.ToUpper(num)
		}
		return "UFC " + num
	}
	return ""
}

func parseEventDate(dateStr string) time.Time {
	parts := strings.Split(dateStr, "/")
	if len(parts) < 2 {
		return time.Time{}
	}

	datePart := strings.TrimSpace(parts[0])
	timePart := strings.TrimSpace(parts[1])

	if idx := strings.Index(datePart, ","); idx != -1 {
		datePart = strings.TrimSpace(datePart[idx+1:])
	}

	loc := extractTimezone(timePart)
	timePart = stripTimezone(timePart)

	combined := fmt.Sprintf("%s %s", datePart, timePart)

	t, err := time.ParseInLocation("Jan 2 3:04 PM", combined, loc)
	if err != nil {
		return time.Time{}
	}

	return adjustYear(t, loc)
}

func extractTimezone(s string) *time.Location {
	zones := map[string]int{
		"EST": -5, "EDT": -4,
		"CST": -6, "CDT": -5,
		"PST": -8, "PDT": -7,
	}

	for tz, offset := range zones {
		if strings.HasSuffix(s, tz) {
			return time.FixedZone(tz, offset*3600)
		}
	}
	return time.UTC
}

func stripTimezone(s string) string {
	for _, tz := range []string{"EST", "EDT", "CST", "CDT", "PST", "PDT"} {
		s = strings.TrimSuffix(s, tz)
	}
	return strings.TrimSpace(s)
}

func adjustYear(t time.Time, loc *time.Location) time.Time {
	now := time.Now()
	t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, loc)

	if t.Before(now.AddDate(0, -2, 0)) {
		t = t.AddDate(1, 0, 0)
	}
	return t
}
