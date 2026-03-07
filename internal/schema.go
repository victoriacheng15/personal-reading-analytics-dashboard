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
	AIDeltaAnalysis              string                       `json:"ai_delta_analysis,omitempty"`
}

// ArticleMeta holds minimal info for backlog/unread analysis
type ArticleMeta struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	Link     string `json:"link"`
	Category string `json:"category"`
	Read     bool   `json:"read"`
}

// SourceMeta tracks when a source was added and its brand color
type SourceMeta struct {
	Added string `json:"added"`
	Color string `json:"color"`
}

type SourceInfo struct {
	Name        string
	Count       int
	Read        int
	Unread      int
	ReadPct     float64
	AuthorCount int
	Color       string
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

type Author struct {
	Name     string `yaml:"name"`
	GitHub   string `yaml:"github"`
	LinkedIn string `yaml:"linkedin"`
}

type Landing struct {
	Header        Header              `yaml:"header"`
	SystemSpec    SystemSpecification `yaml:"system_specification"`
	Hero          Hero                `yaml:"hero"`
	WhatIsReading WhatIsReading       `yaml:"what_is_reading_analytics"`
	KeyFeatures   KeyFeatures         `yaml:"key_features"`
	WhyItMatters  WhyItMatters        `yaml:"why_it_matters"`
	Footer        LandingFooter       `yaml:"footer"`
}

type Header struct {
	ProjectName string `yaml:"project_name"`
	SiteURL     string `yaml:"site_url"`
}

type SystemSpecification struct {
	Objective           string `yaml:"objective"`
	Stack               string `yaml:"stack"`
	Pattern             string `yaml:"pattern"`
	EntryPoint          string `yaml:"entry_point"`
	PersistenceStrategy string `yaml:"persistence_strategy"`
	Observability       string `yaml:"observability"`
	MachineRegistry     string `yaml:"machine_registry"`
}

type Hero struct {
	Headline         string `yaml:"headline"`
	SubHeadline      string `yaml:"sub_headline"`
	BriefDescription string `yaml:"brief_description"`
	CTAText          string `yaml:"cta_text"`
	CTALink          string `yaml:"cta_link"`
	SecondaryCTAText string `yaml:"secondary_cta_text"`
	SecondaryCTALink string `yaml:"secondary_cta_link"`
	TertiaryCTAText  string `yaml:"tertiary_cta_text"`
	TertiaryCTALink  string `yaml:"tertiary_cta_link"`
}

type WhatIsReading struct {
	Title   string   `yaml:"title"`
	Content []string `yaml:"content"`
}

type KeyFeatures struct {
	Title    string    `yaml:"title"`
	Features []Feature `yaml:"features"`
}

type Feature struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

type WhyItMatters struct {
	Title  string   `yaml:"title"`
	Points []string `yaml:"points"`
}

type LandingFooter struct {
	Author       string `yaml:"author"`
	GitHubLink   string `yaml:"github_link"`
	LinkedInLink string `yaml:"linkedin_link"`
}

// Registry represents the machine-readable evolution data
type Registry struct {
	Project            string              `json:"project"`
	Version            string              `json:"version"`
	LastUpdated        string              `json:"last_updated"`
	MachineRegistryURL string              `json:"machine_registry_url"`
	Milestones         []RegistryMilestone `json:"milestones"`
}

type RegistryMilestone struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
