package metrics

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// Constants for Google Sheets column indices
const (
	// Column indices in the Articles sheet
	ColDate     = 0 // Column A: date (YYYY-MM-DD format)
	ColTitle    = 1 // Column B: article title
	ColLink     = 2 // Column C: article link
	ColCategory = 3 // Column D: source/category
	ColRead     = 4 // Column E: read status (TRUE/FALSE)

	// Sheet names
	DefaultArticlesSheet  = "articles"
	DefaultProvidersSheet = "providers"

	// Provider sheet column indices
	ProvidersColName = 0 // Column A: provider name

	// Provider names
	SubstackProvider = "Substack"
)

// SourceMetadataMap holds the addition dates for all known sources
var SourceMetadataMap = map[string]string{
	"freeCodeCamp": "initial",
	"Substack":     "initial",
	"GitHub":       "2024-03-18",
	"Shopify":      "2025-03-05",
	"Stripe":       "2025-11-19",
}

// calculateMonthsDifference calculates the number of months between two dates
func calculateMonthsDifference(earliest, latest time.Time) int {
	years := latest.Year() - earliest.Year()
	months := int(latest.Month()) - int(earliest.Month())
	totalMonths := years*12 + months
	if totalMonths < 1 {
		totalMonths = 1 // At least 1 month to avoid division by zero
	}
	return totalMonths
}

// countSubstackProviders counts the number of Substack providers from the Providers sheet
func countSubstackProviders(client *sheets.Service, spreadsheetID, providersSheet string) (int, error) {
	count := 0
	readRange := fmt.Sprintf("%s!A:B", providersSheet)
	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		// Log error but don't fail - provider counting is optional
		log.Printf("Warning: Unable to read providers sheet: %v\n", err)
		return 0, nil
	}

	if len(resp.Values) == 0 {
		return 0, nil
	}

	// Skip header row and count Substack entries in column A
	for i := 1; i < len(resp.Values); i++ {
		if len(resp.Values[i]) > ProvidersColName {
			provider := fmt.Sprintf("%v", resp.Values[i][ProvidersColName])
			if strings.EqualFold(provider, SubstackProvider) {
				count++
			}
		}
	}

	return count, nil
}

// NormalizeSourceName converts source names to proper capitalization
func NormalizeSourceName(name string) string {
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

// ParsedArticle represents parsed data from a single article row
type ParsedArticle struct {
	Date     time.Time
	Category string // normalized source name
	IsRead   bool
}

// parseArticleRow extracts relevant data from a single article row
func parseArticleRow(row []interface{}) (*ParsedArticle, error) {
	if len(row) < ColRead+1 {
		return nil, fmt.Errorf("incomplete row: expected at least %d columns, got %d", ColRead+1, len(row))
	}

	article := &ParsedArticle{}

	// Parse date (Column A)
	if len(row) > ColDate {
		dateStr := fmt.Sprintf("%v", row[ColDate])
		parsedTime, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %s", dateStr)
		}
		article.Date = parsedTime
	}

	// Parse category/source (Column D)
	if len(row) > ColCategory {
		article.Category = NormalizeSourceName(fmt.Sprintf("%v", row[ColCategory]))
	}

	// Parse read status (Column E)
	if len(row) > ColRead {
		readStatus := fmt.Sprintf("%v", row[ColRead])
		article.IsRead = (readStatus == "TRUE" || readStatus == "true")
	}

	return article, nil
}

// parseArticleRowWithDetails extracts all details from a single article row
func parseArticleRowWithDetails(row []interface{}) (*schema.ArticleMeta, error) {
	if len(row) < ColRead+1 {
		return nil, fmt.Errorf("incomplete row: expected at least %d columns, got %d", ColRead+1, len(row))
	}

	article := &schema.ArticleMeta{}

	// Parse date (Column A)
	if len(row) > ColDate {
		article.Date = fmt.Sprintf("%v", row[ColDate])
	}

	// Parse title (Column B)
	if len(row) > ColTitle {
		article.Title = fmt.Sprintf("%v", row[ColTitle])
	}

	// Parse link (Column C)
	if len(row) > ColLink {
		article.Link = fmt.Sprintf("%v", row[ColLink])
	}

	// Parse category/source (Column D)
	if len(row) > ColCategory {
		article.Category = NormalizeSourceName(fmt.Sprintf("%v", row[ColCategory]))
	}

	// Parse read status (Column E)
	if len(row) > ColRead {
		readStatus := fmt.Sprintf("%v", row[ColRead])
		article.Read = (readStatus == "TRUE" || readStatus == "true")
	}

	return article, nil
}

// updateMetricsByDate updates yearly and monthly aggregate metrics
func updateMetricsByDate(metrics *schema.Metrics, article *ParsedArticle, earliestDate, latestDate *time.Time) {
	if article.Date.IsZero() {
		return
	}

	// Track earliest and latest dates
	if earliestDate.IsZero() || article.Date.Before(*earliestDate) {
		*earliestDate = article.Date
	}
	if latestDate.IsZero() || article.Date.After(*latestDate) {
		*latestDate = article.Date
	}

	year := article.Date.Format("2006")
	month := article.Date.Format("01")
	metrics.ByYear[year]++
	metrics.ByMonth[month]++

	// Track by year and month
	if metrics.ByYearAndMonth[year] == nil {
		metrics.ByYearAndMonth[year] = make(map[string]int)
	}
	metrics.ByYearAndMonth[year][month]++

	// Track by month and source (with read/unread counts)
	if article.Category != "" {
		if metrics.ByMonthAndSource[month] == nil {
			metrics.ByMonthAndSource[month] = make(map[string][2]int)
		}
		status := metrics.ByMonthAndSource[month][article.Category]
		if article.IsRead {
			status[0]++
		} else {
			status[1]++
		}
		metrics.ByMonthAndSource[month][article.Category] = status
	}
}

