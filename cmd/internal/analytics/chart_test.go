package analytics

import (
	"encoding/json"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

// ============================================================================
// formatHex: Formats a number as a 6-digit hex string
// ============================================================================

func TestFormatHex(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "000000",
		},
		{
			name:     "small number",
			input:    255,
			expected: "0000ff",
		},
		{
			name:     "mid range",
			input:    4095,
			expected: "000fff",
		},
		{
			name:     "large number",
			input:    16777215,
			expected: "ffffff",
		},
		{
			name:     "arbitrary value",
			input:    12345,
			expected: "003039",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHex(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// colorHash: Generates a simple hash for generating colors
// ============================================================================

func TestColorHash(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedLength int
		expectNonEmpty bool
	}{
		{
			name:           "simple string",
			input:          "test",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "empty string",
			input:          "",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "long string",
			input:          "this is a much longer test string",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "special characters",
			input:          "@#$%^&*()",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "consistent hash",
			input:          "Substack",
			expectedLength: 6,
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorHash(tt.input)
			if len(result) != tt.expectedLength {
				t.Errorf("expected length %d, got %d", tt.expectedLength, len(result))
			}
			if tt.expectNonEmpty && len(result) == 0 {
				t.Error("expected non-empty result")
			}

			// Verify it's valid hex
			for _, ch := range result {
				if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
					t.Errorf("invalid hex character: %c", ch)
				}
			}
		})
	}
}

func TestColorHashConsistency(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "same input produces same hash",
			input: "GitHub",
		},
		{
			name:  "different inputs produce different hashes",
			input: "Substack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := colorHash(tt.input)
			hash2 := colorHash(tt.input)
			if hash1 != hash2 {
				t.Errorf("expected consistent hash, got %s and %s", hash1, hash2)
			}
		})
	}
}

// ============================================================================
// PrepareYearChartData: Prepares year breakdown chart data
// ============================================================================

func TestPrepareYearChartData(t *testing.T) {
	tests := []struct {
		name             string
		years            []schema.YearInfo
		expectedLabels   []string
		expectedDataLen  int
		shouldHaveLabels bool
		shouldHaveData   bool
	}{
		{
			name: "single year",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
			},
			expectedLabels:   []string{"2025"},
			expectedDataLen:  1,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
		{
			name: "multiple years",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
				{Year: "2024", Count: 80},
				{Year: "2023", Count: 50},
			},
			expectedLabels:   []string{"2025", "2024", "2023"},
			expectedDataLen:  3,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
		{
			name:             "empty years",
			years:            []schema.YearInfo{},
			expectedLabels:   []string{},
			expectedDataLen:  0,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareYearChartData(tt.years)

			if !tt.shouldHaveLabels || result.LabelsJSON == nil {
				if tt.shouldHaveLabels && result.LabelsJSON == nil {
					t.Error("expected LabelsJSON, got nil")
				}
				return
			}

			var labels []string
			err := json.Unmarshal(result.LabelsJSON, &labels)
			if err != nil {
				t.Fatalf("failed to unmarshal labels: %v", err)
			}

			if len(labels) != len(tt.expectedLabels) {
				t.Errorf("expected %d labels, got %d", len(tt.expectedLabels), len(labels))
			}

			for i, expected := range tt.expectedLabels {
				if i < len(labels) && labels[i] != expected {
					t.Errorf("label[%d]: expected %s, got %s", i, expected, labels[i])
				}
			}

			if !tt.shouldHaveData || result.DataJSON == nil {
				if tt.shouldHaveData && result.DataJSON == nil {
					t.Error("expected DataJSON, got nil")
				}
				return
			}

			var data []int
			err = json.Unmarshal(result.DataJSON, &data)
			if err != nil {
				t.Fatalf("failed to unmarshal data: %v", err)
			}

			if len(data) != tt.expectedDataLen {
				t.Errorf("expected %d data points, got %d", tt.expectedDataLen, len(data))
			}
		})
	}
}

func TestPrepareYearChartDataValues(t *testing.T) {
	tests := []struct {
		name         string
		years        []schema.YearInfo
		expectedData []int
	}{
		{
			name: "correct count values",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
				{Year: "2024", Count: 75},
				{Year: "2023", Count: 50},
			},
			expectedData: []int{100, 75, 50},
		},
		{
			name: "zero counts",
			years: []schema.YearInfo{
				{Year: "2025", Count: 0},
				{Year: "2024", Count: 50},
			},
			expectedData: []int{0, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareYearChartData(tt.years)

			var data []int
			err := json.Unmarshal(result.DataJSON, &data)
			if err != nil {
				t.Fatalf("failed to unmarshal data: %v", err)
			}

			for i, expected := range tt.expectedData {
				if i < len(data) && data[i] != expected {
					t.Errorf("data[%d]: expected %d, got %d", i, expected, data[i])
				}
			}
		})
	}
}

// ============================================================================
// PrepareMonthChartData: Prepares month breakdown chart data with source stacking
// ============================================================================

