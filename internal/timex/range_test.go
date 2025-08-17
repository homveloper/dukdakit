package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRange_AlignedSplit_Default(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	// Test default behavior (should be aligned)
	ranges := Range(start, end, duration)

	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)
}

func TestRange_ExactSplit(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithExactSplit())

	require.Len(t, ranges, 3)

	assert.Equal(t, start, ranges[0].Start)
	assert.Equal(t, start.Add(time.Hour), ranges[0].End)

	assert.Equal(t, start.Add(time.Hour), ranges[1].Start)
	assert.Equal(t, start.Add(2*time.Hour), ranges[1].End)

	assert.Equal(t, start.Add(2*time.Hour), ranges[2].Start)
	assert.Equal(t, end, ranges[2].End)
}

func TestRange_AlignedSplit(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithAlignedSplit())

	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)
}

func TestRange_AlignedSplit_Minutes(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 14, 39, 40, 0, time.UTC)
	duration := time.Minute

	ranges := Range(start, end, duration, WithAlignedSplit())

	require.Len(t, ranges, 3)

	expectedStart1 := time.Date(2024, 1, 1, 14, 36, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 14, 37, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 1, 14, 37, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 1, 14, 38, 0, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)

	expectedStart3 := time.Date(2024, 1, 1, 14, 38, 0, 0, time.UTC)
	expectedEnd3 := time.Date(2024, 1, 1, 14, 39, 0, 0, time.UTC)
	assert.Equal(t, expectedStart3, ranges[2].Start)
	assert.Equal(t, expectedEnd3, ranges[2].End)
}

func TestRange_WithTrimFirst(t *testing.T) {
	// Default is aligned split now
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithTrimFirst())

	// Default aligned split creates ranges starting at 15:00, 16:00
	// WithTrimFirst should keep all ranges since they're all full hours
	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)
}

func TestRange_WithTrimLast(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithExactSplit(), WithTrimLast())

	require.Len(t, ranges, 2)

	assert.Equal(t, start, ranges[0].Start)
	assert.Equal(t, start.Add(time.Hour), ranges[0].End)

	assert.Equal(t, start.Add(time.Hour), ranges[1].Start)
	assert.Equal(t, start.Add(2*time.Hour), ranges[1].End)
}

func TestRange_WithTrimBoth(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithTrimFirst(), WithTrimLast())

	// Default aligned split creates ranges starting at 15:00, 16:00
	// All ranges are full hours, so none should be trimmed
	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)
}

func TestRange_WithTrim(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 17, 25, 40, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithTrim())

	// WithTrim should be equivalent to WithTrimFirst() + WithTrimLast()
	// Default aligned split creates ranges starting at 15:00, 16:00
	// All ranges are full hours, so none should be trimmed
	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)
}

func TestRange_EdgeCases(t *testing.T) {
	t.Run("Start after end", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		duration := time.Hour

		ranges := Range(start, end, duration)
		assert.Nil(t, ranges)
	})

	t.Run("Zero duration", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
		duration := time.Duration(0)

		ranges := Range(start, end, duration)
		assert.Nil(t, ranges)
	})

	t.Run("Negative duration", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)
		duration := -time.Hour

		ranges := Range(start, end, duration)
		assert.Nil(t, ranges)
	})

	t.Run("Equal start and end", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		end := start
		duration := time.Hour

		ranges := Range(start, end, duration)
		assert.Empty(t, ranges)
	})
}

func TestRange_ExactDuration(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
	duration := time.Hour

	ranges := Range(start, end, duration, WithExactSplit())

	require.Len(t, ranges, 2)

	assert.Equal(t, start, ranges[0].Start)
	assert.Equal(t, start.Add(time.Hour), ranges[0].End)

	assert.Equal(t, start.Add(time.Hour), ranges[1].Start)
	assert.Equal(t, end, ranges[1].End)
}

