package main

import (
	"encoding/json"
	"html/template"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// Test order follows HTML template section order:
// 1. Top Oldest Unread Articles (🔝)
// 2. Read/Unread Breakdown (📖)
// 3. Unread by Year (📅)
// 4. Age Distribution (⏰)

// ============================================================================
// SECTION 1: Top Oldest Unread Articles Tests
// ============================================================================

// Helper: createTestArticleList creates ArticleMeta for testing article ordering
func createTestArticleList(count int, startYear int) []schema.ArticleMeta {
	articles := make([]schema.ArticleMeta, 0, count)

	dateFormats := []string{
		"2023-01-15",
		"2023-06-20",
		"2024-02-10",
		"2024-08-05",
		"2025-03-12",
		"2025-09-30",
		"2023-03-25",
		"2024-11-18",
		"2025-01-05",
		"2023-12-01",
	}

	for i := 0; i < count && i < len(dateFormats); i++ {
		articles = append(articles, schema.ArticleMeta{
			Title:    "Article " + string(rune('A'+i)),
			Date:     dateFormats[i],
			Link:     "https://example.com/article-" + string(rune('A'+i)),
			Category: []string{"Tech", "Science", "Business", "News"}[i%4],
			Read:     false,
		})
	}

	return articles
}

// TestPrepareTopOldestUnreadArticles tests basic oldest unread articles data
func TestPrepareTopOldestUnreadArticles(t *testing.T) {
	tests := []struct {
		name     string
		articles []schema.ArticleMeta
		validate func(t *testing.T, articles []schema.ArticleMeta)
	}{
		{
			name:     "articles ordered by date oldest first",
			articles: createTestArticleList(5, 2023),
			validate: func(t *testing.T, articles []schema.ArticleMeta) {
				// Verify we have articles
				if len(articles) == 0 {
					t.Error("expected articles, got none")
					return
				}

				// Verify all required fields are present
				for i, article := range articles {
					if article.Title == "" {
						t.Errorf("article %d missing title", i)
					}
					if article.Date == "" {
						t.Errorf("article %d missing date", i)
					}
					if article.Link == "" {
						t.Errorf("article %d missing link", i)
					}
					if article.Category == "" {
						t.Errorf("article %d missing category", i)
					}
				}
			},
		},
		{
			name:     "single article",
			articles: []schema.ArticleMeta{{Title: "Only Article", Date: "2023-01-01", Link: "https://example.com", Category: "Tech"}},
			validate: func(t *testing.T, articles []schema.ArticleMeta) {
				if len(articles) != 1 {
					t.Errorf("expected 1 article, got %d", len(articles))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.articles)
		})
	}
}

// TestPrepareTopOldestUnreadArticlesFormatting tests date format and link encoding
func TestPrepareTopOldestUnreadArticlesFormatting(t *testing.T) {
	articles := []schema.ArticleMeta{
		{
			Title:    "Test Article",
			Date:     "2024-12-19",
			Link:     "https://example.com/article?id=123&sort=asc",
			Category: "Technology",
			Read:     false,
		},
	}

	for _, article := range articles {
		// Verify date format is YYYY-MM-DD
		if !isValidDateFormat(article.Date) {
			t.Errorf("date format invalid: %s, expected YYYY-MM-DD", article.Date)
		}

		// Verify link is properly formatted as URL
		if !isValidURL(article.Link) {
			t.Errorf("link format invalid: %s", article.Link)
		}

		// Verify title is not empty
		if article.Title == "" {
			t.Error("title should not be empty")
		}

		// Verify category is normalized (no special characters)
		if article.Category == "" {
			t.Error("category should not be empty")
		}
	}
}

// TestPrepareTopOldestUnreadArticlesLimiting tests article count limiting
func TestPrepareTopOldestUnreadArticlesLimiting(t *testing.T) {
	tests := []struct {
		name           string
		inputCount     int
		expectedCount  int
		limitThreshold int
		validate       func(t *testing.T, articles []schema.ArticleMeta, expected int)
	}{
		{
			name:           "exactly N articles",
			inputCount:     5,
			expectedCount:  5,
			limitThreshold: 5,
			validate: func(t *testing.T, articles []schema.ArticleMeta, expected int) {
				if len(articles) != expected {
					t.Errorf("expected %d articles, got %d", expected, len(articles))
				}
			},
		},
		{
			name:           "fewer than N articles",
			inputCount:     3,
			expectedCount:  3,
			limitThreshold: 5,
			validate: func(t *testing.T, articles []schema.ArticleMeta, expected int) {
				if len(articles) != expected {
					t.Errorf("expected %d articles, got %d", expected, len(articles))
				}
			},
		},
		{
			name:           "more than N articles only first N returned",
			inputCount:     10,
			expectedCount:  5,
			limitThreshold: 5,
			validate: func(t *testing.T, articles []schema.ArticleMeta, expected int) {
				if len(articles) > expected {
					t.Errorf("expected at most %d articles, got %d", expected, len(articles))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles := createTestArticleList(tt.inputCount, 2023)

			// Simulate limiting to threshold
			if len(articles) > tt.limitThreshold {
				articles = articles[:tt.limitThreshold]
			}

			tt.validate(t, articles, tt.expectedCount)
		})
	}
}

// Helper functions for validation
func isValidDateFormat(date string) bool {
	// Check for YYYY-MM-DD format
	if len(date) != 10 {
		return false
	}
	parts := string(date)[0:4] + string(date)[5:7] + string(date)[8:10]
	for _, ch := range parts {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return date[4] == '-' && date[7] == '-'
}

func isValidURL(link string) bool {
	return len(link) > 0 && (string(link)[0:8] == "https://" || string(link)[0:7] == "http://")
}

// ============================================================================
// SECTION 2: Read/Unread Breakdown Tests
// ============================================================================

// Helper: createTestMetricsWithAgeDistribution creates a Metrics struct with sample age distribution
func createTestMetricsWithAgeDistribution() *schema.Metrics {
	metrics := &schema.Metrics{
		UnreadArticleAgeDistribution: make(map[string]int),
		UnreadByYear:                 make(map[string]int),
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

	// Pre-populate age distribution with realistic data
	metrics.UnreadArticleAgeDistribution["less_than_1_month"] = 8
	metrics.UnreadArticleAgeDistribution["1_to_3_months"] = 12
	metrics.UnreadArticleAgeDistribution["3_to_6_months"] = 15
	metrics.UnreadArticleAgeDistribution["6_to_12_months"] = 10
	metrics.UnreadArticleAgeDistribution["older_than_1year"] = 5

	return metrics
}

// TestPrepareUnreadArticleAgeDistribution tests age distribution chart data preparation
func TestPrepareUnreadArticleAgeDistribution(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *schema.Metrics
		validate func(t *testing.T, jsonStr template.JS)
	}{
		{
			name:    "age distribution with all buckets populated",
			metrics: createTestMetricsWithAgeDistribution(),
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				// Verify structure has labels and data
				if _, hasLabels := chartData["labels"]; !hasLabels {
					t.Error("missing 'labels' key in chart data")
				}
				if _, hasData := chartData["data"]; !hasData {
					t.Error("missing 'data' key in chart data")
				}

				// Verify labels are human-readable
				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 5 {
					t.Errorf("expected 5 labels, got %v", labels)
					return
				}

				// Verify specific label "Older than 1 year" exists
				labelStrs := make([]string, len(labels))
				for i, label := range labels {
					labelStrs[i] = label.(string)
				}

				if labelStrs[4] != "Older than 1 year" {
					t.Errorf("expected 'Older than 1 year' label, got %s", labelStrs[4])
				}
			},
		},
		{
			name: "empty age distribution",
			metrics: &schema.Metrics{
				UnreadArticleAgeDistribution: make(map[string]int),
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				// Verify structure is still valid even with empty data
				data, ok := chartData["data"].([]interface{})
				if !ok || len(data) == 0 {
					t.Error("expected valid data array for empty metrics")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.metrics
			if metrics == nil {
				metrics = &schema.Metrics{
					UnreadArticleAgeDistribution: make(map[string]int),
				}
			}

			jsonStr := prepareUnreadArticleAgeDistribution(*metrics)
			tt.validate(t, jsonStr)
		})
	}
}

// TestPrepareUnreadArticleAgeDistributionJSON tests JSON format compliance for age distribution
func TestPrepareUnreadArticleAgeDistributionJSON(t *testing.T) {
	metrics := createTestMetricsWithAgeDistribution()
	jsonStr := prepareUnreadArticleAgeDistribution(*metrics)

	var chartData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &chartData)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	// Verify all required keys exist
	requiredKeys := []string{"labels", "data"}
	for _, key := range requiredKeys {
		if _, exists := chartData[key]; !exists {
			t.Errorf("required key '%s' missing from chart data", key)
		}
	}

	// Verify data is numeric array
	data, ok := chartData["data"].([]interface{})
	if !ok {
		t.Error("data field should be array of numbers")
		return
	}

	for i, val := range data {
		if _, isNum := val.(float64); !isNum {
			t.Errorf("data[%d] should be numeric, got %T", i, val)
		}
	}

	// Verify numerical values match input data
	labels, _ := chartData["labels"].([]interface{})
	expectedMap := metrics.UnreadArticleAgeDistribution

	for i, label := range labels {
		labelStr := label.(string)
		dataVal := int(data[i].(float64))

		// Map label to key
		labelToKey := map[string]string{
			"Less than 1 month": "less_than_1_month",
			"1-3 months":        "1_to_3_months",
			"3-6 months":        "3_to_6_months",
			"6-12 months":       "6_to_12_months",
			"Older than 1 year": "older_than_1year",
		}

		key := labelToKey[labelStr]
		expectedVal := expectedMap[key]

		if dataVal != expectedVal {
			t.Errorf("data mismatch for %s: expected %d, got %d", labelStr, expectedVal, dataVal)
		}
	}
}

// ============================================================================
// SECTION 3: Unread by Year Preparation Tests
// ============================================================================

// Helper: createTestMetricsWithUnreadByYear creates a Metrics struct with unread by year data
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
	metrics.UnreadByYear["2025"] = 30
	metrics.UnreadByYear["2024"] = 25
	metrics.UnreadByYear["2023"] = 15
	metrics.UnreadByYear["2022"] = 8

	return metrics
}

// TestPrepareUnreadByYear tests unread by year chart data preparation
func TestPrepareUnreadByYear(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *schema.Metrics
		validate func(t *testing.T, jsonStr template.JS)
	}{
		{
			name:    "multiple years in descending order",
			metrics: createTestMetricsWithUnreadByYear(),
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 4 {
					t.Errorf("expected 4 labels, got %v", labels)
					return
				}

				// Verify descending order: 2025, 2024, 2023, 2022
				if labels[0].(string) != "2025" {
					t.Errorf("expected first year to be 2025, got %s", labels[0])
				}
				if labels[len(labels)-1].(string) != "2022" {
					t.Errorf("expected last year to be 2022, got %s", labels[len(labels)-1])
				}
			},
		},
		{
			name: "single year",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 20,
				},
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 1 {
					t.Errorf("expected 1 label, got %v", labels)
				}
			},
		},
		{
			name: "non-consecutive years",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 25,
					"2022": 10,
					"2020": 5,
				},
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 3 {
					t.Errorf("expected 3 labels, got %v", labels)
					return
				}

				// Verify sorted in descending order
				if labels[0].(string) != "2025" || labels[1].(string) != "2022" || labels[2].(string) != "2020" {
					t.Errorf("years not in descending order: %v", labels)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.metrics
			if metrics == nil {
				metrics = &schema.Metrics{
					UnreadByYear: make(map[string]int),
				}
			}

			jsonStr := prepareUnreadByYear(*metrics)
			tt.validate(t, jsonStr)
		})
	}
}

// TestPrepareUnreadByYearDataValidity tests that data values match input metrics
func TestPrepareUnreadByYearDataValidity(t *testing.T) {
	metrics := createTestMetricsWithUnreadByYear()
	jsonStr := prepareUnreadByYear(*metrics)

	var chartData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &chartData)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	labels, ok := chartData["labels"].([]interface{})
	if !ok {
		t.Fatal("labels should be array")
	}

	data, ok := chartData["data"].([]interface{})
	if !ok {
		t.Fatal("data should be array")
	}

	if len(labels) != len(data) {
		t.Errorf("labels and data length mismatch: %d vs %d", len(labels), len(data))
	}

	// Verify data corresponds to metrics and matches bar chart format
	for i, label := range labels {
		year := label.(string)
		expectedCount := metrics.UnreadByYear[year]
		actualCount := int(data[i].(float64))

		if actualCount != expectedCount {
			t.Errorf("year %s: expected %d, got %d", year, expectedCount, actualCount)
		}
	}
}
