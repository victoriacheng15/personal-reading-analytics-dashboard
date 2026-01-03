package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"sort"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
)

var shortMonthNames = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// PrepareReadUnreadByYear creates JSON data for read/unread yearly breakdown chart
func PrepareReadUnreadByYear(metrics schema.Metrics) template.JS {
	// Get sorted years in descending order (latest first)
	years := make([]string, 0)
	for year := range metrics.ByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	readByYearArray := make([]int, 0)
	unreadByYearArray := make([]int, 0)

	for _, year := range years {
		yearRead := 0
		yearUnread := 0

		// Sum up read/unread from all months in this year
		if yearMonthData, exists := metrics.ByYearAndMonth[year]; exists {
			for month, count := range yearMonthData {
				yearRead += count
				// Get unread for this month (if available, otherwise calculate from total)
				if monthUnread, unreadExists := metrics.UnreadByMonth[month]; unreadExists {
					yearUnread += monthUnread
				}
			}
		}

		readByYearArray = append(readByYearArray, yearRead)
		unreadByYearArray = append(unreadByYearArray, yearUnread)
	}

	data := map[string]interface{}{
		"labels":     years,
		"readData":   readByYearArray,
		"unreadData": unreadByYearArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareReadUnreadByMonth creates JSON data for read/unread monthly breakdown chart
func PrepareReadUnreadByMonth(metrics schema.Metrics) template.JS {
	readByMonthArray := make([]int, 12)
	unreadByMonthArray := make([]int, 12)

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		unread := 0
		if u, exists := metrics.UnreadByMonth[monthStr]; exists {
			unread = u
		}
		// Calculate read for this month
		read := 0
		if monthData, exists := metrics.ByMonth[monthStr]; exists {
			read = monthData - unread
		}
		readByMonthArray[month-1] = read
		unreadByMonthArray[month-1] = unread
	}

	data := map[string]interface{}{
		"labels":     shortMonthNames,
		"readData":   readByMonthArray,
		"unreadData": unreadByMonthArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareReadUnreadBySource creates JSON data for read/unread by source chart
func PrepareReadUnreadBySource(sources []schema.SourceInfo) template.JS {
	readUnreadBySourceLabels := make([]string, 0)
	readBySourceData := make([]int, 0)
	unreadBySourceData := make([]int, 0)
	for _, source := range sources {
		readUnreadBySourceLabels = append(readUnreadBySourceLabels, source.Name)
		readBySourceData = append(readBySourceData, source.Read)
		unreadBySourceData = append(unreadBySourceData, source.Unread)
	}

	data := map[string]interface{}{
		"labels":     readUnreadBySourceLabels,
		"readData":   readBySourceData,
		"unreadData": unreadBySourceData,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareUnreadArticleAgeDistribution creates JSON data for unread articles by age chart
func PrepareUnreadArticleAgeDistribution(metrics schema.Metrics) template.JS {
	// Define age bucket labels in display order
	bucketLabels := []struct {
		key   string
		label string
	}{
		{"less_than_1_month", "Less than 1 month"},
		{"1_to_3_months", "1-3 months"},
		{"3_to_6_months", "3-6 months"},
		{"6_to_12_months", "6-12 months"},
		{"older_than_1year", "Older than 1 year"},
	}

	labels := make([]string, 0)
	data := make([]int, 0)

	for _, bucket := range bucketLabels {
		labels = append(labels, bucket.label)
		count := metrics.UnreadArticleAgeDistribution[bucket.key]
		data = append(data, count)
	}

	chartData := map[string]interface{}{
		"labels": labels,
		"data":   data,
	}
	jsonData, _ := json.Marshal(chartData)
	return template.JS(jsonData)
}

// PrepareUnreadByYear creates JSON data for unread articles by year chart
func PrepareUnreadByYear(metrics schema.Metrics) template.JS {
	// Get sorted years in descending order (latest first)
	var years []string
	for year := range metrics.UnreadByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	unreadData := make([]int, 0)
	for _, year := range years {
		unreadData = append(unreadData, metrics.UnreadByYear[year])
	}

	data := map[string]interface{}{
		"labels": years,
		"data":   unreadData,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}
