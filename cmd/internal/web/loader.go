package web

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
	"gopkg.in/yaml.v3"
)

// GetTemplatesDir finds the directory containing HTML templates
// It tries multiple path configurations to handle different execution contexts
func GetTemplatesDir() (string, error) {
	// Define canonical paths in priority order
	possibleDirs := []string{
		// When running from project root (most common during development)
		"cmd/internal/web/templates",
		// When binary is in cmd/web directory
		"internal/web/templates",
		// Fallback: explicit relative path construction
		filepath.Join(".", "cmd", "internal", "web", "templates"),
	}

	var cwd string
	if wd, err := os.Getwd(); err == nil {
		cwd = wd
	}

	// Try each path
	for _, dir := range possibleDirs {
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir, nil
		}
	}

	// Enhanced error message with debugging info
	return "", fmt.Errorf(
		"templates directory not found. Current working directory: %s. Tried paths: %v",
		cwd, possibleDirs,
	)
}

// findAndReadFile searches for a file in a list of possible relative paths and reads it.
func findAndReadFile(possiblePaths []string) ([]byte, string, error) {
	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return content, path, nil
		}
	}
	return nil, "", fmt.Errorf("file not found in any of the paths: %v", possiblePaths)
}

// LoadEvolutionData reads the evolution.yml file and parses it into EvolutionData struct
func LoadEvolutionData() (schema.EvolutionData, error) {
	possiblePaths := []string{
		"cmd/internal/web/content/evolution.yml",
		"internal/web/content/evolution.yml",
		filepath.Join(".", "cmd", "internal", "web", "content", "evolution.yml"),
	}

	var data schema.EvolutionData

	content, _, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("evolution.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("failed to parse evolution.yml: %w", err)
	}

	// Post-process descriptions into lines for each chapter's timeline
	for c := range data.Chapters {
		for i := range data.Chapters[c].Timeline {
			lines := strings.Split(strings.TrimSpace(data.Chapters[c].Timeline[i].Description), "\n")
			data.Chapters[c].Timeline[i].DescriptionLines = make([]string, 0, len(lines))
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Remove leading "- " if present
				line = strings.TrimPrefix(line, "- ")
				line = strings.TrimSpace(line)
				// Remove surrounding quotes if present
				if len(line) >= 2 && line[0] == '"' && line[len(line)-1] == '"' {
					line = line[1 : len(line)-1]
				}
				data.Chapters[c].Timeline[i].DescriptionLines = append(data.Chapters[c].Timeline[i].DescriptionLines, line)
			}
		}
	}

	return data, nil
}

// LoadLanding reads the landing.yml file and parses it into Landing struct
func LoadLanding() (schema.Landing, error) {
	possiblePaths := []string{
		"cmd/internal/web/content/landing.yml",
		"internal/web/content/landing.yml",
		filepath.Join(".", "cmd", "internal", "web", "content", "landing.yml"),
	}

	var data schema.Landing

	content, _, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.Landing{}, fmt.Errorf("landing.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.Landing{}, fmt.Errorf("failed to parse landing.yml: %w", err)
	}

	return data, nil
}
