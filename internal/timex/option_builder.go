package timex

import "time"

// OptionBuilder provides a starting point for building elapsed options
type OptionBuilder struct{}

// Option creates a new option builder
func Option() *OptionBuilder {
	return &OptionBuilder{}
}

// Day returns an ElapsedOption configured for day elapsed checking
func (ob *OptionBuilder) Day() ElapsedOption {
	config := defaultElapsedConfig()
	config.Period = PeriodDay
	return ElapsedOption{config: config}
}

// Week returns an ElapsedOption configured for week elapsed checking
func (ob *OptionBuilder) Week() ElapsedOption {
	config := defaultElapsedConfig()
	config.Period = PeriodWeek
	return ElapsedOption{config: config}
}

// Month returns an ElapsedOption configured for month elapsed checking
func (ob *OptionBuilder) Month() ElapsedOption {
	config := defaultElapsedConfig()
	config.Period = PeriodMonth
	return ElapsedOption{config: config}
}

// Duration returns an ElapsedOption configured for custom duration elapsed checking
func (ob *OptionBuilder) Duration(duration time.Duration) ElapsedOption {
	config := defaultElapsedConfig()
	config.Period = PeriodCustom
	config.Duration = duration
	return ElapsedOption{config: config}
}

// Weekday returns an ElapsedOption configured for specific weekday elapsed checking
func (ob *OptionBuilder) Weekday(weekday time.Weekday) ElapsedOption {
	config := defaultElapsedConfig()
	config.Period = PeriodWeekday
	config.TargetWeekday = weekday
	return ElapsedOption{config: config}
}

// KST9AM returns a preset option for KST 9:00 AM daily reset
func (ob *OptionBuilder) KST9AM() ElapsedOption {
	config := defaultElapsedConfig()
	kst, _ := time.LoadLocation("Asia/Seoul")
	config.Timezone = kst
	config.DailyResetOffset = 9 * time.Hour
	return ElapsedOption{config: config}
}

// KST11AM returns a preset option for KST 11:00 AM daily reset
func (ob *OptionBuilder) KST11AM() ElapsedOption {
	config := defaultElapsedConfig()
	kst, _ := time.LoadLocation("Asia/Seoul")
	config.Timezone = kst
	config.DailyResetOffset = 11 * time.Hour
	return ElapsedOption{config: config}
}

// UTCMidnight returns a preset option for UTC midnight reset
func (ob *OptionBuilder) UTCMidnight() ElapsedOption {
	config := defaultElapsedConfig()
	config.Timezone = time.UTC
	config.DailyResetOffset = 0
	return ElapsedOption{config: config}
}
