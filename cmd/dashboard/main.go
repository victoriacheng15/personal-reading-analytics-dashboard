package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Metrics struct {
	TotalArticles       int               `json:"total_articles"`
	BySource            map[string]int    `json:"by_source"`
	BySourceReadStatus  map[string][2]int `json:"by_source_read_status"`
	ByYear              map[string]int    `json:"by_year"`
	ByMonthOnly         map[string]int    `json:"by_month"`
	ReadCount           int               `json:"read_count"`
	UnreadCount         int               `json:"unread_count"`
	ReadRate            float64           `json:"read_rate"`
	AvgArticlesPerMonth float64           `json:"avg_articles_per_month"`
	LastUpdated         time.Time         `json:"last_updated"`
}

func fetchMetricsFromSheets(ctx context.Context, spreadsheetID, credentialsPath string) (Metrics, error) {
	// Create Sheets service
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to create sheets client: %w", err)
	}

	// Get all sheets to find sheet names
	spreadsheet, err := client.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to retrieve spreadsheet: %w", err)
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

	log.Printf("Reading from sheet: %s\n", articlesSheet)

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
	if substackCount == 0 {
		substackCount = 13 // fallback to 13 if not found
	}

	log.Printf("Found %d Substack providers\n", substackCount)

	// Read all articles data
	readRange = fmt.Sprintf("%s!A:E", articlesSheet)
	resp, err = client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return Metrics{}, fmt.Errorf("no data found in sheet")
	}

	// Parse articles: columns are date, title, link, category, read?
	metrics := Metrics{
		BySource:           make(map[string]int),
		BySourceReadStatus: make(map[string][2]int),
		ByYear:             make(map[string]int),
		ByMonthOnly:        make(map[string]int),
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
			}
		}

		// Column D: category (source)
		var category string
		if len(row) > 3 {
			category = fmt.Sprintf("%v", row[3])
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

	// Display results
	fmt.Println("\n📊 Engineering Reading Analytics")
	fmt.Println("================================")
	fmt.Printf("Total Articles: %d\n", metrics.TotalArticles)
	fmt.Printf("Read: %d | Unread: %d\n", metrics.ReadCount, metrics.UnreadCount)
	fmt.Printf("Read Rate: %.1f%%\n", metrics.ReadRate)
	fmt.Printf("Avg Articles/Month: %.0f\n\n", metrics.AvgArticlesPerMonth)

	if len(metrics.BySource) > 0 {
		fmt.Println("By Source (Top 3):")
		// Sort sources by count (descending) and show top 3
		type sourceCount struct {
			name  string
			count int
		}
		var sorted []sourceCount
		for source, count := range metrics.BySource {
			sorted = append(sorted, sourceCount{source, count})
		}
		// Simple bubble sort for top 3
		for i := 0; i < len(sorted) && i < 3; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].count > sorted[i].count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		for i := 0; i < len(sorted) && i < 3; i++ {
			pct := (float64(sorted[i].count) / float64(metrics.TotalArticles)) * 100
			fmt.Printf("  • %s: %d (%.0f%%)\n", sorted[i].name, sorted[i].count, pct)
		}
	}

	if len(metrics.BySourceReadStatus) > 0 {
		fmt.Println("\nRead/Unread by Source:")
		for source, status := range metrics.BySourceReadStatus {
			if source == "substack_author_count" {
				continue // Skip metadata entry
			}
			total := status[0] + status[1]
			readPct := 0.0
			if total > 0 {
				readPct = (float64(status[0]) / float64(total)) * 100
			}
			fmt.Printf("  • %s: %d read / %d unread (%.0f%% read)\n", source, status[0], status[1], readPct)

			// Normalize Substack by fetched author count
			if source == "substack" {
				authorCount := metrics.BySourceReadStatus["substack_author_count"][0]
				avgPerAuthor := float64(total) / float64(authorCount)
				avgReadPerAuthor := float64(status[0]) / float64(authorCount)
				fmt.Printf("    (normalized by %d authors: %.0f avg/author, %.0f read/author)\n", authorCount, avgPerAuthor, avgReadPerAuthor)
			}
		}
	}

	if len(metrics.ByYear) > 0 {
		fmt.Println("\nBy Year:")
		for year, count := range metrics.ByYear {
			fmt.Printf("  • %s: %d\n", year, count)
		}
	}

	if len(metrics.ByMonthOnly) > 0 {
		fmt.Println("\nBy Month (All Years):")
		monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		for i := 1; i <= 12; i++ {
			monthStr := fmt.Sprintf("%02d", i)
			if count, exists := metrics.ByMonthOnly[monthStr]; exists {
				fmt.Printf("  • %s: %d\n", monthNames[i], count)
			}
		}
	}
	fmt.Println()

	// Save metrics as JSON with timestamp
	os.MkdirAll("metrics", 0755)
	os.MkdirAll("site", 0755)

	metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	// Save to metrics folder with date filename
	dateFilename := metrics.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)
	err = os.WriteFile(metricsFilePath, metricsJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write metrics file: %v", err)
	}

	// Also save current metrics to site for dashboard
	err = os.WriteFile("site/metrics.json", metricsJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write site/metrics.json: %v", err)
	}

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	log.Println("✅ Metrics saved to site/metrics.json")
	log.Println("✅ Successfully processed metrics from Google Sheets")
}
