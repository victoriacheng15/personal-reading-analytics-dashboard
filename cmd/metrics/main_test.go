package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/internal"
)

// MockMetricsFetcher implements MetricsFetcher for testing
type MockMetricsFetcher struct {
	mockMetrics schema.Metrics
	mockError   error
}

func (m *MockMetricsFetcher) FetchMetrics(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error) {
	return m.mockMetrics, m.mockError
}

// createMockMetrics creates sample metrics for testing
func createMockMetrics(lastUpdated time.Time) schema.Metrics {
	return schema.Metrics{
		TotalArticles:       42,
		BySource:            map[string]int{"GitHub": 10, "Substack": 32},
		BySourceReadStatus:  map[string][2]int{"GitHub": {8, 2}, "Substack": {28, 4}},
		ByYear:              map[string]int{"2025": 42},
		ByMonth:             map[string]int{"2025-11": 15, "2025-12": 27},
		ByYearAndMonth:      map[string]map[string]int{"2025": {"11": 15, "12": 27}},
		ReadUnreadTotals:    [2]int{36, 6},
		ReadCount:           36,
		UnreadCount:         6,
		ReadRate:            85.71,
		AvgArticlesPerMonth: 10.5,
		LastUpdated:         lastUpdated,
	}
}

// TestLoadConfiguration tests configuration loading scenarios
func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		envSheetID     string
		envCredPath    string
		expectedSheet  string
		expectedCred   string
		expectError    bool
		errorSubstring string
	}{
		{
			name:          "Success with both env vars",
			envSheetID:    "test-sheet-123",
			envCredPath:   "/path/to/creds.json",
			expectedSheet: "test-sheet-123",
			expectedCred:  "/path/to/creds.json",
			expectError:   false,
		},
		{
			name:          "Success with default credentials path",
			envSheetID:    "test-sheet-123",
			envCredPath:   "",
			expectedSheet: "test-sheet-123",
			expectedCred:  "./credentials.json",
			expectError:   false,
		},
		{
			name:           "Missing SHEET_ID",
			envSheetID:     "",
			envCredPath:    "/path/to/creds.json",
			expectedSheet:  "",
			expectedCred:   "",
			expectError:    true,
			errorSubstring: "SHEET_ID environment variable is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalSheetID := os.Getenv("SHEET_ID")
			originalCredPath := os.Getenv("CREDENTIALS_PATH")
			defer func() {
				os.Setenv("SHEET_ID", originalSheetID)
				os.Setenv("CREDENTIALS_PATH", originalCredPath)
			}()

			// Set test values
			if tt.envSheetID != "" {
				os.Setenv("SHEET_ID", tt.envSheetID)
			} else {
				os.Unsetenv("SHEET_ID")
			}

			if tt.envCredPath != "" {
				os.Setenv("CREDENTIALS_PATH", tt.envCredPath)
			} else {
				os.Unsetenv("CREDENTIALS_PATH")
			}

			sheetID, credentialsPath, err := loadConfiguration()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorSubstring != "" && !contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error containing %q, got %q", tt.errorSubstring, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if sheetID != tt.expectedSheet {
					t.Errorf("Expected sheetID %q, got %q", tt.expectedSheet, sheetID)
				}
				if credentialsPath != tt.expectedCred {
					t.Errorf("Expected credentialsPath %q, got %q", tt.expectedCred, credentialsPath)
				}
			}
		})
	}
}

// TestSaveMetrics tests saving metrics to file
func TestSaveMetrics(t *testing.T) {
	lastUpdated := time.Date(2025, 12, 21, 10, 30, 0, 0, time.UTC)
	mockMetrics := createMockMetrics(lastUpdated)

	tests := []struct {
		name        string
		setup       func(dir string) error
		metrics     schema.Metrics
		expectError bool
		errorSubstr string
	}{
		{
			name:        "Success",
			setup:       func(dir string) error { return nil },
			metrics:     mockMetrics,
			expectError: false,
		},
		{
			name: "Write Error (Directory blocked)",
			setup: func(dir string) error {
				// Block metrics directory creation by creating a file named "metrics"
				return os.WriteFile("metrics", []byte("blocker"), 0644)
			},
			metrics:     mockMetrics,
			expectError: true,
			errorSubstr: "failed to create metrics directory",
		},
		{
			name: "Write Error (File blocked)",
			setup: func(dir string) error {
				os.MkdirAll("metrics", 0755)
				// Create a directory where the file should be to block writing
				return os.Mkdir(filepath.Join("metrics", "2025-12-21.json"), 0755)
			},
			metrics:     mockMetrics,
			expectError: true,
			errorSubstr: "failed to write metrics file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get current directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp directory: %v", err)
			}
			defer os.Chdir(originalDir)

			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			filename, err := saveMetrics(tt.metrics)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorSubstr != "" && !contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Expected error containing %q, got %q", tt.errorSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expectedFilename := "2025-12-21.json"
				if filename != expectedFilename {
					t.Errorf("Expected filename %q, got %q", expectedFilename, filename)
				}
				// Verify file exists
				if _, err := os.Stat(filepath.Join("metrics", filename)); err != nil {
					t.Errorf("File was not created: %v", err)
				}
			}
		})
	}
}

