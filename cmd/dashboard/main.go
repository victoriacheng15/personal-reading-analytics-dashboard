package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
	dashboard "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal/dashboard"
)

const (
	dashboardTitle = "📚 Personal Reading Analytics"
)

var shortMonthNames = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

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

// calculateTopReadRateSource finds the source with the highest read rate
func calculateTopReadRateSource(metrics schema.Metrics) string {
	var topSource string
	var topRate float64
	for name, counts := range metrics.BySourceReadStatus {
		if name == "substack_author_count" {
			continue
		}
		total := counts[0] + counts[1]
		if total > 0 {
			rate := float64(counts[0]) / float64(total) * 100
			if rate > topRate {
				topRate = rate
				topSource = name
			}
		}
	}
	return topSource
}

// calculateMostUnreadSource finds the source with the most unread articles
func calculateMostUnreadSource(metrics schema.Metrics) string {
	var mostUnreadSource string
	var maxUnread int
	for name, unread := range metrics.UnreadBySource {
		if unread > maxUnread {
			maxUnread = unread
			mostUnreadSource = name
		}
	}
	return mostUnreadSource
}

// calculateThisMonthArticles calculates articles read this month (current month)
func calculateThisMonthArticles(metrics schema.Metrics, currentMonth string) int {
	// Sum all read articles from by_month_and_source_read_status for current month
	if monthData, exists := metrics.ByMonthAndSource[currentMonth]; exists {
		total := 0
		for _, counts := range monthData {
			total += counts[0] // read count
		}
		return total
	}
	return 0
}

