package main

import (
	"fmt"
	"time"

	"github.com/homveloper/dukdakit"
)

func main() {
	fmt.Println("🕐 DukDakit Timex Usage Examples")
	fmt.Printf("📦 Version: %s\n", dukdakit.Version)
	fmt.Println()

	// Example 1: Daily Quest Reset Check
	dailyQuestExample()

	// Example 2: Weekly Event Check
	weeklyEventExample()

	// Example 3: Monthly Subscription Check
	monthlySubscriptionExample()

	// Example 4: Skill Cooldown Check
	skillCooldownExample()

	// Example 5: Weekly Event on Specific Day
	weekdayEventExample()

	// Example 6: Timezone-aware Checks
	timezoneExample()
}

func dailyQuestExample() {
	fmt.Println("=== Daily Quest Reset Example ===")

	// Simulate last daily quest completion time (yesterday)
	lastQuestTime := time.Now().AddDate(0, 0, -1) // 1 day ago
	now := time.Now()

	fmt.Printf("Last quest completed: %s\n", lastQuestTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05"))

	if dukdakit.Timex.DayElapsed(lastQuestTime, now) {
		fmt.Println("✅ Day has elapsed - Daily quests can be reset!")
	} else {
		fmt.Println("❌ Same day - Daily quests already completed today")
	}

	fmt.Println()
}

func weeklyEventExample() {
	fmt.Println("=== Weekly Event Check Example ===")

	// Simulate last weekly event time (last week)
	lastWeeklyEvent := time.Now().AddDate(0, 0, -8) // 8 days ago
	now := time.Now()

	fmt.Printf("Last weekly event: %s\n", lastWeeklyEvent.Format("2006-01-02 15:04:05"))
	fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05"))

	if dukdakit.Timex.WeekElapsed(lastWeeklyEvent, now) {
		fmt.Println("✅ Week has elapsed - Weekly event can be triggered!")
	} else {
		fmt.Println("❌ Same week - Weekly event already happened this week")
	}

	fmt.Println()
}

func monthlySubscriptionExample() {
	fmt.Println("=== Monthly Subscription Check Example ===")

	// Simulate subscription start time (last month)
	subscriptionStart := time.Now().AddDate(0, -1, -5) // 1 month and 5 days ago
	now := time.Now()

	fmt.Printf("Subscription started: %s\n", subscriptionStart.Format("2006-01-02 15:04:05"))
	fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05"))

	if dukdakit.Timex.MonthElapsed(subscriptionStart, now) {
		fmt.Println("✅ Month has elapsed - Process monthly subscription!")
	} else {
		fmt.Println("❌ Same month - Subscription still active")
	}

	fmt.Println()
}

func skillCooldownExample() {
	fmt.Println("=== Skill Cooldown Check Example ===")

	// Simulate skill cast time (30 seconds ago)
	skillCastTime := time.Now().Add(-35 * time.Second) // 35 seconds ago
	cooldownDuration := 30 * time.Second              // 30 second cooldown
	now := time.Now()

	fmt.Printf("Skill cast at: %s\n", skillCastTime.Format("15:04:05"))
	fmt.Printf("Current time: %s\n", now.Format("15:04:05"))
	fmt.Printf("Cooldown duration: %v\n", cooldownDuration)

	if dukdakit.Timex.DurationElapsed(skillCastTime, now, cooldownDuration) {
		fmt.Println("✅ Cooldown completed - Skill can be used again!")
	} else {
		remaining := cooldownDuration - now.Sub(skillCastTime)
		fmt.Printf("❌ Still on cooldown - %v remaining\n", remaining.Round(time.Second))
	}

	fmt.Println()
}

func weekdayEventExample() {
	fmt.Println("=== Weekday Event Check Example ===")

	// Simulate event start time (a few days ago)
	eventStartTime := time.Now().AddDate(0, 0, -3) // 3 days ago
	targetWeekday := time.Monday                   // Looking for Monday

	fmt.Printf("Event started: %s (%s)\n", 
		eventStartTime.Format("2006-01-02 15:04:05"), 
		eventStartTime.Weekday())
	fmt.Printf("Current time: %s (%s)\n", 
		time.Now().Format("2006-01-02 15:04:05"), 
		time.Now().Weekday())
	fmt.Printf("Target weekday: %s\n", targetWeekday)

	if dukdakit.Timex.WeekdayElapsed(eventStartTime, targetWeekday) {
		fmt.Printf("✅ %s has passed since event started!\n", targetWeekday)
	} else {
		fmt.Printf("❌ %s has not occurred yet since event started\n", targetWeekday)
	}

	fmt.Println()
}

func timezoneExample() {
	fmt.Println("=== Timezone-aware Example ===")

	// Simulate different timezone scenarios
	kst := dukdakit.Timex.KST() // Korean Standard Time
	utc := dukdakit.Timex.UTC() // UTC

	// Time 25 hours ago
	baseTime := time.Now().Add(-25 * time.Hour)
	now := time.Now()

	fmt.Printf("Base time: %s (UTC)\n", baseTime.UTC().Format("2006-01-02 15:04:05"))
	fmt.Printf("Current time: %s (UTC)\n", now.UTC().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Check day elapsed in different timezones
	fmt.Printf("Day elapsed in UTC: %v\n", 
		dukdakit.Timex.DayElapsedInTZ(baseTime, now, utc))
	fmt.Printf("Day elapsed in KST: %v\n", 
		dukdakit.Timex.DayElapsedInTZ(baseTime, now, kst))

	fmt.Printf("Base time in KST: %s\n", baseTime.In(kst).Format("2006-01-02 15:04:05"))
	fmt.Printf("Current time in KST: %s\n", now.In(kst).Format("2006-01-02 15:04:05"))

	fmt.Println()
}