func TestRange_DayAlignment(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 4, 10, 25, 40, 0, time.UTC)
	duration := 24 * time.Hour

	ranges := Range(start, end, duration, WithAlignedSplit())

	require.Len(t, ranges, 2)

	expectedStart1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)
}

func TestRange_SubMinuteAlignment(t *testing.T) {
	start := time.Date(2024, 1, 1, 14, 35, 20, 500000000, time.UTC) // 500ms
	end := time.Date(2024, 1, 1, 14, 35, 24, 0, time.UTC)
	duration := time.Second

	ranges := Range(start, end, duration, WithAlignedSplit())

	require.Len(t, ranges, 3)

	expectedStart1 := time.Date(2024, 1, 1, 14, 35, 21, 0, time.UTC)
	expectedEnd1 := time.Date(2024, 1, 1, 14, 35, 22, 0, time.UTC)
	assert.Equal(t, expectedStart1, ranges[0].Start)
	assert.Equal(t, expectedEnd1, ranges[0].End)

	expectedStart2 := time.Date(2024, 1, 1, 14, 35, 22, 0, time.UTC)
	expectedEnd2 := time.Date(2024, 1, 1, 14, 35, 23, 0, time.UTC)
	assert.Equal(t, expectedStart2, ranges[1].Start)
	assert.Equal(t, expectedEnd2, ranges[1].End)

	expectedStart3 := time.Date(2024, 1, 1, 14, 35, 23, 0, time.UTC)
	expectedEnd3 := time.Date(2024, 1, 1, 14, 35, 24, 0, time.UTC)
	assert.Equal(t, expectedStart3, ranges[2].Start)
	assert.Equal(t, expectedEnd3, ranges[2].End)
}

func TestRange_WithTrimPartialRanges(t *testing.T) {
	// Create a scenario with actual partial ranges
	start := time.Date(2024, 1, 1, 14, 35, 20, 0, time.UTC)
	end := time.Date(2024, 1, 1, 16, 45, 30, 0, time.UTC) // Partial last range
	duration := time.Hour

	t.Run("ExactSplit with partial last range", func(t *testing.T) {
		ranges := Range(start, end, duration, WithExactSplit(), WithTrimLast())
		
		// Should have 2 full ranges, last partial range trimmed
		require.Len(t, ranges, 2)
		
		assert.Equal(t, start, ranges[0].Start)
		assert.Equal(t, start.Add(time.Hour), ranges[0].End)
		
		assert.Equal(t, start.Add(time.Hour), ranges[1].Start)
		assert.Equal(t, start.Add(2*time.Hour), ranges[1].End)
	})

	t.Run("ExactSplit without trim", func(t *testing.T) {
		ranges := Range(start, end, duration, WithExactSplit())
		
		// Should have 3 ranges including partial last range
		require.Len(t, ranges, 3)
		
		// Last range should be partial
		assert.Equal(t, start.Add(2*time.Hour), ranges[2].Start)
		assert.Equal(t, end, ranges[2].End)
		assert.True(t, ranges[2].End.Sub(ranges[2].Start) < duration)
	})

	t.Run("WithTrim actually removes partial ranges", func(t *testing.T) {
		// Create a scenario where we have actual partial ranges to trim
		start2 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
		end2 := time.Date(2024, 1, 1, 15, 45, 0, 0, time.UTC)
		
		ranges := Range(start2, end2, 30*time.Minute, WithExactSplit(), WithTrim())
		
		// Without trim: [14:30-15:00, 15:00-15:30, 15:30-15:45] (last is partial)
		// With trim: should remove the partial last range
		require.Len(t, ranges, 2)
		
		assert.Equal(t, start2, ranges[0].Start)
		assert.Equal(t, start2.Add(30*time.Minute), ranges[0].End)
		
		assert.Equal(t, start2.Add(30*time.Minute), ranges[1].Start)
		assert.Equal(t, start2.Add(60*time.Minute), ranges[1].End)
	})
}