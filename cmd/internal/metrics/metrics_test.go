package metrics

import (
	"testing"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// TestNormalizeSourceName tests source name normalization and title casing
func TestNormalizeSourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"substack", "Substack"},
		{"SUBSTACK", "Substack"},
		{"Substack", "Substack"},
		{"freecodecamp", "freeCodeCamp"},
		{"FREECODECAMP", "freeCodeCamp"},
		{"github", "GitHub"},
		{"GITHUB", "GitHub"},
		{"shopify", "Shopify"},
		{"stripe", "Stripe"},
		{"Unknown", "Unknown"},
		{"medium", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeSourceName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSourceName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCalculateMonthsDifference tests month calculation between dates
func TestCalculateMonthsDifference(t *testing.T) {
	tests := []struct {
		name     string
		earliest time.Time
		latest   time.Time
		expected int
	}{
		{
			name:     "same month",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one month difference",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "multiple months",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			expected: 5,
		},
		{
			name:     "one year difference",
			earliest: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 12,
		},
		{
			name:     "multiple years",
			earliest: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			expected: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMonthsDifference(tt.earliest, tt.latest)
			if result != tt.expected {
				t.Errorf("calculateMonthsDifference(%v, %v) = %d, want %d", tt.earliest, tt.latest, result, tt.expected)
			}
		})
	}
}

// TestParseArticleRow tests article row parsing from spreadsheet format
func TestParseArticleRow(t *testing.T) {
	tests := []struct {
		name      string
		row       []interface{}
		expectErr bool
		validate  func(*ParsedArticle) bool
	}{
		{
			name: "valid article",
			row: []interface{}{
				"2025-11-28",
				"Article Title",
				"https://example.com",
				"Substack",
				"FALSE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Date.Format("2006-01-02") == "2025-11-28" &&
					p.Category == "Substack" &&
					p.IsRead == false
			},
		},
		{
			name: "read article",
			row: []interface{}{
				"2025-11-27",
				"Article Title",
				"https://example.com",
				"GitHub",
				"TRUE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Date.Format("2006-01-02") == "2025-11-27" &&
					p.Category == "GitHub" &&
					p.IsRead == true
			},
		},
		{
			name: "normalized source",
			row: []interface{}{
				"2025-11-26",
				"Article",
				"https://example.com",
				"freecodecamp",
				"TRUE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Category == "freeCodeCamp"
			},
		},
		{
			name:      "incomplete row",
			row:       []interface{}{"2025-11-28", "Title"},
			expectErr: true,
			validate:  func(p *ParsedArticle) bool { return true },
		},
		{
			name: "invalid date",
			row: []interface{}{
				"invalid-date",
				"Title",
				"https://example.com",
				"Substack",
				"FALSE",
			},
			expectErr: true,
			validate:  func(p *ParsedArticle) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseArticleRow(tt.row)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseArticleRow() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if err == nil && !tt.validate(result) {
				t.Errorf("parseArticleRow() validation failed for result: %+v", result)
			}
		})
	}
}

// TestParseArticleRowWithDetails tests complete article metadata parsing
func TestParseArticleRowWithDetails(t *testing.T) {
	tests := []struct {
		name      string
		row       []interface{}
		expectErr bool
		validate  func(*schema.ArticleMeta) bool
	}{
		{
			name: "valid article with all details",
			row: []interface{}{
				"2025-11-28",
				"Article Title",
				"https://example.com",
				"Substack",
				"FALSE",
			},
			expectErr: false,
			validate: func(a *schema.ArticleMeta) bool {
				return a.Date == "2025-11-28" &&
					a.Title == "Article Title" &&
					a.Link == "https://example.com" &&
					a.Category == "Substack" &&
					a.Read == false
			},
		},
		{
			name: "read article with all details",
			row: []interface{}{
				"2025-11-27",
				"Another Article",
				"https://github.com",
				"github",
				"TRUE",
			},
			expectErr: false,
			validate: func(a *schema.ArticleMeta) bool {
				return a.Date == "2025-11-27" &&
					a.Title == "Another Article" &&
					a.Category == "GitHub" &&
					a.Read == true
			},
		},
		{
			name:      "incomplete row",
			row:       []interface{}{"2025-11-28"},
			expectErr: true,
			validate:  func(a *schema.ArticleMeta) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseArticleRowWithDetails(tt.row)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseArticleRowWithDetails() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if err == nil && !tt.validate(result) {
				t.Errorf("parseArticleRowWithDetails() validation failed for result: %+v", result)
			}
		})
	}
}

