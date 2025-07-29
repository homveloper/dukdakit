package timex

import (
	"time"
)

// DayElapsed checks if a day has passed between baseTime and lastUpdate
// It considers day boundaries (00:00:00) in UTC timezone
func DayElapsed(baseTime, lastUpdate time.Time) bool {
	base := baseTime.UTC()
	last := lastUpdate.UTC()
	
	// Get the start of the day for both times
	baseDay := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
	
	return !baseDay.Equal(lastDay)
}

// DayElapsedInTZ checks if a day has passed between baseTime and lastUpdate in specified timezone
func DayElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	base := baseTime.In(timezone)
	last := lastUpdate.In(timezone)
	
	// Get the start of the day for both times
	baseDay := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, timezone)
	lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, timezone)
	
	return !baseDay.Equal(lastDay)
}

// WeekElapsed checks if a week has passed between baseTime and lastUpdate
// Week starts on Monday (ISO 8601 standard) in UTC timezone
func WeekElapsed(baseTime, lastUpdate time.Time) bool {
	base := baseTime.UTC()
	last := lastUpdate.UTC()
	
	// Get the start of the week (Monday) for both times
	baseWeekStart := getWeekStart(base, time.UTC)
	lastWeekStart := getWeekStart(last, time.UTC)
	
	return !baseWeekStart.Equal(lastWeekStart)
}

// WeekElapsedInTZ checks if a week has passed between baseTime and lastUpdate in specified timezone
func WeekElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	base := baseTime.In(timezone)
	last := lastUpdate.In(timezone)
	
	// Get the start of the week (Monday) for both times
	baseWeekStart := getWeekStart(base, timezone)
	lastWeekStart := getWeekStart(last, timezone)
	
	return !baseWeekStart.Equal(lastWeekStart)
}

// MonthElapsed checks if a month has passed between baseTime and lastUpdate
func MonthElapsed(baseTime, lastUpdate time.Time) bool {
	base := baseTime.UTC()
	last := lastUpdate.UTC()
	
	// Check if year or month is different
	return base.Year() != last.Year() || base.Month() != last.Month()
}

// MonthElapsedInTZ checks if a month has passed between baseTime and lastUpdate in specified timezone
func MonthElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	base := baseTime.In(timezone)
	last := lastUpdate.In(timezone)
	
	// Check if year or month is different
	return base.Year() != last.Year() || base.Month() != last.Month()
}

// DurationElapsed checks if the specified duration has passed between baseTime and lastUpdate
func DurationElapsed(baseTime, lastUpdate time.Time, duration time.Duration) bool {
	return lastUpdate.Sub(baseTime) >= duration
}

// WeekdayElapsed checks if a specific weekday has passed since baseTime (using current time)
// For example, if it's Tuesday and you check for Monday, it returns true
func WeekdayElapsed(baseTime time.Time, targetWeekday time.Weekday) bool {
	return WeekdayElapsedInTZ(baseTime, targetWeekday, time.UTC)
}

// WeekdayElapsedInTZ checks if a specific weekday has passed since baseTime in specified timezone
func WeekdayElapsedInTZ(baseTime time.Time, targetWeekday time.Weekday, timezone *time.Location) bool {
	now := time.Now().In(timezone)
	base := baseTime.In(timezone)
	
	// If base time is in the future, no weekday has elapsed
	if now.Before(base) {
		return false
	}
	
	// Find the next occurrence of target weekday after base time
	nextTargetWeekday := getNextWeekday(base, targetWeekday, timezone)
	
	// Check if current time has passed the next target weekday
	return now.After(nextTargetWeekday) || now.Equal(nextTargetWeekday)
}

// Helper function to get the start of the week (Monday)
func getWeekStart(t time.Time, timezone *time.Location) time.Time {
	// Get the current weekday (0 = Sunday, 1 = Monday, ...)
	weekday := int(t.Weekday())
	
	// Convert Sunday (0) to 7 for easier calculation
	if weekday == 0 {
		weekday = 7
	}
	
	// Calculate days to subtract to get to Monday (1)
	daysToSubtract := weekday - 1
	
	// Get Monday of this week
	monday := t.AddDate(0, 0, -daysToSubtract)
	
	// Set to start of day
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, timezone)
}

// Helper function to get the next occurrence of a specific weekday
func getNextWeekday(from time.Time, targetWeekday time.Weekday, timezone *time.Location) time.Time {
	currentWeekday := from.Weekday()
	
	var daysToAdd int
	if targetWeekday >= currentWeekday {
		daysToAdd = int(targetWeekday - currentWeekday)
	} else {
		daysToAdd = int(7 - currentWeekday + targetWeekday)
	}
	
	// If it's the same weekday, move to next week
	if daysToAdd == 0 {
		daysToAdd = 7
	}
	
	nextWeekday := from.AddDate(0, 0, daysToAdd)
	
	// Set to start of day
	return time.Date(nextWeekday.Year(), nextWeekday.Month(), nextWeekday.Day(), 0, 0, 0, 0, timezone)
}