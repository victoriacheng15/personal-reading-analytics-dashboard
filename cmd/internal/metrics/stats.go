package metrics

import (
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/cmd/internal"
)

// CalculateTopReadRateSource finds the source with the highest read rate
func CalculateTopReadRateSource(metrics schema.Metrics) string {
	var topSource string
	var topRate float64
	for name, counts := range metrics.BySourceReadStatus {
		if name == "substack_author_count" {
			continue
		}
		total := counts[0] + counts[1]
		if total > 0 {
			rate := float64(counts[0]) / float64(total) * 100
			if rate > topRate {
				topRate = rate
				topSource = name
			}
		}
	}
	return topSource
}

// CalculateMostUnreadSource finds the source with the most unread articles
func CalculateMostUnreadSource(metrics schema.Metrics) string {
	var mostUnreadSource string
	var maxUnread int
	for name, unread := range metrics.UnreadBySource {
		if unread > maxUnread {
			maxUnread = unread
			mostUnreadSource = name
		}
	}
	return mostUnreadSource
}

// CalculateThisMonthArticles calculates articles read this month.
// If currentMonth is empty, it uses the current system month.
func CalculateThisMonthArticles(metrics schema.Metrics, currentMonth string) int {
	if currentMonth == "" {
		currentMonth = time.Now().Format("01")
	}

	// Sum all read articles from by_month_and_source_read_status for current month
	if monthData, exists := metrics.ByMonthAndSource[currentMonth]; exists {
		total := 0
		for _, counts := range monthData {
			total += counts[0] // read count
		}
		return total
	}
	return 0
}
