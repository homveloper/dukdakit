package diffit

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures for array testing
type ArrayFields struct {
	StringSlice    []string               `bson:"string_slice"`
	IntSlice       []int                  `bson:"int_slice"`
	FloatSlice     []float64              `bson:"float_slice"`
	BoolSlice      []bool                 `bson:"bool_slice"`
	TimeSlice      []time.Time            `bson:"time_slice"`
	DurationSlice  []time.Duration        `bson:"duration_slice"`
	InterfaceSlice []interface{}          `bson:"interface_slice"`
	PointerSlice   []*string              `bson:"pointer_slice"`
	StructSlice    []SimpleStruct         `bson:"struct_slice"`
	StructPtrSlice []*SimpleStruct        `bson:"struct_ptr_slice"`
	NestedSlice    [][]string             `bson:"nested_slice"`
	MapSlice       []map[string]string    `bson:"map_slice"`
}

type SimpleStruct struct {
	ID   int    `bson:"id"`
	Name string `bson:"name"`
}

func TestDiff_Arrays_PrimitiveTypes(t *testing.T) {
	t.Run("StringSlice", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{"a", "b", "c"},
		}
		newData := ArrayFields{
			StringSlice: []string{"a", "modified", "c", "d"},
		}

		testArrayStrategy(t, oldData, newData, "string_slice")
	})

	t.Run("IntSlice", func(t *testing.T) {
		oldData := ArrayFields{
			IntSlice: []int{1, 2, 3, 4},
		}
		newData := ArrayFields{
			IntSlice: []int{1, 20, 3}, // Modified and removed
		}

		testArrayStrategy(t, oldData, newData, "int_slice")
	})

	t.Run("FloatSlice", func(t *testing.T) {
		oldData := ArrayFields{
			FloatSlice: []float64{1.1, 2.2, 3.3},
		}
		newData := ArrayFields{
			FloatSlice: []float64{1.1, 2.5, 3.3, 4.4}, // Modified and added
		}

		testArrayStrategy(t, oldData, newData, "float_slice")
	})

	t.Run("BoolSlice", func(t *testing.T) {
		oldData := ArrayFields{
			BoolSlice: []bool{true, false, true},
		}
		newData := ArrayFields{
			BoolSlice: []bool{false, false, true, false}, // Modified and added
		}

		testArrayStrategy(t, oldData, newData, "bool_slice")
	})
}

func TestDiff_Arrays_TimeTypes(t *testing.T) {
	base := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	later := base.Add(time.Hour)
	much_later := base.Add(2 * time.Hour)

	t.Run("TimeSlice", func(t *testing.T) {
		oldData := ArrayFields{
			TimeSlice: []time.Time{base, later},
		}
		newData := ArrayFields{
			TimeSlice: []time.Time{base, much_later, base}, // Modified and added
		}

		testArrayStrategy(t, oldData, newData, "time_slice")
	})

	t.Run("DurationSlice", func(t *testing.T) {
		oldData := ArrayFields{
			DurationSlice: []time.Duration{time.Hour, time.Minute},
		}
		newData := ArrayFields{
			DurationSlice: []time.Duration{time.Hour, time.Second, time.Minute}, // Modified and added
		}

		testArrayStrategy(t, oldData, newData, "duration_slice")
	})
}

func TestDiff_Arrays_PointerTypes(t *testing.T) {
	str1 := "first"
	str2 := "second"
	str3 := "third"
	str4 := "fourth"

	t.Run("PointerSlice", func(t *testing.T) {
		oldData := ArrayFields{
			PointerSlice: []*string{&str1, &str2, nil},
		}
		newData := ArrayFields{
			PointerSlice: []*string{&str1, &str3, &str4}, // Modified pointer and replaced nil
		}

		testArrayStrategy(t, oldData, newData, "pointer_slice")
	})
}

