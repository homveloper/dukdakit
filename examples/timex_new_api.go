package main

import (
	"fmt"
	"time"

	"github.com/homveloper/dukdakit"
)

func main() {
	fmt.Println("🎯 DukDakit Timex New API Examples")
	fmt.Printf("📦 Version: %s\n", dukdakit.Version)
	fmt.Println()

	// When users type "dukdakit.Timex." they will see:
	// - Elapsed()
	// - ElapsedSince()
	// - Option()
	// - KST(), JST(), UTC(), etc. (timezone helpers)

	basicUsageExample()
	optionBuilderExample()
	comparisonExample()
}

func basicUsageExample() {
	fmt.Println("=== Basic Usage with New API ===")

	lastLogin := time.Now().AddDate(0, 0, -1)

	// Simple day check - users discover Elapsed first
	if dukdakit.Timex.ElapsedSince(lastLogin) {
		fmt.Println("✅ Daily login bonus available (default midnight UTC)")
	}

	// Then they discover Option() for customization
	if dukdakit.Timex.ElapsedSince(lastLogin, dukdakit.Timex.Option().KST9AM()) {
		fmt.Println("✅ Daily quest reset available (KST 9:00 AM)")
	}

	fmt.Println()
}

func optionBuilderExample() {
	fmt.Println("=== Option Builder Pattern ===")

	// Users type dukdakit.Timex.Option() and discover:
	// - Day()
	// - Week()
	// - Month()
	// - Duration()
	// - Weekday()
	// - KST9AM(), KST11AM(), UTCMidnight() (presets)

	lastWeeklyEvent := time.Now().AddDate(0, 0, -8)

	// Simple week check
	if dukdakit.Timex.ElapsedSince(lastWeeklyEvent, dukdakit.Timex.Option().Week()) {
		fmt.Println("✅ Weekly event ready (simple)")
	}

	// Week check with custom reset time
	if dukdakit.Timex.ElapsedSince(lastWeeklyEvent,
		dukdakit.Timex.Option().Week().
			Timezone(dukdakit.Timex.KST()).
			DailyResetOffset(11*time.Hour)) {
		fmt.Println("✅ Weekly event ready (KST 11:00 AM reset)")
	}

	// Skill cooldown
	skillCastTime := time.Now().Add(-35 * time.Second)
	if dukdakit.Timex.ElapsedSince(skillCastTime, dukdakit.Timex.Option().Duration(30*time.Second)) {
		fmt.Println("✅ Skill off cooldown")
	}

	// Monthly subscription
	subscriptionStart := time.Now().AddDate(0, 0, -35)
	if dukdakit.Timex.ElapsedSince(subscriptionStart, dukdakit.Timex.Option().Month()) {
		fmt.Println("✅ Monthly billing due")
	}

	// Weekday event
	eventStart := time.Now().AddDate(0, 0, -4)
	if dukdakit.Timex.ElapsedSince(eventStart, dukdakit.Timex.Option().Weekday(time.Friday)) {
		fmt.Println("✅ Friday event available")
	}

	fmt.Println()
}

func comparisonExample() {
	fmt.Println("=== API Discoverability Comparison ===")

	fmt.Println("OLD API - Users saw many options at top level:")
	fmt.Println("  dukdakit.WithDay")
	fmt.Println("  dukdakit.WithWeek")
	fmt.Println("  dukdakit.WithMonth")
	fmt.Println("  dukdakit.WithDuration")
	fmt.Println("  dukdakit.WithKST9AM")
	fmt.Println("  dukdakit.WithTimezone")
	fmt.Println("  dukdakit.WithDailyResetOffset")
	fmt.Println("  ... (confusing with too many choices)")
	fmt.Println()

	fmt.Println("NEW API - Users see clear entry points:")
	fmt.Println("  dukdakit.Timex.Elapsed() / ElapsedSince() ← Main functions")
	fmt.Println("  dukdakit.Timex.Option() ← Clear customization entry")
	fmt.Println("  └── .Day() / .Week() / .Month() / .Duration() / .Weekday()")
	fmt.Println("  └── .KST9AM() / .KST11AM() / .UTCMidnight() (presets)")
	fmt.Println("  dukdakit.Timex.KST() / .JST() / .UTC() ← Timezone helpers")
	fmt.Println()

	fmt.Println("Benefits:")
	fmt.Println("  ✅ Elapsed functions are discovered first")
	fmt.Println("  ✅ Options are organized under Option()")
	fmt.Println("  ✅ Clear separation of concerns")
	fmt.Println("  ✅ Better IDE autocomplete experience")
	fmt.Println("  ✅ Follows Go best practices for package design")
}
