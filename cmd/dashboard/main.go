package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
	dashboard "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal/dashboard"
)

// loadLatestMetrics reads the most recent metrics JSON file from metrics/ folder
func loadLatestMetrics() (schema.Metrics, error) {
	entries, err := os.ReadDir("metrics")
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to read metrics directory: %w", err)
	}

	if len(entries) == 0 {
		return schema.Metrics{}, fmt.Errorf("no metrics files found in metrics/ folder")
	}

	// Find the latest metrics file (they are named YYYY-MM-DD.json)
	var jsonFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			jsonFiles = append(jsonFiles, entry.Name())
		}
	}

	if len(jsonFiles) == 0 {
		return schema.Metrics{}, fmt.Errorf("no valid metrics files found")
	}

	// Sort descending (latest first, since YYYY-MM-DD.json is lexicographically ordered)
	sort.Sort(sort.Reverse(sort.StringSlice(jsonFiles)))
	latestFile := jsonFiles[0]

	log.Printf("Loading metrics from: metrics/%s\n", latestFile)

	// Read and parse the JSON file
	data, err := os.ReadFile(fmt.Sprintf("metrics/%s", latestFile))
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to read metrics file: %w", err)
	}

	var metrics schema.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to parse metrics JSON: %w", err)
	}

	return metrics, nil
}

func main() {
	// Load latest metrics from metrics/ folder
	metrics, err := loadLatestMetrics()
	if err != nil {
		log.Fatalf("Failed to load metrics: %v", err)
	}

	// Initialize Dashboard Service
	service := dashboard.NewDashboardService("site")

	// Generate HTML dashboard
	if err := service.Generate(metrics); err != nil {
		log.Fatalf("failed to generate dashboard: %v", err)
	}

	log.Println("✅ Successfully generated dashboard from metrics")
}
