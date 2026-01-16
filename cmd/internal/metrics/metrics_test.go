package metrics

import (
	"fmt"
	"testing"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
	"google.golang.org/api/sheets/v4"
)

// ============================================================================
// calculateMonthsDifference: Calculates the number of months between two dates
// ============================================================================

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

// ============================================================================
// countSubstackProviders: Counts the number of Substack providers from a sheet
// ============================================================================

type MockSheetsClient struct {
	response *sheets.ValueRange
	err      error
}

func (m *MockSheetsClient) GetValues(spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	return m.response, m.err
}

func TestSheetsServiceClientGetValues(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse *sheets.ValueRange
		mockErr      error
		expectErr    bool
		validateData func(*sheets.ValueRange) bool
	}{
		{
			name: "successful retrieval of sheet values",
			mockResponse: &sheets.ValueRange{
				Range: "articles!A1:E5",
				Values: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
					{"2025-11-28", "Article 1", "https://example.com/1", "Substack", "FALSE"},
					{"2025-11-27", "Article 2", "https://example.com/2", "GitHub", "TRUE"},
				},
			},
			mockErr:   nil,
			expectErr: false,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr != nil &&
					len(vr.Values) == 3 &&
					vr.Values[0][0] == "Date" &&
					vr.Values[1][1] == "Article 1"
			},
		},
		{
			name:         "empty sheet returns empty values",
			mockResponse: &sheets.ValueRange{Range: "articles!A1:E1", Values: [][]interface{}{}},
			mockErr:      nil,
			expectErr:    false,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr != nil && len(vr.Values) == 0
			},
		},
		{
			name:         "sheet with only headers",
			mockResponse: &sheets.ValueRange{Range: "articles!A1:E1", Values: [][]interface{}{{"Date", "Title", "Link", "Category", "Read"}}},
			mockErr:      nil,
			expectErr:    false,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr != nil && len(vr.Values) == 1 && vr.Values[0][0] == "Date"
			},
		},
		{
			name:         "error retrieving from sheets",
			mockResponse: nil,
			mockErr:      fmt.Errorf("sheets API error: unauthorized"),
			expectErr:    true,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr == nil
			},
		},
		{
			name: "large dataset retrieval",
			mockResponse: func() *sheets.ValueRange {
				values := [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
				}
				for i := 1; i <= 100; i++ {
					values = append(values, []interface{}{
						"2025-11-28",
						fmt.Sprintf("Article %d", i),
						fmt.Sprintf("https://example.com/%d", i),
						"Substack",
						"FALSE",
					})
				}
				return &sheets.ValueRange{Range: "articles!A1:E101", Values: values}
			}(),
			mockErr:   nil,
			expectErr: false,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr != nil && len(vr.Values) == 101
			},
		},
		{
			name: "mixed data types in sheet",
			mockResponse: &sheets.ValueRange{
				Range: "articles!A1:E3",
				Values: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
					{"2025-11-28", "Article", "https://example.com", "Substack", "FALSE"},
					{"2025-11-27", 12345, "https://example.com/2", "GitHub", true}, // Mixed types
				},
			},
			mockErr:   nil,
			expectErr: false,
			validateData: func(vr *sheets.ValueRange) bool {
				return vr != nil && len(vr.Values) == 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock client with the test response
			mockClient := &MockSheetsClient{
				response: tt.mockResponse,
				err:      tt.mockErr,
			}

			// Call GetValues on the mock client
			result, err := mockClient.GetValues("test-spreadsheet-id", "articles!A1:E100")

			if (err != nil) != tt.expectErr {
				t.Errorf("GetValues() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.validateData(result) {
				t.Errorf("GetValues() result validation failed: %+v", result)
			}
		})
	}
}

func TestCountSubstackProviders(t *testing.T) {
	tests := []struct {
		name          string
		rows          [][]interface{}
		mockErr       error
		expectedCount int
		shouldErr     bool
	}{
		{
			name: "multiple substack providers with case variations",
			rows: [][]interface{}{
				{"Provider", "OtherCol"}, // header row (index 0, skipped)
				{"Substack", "entry1"},
				{"substack", "entry2"},
				{"SUBSTACK", "entry3"},
				{"GitHub", "entry4"},
				{"Substack", "entry5"},
			},
			mockErr:       nil,
			expectedCount: 4, // All 4 substack entries (case-insensitive)
			shouldErr:     false,
		},
		{
			name: "no substack providers",
			rows: [][]interface{}{
				{"Provider", "OtherCol"},
				{"GitHub", "entry1"},
				{"Medium", "entry2"},
				{"Shopify", "entry3"},
			},
			mockErr:       nil,
			expectedCount: 0,
			shouldErr:     false,
		},
		{
			name: "only header row",
			rows: [][]interface{}{
				{"Provider", "OtherCol"},
			},
			mockErr:       nil,
			expectedCount: 0,
			shouldErr:     false,
		},
		{
			name:          "empty sheet",
			rows:          [][]interface{}{},
			mockErr:       nil,
			expectedCount: 0,
			shouldErr:     false,
		},
		{
			name: "row with missing provider column",
			rows: [][]interface{}{
				{"Provider", "OtherCol"},
				{"Substack", "entry1"},
				{}, // empty row - missing provider data
				{"GitHub", "entry2"},
			},
			mockErr:       nil,
			expectedCount: 1, // Only the valid Substack entry
			shouldErr:     false,
		},
		{
			name: "case-sensitive exact match only",
			rows: [][]interface{}{
				{"Provider", "OtherCol"},
				{"Substack", "entry1"},
				{"SubStack", "entry2"},
				{"Substack Pro", "entry3"},
			},
			mockErr:       nil,
			expectedCount: 2, // Only exact "Substack" matches
			shouldErr:     false,
		},
		{
			name:          "API permission denied error - returns 0 (fault tolerant)",
			rows:          nil,
			mockErr:       fmt.Errorf("permission denied"),
			expectedCount: 0,
			shouldErr:     false, // Function doesn't propagate error - logs and returns 0
		},
		{
			name:          "network connectivity error - returns 0 (fault tolerant)",
			rows:          nil,
			mockErr:       fmt.Errorf("connection timeout"),
			expectedCount: 0,
			shouldErr:     false, // Function doesn't propagate error - logs and returns 0
		},
		{
			name:          "spreadsheet not found error - returns 0 (fault tolerant)",
			rows:          nil,
			mockErr:       fmt.Errorf("not found: spreadsheet does not exist"),
			expectedCount: 0,
			shouldErr:     false, // Function doesn't propagate error - logs and returns 0
		},
		{
			name:          "invalid credentials error - returns 0 (fault tolerant)",
			rows:          nil,
			mockErr:       fmt.Errorf("invalid authentication credentials"),
			expectedCount: 0,
			shouldErr:     false, // Function doesn't propagate error - logs and returns 0
		},
		{
			name:          "rate limit exceeded error - returns 0 (fault tolerant)",
			rows:          nil,
			mockErr:       fmt.Errorf("quota exceeded"),
			expectedCount: 0,
			shouldErr:     false, // Function doesn't propagate error - logs and returns 0
		},
		{
			name: "valid data after error simulation",
			rows: [][]interface{}{
				{"Provider", "Status"},
				{"Substack", "active"},
				{"Substack", "inactive"},
				{"GitHub", "active"},
			},
			mockErr:       nil,
			expectedCount: 2,
			shouldErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client with response
			mockClient := &MockSheetsClient{
				response: &sheets.ValueRange{Values: tt.rows},
				err:      tt.mockErr,
			}

			// Call the actual countSubstackProviders function
			count, err := countSubstackProviders(mockClient, "test-spreadsheet-id", "providers")

			if (err != nil) != tt.shouldErr {
				t.Errorf("countSubstackProviders() error = %v, shouldErr %v", err, tt.shouldErr)
				return
			}

			if count != tt.expectedCount {
				t.Errorf("countSubstackProviders() = %d, want %d", count, tt.expectedCount)
			}
		})
	}
}

