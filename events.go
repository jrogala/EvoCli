package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type eventsFile struct {
	Events []EventItem `json:"events"`
}

type carsFile struct {
	Cars []carEntry `json:"cars"`
}

type carEntry struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// loadAllCars reads cars.json and returns every car_name. Used by the
// -cars all shortcut so server operators don't have to enumerate all 89
// presets by hand.
func loadAllCars(serverDir string) ([]AllowedCar, error) {
	data, err := os.ReadFile(filepath.Join(serverDir, "cars.json"))
	if err != nil {
		return nil, err
	}
	var f carsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse cars.json: %w", err)
	}
	out := make([]AllowedCar, len(f.Cars))
	for i, c := range f.Cars {
		out[i] = AllowedCar{CarName: c.Name}
	}
	return out, nil
}

func loadEvents(path string) ([]EventItem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f eventsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return f.Events, nil
}

func findEvent(events []EventItem, track, layout string) (EventItem, error) {
	var matchesForTrack []string
	for _, e := range events {
		if strings.EqualFold(e.Track, track) {
			matchesForTrack = append(matchesForTrack, e.Layout)
			if strings.EqualFold(e.Layout, layout) {
				return e, nil
			}
		}
	}
	if len(matchesForTrack) == 0 {
		return EventItem{}, fmt.Errorf("track %q not found in catalogue", track)
	}
	return EventItem{}, fmt.Errorf("layout %q not found for track %q (available: %s)",
		layout, track, strings.Join(matchesForTrack, ", "))
}

// resolveEvent picks the right events_*.json for a game mode and returns
// the matching EventItem. The game mode is the short form ("practice"),
// not the proto enum name.
func resolveEvent(serverDir, gameMode, track, layout string) (EventItem, error) {
	var catalogue string
	switch gameMode {
	case "practice":
		catalogue = "events_practice.json"
	case "race_weekend":
		catalogue = "events_race_weekend.json"
	default:
		return EventItem{}, fmt.Errorf("no events catalogue for game mode %q", gameMode)
	}
	events, err := loadEvents(filepath.Join(serverDir, catalogue))
	if err != nil {
		return EventItem{}, err
	}
	return findEvent(events, track, layout)
}
