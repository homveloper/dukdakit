package dukdakit

import (
	"time"

	"github.com/homveloper/dukdakit/internal/timex"
)

// TimexCategory provides time-related utilities for game development
type TimexCategory struct{}

// OptionBuilder provides access to option building methods
type OptionBuilder = timex.OptionBuilder

// Timex is the global instance for time features
var Timex = &TimexCategory{}

// Elapsed checks if time has elapsed between baseTime and lastUpdate based on options
// This is the main function that handles all time elapsed checking scenarios
//
// Example usage:
//
//	// Basic day elapsed (midnight reset in UTC)
//	if dukdakit.Timex.Elapsed(lastLoginTime, now) {
//	    // Daily reset occurred
//	}
//
//	// Week elapsed
//	if dukdakit.Timex.Elapsed(lastWeeklyTime, now, timex.WithWeek()) {
//	    // Weekly reset occurred
//	}
//
//	// Month elapsed
//	if dukdakit.Timex.Elapsed(lastMonthlyTime, now, timex.WithMonth()) {
//	    // Monthly reset occurred
//	}
//
//	// Custom duration (30 minutes)
//	if dukdakit.Timex.Elapsed(skillCastTime, now, timex.WithDuration(30*time.Minute)) {
//	    // 30 minutes have passed
//	}
//
//	// KST 9:00 AM daily reset (common for mobile games)
//	if dukdakit.Timex.Elapsed(lastQuestTime, now, timex.WithKST9AM()) {
//	    // Daily quest reset at 9 AM KST
//	}
//
//	// Custom daily reset time with builder pattern
//	if dukdakit.Timex.Elapsed(lastTime, now,
//	    dukdakit.WithDay().WithTimezone(dukdakit.Timex.KST()).WithDailyResetOffset(11*time.Hour)) {
//	    // Daily reset at 11:00 AM KST
//	}
//
//	// Weekday elapsed (check if Monday passed)
//	if dukdakit.Timex.Elapsed(eventStartTime, now, timex.WithWeekday(time.Monday)) {
//	    // Monday has passed since event started
//	}
func (t *TimexCategory) Elapsed(baseTime, lastUpdate time.Time, options ...timex.ElapsedOption) bool {
	return timex.Elapsed(baseTime, lastUpdate, options...)
}

// ElapsedSince checks if time has elapsed since baseTime using current time
// This is a convenience function that uses time.Now() as the comparison point
//
// Example usage:
//
//	// Check if a day has passed since last login
//	if dukdakit.Timex.ElapsedSince(player.LastLoginTime) {
//	    // Grant daily bonus
//	}
//
//	// Check if skill cooldown is over
//	if dukdakit.Timex.ElapsedSince(skill.LastUsedTime, timex.WithDuration(30*time.Second)) {
//	    // Skill is off cooldown
//	}
//
//	// Check if weekly event should trigger with builder pattern
//	if dukdakit.Timex.ElapsedSince(event.LastTriggerTime, dukdakit.WithWeek().WithTimezone(dukdakit.Timex.KST()).WithDailyResetOffset(9*time.Hour)) {
//	    // Weekly event at KST 9 AM reset
//	}
func (t *TimexCategory) ElapsedSince(baseTime time.Time, options ...timex.ElapsedOption) bool {
	return timex.ElapsedSince(baseTime, options...)
}

// Option returns an OptionBuilder for creating elapsed time options
// This is the main entry point for building options with method chaining
//
// Example usage:
//
//	// Basic day elapsed with custom reset time
//	if dukdakit.Timex.ElapsedSince(lastTime, dukdakit.Timex.Option().Day().WithTimezone(kst).WithDailyResetOffset(9*time.Hour)) {
//	    // Day has elapsed with KST 9:00 AM reset
//	}
//
//	// Week elapsed
//	if dukdakit.Timex.ElapsedSince(lastTime, dukdakit.Timex.Option().Week()) {
//	    // Week has elapsed
//	}
//
//	// Custom duration
//	if dukdakit.Timex.ElapsedSince(lastTime, dukdakit.Timex.Option().Duration(30*time.Second)) {
//	    // 30 seconds have passed
//	}
//
//	// Preset options
//	if dukdakit.Timex.ElapsedSince(lastTime, dukdakit.Timex.Option().KST9AM()) {
//	    // KST 9:00 AM daily reset
//	}
func (t *TimexCategory) Option() *OptionBuilder {
	return timex.Option()
}

// ElapsedOption type for backward compatibility and direct usage
type ElapsedOption = timex.ElapsedOption

// Timezone helpers
func (t *TimexCategory) KST() *time.Location {
	loc, _ := time.LoadLocation("Asia/Seoul")
	return loc
}

func (t *TimexCategory) JST() *time.Location {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return loc
}

func (t *TimexCategory) PST() *time.Location {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	return loc
}

func (t *TimexCategory) EST() *time.Location {
	loc, _ := time.LoadLocation("America/New_York")
	return loc
}

func (t *TimexCategory) UTC() *time.Location {
	return time.UTC
}
