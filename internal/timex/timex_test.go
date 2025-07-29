package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElapsed_BasicDay(t *testing.T) {
	// Yesterday to today
	yesterday := time.Now().AddDate(0, 0, -1)
	now := time.Now()

	// Simple day check (UTC midnight)
	assert.True(t, Elapsed(yesterday, now), "Expected day elapsed for yesterday")

	// Same time should not be elapsed
	assert.False(t, Elapsed(now, now), "Expected same time to not be elapsed")
}

func TestElapsedSince_BasicUsage(t *testing.T) {
	// Yesterday should be elapsed
	yesterday := time.Now().AddDate(0, 0, -1)
	assert.True(t, ElapsedSince(yesterday), "Expected day elapsed since yesterday")

	// Current time should not be elapsed
	now := time.Now()
	assert.False(t, ElapsedSince(now), "Expected current time to not be elapsed")
}

func TestElapsed_CustomResetTime(t *testing.T) {
	// Create specific times in KST
	kst, err := time.LoadLocation("Asia/Seoul")
	require.NoError(t, err)

	today8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
	today10AM := time.Date(2024, 1, 15, 10, 0, 0, 0, kst)

	// Check with KST 9:00 AM reset
	// 8 AM -> 10 AM should cross 9 AM reset boundary
	assert.True(t, Elapsed(today8AM, today10AM, Option().KST9AM()), "Expected elapsed with KST 9:00 AM reset")

	// 10 AM -> 10:30 AM should not cross reset boundary
	today1030AM := time.Date(2024, 1, 15, 10, 30, 0, 0, kst)
	assert.False(t, Elapsed(today10AM, today1030AM, Option().KST9AM()), "Expected no elapsed within same reset period")
}

func TestElapsed_WeekCheck(t *testing.T) {
	// Create specific times to avoid week boundary issues
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)     // Monday
	nextMonday := time.Date(2024, 1, 22, 10, 0, 0, 0, time.UTC) // Next Monday

	// Week should be elapsed
	builder := Option()
	assert.True(t, Elapsed(monday, nextMonday, builder.Week()), "Expected week elapsed between consecutive Mondays")

	// Same week should not be elapsed
	tuesday := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC) // Tuesday same week
	assert.False(t, Elapsed(monday, tuesday, builder.Week()), "Expected no week elapsed within same week")
}

func TestElapsed_MonthCheck(t *testing.T) {
	// January to February
	january := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	february := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)

	// Monthly check should be elapsed
	builder := Option()
	assert.True(t, Elapsed(january, february, builder.Month()), "Expected month elapsed from January to February")

	// Same month should not be elapsed
	january20 := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	assert.False(t, Elapsed(january, january20, builder.Month()), "Expected no month elapsed within same month")
}

func TestElapsed_CustomDuration(t *testing.T) {
	// Use fixed times for reliable testing
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 35, 0, time.UTC) // 35 seconds later
	cooldown := 30 * time.Second

	// Duration should be elapsed
	builder := Option()
	assert.True(t, Elapsed(start, end, builder.Duration(cooldown)), "Expected duration elapsed after 35 seconds with 30 second duration")

	// 20 seconds later should not be elapsed
	shortEnd := time.Date(2024, 1, 15, 10, 0, 20, 0, time.UTC) // 20 seconds later
	assert.False(t, Elapsed(start, shortEnd, builder.Duration(cooldown)), "Expected duration not elapsed after 20 seconds with 30 second duration")
}

func TestElapsed_WeekdayCheck(t *testing.T) {
	// Monday to Saturday (Friday should have passed)
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)   // Monday
	saturday := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC) // Saturday

	// Friday should have passed since Monday by Saturday
	builder := Option()
	assert.True(t, Elapsed(monday, saturday, builder.Weekday(time.Friday)), "Expected Friday to have passed since Monday by Saturday")

	// Thursday should not have Friday passed yet
	thursday := time.Date(2024, 1, 18, 10, 0, 0, 0, time.UTC) // Thursday
	assert.False(t, Elapsed(monday, thursday, builder.Weekday(time.Friday)), "Expected Friday not to have passed yet on Thursday")
}

func TestElapsed_CombinedOptions(t *testing.T) {
	// Test combining multiple options: weekly reset with custom time
	kst, err := time.LoadLocation("Asia/Seoul")
	require.NoError(t, err)

	lastWeek := time.Date(2024, 1, 8, 10, 0, 0, 0, kst)  // Monday last week
	thisWeek := time.Date(2024, 1, 16, 10, 0, 0, 0, kst) // Tuesday this week

	opts := Option()

	// Weekly check with KST timezone and custom reset offset using builder pattern
	assert.True(t, Elapsed(lastWeek, thisWeek, opts.Week().Timezone(kst).DailyResetOffset(9*time.Hour)), "Expected week elapsed with KST timezone and custom reset")
}

