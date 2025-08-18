package diffit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDebug_ArrayAppendStrategy(t *testing.T) {
	t.Run("PureAppend", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{"a", "b"},
		}
		newData := ArrayFields{
			StringSlice: []string{"a", "b", "c", "d"}, // Pure append
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayAppend))
		require.NoError(t, err)
		
		t.Logf("Pure append patch: %s", patch)
		
		operations := patch.Operations()
		t.Logf("Operations: %+v", operations)
		
		// Should use $push operations
		if pushOps, ok := operations["$push"]; ok {
			t.Logf("Push operations found: %+v", pushOps)
		}
	})

	t.Run("ModifiedAndAppend", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{"a", "b"},
		}
		newData := ArrayFields{
			StringSlice: []string{"modified", "b", "c"}, // Modified first element
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayAppend))
		require.NoError(t, err)
		
		t.Logf("Modified and append patch: %s", patch)
		
		// Should fall back to replace since elements were modified
		operations := patch.Operations()
		if setOps, ok := operations["$set"]; ok {
			t.Logf("Fell back to $set (replace): %+v", setOps)
		}
	})

	t.Run("StructAppend", func(t *testing.T) {
		oldData := ArrayFields{
			StructSlice: []SimpleStruct{
				{ID: 1, Name: "first"},
			},
		}
		newData := ArrayFields{
			StructSlice: []SimpleStruct{
				{ID: 1, Name: "first"},
				{ID: 2, Name: "second"}, // Appended struct
				{ID: 3, Name: "third"},  // Appended struct
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayAppend))
		require.NoError(t, err)
		
		t.Logf("Struct append patch: %s", patch)
		
		operations := patch.Operations()
		t.Logf("Operations: %+v", operations)
	})
}