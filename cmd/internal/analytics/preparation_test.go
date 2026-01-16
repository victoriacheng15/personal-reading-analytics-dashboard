package analytics

import (
	"encoding/json"
	"html/template"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

func TestPrepareReadUnreadByYear(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedYear0   string
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
		expectEmpty     bool
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
			expectEmpty:     false,
		},
		{
			name:        "empty metrics",
			metrics:     schema.Metrics{ByYear: map[string]int{}},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadByYear(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if tt.expectEmpty {
				if len(labels) != 0 {
					t.Errorf("expected empty labels, got %d", len(labels))
				}
				if len(readData) != 0 {
					t.Errorf("expected empty readData, got %d", len(readData))
				}
				return
			}

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

func TestPrepareReadUnreadByMonth(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
		expectedRead2   float64
		isAllZero       bool
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
			isAllZero:       false,
		},
		{
			name:      "empty metrics returns zeroed arrays",
			metrics:   schema.Metrics{},
			isAllZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadByMonth(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if tt.isAllZero {
				if len(readData) != 12 {
					t.Errorf("expected 12 months, got %d", len(readData))
				}
				for i := 0; i < 12; i++ {
					if readData[i].(float64) != 0 || unreadData[i].(float64) != 0 {
						t.Errorf("expected zero at index %d", i)
					}
				}
				return
			}

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
		{
			name:           "nil sources list",
			sources:        nil,
			expectedLabels: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadBySource(tt.sources)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if len(labels) != tt.expectedLabels {
				t.Errorf("expected %d labels, got %d", tt.expectedLabels, len(labels))
			}
			if tt.expectedLabels > 0 {
				if labels[0].(string) != tt.expectedFirstLabel {
					t.Errorf("expected %s, got %s", tt.expectedFirstLabel, labels[0])
				}
				if readData[0].(float64) != tt.expectedRead {
					t.Errorf("expected %v read, got %v", tt.expectedRead, readData[0])
				}
				if unreadData[0].(float64) != tt.expectedUnread {
					t.Errorf("expected %v unread, got %v", tt.expectedUnread, unreadData[0])
				}
			}
		})
	}
}

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
				// Also verify values are 0
				for _, val := range data {
					if val.(float64) != 0 {
						t.Error("expected 0 for empty distribution buckets")
					}
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

			jsonStr := PrepareUnreadArticleAgeDistribution(*metrics)
			tt.validate(t, jsonStr)
		})
	}
}

func TestPrepareUnreadArticleAgeDistributionJSON(t *testing.T) {
	metrics := createTestMetricsWithAgeDistribution()
	jsonStr := PrepareUnreadArticleAgeDistribution(*metrics)

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

			jsonStr := PrepareUnreadByYear(*metrics)
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
			jsonStr := PrepareUnreadByYear(*tt.metrics)

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
