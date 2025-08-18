package diffit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDebug_StructArrayStrategies(t *testing.T) {
	oldData := ArrayFields{
		StructSlice: []SimpleStruct{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
			{ID: 3, Name: "third"},
		},
	}
	newData := ArrayFields{
		StructSlice: []SimpleStruct{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "modified"}, // Modified struct field
			{ID: 3, Name: "third"},
		},
	}

	strategies := []struct {
		name     string
		strategy ArrayStrategy
	}{
		{"ArrayReplace", ArrayReplace},
		{"ArraySmart", ArraySmart},
	}

	for _, s := range strategies {
		t.Run(s.name, func(t *testing.T) {
			patch, err := Diff(oldData, newData, WithArrayStrategy(s.strategy))
			require.NoError(t, err, "Diff should not return error for %s", s.name)
			
			t.Logf("Strategy: %s", s.name)
			t.Logf("Patch: %s", patch)
			
			if !patch.IsEmpty() {
				operations := patch.Operations()
				t.Logf("Operations: %+v", operations)
				
				// Check if patch has array filters when expected
				if patch.HasArrayFilters() {
					arrayFilters := patch.ArrayFilters()
					t.Logf("Array filters: %+v", arrayFilters)
				}
			}
		})
	}
}