// prepareReadUnreadByYear creates JSON data for read/unread yearly breakdown chart
func prepareReadUnreadByYear(metrics schema.Metrics) template.JS {
	// Get sorted years in descending order (latest first)
	var years []string
	for year := range metrics.ByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	readByYearArray := make([]int, 0)
	unreadByYearArray := make([]int, 0)

	for _, year := range years {
		yearRead := 0
		yearUnread := 0

		// Sum up read/unread from all months in this year
		if yearMonthData, exists := metrics.ByYearAndMonth[year]; exists {
			for month, count := range yearMonthData {
				yearRead += count
				// Get unread for this month (if available, otherwise calculate from total)
				if monthUnread, unreadExists := metrics.UnreadByMonth[month]; unreadExists {
					yearUnread += monthUnread
				}
			}
		}

		readByYearArray = append(readByYearArray, yearRead)
		unreadByYearArray = append(unreadByYearArray, yearUnread)
	}

	data := map[string]interface{}{
		"labels":     years,
		"readData":   readByYearArray,
		"unreadData": unreadByYearArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// prepareReadUnreadByMonth creates JSON data for read/unread monthly breakdown chart
func prepareReadUnreadByMonth(metrics schema.Metrics) template.JS {
	readByMonthArray := make([]int, 12)
	unreadByMonthArray := make([]int, 12)

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		unread := 0
		if u, exists := metrics.UnreadByMonth[monthStr]; exists {
			unread = u
		}
		// Calculate read for this month
		read := 0
		if monthData, exists := metrics.ByMonth[monthStr]; exists {
			read = monthData - unread
		}
		readByMonthArray[month-1] = read
		unreadByMonthArray[month-1] = unread
	}

	data := map[string]interface{}{
		"labels":     shortMonthNames,
		"readData":   readByMonthArray,
		"unreadData": unreadByMonthArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// prepareReadUnreadBySource creates JSON data for read/unread by source chart
func prepareReadUnreadBySource(sources []schema.SourceInfo) template.JS {
	readUnreadBySourceLabels := make([]string, 0)
	readBySourceData := make([]int, 0)
	unreadBySourceData := make([]int, 0)
	for _, source := range sources {
		readUnreadBySourceLabels = append(readUnreadBySourceLabels, source.Name)
		readBySourceData = append(readBySourceData, source.Read)
		unreadBySourceData = append(unreadBySourceData, source.Unread)
	}

	data := map[string]interface{}{
		"labels":     readUnreadBySourceLabels,
		"readData":   readBySourceData,
		"unreadData": unreadBySourceData,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// generateHTMLDashboard creates and saves the HTML dashboard file
func generateHTMLDashboard(metrics schema.Metrics) error {
	// Sort sources by count
	var sources []schema.SourceInfo
	for name, count := range metrics.BySource {
		readStatus := metrics.BySourceReadStatus[name]
		read := readStatus[0]
		unread := readStatus[1]
		readPct := 0.0
		if count > 0 {
			readPct = (float64(read) / float64(count)) * 100
		}

		authorCount := 0
		if name == "Substack" {
			authorCount = metrics.BySourceReadStatus["substack_author_count"][0]
		}

		sources = append(sources, schema.SourceInfo{
			Name:        name,
			Count:       count,
			Read:        read,
			Unread:      unread,
			ReadPct:     readPct,
			AuthorCount: authorCount,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	// Build year info
	var years []schema.YearInfo
	for year, count := range metrics.ByYear {
		years = append(years, schema.YearInfo{Year: year, Count: count})
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Year > years[j].Year
	})

	// Create aggregated monthly data (Jan-Dec, all years combined)
	var monthlyAggregated []schema.MonthInfo

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		monthShort := shortMonthNames[month-1]

		// Get source data for this month from ByMonthAndSource (aggregated across all years)
		if monthSourceData, exists := metrics.ByMonthAndSource[monthStr]; exists {
			total := 0
			sources := make(map[string]int)

			for source, counts := range monthSourceData {
				articleCount := counts[0] + counts[1] // read + unread
				sources[source] = articleCount
				total += articleCount
			}

			if total > 0 {
				monthlyAggregated = append(monthlyAggregated, schema.MonthInfo{
					Name:    monthShort,
					Month:   monthStr,
					Year:    "", // No year for aggregated monthly view
					Total:   total,
					Sources: sources,
				})
			}
		}
	}

	// Extract all unique years for filtering
	var allYears []string
	for _, year := range years {
		allYears = append(allYears, year.Year)
	}

	// Extract all unique sources for filtering
	var allSources []string
	for _, source := range sources {
		allSources = append(allSources, source.Name)
	}

	// Determine current month (MM format) for badge calculation
	// Using the latest month from ByMonth as current month
	currentMonth := "11" // Default to November
	for month := 12; month >= 1; month-- {
		monthStr := fmt.Sprintf("%02d", month)
		if _, exists := metrics.ByMonth[monthStr]; exists {
			currentMonth = monthStr
			break
		}
	}

	// Calculate badges using new aggregates
	topReadRateSource := calculateTopReadRateSource(metrics)
	mostUnreadSource := calculateMostUnreadSource(metrics)
	thisMonthArticles := calculateThisMonthArticles(metrics, currentMonth)

	// Load HTML template from file
	templateContent, err := dashboard.LoadTemplateContent()
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Parse and execute template
	funcMap := template.FuncMap{
		"divideFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
	}

	tmpl := template.New("dashboard").Funcs(funcMap)

	tmpl, err = tmpl.Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Create site directory
	os.MkdirAll("site", 0755)

	// Create output file
	file, err := os.Create("site/index.html")
	if err != nil {
		return fmt.Errorf("failed to create site/index.html: %w", err)
	}
	defer file.Close()

	// Prepare chart data using dashboard helpers
	yearChartData := dashboard.PrepareYearChartData(years)
	monthChartData := dashboard.PrepareMonthChartData(monthlyAggregated, sources)

	// Prepare read/unread data for both month and source views using helper functions
	readUnreadByMonthJSON := prepareReadUnreadByMonth(metrics)
	readUnreadBySourceJSON := prepareReadUnreadBySource(sources)
	readUnreadByYearJSON := prepareReadUnreadByYear(metrics)

	// Marshal AllYears and AllSources to JSON for JavaScript
	allYearsJSON, _ := json.Marshal(allYears)
	allSourcesJSON, _ := json.Marshal(allSources)

	// Prepare key metrics (formatted strings) for template loop
	keyMetrics := []schema.KeyMetric{
		{Title: "Total Articles", Value: fmt.Sprintf("%d", metrics.TotalArticles)},
		{Title: "Read Rate", Value: fmt.Sprintf("%.1f%%", metrics.ReadRate)},
		{Title: "Read", Value: fmt.Sprintf("%d", metrics.ReadCount)},
		{Title: "Unread", Value: fmt.Sprintf("%d", metrics.UnreadCount)},
		{Title: "Avg/Month", Value: fmt.Sprintf("%.0f", metrics.AvgArticlesPerMonth)},
	}

	highlightMetrics := []schema.HightlightMetric{
		{Title: "🎯 Top Read Rate Source", Value: topReadRateSource},
		{Title: "📚 Most Unread Source", Value: mostUnreadSource},
		{Title: "✅ This Month's Articles", Value: fmt.Sprintf("%d", thisMonthArticles)},
	}

	// Execute template
	data := map[string]interface{}{
		"DashboardTitle":         dashboardTitle,
		"KeyMetrics":             keyMetrics,
		"highlightMetrics":       highlightMetrics,
		"TotalArticles":          metrics.TotalArticles,
		"ReadCount":              metrics.ReadCount,
		"UnreadCount":            metrics.UnreadCount,
		"ReadRate":               metrics.ReadRate,
		"AvgArticlesPerMonth":    metrics.AvgArticlesPerMonth,
		"LastUpdated":            metrics.LastUpdated,
		"Sources":                sources,
		"Months":                 monthlyAggregated,
		"Years":                  years,
		"AllYears":               allYears,
		"AllSources":             allSources,
		"AllYearsJSON":           template.JS(allYearsJSON),
		"AllSourcesJSON":         template.JS(allSourcesJSON),
		"YearChartLabels":        template.JS(yearChartData.LabelsJSON),
		"YearChartData":          template.JS(yearChartData.DataJSON),
		"MonthChartLabels":       template.JS(monthChartData.LabelsJSON),
		"MonthChartDatasets":     template.JS(monthChartData.DatasetsJSON),
		"MonthTotalData":         template.JS(monthChartData.TotalDataJSON),
		"ReadUnreadByMonthJSON":  template.JS(readUnreadByMonthJSON),
		"ReadUnreadBySourceJSON": template.JS(readUnreadBySourceJSON),
		"ReadUnreadByYearJSON":   template.JS(readUnreadByYearJSON),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Printf("❌ Template execution error: %v\n", err)
		log.Printf("Error type: %T\n", err)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Println("✅ HTML dashboard generated at site/index.html")
	return nil
}

func main() {
	// Load latest metrics from metrics/ folder
	metrics, err := loadLatestMetrics()
	if err != nil {
		log.Fatalf("Failed to load metrics: %v", err)
	}

	// Generate HTML dashboard
	if err := generateHTMLDashboard(metrics); err != nil {
		log.Fatalf("failed to generate dashboard: %v", err)
	}

	log.Println("✅ Successfully generated dashboard from metrics")
}