func TestDiff_Arrays_InterfaceTypes(t *testing.T) {
	t.Run("InterfaceSlice", func(t *testing.T) {
		oldData := ArrayFields{
			InterfaceSlice: []interface{}{1, "string", true, 3.14},
		}
		newData := ArrayFields{
			InterfaceSlice: []interface{}{1, "modified", false, 3.14, "new"}, // Mixed modifications
		}

		testArrayStrategy(t, oldData, newData, "interface_slice")
	})
}

func TestDiff_Arrays_StructTypes(t *testing.T) {
	t.Run("StructSlice", func(t *testing.T) {
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
				{ID: 4, Name: "fourth"}, // Added struct
			},
		}

		testArrayStrategy(t, oldData, newData, "struct_slice")
	})

	t.Run("StructPtrSlice", func(t *testing.T) {
		oldData := ArrayFields{
			StructPtrSlice: []*SimpleStruct{
				{ID: 1, Name: "first"},
				{ID: 2, Name: "second"},
				nil,
			},
		}
		newData := ArrayFields{
			StructPtrSlice: []*SimpleStruct{
				{ID: 1, Name: "modified"}, // Modified struct field
				{ID: 2, Name: "second"},
				{ID: 3, Name: "third"}, // Replaced nil with struct
			},
		}

		testArrayStrategy(t, oldData, newData, "struct_ptr_slice")
	})
}

func TestDiff_Arrays_NestedTypes(t *testing.T) {
	t.Run("NestedSlice", func(t *testing.T) {
		oldData := ArrayFields{
			NestedSlice: [][]string{
				{"a", "b"},
				{"c", "d"},
			},
		}
		newData := ArrayFields{
			NestedSlice: [][]string{
				{"a", "modified"}, // Modified nested array
				{"c", "d"},
				{"e", "f"}, // Added nested array
			},
		}

		testArrayStrategy(t, oldData, newData, "nested_slice")
	})

	t.Run("MapSlice", func(t *testing.T) {
		oldData := ArrayFields{
			MapSlice: []map[string]string{
				{"key1": "value1", "key2": "value2"},
				{"key3": "value3"},
			},
		}
		newData := ArrayFields{
			MapSlice: []map[string]string{
				{"key1": "modified", "key2": "value2"}, // Modified map value
				{"key3": "value3", "key4": "value4"},   // Added map key
			},
		}

		testArrayStrategy(t, oldData, newData, "map_slice")
	})
}

func TestDiff_Arrays_EdgeCases(t *testing.T) {
	t.Run("EmptyToNonEmpty", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{},
		}
		newData := ArrayFields{
			StringSlice: []string{"new"},
		}

		testArrayStrategy(t, oldData, newData, "string_slice")
	})

	t.Run("NonEmptyToEmpty", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{"existing"},
		}
		newData := ArrayFields{
			StringSlice: []string{},
		}

		testArrayStrategy(t, oldData, newData, "string_slice")
	})

	t.Run("NilToSlice", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: nil,
		}
		newData := ArrayFields{
			StringSlice: []string{"new"},
		}

		testArrayStrategy(t, oldData, newData, "string_slice")
	})

	t.Run("SliceToNil", func(t *testing.T) {
		oldData := ArrayFields{
			StringSlice: []string{"existing"},
		}
		newData := ArrayFields{
			StringSlice: nil,
		}

		testArrayStrategy(t, oldData, newData, "string_slice")
	})

	t.Run("IdenticalSlices", func(t *testing.T) {
		data := ArrayFields{
			StringSlice: []string{"a", "b", "c"},
		}

		patch, err := Diff(data, data)
		require.NoError(t, err)
		assert.True(t, patch.IsEmpty(), "Identical slices should result in empty patch")
	})
}

func TestDiff_Arrays_Strategies(t *testing.T) {
	oldData := ArrayFields{
		StringSlice: []string{"a", "b", "c"},
	}
	newData := ArrayFields{
		StringSlice: []string{"a", "modified", "c", "d"},
	}

	strategies := []struct {
		name     string
		strategy ArrayStrategy
	}{
		{"ArrayReplace", ArrayReplace},
		{"ArraySmart", ArraySmart},
		{"ArrayAppend", ArrayAppend},
		{"ArrayMerge", ArrayMerge},
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
					assert.Greater(t, len(arrayFilters), 0, "Should have array filters")
				}
			}
		})
	}
}