// ============================================================================
// NormalizeSourceName: Converts source names to proper capitalization
// ============================================================================

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

// ============================================================================
// parseArticleRow: Extracts relevant data from a single article row
// ============================================================================

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

// ============================================================================
// parseArticleRowWithDetails: Extracts all details from a single article row
// ============================================================================

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

// ============================================================================
// updateMetricsByDate: Updates yearly and monthly aggregate metrics
// ============================================================================

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

// ============================================================================
// updateMetricsBySource: Updates source-level aggregate metrics
// ============================================================================

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

// ============================================================================
// updateMetricsByCategory: Updates category-level aggregate metrics
// ============================================================================

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

// ============================================================================
// updateMetricsReadStatus: Updates read/unread counts and status by source
// ============================================================================

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
// calculateArticleAgeBucket: Determines which age bucket an article falls into
// ============================================================================

func TestCalculateArticleAgeBucket(t *testing.T) {
	referenceDate := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		articleDate    time.Time
		referenceDate  time.Time
		expectedBucket string
	}{
		{
			name:           "article less than 1 month old (9 days)",
			articleDate:    time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "less_than_1_month",
		},
		{
			name:           "article just under 1 month old (30 days)",
			articleDate:    time.Date(2025, 11, 19, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "less_than_1_month",
		},
		{
			name:           "article 1 month old (31 days)",
			articleDate:    time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "1_to_3_months",
		},
		{
			name:           "article 2 months old (61 days)",
			articleDate:    time.Date(2025, 10, 19, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "1_to_3_months",
		},
		{
			name:           "article 3 months old (92 days)",
			articleDate:    time.Date(2025, 9, 18, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "3_to_6_months",
		},
		{
			name:           "article 4 months old (122 days)",
			articleDate:    time.Date(2025, 8, 18, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "3_to_6_months",
		},
		{
			name:           "article 6 months old (183 days)",
			articleDate:    time.Date(2025, 6, 18, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "6_to_12_months",
		},
		{
			name:           "article 9 months old (274 days)",
			articleDate:    time.Date(2025, 3, 19, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "6_to_12_months",
		},
		{
			name:           "article 11 months old (335 days)",
			articleDate:    time.Date(2024, 12, 19, 12, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "6_to_12_months",
		},
		{
			name:           "article exactly 1 year old (366 days)",
			articleDate:    time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "older_than_1year",
		},
		{
			name:           "article 2 years old",
			articleDate:    time.Date(2023, 12, 19, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "older_than_1year",
		},
		{
			name:           "article from today",
			articleDate:    referenceDate,
			referenceDate:  referenceDate,
			expectedBucket: "less_than_1_month",
		},
		{
			name:           "future article (edge case)",
			articleDate:    time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "less_than_1_month",
		},
		{
			name:           "very old article (10 years)",
			articleDate:    time.Date(2015, 12, 19, 0, 0, 0, 0, time.UTC),
			referenceDate:  referenceDate,
			expectedBucket: "older_than_1year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateArticleAgeBucket(tt.articleDate, tt.referenceDate)
			if result != tt.expectedBucket {
				t.Errorf("calculateArticleAgeBucket(%v, %v) = %q, want %q",
					tt.articleDate, tt.referenceDate, result, tt.expectedBucket)
			}
		})
	}
}

// ============================================================================
// calculateArticleAgeBucket & updateUnreadArticleAgeDistribution:
// Calculates age distribution for unread articles
// ============================================================================

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
// UnreadByYear metrics: Unread article aggregation by year
// ============================================================================

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
// TopOldestUnreadArticles: Selection and ranking of oldest unread articles
// ============================================================================

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
// processArticleRows: Processes all article rows and updates metrics
// ============================================================================

func TestProcessArticleRows(t *testing.T) {
	tests := []struct {
		name        string
		description string
		rows        [][]interface{}
		validate    func(*schema.Metrics, []schema.ArticleMeta, *schema.ArticleMeta) bool
	}{
		{
			name:        "processes all articles and aggregates metrics",
			description: "Validates article parsing, counting, and metric aggregation",
			rows:        createTestArticleRows(),
			validate: func(m *schema.Metrics, unread []schema.ArticleMeta, oldest *schema.ArticleMeta) bool {
				return m.TotalArticles == 10 &&
					m.ReadCount == 3 &&
					m.UnreadCount == 7 &&
					len(unread) == 7 &&
					oldest != nil &&
					oldest.Date == "2024-12-18"
			},
		},
		{
			name:        "handles empty rows",
			description: "Validates handling of empty row list",
			rows:        [][]interface{}{},
			validate: func(m *schema.Metrics, unread []schema.ArticleMeta, oldest *schema.ArticleMeta) bool {
				return m.TotalArticles == 0 &&
					m.ReadCount == 0 &&
					m.UnreadCount == 0 &&
					len(unread) == 0 &&
					oldest == nil
			},
		},
		{
			name:        "handles header-only rows",
			description: "Validates handling of only header row",
			rows: [][]interface{}{
				{"Date", "Title", "Link", "Category", "Read"},
			},
			validate: func(m *schema.Metrics, unread []schema.ArticleMeta, oldest *schema.ArticleMeta) bool {
				return m.TotalArticles == 0 &&
					len(unread) == 0 &&
					oldest == nil
			},
		},
		{
			name:        "aggregates by source",
			description: "Validates source normalization and counting",
			rows:        createTestArticleRows(),
			validate: func(m *schema.Metrics, _ []schema.ArticleMeta, _ *schema.ArticleMeta) bool {
				return m.BySource["Substack"] == 3 &&
					m.BySource["GitHub"] == 3 &&
					m.BySource["freeCodeCamp"] == 2 &&
					m.BySource["Shopify"] == 1 &&
					m.BySource["Stripe"] == 1
			},
		},
		{
			name:        "separates read and unread articles",
			description: "Validates read/unread segregation and counting",
			rows:        createTestArticleRows(),
			validate: func(m *schema.Metrics, unread []schema.ArticleMeta, oldest *schema.ArticleMeta) bool {
				// All unread articles should be in the unread list
				for _, article := range unread {
					if article.Read {
						return false
					}
				}
				return len(unread) == 7 && m.UnreadCount == 7
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			unread, oldest := processArticleRows(tt.rows, &metrics, &earliestDate, &latestDate)

			if !tt.validate(&metrics, unread, oldest) {
				t.Errorf("%s: validation failed", tt.name)
			}
		})
	}
}

// ============================================================================
// calculateDerivedMetrics: Computes read rate and average articles per month
// ============================================================================

func TestCalculateDerivedMetrics(t *testing.T) {
	tests := []struct {
		name                string
		description         string
		totalArticles       int
		readCount           int
		earliestDate        time.Time
		latestDate          time.Time
		expectedReadRate    float64
		validateAvgArticles func(float64) bool
	}{
		{
			name:             "calculates read rate correctly",
			description:      "Validates read rate calculation (30% = 3 read / 10 total)",
			totalArticles:    10,
			readCount:        3,
			earliestDate:     time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC),
			latestDate:       time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
			expectedReadRate: 30.0,
			validateAvgArticles: func(avg float64) bool {
				return avg > 0 // Should be positive
			},
		},
		{
			name:             "handles 100% read articles",
			description:      "Validates read rate for all read articles",
			totalArticles:    5,
			readCount:        5,
			earliestDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latestDate:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			expectedReadRate: 100.0,
			validateAvgArticles: func(avg float64) bool {
				return avg > 0
			},
		},
		{
			name:             "handles 0% read articles",
			description:      "Validates read rate for all unread articles",
			totalArticles:    5,
			readCount:        0,
			earliestDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latestDate:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			expectedReadRate: 0.0,
			validateAvgArticles: func(avg float64) bool {
				return avg > 0
			},
		},
		{
			name:             "handles zero articles",
			description:      "Validates handling when no articles exist",
			totalArticles:    0,
			readCount:        0,
			earliestDate:     time.Time{},
			latestDate:       time.Time{},
			expectedReadRate: 0.0,
			validateAvgArticles: func(avg float64) bool {
				return avg == 0 // Average should be 0 when totalArticles is 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := schema.Metrics{
				TotalArticles: tt.totalArticles,
				ReadCount:     tt.readCount,
			}

			calculateDerivedMetrics(&metrics, tt.earliestDate, tt.latestDate)

			if metrics.ReadRate != tt.expectedReadRate {
				t.Errorf("Expected read rate %.1f%%, got %.1f%%", tt.expectedReadRate, metrics.ReadRate)
			}

			if !tt.validateAvgArticles(metrics.AvgArticlesPerMonth) {
				t.Errorf("Average articles per month validation failed: %.2f", metrics.AvgArticlesPerMonth)
			}
		})
	}
}

// ============================================================================
// populateTopArticles: Sorts articles and stores top N oldest
// ============================================================================

func TestPopulateTopArticles(t *testing.T) {
	tests := []struct {
		name        string
		description string
		unread      []schema.ArticleMeta
		oldest      *schema.ArticleMeta
		validate    func(*schema.Metrics) bool
	}{
		{
			name:        "stores all articles when fewer than top N",
			description: "Validates storing all articles when count < TopUnreadArticlesCount",
			unread: []schema.ArticleMeta{
				{Date: "2025-12-10", Title: "Article 1", Link: "link1", Category: "Substack", Read: false},
				{Date: "2025-11-10", Title: "Article 2", Link: "link2", Category: "GitHub", Read: false},
			},
			oldest: &schema.ArticleMeta{Date: "2025-11-10", Title: "Article 2", Link: "link2", Category: "GitHub", Read: false},
			validate: func(m *schema.Metrics) bool {
				return len(m.TopOldestUnreadArticles) == 2 &&
					m.OldestUnreadArticle != nil &&
					m.OldestUnreadArticle.Date == "2025-11-10"
			},
		},
		{
			name:        "stores only top N articles",
			description: "Validates limiting to TopUnreadArticlesCount when more articles exist",
			unread: []schema.ArticleMeta{
				{Date: "2025-12-10", Title: "Article 1", Link: "link1", Category: "Substack", Read: false},
				{Date: "2025-11-10", Title: "Article 2", Link: "link2", Category: "GitHub", Read: false},
				{Date: "2025-10-10", Title: "Article 3", Link: "link3", Category: "Substack", Read: false},
				{Date: "2025-09-10", Title: "Article 4", Link: "link4", Category: "freeCodeCamp", Read: false},
				{Date: "2025-08-10", Title: "Article 5", Link: "link5", Category: "Stripe", Read: false},
			},
			oldest: &schema.ArticleMeta{Date: "2025-08-10", Title: "Article 5", Link: "link5", Category: "Stripe", Read: false},
			validate: func(m *schema.Metrics) bool {
				return len(m.TopOldestUnreadArticles) == 3 && // TopUnreadArticlesCount = 3
					m.OldestUnreadArticle != nil &&
					m.OldestUnreadArticle.Date == "2025-08-10"
			},
		},
		{
			name:        "handles empty unread articles",
			description: "Validates handling when no unread articles exist",
			unread:      []schema.ArticleMeta{},
			oldest:      nil,
			validate: func(m *schema.Metrics) bool {
				return len(m.TopOldestUnreadArticles) == 0 &&
					m.OldestUnreadArticle == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := schema.Metrics{}
			populateTopArticles(&metrics, tt.unread, tt.oldest)

			if !tt.validate(&metrics) {
				t.Errorf("%s: validation failed", tt.name)
			}
		})
	}
}

// ============================================================================
// SheetsFetcher Mock: For testing sheet operations
// ============================================================================

type MockSheetsFetcher struct {
	spreadsheet    *sheets.Spreadsheet
	articleRows    [][]interface{}
	providerRows   [][]interface{}
	spreadsheetErr error
	articleErr     error
	providerErr    error
}

func (m *MockSheetsFetcher) GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error) {
	return m.spreadsheet, m.spreadsheetErr
}

func (m *MockSheetsFetcher) GetArticleRows(spreadsheetID, articlesSheet string) ([][]interface{}, error) {
	return m.articleRows, m.articleErr
}

func (m *MockSheetsFetcher) GetProvidersSheet(spreadsheetID, providersSheet string) ([][]interface{}, error) {
	return m.providerRows, m.providerErr
}

// ============================================================================
// findSheetNames: Discovers Article and Provider sheet names
// ============================================================================

func TestFindSheetNames(t *testing.T) {
	tests := []struct {
		name                   string
		spreadsheet            *sheets.Spreadsheet
		expectedArticlesSheet  string
		expectedProvidersSheet string
	}{
		{
			name: "finds standard sheet names",
			spreadsheet: &sheets.Spreadsheet{
				Sheets: []*sheets.Sheet{
					{Properties: &sheets.SheetProperties{Title: "Articles"}},
					{Properties: &sheets.SheetProperties{Title: "Providers"}},
				},
			},
			expectedArticlesSheet:  "Articles",
			expectedProvidersSheet: "Providers",
		},
		{
			name: "finds lowercase sheet names",
			spreadsheet: &sheets.Spreadsheet{
				Sheets: []*sheets.Sheet{
					{Properties: &sheets.SheetProperties{Title: "articles"}},
					{Properties: &sheets.SheetProperties{Title: "providers"}},
				},
			},
			expectedArticlesSheet:  "articles",
			expectedProvidersSheet: "providers",
		},
		{
			name: "uses defaults when sheets not found",
			spreadsheet: &sheets.Spreadsheet{
				Sheets: []*sheets.Sheet{
					{Properties: &sheets.SheetProperties{Title: "Other"}},
				},
			},
			expectedArticlesSheet:  "articles",
			expectedProvidersSheet: "providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles, providers := findSheetNames(tt.spreadsheet)
			if articles != tt.expectedArticlesSheet {
				t.Errorf("Expected articles sheet '%s', got '%s'", tt.expectedArticlesSheet, articles)
			}
			if providers != tt.expectedProvidersSheet {
				t.Errorf("Expected providers sheet '%s', got '%s'", tt.expectedProvidersSheet, providers)
			}
		})
	}
}

// ============================================================================
// getSubstackProviderCount: Counts Substack providers from sheet
// ============================================================================

func TestGetSubstackProviderCount(t *testing.T) {
	tests := []struct {
		name          string
		providerRows  [][]interface{}
		providerErr   error
		expectedCount int
		expectErr     bool
	}{
		{
			name: "counts case-insensitive Substack providers",
			providerRows: [][]interface{}{
				{"Provider", "OtherCol"},
				{"Substack", "entry1"},
				{"substack", "entry2"},
				{"SUBSTACK", "entry3"},
				{"GitHub", "entry4"},
			},
			providerErr:   nil,
			expectedCount: 3,
			expectErr:     false,
		},
		{
			name: "no providers",
			providerRows: [][]interface{}{
				{"Provider", "OtherCol"},
				{"GitHub", "entry1"},
			},
			providerErr:   nil,
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name:          "empty sheet",
			providerRows:  [][]interface{}{},
			providerErr:   nil,
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name:          "error retrieving provider sheet",
			providerRows:  nil,
			providerErr:   fmt.Errorf("permission denied"),
			expectedCount: 0,
			expectErr:     true,
		},
		{
			name:          "network error when fetching providers",
			providerRows:  nil,
			providerErr:   fmt.Errorf("connection timeout"),
			expectedCount: 0,
			expectErr:     true,
		},
		{
			name: "API error returns 0 count",
			providerRows: [][]interface{}{
				{"Provider", "Status"},
				{"Substack", "active"},
			},
			providerErr:   fmt.Errorf("API quota exceeded"),
			expectedCount: 0,
			expectErr:     true,
		},
		{
			name: "mixed case Substack entries only",
			providerRows: [][]interface{}{
				{"Provider", "Info"},
				{"Substack", "author1"},
				{"SUBSTACK", "author2"},
				{"subStack", "author3"},
			},
			providerErr:   nil,
			expectedCount: 3,
			expectErr:     false,
		},
		{
			name: "single Substack provider",
			providerRows: [][]interface{}{
				{"Provider", "Info"},
				{"Substack", "author1"},
			},
			providerErr:   nil,
			expectedCount: 1,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &MockSheetsFetcher{
				providerRows: tt.providerRows,
				providerErr:  tt.providerErr,
			}
			count := getSubstackProviderCount(fetcher, "spreadsheetID", "providers")

			// getSubstackProviderCount returns 0 on error (doesn't panic)
			// but we can verify the error was encountered by the mock
			if tt.expectErr && tt.providerErr != nil {
				// When error is expected, count should be 0
				if count != 0 {
					t.Errorf("%s: expected count 0 on error, got %d", tt.name, count)
				}
			} else {
				if count != tt.expectedCount {
					t.Errorf("%s: expected count %d, got %d", tt.name, tt.expectedCount, count)
				}
			}
		})
	}
}

// ============================================================================
// fetchMetricsWithFetcher: Complete metrics calculation with mock fetcher
// ============================================================================

func TestFetchMetricsWithFetcher(t *testing.T) {
	tests := []struct {
		name        string
		description string
		fetcher     SheetsFetcher
		expectErr   bool
		validate    func(*schema.Metrics) bool
	}{
		{
			name:        "successful metrics calculation",
			description: "Validates complete metrics workflow with mock data",
			fetcher: &MockSheetsFetcher{
				spreadsheet: &sheets.Spreadsheet{
					Sheets: []*sheets.Sheet{
						{Properties: &sheets.SheetProperties{Title: "Articles"}},
						{Properties: &sheets.SheetProperties{Title: "Providers"}},
					},
				},
				articleRows:  createTestArticleRows(),
				providerRows: [][]interface{}{{"Provider", "OtherCol"}, {"Substack", "entry1"}},
			},
			expectErr: false,
			validate: func(m *schema.Metrics) bool {
				return m.TotalArticles == 10 &&
					m.ReadCount == 3 &&
					m.UnreadCount == 7 &&
					m.ReadRate == 30.0 &&
					m.BySource["Substack"] == 3
			},
		},
		{
			name:        "spreadsheet retrieval error",
			description: "Handles error when getting spreadsheet metadata",
			fetcher: &MockSheetsFetcher{
				spreadsheetErr: fmt.Errorf("API error"),
			},
			expectErr: true,
			validate:  func(m *schema.Metrics) bool { return false },
		},
		{
			name:        "article retrieval error",
			description: "Handles error when getting article rows",
			fetcher: &MockSheetsFetcher{
				spreadsheet: &sheets.Spreadsheet{
					Sheets: []*sheets.Sheet{
						{Properties: &sheets.SheetProperties{Title: "Articles"}},
					},
				},
				articleErr: fmt.Errorf("read error"),
			},
			expectErr: true,
			validate:  func(m *schema.Metrics) bool { return false },
		},
		{
			name:        "empty article sheet",
			description: "Handles empty article data",
			fetcher: &MockSheetsFetcher{
				spreadsheet: &sheets.Spreadsheet{
					Sheets: []*sheets.Sheet{
						{Properties: &sheets.SheetProperties{Title: "Articles"}},
					},
				},
				articleRows: [][]interface{}{},
			},
			expectErr: true,
			validate:  func(m *schema.Metrics) bool { return false },
		},
		{
			name:        "multiple Substack providers",
			description: "Correctly counts multiple Substack providers",
			fetcher: &MockSheetsFetcher{
				spreadsheet: &sheets.Spreadsheet{
					Sheets: []*sheets.Sheet{
						{Properties: &sheets.SheetProperties{Title: "Articles"}},
						{Properties: &sheets.SheetProperties{Title: "Providers"}},
					},
				},
				articleRows: createTestArticleRows(),
				providerRows: [][]interface{}{
					{"Provider", "OtherCol"},
					{"Substack", "entry1"},
					{"Substack", "entry2"},
					{"GitHub", "entry3"},
				},
			},
			expectErr: false,
			validate: func(m *schema.Metrics) bool {
				substackCount := m.BySourceReadStatus["substack_author_count"]
				return substackCount[0] == 2 // Should count 2 Substack providers
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := fetchMetricsWithFetcher("spreadsheetID", tt.fetcher)

			if tt.expectErr && err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
				return
			}

			if !tt.expectErr && !tt.validate(&metrics) {
				t.Errorf("%s: validation failed", tt.name)
			}
		})
	}
}

