package timex

import "time"

// Elapsed checks if time has elapsed based on the given options
func Elapsed(baseTime, lastUpdate time.Time, options ...ElapsedOption) bool {
	config := defaultElapsedConfig()
	for _, opt := range options {
		opt.apply(&config)
	}

	return checkElapsed(baseTime, lastUpdate, config)
}

// ElapsedSince checks if time has elapsed since baseTime using current time
func ElapsedSince(baseTime time.Time, options ...ElapsedOption) bool {
	return Elapsed(baseTime, time.Now(), options...)
}

// Core elapsed checking logic
func checkElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	switch config.Period {
	case PeriodDay:
		return checkDayElapsed(baseTime, lastUpdate, config)
	case PeriodWeek:
		return checkWeekElapsed(baseTime, lastUpdate, config)
	case PeriodMonth:
		return checkMonthElapsed(baseTime, lastUpdate, config)
	case PeriodCustom:
		return checkDurationElapsed(baseTime, lastUpdate, config)
	case PeriodWeekday:
		return checkWeekdayElapsed(baseTime, lastUpdate, config)
	default:
		return false
	}
}

func checkDayElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	base := baseTime.In(config.Timezone)
	last := lastUpdate.In(config.Timezone)

	// Handle custom daily reset time
	if config.DailyResetOffset > 0 {
		// Get the daily reset time for both dates
		baseResetTime := getDailyResetTimeWithOffset(base, config.DailyResetOffset, config.Timezone)
		lastResetTime := getDailyResetTimeWithOffset(last, config.DailyResetOffset, config.Timezone)

		// If time is before reset time, use previous day's reset
		if base.Before(baseResetTime) {
			baseResetTime = baseResetTime.AddDate(0, 0, -1)
		}
		if last.Before(lastResetTime) {
			lastResetTime = lastResetTime.AddDate(0, 0, -1)
		}

		return !baseResetTime.Equal(lastResetTime)
	}

	// Standard midnight-based day check
	baseDay := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, config.Timezone)
	lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, config.Timezone)

	return !baseDay.Equal(lastDay)
}

func checkWeekElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	base := baseTime.In(config.Timezone)
	last := lastUpdate.In(config.Timezone)

	baseWeekStart := getWeekStart(base, config.Timezone)
	lastWeekStart := getWeekStart(last, config.Timezone)

	return !baseWeekStart.Equal(lastWeekStart)
}

func checkMonthElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	base := baseTime.In(config.Timezone)
	last := lastUpdate.In(config.Timezone)

	return base.Year() != last.Year() || base.Month() != last.Month()
}

func checkDurationElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	return lastUpdate.Sub(baseTime) >= config.Duration
}

func checkWeekdayElapsed(baseTime, lastUpdate time.Time, config ElapsedConfig) bool {
	base := baseTime.In(config.Timezone)
	last := lastUpdate.In(config.Timezone)

	if last.Before(base) {
		return false
	}

	nextTargetWeekday := getNextWeekday(base, config.TargetWeekday, config.Timezone)
	return last.After(nextTargetWeekday) || last.Equal(nextTargetWeekday)
}

// Helper functions
func getDailyResetTimeWithOffset(t time.Time, offset time.Duration, timezone *time.Location) time.Time {
	// Start of day + offset
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timezone)
	return startOfDay.Add(offset)
}

func getWeekStart(t time.Time, timezone *time.Location) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysToSubtract := weekday - 1
	monday := t.AddDate(0, 0, -daysToSubtract)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, timezone)
}

func getNextWeekday(from time.Time, targetWeekday time.Weekday, timezone *time.Location) time.Time {
	currentWeekday := from.Weekday()

	var daysToAdd int
	if targetWeekday >= currentWeekday {
		daysToAdd = int(targetWeekday - currentWeekday)
	} else {
		daysToAdd = int(7 - currentWeekday + targetWeekday)
	}

	if daysToAdd == 0 {
		daysToAdd = 7
	}

	nextWeekday := from.AddDate(0, 0, daysToAdd)
	return time.Date(nextWeekday.Year(), nextWeekday.Month(), nextWeekday.Day(), 0, 0, 0, 0, timezone)
}
