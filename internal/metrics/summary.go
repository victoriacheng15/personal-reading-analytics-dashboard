package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/victoriacheng15/personal-reading-analytics/internal"
	"github.com/victoriacheng15/personal-reading-analytics/internal/ai"
)

// GenerateAndSaveDeltaAnalysis generates an AI delta analysis comparing the current metrics with the previous week's.
func GenerateAndSaveDeltaAnalysis(ctx context.Context, metricsDir string, currentFilename string, currentMetrics *internal.Metrics) error {
	prevMetrics, err := loadPreviousMetrics(metricsDir, currentFilename)
	if err != nil {
		// Log warning but don't fail, just return.
		// In a real logger we'd log this. For now printing to stderr is acceptable for CLI.
		fmt.Fprintf(os.Stderr, "Warning: Could not load previous metrics for comparison: %v\n", err)
	}

	prompt := constructPrompt(currentMetrics, prevMetrics)

	client, err := ai.NewClient(ctx)
	if err != nil {
		// If client init fails (e.g. no key), we silently skip delta analysis
		fmt.Fprintf(os.Stderr, "Skipping AI delta analysis: %v\n", err)
		return nil
	}
	defer client.Close()

	deltaAnalysis, err := client.GenerateContent(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating AI delta analysis: %v\n", err)
		currentMetrics.AIDeltaAnalysis = "AI delta analysis unavailable at this time."
	} else {
		currentMetrics.AIDeltaAnalysis = deltaAnalysis
	}

	// Save the updated metrics back to the file
	return saveMetrics(metricsDir, currentFilename, currentMetrics)
}

func loadPreviousMetrics(dir, currentFilename string) (*internal.Metrics, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var jsonFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") && f.Name() != ".gitkeep" {
			jsonFiles = append(jsonFiles, f.Name())
		}
	}

	sort.Strings(jsonFiles)

	// Find the index of the current file
	currentIndex := -1
	for i, f := range jsonFiles {
		if f == currentFilename {
			currentIndex = i
			break
		}
	}

	if currentIndex <= 0 {
		return nil, fmt.Errorf("no previous metrics file found before %s", currentFilename)
	}

	prevFilename := jsonFiles[currentIndex-1]
	content, err := os.ReadFile(filepath.Join(dir, prevFilename))
	if err != nil {
		return nil, err
	}

	var metrics internal.Metrics
	if err := json.Unmarshal(content, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

func saveMetrics(dir, filename string, metrics *internal.Metrics) error {
	path := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func constructPrompt(curr, prev *internal.Metrics) string {
	currJSON, _ := json.MarshalIndent(curr, "", "  ")

	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are a personal reading analytics assistant. Analyze the user's reading habits.\n\n")

	if prev != nil {
		prevJSON, _ := json.MarshalIndent(prev, "", "  ")
		promptBuilder.WriteString("Compare the following two JSON metrics files (Previous vs Current):\n\n")
		promptBuilder.WriteString("PREVIOUS WEEK:\n")
		promptBuilder.Write(prevJSON)
		promptBuilder.WriteString("\n\nCURRENT WEEK:\n")
		promptBuilder.Write(currJSON)
		promptBuilder.WriteString("\n\n")
		promptBuilder.WriteString("Provide a concise, qualitative delta analysis (2-3 sentences) of the changes. ")
		promptBuilder.WriteString("Focus on these three dimensions: ")
		promptBuilder.WriteString("1. Velocity: Changes in reading pace or read rate. ")
		promptBuilder.WriteString("2. Backlog Health: Whether you are clearing old debt (items older than 1 year) or adding new unread noise. ")
		promptBuilder.WriteString("3. Chronology: The specific years of content you focused on reading this week. ")
		promptBuilder.WriteString("Do not mention source names (like Substack). Interpret the trends into a narrative. ")
		promptBuilder.WriteString("Do not use personal pronouns like 'you' or 'your'; maintain an objective, third-person perspective. ")
		promptBuilder.WriteString("IMPORTANT: Provide the delta analysis in plain text only. Do not use any markdown formatting (no bolding, no italics, no bullet points, no headers).")
	} else {
		promptBuilder.WriteString("Analyze the following reading metrics:\n\n")
		promptBuilder.Write(currJSON)
		promptBuilder.WriteString("\n\n")
		promptBuilder.WriteString("Provide a concise delta analysis (2-3 sentences) of the reading profile. ")
		promptBuilder.WriteString("Focus on reading velocity, backlog age distribution, and the chronological era of the collection. ")
		promptBuilder.WriteString("Do not use personal pronouns like 'you' or 'your'; maintain an objective, third-person perspective. ")
		promptBuilder.WriteString("IMPORTANT: Provide the analysis in plain text only. Do not use any markdown formatting.")
	}

	return promptBuilder.String()
}
