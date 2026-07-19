package ufc

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"valid epoch", "1784419200", time.Unix(1784419200, 0).UTC()},
		{"surrounding whitespace", "  1784419200  ", time.Unix(1784419200, 0).UTC()},
		{"empty", "", time.Time{}},
		{"zero", "0", time.Time{}},
		{"non-numeric", "Sat, May 9", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTimestamp(tt.in); !got.Equal(tt.want) {
				t.Errorf("parseTimestamp(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// parseEventCard must read the authoritative Unix timestamp, not the year-less
// display text. UFC 328 (data-main-card-timestamp 1778374800) is 2026-05-09.
func TestParseEventCardUsesTimestamp(t *testing.T) {
	const card = `
<article class="c-card-event--result">
  <h3 class="c-card-event--result__headline"><a href="/event/ufc-328">Chimaev vs Strickland</a></h3>
  <div class="c-card-event--result__date tz-change-data"
       data-main-card="Sat, May 9 / 9:00 PM EDT"
       data-main-card-timestamp="1778374800">Sat, May 9 / 9:00 PM EDT / Main Card</div>
</article>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(card))
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}

	s := NewScraper()
	event := s.parseEventCard(doc.Find("article.c-card-event--result"))

	want := time.Unix(1778374800, 0).UTC() // 2026-05-10T01:00Z == 9:00 PM EDT, May 9
	if !event.ParsedDate.Equal(want) {
		t.Errorf("ParsedDate = %v, want %v (must derive year from timestamp, not display text)", event.ParsedDate, want)
	}
	if event.ParsedDate.Year() != 2026 {
		t.Errorf("year = %d, want 2026 (regression: year-less text must not be promoted to a future year)", event.ParsedDate.Year())
	}
}

// filterUpcoming is the guard against stale events: a card that already
// happened must never appear as upcoming.
func TestFilterUpcomingDropsPastEvents(t *testing.T) {
	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)

	events := []Event{
		{Name: "UFC 328", ParsedDate: time.Date(2026, time.May, 9, 21, 0, 0, 0, time.UTC)},         // past
		{Name: "Allen vs Costa", ParsedDate: time.Date(2026, time.May, 16, 20, 0, 0, 0, time.UTC)}, // past
		{Name: "Today", ParsedDate: time.Date(2026, time.July, 18, 20, 0, 0, 0, time.UTC)},         // today, keep
		{Name: "Future", ParsedDate: time.Date(2026, time.August, 1, 17, 0, 0, 0, time.UTC)},       // future, keep
		{Name: "No date", ParsedDate: time.Time{}},                                                 // zero, drop
	}

	got := filterUpcoming(events, now)

	want := []string{"Today", "Future"}
	if len(got) != len(want) {
		t.Fatalf("got %d events %v, want %d %v", len(got), names(got), len(want), want)
	}
	for i, e := range got {
		if e.Name != want[i] {
			t.Errorf("event[%d] = %q, want %q", i, e.Name, want[i])
		}
	}
}

func names(events []Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.Name
	}
	return out
}
