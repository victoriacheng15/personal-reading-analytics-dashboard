package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Metrics represents the calculated reading analytics
type Metrics struct {
	TotalArticles       int                       `json:"total_articles"`
	BySource            map[string]int            `json:"by_source"`
	BySourceReadStatus  map[string][2]int         `json:"by_source_read_status"`
	ByYear              map[string]int            `json:"by_year"`
	ByMonthOnly         map[string]int            `json:"by_month"`
	ByMonthAndSource    map[string]map[string]int `json:"by_month_and_source"`
	ReadCount           int                       `json:"read_count"`
	UnreadCount         int                       `json:"unread_count"`
	ReadRate            float64                   `json:"read_rate"`
	AvgArticlesPerMonth float64                   `json:"avg_articles_per_month"`
	LastUpdated         time.Time                 `json:"last_updated"`
}

// SourceInfo represents statistics for a single source
type SourceInfo struct {
	Name        string
	Count       int
	Read        int
	Unread      int
	ReadPct     float64
	AuthorCount int
}

// MonthInfo represents aggregated data for a month
type MonthInfo struct {
	Name    string
	Month   string
	Total   int
	Sources map[string]int
}

// YearInfo represents aggregated data for a year
type YearInfo struct {
	Year  string
	Count int
}

const (
	articlesCols   = 5  // Expected number of columns: date, title, link, category, read
	monthsInPeriod = 36 // 3 years of data for average calculation
	dashboardTitle = "📚 Personal Reading Analytics"
)

var monthNames = []string{"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

var sourceColors = map[string]string{
	"Substack":     "#667eea",
	"freeCodeCamp": "#764ba2",
	"GitHub":       "#f093fb",
	"Shopify":      "#4facfe",
	"Stripe":       "#00f2fe",
}

// normalizeSourceName converts source names to proper capitalization
func normalizeSourceName(name string) string {
	sourceMap := map[string]string{
		"substack":     "Substack",
		"freecodecamp": "freeCodeCamp",
		"github":       "GitHub",
		"shopify":      "Shopify",
		"stripe":       "Stripe",
	}

	// Convert to lowercase for comparison
	lower := strings.ToLower(name)

	// Return normalized name if found, otherwise return original
	if normalized, exists := sourceMap[lower]; exists {
		return normalized
	}
	return name
}

// hash generates a simple hash for generating colors
func hash(s string) uint32 {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return h
}

// fetchMetricsFromSheets retrieves and calculates metrics from Google Sheets

func fetchMetricsFromSheets(ctx context.Context, spreadsheetID, credentialsPath string) (Metrics, error) {
	// Create Sheets service
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to create sheets client: %w", err)
	}

	// Get all sheets to find sheet names
	spreadsheet, err := client.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to retrieve spreadsheet: %w", err)
	}

	// Find Articles and Providers sheets
	articlesSheet := "articles"
	providersSheet := "providers"
	for _, sheet := range spreadsheet.Sheets {
		title := sheet.Properties.Title
		if title == "Articles" || title == "articles" {
			articlesSheet = title
		}
		if title == "Providers" || title == "providers" {
			providersSheet = title
		}
	}

	// Count Substack providers
	substackCount := 0
	readRange := fmt.Sprintf("%s!A:B", providersSheet)
	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err == nil && len(resp.Values) > 0 {
		// Assuming column A has provider name, count rows with "substack"
		for i := 1; i < len(resp.Values); i++ {
			if len(resp.Values[i]) > 0 {
				provider := fmt.Sprintf("%v", resp.Values[i][0])
				if provider == "substack" || provider == "Substack" {
					substackCount++
				}
			}
		}
	}

	// Read all articles data
	readRange = fmt.Sprintf("%s!A:E", articlesSheet)
	resp, err = client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return Metrics{}, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return Metrics{}, fmt.Errorf("no data found in sheet")
	}

	// Parse articles: columns are date, title, link, category, read?
	metrics := Metrics{
		BySource:           make(map[string]int),
		BySourceReadStatus: make(map[string][2]int),
		ByYear:             make(map[string]int),
		ByMonthOnly:        make(map[string]int),
		ByMonthAndSource:   make(map[string]map[string]int),
	}

	// Skip header row (row 0)
	for i := 1; i < len(resp.Values); i++ {
		row := resp.Values[i]
		if len(row) < 5 {
			continue // Skip incomplete rows
		}

		metrics.TotalArticles++

		// Column A: date (YYYY-MM-DD format)
		if len(row) > 0 {
			dateStr := fmt.Sprintf("%v", row[0])
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				year := t.Format("2006")
				month := t.Format("01")
				metrics.ByYear[year]++
				metrics.ByMonthOnly[month]++

				// Track by month and source
				if len(row) > 3 {
					category := normalizeSourceName(fmt.Sprintf("%v", row[3]))
					if metrics.ByMonthAndSource[month] == nil {
						metrics.ByMonthAndSource[month] = make(map[string]int)
					}
					metrics.ByMonthAndSource[month][category]++
				}
			}
		}

		// Column D: category (source)
		var category string
		if len(row) > 3 {
			category = normalizeSourceName(fmt.Sprintf("%v", row[3]))
			metrics.BySource[category]++
		}

		// Column E: read? (checkbox - TRUE/FALSE)
		isRead := false
		if len(row) > 4 {
			readStatus := fmt.Sprintf("%v", row[4])
			// Checkbox returns TRUE or FALSE (case-insensitive)
			if readStatus == "TRUE" || readStatus == "true" {
				metrics.ReadCount++
				isRead = true
			} else {
				metrics.UnreadCount++
			}
		}

		// Track read/unread by source
		if category != "" {
			status := metrics.BySourceReadStatus[category]
			if isRead {
				status[0]++ // read
			} else {
				status[1]++ // unread
			}
			metrics.BySourceReadStatus[category] = status
		}
	}

	// Calculate derived metrics
	if metrics.TotalArticles > 0 {
		metrics.ReadRate = (float64(metrics.ReadCount) / float64(metrics.TotalArticles)) * 100
	}
	// Assume 36 months (3 years of data)
	metrics.AvgArticlesPerMonth = float64(metrics.TotalArticles) / 36

	// Store substack count for later use in display
	metrics.BySourceReadStatus["substack_author_count"] = [2]int{substackCount, 0}

	// Set timestamp
	metrics.LastUpdated = time.Now()

	return metrics, nil
}

