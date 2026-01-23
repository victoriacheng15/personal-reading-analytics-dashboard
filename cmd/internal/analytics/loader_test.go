package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetTemplatesDir tests the GetTemplatesDir function
func TestGetTemplatesDir(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns temp dir path
		expectError bool
		expectEmpty bool
	}{
		{
			name: "finds templates directory from primary path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for primary path
				templatesDir := filepath.Join("cmd", "internal", "analytics", "templates")
				if err := os.MkdirAll(templatesDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "finds templates directory from secondary path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for secondary path
				templatesDir := filepath.Join("internal", "analytics", "templates")
				if err := os.MkdirAll(templatesDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "returns error when templates directory not found",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				return tmpDir
			},
			expectError: true,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)
			defer os.RemoveAll(tmpDir)

			dir, err := GetTemplatesDir()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tt.expectEmpty && dir != "" {
					t.Errorf("expected empty path on error, got: %v", dir)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectEmpty && dir == "" {
				t.Errorf("expected non-empty path, got empty string")
			}

			if !tt.expectEmpty && dir == "" {
				t.Errorf("expected non-empty path, got empty string")
			}
		})
	}
}

// TestLoadEvolutionData tests the LoadEvolutionData function
func TestLoadEvolutionData(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "loads evolution data successfully",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				dir := filepath.Join("cmd", "internal", "analytics", "content")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}

				yamlContent := `
chapters:
  - title: "Chapter 1"
    timeline:
      - date: "2024-01"
        title: "Test Event"
        description: |
          - "Detail 1"
`
				if err := os.WriteFile(filepath.Join(dir, "evolution.yml"), []byte(yamlContent), 0644); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			expectError: false,
		},
		{
			name: "returns error when file missing",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				return tmpDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)
			defer os.RemoveAll(tmpDir)

			data, err := LoadEvolutionData()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(data.Chapters) > 0 && len(data.Chapters[0].Timeline) > 0 && data.Chapters[0].Timeline[0].Title != "Test Event" {
				t.Errorf("expected title 'Test Event', got %s", data.Chapters[0].Timeline[0].Title)
			}

			if len(data.Chapters) > 0 && len(data.Chapters[0].Timeline) > 0 {
				if len(data.Chapters[0].Timeline[0].DescriptionLines) == 0 || data.Chapters[0].Timeline[0].DescriptionLines[0] != "Detail 1" {
					t.Errorf("expected DescriptionLines[0] to be 'Detail 1', got %v", data.Chapters[0].Timeline[0].DescriptionLines)
				}
			}
		})
	}
}
