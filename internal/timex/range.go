package timex

import "time"

// TimeRange represents a time interval with start and end times
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// RangeConfig holds configuration for time range splitting
type RangeConfig struct {
	// SplitMode determines how the time range is split
	SplitMode SplitMode

	// TrimFirst removes partial duration at the beginning
	TrimFirst bool

	// TrimLast removes partial duration at the end
	TrimLast bool
}

// SplitMode defines how time ranges are split
type SplitMode int

const (
	// SplitExact splits from start time with exact durations
	SplitExact SplitMode = iota
	// SplitAligned splits aligned to natural boundaries (e.g., top of the hour)
	SplitAligned
)

// RangeOption configures time range splitting behavior
type RangeOption func(*RangeConfig)

// WithExactSplit splits the time range starting exactly from the start time
func WithExactSplit() RangeOption {
	return func(config *RangeConfig) {
		config.SplitMode = SplitExact
	}
}

// WithAlignedSplit splits the time range aligned to natural boundaries
// For example, if splitting by 1 hour and start is 14:35:20, the first range
// starts at 15:00:00 (aligned to the hour)
func WithAlignedSplit() RangeOption {
	return func(config *RangeConfig) {
		config.SplitMode = SplitAligned
	}
}

// WithTrimFirst removes the first partial interval if it's shorter than the full duration
func WithTrimFirst() RangeOption {
	return func(config *RangeConfig) {
		config.TrimFirst = true
	}
}

// WithTrimLast removes the last partial interval if it's shorter than the full duration
func WithTrimLast() RangeOption {
	return func(config *RangeConfig) {
		config.TrimLast = true
	}
}

// WithTrim removes both first and last partial intervals if they're shorter than the full duration
// This is equivalent to combining WithTrimFirst() and WithTrimLast()
func WithTrim() RangeOption {
	return func(config *RangeConfig) {
		config.TrimFirst = true
		config.TrimLast = true
	}
}

// Range splits a time interval into smaller ranges based on the given duration
//
// Parameters:
//   - start: The start time of the range
//   - end: The end time of the range
//   - duration: The duration to split by
//   - options: Configuration options for splitting behavior
//
// Returns:
//   - []TimeRange: Array of time ranges split according to the configuration
//
// Example usage:
//
//	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
//	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
//
//	// Aligned splitting (default - aligns to hour boundaries)
//	ranges := timex.Range(start, end, time.Hour)
//	// Result: [15:00:00-16:00:00, 16:00:00-17:00:00]
//
//	// Exact splitting (starts from 14:35:20)
//	ranges = timex.Range(start, end, time.Hour, WithExactSplit())
//	// Result: [14:35:20-15:35:20, 15:35:20-16:35:20, 16:35:20-17:25:40]
//
//	// With trimming options (removes partial ranges)
//	ranges = timex.Range(start, end, time.Hour, WithTrim())
//	// Result: [15:00:00-16:00:00, 16:00:00-17:00:00] (partial ranges removed)
func Range(start, end time.Time, duration time.Duration, options ...RangeOption) []TimeRange {
	if start.After(end) || duration <= 0 {
		return nil
	}

	config := &RangeConfig{
		SplitMode: SplitAligned,
		TrimFirst: false,
		TrimLast:  false,
	}

	for _, option := range options {
		option(config)
	}

	var ranges []TimeRange

	switch config.SplitMode {
	case SplitExact:
		ranges = splitExact(start, end, duration)
	case SplitAligned:
		ranges = splitAligned(start, end, duration)
	}

	if config.TrimFirst && len(ranges) > 0 {
		firstRange := ranges[0]
		if firstRange.End.Sub(firstRange.Start) < duration {
			ranges = ranges[1:]
		}
	}

	if config.TrimLast && len(ranges) > 0 {
		lastRange := ranges[len(ranges)-1]
		if lastRange.End.Sub(lastRange.Start) < duration {
			ranges = ranges[:len(ranges)-1]
		}
	}

	return ranges
}

// splitExact splits time range starting exactly from the start time
func splitExact(start, end time.Time, duration time.Duration) []TimeRange {
	var ranges []TimeRange
	current := start

	for current.Before(end) {
		rangeEnd := current.Add(duration)
		if rangeEnd.After(end) {
			rangeEnd = end
		}

		ranges = append(ranges, TimeRange{
			Start: current,
			End:   rangeEnd,
		})

		current = current.Add(duration)
	}

	return ranges
}

// splitAligned splits time range aligned to natural boundaries
func splitAligned(start, end time.Time, duration time.Duration) []TimeRange {
	var ranges []TimeRange

	// Find the first aligned boundary after start
	current := alignToNext(start, duration)

	for current.Before(end) {
		rangeEnd := current.Add(duration)
		if rangeEnd.After(end) {
			break // Don't include partial ranges at the end for aligned mode
		}

		ranges = append(ranges, TimeRange{
			Start: current,
			End:   rangeEnd,
		})

		current = rangeEnd
	}

	return ranges
}

// alignToNext finds the next aligned time boundary for the given duration
func alignToNext(t time.Time, duration time.Duration) time.Time {
	switch {
	case duration >= 24*time.Hour:
		// For day or longer durations, align to start of next day
		year, month, day := t.Date()
		return time.Date(year, month, day+1, 0, 0, 0, 0, t.Location())

	case duration >= time.Hour:
		// For hour durations, align to next hour
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())

	case duration >= time.Minute:
		// For minute durations, align to next minute
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, t.Location())

	default:
		// For sub-minute durations, align to next second
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()+1, 0, t.Location())
	}
}