func main() {
	ctx := context.Background()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, will use environment variables")
	}

	sheetID := os.Getenv("SHEET_ID")
	credentialsPath := os.Getenv("CREDENTIALS_PATH")

	if sheetID == "" {
		log.Fatal("SHEET_ID environment variable is required")
	}
	if credentialsPath == "" {
		credentialsPath = "./credentials.json"
	}

	metrics, err := fetchMetricsFromSheets(ctx, sheetID, credentialsPath)
	if err != nil {
		log.Fatalf("Failed to fetch metrics: %v", err)
	}

	// Save metrics as JSON with timestamp
	os.MkdirAll("metrics", 0755)

	metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	// Save to metrics folder with date filename (YYYY-MM-DD.json)
	dateFilename := metrics.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)
	err = os.WriteFile(metricsFilePath, metricsJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write metrics file: %v", err)
	}

	// Generate HTML dashboard
	generateHTMLDashboard(metrics)

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	log.Println("✅ Successfully processed metrics from Google Sheets")
}

// generateHTMLDashboard creates and saves the HTML dashboard file
func generateHTMLDashboard(metrics Metrics) {
	// Sort sources by count
	var sources []SourceInfo
	for name, count := range metrics.BySource {
		readStatus := metrics.BySourceReadStatus[name]
		read := readStatus[0]
		unread := readStatus[1]
		readPct := 0.0
		if count > 0 {
			readPct = (float64(read) / float64(count)) * 100
		}

		authorCount := 0
		if name == "Substack" {
			authorCount = metrics.BySourceReadStatus["substack_author_count"][0]
		}

		sources = append(sources, SourceInfo{
			Name:        name,
			Count:       count,
			Read:        read,
			Unread:      unread,
			ReadPct:     readPct,
			AuthorCount: authorCount,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	// Build month info
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	var months []MonthInfo
	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		if total, exists := metrics.ByMonthOnly[monthStr]; exists && total > 0 {
			months = append(months, MonthInfo{
				Name:    monthNames[month],
				Month:   monthStr,
				Total:   total,
				Sources: metrics.ByMonthAndSource[monthStr],
			})
		}
	}

	// Build year info
	var years []YearInfo
	for year, count := range metrics.ByYear {
		years = append(years, YearInfo{Year: year, Count: count})
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Year > years[j].Year
	})

	// HTML template with semantic HTML
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.DashboardTitle}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', sans-serif;
            background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
            min-height: 100vh;
            padding: 2rem 1rem;
        }
        
        main {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 3rem;
        }
        
        header {
            margin-bottom: 2rem;
        }
        
        h1 {
            color: #2d3748;
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
        }
        
        time {
            color: #718096;
            font-size: 0.95rem;
            display: block;
            margin-bottom: 2rem;
        }
        
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0 3rem 0;
        }
        
        article.metric-card {
            background: linear-gradient(135deg, #0369a1 0%, #0284c7 100%);
            padding: 1.5rem;
            border-radius: 12px;
            text-align: center;
            border: 2px solid #0ea5e9;
            color: white;
        }
        
        article.metric-card h2 {
            color: white;
        }
        
        article.metric-card.highlight {
            background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
            color: white;
        }
        
        article.metric-card h2 {
            font-size: 0.85rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            opacity: 0.8;
            margin-bottom: 0.5rem;
        }
        
        .metric-value {
            font-size: 2.2rem;
            font-weight: 700;
        }
        
        section {
            margin: 3rem 0;
        }
        
        section > h2 {
            font-size: 1.4rem;
            color: #2d3748;
            margin-bottom: 1.5rem;
            font-weight: 600;
            border-bottom: 3px solid #0369a1;
            padding-bottom: 0.5rem;
            display: inline-block;
        }
        
        .sources-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
        }
        
        article.source-card {
            background: #f7fafc;
            border-left: 5px solid #0369a1;
            padding: 1.5rem;
            border-radius: 8px;
        }
        
        article.source-card h3 {
            font-size: 1.1rem;
            font-weight: 600;
            color: #2d3748;
            margin-bottom: 0.75rem;
        }
        
        .source-stats {
            font-size: 0.9rem;
            color: #718096;
            line-height: 1.6;
        }
        
        .stat-line {
            display: flex;
            justify-content: space-between;
            margin-bottom: 0.4rem;
        }
        
        .months-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
        }
        
        article.month-card {
            background: #f7fafc;
            padding: 1rem;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
        }
        
        article.month-card h3 {
            font-weight: 600;
            color: #2d3748;
            margin-bottom: 0.8rem;
        }
        
        .month-sources {
            font-size: 0.85rem;
            color: #4a5568;
            line-height: 1.5;
        }
        
        .year-breakdown {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
        }
        
        .year-badge {
            background: #edf2f7;
            padding: 0.75rem 1.25rem;
            border-radius: 20px;
            font-size: 0.9rem;
            font-weight: 600;
            color: #2d3748;
        }
        
        footer {
            margin-top: 3rem;
            padding-top: 2rem;
            border-top: 2px solid #e2e8f0;
            text-align: center;
            color: #718096;
            font-size: 0.9rem;
        }
        
        @media (max-width: 768px) {
            main { padding: 1.5rem; }
            h1 { font-size: 1.8rem; }
            .metrics-grid { grid-template-columns: repeat(2, 1fr); }
        }
        .chart-container {
            position: relative;
            width: 100%;
            margin: 2rem 0;
            padding: 1.5rem;
            background: #f7fafc;
            border-radius: 12px;
            border: 1px solid #e2e8f0;
        }
        
        .chart-wrapper {
            position: relative;
            height: 400px;
        }
        
        @media (max-width: 1024px) {
            .chart-wrapper {
                height: 300px;
            }
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
</head>
<body>
    <main>
        <header>
            <h1>{{.DashboardTitle}}</h1>
            <time>Last updated: {{.LastUpdated.Format "Jan 02, 2006 at 3:04 PM"}}</time>
        </header>
        
        <section aria-label="Key Metrics">
            <div class="metrics-grid">
                <article class="metric-card">
                    <h2>Total Articles</h2>
                    <div class="metric-value">{{.TotalArticles}}</div>
                </article>
                <article class="metric-card highlight">
                    <h2>Read Rate</h2>
                    <div class="metric-value">{{printf "%.1f" .ReadRate}}%</div>
                </article>
                <article class="metric-card">
                    <h2>Read</h2>
                    <div class="metric-value">{{.ReadCount}}</div>
                </article>
                <article class="metric-card">
                    <h2>Unread</h2>
                    <div class="metric-value">{{.UnreadCount}}</div>
                </article>
                <article class="metric-card">
                    <h2>Avg/Month</h2>
                    <div class="metric-value">{{printf "%.0f" .AvgArticlesPerMonth}}</div>
                </article>
            </div>
        </section>
        
        <section aria-label="Sources">
            <h2>📌 Sources</h2>
            <div class="sources-grid">
                {{range .Sources}}
                <article class="source-card">
                    <h3>{{.Name}}</h3>
                    <div class="source-stats">
                        <div class="stat-line">
                            <span>Total:</span>
                            <strong>{{.Count}}</strong>
                        </div>
                        <div class="stat-line">
                            <span>Read:</span>
                            <strong>{{.Read}} ({{printf "%.1f" .ReadPct}}%)</strong>
                        </div>
                        <div class="stat-line">
                            <span>Unread:</span>
                            <strong>{{.Unread}}</strong>
                        </div>
                        {{if gt .AuthorCount 0}}
                        <div class="stat-line" style="margin-top: 0.5rem; padding-top: 0.5rem; border-top: 1px solid #cbd5e0; font-size: 0.8rem; opacity: 0.7;">
                            <span>Per author:</span>
                            <strong>{{printf "%.0f" (divideFloat .Count .AuthorCount)}} articles</strong>
                        </div>
                        {{end}}
                    </div>
                </article>
                {{end}}
            </div>
        </section>
        
        <section aria-label="Yearly Breakdown">
            <h2>📅 By Year</h2>
            <div class="chart-container">
                <div class="chart-wrapper">
                    <canvas id="yearChart"></canvas>
                </div>
            </div>
        </section>
        
        <section aria-label="Monthly Breakdown">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                <h2>📊 Monthly Breakdown</h2>
                <select id="monthViewToggle" style="padding: 0.5rem 1rem; border-radius: 6px; border: 2px solid #667eea; background: white; color: #2d3748; font-weight: 600; cursor: pointer; font-size: 0.95rem;">
                    <option value="total">Total Articles</option>
                    <option value="stacked">By Source</option>
                </select>
            </div>
            <div class="chart-container">
                <div class="chart-wrapper">
                    <canvas id="monthChart"></canvas>
                </div>
            </div>
        </section>
        
        <footer>
            <p>📈 Automatically generated from Google Sheets • Updated daily via GitHub Actions</p>
        </footer>
    </main>
    
    <script>
        // Year chart
        const yearCtx = document.getElementById('yearChart').getContext('2d');
        new Chart(yearCtx, {
            type: 'bar',
            data: {
                labels: {{.YearChartLabels}},
                datasets: [{
                    label: 'Articles by Year',
                    data: {{.YearChartData}},
                    backgroundColor: [
                        '#667eea',
                        '#764ba2',
                        '#f093fb',
                        '#4facfe',
                        '#00f2fe',
                        '#43e97b',
                        '#fa709a'
                    ],
                    borderColor: '#2d3748',
                    borderWidth: 2,
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            font: { size: 12 }
                        }
                    },
                    x: {
                        ticks: {
                            font: { size: 12 }
                        }
                    }
                }
            }
        });
        
        // Month chart data
        const monthChartLabels = {{.MonthChartLabels}};
        const monthChartDatasets = {{.MonthChartDatasets}};
        const monthTotalData = {{.MonthTotalData}};
        
        let monthChart = null;
        
        function updateMonthChart(view) {
            if (monthChart) {
                monthChart.destroy();
            }
            
            const monthCtx = document.getElementById('monthChart').getContext('2d');
            
            if (view === 'total') {
                monthChart = new Chart(monthCtx, {
                    type: 'line',
                    data: {
                        labels: monthChartLabels,
                        datasets: [{
                            label: 'Total Articles',
                            data: monthTotalData,
                            borderColor: '#667eea',
                            backgroundColor: 'rgba(102, 126, 234, 0.1)',
                            borderWidth: 3,
                            fill: true,
                            tension: 0.4,
                            pointRadius: 5,
                            pointBackgroundColor: '#764ba2',
                            pointBorderColor: '#fff',
                            pointBorderWidth: 2,
                            pointHoverRadius: 7
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                display: true,
                                labels: {
                                    font: { size: 12 },
                                    usePointStyle: true
                                }
                            }
                        },
                        scales: {
                            y: {
                                beginAtZero: true,
                                ticks: {
                                    font: { size: 12 }
                                }
                            },
                            x: {
                                ticks: {
                                    font: { size: 11 }
                                }
                            }
                        }
                    }
                });
            } else {
                monthChart = new Chart(monthCtx, {
                    type: 'bar',
                    data: {
                        labels: monthChartLabels,
                        datasets: monthChartDatasets
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            x: {
                                stacked: true,
                                ticks: {
                                    font: { size: 11 }
                                }
                            },
                            y: {
                                stacked: true,
                                beginAtZero: true,
                                ticks: {
                                    font: { size: 12 }
                                }
                            }
                        },
                        plugins: {
                            legend: {
                                display: true,
                                labels: {
                                    font: { size: 12 },
                                    usePointStyle: true
                                }
                            }
                        }
                    }
                });
            }
        }
        
        // Initialize with total view
        updateMonthChart('total');
        
        // Add event listener to dropdown
        document.getElementById('monthViewToggle').addEventListener('change', function(e) {
            updateMonthChart(e.target.value);
        });
    </script>
