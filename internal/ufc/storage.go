package ufc

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
)

const (
	DefaultDataFile = "assets/events.json"
	DefaultICSFile  = "assets/events.ics"
)

func SaveEvents(events []*EventDetails, filename string) error {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling events: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func LoadEvents(filename string) ([]*EventDetails, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var events []*EventDetails
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("unmarshaling events: %w", err)
	}

	return events, nil
}

func LoadEventsFromFS(fsys fs.FS, filename string) ([]*EventDetails, error) {
	data, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var events []*EventDetails
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("unmarshaling events: %w", err)
	}

	return events, nil
}
