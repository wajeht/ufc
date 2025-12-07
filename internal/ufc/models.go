package ufc

import "time"

type Event struct {
	Name       string    `json:"name"`
	Headline   string    `json:"headline"`
	Date       string    `json:"date"`
	ParsedDate time.Time `json:"parsed_date"`
	Venue      string    `json:"venue"`
	Location   string    `json:"location"`
	URL        string    `json:"url"`
}

type Fight struct {
	Fighter1    string `json:"fighter1"`
	Fighter2    string `json:"fighter2"`
	WeightClass string `json:"weight_class"`
	Winner      int    `json:"winner,omitempty"`  // 0=no result, 1=fighter1, 2=fighter2
	Method      string `json:"method,omitempty"`  // KO/TKO, Submission, Decision, etc.
	Round       string `json:"round,omitempty"`   // Round number
	Time        string `json:"time,omitempty"`    // Time in round
	Odds1       string `json:"odds1,omitempty"`   // Odds for fighter 1
	Odds2       string `json:"odds2,omitempty"`   // Odds for fighter 2
	Country1    string `json:"country1,omitempty"` // Country for fighter 1
	Country2    string `json:"country2,omitempty"` // Country for fighter 2
}

type EventDetails struct {
	Event
	Fights []Fight `json:"fights"`
}
