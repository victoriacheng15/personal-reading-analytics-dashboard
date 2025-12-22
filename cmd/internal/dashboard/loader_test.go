package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTemplateContent tests the LoadTemplateContent function
func TestLoadTemplateContent(t *testing.T) {
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
			name: "loads template from primary path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for primary path
				dashboardDir := filepath.Join("cmd", "internal", "dashboard")
				if err := os.MkdirAll(dashboardDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				// Create template file
				templatePath := filepath.Join(dashboardDir, "template.html")
				templateContent := "<html><body>Test Template</body></html>"
				if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
					t.Fatalf("failed to write template file: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "loads template from secondary path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for secondary path
				dashboardDir := filepath.Join("internal", "dashboard")
				if err := os.MkdirAll(dashboardDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				// Create template file
				templatePath := filepath.Join(dashboardDir, "template.html")
				templateContent := "<html><body>Secondary Path Template</body></html>"
				if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
					t.Fatalf("failed to write template file: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "returns error when template not found",
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

			content, err := LoadTemplateContent()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tt.expectEmpty && content != "" {
					t.Errorf("expected empty content on error, got: %v", content)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectEmpty && content == "" {
				t.Errorf("expected non-empty content, got empty string")
			}

			if !tt.expectEmpty && content == "" {
				t.Errorf("expected non-empty content, got empty string")
			}
		})
	}
}
