package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	schema "github.com/victoriacheng15/personal-reading-analytics/internal"
	metrics "github.com/victoriacheng15/personal-reading-analytics/internal/metrics"
)

// MetricsFetcher defines the interface for fetching metrics
type MetricsFetcher interface {
	FetchMetrics(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error)
}

// DefaultMetricsFetcher implements MetricsFetcher
type DefaultMetricsFetcher struct{}

// fetchMetricsFunc is a package-level variable that can be mocked in tests
var fetchMetricsFunc = metrics.FetchMetricsFromSheets

// logFatalf is a package-level variable that can be mocked in tests
var logFatalf = log.Fatalf

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, will use environment variables")
	}

	fetchFlag := flag.Bool("fetch", false, "Only fetch metrics from Google Sheets")
	summarizeFlag := flag.Bool("summarize", false, "Only generate AI delta analysis for the latest metrics")
	flag.Parse()

	ctx := context.Background()
	fetcher := &DefaultMetricsFetcher{}

	if err := execute(ctx, fetcher, *fetchFlag, *summarizeFlag); err != nil {
		logFatalf("%v", err)
	}
}

// FetchMetrics fetches metrics from Google Sheets
func (d *DefaultMetricsFetcher) FetchMetrics(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error) {
	return fetchMetricsFunc(ctx, sheetID, credentialsPath)
}

// loadConfiguration loads environment variables and returns sheetID and credentialsPath
func loadConfiguration() (string, string, error) {
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
func saveMetrics(metricsData schema.Metrics) (string, error) {
	// Create metrics directory
	if err := os.MkdirAll("metrics", 0755); err != nil {
		return "", fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Marshal to JSON
	metricsJSON, err := json.MarshalIndent(metricsData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Generate filename with date

	dateFilename := metricsData.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)

	// Write to file
	if err := os.WriteFile(metricsFilePath, metricsJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write metrics file: %w", err)
	}

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	return dateFilename, nil
}

// runFetch executes the fetch logic
func runFetch(ctx context.Context, fetcher MetricsFetcher) (string, *schema.Metrics, error) {
	// Load configuration
	sheetID, credentialsPath, err := loadConfiguration()
	if err != nil {
		return "", nil, err
	}

	// Fetch metrics from Google Sheets
	metricsData, err := fetcher.FetchMetrics(ctx, sheetID, credentialsPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// Save metrics
	filename, err := saveMetrics(metricsData)
	if err != nil {
		return "", nil, err
	}

	log.Println("✅ Successfully generated metrics from Google Sheets")
	return filename, &metricsData, nil
}

// runDeltaAnalysis executes the AI delta analysis logic
func runDeltaAnalysis(ctx context.Context, filename string, metricsData *schema.Metrics) error {
	if filename == "" || metricsData == nil {
		return fmt.Errorf("metrics data not provided for delta analysis")
	}

	// Generate AI Delta Analysis
	if err := metrics.GenerateAndSaveDeltaAnalysis(ctx, "metrics", filename, metricsData); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating AI delta analysis: %v\n", err)
	}
	log.Println("✅ AI Delta Analysis generated and saved.")
	return nil
}

// execute runs the application logic based on flags
func execute(ctx context.Context, fetcher MetricsFetcher, fetchFlag, summarizeFlag bool) error {
	// Default behavior: Run both
	runBoth := !fetchFlag && !summarizeFlag

	var filename string
	var metricsData *schema.Metrics
	var err error

	if runBoth || fetchFlag {
		filename, metricsData, err = runFetch(ctx, fetcher)
		if err != nil {
			return fmt.Errorf("Error fetching metrics: %w", err)
		}
	}

	if runBoth || summarizeFlag {
		if summarizeFlag && filename == "" {
			// Standalone mode: Find latest file in metrics/ dir
			entries, err := os.ReadDir("metrics")
			if err == nil {
				// Find last one
				var lastFile string
				for _, e := range entries {
					if !e.IsDir() {
						lastFile = e.Name()
					}
				}
				if lastFile != "" {
					filename = lastFile
					// Load it
					bytes, _ := os.ReadFile("metrics/" + lastFile)
					var m schema.Metrics
					if json.Unmarshal(bytes, &m) == nil {
						metricsData = &m
					}
				}
			}
		}

		if metricsData != nil {
			if err := runDeltaAnalysis(ctx, filename, metricsData); err != nil {
				log.Printf("Warning: AI delta analysis failed: %v", err)
				// Don't error here, as the primary metrics are safe
			}
		} else {
			log.Println("No metrics data available to perform delta analysis.")
		}
	}
	return nil
}
