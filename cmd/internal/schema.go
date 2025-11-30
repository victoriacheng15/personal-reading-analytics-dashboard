package internal

import "time"

type Metrics struct {
	TotalArticles       int                          `json:"total_articles"`
	BySource            map[string]int               `json:"by_source"`
	BySourceReadStatus  map[string][2]int            `json:"by_source_read_status"`
	ByYear              map[string]int               `json:"by_year"`
	ByMonth             map[string]int               `json:"by_month"`
	ByYearAndMonth      map[string]map[string]int    `json:"by_year_and_month"`               // year -> month -> count
	ByMonthAndSource    map[string]map[string][2]int `json:"by_month_and_source_read_status"` // month -> source -> [read, unread]
	ByCategory          map[string][2]int            `json:"by_category"`                     // category -> [read, unread]
	ByCategoryAndSource map[string]map[string][2]int `json:"by_category_and_source"`          // category -> source -> [read, unread]
	ReadUnreadTotals    [2]int                       `json:"read_unread_totals"`              // [read, unread]
	UnreadByMonth       map[string]int               `json:"unread_by_month"`
	UnreadByCategory    map[string]int               `json:"unread_by_category"`
	UnreadBySource      map[string]int               `json:"unread_by_source"`
	OldestUnreadArticle *ArticleMeta                 `json:"oldest_unread_article,omitempty"`
	SourceMetadata      map[string]SourceMeta        `json:"source_metadata"`
	ReadCount           int                          `json:"read_count"`
	UnreadCount         int                          `json:"unread_count"`
	ReadRate            float64                      `json:"read_rate"`
	AvgArticlesPerMonth float64                      `json:"avg_articles_per_month"`
	LastUpdated         time.Time                    `json:"last_updated"`
	Articles            []ArticleMeta                `json:"articles,omitempty"`
}

// ArticleMeta holds minimal info for backlog/unread analysis
type ArticleMeta struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	Link     string `json:"link"`
	Category string `json:"category"`
	Read     bool   `json:"read"`
}

// SourceMeta tracks when a source was added
type SourceMeta struct {
	Added string `json:"added"`
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
	Year    string
	Total   int
	Sources map[string]int
}

type YearInfo struct {
	Year  string
	Count int
}
