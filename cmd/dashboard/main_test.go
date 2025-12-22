package main

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

func isValidDateFormat(date string) bool {
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
// loadLatestMetrics: Loads the latest metrics JSON file from the metrics directory
// ============================================================================

func TestLoadLatestMetrics(t *testing.T) {
	tests := []struct {
		name             string
		fileNames        []string
		fileContents     []string
		expectedArticles int
		expectError      bool
	}{
		{
			name:             "loads latest metrics file",
			fileNames:        []string{"2025-01-01.json", "2024-01-01.json"},
			fileContents:     []string{`{"total_articles": 100}`, `{"total_articles": 50}`},
			expectedArticles: 100,
			expectError:      false,
		},
		{
			name:             "single metrics file",
			fileNames:        []string{"2024-01-01.json"},
			fileContents:     []string{`{"total_articles": 50}`},
			expectedArticles: 50,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_metrics")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.Mkdir(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			for i, fileName := range tt.fileNames {
				if err := os.WriteFile(filepath.Join(metricsDir, fileName), []byte(tt.fileContents[i]), 0644); err != nil {
					t.Fatal(err)
				}
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			metrics, err := loadLatestMetrics()
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}
			if metrics.TotalArticles != tt.expectedArticles {
				t.Errorf("expected %d articles, got %d", tt.expectedArticles, metrics.TotalArticles)
			}
		})
	}
}

// ============================================================================
// calculateTopReadRateSource: Calculates which source has the highest read rate
// ============================================================================

func TestCalculateTopReadRateSource(t *testing.T) {
	tests := []struct {
		name           string
		metrics        schema.Metrics
		expectedSource string
	}{
		{
			name: "identifies highest read rate",
			metrics: schema.Metrics{
				BySourceReadStatus: map[string][2]int{
					"SourceA":               {10, 90},
					"SourceB":               {80, 20},
					"SourceC":               {50, 50},
					"substack_author_count": {100, 0},
				},
			},
			expectedSource: "SourceB",
		},
		{
			name: "ignores substack_author_count",
			metrics: schema.Metrics{
				BySourceReadStatus: map[string][2]int{
					"SourceA":               {30, 70},
					"SourceB":               {20, 80},
					"substack_author_count": {100, 0},
				},
			},
			expectedSource: "SourceA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topSource := calculateTopReadRateSource(tt.metrics)
			if topSource != tt.expectedSource {
				t.Errorf("expected %s, got %s", tt.expectedSource, topSource)
			}
		})
	}
}

// ============================================================================
// calculateMostUnreadSource: Identifies the source with the most unread articles
// ============================================================================

func TestCalculateMostUnreadSource(t *testing.T) {
	tests := []struct {
		name           string
		unreadBySource map[string]int
		expectedSource string
	}{
		{
			name: "identifies source with most unread",
			unreadBySource: map[string]int{
				"SourceA": 10,
				"SourceB": 50,
				"SourceC": 5,
			},
			expectedSource: "SourceB",
		},
		{
			name: "single source",
			unreadBySource: map[string]int{
				"SourceA": 100,
			},
			expectedSource: "SourceA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := schema.Metrics{
				UnreadBySource: tt.unreadBySource,
			}
			mostUnread := calculateMostUnreadSource(metrics)
			if mostUnread != tt.expectedSource {
				t.Errorf("expected %s, got %s", tt.expectedSource, mostUnread)
			}
		})
	}
}

// ============================================================================
// calculateThisMonthArticles: Counts articles added in the current month
// ============================================================================