// updateMetricsBySource updates source-level aggregate metrics
func updateMetricsBySource(metrics *schema.Metrics, category string) {
	if category != "" {
		metrics.BySource[category]++
	}
}

// updateMetricsByCategory updates category-level aggregate metrics
func updateMetricsByCategory(metrics *schema.Metrics, article *ParsedArticle) {
	if article.Category != "" {
		status := metrics.ByCategory[article.Category]
		if article.IsRead {
			status[0]++
		} else {
			status[1]++
		}
		metrics.ByCategory[article.Category] = status

		// Track unread by category
		if !article.IsRead {
			metrics.UnreadByCategory[article.Category]++
		}
	}
}

// updateMetricsReadStatus updates read/unread counts and status by source
func updateMetricsReadStatus(metrics *schema.Metrics, article *ParsedArticle) {
	if article.IsRead {
		metrics.ReadCount++
	} else {
		metrics.UnreadCount++
	}

	// Track read/unread by source
	if article.Category != "" {
		status := metrics.BySourceReadStatus[article.Category]
		if article.IsRead {
			status[0]++ // read
		} else {
			status[1]++ // unread
		}
		metrics.BySourceReadStatus[article.Category] = status

		// Track unread by source
		if !article.IsRead {
			metrics.UnreadBySource[article.Category]++
		}
	}
}

// FetchMetricsFromSheets retrieves and calculates metrics from Google Sheets
func FetchMetricsFromSheets(ctx context.Context, spreadsheetID, credentialsPath string) (schema.Metrics, error) {
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
	articlesSheet := DefaultArticlesSheet
	providersSheet := DefaultProvidersSheet
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
	substackCount, err := countSubstackProviders(client, spreadsheetID, providersSheet)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to count providers: %w", err)
	}

	// Read all articles data
	readRange := fmt.Sprintf("%s!A:E", articlesSheet)
	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return schema.Metrics{}, fmt.Errorf("no data found in sheet")
	}

	// Parse articles: columns are date, title, link, category, read?
	metrics := schema.Metrics{
		BySource:            make(map[string]int),
		BySourceReadStatus:  make(map[string][2]int),
		ByYear:              make(map[string]int),
		ByMonth:             make(map[string]int),
		ByYearAndMonth:      make(map[string]map[string]int),
		ByMonthAndSource:    make(map[string]map[string][2]int),
		ByCategory:          make(map[string][2]int),
		ByCategoryAndSource: make(map[string]map[string][2]int),
		UnreadByMonth:       make(map[string]int),
		UnreadByCategory:    make(map[string]int),
		UnreadBySource:      make(map[string]int),
		SourceMetadata:      make(map[string]schema.SourceMeta),
	}

	var earliestDate, latestDate time.Time
	var oldestUnreadArticle *schema.ArticleMeta

	// Skip header row (row 0) and process each article
	for i := 1; i < len(resp.Values); i++ {
		row := resp.Values[i]

		// Parse the article row into structured data
		article, err := parseArticleRow(row)
		if err != nil {
			// Skip incomplete or invalid rows
			continue
		}

		metrics.TotalArticles++

		// Update metrics by date (year, month, month+source aggregates)
		updateMetricsByDate(&metrics, article, &earliestDate, &latestDate)

		// Update source-level aggregates
		updateMetricsBySource(&metrics, article.Category)

		// Update category-level aggregates
		updateMetricsByCategory(&metrics, article)

		// Update read/unread counts and by-source read status
		updateMetricsReadStatus(&metrics, article)

		// Track unread by month
		if !article.IsRead {
			month := article.Date.Format("01")
			metrics.UnreadByMonth[month]++

			// Track oldest unread article
			articleDetail, _ := parseArticleRowWithDetails(row)
			if articleDetail != nil && oldestUnreadArticle == nil {
				oldestUnreadArticle = articleDetail
			} else if articleDetail != nil && oldestUnreadArticle != nil {
				// Compare dates to find oldest
				oldestDate, _ := time.Parse("2006-01-02", oldestUnreadArticle.Date)
				currentDate, _ := time.Parse("2006-01-02", articleDetail.Date)
				if currentDate.Before(oldestDate) {
					oldestUnreadArticle = articleDetail
				}
			}
		}
	}

	// Calculate derived metrics
	if metrics.TotalArticles > 0 {
		metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
	}
	// Calculate average articles per month based on actual data span
	monthsSpan := 1
	if !earliestDate.IsZero() && !latestDate.IsZero() {
		monthsSpan = calculateMonthsDifference(earliestDate, latestDate) + 1 // +1 to include both start and end month
		log.Printf("📊 Data span: %s to %s (%d months)\n", earliestDate.Format("2006-01-02"), latestDate.Format("2006-01-02"), monthsSpan)
	}
	metrics.AvgArticlesPerMonth = float64(metrics.TotalArticles) / float64(monthsSpan)

	// Populate read/unread totals
	metrics.ReadUnreadTotals = [2]int{metrics.ReadCount, metrics.UnreadCount}

	// Populate oldest unread article
	if oldestUnreadArticle != nil {
		metrics.OldestUnreadArticle = oldestUnreadArticle
	}

	// Store substack count for later use in display
	metrics.BySourceReadStatus["substack_author_count"] = [2]int{substackCount, 0}

	// Populate source metadata
	for source, addedDate := range SourceMetadataMap {
		metrics.SourceMetadata[source] = schema.SourceMeta{Added: addedDate}
	}

	// Set timestamp
	metrics.LastUpdated = time.Now()

	return metrics, nil
}
