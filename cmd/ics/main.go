package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wajeht/ufc/internal/ufc"
)

func main() {
	input := flag.String("i", ufc.DefaultDataFile, "input JSON file")
	output := flag.String("o", ufc.DefaultICSFile, "output ICS file")
	flag.Parse()

	events, err := ufc.LoadEvents(*input)
	if err != nil {
		log.Fatalf("Failed to load events: %v", err)
	}

	fmt.Printf("Loaded %d events from %s\n", len(events), *input)

	cal := ufc.NewCalendar(events)

	if err := os.WriteFile(*output, []byte(cal.String()), 0644); err != nil {
		log.Fatalf("Failed to write ICS file: %v", err)
	}

	fmt.Printf("Generated %s\n", *output)
}
