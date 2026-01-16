package analytics

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
	"github.com/victoriacheng15/personal-reading-analytics/cmd/internal/metrics"
)

const (
	AnalyticsTitle = "📚 Personal Reading Analytics"
)

// AnalyticsService handles the generation of the HTML analytics
type AnalyticsService struct {
	outputDir string
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(outputDir string) *AnalyticsService {
	return &AnalyticsService{outputDir: outputDir}
}

// Generate creates the analytics files from the provided metrics
func (s *AnalyticsService) Generate(m schema.Metrics) error {
	vm, err := s.prepareViewModel(m)
	if err != nil {
		return fmt.Errorf("failed to prepare view model: %w", err)
	}

	return s.render(vm)
}

func (s *AnalyticsService) prepareViewModel(m schema.Metrics) (ViewModel, error) {
	// Sort sources by count
	var sources []schema.SourceInfo
	for name, count := range m.BySource {
		readStatus := m.BySourceReadStatus[name]
		read := readStatus[0]
		unread := readStatus[1]
		readPct := 0.0
		if count > 0 {
			readPct = (float64(read) / float64(count)) * 100
		}

		authorCount := 0
		if name == "Substack" {
			authorCount = m.BySourceReadStatus["substack_author_count"][0]
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
	for year, count := range m.ByYear {
		years = append(years, schema.YearInfo{Year: year, Count: count})
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Year > years[j].Year
	})

	// Create aggregated monthly data (Jan-Dec, all years combined)
	var monthlyAggregated []schema.MonthInfo
	// shortMonthNames is defined in preparation.go (same package)

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		monthShort := shortMonthNames[month-1]

		// Get source data for this month from ByMonthAndSource (aggregated across all years)
		if monthSourceData, exists := m.ByMonthAndSource[monthStr]; exists {
			total := 0
			monthSources := make(map[string]int)

			for source, counts := range monthSourceData {
				articleCount := counts[0] + counts[1] // read + unread
				monthSources[source] = articleCount
				total += articleCount
			}

			if total > 0 {
				monthlyAggregated = append(monthlyAggregated, schema.MonthInfo{
					Name:    monthShort,
					Month:   monthStr,
					Year:    "", // No year for aggregated monthly view
					Total:   total,
					Sources: monthSources,
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
	now := time.Now()
	currentMonth := now.Format("01")

	// If the current month (from system time) has no data,
	// fall back to the latest month available in the metrics to provide
	// a better "latest snapshot" view.
	if _, exists := m.ByMonth[currentMonth]; !exists {
		for month := 12; month >= 1; month-- {
			monthStr := fmt.Sprintf("%02d", month)
			if _, exists := m.ByMonth[monthStr]; exists {
				currentMonth = monthStr
				break
			}
		}
	}

	// Calculate badges using metrics package helpers
	topReadRateSource := metrics.CalculateTopReadRateSource(m)
	mostUnreadSource := metrics.CalculateMostUnreadSource(m)
	thisMonthArticles := metrics.CalculateThisMonthArticles(m, currentMonth)

	// Prepare chart data using analytics helpers
	yearChartData := PrepareYearChartData(years)
	monthChartData := PrepareMonthChartData(monthlyAggregated, sources)

	// Prepare read/unread data for both month and source views
	readUnreadByMonthJSON := PrepareReadUnreadByMonth(m)
	readUnreadBySourceJSON := PrepareReadUnreadBySource(sources)
	readUnreadByYearJSON := PrepareReadUnreadByYear(m)
	unreadArticleAgeDistributionJSON := PrepareUnreadArticleAgeDistribution(m)
	unreadByYearJSON := PrepareUnreadByYear(m)

	// Marshal AllYears and AllSources to JSON for JavaScript
	allYearsJSON, _ := json.Marshal(allYears)
	allSourcesJSON, _ := json.Marshal(allSources)

	// Prepare key metrics
	keyMetrics := []schema.KeyMetric{
		{Title: "Total Articles", Value: fmt.Sprintf("%d", m.TotalArticles)},
		{Title: "Read Rate", Value: fmt.Sprintf("%.1f%%", m.ReadRate)},
		{Title: "Read", Value: fmt.Sprintf("%d", m.ReadCount)},
		{Title: "Unread", Value: fmt.Sprintf("%d", m.UnreadCount)},
		{Title: "Avg/Month", Value: fmt.Sprintf("%.0f", m.AvgArticlesPerMonth)},
	}

	highlightMetrics := []schema.HightlightMetric{
		{Title: "🎯 Top Read Rate Source", Value: topReadRateSource},
		{Title: "📚 Most Unread Source", Value: mostUnreadSource},
		{Title: "✅ This Month's Articles", Value: fmt.Sprintf("%d", thisMonthArticles)},
	}

	// Load evolution data
	evolutionData, err := LoadEvolutionData()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to load evolution data: %v", err)
	} else {
		// Sort events by date descending (latest first)
		sort.Slice(evolutionData.Events, func(i, j int) bool {
			return evolutionData.Events[i].Date > evolutionData.Events[j].Date
		})
	}

	return ViewModel{
		AnalyticsTitle:                   AnalyticsTitle,
		KeyMetrics:                       keyMetrics,
		HighlightMetrics:                 highlightMetrics,
		TotalArticles:                    m.TotalArticles,
		ReadCount:                        m.ReadCount,
		UnreadCount:                      m.UnreadCount,
		ReadRate:                         m.ReadRate,
		AvgArticlesPerMonth:              m.AvgArticlesPerMonth,
		LastUpdated:                      m.LastUpdated,
		Sources:                          sources,
		Months:                           monthlyAggregated,
		Years:                            years,
		AllYears:                         allYears,
		AllSources:                       allSources,
		AllYearsJSON:                     template.JS(allYearsJSON),
		AllSourcesJSON:                   template.JS(allSourcesJSON),
		YearChartLabels:                  template.JS(yearChartData.LabelsJSON),
		YearChartData:                    template.JS(yearChartData.DataJSON),
		MonthChartLabels:                 template.JS(monthChartData.LabelsJSON),
		MonthChartDatasets:               template.JS(monthChartData.DatasetsJSON),
		MonthTotalData:                   template.JS(monthChartData.TotalDataJSON),
		ReadUnreadByMonthJSON:            readUnreadByMonthJSON,
		ReadUnreadBySourceJSON:           readUnreadBySourceJSON,
		ReadUnreadByYearJSON:             readUnreadByYearJSON,
		UnreadArticleAgeDistributionJSON: unreadArticleAgeDistributionJSON,
		UnreadByYearJSON:                 unreadByYearJSON,
		TopOldestUnreadArticles:          m.TopOldestUnreadArticles,
		EvolutionData:                    evolutionData,
	}, nil
}

func (s *AnalyticsService) render(vm ViewModel) error {
	// Get templates directory
	tmplDir, err := GetTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	// Common function map
	funcMap := template.FuncMap{
		"divideFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
	}

	// Create output directory
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy CSS directory
	cssSrc := filepath.Join(tmplDir, "css")
	cssDst := filepath.Join(s.outputDir, "css")
	if err := copyDir(cssSrc, cssDst); err != nil {
		// Log warning but don't fail, in case css dir doesn't exist
		log.Printf("⚠️ Warning: Failed to copy CSS directory: %v", err)
	} else {
		log.Printf("✅ Copied CSS to %s", cssDst)
	}

	// Pages to generate
	pages := []struct {
		Filename string
		Title    string
	}{
		{"index.html", AnalyticsTitle},
		{"analytics.html", "📊 Analytics"},
		{"evolution.html", "⏳ Evolution"},
	}

	// Loop and generate each page
	for _, page := range pages {
		// Create new template instance for this page
		tmpl := template.New("").Funcs(funcMap)

		// Parse shared templates and the specific page template
		files := []string{
			filepath.Join(tmplDir, "base.html"),
			filepath.Join(tmplDir, "header.html"),
			filepath.Join(tmplDir, "footer.html"),
			filepath.Join(tmplDir, page.Filename),
		}

		// Parse files
		tmpl, err = tmpl.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse templates for %s: %w", page.Filename, err)
		}

		// Create output file
		outPath := filepath.Join(s.outputDir, page.Filename)
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", outPath, err)
		}
		defer f.Close()

		// Update PageTitle in ViewModel for this page
		vm.PageTitle = page.Title

		// Execute the template matching the filename
		err = tmpl.ExecuteTemplate(f, page.Filename, vm)
		if err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", page.Filename, err)
		}

		log.Printf("✅ Generated %s", outPath)
	}

	return nil
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
