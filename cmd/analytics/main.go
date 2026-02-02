package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
	analytics "github.com/victoriacheng15/personal-reading-analytics/cmd/internal/analytics"
)

// getMetricsDates returns all YYYY-MM-DD dates from JSON files in metrics/ folder, sorted descending
func getMetricsDates() ([]string, error) {
	entries, err := os.ReadDir("metrics")
	if err != nil {
		return nil, fmt.Errorf("unable to read metrics directory: %w", err)
	}

	var dates []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			date := strings.TrimSuffix(entry.Name(), ".json")
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("no valid metrics files found")
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates, nil
}

// loadMetricsByDate reads a specific metrics JSON file from metrics/ folder
func loadMetricsByDate(date string) (schema.Metrics, error) {
	filename := fmt.Sprintf("metrics/%s.json", date)
	data, err := os.ReadFile(filename)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to read metrics file %s: %w", filename, err)
	}

	var metrics schema.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to parse metrics JSON from %s: %w", filename, err)
	}

	return metrics, nil
}

func main() {
	// 1. Get all available metrics dates
	dates, err := getMetricsDates()
	if err != nil {
		log.Fatalf("Failed to discover metrics: %v", err)
	}

	// 2. Initialize Analytics Service
	service := analytics.NewAnalyticsService("site")

	// 3. Multi-pass generation
	for i, date := range dates {
		log.Printf("[%d/%d] Generating reports for %s...\n", i+1, len(dates), date)

		metrics, err := loadMetricsByDate(date)
		if err != nil {
			log.Printf("⚠️ Warning: Skipping %s: %v\n", date, err)
			continue
		}

		// Historical: ONLY analytics.html in site/history/YYYY-MM-DD
		err = service.GenerateAnalyticsOnly(metrics, analytics.GenConfig{
			OutputDir:    filepath.Join("site", "history", date),
			BaseURL:      "../../",
			IsHistorical: true,
			HistoryDates: dates,
			ReportDate:   date,
		})
		if err != nil {
			log.Printf("⚠️ Warning: Failed historical generation for %s: %v\n", date, err)
		}

		// Latest (root): ALL pages in site/
		if i == 0 {
			err = service.GenerateFullSite(metrics, analytics.GenConfig{
				OutputDir:    "site",
				BaseURL:      "./",
				IsHistorical: false,
				HistoryDates: dates,
				ReportDate:   date,
			})
			if err != nil {
				log.Fatalf("Failed to generate latest site: %v", err)
			}
		}
	}

	log.Println("✅ Successfully generated all historical and latest analytics")
}