func TestCalculateThisMonthArticles(t *testing.T) {
	tests := []struct {
		name          string
		metrics       schema.Metrics
		month         string
		expectedCount int
	}{
		{
			name: "multiple sources in month",
			metrics: schema.Metrics{
				ByMonthAndSource: map[string]map[string][2]int{
					"01": {
						"SourceA": {5, 2},
						"SourceB": {3, 1},
					},
					"02": {
						"SourceA": {10, 5},
					},
				},
			},
			month:         "01",
			expectedCount: 8,
		},
		{
			name: "single source in month",
			metrics: schema.Metrics{
				ByMonthAndSource: map[string]map[string][2]int{
					"02": {
						"SourceA": {10, 5},
					},
				},
			},
			month:         "02",
			expectedCount: 10,
		},
		{
			name: "month with no data",
			metrics: schema.Metrics{
				ByMonthAndSource: map[string]map[string][2]int{
					"01": {
						"SourceA": {5, 2},
					},
				},
			},
			month:         "03",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := calculateThisMonthArticles(tt.metrics, tt.month)
			if count != tt.expectedCount {
				t.Errorf("expected %d articles, got %d", tt.expectedCount, count)
			}
		})
	}
}

// ============================================================================
// prepareReadUnreadByYear: Generates read/unread breakdown organized by year
// ============================================================================

func TestPrepareReadUnreadByYear(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedYear0   string
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
	}{
		{
			name: "multiple years with correct values",
			metrics: schema.Metrics{
				ByYear: map[string]int{
					"2024": 100,
					"2023": 50,
				},
				ByYearAndMonth: map[string]map[string]int{
					"2024": {"01": 10, "02": 20},
					"2023": {"01": 5},
				},
				UnreadByMonth: map[string]int{
					"01": 2,
					"02": 3,
				},
			},
			expectedYear0:   "2024",
			expectedRead0:   30,
			expectedUnread0: 5,
			expectedRead1:   5,
			expectedUnread1: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := prepareReadUnreadByYear(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if labels[0].(string) != tt.expectedYear0 {
				t.Errorf("expected year %s first, got %s", tt.expectedYear0, labels[0])
			}
			if readData[0].(float64) != tt.expectedRead0 {
				t.Errorf("expected %v read, got %v", tt.expectedRead0, readData[0])
			}
			if unreadData[0].(float64) != tt.expectedUnread0 {
				t.Errorf("expected %v unread, got %v", tt.expectedUnread0, unreadData[0])
			}
			if readData[1].(float64) != tt.expectedRead1 {
				t.Errorf("expected %v read, got %v", tt.expectedRead1, readData[1])
			}
			if unreadData[1].(float64) != tt.expectedUnread1 {
				t.Errorf("expected %v unread, got %v", tt.expectedUnread1, unreadData[1])
			}
		})
	}
}

// ============================================================================
// prepareReadUnreadByMonth: Generates read/unread breakdown organized by month
// ============================================================================

func TestPrepareReadUnreadByMonth(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
		expectedRead2   float64
	}{
		{
			name: "monthly breakdown with correct calculations",
			metrics: schema.Metrics{
				UnreadByMonth: map[string]int{
					"01": 5,
					"02": 10,
				},
				ByMonth: map[string]int{
					"01": 20,
					"02": 30,
				},
			},
			expectedRead0:   15,
			expectedUnread0: 5,
			expectedRead1:   20,
			expectedUnread1: 10,
			expectedRead2:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := prepareReadUnreadByMonth(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if readData[0].(float64) != tt.expectedRead0 {
				t.Errorf("expected %v read for Jan, got %v", tt.expectedRead0, readData[0])
			}
			if unreadData[0].(float64) != tt.expectedUnread0 {
				t.Errorf("expected %v unread for Jan, got %v", tt.expectedUnread0, unreadData[0])
			}
			if readData[1].(float64) != tt.expectedRead1 {
				t.Errorf("expected %v read for Feb, got %v", tt.expectedRead1, readData[1])
			}
			if unreadData[1].(float64) != tt.expectedUnread1 {
				t.Errorf("expected %v unread for Feb, got %v", tt.expectedUnread1, unreadData[1])
			}
			if readData[2].(float64) != tt.expectedRead2 {
				t.Errorf("expected %v read for Mar, got %v", tt.expectedRead2, readData[2])
			}
		})
	}
}

