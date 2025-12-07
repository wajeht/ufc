package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wajeht/ufc/assets"
	"github.com/wajeht/ufc/internal/ufc"
)

func main() {
	port := flag.String("port", "80", "port to listen on")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		*port = p
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /events.ics", func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.FS.ReadFile("events.ics")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>404 Not Found</title>
    <meta name="robots" content="noindex, nofollow, noarchive, nosnippet, noimageindex">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
</head>
<body>
    <h1>404 Not Found</h1>
    <p>Calendar not found.</p>
    <p><a href="/">Go to homepage</a></p>
</body>
</html>`)
			return
		}

		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\"events.ics\"")
		w.Write(data)
	})

	mux.HandleFunc("GET /events.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.FS.ReadFile("events.json")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>404 Not Found</title>
    <meta name="robots" content="noindex, nofollow, noarchive, nosnippet, noimageindex">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
</head>
<body>
    <h1>404 Not Found</h1>
    <p>Events not found.</p>
    <p><a href="/">Go to homepage</a></p>
</body>
</html>`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.FS.ReadFile("favicon.ico")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})

	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, `User-agent: *
Disallow: /

User-agent: GPTBot
Disallow: /

User-agent: ChatGPT-User
Disallow: /

User-agent: CCBot
Disallow: /

User-agent: anthropic-ai
Disallow: /

User-agent: Google-Extended
Disallow: /
`)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>404 Not Found</title>
    <meta name="robots" content="noindex, nofollow, noarchive, nosnippet, noimageindex">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
</head>
<body>
    <h1>404 Not Found</h1>
    <p>The page you requested could not be found.</p>
    <p><a href="/">Go to homepage</a></p>
</body>
</html>`)
			return
		}

		events, err := ufc.LoadEventsFromFS(assets.FS, "events.json")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>500 Internal Server Error</title>
    <meta name="robots" content="noindex, nofollow, noarchive, nosnippet, noimageindex">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
</head>
<body>
    <h1>500 Internal Server Error</h1>
    <p>Failed to load events.</p>
    <p><a href="/">Go to homepage</a></p>
</body>
</html>`)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>UFC Calendar</title>
    <meta name="robots" content="noindex, nofollow, noarchive, nosnippet, noimageindex">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
</head>
<body>
    <h1>UFC Calendar</h1>
    <p>Subscribe to upcoming UFC events in your calendar app.</p>
    <p><a href="/events.ics">Download Calendar (.ics)</a></p>
    <p>Or subscribe via URL: <code>%s/events.ics</code></p>
    <h2>Upcoming Events (%d)</h2>
    <table border="1" cellpadding="5" cellspacing="0">
    <thead>
        <tr>
            <th>Event</th>
            <th>Date</th>
            <th>Venue</th>
            <th>Location</th>
            <th>Fights</th>
        </tr>
    </thead>
    <tbody>
`, r.Host, len(events))

		for _, e := range events {
			fmt.Fprintf(w, `<tr>`)
			fmt.Fprintf(w, `<td><a href="https://www.ufc.com%s"><strong>%s</strong><br>%s</a></td>`, e.URL, e.Name, e.Headline)
			fmt.Fprintf(w, `<td>%s</td>`, e.Date)
			fmt.Fprintf(w, `<td>%s</td>`, e.Venue)
			fmt.Fprintf(w, `<td>%s</td>`, e.Location)

			// Fights column
			fmt.Fprintf(w, `<td>`)
			if len(e.Fights) > 0 {
				fmt.Fprintf(w, `<details><summary>%d fights</summary>`, len(e.Fights))
				fmt.Fprintf(w, `<table border="1" cellpadding="3" cellspacing="0">`)
				fmt.Fprintf(w, `<thead><tr><th>Weight Class</th><th>Fighter 1</th><th>Odds</th><th></th><th>Fighter 2</th><th>Odds</th><th>Method</th><th>R</th><th>Time</th></tr></thead>`)
				fmt.Fprintf(w, `<tbody>`)
				for _, f := range e.Fights {
					// Fighter names with links and winner indicator
					f1 := f.Fighter1
					f2 := f.Fighter2
					if f.Fighter1URL != "" {
						f1 = fmt.Sprintf(`<a href="%s">%s</a>`, f.Fighter1URL, f.Fighter1)
					}
					if f.Fighter2URL != "" {
						f2 = fmt.Sprintf(`<a href="%s">%s</a>`, f.Fighter2URL, f.Fighter2)
					}
					if f.Country1 != "" {
						f1 += "<br><small>" + f.Country1 + "</small>"
					}
					if f.Country2 != "" {
						f2 += "<br><small>" + f.Country2 + "</small>"
					}
					if f.Winner == 1 {
						f1 = "<strong>" + f1 + "</strong>"
					} else if f.Winner == 2 {
						f2 = "<strong>" + f2 + "</strong>"
					}

					// Odds - show dash if empty or just "-"
					odds1 := f.Odds1
					odds2 := f.Odds2
					if odds1 == "" || odds1 == "-" {
						odds1 = "-"
					}
					if odds2 == "" || odds2 == "-" {
						odds2 = "-"
					}

					// Result
					method := f.Method
					if method == "" {
						method = "-"
					}
					round := f.Round
					if round == "" {
						round = "-"
					}
					ftime := f.Time
					if ftime == "" {
						ftime = "-"
					}

					fmt.Fprintf(w, `<tr>`)
					fmt.Fprintf(w, `<td>%s</td>`, f.WeightClass)
					fmt.Fprintf(w, `<td>%s</td>`, f1)
					fmt.Fprintf(w, `<td>%s</td>`, odds1)
					fmt.Fprintf(w, `<td>vs</td>`)
					fmt.Fprintf(w, `<td>%s</td>`, f2)
					fmt.Fprintf(w, `<td>%s</td>`, odds2)
					fmt.Fprintf(w, `<td>%s</td>`, method)
					fmt.Fprintf(w, `<td>%s</td>`, round)
					fmt.Fprintf(w, `<td>%s</td>`, ftime)
					fmt.Fprintf(w, `</tr>`)
				}
				fmt.Fprintf(w, `</tbody></table></details>`)
			} else {
				fmt.Fprintf(w, `0 fights`)
			}
			fmt.Fprintf(w, `</td></tr>
`)
		}

		fmt.Fprintf(w, `</tbody></table></body></html>`)
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