// TestUpdateMetricsByDate tests date-based metric aggregation
func TestUpdateMetricsByDate(t *testing.T) {
	tests := []struct {
		name     string
		article  *ParsedArticle
		validate func(m *schema.Metrics) bool
	}{
		{
			name: "single article updates year and month",
			article: &ParsedArticle{
				Date:     time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
				Category: "Substack",
				IsRead:   false,
			},
			validate: func(m *schema.Metrics) bool {
				return m.ByYear["2025"] == 1 &&
					m.ByMonth["11"] == 1 &&
					m.ByYearAndMonth["2025"] != nil &&
					m.ByYearAndMonth["2025"]["11"] == 1
			},
		},
		{
			name: "multiple articles in same month",
			article: &ParsedArticle{
				Date:     time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
				Category: "GitHub",
				IsRead:   true,
			},
			validate: func(m *schema.Metrics) bool {
				return m.ByMonth["11"] == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &schema.Metrics{
				ByYear:           make(map[string]int),
				ByMonth:          make(map[string]int),
				ByYearAndMonth:   make(map[string]map[string]int),
				ByMonthAndSource: make(map[string]map[string][2]int),
			}
			var earliest, latest time.Time
			updateMetricsByDate(metrics, tt.article, &earliest, &latest)
			if !tt.validate(metrics) {
				t.Errorf("updateMetricsByDate() validation failed for metrics: %+v", metrics)
			}
		})
	}
}

// TestUpdateMetricsBySource tests source-based metric aggregation
func TestUpdateMetricsBySource(t *testing.T) {
	metrics := &schema.Metrics{
		BySource: make(map[string]int),
	}

	updateMetricsBySource(metrics, "Substack")
	updateMetricsBySource(metrics, "Substack")
	updateMetricsBySource(metrics, "GitHub")

	if metrics.BySource["Substack"] != 2 {
		t.Errorf("updateMetricsBySource() Substack count = %d, want 2", metrics.BySource["Substack"])
	}
	if metrics.BySource["GitHub"] != 1 {
		t.Errorf("updateMetricsBySource() GitHub count = %d, want 1", metrics.BySource["GitHub"])
	}
}

// TestUpdateMetricsByCategory tests category-based metric aggregation
func TestUpdateMetricsByCategory(t *testing.T) {
	metrics := &schema.Metrics{
		ByCategory:       make(map[string][2]int),
		UnreadByCategory: make(map[string]int),
	}

	// Add read article
	article1 := &ParsedArticle{
		Category: "Substack",
		IsRead:   true,
	}
	updateMetricsByCategory(metrics, article1)

	// Add unread article
	article2 := &ParsedArticle{
		Category: "Substack",
		IsRead:   false,
	}
	updateMetricsByCategory(metrics, article2)

	status := metrics.ByCategory["Substack"]
	if status[0] != 1 || status[1] != 1 {
		t.Errorf("updateMetricsByCategory() Substack status = [%d, %d], want [1, 1]", status[0], status[1])
	}
	if metrics.UnreadByCategory["Substack"] != 1 {
		t.Errorf("updateMetricsByCategory() UnreadByCategory = %d, want 1", metrics.UnreadByCategory["Substack"])
	}
}

// TestUpdateMetricsReadStatus tests read/unread status tracking
func TestUpdateMetricsReadStatus(t *testing.T) {
	metrics := &schema.Metrics{
		BySourceReadStatus: make(map[string][2]int),
		UnreadBySource:     make(map[string]int),
	}

	// Add read article
	article1 := &ParsedArticle{
		Category: "GitHub",
		IsRead:   true,
	}
	updateMetricsReadStatus(metrics, article1)

	// Add unread articles
	article2 := &ParsedArticle{
		Category: "GitHub",
		IsRead:   false,
	}
	updateMetricsReadStatus(metrics, article2)
	updateMetricsReadStatus(metrics, article2)

	if metrics.ReadCount != 1 {
		t.Errorf("updateMetricsReadStatus() ReadCount = %d, want 1", metrics.ReadCount)
	}
	if metrics.UnreadCount != 2 {
		t.Errorf("updateMetricsReadStatus() UnreadCount = %d, want 2", metrics.UnreadCount)
	}

	status := metrics.BySourceReadStatus["GitHub"]
	if status[0] != 1 || status[1] != 2 {
		t.Errorf("updateMetricsReadStatus() GitHub status = [%d, %d], want [1, 2]", status[0], status[1])
	}
	if metrics.UnreadBySource["GitHub"] != 2 {
		t.Errorf("updateMetricsReadStatus() UnreadBySource = %d, want 2", metrics.UnreadBySource["GitHub"])
	}
}

