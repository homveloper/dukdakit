package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test OptionBuilder functionality
func TestOptionBuilder(t *testing.T) {
	builder := Option()

	t.Run("Day option", func(t *testing.T) {
		opt := builder.Day()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodDay, config.Period, "Expected period to be PeriodDay")
	})

	t.Run("Week option", func(t *testing.T) {
		opt := builder.Week()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodWeek, config.Period, "Expected period to be PeriodWeek")
	})

	t.Run("Month option", func(t *testing.T) {
		opt := builder.Month()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodMonth, config.Period, "Expected period to be PeriodMonth")
	})

	t.Run("Duration option", func(t *testing.T) {
		testDuration := 30 * time.Minute
		opt := builder.Duration(testDuration)
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodCustom, config.Period, "Expected period to be PeriodCustom")
		assert.Equal(t, testDuration, config.Duration, "Expected duration to match test duration")
	})

	t.Run("Weekday option", func(t *testing.T) {
		opt := builder.Weekday(time.Friday)
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodWeekday, config.Period, "Expected period to be PeriodWeekday")
		assert.Equal(t, time.Friday, config.TargetWeekday, "Expected target weekday to be Friday")
	})
}

// Test OptionBuilder preset methods
func TestOptionBuilderPresets(t *testing.T) {
	builder := Option()

	t.Run("KST9AM preset", func(t *testing.T) {
		opt := builder.KST9AM()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone Asia/Seoul")
		assert.Equal(t, 9*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 9 hours")
	})

	t.Run("KST11AM preset", func(t *testing.T) {
		opt := builder.KST11AM()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone Asia/Seoul")
		assert.Equal(t, 11*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 11 hours")
	})

	t.Run("UTCMidnight preset", func(t *testing.T) {
		opt := builder.UTCMidnight()
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, time.UTC, config.Timezone, "Expected timezone to be UTC")
		assert.Equal(t, time.Duration(0), config.DailyResetOffset, "Expected daily reset offset to be 0")
	})
}

// Test OptionBuilder with chaining
func TestOptionBuilderChaining(t *testing.T) {
	builder := Option()

	t.Run("Day with custom reset time", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		opt := builder.Day().Timezone(kst).DailyResetOffset(9 * time.Hour)

		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodDay, config.Period, "Expected period to be PeriodDay")
		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone Asia/Seoul")
		assert.Equal(t, 9*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 9 hours")
	})

	t.Run("Week with timezone", func(t *testing.T) {
		jst, err := time.LoadLocation("Asia/Tokyo")
		require.NoError(t, err)

		opt := builder.Week().Timezone(jst)

		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodWeek, config.Period, "Expected period to be PeriodWeek")
		assert.Equal(t, "Asia/Tokyo", config.Timezone.String(), "Expected timezone Asia/Tokyo")
	})
}

// Test OptionBuilder with actual elapsed checking
func TestOptionBuilderWithElapsed(t *testing.T) {
	builder := Option()

	t.Run("Day elapsed with OptionBuilder", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		yesterday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
		today10AM := time.Date(2024, 1, 16, 10, 0, 0, 0, kst)

		assert.True(t, Elapsed(yesterday8AM, today10AM, builder.Day().Timezone(kst).DailyResetOffset(9*time.Hour)), "Expected elapsed with OptionBuilder Day method")
	})

	t.Run("Duration elapsed with OptionBuilder", func(t *testing.T) {
		start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 15, 10, 0, 35, 0, time.UTC)

		assert.True(t, Elapsed(start, end, builder.Duration(30*time.Second)), "Expected duration elapsed with OptionBuilder Duration method")
	})

	t.Run("Preset option with OptionBuilder", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		yesterday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
		today10AM := time.Date(2024, 1, 16, 10, 0, 0, 0, kst)

		assert.True(t, Elapsed(yesterday8AM, today10AM, builder.KST9AM()), "Expected elapsed with OptionBuilder KST9AM preset")
	})
}
