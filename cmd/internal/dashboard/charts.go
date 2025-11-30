package dashboard

import (
	"encoding/json"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// ChartDataset represents a single dataset for Chart.js
type ChartDataset struct {
	Label           string      `json:"label"`
	Data            interface{} `json:"data"`
	BackgroundColor interface{} `json:"backgroundColor,omitempty"`
	BorderColor     string      `json:"borderColor,omitempty"`
	BorderWidth     int         `json:"borderWidth,omitempty"`
}

// YearChartData holds prepared year chart data
type YearChartData struct {
	LabelsJSON json.RawMessage
	DataJSON   json.RawMessage
}

// MonthChartData holds prepared month chart data
type MonthChartData struct {
	LabelsJSON    json.RawMessage
	DatasetsJSON  json.RawMessage
	TotalDataJSON json.RawMessage
}

// PrepareYearChartData prepares year breakdown chart data
func PrepareYearChartData(years []schema.YearInfo) *YearChartData {
	labels := make([]string, 0)
	data := make([]int, 0)

	for _, year := range years {
		labels = append(labels, year.Year)
		data = append(data, year.Count)
	}

	labelsJSON, _ := json.Marshal(labels)
	dataJSON, _ := json.Marshal(data)

	return &YearChartData{
		LabelsJSON: labelsJSON,
		DataJSON:   dataJSON,
	}
}

// PrepareMonthChartData prepares month breakdown chart data with source stacking
func PrepareMonthChartData(months []schema.MonthInfo, sources []schema.SourceInfo) *MonthChartData {
	monthLabels := make([]string, 0)
	for _, month := range months {
		// Just use the month name for aggregated monthly view (no year)
		monthLabels = append(monthLabels, month.Name)
	}
	monthLabelsJSON, _ := json.Marshal(monthLabels)

	// Build datasets for each source
	sourceColors := map[string]string{
		"Substack":     "#667eea",
		"freeCodeCamp": "#764ba2",
		"GitHub":       "#f093fb",
		"Shopify":      "#4facfe",
		"Stripe":       "#00f2fe",
	}

	datasetsMap := make(map[string][]int)

	// Initialize all sources with data for each month
	for _, source := range sources {
		datasetsMap[source.Name] = make([]int, len(months))
	}

	// Populate data from month.Sources
	for monthIdx, month := range months {
		for sourceName, articleCount := range month.Sources {
			if _, exists := datasetsMap[sourceName]; exists {
				datasetsMap[sourceName][monthIdx] = articleCount
			}
		}
	}

	// Create Chart.js datasets
	var datasets []map[string]interface{}
	for _, source := range sources {
		if data, exists := datasetsMap[source.Name]; exists && len(data) > 0 {
			color := sourceColors[source.Name]
			if color == "" {
				color = "#" + colorHash(source.Name)
			}
			dataset := map[string]interface{}{
				"label":           source.Name,
				"data":            data,
				"backgroundColor": color,
				"borderColor":     "#2d3748",
				"borderWidth":     1,
			}
			datasets = append(datasets, dataset)
		}
	}

	datasetsJSON, _ := json.Marshal(datasets)

	// Prepare total data for months (for the line chart view)
	monthTotalData := make([]int, 0)
	for _, month := range months {
		monthTotalData = append(monthTotalData, month.Total)
	}
	monthTotalDataJSON, _ := json.Marshal(monthTotalData)

	return &MonthChartData{
		LabelsJSON:    monthLabelsJSON,
		DatasetsJSON:  datasetsJSON,
		TotalDataJSON: monthTotalDataJSON,
	}
}

// colorHash generates a simple hash for generating colors
func colorHash(s string) string {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return formatHex(h % 16777215)
}

// formatHex formats a number as a 6-digit hex string
func formatHex(n uint32) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		b[i] = hex[n%16]
		n /= 16
	}
	return string(b)
}
