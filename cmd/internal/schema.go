package internal

import "time"

type Metrics struct {
	TotalArticles                int                          `json:"total_articles"`
	BySource                     map[string]int               `json:"by_source"`
	BySourceReadStatus           map[string][2]int            `json:"by_source_read_status"`
	ByYear                       map[string]int               `json:"by_year"`
	ByMonth                      map[string]int               `json:"by_month"`
	ByYearAndMonth               map[string]map[string]int    `json:"by_year_and_month"`               // year -> month -> count
	ByMonthAndSource             map[string]map[string][2]int `json:"by_month_and_source_read_status"` // month -> source -> [read, unread]
	ByCategory                   map[string][2]int            `json:"by_category"`                     // category -> [read, unread]
	ByCategoryAndSource          map[string]map[string][2]int `json:"by_category_and_source"`          // category -> source -> [read, unread]
	ReadUnreadTotals             [2]int                       `json:"read_unread_totals"`              // [read, unread]
	UnreadByMonth                map[string]int               `json:"unread_by_month"`
	UnreadByCategory             map[string]int               `json:"unread_by_category"`
	UnreadBySource               map[string]int               `json:"unread_by_source"`
	UnreadByYear                 map[string]int               `json:"unread_by_year"`
	UnreadArticleAgeDistribution map[string]int               `json:"unread_article_age_distribution"`
	OldestUnreadArticle          *ArticleMeta                 `json:"oldest_unread_article,omitempty"`
	TopOldestUnreadArticles      []ArticleMeta                `json:"top_oldest_unread_articles,omitempty"`
	SourceMetadata               map[string]SourceMeta        `json:"source_metadata"`
	ReadCount                    int                          `json:"read_count"`
	UnreadCount                  int                          `json:"unread_count"`
	ReadRate                     float64                      `json:"read_rate"`
	AvgArticlesPerMonth          float64                      `json:"avg_articles_per_month"`
	LastUpdated                  time.Time                    `json:"last_updated"`
	AISummary                    string                       `json:"ai_summary,omitempty"`
	Articles                     []ArticleMeta                `json:"articles,omitempty"`
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

type KeyMetric struct {
	Title string
	Value string
}

type HightlightMetric struct {
	Title string
	Value string
}

type EvolutionData struct {
	Chapters []Chapter `yaml:"chapters"`
}

type Chapter struct {
	Title    string      `yaml:"title"`
	Period   string      `yaml:"period"`
	Intro    string      `yaml:"intro"`
	Timeline []Milestone `yaml:"timeline"`
}

type Artifact struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Milestone struct {
	Date             string     `yaml:"date"`
	Title            string     `yaml:"title"`
	Artifacts        []Artifact `yaml:"artifacts,omitempty"`
	Description      string     `yaml:"description"`
	DescriptionLines []string   `yaml:"-"`
}

type IndexContent struct {
	IntroSection                 IntroSection                 `yaml:"intro_section"`
	OriginStorySection           OriginStorySection           `yaml:"origin_story_section"`
	EngineeringPrinciplesSection EngineeringPrinciplesSection `yaml:"engineering_principles_section"`
}

type IntroSection struct {
	Heading    string      `yaml:"heading"`
	CTAButtons []CTAButton `yaml:"cta_buttons"`
}

type CTAButton struct {
	Text string `yaml:"text"`
	URL  string `yaml:"url"`
}

type OriginStorySection struct {
	Title      string   `yaml:"title"`
	Paragraphs []string `yaml:"paragraphs"`
}

type EngineeringPrinciplesSection struct {
	Title      string      `yaml:"title"`
	Principles []Principle `yaml:"principles"`
}

type Principle struct {
	Icon        string `yaml:"icon"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
