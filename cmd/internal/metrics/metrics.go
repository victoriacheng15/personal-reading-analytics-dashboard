package metrics

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

// SheetsClient interface for dependency injection in testing
type SheetsClient interface {
	GetValues(spreadsheetID, readRange string) (*sheets.ValueRange, error)
}

// SheetsServiceClient wraps sheets.Service to implement SheetsClient
type SheetsServiceClient struct {
	service *sheets.Service
}

// GetValues retrieves values from a Google Sheet
func (s *SheetsServiceClient) GetValues(spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	return s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
}

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

	// Top oldest unread articles count
	TopUnreadArticlesCount = 3
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
func countSubstackProviders(client SheetsClient, spreadsheetID, providersSheet string) (int, error) {
	count := 0
	readRange := fmt.Sprintf("%s!A:B", providersSheet)
	resp, err := client.GetValues(spreadsheetID, readRange)
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

// calculateArticleAgeBucket determines which age bucket an article falls into
func calculateArticleAgeBucket(articleDate, referenceDate time.Time) string {
	if articleDate.After(referenceDate) {
		return "less_than_1_month" // Handle future dates just in case
	}

	daysDiff := referenceDate.Sub(articleDate).Hours() / 24
	monthsDiff := daysDiff / 30.44 // Average days per month
	yearsDiff := daysDiff / 365.25 // Average days per year

	if yearsDiff >= 1 {
		return "older_than_1year"
	} else if monthsDiff >= 6 {
		return "6_to_12_months"
	} else if monthsDiff >= 3 {
		return "3_to_6_months"
	} else if monthsDiff >= 1 {
		return "1_to_3_months"
	}
	return "less_than_1_month"
}

// updateUnreadArticleAgeDistribution updates the age distribution for unread articles
func updateUnreadArticleAgeDistribution(metrics *schema.Metrics, article *ParsedArticle, referenceDate time.Time) {
	if !article.IsRead && !article.Date.IsZero() {
		bucket := calculateArticleAgeBucket(article.Date, referenceDate)
		metrics.UnreadArticleAgeDistribution[bucket]++
	}
}

// processArticleRows processes all article rows and updates metrics
func processArticleRows(rows [][]interface{}, metrics *schema.Metrics, earliestDate, latestDate *time.Time) ([]schema.ArticleMeta, *schema.ArticleMeta) {
	var unreadArticles []schema.ArticleMeta
	var oldestUnreadArticle *schema.ArticleMeta

	// Skip header row (row 0) and process each article
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Parse the article row into structured data
		article, err := parseArticleRow(row)
		if err != nil {
			// Skip incomplete or invalid rows
			continue
		}

		metrics.TotalArticles++

		// Update metrics by date (year, month, month+source aggregates)
		updateMetricsByDate(metrics, article, earliestDate, latestDate)

		// Update source-level aggregates
		updateMetricsBySource(metrics, article.Category)

		// Update category-level aggregates
		updateMetricsByCategory(metrics, article)

		// Update read/unread counts and by-source read status
		updateMetricsReadStatus(metrics, article)

		// Track unread by month and age distribution
		if !article.IsRead {
			month := article.Date.Format("01")
			metrics.UnreadByMonth[month]++

			// Track unread by year
			year := article.Date.Format("2006")
			metrics.UnreadByYear[year]++

			// Update age distribution for unread articles
			updateUnreadArticleAgeDistribution(metrics, article, time.Now())

			// Collect unread article details
			articleDetail, _ := parseArticleRowWithDetails(row)
			if articleDetail != nil {
				unreadArticles = append(unreadArticles, *articleDetail)

				// Track oldest unread article
				if oldestUnreadArticle == nil {
					oldestUnreadArticle = articleDetail
				} else {
					// Compare dates to find oldest
					oldestDate, _ := time.Parse("2006-01-02", oldestUnreadArticle.Date)
					currentDate, _ := time.Parse("2006-01-02", articleDetail.Date)
					if currentDate.Before(oldestDate) {
						oldestUnreadArticle = articleDetail
					}
				}
			}
		}
	}

	return unreadArticles, oldestUnreadArticle
}

