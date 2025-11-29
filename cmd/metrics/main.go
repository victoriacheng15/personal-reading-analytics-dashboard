package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// normalizeSourceName converts source names to proper capitalization
func normalizeSourceName(name string) string {
	sourceMap := map[string]string{
		"substack":     "Substack",
		"freecodecamp": "freeCodeCamp",
		"github":       "GitHub",
		"shopify":      "Shopify",
		"stripe":       "Stripe",
	}

	// Convert to lowercase for comparison
	lower := strings.ToLower(name)

	// Return normalized name if found, otherwise return original
	if normalized, exists := sourceMap[lower]; exists {
		return normalized
	}
	return name
}

// fetchMetricsFromSheets retrieves and calculates metrics from Google Sheets
func fetchMetricsFromSheets(ctx context.Context, spreadsheetID, credentialsPath string) (schema.Metrics, error) {
	// Create Sheets service
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to create sheets client: %w", err)
	}

	// Get all sheets to find sheet names
	spreadsheet, err := client.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to retrieve spreadsheet: %w", err)
	}

	// Find Articles and Providers sheets
	articlesSheet := "articles"
	providersSheet := "providers"
	for _, sheet := range spreadsheet.Sheets {
		title := sheet.Properties.Title
		if title == "Articles" || title == "articles" {
			articlesSheet = title
		}
		if title == "Providers" || title == "providers" {
			providersSheet = title
		}
	}

	// Count Substack providers
	substackCount := 0
	readRange := fmt.Sprintf("%s!A:B", providersSheet)
	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err == nil && len(resp.Values) > 0 {
		// Assuming column A has provider name, count rows with "substack"
		for i := 1; i < len(resp.Values); i++ {
			if len(resp.Values[i]) > 0 {
				provider := fmt.Sprintf("%v", resp.Values[i][0])
				if provider == "substack" || provider == "Substack" {
					substackCount++
				}
			}
		}
	}

	// Read all articles data
	readRange = fmt.Sprintf("%s!A:E", articlesSheet)
	resp, err = client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return schema.Metrics{}, fmt.Errorf("no data found in sheet")
	}

	// Parse articles: columns are date, title, link, category, read?
	metrics := schema.Metrics{
		BySource:           make(map[string]int),
		BySourceReadStatus: make(map[string][2]int),
		ByYear:             make(map[string]int),
		ByMonthOnly:        make(map[string]int),
		ByMonthAndSource:   make(map[string]map[string]int),
	}

	// Skip header row (row 0)
	for i := 1; i < len(resp.Values); i++ {
		row := resp.Values[i]
		if len(row) < 5 {
			continue // Skip incomplete rows
		}

		metrics.TotalArticles++

		// Column A: date (YYYY-MM-DD format)
		if len(row) > 0 {
			dateStr := fmt.Sprintf("%v", row[0])
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				year := t.Format("2006")
				month := t.Format("01")
				metrics.ByYear[year]++
				metrics.ByMonthOnly[month]++

				// Track by month and source
				if len(row) > 3 {
					category := normalizeSourceName(fmt.Sprintf("%v", row[3]))
					if metrics.ByMonthAndSource[month] == nil {
						metrics.ByMonthAndSource[month] = make(map[string]int)
					}
					metrics.ByMonthAndSource[month][category]++
				}
			}
		}

		// Column D: category (source)
		var category string
		if len(row) > 3 {
			category = normalizeSourceName(fmt.Sprintf("%v", row[3]))
			metrics.BySource[category]++
		}

		// Column E: read? (checkbox - TRUE/FALSE)
		isRead := false
		if len(row) > 4 {
			readStatus := fmt.Sprintf("%v", row[4])
			// Checkbox returns TRUE or FALSE (case-insensitive)
			if readStatus == "TRUE" || readStatus == "true" {
				metrics.ReadCount++
				isRead = true
			} else {
				metrics.UnreadCount++
			}
		}

		// Track read/unread by source
		if category != "" {
			status := metrics.BySourceReadStatus[category]
			if isRead {
				status[0]++ // read
			} else {
				status[1]++ // unread
			}
			metrics.BySourceReadStatus[category] = status
		}
	}

	// Calculate derived metrics
	if metrics.TotalArticles > 0 {
		metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
	}
	// Assume 36 months (3 years of data)
	metrics.AvgArticlesPerMonth = float64(metrics.TotalArticles) / 36

	// Store substack count for later use in display
	metrics.BySourceReadStatus["substack_author_count"] = [2]int{substackCount, 0}

	// Set timestamp
	metrics.LastUpdated = time.Now()

	return metrics, nil
}

func main() {
	ctx := context.Background()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, will use environment variables")
	}

	sheetID := os.Getenv("SHEET_ID")
	credentialsPath := os.Getenv("CREDENTIALS_PATH")

	if sheetID == "" {
		log.Fatal("SHEET_ID environment variable is required")
	}
	if credentialsPath == "" {
		credentialsPath = "./credentials.json"
	}

	metrics, err := fetchMetricsFromSheets(ctx, sheetID, credentialsPath)
	if err != nil {
		log.Fatalf("Failed to fetch metrics: %v", err)
	}

	// Save metrics as JSON with timestamp
	os.MkdirAll("metrics", 0755)

	metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	// Save to metrics folder with date filename (YYYY-MM-DD.json)
	dateFilename := metrics.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)
	err = os.WriteFile(metricsFilePath, metricsJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write metrics file: %v", err)
	}

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	log.Println("✅ Successfully generated metrics from Google Sheets")
}
