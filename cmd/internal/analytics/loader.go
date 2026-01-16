package analytics

import (
	"fmt"
	"log"
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
		"cmd/internal/analytics/templates",
		// When binary is in cmd/analytics directory
		"internal/analytics/templates",
		// Fallback: explicit relative path construction
		filepath.Join(".", "cmd", "internal", "analytics", "templates"),
	}

	var cwd string
	if wd, err := os.Getwd(); err == nil {
		cwd = wd
	}

	// Try each path
	for _, dir := range possibleDirs {
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			log.Printf("✅ Found templates directory: %s\n", dir)
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
		"cmd/internal/analytics/content/evolution.yml",
		"internal/analytics/content/evolution.yml",
		filepath.Join(".", "cmd", "internal", "analytics", "content", "evolution.yml"),
	}

	var data schema.EvolutionData

	content, path, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("evolution.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("failed to parse evolution.yml: %w", err)
	}

	// Post-process descriptions into lines
	for i := range data.Events {
		lines := strings.Split(strings.TrimSpace(data.Events[i].Description), "\n")
		data.Events[i].DescriptionLines = make([]string, 0, len(lines))
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
			data.Events[i].DescriptionLines = append(data.Events[i].DescriptionLines, line)
		}
	}

	log.Printf("✅ Loaded evolution data from: %s\n", path)
	return data, nil
}

// LoadIndexContent reads the index.yml file and parses it into IndexContent struct
func LoadIndexContent() (schema.IndexContent, error) {
	possiblePaths := []string{
		"cmd/internal/analytics/content/index.yml",
		"internal/analytics/content/index.yml",
		filepath.Join(".", "cmd", "internal", "analytics", "content", "index.yml"),
	}

	var data schema.IndexContent

	content, path, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.IndexContent{}, fmt.Errorf("index.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.IndexContent{}, fmt.Errorf("failed to parse index.yml: %w", err)
	}

	log.Printf("✅ Loaded index content from: %s\n", path)
	return data, nil
}