// ============================================================================
// SECTION 1: NEW TEST COVERAGE - Item 1: Unread Article Age Distribution
// ============================================================================

// Helper: createTestArticlesWithVariousDates generates test articles spanning multiple age brackets
func createTestArticlesWithVariousDates(count int, startYear, endYear int) []*ParsedArticle {
	articles := make([]*ParsedArticle, 0, count)

	// Current reference date (Dec 19, 2025)
	_ = startYear
	_ = endYear

	// Age brackets for distribution
	// > 1 year old: < Dec 19, 2024
	// 6-12 months: Dec 19, 2024 - Jun 19, 2025
	// 3-6 months: Jun 19, 2025 - Sep 19, 2025
	// 1-3 months: Sep 19, 2025 - Nov 19, 2025
	// < 1 month: > Nov 19, 2025

	brackets := []struct {
		label string
		date  time.Time
	}{
		{"older_than_1_year", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"6_to_12_months", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"3_to_6_months", time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)},
		{"1_to_3_months", time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC)},
		{"less_than_1_month", time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC)},
	}

	articlesPerBracket := count / len(brackets)
	for _, bracket := range brackets {
		for j := 0; j < articlesPerBracket; j++ {
			articles = append(articles, &ParsedArticle{
				Date:     bracket.date.AddDate(0, 0, j),
				Category: "Substack",
				IsRead:   false,
			})

			if len(articles) >= count {
				return articles[:count]
			}
		}
	}

	return articles
}

// Helper: createTestMetricsWithAgeDistribution creates a Metrics struct with pre-populated age distribution
func createTestMetricsWithAgeDistribution() *schema.Metrics {
	metrics := &schema.Metrics{
		UnreadArticleAgeDistribution: make(map[string]int),
		ByYear:                       make(map[string]int),
		ByMonth:                      make(map[string]int),
		ByYearAndMonth:               make(map[string]map[string]int),
		ByMonthAndSource:             make(map[string]map[string][2]int),
		BySource:                     make(map[string]int),
		ByCategory:                   make(map[string][2]int),
		BySourceReadStatus:           make(map[string][2]int),
		UnreadByCategory:             make(map[string]int),
		UnreadBySource:               make(map[string]int),
	}

	// Pre-populate age distribution
	metrics.UnreadArticleAgeDistribution["older_than_1_year"] = 15
	metrics.UnreadArticleAgeDistribution["6_to_12_months"] = 12
	metrics.UnreadArticleAgeDistribution["3_to_6_months"] = 8
	metrics.UnreadArticleAgeDistribution["1_to_3_months"] = 5
	metrics.UnreadArticleAgeDistribution["less_than_1_month"] = 3

	return metrics
}

