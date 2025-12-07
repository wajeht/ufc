package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wajeht/ufc/internal/ufc"
)

func main() {
	port := flag.String("port", "80", "port to listen on")
	assetsDir := flag.String("assets", "assets", "assets directory")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		*port = p
	}

	mux := http.NewServeMux()

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

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

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
</head>
<body>
    <h1>UFC Calendar</h1>
    <p>Subscribe to upcoming UFC events in your calendar app.</p>
    <p><a href="/events.ics">Download Calendar (.ics)</a></p>
    <p>Or subscribe via URL: <code>%s/events.ics</code></p>
    <h2>Upcoming Events (%d)</h2>
    <ul>
`, r.Host, len(events))

		for _, e := range events {
			fmt.Fprintf(w, `<li><strong>%s: %s</strong><br>%s<br>%s, %s<br>%d fights</li>
`, e.Name, e.Headline, e.Date, e.Venue, e.Location, len(e.Fights))
		}

		fmt.Fprintf(w, `</ul></body></html>`)
	})

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Server listening on http://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server stopped")
}
