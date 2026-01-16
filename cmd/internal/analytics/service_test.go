package analytics

import (
	"os"
	"path/filepath"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

func TestAnalyticsService_Generate(t *testing.T) {
	tests := []struct {
		name          string
		metrics       schema.Metrics
		expectSuccess bool
	}{
		{
			name: "generates html analytics with metrics",
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
			tmpDir, err := os.MkdirTemp("", "analytics_service_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// The service looks for templates relative to CWD or in specific paths.
			// For testing, we'll create a mock structure.
			templateDir := filepath.Join(tmpDir, "cmd", "internal", "analytics", "templates")
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Create required template files
			baseTmpl := `{{define "base"}}<html><body>{{block "content" .}}{{end}}</body></html>{{end}}`
			headerTmpl := `{{define "header"}}<header></header>{{end}}`
			footerTmpl := `{{define "footer"}}<footer></footer>{{end}}`
			indexTmpl := `{{define "content"}}{{template "header" .}}<h1>Home</h1>{{template "footer" .}}{{end}}{{template "base" .}}`
			analyticsTmpl := `{{define "content"}}{{template "header" .}}<h1>Analytics</h1>{{template "footer" .}}{{end}}{{template "base" .}}`
			evolutionTmpl := `{{define "content"}}{{template "header" .}}<h1>Evolution</h1>{{template "footer" .}}{{end}}{{template "base" .}}`

			templates := map[string]string{
				"base.html":      baseTmpl,
				"header.html":    headerTmpl,
				"footer.html":    footerTmpl,
				"index.html":     indexTmpl,
				"analytics.html": analyticsTmpl,
				"evolution.html": evolutionTmpl,
			}

			for name, content := range templates {
				if err := os.WriteFile(filepath.Join(templateDir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to create template %s: %v", name, err)
				}
			}

			// Mock evolution.yml
			evolutionData := `events: []`
			contentDir := filepath.Join(tmpDir, "cmd", "internal", "analytics", "content")
			if err := os.MkdirAll(contentDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(contentDir, "evolution.yml"), []byte(evolutionData), 0644); err != nil {
				t.Fatal(err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			service := NewAnalyticsService("site")
			err = service.Generate(tt.metrics)
			if (err == nil) != tt.expectSuccess {
				t.Errorf("AnalyticsService.Generate() error = %v, expectSuccess %v", err, tt.expectSuccess)
			}

			if _, err := os.Stat("site/index.html"); os.IsNotExist(err) {
				t.Error("site/index.html was not created")
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		setup     func(t *testing.T, src, dst string)
		expectErr bool
	}{
		{
			name:    "successfully copies file",
			content: "hello world",
			setup: func(t *testing.T, src, dst string) {
				// Normal setup
			},
			expectErr: false,
		},
		{
			name:    "source does not exist",
			content: "",
			setup: func(t *testing.T, src, dst string) {
				os.Remove(src) // Ensure source is missing
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcPath := filepath.Join(tmpDir, "source.txt")
			dstPath := filepath.Join(tmpDir, "dest.txt")

			if tt.content != "" {
				if err := os.WriteFile(srcPath, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			tt.setup(t, srcPath, dstPath)

			err := copyFile(srcPath, dstPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("copyFile() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr {
				content, err := os.ReadFile(dstPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src string)
		expectErr bool
		verify    func(t *testing.T, dst string)
	}{
		{
			name: "recursively copies directory",
			setup: func(t *testing.T, src string) {
				// Create file in root
				os.WriteFile(filepath.Join(src, "root.txt"), []byte("root"), 0644)
				// Create subdir
				subDir := filepath.Join(src, "subdir")
				os.Mkdir(subDir, 0755)
				// Create file in subdir
				os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("sub"), 0644)
			},
			expectErr: false,
			verify: func(t *testing.T, dst string) {
				// Check root file
				if _, err := os.Stat(filepath.Join(dst, "root.txt")); os.IsNotExist(err) {
					t.Error("root.txt not copied")
				}
				// Check subdir
				if _, err := os.Stat(filepath.Join(dst, "subdir")); os.IsNotExist(err) {
					t.Error("subdir not copied")
				}
				// Check subdir file
				if _, err := os.Stat(filepath.Join(dst, "subdir", "sub.txt")); os.IsNotExist(err) {
					t.Error("sub.txt not copied")
				}
			},
		},
		{
			name: "source does not exist",
			setup: func(t *testing.T, src string) {
				os.RemoveAll(src)
			},
			expectErr: true,
			verify:    func(t *testing.T, dst string) {},
		},
		{
			name: "source is a file",
			setup: func(t *testing.T, src string) {
				os.RemoveAll(src)
				os.WriteFile(src, []byte("file"), 0644)
			},
			expectErr: true,
			verify:    func(t *testing.T, dst string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcDir := filepath.Join(tmpDir, "src")
			dstDir := filepath.Join(tmpDir, "dst")

			// Create src dir by default (tests can remove it)
			os.Mkdir(srcDir, 0755)

			tt.setup(t, srcDir)

			err := copyDir(srcDir, dstDir)
			if (err != nil) != tt.expectErr {
				t.Errorf("copyDir() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr {
				tt.verify(t, dstDir)
			}
		})
	}
}