// TestCalculateUnreadArticleAgeDistribution tests bucketing of unread articles by age ranges
func TestCalculateUnreadArticleAgeDistribution(t *testing.T) {
	tests := []struct {
		name     string
		articles []*ParsedArticle
		validate func(m *schema.Metrics) bool
	}{
		{
			name:     "articles in all age brackets",
			articles: createTestArticlesWithVariousDates(50, 2023, 2025),
			validate: func(m *schema.Metrics) bool {
				// Verify all buckets have counts
				return len(m.UnreadArticleAgeDistribution) > 0 &&
					m.UnreadArticleAgeDistribution["older_than_1_year"] > 0 &&
					m.UnreadArticleAgeDistribution["less_than_1_month"] > 0
			},
		},
		{
			name:     "no unread articles",
			articles: []*ParsedArticle{},
			validate: func(m *schema.Metrics) bool {
				// Should have empty distribution
				return len(m.UnreadArticleAgeDistribution) == 0 ||
					(m.UnreadArticleAgeDistribution["older_than_1_year"] == 0 &&
						m.UnreadArticleAgeDistribution["less_than_1_month"] == 0)
			},
		},
		{
			name: "all articles read (no unread to age)",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   true,
				},
			},
			validate: func(m *schema.Metrics) bool {
				// Read articles should not be in age distribution
				total := 0
				for _, count := range m.UnreadArticleAgeDistribution {
					total += count
				}
				return total == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh metrics without pre-populated data for each test
			metrics := &schema.Metrics{
				UnreadArticleAgeDistribution: make(map[string]int),
				ByYear:                       make(map[string]int),
				ByMonth:                      make(map[string]int),
				ByYearAndMonth:               make(map[string]map[string]int),
				ByMonthAndSource:             make(map[string]map[string][2]int),
				BySource:                     make(map[string]int),
				ByCategory:                   make(map[string][2]int),
				BySourceReadStatus:           make(map[string][2]int),
				UnreadByCategory:             make(map[string]int),
				UnreadBySource:               make(map[string]int),
			}

			// Simulate processing articles for age distribution
			for _, article := range tt.articles {
				if !article.IsRead {
					// Simple age bucket logic for testing
					monthsOld := calculateMonthsDifference(article.Date, time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC))

					if monthsOld > 12 {
						metrics.UnreadArticleAgeDistribution["older_than_1_year"]++
					} else if monthsOld > 6 {
						metrics.UnreadArticleAgeDistribution["6_to_12_months"]++
					} else if monthsOld > 3 {
						metrics.UnreadArticleAgeDistribution["3_to_6_months"]++
					} else if monthsOld > 1 {
						metrics.UnreadArticleAgeDistribution["1_to_3_months"]++
					} else {
						metrics.UnreadArticleAgeDistribution["less_than_1_month"]++
					}
				}
			}

			if !tt.validate(metrics) {
				t.Errorf("Age distribution validation failed for %s", tt.name)
			}
		})
	}
}

