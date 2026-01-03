package dashboard

import (
	"html/template"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

// ViewModel represents the data structure passed to HTML templates
type ViewModel struct {
	DashboardTitle                   string
	PageTitle                        string
	KeyMetrics                       []schema.KeyMetric
	HighlightMetrics                 []schema.HightlightMetric
	TotalArticles                    int
	ReadCount                        int
	UnreadCount                      int
	ReadRate                         float64
	AvgArticlesPerMonth              float64
	LastUpdated                      time.Time
	Sources                          []schema.SourceInfo
	Months                           []schema.MonthInfo
	Years                            []schema.YearInfo
	AllYears                         []string
	AllSources                       []string
	AllYearsJSON                     template.JS
	AllSourcesJSON                   template.JS
	YearChartLabels                  template.JS
	YearChartData                    template.JS
	MonthChartLabels                 template.JS
	MonthChartDatasets               template.JS
	MonthTotalData                   template.JS
	ReadUnreadByMonthJSON            template.JS
	ReadUnreadBySourceJSON           template.JS
	ReadUnreadByYearJSON             template.JS
	UnreadArticleAgeDistributionJSON template.JS
	UnreadByYearJSON                 template.JS
	TopOldestUnreadArticles          []schema.ArticleMeta
	EvolutionData                    schema.EvolutionData
}
