package metrics

import (
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

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
					"SourceA":               {10, 90}, // 10%
					"SourceB":               {80, 20}, // 80% (Winner)
					"SourceC":               {50, 50}, // 50%
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
					"substack_author_count": {100, 0}, // Would be 100%, but must be ignored
				},
			},
			expectedSource: "SourceA",
		},
		{
			name:           "empty metrics returns empty string",
			metrics:        schema.Metrics{BySourceReadStatus: map[string][2]int{}},
			expectedSource: "",
		},
		{
			name: "handles source with zero total articles (avoid div by zero)",
			metrics: schema.Metrics{
				BySourceReadStatus: map[string][2]int{
					"SourceA": {0, 0}, // 0 total
					"SourceB": {5, 5}, // 50%
				},
			},
			expectedSource: "SourceB",
		},
		{
			name: "handles tie breaking (first encountered or unstable, but safe)",
			metrics: schema.Metrics{
				BySourceReadStatus: map[string][2]int{
					"SourceA": {10, 0}, // 100%
					"SourceB": {10, 0}, // 100%
				},
			},
			// Note: Map iteration order is random in Go, so either is valid.
			// We just ensure it returns *one* of them and doesn't crash.
			// In a real deterministic requirement, we'd sort keys first.
			expectedSource: "SourceA", // Or SourceB, logic implies > topRate, so first one wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topSource := CalculateTopReadRateSource(tt.metrics)
			// For the tie-breaker case, we accept either valid winner
			if tt.name == "handles tie breaking (first encountered or unstable, but safe)" {
				if topSource != "SourceA" && topSource != "SourceB" {
					t.Errorf("expected SourceA or SourceB, got %s", topSource)
				}
			} else if topSource != tt.expectedSource {
				t.Errorf("expected %s, got %s", tt.expectedSource, topSource)
			}
		})
	}
}

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
		{
			name:           "empty metrics returns empty string",
			unreadBySource: map[string]int{},
			expectedSource: "",
		},
		{
			name: "tie breaker returns one of the top sources",
			unreadBySource: map[string]int{
				"SourceA": 50,
				"SourceB": 50,
			},
			expectedSource: "SourceA", // Random map order, but checking for safety
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := schema.Metrics{
				UnreadBySource: tt.unreadBySource,
			}
			mostUnread := CalculateMostUnreadSource(metrics)

			if tt.name == "tie breaker returns one of the top sources" {
				if mostUnread != "SourceA" && mostUnread != "SourceB" {
					t.Errorf("expected SourceA or SourceB, got %s", mostUnread)
				}
			} else if mostUnread != tt.expectedSource {
				t.Errorf("expected %s, got %s", tt.expectedSource, mostUnread)
			}
		})
	}
}

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
		{
			name:          "empty metrics",
			metrics:       schema.Metrics{ByMonthAndSource: map[string]map[string][2]int{}},
			month:         "01",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CalculateThisMonthArticles(tt.metrics, tt.month)
			if count != tt.expectedCount {
				t.Errorf("expected %d articles, got %d", tt.expectedCount, count)
			}
		})
	}
}
