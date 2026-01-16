package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
	metrics "github.com/victoriacheng15/personal-reading-analytics/cmd/internal/metrics"
)

// MetricsFetcher defines the interface for fetching metrics
type MetricsFetcher interface {
	FetchMetrics(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error)
}

// DefaultMetricsFetcher implements MetricsFetcher
type DefaultMetricsFetcher struct{}

// fetchMetricsFunc is a package-level variable that can be mocked in tests
var fetchMetricsFunc = metrics.FetchMetricsFromSheets

// FetchMetrics fetches metrics from Google Sheets
func (d *DefaultMetricsFetcher) FetchMetrics(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error) {
	return fetchMetricsFunc(ctx, sheetID, credentialsPath)
}

// loadConfiguration loads environment variables and returns sheetID and credentialsPath
func loadConfiguration() (string, string, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, will use environment variables")
	}

	sheetID := os.Getenv("SHEET_ID")
	credentialsPath := os.Getenv("CREDENTIALS_PATH")

	if sheetID == "" {
		return "", "", fmt.Errorf("SHEET_ID environment variable is required")
	}
	if credentialsPath == "" {
		credentialsPath = "./credentials.json"
	}

	return sheetID, credentialsPath, nil
}

// saveMetrics saves metrics to a JSON file
func saveMetrics(metricsData schema.Metrics) error {
	// Create metrics directory
	if err := os.MkdirAll("metrics", 0755); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Marshal to JSON
	metricsJSON, err := json.MarshalIndent(metricsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Generate filename with date
	dateFilename := metricsData.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)

	// Write to file
	if err := os.WriteFile(metricsFilePath, metricsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	return nil
}

// run executes the main logic and returns an error
func run(ctx context.Context, fetcher MetricsFetcher) error {
	// Load configuration
	sheetID, credentialsPath, err := loadConfiguration()
	if err != nil {
		return err
	}

	// Fetch metrics from Google Sheets
	metricsData, err := fetcher.FetchMetrics(ctx, sheetID, credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// Save metrics
	if err := saveMetrics(metricsData); err != nil {
		return err
	}

	log.Println("✅ Successfully generated metrics from Google Sheets")
	return nil
}

// logFatalf is a package-level variable that can be mocked in tests
var logFatalf = log.Fatalf

func main() {
	ctx := context.Background()
	fetcher := &DefaultMetricsFetcher{}

	if err := run(ctx, fetcher); err != nil {
		logFatalf("Error: %v", err)
	}
}