func TestDiff_Arrays_ArrayFilters(t *testing.T) {
	t.Run("StructSliceWithFilters", func(t *testing.T) {
		oldData := ArrayFields{
			StructSlice: []SimpleStruct{
				{ID: 1, Name: "first"},
				{ID: 2, Name: "second"},
				{ID: 3, Name: "third"},
			},
		}
		newData := ArrayFields{
			StructSlice: []SimpleStruct{
				{ID: 1, Name: "modified_first"}, // Should generate array filter
				{ID: 2, Name: "second"},
				{ID: 3, Name: "modified_third"}, // Should generate array filter
			},
		}

		// Test with smart strategy to see if array filters are generated
		patch, err := Diff(oldData, newData, WithArrayStrategy(ArraySmart))
		require.NoError(t, err)
		
		t.Logf("Smart array patch: %s", patch)
		
		// For now, check if we have operations (array filters implementation is TODO)
		assert.False(t, patch.IsEmpty(), "Should have operations for struct slice modifications")
		
		operations := patch.Operations()
		t.Logf("Operations: %+v", operations)
		
		if patch.HasArrayFilters() {
			arrayFilters := patch.ArrayFilters()
			t.Logf("Array filters generated: %+v", arrayFilters)
			// When implemented, we should have filters for matching specific array elements
		}
	})
}

// Helper function to test different array strategies
func testArrayStrategy(t *testing.T, oldData, newData ArrayFields, fieldName string) {
	t.Helper()
	
	// Test with default strategy
	patch, err := Diff(oldData, newData)
	require.NoError(t, err, "Diff should not return error")
	
	t.Logf("Field: %s", fieldName)
	t.Logf("Old: %+v", getFieldValue(oldData, fieldName))
	t.Logf("New: %+v", getFieldValue(newData, fieldName))
	
	if !patch.IsEmpty() {
		t.Logf("Patch: %s", patch)
		
		operations := patch.Operations()
		t.Logf("Operations: %+v", operations)
		
		// For arrays, we expect $set operations (either direct replace or array filters)
		if setOps, ok := operations["$set"]; ok {
			setMap := setOps.(map[string]interface{})
			
			// Check if direct field replacement is used
			hasDirectField := false
			hasArrayFilter := false
			
			if _, exists := setMap[fieldName]; exists {
				hasDirectField = true
			}
			
			// Check if array filter patterns are used
			for key := range setMap {
				if strings.HasPrefix(key, fieldName+".$[") {
					hasArrayFilter = true
					break
				}
			}
			
			assert.True(t, hasDirectField || hasArrayFilter, 
				"Should contain either direct field '%s' or array filter pattern '%s.$[...]'", 
				fieldName, fieldName)
		}
		
		// Check if array filters were generated
		if patch.HasArrayFilters() {
			arrayFilters := patch.ArrayFilters()
			t.Logf("Array filters: %+v", arrayFilters)
		}
	} else {
		t.Log("Patch is empty - no changes detected")
	}
}

// Helper to get field value using field name
func getFieldValue(data ArrayFields, fieldName string) interface{} {
	switch fieldName {
	case "string_slice":
		return data.StringSlice
	case "int_slice":
		return data.IntSlice
	case "float_slice":
		return data.FloatSlice
	case "bool_slice":
		return data.BoolSlice
	case "time_slice":
		return data.TimeSlice
	case "duration_slice":
		return data.DurationSlice
	case "interface_slice":
		return data.InterfaceSlice
	case "pointer_slice":
		return data.PointerSlice
	case "struct_slice":
		return data.StructSlice
	case "struct_ptr_slice":
		return data.StructPtrSlice
	case "nested_slice":
		return data.NestedSlice
	case "map_slice":
		return data.MapSlice
	default:
		return nil
	}
}