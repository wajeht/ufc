# ufc

Subscribe to UFC events in your calendar app.

Scrapes upcoming events from ufc.com, generates an `.ics` calendar file, and serves it via a simple web server.

## Usage

```
# fetch events
go run ./cmd/fetch

# generate ics
go run ./cmd/ics

# start web server
go run ./cmd/web
```

## Endpoints

- `/` — upcoming events
- `/events.ics` — calendar subscription
- `/events.json` — raw event data
- `/health` — health check

## Subscribe

Add the `/events.ics` URL to your calendar app (Google Calendar, Apple Calendar, etc.) as a subscription.
