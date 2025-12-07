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
    <ul>
`, r.Host, len(events))

		for _, e := range events {
			fmt.Fprintf(w, `<li><a href="https://www.ufc.com%s"><strong>%s: %s</strong></a><br>%s<br>%s, %s<br>`, e.URL, e.Name, e.Headline, e.Date, e.Venue, e.Location)
			if len(e.Fights) > 0 {
				fmt.Fprintf(w, `<details><summary>%d fights</summary><ul>`, len(e.Fights))
				for _, f := range e.Fights {
					// Fighter names with winner indicator
					f1 := f.Fighter1
					f2 := f.Fighter2
					if f.Winner == 1 {
						f1 = "<strong>" + f1 + " (W)</strong>"
					} else if f.Winner == 2 {
						f2 = "<strong>" + f2 + " (W)</strong>"
					}

					// Build fight details line
					fmt.Fprintf(w, `<li>%s`, f1)
					if f.Country1 != "" {
						fmt.Fprintf(w, ` <small>(%s)</small>`, f.Country1)
					}
					if f.Odds1 != "" {
						fmt.Fprintf(w, ` [%s]`, f.Odds1)
					}
					fmt.Fprintf(w, ` vs %s`, f2)
					if f.Country2 != "" {
						fmt.Fprintf(w, ` <small>(%s)</small>`, f.Country2)
					}
					if f.Odds2 != "" {
						fmt.Fprintf(w, ` [%s]`, f.Odds2)
					}
					fmt.Fprintf(w, ` - <em>%s</em>`, f.WeightClass)

					// Result details if available
					if f.Method != "" {
						fmt.Fprintf(w, ` | %s`, f.Method)
						if f.Round != "" {
							fmt.Fprintf(w, ` R%s`, f.Round)
						}
						if f.Time != "" {
							fmt.Fprintf(w, ` %s`, f.Time)
						}
					}
					fmt.Fprintf(w, `</li>`)
				}
				fmt.Fprintf(w, `</ul></details>`)
			} else {
				fmt.Fprintf(w, `0 fights`)
			}
			fmt.Fprintf(w, `</li>
`)
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
