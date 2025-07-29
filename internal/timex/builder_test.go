package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test builder pattern functionality
func TestBuilderPattern(t *testing.T) {
	builder := Option()

	t.Run("Basic builder chaining", func(t *testing.T) {
		// Test Day().WithTimezone().WithDailyResetOffset() chaining
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		opt := builder.Day().Timezone(kst).DailyResetOffset(9 * time.Hour)

		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodDay, config.Period, "Expected period to be PeriodDay")
		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone Asia/Seoul")
		assert.Equal(t, 9*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 9 hours")
	})

	t.Run("Week with custom reset time", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		opt := builder.Week().Timezone(kst).DailyResetOffset(11 * time.Hour)

		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodWeek, config.Period, "Expected period to be PeriodWeek")
		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone Asia/Seoul")
		assert.Equal(t, 11*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 11 hours")
	})

	t.Run("Month with timezone", func(t *testing.T) {
		jst, err := time.LoadLocation("Asia/Tokyo")
		require.NoError(t, err)

		opt := builder.Month().Timezone(jst)

		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodMonth, config.Period, "Expected period to be PeriodMonth")
		assert.Equal(t, "Asia/Tokyo", config.Timezone.String(), "Expected timezone Asia/Tokyo")
	})
}

// Test builder pattern with actual elapsed checking
func TestBuilderPatternWithElapsed(t *testing.T) {
	builder := Option()

	t.Run("Day elapsed with custom reset time", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		// 8 AM yesterday to 10 AM today, with 9 AM reset
		yesterday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
		today10AM := time.Date(2024, 1, 16, 10, 0, 0, 0, kst)

		assert.True(t, Elapsed(yesterday8AM, today10AM, builder.Day().Timezone(kst).DailyResetOffset(9*time.Hour)), "Expected elapsed with custom KST 9 AM reset")
	})

	t.Run("Week elapsed with custom reset time", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		// Last Monday to this Tuesday, with KST timezone and 11 AM reset
		lastMonday := time.Date(2024, 1, 8, 10, 0, 0, 0, kst)
		thisTuesday := time.Date(2024, 1, 16, 12, 0, 0, 0, kst)

		assert.True(t, Elapsed(lastMonday, thisTuesday, builder.Week().Timezone(kst).DailyResetOffset(11*time.Hour)), "Expected week elapsed with KST timezone and 11 AM reset")
	})

	t.Run("Preset options should work the same", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		yesterday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
		today10AM := time.Date(2024, 1, 16, 10, 0, 0, 0, kst)

		opts := Option()

		// Test preset option
		presetResult := Elapsed(yesterday8AM, today10AM, opts.KST9AM())

		// Test builder pattern equivalent
		builderResult := Elapsed(yesterday8AM, today10AM, builder.Day().Timezone(kst).DailyResetOffset(9*time.Hour))

		assert.Equal(t, presetResult, builderResult, "Preset and builder pattern should give same result")
		assert.True(t, presetResult, "Both preset and builder should return true")
	})
}

// Test that method chaining returns correct type
func TestBuilderReturnTypes(t *testing.T) {
	builder := Option()

	t.Run("Method chaining returns ElapsedOption", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		// This should compile and be chainable
		var opt ElapsedOption
		opt = builder.Day()
		opt = opt.Timezone(kst)
		opt = opt.DailyResetOffset(9 * time.Hour)

		// Verify the final configuration
		config := defaultElapsedConfig()
		opt.apply(&config)

		assert.Equal(t, PeriodDay, config.Period, "Expected period to be PeriodDay")
		assert.Equal(t, "Asia/Seoul", config.Timezone.String(), "Expected timezone to be Asia/Seoul")
		assert.Equal(t, 9*time.Hour, config.DailyResetOffset, "Expected daily reset offset to be 9 hours")
	})
}
