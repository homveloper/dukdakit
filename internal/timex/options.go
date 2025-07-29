package timex

import "time"

// ElapsedOption configures how elapsed time is calculated using builder pattern
type ElapsedOption struct {
	config ElapsedConfig
}

// Build applies the option to the config
func (opt ElapsedOption) apply(config *ElapsedConfig) {
	*config = opt.config
}

// ElapsedConfig holds configuration for elapsed time checking
type ElapsedConfig struct {
	// Period type
	Period PeriodType

	// Timezone for calculations (default: UTC)
	Timezone *time.Location

	// Custom duration for PeriodCustom
	Duration time.Duration

	// Daily reset time offset from midnight
	// For example: 9*time.Hour means reset at 09:00
	DailyResetOffset time.Duration

	// Target weekday for PeriodWeekday
	TargetWeekday time.Weekday
}

// PeriodType defines different types of time periods
type PeriodType int

const (
	PeriodDay PeriodType = iota
	PeriodWeek
	PeriodMonth
	PeriodCustom
	PeriodWeekday
)

// Default configuration
func defaultElapsedConfig() ElapsedConfig {
	return ElapsedConfig{
		Period:           PeriodDay,
		Timezone:         time.UTC,
		Duration:         24 * time.Hour,
		DailyResetOffset: 0, // Midnight
		TargetWeekday:    time.Monday,
	}
}

// Chainable builder methods for ElapsedOption

// Timezone sets the timezone for calculations
func (opt ElapsedOption) Timezone(tz *time.Location) ElapsedOption {
	opt.config.Timezone = tz
	return opt
}

// DailyResetOffset sets the daily reset time offset from midnight
// For example: DailyResetOffset(9*time.Hour) for 09:00 reset
func (opt ElapsedOption) DailyResetOffset(offset time.Duration) ElapsedOption {
	opt.config.DailyResetOffset = offset
	return opt
}
