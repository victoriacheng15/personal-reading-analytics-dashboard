package main

import (
	"os"
	"path/filepath"
	"testing"
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
// getMetricsDates: Returns all YYYY-MM-DD dates from JSON files in metrics/ folder
// loadMetricsByDate: Reads a specific metrics JSON file from metrics/ folder
// ============================================================================

func TestGetMetricsDates(t *testing.T) {
	tests := []struct {
		name          string
		fileNames     []string
		expectedDates []string
		expectError   bool
	}{
		{
			name:          "returns sorted dates",
			fileNames:     []string{"2025-01-01.json", "2024-01-01.json", "invalid.txt"},
			expectedDates: []string{"2025-01-01", "2024-01-01"},
			expectError:   false,
		},
		{
			name:          "no valid metrics files",
			fileNames:     []string{"not-a-date.txt"},
			expectedDates: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_metrics_dates")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.Mkdir(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			for _, fileName := range tt.fileNames {
				if err := os.WriteFile(filepath.Join(metricsDir, fileName), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			dates, err := getMetricsDates()
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}

			if len(dates) != len(tt.expectedDates) {
				t.Errorf("expected %d dates, got %d", len(tt.expectedDates), len(dates))
			}

			for i := range dates {
				if dates[i] != tt.expectedDates[i] {
					t.Errorf("expected date %s, got %s", tt.expectedDates[i], dates[i])
				}
			}
		})
	}
}

func TestLoadMetricsByDate(t *testing.T) {
	tests := []struct {
		name             string
		date             string
		fileContent      string
		expectedArticles int
		expectError      bool
	}{
		{
			name:             "loads metrics for specific date",
			date:             "2025-01-01",
			fileContent:      `{"total_articles": 100}`,
			expectedArticles: 100,
			expectError:      false,
		},
		{
			name:             "non-existent date",
			date:             "2000-01-01",
			fileContent:      "",
			expectedArticles: 0,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_metrics_by_date")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.Mkdir(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if tt.fileContent != "" {
				if err := os.WriteFile(filepath.Join(metricsDir, tt.date+".json"), []byte(tt.fileContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			metrics, err := loadMetricsByDate(tt.date)
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}
			if metrics.TotalArticles != tt.expectedArticles {
				t.Errorf("expected %d articles, got %d", tt.expectedArticles, metrics.TotalArticles)
			}
		})
	}
}
