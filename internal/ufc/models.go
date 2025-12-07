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
}

type EventDetails struct {
	Event
	Fights []Fight `json:"fights"`
}