</body>
</html>`

	// Parse and execute template
	tmpl := template.New("dashboard")
	tmpl.Funcs(template.FuncMap{
		"divideFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
	})

	tmpl, err := tmpl.Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Failed to parse HTML template: %v", err)
	}

	// Create site directory
	os.MkdirAll("site", 0755)

	// Create output file
	file, err := os.Create("site/index.html")
	if err != nil {
		log.Fatalf("Failed to create site/index.html: %v", err)
	}
	defer file.Close()

	// Prepare chart data for years
	yearLabels := make([]string, 0)
	yearData := make([]int, 0)
	for _, year := range years {
		yearLabels = append(yearLabels, year.Year)
		yearData = append(yearData, year.Count)
	}
	yearLabelsJSON, _ := json.Marshal(yearLabels)
	yearDataJSON, _ := json.Marshal(yearData)

	// Prepare chart data for months
	monthLabels := make([]string, 0)
	for _, month := range months {
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
	for _, month := range months {
		for source, count := range month.Sources {
			if _, exists := datasetsMap[source]; !exists {
				datasetsMap[source] = make([]int, 0)
			}
			datasetsMap[source] = append(datasetsMap[source], count)
		}
	}

	// Ensure all sources have data for all months (fill with 0)
	for source := range datasetsMap {
		if len(datasetsMap[source]) < len(months) {
			datasetsMap[source] = append(datasetsMap[source], make([]int, len(months)-len(datasetsMap[source]))...)
		}
	}

	// Create Chart.js datasets
	var datasets []map[string]interface{}
	for _, source := range sources {
		if data, exists := datasetsMap[source.Name]; exists && len(data) > 0 {
			color := sourceColors[source.Name]
			if color == "" {
				color = "#" + fmt.Sprintf("%06x", int64(hash(source.Name))%16777215)
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

	// Execute template
	data := map[string]interface{}{
		"DashboardTitle":      dashboardTitle,
		"TotalArticles":       metrics.TotalArticles,
		"ReadCount":           metrics.ReadCount,
		"UnreadCount":         metrics.UnreadCount,
		"ReadRate":            metrics.ReadRate,
		"AvgArticlesPerMonth": metrics.AvgArticlesPerMonth,
		"LastUpdated":         metrics.LastUpdated,
		"Sources":             sources,
		"Months":              months,
		"Years":               years,
		"YearChartLabels":     template.JS(yearLabelsJSON),
		"YearChartData":       template.JS(yearDataJSON),
		"MonthChartLabels":    template.JS(monthLabelsJSON),
		"MonthChartDatasets":  template.JS(datasetsJSON),
		"MonthTotalData":      template.JS(monthTotalDataJSON),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Println("✅ HTML dashboard generated at site/index.html")
}