// ============================================================================
// prepareReadUnreadBySource: Generates read/unread breakdown organized by source
// ============================================================================

func TestPrepareReadUnreadBySource(t *testing.T) {
	tests := []struct {
		name               string
		sources            []schema.SourceInfo
		expectedLabels     int
		expectedFirstLabel string
		expectedRead       float64
		expectedUnread     float64
	}{
		{
			name: "multiple sources",
			sources: []schema.SourceInfo{
				{Name: "SourceA", Read: 10, Unread: 5},
				{Name: "SourceB", Read: 20, Unread: 0},
			},
			expectedLabels:     2,
			expectedFirstLabel: "SourceA",
			expectedRead:       10,
			expectedUnread:     5,
		},
		{
			name: "single source",
			sources: []schema.SourceInfo{
				{Name: "SourceX", Read: 15, Unread: 3},
			},
			expectedLabels:     1,
			expectedFirstLabel: "SourceX",
			expectedRead:       15,
			expectedUnread:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := prepareReadUnreadBySource(tt.sources)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if len(labels) != tt.expectedLabels {
				t.Errorf("expected %d labels, got %d", tt.expectedLabels, len(labels))
			}
			if labels[0].(string) != tt.expectedFirstLabel {
				t.Errorf("expected %s, got %s", tt.expectedFirstLabel, labels[0])
			}
			if readData[0].(float64) != tt.expectedRead {
				t.Errorf("expected %v read, got %v", tt.expectedRead, readData[0])
			}
			if unreadData[0].(float64) != tt.expectedUnread {
				t.Errorf("expected %v unread, got %v", tt.expectedUnread, unreadData[0])
			}
		})
	}
}

// ============================================================================
// prepareUnreadArticleAgeDistribution: Categorizes unread articles by age buckets
// ============================================================================

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

	metrics.UnreadArticleAgeDistribution["less_than_1_month"] = 8
	metrics.UnreadArticleAgeDistribution["1_to_3_months"] = 12
	metrics.UnreadArticleAgeDistribution["3_to_6_months"] = 15
	metrics.UnreadArticleAgeDistribution["6_to_12_months"] = 10
	metrics.UnreadArticleAgeDistribution["older_than_1year"] = 5

	return metrics
}

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

				if _, hasLabels := chartData["labels"]; !hasLabels {
					t.Error("missing 'labels' key in chart data")
				}
				if _, hasData := chartData["data"]; !hasData {
					t.Error("missing 'data' key in chart data")
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 5 {
					t.Errorf("expected 5 labels, got %v", labels)
					return
				}

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

func TestPrepareUnreadArticleAgeDistributionJSON(t *testing.T) {
	metrics := createTestMetricsWithAgeDistribution()
	jsonStr := prepareUnreadArticleAgeDistribution(*metrics)

	var chartData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &chartData)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	requiredKeys := []string{"labels", "data"}
	for _, key := range requiredKeys {
		if _, exists := chartData[key]; !exists {
			t.Errorf("required key '%s' missing from chart data", key)
		}
	}

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

	labels, _ := chartData["labels"].([]interface{})
	expectedMap := metrics.UnreadArticleAgeDistribution

	for i, label := range labels {
		labelStr := label.(string)
		dataVal := int(data[i].(float64))

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
// prepareUnreadByYear: Generates unread article counts organized by year
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

	metrics.UnreadByYear["2025"] = 30
	metrics.UnreadByYear["2024"] = 25
	metrics.UnreadByYear["2023"] = 15
	metrics.UnreadByYear["2022"] = 8

	return metrics
}

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

func TestPrepareUnreadByYearDataValidity(t *testing.T) {
	tests := []struct {
		name          string
		metrics       *schema.Metrics
		expectedValid bool
	}{
		{
			name:          "data matches input metrics",
			metrics:       createTestMetricsWithUnreadByYear(),
			expectedValid: true,
		},
		{
			name: "single year",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 100,
				},
			},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := prepareUnreadByYear(*tt.metrics)

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

			for i, label := range labels {
				year := label.(string)
				expectedCount := tt.metrics.UnreadByYear[year]
				actualCount := int(data[i].(float64))
				if actualCount != expectedCount {
					t.Errorf("year %s: expected %d, got %d", year, expectedCount, actualCount)
				}
			}
		})
	}
}