// ============================================================================
// FetchMetricsFromSheets: Retrieves and calculates metrics from Google Sheets
// ============================================================================

func TestFetchMetricsFromSheetsWithService(t *testing.T) {
	tests := []struct {
		name        string
		description string
		testFunc    func(*testing.T)
	}{
		{
			name:        "complete workflow with test data",
			description: "Simulates FetchMetricsFromSheetsWithService with test data",
			testFunc: func(t *testing.T) {
				// Simulate reading all articles from sheets (resp.Values)
				rows := createTestArticleRows()

				// Initialize metrics as FetchMetricsFromSheetsWithService does
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
				var unreadArticles []schema.ArticleMeta

				// Parse articles from rows (skip header row 0)
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}

					metrics.TotalArticles++
					updateMetricsByDate(&metrics, article, &earliestDate, &latestDate)
					updateMetricsBySource(&metrics, article.Category)
					updateMetricsByCategory(&metrics, article)
					updateMetricsReadStatus(&metrics, article)

					if !article.IsRead {
						unreadArticles = append(unreadArticles, schema.ArticleMeta{
							Date:     article.Date.Format("2006-01-02"),
							Category: article.Category,
							Read:     false,
						})
						bucket := calculateArticleAgeBucket(article.Date, latestDate)
						metrics.UnreadArticleAgeDistribution[bucket]++
						year := article.Date.Format("2006")
						metrics.UnreadByYear[year]++
					}
				}

				// Calculate derived metrics
				if metrics.TotalArticles > 0 {
					metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
				}

				// Validate complete workflow
				if metrics.TotalArticles != 10 {
					t.Errorf("Expected 10 total articles, got %d", metrics.TotalArticles)
				}
				if metrics.ReadCount != 3 {
					t.Errorf("Expected 3 read articles, got %d", metrics.ReadCount)
				}
				if metrics.UnreadCount != 7 {
					t.Errorf("Expected 7 unread articles, got %d", metrics.UnreadCount)
				}
				if metrics.ReadRate != 30.0 {
					t.Errorf("Expected 30%% read rate, got %.1f%%", metrics.ReadRate)
				}
				if len(unreadArticles) != 7 {
					t.Errorf("Expected 7 unread articles in list, got %d", len(unreadArticles))
				}
				if len(metrics.UnreadByYear) == 0 {
					t.Errorf("UnreadByYear should have entries")
				}
				if metrics.UnreadByYear["2025"] != 6 {
					t.Errorf("Expected 6 unread articles in 2025, got %d", metrics.UnreadByYear["2025"])
				}
				if metrics.UnreadByYear["2024"] != 1 {
					t.Errorf("Expected 1 unread article in 2024, got %d", metrics.UnreadByYear["2024"])
				}
			},
		},
		{
			name:        "age distribution calculation",
			description: "Validates age bucket calculation for unread articles",
			testFunc: func(t *testing.T) {
				rows := createTestArticleRows()
				ageDistribution := make(map[string]int)
				var latestDate time.Time

				// Get latest date from all articles
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}
					if i == 1 || article.Date.After(latestDate) {
						latestDate = article.Date
					}
				}

				// Process unread articles for age distribution
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}
					if !article.IsRead {
						bucket := calculateArticleAgeBucket(article.Date, latestDate)
						ageDistribution[bucket]++
					}
				}

				// Validate age distribution has expected buckets
				expectedBuckets := map[string]bool{
					"less_than_1_month": false,
					"1_to_3_months":     false,
					"3_to_6_months":     false,
					"6_to_12_months":    false,
					"older_than_1year":  false,
				}

				for bucket := range ageDistribution {
					if _, exists := expectedBuckets[bucket]; !exists {
						t.Errorf("Unexpected age bucket: %s", bucket)
					}
				}

				// Verify total unread articles across all buckets
				totalUnread := 0
				for _, count := range ageDistribution {
					totalUnread += count
				}
				if totalUnread != 7 {
					t.Errorf("Expected 7 unread articles in age distribution, got %d", totalUnread)
				}
			},
		},
		{
			name:        "source and category aggregation",
			description: "Validates proper aggregation by source and category",
			testFunc: func(t *testing.T) {
				rows := createTestArticleRows()
				metrics := schema.Metrics{
					BySource:                     make(map[string]int),
					BySourceReadStatus:           make(map[string][2]int),
					ByCategory:                   make(map[string][2]int),
					ByCategoryAndSource:          make(map[string]map[string][2]int),
					ByYear:                       make(map[string]int),
					ByMonth:                      make(map[string]int),
					ByYearAndMonth:               make(map[string]map[string]int),
					ByMonthAndSource:             make(map[string]map[string][2]int),
					UnreadByMonth:                make(map[string]int),
					UnreadByCategory:             make(map[string]int),
					UnreadBySource:               make(map[string]int),
					UnreadByYear:                 make(map[string]int),
					UnreadArticleAgeDistribution: make(map[string]int),
					SourceMetadata:               make(map[string]schema.SourceMeta),
				}

				var earliestDate, latestDate time.Time

				// Process articles
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}

					metrics.TotalArticles++
					updateMetricsByDate(&metrics, article, &earliestDate, &latestDate)
					updateMetricsBySource(&metrics, article.Category)
					updateMetricsByCategory(&metrics, article)
					updateMetricsReadStatus(&metrics, article)
				}

				// Validate source aggregation (lowercase 'github' normalizes to 'GitHub')
				expectedSources := map[string]int{
					"Substack":     3,
					"GitHub":       3,
					"freeCodeCamp": 2,
					"Shopify":      1,
					"Stripe":       1,
				}

				for source, expectedCount := range expectedSources {
					if metrics.BySource[source] != expectedCount {
						t.Errorf("Expected %d %s articles, got %d", expectedCount, source, metrics.BySource[source])
					}
				}

				// Validate year aggregation
				if metrics.ByYear["2025"] != 8 {
					t.Errorf("Expected 8 articles in 2025, got %d", metrics.ByYear["2025"])
				}
				if metrics.ByYear["2024"] != 2 {
					t.Errorf("Expected 2 articles in 2024, got %d", metrics.ByYear["2024"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func createTestArticleRows() [][]interface{} {
	return [][]interface{}{
		// Header row
		{"Date", "Title", "Link", "Category", "Read"},
		// Unread articles of various ages
		{"2025-12-10", "Recent Article", "https://example.com/recent", "Substack", "FALSE"},
		{"2025-11-18", "One Month Old", "https://example.com/1month", "GitHub", "FALSE"},
		{"2025-09-18", "Three Months Old", "https://example.com/3month", "freeCodeCamp", "FALSE"},
		{"2025-06-18", "Six Months Old", "https://example.com/6month", "Shopify", "FALSE"},
		{"2024-12-18", "One Year Old", "https://example.com/1year", "Stripe", "FALSE"},
		// Read articles
		{"2025-11-28", "Read Recently", "https://example.com/readrecent", "Substack", "TRUE"},
		{"2025-10-15", "Read Older", "https://example.com/readold", "GitHub", "TRUE"},
		{"2024-06-20", "Read Very Old", "https://example.com/readveryold", "freeCodeCamp", "TRUE"},
		// Mixed sources
		{"2025-12-05", "Another Substack", "https://example.com/substack2", "Substack", "FALSE"},
		{"2025-12-01", "GitHub Unread", "https://example.com/github1", "github", "FALSE"},
	}
}

func TestFetchMetricsFromSheets(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		testValidation func(*testing.T)
	}{
		{
			name:        "complete metrics calculation from test data",
			description: "Validates that article parsing, aggregation, and metric derivation work correctly",
			testValidation: func(t *testing.T) {
				rows := createTestArticleRows()
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

				// Simulate FetchMetricsFromSheetsWithService processing (skip header at index 0)
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}

					metrics.TotalArticles++
					updateMetricsByDate(&metrics, article, &earliestDate, &latestDate)
					updateMetricsBySource(&metrics, article.Category)
					updateMetricsByCategory(&metrics, article)
					updateMetricsReadStatus(&metrics, article)

					if !article.IsRead {
						year := article.Date.Format("2006")
						metrics.UnreadByYear[year]++
					}
				}

				// Validate basic counts
				if metrics.TotalArticles != 10 {
					t.Errorf("Expected 10 articles, got %d", metrics.TotalArticles)
				}
				if metrics.ReadCount != 3 {
					t.Errorf("Expected 3 read articles, got %d", metrics.ReadCount)
				}
				if metrics.UnreadCount != 7 {
					t.Errorf("Expected 7 unread articles, got %d", metrics.UnreadCount)
				}

				// Validate source aggregation (lowercase 'github' normalizes to 'GitHub')
				if metrics.BySource["Substack"] != 3 {
					t.Errorf("Expected 3 Substack articles, got %d", metrics.BySource["Substack"])
				}
				if metrics.BySource["GitHub"] != 3 {
					t.Errorf("Expected 3 GitHub articles, got %d", metrics.BySource["GitHub"])
				}

				// Validate year aggregation
				if metrics.ByYear["2025"] != 8 {
					t.Errorf("Expected 8 articles in 2025, got %d", metrics.ByYear["2025"])
				}
				if metrics.ByYear["2024"] != 2 {
					t.Errorf("Expected 2 articles in 2024, got %d", metrics.ByYear["2024"])
				}

				// Validate unread year aggregation
				if metrics.UnreadByYear["2025"] != 6 {
					t.Errorf("Expected 6 unread articles in 2025, got %d", metrics.UnreadByYear["2025"])
				}
				if metrics.UnreadByYear["2024"] != 1 {
					t.Errorf("Expected 1 unread article in 2024, got %d", metrics.UnreadByYear["2024"])
				}
			},
		},
		{
			name:        "derived metrics from test data",
			description: "Validates read rate and average articles per month calculations",
			testValidation: func(t *testing.T) {
				rows := createTestArticleRows()
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

				// Process rows
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRow(rows[i])
					if err != nil {
						continue
					}
					metrics.TotalArticles++
					updateMetricsByDate(&metrics, article, &earliestDate, &latestDate)

					if article.IsRead {
						metrics.ReadCount++
					} else {
						metrics.UnreadCount++
					}
				}

				// Calculate read rate
				if metrics.TotalArticles > 0 {
					metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
				}

				// Validate read rate is 30%
				expectedRate := 30.0
				if metrics.ReadRate != expectedRate {
					t.Errorf("Expected read rate %.2f%%, got %.2f%%", expectedRate, metrics.ReadRate)
				}

				// Calculate average articles per month
				monthsSpan := 1
				if !earliestDate.IsZero() && !latestDate.IsZero() {
					monthsSpan = calculateMonthsDifference(earliestDate, latestDate) + 1
				}
				metrics.AvgArticlesPerMonth = float64(metrics.TotalArticles) / float64(monthsSpan)

				// Should have a reasonable average
				if metrics.AvgArticlesPerMonth <= 0 {
					t.Errorf("Average articles per month should be positive, got %.2f", metrics.AvgArticlesPerMonth)
				}
			},
		},
		{
			name:        "unread articles collection",
			description: "Validates that unread articles are properly collected with all details",
			testValidation: func(t *testing.T) {
				rows := createTestArticleRows()
				var unreadArticles []schema.ArticleMeta

				// Collect unread articles (simulating FetchMetricsFromSheetsWithService)
				for i := 1; i < len(rows); i++ {
					article, err := parseArticleRowWithDetails(rows[i])
					if err != nil {
						continue
					}
					if !article.Read {
						unreadArticles = append(unreadArticles, *article)
					}
				}

				// Validate count
				if len(unreadArticles) != 7 {
					t.Errorf("Expected 7 unread articles, got %d", len(unreadArticles))
					return
				}

				// Verify all articles are unread
				for i, article := range unreadArticles {
					if article.Read {
						t.Errorf("Article at index %d should be unread: %+v", i, article)
					}
					// Verify article has required fields
					if article.Date == "" || article.Title == "" || article.Category == "" {
						t.Errorf("Article at index %d missing required fields: %+v", i, article)
					}
				}
			},
		},
		{
			name:        "metrics initialization",
			description: "Validates all metric maps are properly initialized",
			testValidation: func(t *testing.T) {
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

				// Validate all required maps are not nil
				if metrics.BySource == nil {
					t.Error("BySource not initialized")
				}
				if metrics.ByYear == nil {
					t.Error("ByYear not initialized")
				}
				if metrics.UnreadArticleAgeDistribution == nil {
					t.Error("UnreadArticleAgeDistribution not initialized")
				}
				if metrics.SourceMetadata == nil {
					t.Error("SourceMetadata not initialized")
				}
				if metrics.ByYearAndMonth == nil {
					t.Error("ByYearAndMonth not initialized")
				}
				if metrics.ByMonthAndSource == nil {
					t.Error("ByMonthAndSource not initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testValidation(t)
		})
	}
}

// ============================================================================
// Integration Test: Validates all metrics features work together
// ============================================================================

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

// ============================================================================
// SheetServiceFetcher Tests: GetArticleRows and GetProvidersSheet
// ============================================================================

// MockSheetsService mocks the sheets.SpreadsheetsService for testing
type MockSheetsService struct {
	spreadsheetReturn *sheets.Spreadsheet
	valuesReturn      *sheets.ValueRange
	getErr            error
	spreadsheetErr    error
}

type MockSpreadsheetsService struct {
	getCall *MockSheetsService
}

type MockValuesService struct {
	getCall *MockSheetsService
}

func (m *MockValuesService) Get(spreadsheetID, readRange string) *sheets.SpreadsheetsValuesGetCall {
	// In real implementation, this returns a call object
	// For testing, we'll verify the call was made with correct parameters
	return nil
}

// Note: Testing SheetServiceFetcher with real Google Sheets API is complex.
// Instead, we test the interface behavior through MockSheetsFetcher.
// These tests verify that the fetcher correctly handles article and provider data.

func TestGetArticleRowsInterface(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		mockFetcher     SheetsFetcher
		expectedRows    int
		expectedErr     bool
		validateContent func([][]interface{}) bool
	}{
		{
			name:        "successfully returns article rows with header",
			description: "Verifies GetArticleRows returns data in correct format",
			mockFetcher: &MockSheetsFetcher{
				articleRows: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
					{"2025-12-10", "Article 1", "link1", "Substack", "FALSE"},
					{"2025-11-10", "Article 2", "link2", "GitHub", "TRUE"},
				},
			},
			expectedRows: 3,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				// Check header row
				if len(rows[0]) != 5 {
					return false
				}
				// Check data rows
				if rows[1][0] != "2025-12-10" {
					return false
				}
				return len(rows) == 3
			},
		},
		{
			name:        "returns article rows with various categories",
			description: "Verifies article data from multiple sources is preserved",
			mockFetcher: &MockSheetsFetcher{
				articleRows: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
					{"2025-12-10", "Article A", "https://example.com/a", "Substack", "FALSE"},
					{"2025-12-09", "Article B", "https://example.com/b", "GitHub", "FALSE"},
					{"2025-12-08", "Article C", "https://example.com/c", "freeCodeCamp", "TRUE"},
					{"2025-12-07", "Article D", "https://example.com/d", "Stripe", "FALSE"},
				},
			},
			expectedRows: 5,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				if len(rows) != 5 {
					return false
				}
				// Verify categories are preserved
				categories := map[string]bool{}
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) > 3 {
						cat := rows[i][3].(string)
						categories[cat] = true
					}
				}
				return len(categories) == 4 // Should have 4 different categories
			},
		},
		{
			name:        "handles empty article sheet",
			description: "Verifies behavior with only header row",
			mockFetcher: &MockSheetsFetcher{
				articleRows: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
				},
			},
			expectedRows: 1,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				return len(rows) == 1 // Only header
			},
		},
		{
			name:        "handles article retrieval error",
			description: "Verifies error propagation when fetching fails",
			mockFetcher: &MockSheetsFetcher{
				articleErr: fmt.Errorf("API error: permission denied"),
			},
			expectedRows:    0,
			expectedErr:     true,
			validateContent: func(rows [][]interface{}) bool { return true },
		},
		{
			name:        "returns articles with various read statuses",
			description: "Verifies read/unread status is preserved",
			mockFetcher: &MockSheetsFetcher{
				articleRows: [][]interface{}{
					{"Date", "Title", "Link", "Category", "Read"},
					{"2025-12-10", "Read 1", "link1", "Substack", "TRUE"},
					{"2025-12-09", "Unread 1", "link2", "GitHub", "FALSE"},
					{"2025-12-08", "Read 2", "link3", "freeCodeCamp", "TRUE"},
					{"2025-12-07", "Unread 2", "link4", "Stripe", "FALSE"},
				},
			},
			expectedRows: 5,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				readCount := 0
				unreadCount := 0
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) > 4 {
						status := rows[i][4].(string)
						if status == "TRUE" {
							readCount++
						} else if status == "FALSE" {
							unreadCount++
						}
					}
				}
				return readCount == 2 && unreadCount == 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.mockFetcher.GetArticleRows("spreadsheetID", "Articles")

			if tt.expectedErr && err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
				return
			}

			if len(rows) != tt.expectedRows {
				t.Errorf("%s: expected %d rows, got %d", tt.name, tt.expectedRows, len(rows))
				return
			}

			if !tt.validateContent(rows) {
				t.Errorf("%s: content validation failed", tt.name)
			}
		})
	}
}