// calculateDerivedMetrics computes read rate and average articles per month
func calculateDerivedMetrics(metrics *schema.Metrics, earliestDate, latestDate time.Time) {
	if metrics.TotalArticles > 0 {
		metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
	}

	// Calculate average articles per month based on actual data span
	monthsSpan := 1.0
	if !earliestDate.IsZero() && !latestDate.IsZero() {
		monthsDiff := calculateMonthsDifference(earliestDate, latestDate)

		// Handle partial month for the latest month
		// If latestDate is in the current month, we calculate the fraction of the month passed
		now := time.Now()
		if latestDate.Year() == now.Year() && latestDate.Month() == now.Month() {
			daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			fraction := float64(now.Day()) / float64(daysInMonth)
			monthsSpan = float64(monthsDiff) + fraction
		} else {
			monthsSpan = float64(monthsDiff) + 1.0
		}

		log.Printf("📊 Data span: %s to %s (%.2f months)\n", earliestDate.Format("2006-01-02"), latestDate.Format("2006-01-02"), monthsSpan)
	}

	if monthsSpan > 0 {
		metrics.AvgArticlesPerMonth = float64(metrics.TotalArticles) / monthsSpan
	}
}

// populateTopArticles stores the top oldest unread articles and the oldest unread article
func populateTopArticles(metrics *schema.Metrics, unreadArticles []schema.ArticleMeta, oldestUnreadArticle *schema.ArticleMeta) {
	// Store oldest unread article
	metrics.OldestUnreadArticle = oldestUnreadArticle

	// Sort by date (oldest first) and store only top TopUnreadArticlesCount
	if len(unreadArticles) > 0 {
		// Sort articles by date (oldest first)
		sort.Slice(unreadArticles, func(i, j int) bool {
			dateI, _ := time.Parse("2006-01-02", unreadArticles[i].Date)
			dateJ, _ := time.Parse("2006-01-02", unreadArticles[j].Date)
			return dateI.Before(dateJ)
		})

		// Store top N articles
		if len(unreadArticles) > TopUnreadArticlesCount {
			metrics.TopOldestUnreadArticles = unreadArticles[:TopUnreadArticlesCount]
		} else {
			metrics.TopOldestUnreadArticles = unreadArticles
		}
	}
}

// SheetsFetcher interface abstracts sheet operations for testability
type SheetsFetcher interface {
	GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error)
	GetArticleRows(spreadsheetID, articlesSheet string) ([][]interface{}, error)
	GetProvidersSheet(spreadsheetID, providersSheet string) ([][]interface{}, error)
}

// SheetServiceFetcher implements SheetsFetcher using sheets.Service
type SheetServiceFetcher struct {
	service *sheets.Service
}

// GetSpreadsheet retrieves spreadsheet metadata
func (s *SheetServiceFetcher) GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error) {
	return s.service.Spreadsheets.Get(spreadsheetID).Do()
}

// GetArticleRows retrieves article data from the Articles sheet
func (s *SheetServiceFetcher) GetArticleRows(spreadsheetID, articlesSheet string) ([][]interface{}, error) {
	readRange := fmt.Sprintf("%s!A:E", articlesSheet)
	resp, err := s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// GetProvidersSheet retrieves provider data from the Providers sheet
func (s *SheetServiceFetcher) GetProvidersSheet(spreadsheetID, providersSheet string) ([][]interface{}, error) {
	readRange := fmt.Sprintf("%s!A:B", providersSheet)
	resp, err := s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// findSheetNames discovers Article and Provider sheet names from spreadsheet
func findSheetNames(spreadsheet *sheets.Spreadsheet) (string, string) {
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

	return articlesSheet, providersSheet
}

// getSubstackProviderCount retrieves the count of Substack providers
func getSubstackProviderCount(fetcher SheetsFetcher, spreadsheetID, providersSheet string) int {
	rows, err := fetcher.GetProvidersSheet(spreadsheetID, providersSheet)
	if err != nil {
		log.Printf("Warning: Unable to read providers sheet: %v\n", err)
		return 0
	}

	count := 0
	if len(rows) == 0 {
		return 0
	}

	// Skip header row and count Substack entries in column A
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) > ProvidersColName {
			provider := fmt.Sprintf("%v", rows[i][ProvidersColName])
			if strings.EqualFold(provider, SubstackProvider) {
				count++
			}
		}
	}

	return count
}

