package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wajeht/ufc/internal/ufc"
)

func main() {
	port := flag.String("port", "8080", "port to listen on")
	assetsDir := flag.String("assets", "assets", "assets directory")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		*port = p
	}

	mux := http.NewServeMux()

	// Serve ICS calendar
	mux.HandleFunc("GET /events.ics", func(w http.ResponseWriter, r *http.Request) {
		icsPath := filepath.Join(*assetsDir, "events.ics")
		data, err := os.ReadFile(icsPath)
		if err != nil {
			http.Error(w, "Calendar not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\"events.ics\"")
		w.Write(data)
	})

	// Serve JSON events
	mux.HandleFunc("GET /events.json", func(w http.ResponseWriter, r *http.Request) {
		jsonPath := filepath.Join(*assetsDir, "events.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			http.Error(w, "Events not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Home page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		events, err := ufc.LoadEvents(filepath.Join(*assetsDir, "events.json"))
		if err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>UFC Calendar</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: system-ui, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1 { margin-bottom: 10px; }
        .subscribe { background: #d20a0a; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; margin: 10px 0; border-radius: 4px; }
        .event { border-bottom: 1px solid #eee; padding: 15px 0; }
        .event-name { font-weight: bold; font-size: 1.1em; }
        .event-date { color: #666; margin: 5px 0; }
        .event-location { color: #888; font-size: 0.9em; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>UFC Calendar</h1>
    <p>Subscribe to upcoming UFC events in your calendar app.</p>
    <a class="subscribe" href="/events.ics">Download Calendar (.ics)</a>
    <p>Or subscribe via URL: <code>%s/events.ics</code></p>
    <h2>Upcoming Events (%d)</h2>
`, r.Host, len(events))

		for _, e := range events {
			fmt.Fprintf(w, `<div class="event">
        <div class="event-name">%s: %s</div>
        <div class="event-date">%s</div>
        <div class="event-location">%s, %s</div>
        <div>%d fights</div>
    </div>
`, e.Name, e.Headline, e.Date, e.Venue, e.Location, len(e.Fights))
		}

		fmt.Fprintf(w, `</body></html>`)
	})

	addr := ":" + *port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