// ============================================================================
// generateHTMLDashboard: Renders the complete dashboard HTML from metrics data
// ============================================================================

func TestGenerateHTMLDashboard(t *testing.T) {
	tests := []struct {
		name          string
		metrics       schema.Metrics
		expectSuccess bool
	}{
		{
			name: "generates html dashboard with metrics",
			metrics: schema.Metrics{
				TotalArticles: 10,
				BySource:      map[string]int{"SourceA": 10},
				BySourceReadStatus: map[string][2]int{
					"SourceA":               {5, 5},
					"substack_author_count": {0, 0},
				},
				ByYear:  map[string]int{"2024": 10},
				ByMonth: map[string]int{"01": 10},
				ByMonthAndSource: map[string]map[string][2]int{
					"01": {"SourceA": {5, 5}},
				},
				UnreadByMonth: map[string]int{"01": 5},
				UnreadByYear:  map[string]int{"2024": 5},
				UnreadArticleAgeDistribution: map[string]int{
					"less_than_1_month": 5,
					"1_to_3_months":     0,
					"3_to_6_months":     0,
					"6_to_12_months":    0,
					"older_than_1year":  0,
				},
			},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "dashboard_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			templateDir := filepath.Join(tmpDir, "cmd", "internal", "dashboard")
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatal(err)
			}

			dummyTemplate := `<html><body><h1>{{.DashboardTitle}}</h1></body></html>`
			if err := os.WriteFile(filepath.Join(templateDir, "template.html"), []byte(dummyTemplate), 0644); err != nil {
				t.Fatal(err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			err = generateHTMLDashboard(tt.metrics)
			if (err == nil) != tt.expectSuccess {
				t.Errorf("generateHTMLDashboard error = %v, expectSuccess %v", err, tt.expectSuccess)
			}

			if _, err := os.Stat("site/index.html"); os.IsNotExist(err) {
				t.Error("site/index.html was not created")
			}
		})
	}
}

// ============================================================================
// main: Orchestrates the complete dashboard generation pipeline
// ============================================================================

func TestMainExecution(t *testing.T) {
	tests := []struct {
		name                 string
		metricsJSON          string
		expectHTMLGeneration bool
	}{
		{
			name:                 "main generates html from latest metrics",
			metricsJSON:          `{"total_articles": 100, "by_source": {"Test": 100}, "by_source_read_status": {"Test": [50, 50], "substack_author_count": [0,0]}, "unread_article_age_distribution": {"less_than_1_month": 10, "1_to_3_months": 0, "3_to_6_months": 0, "6_to_12_months": 0, "older_than_1year": 0}}`,
			expectHTMLGeneration: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "main_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.MkdirAll(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(metricsDir, "2025-01-01.json"), []byte(tt.metricsJSON), 0644); err != nil {
				t.Fatal(err)
			}

			templateDir := filepath.Join(tmpDir, "cmd", "internal", "dashboard")
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatal(err)
			}
			dummyTemplate := `<html><body><h1>Main Test</h1></body></html>`
			if err := os.WriteFile(filepath.Join(templateDir, "template.html"), []byte(dummyTemplate), 0644); err != nil {
				t.Fatal(err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			main()

			_, err = os.Stat("site/index.html")
			if tt.expectHTMLGeneration && os.IsNotExist(err) {
				t.Error("site/index.html was not created by main()")
			}
		})
	}
}