// FetchMetricsFromSheetsWithService retrieves and calculates metrics using an existing Sheets service client.
// This is now a thin orchestrator that delegates to smaller, testable functions.
func FetchMetricsFromSheetsWithService(ctx context.Context, client *sheets.Service, spreadsheetID string) (schema.Metrics, error) {
	fetcher := &SheetServiceFetcher{service: client}
	return fetchMetricsWithFetcher(spreadsheetID, fetcher)
}

// fetchMetricsWithFetcher performs metrics calculation with a pluggable sheet fetcher for testability
func fetchMetricsWithFetcher(spreadsheetID string, fetcher SheetsFetcher) (schema.Metrics, error) {
	// Get spreadsheet metadata to find sheet names
	spreadsheet, err := fetcher.GetSpreadsheet(spreadsheetID)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to retrieve spreadsheet: %w", err)
	}

	// Find Article and Provider sheet names
	articlesSheet, providersSheet := findSheetNames(spreadsheet)

	// Get Substack provider count
	substackCount := getSubstackProviderCount(fetcher, spreadsheetID, providersSheet)

	// Read all articles data
	articleRows, err := fetcher.GetArticleRows(spreadsheetID, articlesSheet)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	if len(articleRows) == 0 {
		return schema.Metrics{}, fmt.Errorf("no data found in sheet")
	}

	// Initialize metrics
	metrics := schema.Metrics{
		BySource:                     make(map[string]int),
		BySourceReadStatus:           make(map[string][2]int),
		ByYear:                       make(map[string]int),
		ByMonth:                      make(map[string]int),
		ByYearAndMonth:               make(map[string]map[string]int),
		ByMonthAndSource:             make(map[string]map[string][2]int),
		ByCategory:                   make(map[string][2]int),
		ByCategoryAndSource:          make(map[string]map[string][2]int),
		UnreadByMonth:                make(map[string]int),
		UnreadByCategory:             make(map[string]int),
		UnreadBySource:               make(map[string]int),
		UnreadByYear:                 make(map[string]int),
		UnreadArticleAgeDistribution: make(map[string]int),
		SourceMetadata:               make(map[string]schema.SourceMeta),
	}

	var earliestDate, latestDate time.Time

	// Process all articles
	unreadArticles, oldestUnreadArticle := processArticleRows(articleRows, &metrics, &earliestDate, &latestDate)

	// Calculate derived metrics
	calculateDerivedMetrics(&metrics, earliestDate, latestDate)

	// Populate read/unread totals
	metrics.ReadUnreadTotals = [2]int{metrics.ReadCount, metrics.UnreadCount}

	// Populate top articles
	populateTopArticles(&metrics, unreadArticles, oldestUnreadArticle)

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

// FetchMetricsFromSheets is a backward-compatible wrapper that creates a Sheets service
// and delegates to FetchMetricsFromSheetsWithService.
func FetchMetricsFromSheets(ctx context.Context, spreadsheetID, credentialsPath string) (schema.Metrics, error) {
	// Create Sheets service
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to create sheets client: %w", err)
	}

	return FetchMetricsFromSheetsWithService(ctx, client, spreadsheetID)
}