// TestRunFetch tests the runFetch function
func TestRunFetch(t *testing.T) {
	lastUpdated := time.Date(2025, 12, 21, 10, 30, 0, 0, time.UTC)
	mockMetrics := createMockMetrics(lastUpdated)

	tests := []struct {
		name           string
		sheetID        string
		fetcher        MetricsFetcher
		expectError    bool
		errorSubstring string
	}{
		{
			name:    "Success",
			sheetID: "valid-sheet",
			fetcher: &MockMetricsFetcher{
				mockMetrics: mockMetrics,
				mockError:   nil,
			},
			expectError: false,
		},
		{
			name:    "Missing Configuration",
			sheetID: "", // Triggers config error
			fetcher: &MockMetricsFetcher{
				mockMetrics: mockMetrics,
				mockError:   nil,
			},
			expectError:    true,
			errorSubstring: "SHEET_ID",
		},
		{
			name:    "Fetch Error",
			sheetID: "valid-sheet",
			fetcher: &MockMetricsFetcher{
				mockError: fmt.Errorf("API error"),
			},
			expectError:    true,
			errorSubstring: "failed to fetch metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get current directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp directory: %v", err)
			}
			defer os.Chdir(originalDir)

			// Setup Env
			originalSheetID := os.Getenv("SHEET_ID")
			defer os.Setenv("SHEET_ID", originalSheetID)

			if tt.sheetID != "" {
				os.Setenv("SHEET_ID", tt.sheetID)
			} else {
				os.Unsetenv("SHEET_ID")
			}
			os.Setenv("CREDENTIALS_PATH", "dummy.json")

			filename, metrics, err := runFetch(context.Background(), tt.fetcher)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorSubstring != "" && !contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error containing %q, got %q", tt.errorSubstring, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if filename == "" {
					t.Error("Expected filename to be returned")
				}
				if metrics == nil {
					t.Error("Expected metrics to be returned")
				}
			}
		})
	}
}

// TestMainBehavior tests the main execution logic via execute()
func TestMainBehavior(t *testing.T) {
	tests := []struct {
		name              string
		sheetID           string
		fetchSuccess      bool
		expectError       bool
		expectFileCreated bool
	}{
		{
			name:              "Success path",
			sheetID:           "test-sheet-123",
			fetchSuccess:      true,
			expectError:       false,
			expectFileCreated: true,
		},
		{
			name:              "Config failure",
			sheetID:           "",
			fetchSuccess:      false, // Irrelevant
			expectError:       true,
			expectFileCreated: false,
		},
		{
			name:              "Fetch failure",
			sheetID:           "test-sheet-123",
			fetchSuccess:      false,
			expectError:       true,
			expectFileCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get current directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp directory: %v", err)
			}
			defer os.Chdir(originalDir)

			// Mock dependencies
			originalSheetID := os.Getenv("SHEET_ID")
			originalCredPath := os.Getenv("CREDENTIALS_PATH")
			originalFetchMetricsFunc := fetchMetricsFunc
			defer func() {
				os.Setenv("SHEET_ID", originalSheetID)
				os.Setenv("CREDENTIALS_PATH", originalCredPath)
				fetchMetricsFunc = originalFetchMetricsFunc
			}()

			if tt.sheetID != "" {
				os.Setenv("SHEET_ID", tt.sheetID)
			} else {
				os.Unsetenv("SHEET_ID")
			}
			os.Setenv("CREDENTIALS_PATH", "creds.json")

			// Mock FetchMetrics
			fetchMetricsFunc = func(ctx context.Context, sheetID, credentialsPath string) (schema.Metrics, error) {
				if tt.fetchSuccess {
					return createMockMetrics(time.Date(2025, 12, 21, 10, 30, 0, 0, time.UTC)), nil
				}
				return schema.Metrics{}, fmt.Errorf("fetch failed")
			}

			// Call execute() directly instead of main() to avoid flag redefinition
			fetcher := &DefaultMetricsFetcher{}
			// Default flags: fetch=false, summarize=false -> runs both
			err = execute(context.Background(), fetcher, false, false)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Check file existence
			expectedFile := filepath.Join("metrics", "2025-12-21.json")
			_, err = os.Stat(expectedFile)
			fileCreated := err == nil

			if tt.expectFileCreated != fileCreated {
				t.Errorf("File created: %v, expected: %v", fileCreated, tt.expectFileCreated)
			}
		})
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
