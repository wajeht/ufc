package ufc

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ufc-cal/1.0)")

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

	return filterUpcoming(events, time.Now()), nil
}

// filterUpcoming returns events on or after the start of the given day, sorted
// chronologically. Events with a zero ParsedDate sort first and are dropped.
func filterUpcoming(events []Event, now time.Time) []Event {
	sort.Slice(events, func(i, j int) bool {
		return events[i].ParsedDate.Before(events[j].ParsedDate)
	})

	today := now.Truncate(24 * time.Hour)
	var upcoming []Event
	for _, e := range events {
		if !e.ParsedDate.Before(today) {
			upcoming = append(upcoming, e)
		}
	}

	return upcoming
}

func (s *Scraper) parseEventCard(sel *goquery.Selection) Event {
	dateSel := sel.Find(".c-card-event--result__date")

	event := Event{
		URL:      sel.Find("a").First().AttrOr("href", ""),
		Name:     cleanText(sel.Find(".c-card-event--result__logo img").AttrOr("alt", "")),
		Headline: cleanText(sel.Find(".c-card-event--result__headline").Text()),
		Date:     cleanText(dateSel.Text()),
		Venue:    cleanText(sel.Find(".field--name-taxonomy-term-title").Text()),
		Location: cleanText(sel.Find(".address").Text()),
	}

	if event.Name == "" && event.URL != "" {
		event.Name = parseEventNameFromURL(event.URL)
	}

	// UFC embeds the authoritative event time as a Unix timestamp on the date
	// element. Use it directly rather than reconstructing a year from the
	// year-less display text.
	event.ParsedDate = parseTimestamp(dateSel.AttrOr("data-main-card-timestamp", ""))
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

	fight := &Fight{
		Fighter1:    fighter1,
		Fighter2:    fighter2,
		WeightClass: weightClass,
	}

	// Determine winner (red corner is fighter1, blue corner is fighter2)
	if sel.Find(".c-listing-fight__corner--red .c-listing-fight__outcome--win").Length() > 0 {
		fight.Winner = 1
	} else if sel.Find(".c-listing-fight__corner--blue .c-listing-fight__outcome--win").Length() > 0 {
		fight.Winner = 2
	}

	// Extract round, time, method from result sections
	sel.Find(".c-listing-fight__result").Each(func(_ int, result *goquery.Selection) {
		label := cleanText(result.Find(".c-listing-fight__result-label").Text())
		value := cleanText(result.Find(".c-listing-fight__result-text").Text())
		switch label {
		case "Round":
			fight.Round = value
		case "Time":
			fight.Time = value
		case "Method":
			fight.Method = value
		}
	})

	// Extract odds
	odds := sel.Find(".c-listing-fight__odds-amount")
	if odds.Length() >= 2 {
		fight.Odds1 = cleanText(odds.Eq(0).Text())
		fight.Odds2 = cleanText(odds.Eq(1).Text())
	}

	// Extract countries
	countries := sel.Find(".c-listing-fight__country-text")
	if countries.Length() >= 2 {
		fight.Country1 = cleanText(countries.Eq(0).Text())
		fight.Country2 = cleanText(countries.Eq(1).Text())
	}

	fight.Fighter1Image = absoluteURL(sel.Find(".c-listing-fight__corner-image--red img").First().AttrOr("src", ""))
	fight.Fighter2Image = absoluteURL(sel.Find(".c-listing-fight__corner-image--blue img").First().AttrOr("src", ""))

	// Extract athlete URLs from corner name links
	fighterLinks := sel.Find(".c-listing-fight__corner-name a[href*='athlete']")
	if fighterLinks.Length() >= 2 {
		fight.Fighter1URL = absoluteURL(fighterLinks.Eq(0).AttrOr("href", ""))
		fight.Fighter2URL = absoluteURL(fighterLinks.Eq(1).AttrOr("href", ""))
	}

	return fight
}

func absoluteURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "//") {
		return "https:" + s
	}
	if strings.HasPrefix(s, "/") {
		return BaseURL + s
	}
	return s
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
	if after, ok := strings.CutPrefix(slug, "ufc-"); ok {
		num := after
		if _, err := fmt.Sscanf(num, "%d", new(int)); err == nil {
			return "UFC " + strings.ToUpper(num)
		}
		return "UFC " + num
	}
	return ""
}

// parseTimestamp converts UFC's Unix epoch (seconds) into a UTC time. The
// value carries the full, timezone-correct date, so no year inference is
// needed. Returns the zero time when the attribute is missing or malformed,
// which callers treat as "past" and filter out.
func parseTimestamp(ts string) time.Time {
	secs, err := strconv.ParseInt(strings.TrimSpace(ts), 10, 64)
	if err != nil || secs == 0 {
		return time.Time{}
	}
	return time.Unix(secs, 0).UTC()
}
