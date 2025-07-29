package dukdakit

import (
	"time"

	"github.com/homveloper/dukdakit/internal/timex"
)

// TimexCategory provides time-related utilities for game development
type TimexCategory struct{}

// Timex is the global instance for time features
var Timex = &TimexCategory{}

// DayElapsed checks if a day has passed between baseTime and lastUpdate
// Useful for daily quests, daily rewards, etc.
//
// Example usage:
//   lastLoginTime := player.LastLoginTime
//   now := time.Now()
//   if dukdakit.Timex.DayElapsed(lastLoginTime, now) {
//       // Grant daily rewards
//   }
func (t *TimexCategory) DayElapsed(baseTime, lastUpdate time.Time) bool {
	return timex.DayElapsed(baseTime, lastUpdate)
}

// DayElapsedInTZ checks if a day has passed in a specific timezone
// Useful for global games with players in different timezones
func (t *TimexCategory) DayElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	return timex.DayElapsedInTZ(baseTime, lastUpdate, timezone)
}

// WeekElapsed checks if a week has passed between baseTime and lastUpdate
// Week starts on Monday (ISO 8601 standard)
// Useful for weekly events, weekly rewards, etc.
//
// Example usage:
//   lastWeeklyResetTime := event.LastWeeklyResetTime
//   now := time.Now()
//   if dukdakit.Timex.WeekElapsed(lastWeeklyResetTime, now) {
//       // Reset weekly progress, grant weekly rewards
//   }
func (t *TimexCategory) WeekElapsed(baseTime, lastUpdate time.Time) bool {
	return timex.WeekElapsed(baseTime, lastUpdate)
}

// WeekElapsedInTZ checks if a week has passed in a specific timezone
func (t *TimexCategory) WeekElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	return timex.WeekElapsedInTZ(baseTime, lastUpdate, timezone)
}

// MonthElapsed checks if a month has passed between baseTime and lastUpdate
// Useful for monthly subscriptions, monthly events, etc.
//
// Example usage:
//   subscriptionStartTime := player.SubscriptionStartTime
//   now := time.Now()
//   if dukdakit.Timex.MonthElapsed(subscriptionStartTime, now) {
//       // Process monthly subscription
//   }
func (t *TimexCategory) MonthElapsed(baseTime, lastUpdate time.Time) bool {
	return timex.MonthElapsed(baseTime, lastUpdate)
}

// MonthElapsedInTZ checks if a month has passed in a specific timezone
func (t *TimexCategory) MonthElapsedInTZ(baseTime, lastUpdate time.Time, timezone *time.Location) bool {
	return timex.MonthElapsedInTZ(baseTime, lastUpdate, timezone)
}

// DurationElapsed checks if the specified duration has passed between baseTime and lastUpdate
// Useful for custom cooldowns, temporary buffs, etc.
//
// Example usage:
//   skillCastTime := player.LastSkillCastTime
//   cooldown := 30 * time.Second
//   now := time.Now()
//   if dukdakit.Timex.DurationElapsed(skillCastTime, now, cooldown) {
//       // Skill is off cooldown
//   }
func (t *TimexCategory) DurationElapsed(baseTime, lastUpdate time.Time, duration time.Duration) bool {
	return timex.DurationElapsed(baseTime, lastUpdate, duration)
}

// WeekdayElapsed checks if a specific weekday has passed since baseTime
// Uses current time as the comparison point
// Useful for weekly events that happen on specific days
//
// Example usage:
//   eventStartTime := event.StartTime
//   if dukdakit.Timex.WeekdayElapsed(eventStartTime, time.Monday) {
//       // Monday has passed since the event started
//   }
func (t *TimexCategory) WeekdayElapsed(baseTime time.Time, targetWeekday time.Weekday) bool {
	return timex.WeekdayElapsed(baseTime, targetWeekday)
}

// WeekdayElapsedInTZ checks if a specific weekday has passed since baseTime in a specific timezone
func (t *TimexCategory) WeekdayElapsedInTZ(baseTime time.Time, targetWeekday time.Weekday, timezone *time.Location) bool {
	return timex.WeekdayElapsedInTZ(baseTime, targetWeekday, timezone)
}

// Timezone helpers for common game server locations

// KST returns Korean Standard Time timezone
func (t *TimexCategory) KST() *time.Location {
	loc, _ := time.LoadLocation("Asia/Seoul")
	return loc
}

// JST returns Japan Standard Time timezone
func (t *TimexCategory) JST() *time.Location {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return loc
}

// PST returns Pacific Standard Time timezone
func (t *TimexCategory) PST() *time.Location {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	return loc
}

// EST returns Eastern Standard Time timezone
func (t *TimexCategory) EST() *time.Location {
	loc, _ := time.LoadLocation("America/New_York")
	return loc
}

// UTC returns UTC timezone
func (t *TimexCategory) UTC() *time.Location {
	return time.UTC
}