func TestPresetOptions(t *testing.T) {
	// Test KST 9 AM preset
	kst, err := time.LoadLocation("Asia/Seoul")
	require.NoError(t, err)

	yesterday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, kst)
	today10AM := time.Date(2024, 1, 16, 10, 0, 0, 0, kst)

	assert.True(t, Elapsed(yesterday8AM, today10AM, Option().KST9AM()), "Expected elapsed with KST 9:00 AM preset")

	// Test KST 11 AM preset
	assert.True(t, Elapsed(yesterday8AM, today10AM, Option().KST11AM()), "Expected elapsed with KST 11:00 AM preset")

	// Test UTC midnight preset
	yesterdayUTC := time.Date(2024, 1, 15, 23, 0, 0, 0, time.UTC)
	todayUTC := time.Date(2024, 1, 16, 1, 0, 0, 0, time.UTC)

	assert.True(t, Elapsed(yesterdayUTC, todayUTC, Option().UTCMidnight()), "Expected elapsed with UTC midnight preset")
}

// Game scenario tests demonstrating usage patterns
func TestGameScenarios(t *testing.T) {
	t.Run("Daily Login Bonus", func(t *testing.T) {
		lastLogin := time.Now().AddDate(0, 0, -1)
		assert.True(t, ElapsedSince(lastLogin), "Player should receive daily login bonus")
	})

	t.Run("Daily Quest Reset KST 9AM", func(t *testing.T) {
		lastQuestTime := time.Now().Add(-25 * time.Hour)
		assert.True(t, ElapsedSince(lastQuestTime, Option().KST9AM()), "Daily quests should reset at KST 9:00 AM")
	})

	t.Run("Skill Cooldown", func(t *testing.T) {
		skillCastTime := time.Now().Add(-35 * time.Second)
		assert.True(t, ElapsedSince(skillCastTime, Option().Duration(30*time.Second)), "Skill should be off cooldown")
	})

	t.Run("Monthly Subscription", func(t *testing.T) {
		subscriptionStart := time.Now().AddDate(0, 0, -35)
		assert.True(t, ElapsedSince(subscriptionStart, Option().Month()), "Monthly subscription should be due for billing")
	})

	t.Run("Weekly Event with Custom Reset", func(t *testing.T) {
		lastWeeklyEvent := time.Now().AddDate(0, 0, -8)
		assert.True(t, ElapsedSince(lastWeeklyEvent, Option().KST9AM()), "Weekly event should trigger")
	})
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("getDailyResetTimeWithOffset", func(t *testing.T) {
		kst, err := time.LoadLocation("Asia/Seoul")
		require.NoError(t, err)

		testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, kst)

		resetTime := getDailyResetTimeWithOffset(testTime, 9*time.Hour, kst)
		expected := time.Date(2024, 1, 15, 9, 0, 0, 0, kst)

		assert.True(t, resetTime.Equal(expected), "Expected reset time %v, got %v", expected, resetTime)
	})

	t.Run("getWeekStart", func(t *testing.T) {
		// Test with Wednesday
		wednesday := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC) // Wednesday

		weekStart := getWeekStart(wednesday, time.UTC)
		expectedMonday := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // Monday

		assert.True(t, weekStart.Equal(expectedMonday), "Expected week start %v, got %v", expectedMonday, weekStart)
	})

	t.Run("getNextWeekday", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Monday

		nextFriday := getNextWeekday(monday, time.Friday, time.UTC)
		expectedFriday := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC) // Friday same week

		assert.True(t, nextFriday.Equal(expectedFriday), "Expected next Friday %v, got %v", expectedFriday, nextFriday)
	})
}

// Test configuration and options
func TestElapsedConfig(t *testing.T) {
	t.Run("defaultElapsedConfig", func(t *testing.T) {
		config := defaultElapsedConfig()

		assert.Equal(t, PeriodDay, config.Period, "Expected default period to be PeriodDay")
		assert.Equal(t, time.UTC, config.Timezone, "Expected default timezone to be UTC")
		assert.Equal(t, 24*time.Hour, config.Duration, "Expected default duration to be 24 hours")
		assert.Equal(t, time.Duration(0), config.DailyResetOffset, "Expected default daily reset offset to be 0")
	})
}
