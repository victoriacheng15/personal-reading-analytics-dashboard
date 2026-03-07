package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/victoriacheng15/personal-reading-analytics/internal"
)

func TestConstructPrompt(t *testing.T) {
	curr := &internal.Metrics{
		TotalArticles: 10,
		ReadCount:     5,
		ReadRate:      50.0,
	}

	t.Run("with previous metrics", func(t *testing.T) {
		prev := &internal.Metrics{
			TotalArticles: 8,
			ReadCount:     4,
			ReadRate:      50.0,
		}
		prompt := constructPrompt(curr, prev)
		if !contains(prompt, "Compare the following") {
			t.Errorf("expected comparison prompt, got: %s", prompt)
		}
		if !contains(prompt, "PREVIOUS WEEK") {
			t.Errorf("expected PREVIOUS WEEK section, got: %s", prompt)
		}
	})

	t.Run("without previous metrics", func(t *testing.T) {
		prompt := constructPrompt(curr, nil)
		if !contains(prompt, "Analyze the following reading metrics") {
			t.Errorf("expected snapshot prompt, got: %s", prompt)
		}
		if contains(prompt, "PREVIOUS WEEK") {
			t.Errorf("did not expect PREVIOUS WEEK section")
		}
	})
}

func TestLoadPreviousMetrics(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some mock metrics files
	files := []struct {
		name string
		data internal.Metrics
	}{
		{"2026-01-01.json", internal.Metrics{TotalArticles: 100}},
		{"2026-01-08.json", internal.Metrics{TotalArticles: 110}},
		{"2026-01-15.json", internal.Metrics{TotalArticles: 120}},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		bytes, _ := json.Marshal(f.data)
		os.WriteFile(path, bytes, 0644)
	}

	t.Run("find immediate predecessor", func(t *testing.T) {
		prev, err := loadPreviousMetrics(tmpDir, "2026-01-15.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prev.TotalArticles != 110 {
			t.Errorf("expected 110 articles from 2026-01-08.json, got %d", prev.TotalArticles)
		}
	})

	t.Run("first file has no predecessor", func(t *testing.T) {
		_, err := loadPreviousMetrics(tmpDir, "2026-01-01.json")
		if err == nil {
			t.Error("expected error for first file, got nil")
		}
	})

	t.Run("file not in list", func(t *testing.T) {
		_, err := loadPreviousMetrics(tmpDir, "2026-01-22.json")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

func TestSaveUpdatedMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	filename := "test.json"
	m := &internal.Metrics{
		TotalArticles:   10,
		AIDeltaAnalysis: "Looks good!",
		LastUpdated:     time.Now(),
	}

	err := saveMetrics(tmpDir, filename, m)
	if err != nil {
		t.Fatalf("saveMetrics failed: %v", err)
	}

	// Read back and verify
	bytes, err := os.ReadFile(filepath.Join(tmpDir, filename))
	if err != nil {
		t.Fatalf("failed to read back: %v", err)
	}

	var result internal.Metrics
	json.Unmarshal(bytes, &result)

	if result.AIDeltaAnalysis != "Looks good!" {
		t.Errorf("AIDeltaAnalysis mismatch: got %s", result.AIDeltaAnalysis)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
