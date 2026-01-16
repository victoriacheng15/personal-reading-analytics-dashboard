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
