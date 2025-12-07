package ufc

import (
	"fmt"
	"strings"
	"time"
)

type Calendar struct {
	events []*EventDetails
}

func NewCalendar(events []*EventDetails) *Calendar {
	return &Calendar{events: events}
}

func (c *Calendar) String() string {
	var b strings.Builder

	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//ufc//UFC Events//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString("X-WR-CALNAME:UFC Events\r\n")

	for _, e := range c.events {
		b.WriteString(c.formatEvent(e))
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func (c *Calendar) formatEvent(e *EventDetails) string {
	var b strings.Builder

	uid := strings.ReplaceAll(e.URL, "/", "-") + "@ufc.com"

	startUTC := e.ParsedDate.UTC()
	endUTC := startUTC.Add(4 * time.Hour)

	var desc strings.Builder
	desc.WriteString("Main Card:\\n")
	for _, f := range e.Fights {
		desc.WriteString(fmt.Sprintf("• %s vs %s (%s)\\n", f.Fighter1, f.Fighter2, f.WeightClass))
	}
	desc.WriteString(fmt.Sprintf("\\nMore info: %s%s", BaseURL, e.URL))

	location := e.Venue
	if e.Location != "" && !strings.Contains(e.Venue, e.Location) {
		location += ", " + e.Location
	}

	summary := e.Headline
	if e.Name != "" {
		summary = e.Name + ": " + e.Headline
	}

	b.WriteString("BEGIN:VEVENT\r\n")
	b.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	b.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
	b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startUTC.Format("20060102T150405Z")))
	b.WriteString(fmt.Sprintf("DTEND:%s\r\n", endUTC.Format("20060102T150405Z")))
	b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(summary)))
	b.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(location)))
	b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", desc.String()))
	b.WriteString(fmt.Sprintf("URL:%s%s\r\n", BaseURL, e.URL))
	b.WriteString("END:VEVENT\r\n")

	return b.String()
}

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ";", "\\;")
	return s
}
