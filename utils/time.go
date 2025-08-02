package utils

import (
	"time"
)

// TimeNow returns the current time (useful for testing)
var TimeNow = time.Now

// FormatTime formats time in ISO 8601 format
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses ISO 8601 formatted time string
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// StartOfDay returns the start of day for given time
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of day for given time
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of week (Monday) for given time
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -weekday+1))
}

// EndOfWeek returns the end of week (Sunday) for given time
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth returns the start of month for given time
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of month for given time
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// DurationSince returns the duration since the given time
func DurationSince(t time.Time) time.Duration {
	return TimeNow().Sub(t)
}

// IsWeekend checks if the given time is a weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsBusinessDay checks if the given time is a business day (Monday-Friday)
func IsBusinessDay(t time.Time) bool {
	return !IsWeekend(t)
}

// AddBusinessDays adds business days to the given time
func AddBusinessDays(t time.Time, days int) time.Time {
	if days == 0 {
		return t
	}
	
	sign := 1
	if days < 0 {
		sign = -1
		days = -days
	}
	
	result := t
	for days > 0 {
		result = result.AddDate(0, 0, sign)
		if IsBusinessDay(result) {
			days--
		}
	}
	
	return result
}

// BusinessDaysBetween calculates the number of business days between two dates
func BusinessDaysBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	
	count := 0
	current := StartOfDay(start)
	endDay := StartOfDay(end)
	
	for current.Before(endDay) {
		if IsBusinessDay(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}
	
	return count
}

// UnixNow returns current Unix timestamp
func UnixNow() int64 {
	return TimeNow().Unix()
}

// UnixMilliNow returns current Unix timestamp in milliseconds
func UnixMilliNow() int64 {
	return TimeNow().UnixMilli()
}

// UnixNanoNow returns current Unix timestamp in nanoseconds
func UnixNanoNow() int64 {
	return TimeNow().UnixNano()
}

// FromUnix converts Unix timestamp to time.Time
func FromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FromUnixMilli converts Unix timestamp in milliseconds to time.Time
func FromUnixMilli(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}

// FromUnixNano converts Unix timestamp in nanoseconds to time.Time
func FromUnixNano(timestamp int64) time.Time {
	return time.Unix(0, timestamp)
}