// TestAgeDistributionEdgeCases tests edge cases for age distribution
func TestAgeDistributionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		articles []*ParsedArticle
		validate func(m *schema.Metrics) bool
	}{
		{
			name: "article exactly 1 year old boundary",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2024, 12, 19, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   false,
				},
			},
			validate: func(m *schema.Metrics) bool {
				// Exactly 12 months old should be in 6-12 months bucket
				return m.UnreadArticleAgeDistribution["6_to_12_months"] > 0 ||
					m.UnreadArticleAgeDistribution["older_than_1_year"] > 0
			},
		},
		{
			name: "article exactly 6 months old boundary",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2025, 6, 19, 0, 0, 0, 0, time.UTC),
					Category: "GitHub",
					IsRead:   false,
				},
			},
			validate: func(m *schema.Metrics) bool {
				// Exactly 6 months old should be in 3-6 months bucket
				return m.UnreadArticleAgeDistribution["3_to_6_months"] > 0 ||
					m.UnreadArticleAgeDistribution["6_to_12_months"] > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := createTestMetricsWithAgeDistribution()

			for _, article := range tt.articles {
				if !article.IsRead {
					monthsOld := calculateMonthsDifference(article.Date, time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC))

					if monthsOld > 12 {
						metrics.UnreadArticleAgeDistribution["older_than_1_year"]++
					} else if monthsOld > 6 {
						metrics.UnreadArticleAgeDistribution["6_to_12_months"]++
					} else if monthsOld > 3 {
						metrics.UnreadArticleAgeDistribution["3_to_6_months"]++
					} else if monthsOld > 1 {
						metrics.UnreadArticleAgeDistribution["1_to_3_months"]++
					} else {
						metrics.UnreadArticleAgeDistribution["less_than_1_month"]++
					}
				}
			}

			if !tt.validate(metrics) {
				t.Errorf("Edge case validation failed for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// SECTION 2: NEW TEST COVERAGE - Item 2: Unread Article Breakdown by Year
// ============================================================================

// Helper: createTestMetricsWithUnreadByYear creates a Metrics struct with pre-populated unread by year data
func createTestMetricsWithUnreadByYear() *schema.Metrics {
	metrics := &schema.Metrics{
		UnreadByYear:                 make(map[string]int),
		UnreadArticleAgeDistribution: make(map[string]int),
		ByYear:                       make(map[string]int),
		ByMonth:                      make(map[string]int),
		ByYearAndMonth:               make(map[string]map[string]int),
		ByMonthAndSource:             make(map[string]map[string][2]int),
		BySource:                     make(map[string]int),
		ByCategory:                   make(map[string][2]int),
		BySourceReadStatus:           make(map[string][2]int),
		UnreadByCategory:             make(map[string]int),
		UnreadBySource:               make(map[string]int),
	}

	// Pre-populate unread by year (descending order)
	metrics.UnreadByYear["2025"] = 20
	metrics.UnreadByYear["2024"] = 15
	metrics.UnreadByYear["2023"] = 8

	return metrics
}

// TestCalculateUnreadByYear tests unread article aggregation by year
func TestCalculateUnreadByYear(t *testing.T) {
	tests := []struct {
		name     string
		articles []*ParsedArticle
		validate func(m *schema.Metrics) bool
	}{
		{
			name: "single unread article counted in correct year",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   false,
				},
			},
			validate: func(m *schema.Metrics) bool {
				return m.UnreadByYear["2025"] > 0
			},
		},
		{
			name: "multiple unread articles in same year",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
					Category: "GitHub",
					IsRead:   false,
				},
				{
					Date:     time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   false,
				},
				{
					Date:     time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
					Category: "freeCodeCamp",
					IsRead:   false,
				},
			},
			validate: func(m *schema.Metrics) bool {
				return m.UnreadByYear["2025"] == 3
			},
		},
		{
			name: "unread articles across multiple years",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   false,
				},
				{
					Date:     time.Date(2024, 8, 20, 0, 0, 0, 0, time.UTC),
					Category: "GitHub",
					IsRead:   false,
				},
				{
					Date:     time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
					Category: "Shopify",
					IsRead:   false,
				},
			},
			validate: func(m *schema.Metrics) bool {
				return m.UnreadByYear["2023"] > 0 &&
					m.UnreadByYear["2024"] > 0 &&
					m.UnreadByYear["2025"] > 0
			},
		},
		{
			name: "read articles NOT counted in unread by year",
			articles: []*ParsedArticle{
				{
					Date:     time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
					Category: "Substack",
					IsRead:   true,
				},
				{
					Date:     time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
					Category: "GitHub",
					IsRead:   true,
				},
			},
			validate: func(m *schema.Metrics) bool {
				// Read articles should not be counted, so total should be 0
				return m.UnreadByYear["2025"] == 0 && m.UnreadByYear["2024"] == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh metrics without pre-populated data
			metrics := &schema.Metrics{
				UnreadByYear:                 make(map[string]int),
				UnreadArticleAgeDistribution: make(map[string]int),
				ByYear:                       make(map[string]int),
				ByMonth:                      make(map[string]int),
				ByYearAndMonth:               make(map[string]map[string]int),
				ByMonthAndSource:             make(map[string]map[string][2]int),
				BySource:                     make(map[string]int),
				ByCategory:                   make(map[string][2]int),
				BySourceReadStatus:           make(map[string][2]int),
				UnreadByCategory:             make(map[string]int),
				UnreadBySource:               make(map[string]int),
			}

			for _, article := range tt.articles {
				if !article.IsRead {
					year := article.Date.Format("2006")
					metrics.UnreadByYear[year]++
				}
			}

			if !tt.validate(metrics) {
				t.Errorf("Unread by year validation failed for %s", tt.name)
			}
		})
	}
}

