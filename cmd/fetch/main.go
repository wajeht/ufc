package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/wajeht/ufc/internal/ufc"
)

func main() {
	output := flag.String("o", ufc.DefaultDataFile, "output JSON file")
	flag.Parse()

	scraper := ufc.NewScraper()

	fmt.Println("Fetching UFC events...")

	events, err := scraper.GetEvents()
	if err != nil {
		log.Fatalf("Failed to fetch events: %v", err)
	}

	fmt.Printf("Found %d upcoming events\n\n", len(events))

	var allDetails []*ufc.EventDetails
	for i, event := range events {
		fmt.Printf("[%d/%d] %s - %s\n", i+1, len(events), event.Name, event.Headline)

		details, err := scraper.GetEventDetails(event)
		if err != nil {
			fmt.Printf("        Error: %v\n", err)
			continue
		}

		fmt.Printf("        %d fights\n", len(details.Fights))
		allDetails = append(allDetails, details)
	}

	if err := ufc.SaveEvents(allDetails, *output); err != nil {
		log.Fatalf("Failed to save events: %v", err)
	}

	fmt.Printf("\nSaved %d events to %s\n", len(allDetails), *output)
}