func TestPrepareMonthChartData(t *testing.T) {
	tests := []struct {
		name            string
		months          []schema.MonthInfo
		sources         []schema.SourceInfo
		expectedLabels  []string
		expectDatasets  bool
		expectTotalData bool
	}{
		{
			name: "single month with sources",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack":     30,
						"freeCodeCamp": 20,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 10, Unread: 20},
				{Name: "freeCodeCamp", Read: 8, Unread: 12},
			},
			expectedLabels:  []string{"January"},
			expectDatasets:  true,
			expectTotalData: true,
		},
		{
			name: "multiple months",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack": 30,
						"GitHub":   20,
					},
				},
				{
					Name:  "February",
					Total: 60,
					Sources: map[string]int{
						"Substack": 40,
						"GitHub":   20,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 15, Unread: 25},
				{Name: "GitHub", Read: 20, Unread: 0},
			},
			expectedLabels:  []string{"January", "February"},
			expectDatasets:  true,
			expectTotalData: true,
		},
		{
			name:            "empty months",
			months:          []schema.MonthInfo{},
			sources:         []schema.SourceInfo{},
			expectedLabels:  []string{},
			expectDatasets:  true,
			expectTotalData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareMonthChartData(tt.months, tt.sources)

			// Verify labels
			var labels []string
			err := json.Unmarshal(result.LabelsJSON, &labels)
			if err != nil {
				t.Fatalf("failed to unmarshal labels: %v", err)
			}

			if len(labels) != len(tt.expectedLabels) {
				t.Errorf("expected %d labels, got %d", len(tt.expectedLabels), len(labels))
			}

			for i, expected := range tt.expectedLabels {
				if i < len(labels) && labels[i] != expected {
					t.Errorf("label[%d]: expected %s, got %s", i, expected, labels[i])
				}
			}

			// Verify datasets
			if tt.expectDatasets {
				var datasets []map[string]interface{}
				err := json.Unmarshal(result.DatasetsJSON, &datasets)
				if err != nil {
					t.Fatalf("failed to unmarshal datasets: %v", err)
				}

				for _, dataset := range datasets {
					if _, hasLabel := dataset["label"]; !hasLabel {
						t.Error("dataset missing label")
					}
					if _, hasData := dataset["data"]; !hasData {
						t.Error("dataset missing data")
					}
				}
			}

			// Verify total data
			if tt.expectTotalData {
				var totalData []int
				err := json.Unmarshal(result.TotalDataJSON, &totalData)
				if err != nil {
					t.Fatalf("failed to unmarshal total data: %v", err)
				}

				if len(totalData) != len(tt.months) {
					t.Errorf("expected %d total data points, got %d", len(tt.months), len(totalData))
				}
			}
		})
	}
}

func TestPrepareMonthChartDataDataValues(t *testing.T) {
	tests := []struct {
		name          string
		months        []schema.MonthInfo
		sources       []schema.SourceInfo
		expectedTotal []int
	}{
		{
			name: "correct monthly totals",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack": 30,
						"GitHub":   20,
					},
				},
				{
					Name:  "February",
					Total: 75,
					Sources: map[string]int{
						"Substack": 50,
						"GitHub":   25,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 15, Unread: 65},
				{Name: "GitHub", Read: 30, Unread: 15},
			},
			expectedTotal: []int{50, 75},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareMonthChartData(tt.months, tt.sources)

			var totalData []int
			err := json.Unmarshal(result.TotalDataJSON, &totalData)
			if err != nil {
				t.Fatalf("failed to unmarshal total data: %v", err)
			}

			for i, expected := range tt.expectedTotal {
				if i < len(totalData) && totalData[i] != expected {
					t.Errorf("total[%d]: expected %d, got %d", i, expected, totalData[i])
				}
			}
		})
	}
}

func TestPrepareMonthChartDataColorAssignment(t *testing.T) {
	tests := []struct {
		name          string
		sourceName    string
		providedColor string
		expectedColor string
	}{
		{
			name:          "Source with provided color uses that color",
			sourceName:    "Substack",
			providedColor: "#667eea",
			expectedColor: "#667eea",
		},
		{
			name:          "Source without provided color uses hash-generated color",
			sourceName:    "UnknownSource",
			providedColor: "",
			expectedColor: "#", // Should start with #
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			months := []schema.MonthInfo{
				{
					Name:  "January",
					Total: 30,
					Sources: map[string]int{
						tt.sourceName: 30,
					},
				},
			}

			sources := []schema.SourceInfo{
				{Name: tt.sourceName, Read: 10, Unread: 20, Color: tt.providedColor},
			}

			result := PrepareMonthChartData(months, sources)

			var datasets []map[string]interface{}
			err := json.Unmarshal(result.DatasetsJSON, &datasets)
			if err != nil {
				t.Fatalf("failed to unmarshal datasets: %v", err)
			}

			if len(datasets) > 0 {
				bgColor := datasets[0]["backgroundColor"]
				if bgColor == nil {
					t.Error("backgroundColor is missing")
					return
				}

				colorStr := bgColor.(string)
				if tt.providedColor != "" && colorStr != tt.expectedColor {
					t.Errorf("expected color %s, got %s", tt.expectedColor, colorStr)
				}

				if colorStr[0] != '#' {
					t.Errorf("backgroundColor should start with #, got %s", colorStr)
				}
			}
		})
	}
}