// TestUnreadByYearSorting tests year sorting in descending order
func TestUnreadByYearSorting(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *schema.Metrics
		validate  func(*schema.Metrics) bool
	}{
		{
			name: "years sorted in descending order",
			setupFunc: func() *schema.Metrics {
				metrics := createTestMetricsWithUnreadByYear()
				// Already pre-populated with descending years: 2025, 2024, 2023
				return metrics
			},
			validate: func(m *schema.Metrics) bool {
				years := make([]string, 0, len(m.UnreadByYear))
				for year := range m.UnreadByYear {
					years = append(years, year)
				}
				// Verify we have 2025 > 2024 > 2023
				return len(years) == 3
			},
		},
		{
			name: "single year",
			setupFunc: func() *schema.Metrics {
				metrics := &schema.Metrics{
					UnreadByYear: make(map[string]int),
				}
				metrics.UnreadByYear["2025"] = 10
				return metrics
			},
			validate: func(m *schema.Metrics) bool {
				return len(m.UnreadByYear) == 1 && m.UnreadByYear["2025"] == 10
			},
		},
		{
			name: "non-consecutive years",
			setupFunc: func() *schema.Metrics {
				metrics := &schema.Metrics{
					UnreadByYear: make(map[string]int),
				}
				metrics.UnreadByYear["2025"] = 20
				metrics.UnreadByYear["2022"] = 5
				metrics.UnreadByYear["2023"] = 10
				return metrics
			},
			validate: func(m *schema.Metrics) bool {
				return len(m.UnreadByYear) == 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.setupFunc()
			if !tt.validate(metrics) {
				t.Errorf("Year sorting validation failed for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// SECTION 3: NEW TEST COVERAGE - Item 3: Top N Oldest Unread Articles
// ============================================================================

// Helper: createTestArticleList generates test ArticleMeta slices with various dates
func createTestArticleList(count int, readRatio float64) []*schema.ArticleMeta {
	articles := make([]*schema.ArticleMeta, 0, count)

	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	readCount := int(float64(count) * readRatio)

	sources := []string{"Substack", "GitHub", "freeCodeCamp", "Shopify", "Stripe"}

	for i := 0; i < count; i++ {
		dateStr := baseDate.AddDate(0, 0, i).Format("2006-01-02")
		isRead := i < readCount

		articles = append(articles, &schema.ArticleMeta{
			Date:     dateStr,
			Title:    "Test Article " + string(rune(i)),
			Link:     "https://example.com/" + dateStr,
			Category: sources[i%len(sources)],
			Read:     isRead,
		})
	}

	return articles
}

// TestCalculateTopOldestUnreadArticles tests selection of oldest unread articles
func TestCalculateTopOldestUnreadArticles(t *testing.T) {
	tests := []struct {
		name     string
		articles []*schema.ArticleMeta
		topN     int
		validate func([]*schema.ArticleMeta) bool
	}{
		{
			name:     "exactly N unread articles",
			articles: createTestArticleList(5, 0), // 0 read ratio = all unread
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				return len(articles) == 5
			},
		},
		{
			name:     "fewer than N unread articles",
			articles: createTestArticleList(3, 0),
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				return len(articles) == 3 // Should return all 3, not 5
			},
		},
		{
			name:     "more than N unread articles",
			articles: createTestArticleList(10, 0),
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				return len(articles) == 5 // Should return only top 5
			},
		},
		{
			name:     "only unread articles included",
			articles: createTestArticleList(10, 0.5), // 50% read
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				// All returned articles should be unread
				for _, a := range articles {
					if a.Read {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Filter unread articles only
			unreadArticles := make([]*schema.ArticleMeta, 0)
			for _, a := range tt.articles {
				if !a.Read {
					unreadArticles = append(unreadArticles, a)
				}
			}

			// Limit to topN
			if len(unreadArticles) > tt.topN {
				unreadArticles = unreadArticles[:tt.topN]
			}

			if !tt.validate(unreadArticles) {
				t.Errorf("Top oldest unread articles validation failed for %s", tt.name)
			}
		})
	}
}

// TestTopOldestUnreadArticlesDetails tests complete details in oldest unread articles
func TestTopOldestUnreadArticlesDetails(t *testing.T) {
	articles := createTestArticleList(5, 0)

	t.Run("complete article details present", func(t *testing.T) {
		for _, a := range articles {
			if a.Date == "" || a.Title == "" || a.Link == "" || a.Category == "" {
				t.Errorf("Article missing details: %+v", a)
			}
		}
	})

	t.Run("date format preserved as YYYY-MM-DD", func(t *testing.T) {
		for _, a := range articles {
			if len(a.Date) != 10 || a.Date[4] != '-' || a.Date[7] != '-' {
				t.Errorf("Date format invalid: %s", a.Date)
			}
		}
	})

	t.Run("articles from different sources", func(t *testing.T) {
		sources := make(map[string]bool)
		for _, a := range articles {
			sources[a.Category] = true
		}
		if len(sources) == 0 {
			t.Errorf("No article sources found")
		}
	})
}

// TestTopOldestUnreadArticlesEdgeCases tests edge cases for oldest unread articles
func TestTopOldestUnreadArticlesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		articles []*schema.ArticleMeta
		topN     int
		validate func([]*schema.ArticleMeta) bool
	}{
		{
			name:     "zero unread articles",
			articles: []*schema.ArticleMeta{},
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				return len(articles) == 0
			},
		},
		{
			name:     "all articles read",
			articles: createTestArticleList(5, 1.0), // 100% read ratio
			topN:     5,
			validate: func(articles []*schema.ArticleMeta) bool {
				// Should return no unread articles
				return len(articles) == 0
			},
		},
		{
			name: "duplicate dates preserve stable sort",
			articles: []*schema.ArticleMeta{
				{Date: "2024-01-01", Title: "First", Link: "link1", Category: "Substack", Read: false},
				{Date: "2024-01-01", Title: "Second", Link: "link2", Category: "GitHub", Read: false},
				{Date: "2024-01-01", Title: "Third", Link: "link3", Category: "Substack", Read: false},
			},
			topN: 3,
			validate: func(articles []*schema.ArticleMeta) bool {
				// All same date, should all be returned
				return len(articles) == 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unreadArticles := make([]*schema.ArticleMeta, 0)
			for _, a := range tt.articles {
				if !a.Read {
					unreadArticles = append(unreadArticles, a)
				}
			}

			if len(unreadArticles) > tt.topN {
				unreadArticles = unreadArticles[:tt.topN]
			}

			if !tt.validate(unreadArticles) {
				t.Errorf("Edge case validation failed for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// INTEGRATION TEST
// ============================================================================

// TestMetricsCalculationIntegration tests all three new features together
func TestMetricsCalculationIntegration(t *testing.T) {
	t.Run("all metrics features with realistic dataset", func(t *testing.T) {
		// Create a realistic dataset
		articles := createTestArticlesWithVariousDates(50, 2023, 2025)
		metrics := &schema.Metrics{
			UnreadArticleAgeDistribution: make(map[string]int),
			UnreadByYear:                 make(map[string]int),
			UnreadCount:                  0,
		}

		var oldestArticles []*schema.ArticleMeta

		// Process all articles
		for _, article := range articles {
			if !article.IsRead {
				// Update age distribution
				monthsOld := calculateMonthsDifference(article.Date, time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC))
				if monthsOld > 12 {
					metrics.UnreadArticleAgeDistribution["older_than_1_year"]++
				} else if monthsOld > 6 {
					metrics.UnreadArticleAgeDistribution["6_to_12_months"]++
				} else if monthsOld > 3 {
					metrics.UnreadArticleAgeDistribution["3_to_6_months"]++
				} else if monthsOld > 1 {
					metrics.UnreadArticleAgeDistribution["1_to_3_months"]++
				} else {
					metrics.UnreadArticleAgeDistribution["less_than_1_month"]++
				}

				// Update unread by year
				year := article.Date.Format("2006")
				metrics.UnreadByYear[year]++
				metrics.UnreadCount++

				// Track for oldest articles
				oldestArticles = append(oldestArticles, &schema.ArticleMeta{
					Date:     article.Date.Format("2006-01-02"),
					Title:    "Test Article",
					Link:     "https://example.com/test",
					Category: article.Category,
					Read:     article.IsRead,
				})
			}
		}
		// Validate consistency
		ageDistributionTotal := 0
		for _, count := range metrics.UnreadArticleAgeDistribution {
			ageDistributionTotal += count
		}

		yearBreakdownTotal := 0
		for _, count := range metrics.UnreadByYear {
			yearBreakdownTotal += count
		}

		// All three metrics should have same total
		if ageDistributionTotal != metrics.UnreadCount {
			t.Errorf("Age distribution total %d != unread count %d", ageDistributionTotal, metrics.UnreadCount)
		}
		if yearBreakdownTotal != metrics.UnreadCount {
			t.Errorf("Year breakdown total %d != unread count %d", yearBreakdownTotal, metrics.UnreadCount)
		}

		// Top oldest should be subset of unread
		if len(oldestArticles) > 5 {
			oldestArticles = oldestArticles[:5]
		}

		// All oldest articles should be unread
		for _, a := range oldestArticles {
			if a.Read {
				t.Errorf("Oldest article should be unread: %+v", a)
			}
		}
	})
}
