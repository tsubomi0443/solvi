package datetime

import "time"

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// ParseAnswerDueDate parses YYYY-MM-DD and returns that day at 23:59:59 in Asia/Tokyo.
func ParseAnswerDueDate(dateStr string) (time.Time, error) {
	d, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, jst), nil
}
