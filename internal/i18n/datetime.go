package i18n

import "time"

// FormatDate converts the specified date (year, month, and day)
// according to the current locale
func FormatDate(t time.Time) string {
	// TODO
	return t.Format("2006-01-02")
}

// FormatDatetime formats the specified datetime (date, hour, and minute)
// according to the current locale.
func FormatDatetime(t time.Time) string {
	// TODO
	return t.Format("2006-01-02 15:04:05 MST")
}
