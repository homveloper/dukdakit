package diffit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple struct with just time fields for debugging
type TimeTestStruct struct {
	TimeField time.Time `bson:"time_field"`
	TimePtr   *time.Time `bson:"time_ptr"`
	Duration  time.Duration `bson:"duration"`
}

func TestDebug_TimeComparison(t *testing.T) {
	base := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	later := base.Add(2 * time.Hour)

	oldStruct := TimeTestStruct{
		TimeField: base,
		TimePtr:   &base,
		Duration:  time.Hour,
	}

	newStruct := TimeTestStruct{
		TimeField: later,
		TimePtr:   &later,
		Duration:  time.Hour * 2,
	}

	t.Logf("Old TimeField: %v", oldStruct.TimeField)
	t.Logf("New TimeField: %v", newStruct.TimeField)
	t.Logf("Are they equal? %v", oldStruct.TimeField.Equal(newStruct.TimeField))

	patch, err := Diff(oldStruct, newStruct)
	require.NoError(t, err, "Diff should not return error")
	
	t.Logf("Simple time patch: %s", patch)
	t.Logf("Operations: %+v", patch.Operations())
	
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")
	
	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")
	
	setMap := setOps.(map[string]interface{})
	assert.Contains(t, setMap, "time_field", "TimeField change should be detected")
	assert.Contains(t, setMap, "time_ptr", "TimePtr change should be detected")
	assert.Contains(t, setMap, "duration", "Duration change should be detected")
}