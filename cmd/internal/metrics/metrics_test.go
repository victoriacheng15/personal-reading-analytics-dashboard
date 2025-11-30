package metrics

import (
	"testing"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// Test NormalizeSourceName
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

// Test calculateMonthsDifference
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

// Test parseArticleRow
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

// Test parseArticleRowWithDetails
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

// Test updateMetricsByDate
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

// Test updateMetricsBySource
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

// Test updateMetricsByCategory
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

// Test updateMetricsReadStatus
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
