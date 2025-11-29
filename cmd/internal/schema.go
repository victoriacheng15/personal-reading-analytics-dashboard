package internal

import "time"

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

type SourceInfo struct {
	Name        string
	Count       int
	Read        int
	Unread      int
	ReadPct     float64
	AuthorCount int
}

type MonthInfo struct {
	Name    string
	Month   string
	Total   int
	Sources map[string]int
}

type YearInfo struct {
	Year  string
	Count int
}