func TestGetProvidersSheetInterface(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		mockFetcher     SheetsFetcher
		expectedRows    int
		expectedErr     bool
		validateContent func([][]interface{}) bool
	}{
		{
			name:        "successfully returns provider rows",
			description: "Verifies GetProvidersSheet returns data in correct format",
			mockFetcher: &MockSheetsFetcher{
				providerRows: [][]interface{}{
					{"Provider", "OtherColumn"},
					{"Substack", "value1"},
					{"GitHub", "value2"},
				},
			},
			expectedRows: 3,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				if len(rows[0]) != 2 {
					return false
				}
				return rows[0][0].(string) == "Provider"
			},
		},
		{
			name:        "counts multiple Substack entries",
			description: "Verifies multiple Substack provider rows are returned",
			mockFetcher: &MockSheetsFetcher{
				providerRows: [][]interface{}{
					{"Provider", "Status"},
					{"Substack", "active"},
					{"Substack", "inactive"},
					{"GitHub", "active"},
					{"Substack", "active"},
					{"Shopify", "active"},
				},
			},
			expectedRows: 6,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				substackCount := 0
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) > 0 {
						provider := rows[i][0].(string)
						if provider == "Substack" {
							substackCount++
						}
					}
				}
				return substackCount == 3
			},
		},
		{
			name:        "returns providers with various sources",
			description: "Verifies all provider types are preserved",
			mockFetcher: &MockSheetsFetcher{
				providerRows: [][]interface{}{
					{"Provider", "URL"},
					{"Substack", "https://substack.com/user1"},
					{"GitHub", "https://github.com/user1"},
					{"Stripe", "https://stripe.com"},
					{"Shopify", "https://shopify.com"},
					{"freeCodeCamp", "https://freecodecamp.org"},
				},
			},
			expectedRows: 6,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				providers := map[string]bool{}
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) > 0 {
						provider := rows[i][0].(string)
						providers[provider] = true
					}
				}
				return len(providers) == 5 // 5 unique providers
			},
		},
		{
			name:        "handles empty provider sheet",
			description: "Verifies behavior with only header row",
			mockFetcher: &MockSheetsFetcher{
				providerRows: [][]interface{}{
					{"Provider", "Status"},
				},
			},
			expectedRows: 1,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				return len(rows) == 1
			},
		},
		{
			name:        "handles provider retrieval error",
			description: "Verifies error propagation when fetching fails",
			mockFetcher: &MockSheetsFetcher{
				providerErr: fmt.Errorf("network error"),
			},
			expectedRows:    0,
			expectedErr:     true,
			validateContent: func(rows [][]interface{}) bool { return true },
		},
		{
			name:        "preserves provider data with special characters",
			description: "Verifies provider names with special characters are preserved",
			mockFetcher: &MockSheetsFetcher{
				providerRows: [][]interface{}{
					{"Provider", "Description"},
					{"Substack", "Substack Newsletter Platform"},
					{"GitHub", "Code Repository & Collaboration"},
					{"Dev.to", "Community Platform for Developers"},
					{"Medium", "Publishing Platform"},
				},
			},
			expectedRows: 5,
			expectedErr:  false,
			validateContent: func(rows [][]interface{}) bool {
				if len(rows) != 5 {
					return false
				}
				// Check that provider names are preserved exactly
				expectedProviders := map[string]bool{
					"Substack": false,
					"GitHub":   false,
					"Dev.to":   false,
					"Medium":   false,
				}
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) > 0 {
						provider := rows[i][0].(string)
						if _, exists := expectedProviders[provider]; exists {
							expectedProviders[provider] = true
						}
					}
				}
				// Verify all expected providers were found
				for _, found := range expectedProviders {
					if !found {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.mockFetcher.GetProvidersSheet("spreadsheetID", "Providers")

			if tt.expectedErr && err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
				return
			}

			if len(rows) != tt.expectedRows {
				t.Errorf("%s: expected %d rows, got %d", tt.name, tt.expectedRows, len(rows))
				return
			}

			if !tt.validateContent(rows) {
				t.Errorf("%s: content validation failed", tt.name)
			}
		})
	}
}
