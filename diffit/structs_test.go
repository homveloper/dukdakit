package diffit

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DiverseStruct contains various field types for comprehensive testing
type DiverseStruct struct {
	// Primitive types
	StringField  string  `bson:"string_field"`
	IntField     int     `bson:"int_field"`
	Int32Field   int32   `bson:"int32_field"`
	Int64Field   int64   `bson:"int64_field"`
	Float32Field float32 `bson:"float32_field"`
	Float64Field float64 `bson:"float64_field"`
	BoolField    bool    `bson:"bool_field"`
	UintField    uint    `bson:"uint_field"`

	// Pointer types
	StringPtr  *string  `bson:"string_ptr,omitempty"`
	IntPtr     *int     `bson:"int_ptr,omitempty"`
	Float64Ptr *float64 `bson:"float64_ptr,omitempty"`
	BoolPtr    *bool    `bson:"bool_ptr,omitempty"`

	// Time types
	TimeField     time.Time     `bson:"time_field"`
	TimePtr       *time.Time    `bson:"time_ptr,omitempty"`
	DurationField time.Duration `bson:"duration_field"`
	DurationPtr   *time.Duration `bson:"duration_ptr,omitempty"`

	// Collection types
	StringSlice  []string               `bson:"string_slice"`
	IntSlice     []int                  `bson:"int_slice"`
	StringMap    map[string]string      `bson:"string_map"`
	InterfaceMap map[string]interface{} `bson:"interface_map"`

	// Nested struct
	Nested    NestedStruct  `bson:"nested"`
	NestedPtr *NestedStruct `bson:"nested_ptr,omitempty"`
}

// NestedStruct for testing deep nesting scenarios
type NestedStruct struct {
	Level1Field string    `bson:"level1_field"`
	Level1Int   int       `bson:"level1_int"`
	Level1Bool  bool      `bson:"level1_bool"`
	Level1Time  time.Time `bson:"level1_time"`
	Level1Ptr   *string   `bson:"level1_ptr,omitempty"`

	// Deep nesting
	Deep DeepNested `bson:"deep"`
}

// DeepNested for testing very deep nesting
type DeepNested struct {
	Level2Field    string           `bson:"level2_field"`
	Level2Duration time.Duration    `bson:"level2_duration"`
	Level2Map      map[string]int   `bson:"level2_map"`
}

func TestDiff_Structs_DiverseFieldTypes(t *testing.T) {
	str1 := "hello"
	int1 := 42
	float1 := 3.14
	bool1 := true
	duration1 := time.Hour
	now := time.Now()

	oldStruct := DiverseStruct{
		// Primitives
		StringField:  "original",
		IntField:     10,
		Int32Field:   20,
		Int64Field:   100,
		Float32Field: 1.5,
		Float64Field: 2.5,
		BoolField:    false,
		UintField:    50,

		// Pointers
		StringPtr:  &str1,
		IntPtr:     &int1,
		Float64Ptr: &float1,
		BoolPtr:    &bool1,

		// Time types
		TimeField:     now,
		TimePtr:       &now,
		DurationField: time.Minute,
		DurationPtr:   &duration1,

		// Collections
		StringSlice:  []string{"a", "b", "c"},
		IntSlice:     []int{1, 2, 3},
		StringMap:    map[string]string{"key1": "value1"},
		InterfaceMap: map[string]interface{}{"count": 5, "active": true},

		// Nested
		Nested: NestedStruct{
			Level1Field: "nested_original",
			Level1Int:   99,
			Level1Bool:  true,
			Level1Time:  now,
			Level1Ptr:   &str1,
		},
	}

	str2 := "world"
	float2 := 6.28
	duration2 := time.Minute * 30
	later := now.Add(time.Hour)

	newStruct := oldStruct
	// Change various field types
	newStruct.StringField = "changed"
	newStruct.IntField = 15      // +5
	newStruct.Float64Field = 3.5 // +1.0
	newStruct.BoolField = true

	// Pointer changes
	newStruct.StringPtr = &str2   // Change pointer value
	newStruct.IntPtr = nil        // Set to nil
	newStruct.Float64Ptr = &float2 // Change pointer value

	// Time changes
	newStruct.TimePtr = &later
	newStruct.DurationField = time.Hour * 2
	newStruct.DurationPtr = &duration2

	// Collection changes
	newStruct.StringSlice = []string{"x", "y", "z"}
	newStruct.StringMap = map[string]string{"key1": "changed_value", "key2": "new_value"}
	newStruct.InterfaceMap = map[string]interface{}{"count": 10, "status": "updated"}

	// Nested changes
	newStruct.Nested.Level1Field = "nested_changed"
	newStruct.Nested.Level1Int = 150 // +51
	newStruct.Nested.Level1Bool = false

	patch, err := Diff(oldStruct, newStruct, WithNumericOptimization(true))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()

	// Numeric optimization is disabled - all changes should be in $set

	// Check $set operations for other changes
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})

	// Primitive changes
	assert.Equal(t, "changed", setMap["string_field"], "StringField should be changed")
	assert.Equal(t, 15, setMap["int_field"], "IntField should be changed")
	assert.Equal(t, 3.5, setMap["float64_field"], "Float64Field should be changed")
	assert.Equal(t, true, setMap["bool_field"], "BoolField should be changed")

	// Pointer changes
	assert.Equal(t, "world", setMap["string_ptr"], "StringPtr should be changed")
	assert.Equal(t, 6.28, setMap["float64_ptr"], "Float64Ptr should be changed")

	// Time changes
	assert.Equal(t, later, setMap["time_ptr"], "TimePtr should be changed")
	assert.Equal(t, time.Hour*2, setMap["duration_field"], "DurationField should be changed")
	assert.Equal(t, duration2, setMap["duration_ptr"], "DurationPtr should be changed")

	// Collection changes (for now, arrays are replaced entirely)
	assert.Contains(t, setMap, "string_slice", "StringSlice should be updated")

	// Map field changes
	assert.Equal(t, "changed_value", setMap["string_map.key1"], "StringMap key1 should be changed")
	assert.Equal(t, "new_value", setMap["string_map.key2"], "StringMap key2 should be added")
	assert.Equal(t, 10, setMap["interface_map.count"], "InterfaceMap count should be changed")
	assert.Equal(t, "updated", setMap["interface_map.status"], "InterfaceMap status should be added")

	// Nested struct changes
	assert.Equal(t, "nested_changed", setMap["nested.level1_field"], "Nested field should be changed")
	assert.Equal(t, 150, setMap["nested.level1_int"], "Nested int should be changed")
	assert.Equal(t, false, setMap["nested.level1_bool"], "Nested bool should be changed")

	// Check $unset operations for nil pointers
	if unsetOps, ok := operations["$unset"]; ok {
		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "int_ptr", "IntPtr should be unset to nil")
	}

	t.Logf("Diverse field types patch: %s", patch)
}

func TestDiff_Structs_DeepNesting(t *testing.T) {
	str1 := "deep_original"
	now := time.Now()

	oldStruct := DiverseStruct{
		StringField: "root_field",
		IntField:    100,
		Nested: NestedStruct{
			Level1Field: "level1_original",
			Level1Int:   50,
			Level1Bool:  true,
			Level1Time:  now,
			Level1Ptr:   &str1,
			Deep: DeepNested{
				Level2Field:    "level2_original",
				Level2Duration: time.Minute * 5,
				Level2Map:      map[string]int{"score": 100, "lives": 3},
			},
		},
	}

	str2 := "deep_changed"
	later := now.Add(time.Hour)

	newStruct := oldStruct
	// Change fields at different nesting levels
	newStruct.StringField = "root_changed"              // Root level
	newStruct.IntField = 150                           // Root level, +50
	newStruct.Nested.Level1Field = "level1_changed"   // Level 1
	newStruct.Nested.Level1Int = 75                    // Level 1, +25
	newStruct.Nested.Level1Bool = false                // Level 1
	newStruct.Nested.Level1Time = later                // Level 1
	newStruct.Nested.Level1Ptr = &str2                 // Level 1 pointer
	newStruct.Nested.Deep.Level2Field = "level2_changed" // Level 2
	newStruct.Nested.Deep.Level2Duration = time.Minute * 10 // Level 2
	newStruct.Nested.Deep.Level2Map = map[string]int{"score": 200, "lives": 5, "bonus": 50} // Level 2 map

	patch, err := Diff(oldStruct, newStruct, WithNumericOptimization(true))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()

	// Check $set operations
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})
	
	// Numeric optimization is disabled - all changes should be in $set
	// Check for numeric field changes in $set operations
	assert.Equal(t, 150, setMap["int_field"], "IntField should be changed to 150")
	assert.Equal(t, 75, setMap["nested.level1_int"], "Level1 int should be changed to 75")

	// Root level changes
	assert.Equal(t, "root_changed", setMap["string_field"], "Root string should be changed")

	// Level 1 changes
	assert.Equal(t, "level1_changed", setMap["nested.level1_field"], "Level1 field should be changed")
	assert.Equal(t, false, setMap["nested.level1_bool"], "Level1 bool should be changed")
	assert.Equal(t, later, setMap["nested.level1_time"], "Level1 time should be changed")
	assert.Equal(t, "deep_changed", setMap["nested.level1_ptr"], "Level1 pointer should be changed")

	// Level 2 changes
	assert.Equal(t, "level2_changed", setMap["nested.deep.level2_field"], "Level2 field should be changed")
	assert.Equal(t, time.Minute*10, setMap["nested.deep.level2_duration"], "Level2 duration should be changed")

	// Level 2 map changes
	assert.Equal(t, 200, setMap["nested.deep.level2_map.score"], "Level2 map score should be changed")
	assert.Equal(t, 5, setMap["nested.deep.level2_map.lives"], "Level2 map lives should be changed")
	assert.Equal(t, 50, setMap["nested.deep.level2_map.bonus"], "Level2 map bonus should be added")

	t.Logf("Deep nesting patch: %s", patch)
}

func TestDiff_Structs_PointerHandling(t *testing.T) {
	str1 := "original"
	float1 := 3.14
	bool1 := true
	now := time.Now()

	oldStruct := DiverseStruct{
		StringField: "test",
		// Mix of nil and non-nil pointers
		StringPtr:   &str1,
		IntPtr:      nil,
		Float64Ptr:  &float1,
		BoolPtr:     &bool1,
		TimePtr:     &now,
		DurationPtr: nil,
		NestedPtr:   nil,
	}

	str2 := "changed"
	int2 := 200
	bool2 := false
	duration2 := time.Minute * 30
	later := now.Add(time.Hour)

	newStruct := oldStruct
	// Various pointer operations
	newStruct.StringPtr = &str2      // Change pointer value
	newStruct.IntPtr = &int2         // nil to value
	newStruct.Float64Ptr = nil       // value to nil
	newStruct.BoolPtr = &bool2       // Change pointer value
	newStruct.TimePtr = &later       // Change time pointer
	newStruct.DurationPtr = &duration2 // nil to value
	newStruct.NestedPtr = &NestedStruct{ // nil to struct
		Level1Field: "new_nested",
		Level1Int:   999,
		Level1Bool:  true,
		Level1Time:  now,
	}

	patch, err := Diff(oldStruct, newStruct)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()

	// Check $set operations for pointer changes and nil-to-value
	if setOps, ok := operations["$set"]; ok {
		setMap := setOps.(map[string]interface{})

		// Pointer value changes
		assert.Equal(t, "changed", setMap["string_ptr"], "StringPtr should change value")
		assert.Equal(t, 200, setMap["int_ptr"], "IntPtr should be set from nil")
		assert.Equal(t, false, setMap["bool_ptr"], "BoolPtr should change value")
		assert.Equal(t, later, setMap["time_ptr"], "TimePtr should change time")
		assert.Equal(t, duration2, setMap["duration_ptr"], "DurationPtr should be set from nil")

		// Nested struct pointer (nil to struct)
		assert.Equal(t, "new_nested", setMap["nested_ptr.level1_field"], "NestedPtr field should be set")
		assert.Equal(t, 999, setMap["nested_ptr.level1_int"], "NestedPtr int should be set")
		assert.Equal(t, true, setMap["nested_ptr.level1_bool"], "NestedPtr bool should be set")
		assert.Equal(t, now, setMap["nested_ptr.level1_time"], "NestedPtr time should be set")
	}

	// Check $unset operations for value-to-nil
	if unsetOps, ok := operations["$unset"]; ok {
		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "float64_ptr", "Float64Ptr should be unset")
	}

	t.Logf("Pointer handling patch: %s", patch)
}

func TestDiff_Structs_TimeTypes(t *testing.T) {
	base := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	later := base.Add(2 * time.Hour)
	duration1 := time.Hour * 2
	duration2 := time.Minute * 30

	oldStruct := DiverseStruct{
		StringField:   "time_test",
		TimeField:     base,
		TimePtr:       &base,
		DurationField: duration1,
		DurationPtr:   &duration1,
		Nested: NestedStruct{
			Level1Field: "nested_time",
			Level1Time:  base,
			Deep: DeepNested{
				Level2Field:    "deep_time",
				Level2Duration: duration2,
			},
		},
	}

	duration3 := time.Hour * 5
	duration4 := time.Minute * 45

	newStruct := oldStruct
	// Change time-related fields
	newStruct.TimeField = later
	newStruct.TimePtr = &later
	newStruct.DurationField = duration3
	newStruct.DurationPtr = &duration3
	newStruct.Nested.Level1Time = later
	newStruct.Nested.Deep.Level2Duration = duration4

	t.Logf("Old TimeField: %v", oldStruct.TimeField)
	t.Logf("New TimeField: %v", newStruct.TimeField)
	t.Logf("Old Level1Time: %v", oldStruct.Nested.Level1Time)
	t.Logf("New Level1Time: %v", newStruct.Nested.Level1Time)

	patch, err := Diff(oldStruct, newStruct)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	t.Logf("Time patch operations: %s", patch)
	t.Logf("Operations detail: %+v", patch.Operations())

	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})

	// Check time field changes
	assert.Equal(t, later, setMap["time_field"], "TimeField should be changed")
	assert.Equal(t, later, setMap["time_ptr"], "TimePtr should be changed")
	assert.Equal(t, duration3, setMap["duration_field"], "DurationField should be changed")
	assert.Equal(t, duration3, setMap["duration_ptr"], "DurationPtr should be changed")

	// Check nested time changes
	assert.Equal(t, later, setMap["nested.level1_time"], "Nested time should be changed")
	assert.Equal(t, duration4, setMap["nested.deep.level2_duration"], "Deep nested duration should be changed")

	t.Logf("Time types patch: %s", patch)
}

func TestDiff_Structs_CollectionFields(t *testing.T) {
	oldStruct := DiverseStruct{
		StringField: "collection_test",
		StringSlice: []string{"a", "b", "c"},
		IntSlice:    []int{1, 2, 3},
		StringMap: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		InterfaceMap: map[string]interface{}{
			"count":  10,
			"active": true,
			"score":  95.5,
		},
		Nested: NestedStruct{
			Level1Field: "nested_collection",
			Deep: DeepNested{
				Level2Map: map[string]int{
					"hp":    100,
					"mp":    50,
					"level": 5,
				},
			},
		},
	}

	newStruct := oldStruct
	// Change collection fields
	newStruct.StringSlice = []string{"x", "y", "z", "w"} // Different array
	newStruct.IntSlice = []int{10, 20, 30}              // Different array
	newStruct.StringMap = map[string]string{
		"key1": "changed_value1", // Changed
		"key2": "value2",         // Same
		"key3": "new_value3",     // Added
		// key2 removed implicitly
	}
	newStruct.InterfaceMap = map[string]interface{}{
		"count":  15,            // Changed
		"active": false,         // Changed
		"score":  95.5,          // Same
		"bonus":  25,            // Added
		"status": "completed",   // Added
	}
	newStruct.Nested.Deep.Level2Map = map[string]int{
		"hp":      120, // Changed
		"mp":      50,  // Same
		"level":   6,   // Changed
		"defense": 25,  // Added
	}

	patch, err := Diff(oldStruct, newStruct)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})

	// Arrays are replaced entirely (for now)
	assert.Contains(t, setMap, "string_slice", "StringSlice should be replaced")
	assert.Contains(t, setMap, "int_slice", "IntSlice should be replaced")

	// Map field changes
	assert.Equal(t, "changed_value1", setMap["string_map.key1"], "StringMap key1 should be changed")
	assert.Equal(t, "new_value3", setMap["string_map.key3"], "StringMap key3 should be added")

	assert.Equal(t, 15, setMap["interface_map.count"], "InterfaceMap count should be changed")
	assert.Equal(t, false, setMap["interface_map.active"], "InterfaceMap active should be changed")
	assert.Equal(t, 25, setMap["interface_map.bonus"], "InterfaceMap bonus should be added")
	assert.Equal(t, "completed", setMap["interface_map.status"], "InterfaceMap status should be added")

	// Nested map changes
	assert.Equal(t, 120, setMap["nested.deep.level2_map.hp"], "Deep map hp should be changed")
	assert.Equal(t, 6, setMap["nested.deep.level2_map.level"], "Deep map level should be changed")
	assert.Equal(t, 25, setMap["nested.deep.level2_map.defense"], "Deep map defense should be added")

	// Unchanged values should not be in patch
	assert.NotContains(t, setMap, "string_map.key2", "Unchanged string map value should not be in patch")
	assert.NotContains(t, setMap, "interface_map.score", "Unchanged interface map value should not be in patch")
	assert.NotContains(t, setMap, "nested.deep.level2_map.mp", "Unchanged deep map value should not be in patch")

	t.Logf("Collection fields patch: %s", patch)
}

func TestDiff_Structs_IgnoreFields(t *testing.T) {
	str1 := "original"
	now := time.Now()
	later := now.Add(time.Hour)

	oldStruct := DiverseStruct{
		StringField:   "test",
		IntField:      100,
		BoolField:     false,
		StringPtr:     &str1,
		TimeField:     now,
		DurationField: time.Hour,
		StringMap:     map[string]string{"key": "value"},
		Nested: NestedStruct{
			Level1Field: "nested_original",
			Level1Int:   50,
			Level1Bool:  true,
			Deep: DeepNested{
				Level2Field: "deep_original",
			},
		},
	}

	str2 := "changed"
	newStruct := oldStruct
	// Change multiple fields
	newStruct.StringField = "changed"
	newStruct.IntField = 200       // Should be ignored
	newStruct.BoolField = true
	newStruct.StringPtr = &str2
	newStruct.TimeField = later    // Should be ignored
	newStruct.DurationField = time.Minute
	newStruct.StringMap = map[string]string{"key": "new_value"}
	newStruct.Nested.Level1Field = "nested_changed"
	newStruct.Nested.Level1Int = 100 // Should be ignored
	newStruct.Nested.Level1Bool = false
	newStruct.Nested.Deep.Level2Field = "deep_changed" // Should be ignored

	// Ignore specific fields including nested ones
	patch, err := Diff(oldStruct, newStruct, WithIgnoreFields(
		"int_field",                    // Root level field
		"time_field",                   // Root level time field
		"nested.level1_int",            // Nested field
		"nested.deep.level2_field",     // Deep nested field
	))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})

	// Fields that should be changed
	assert.Equal(t, "changed", setMap["string_field"], "StringField should be changed")
	assert.Equal(t, true, setMap["bool_field"], "BoolField should be changed")
	assert.Equal(t, "changed", setMap["string_ptr"], "StringPtr should be changed")
	assert.Equal(t, time.Minute, setMap["duration_field"], "DurationField should be changed")
	assert.Equal(t, "new_value", setMap["string_map.key"], "StringMap should be changed")
	assert.Equal(t, "nested_changed", setMap["nested.level1_field"], "Nested field should be changed")
	assert.Equal(t, false, setMap["nested.level1_bool"], "Nested bool should be changed")

	// Fields that should be ignored
	assert.NotContains(t, setMap, "int_field", "IntField should be ignored")
	assert.NotContains(t, setMap, "time_field", "TimeField should be ignored")
	assert.NotContains(t, setMap, "nested.level1_int", "Nested int should be ignored")
	assert.NotContains(t, setMap, "nested.deep.level2_field", "Deep nested field should be ignored")

	// Verify ignored fields are not in metadata
	fieldList := patch.metadata.FieldsChanged
	assert.NotContains(t, fieldList, "int_field", "Ignored field should not be in metadata")
	assert.NotContains(t, fieldList, "time_field", "Ignored time field should not be in metadata")
	assert.NotContains(t, fieldList, "nested.level1_int", "Ignored nested field should not be in metadata")
	assert.NotContains(t, fieldList, "nested.deep.level2_field", "Ignored deep field should not be in metadata")

	t.Logf("Ignore fields patch: %s", patch)
}

func TestDiff_Structs_ZeroValueHandling(t *testing.T) {
	t.Run("ZeroAsSet", func(t *testing.T) {
		str1 := "original"
		now := time.Now()
		duration1 := time.Hour

		oldStruct := DiverseStruct{
			StringField:   "test",
			IntField:      100,
			BoolField:     true,
			StringPtr:     &str1,
			TimeField:     now,
			DurationField: duration1,
			StringSlice:   []string{"a", "b"},
			StringMap:     map[string]string{"key": "value"},
		}

		newStruct := oldStruct
		// Set fields to zero values
		newStruct.StringField = ""        // Zero string
		newStruct.IntField = 0           // Zero int
		newStruct.BoolField = false      // Zero bool
		newStruct.TimeField = time.Time{} // Zero time
		newStruct.DurationField = 0      // Zero duration
		newStruct.StringSlice = nil      // Zero slice
		newStruct.StringMap = nil        // Zero map

		patch, err := Diff(oldStruct, newStruct, WithZeroValueHandling(ZeroAsSet))
		require.NoError(t, err, "Diff should not return error")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")

		setMap := setOps.(map[string]interface{})
		assert.Equal(t, "", setMap["string_field"], "String should be set to empty")
		assert.Equal(t, 0, setMap["int_field"], "Int should be set to zero")
		assert.Equal(t, false, setMap["bool_field"], "Bool should be set to false")
		assert.Equal(t, time.Time{}, setMap["time_field"], "Time should be set to zero")
		assert.Equal(t, time.Duration(0), setMap["duration_field"], "Duration should be set to zero")
	})

	t.Run("ZeroAsUnset", func(t *testing.T) {
		str1 := "original"
		now := time.Now()

		oldStruct := DiverseStruct{
			StringField: "test",
			IntField:    100,
			BoolField:   true,
			StringPtr:   &str1,
			TimeField:   now,
		}

		newStruct := oldStruct
		// Set fields to zero values
		newStruct.StringField = ""
		newStruct.IntField = 0
		newStruct.BoolField = false
		newStruct.TimeField = time.Time{}

		patch, err := Diff(oldStruct, newStruct, WithZeroValueHandling(ZeroAsUnset))
		require.NoError(t, err, "Diff should not return error")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		unsetOps, ok := operations["$unset"]
		require.True(t, ok, "Should have $unset operation")

		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "string_field", "String should be unset")
		assert.Contains(t, unsetMap, "int_field", "Int should be unset")
		assert.Contains(t, unsetMap, "bool_field", "Bool should be unset")
		assert.Contains(t, unsetMap, "time_field", "Time should be unset")
	})

	t.Run("ZeroIgnore", func(t *testing.T) {
		oldStruct := DiverseStruct{
			StringField: "test",
			IntField:    100,
			BoolField:   true,
		}

		newStruct := oldStruct
		// Set fields to zero values
		newStruct.StringField = ""
		newStruct.IntField = 0
		newStruct.BoolField = false

		patch, err := Diff(oldStruct, newStruct, WithZeroValueHandling(ZeroIgnore))
		require.NoError(t, err, "Diff should not return error")
		assert.True(t, patch.IsEmpty(), "Patch should be empty when ignoring zero values")
	})
}

func TestDiff_Structs_EdgeCases(t *testing.T) {
	t.Run("NilStructComparison", func(t *testing.T) {
		newStruct := DiverseStruct{
			StringField: "test",
			IntField:    100,
		}

		patch, err := Diff(nil, newStruct, WithNumericOptimization(false))
		require.NoError(t, err, "Should handle nil to struct comparison")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		t.Logf("Nil to struct patch: %s", patch)
		t.Logf("Operations: %+v", patch.Operations())

		operations := patch.Operations()
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")

		setMap := setOps.(map[string]interface{})
		assert.Equal(t, "test", setMap["string_field"], "Should set string field")
		assert.Equal(t, 100, setMap["int_field"], "Should set int field")
	})

	t.Run("StructToNil", func(t *testing.T) {
		oldStruct := DiverseStruct{
			StringField: "test",
			IntField:    100,
			BoolField:   true,
		}

		patch, err := Diff(oldStruct, nil)
		require.NoError(t, err, "Should handle struct to nil comparison")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		t.Logf("Struct to nil patch: %s", patch)
		t.Logf("Operations: %+v", patch.Operations())

		operations := patch.Operations()
		unsetOps, ok := operations["$unset"]
		require.True(t, ok, "Should have $unset operation")

		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "string_field", "Should unset string field")
		assert.Contains(t, unsetMap, "int_field", "Should unset int field")
		assert.Contains(t, unsetMap, "bool_field", "Should unset bool field")
	})

	t.Run("EmptyStruct", func(t *testing.T) {
		oldStruct := DiverseStruct{}
		newStruct := DiverseStruct{}

		patch, err := Diff(oldStruct, newStruct)
		require.NoError(t, err, "Should handle empty struct comparison")
		
		t.Logf("Empty struct patch: %s", patch)
		t.Logf("IsEmpty: %v", patch.IsEmpty())
		t.Logf("Operations: %+v", patch.Operations())
		
		assert.True(t, patch.IsEmpty(), "Patch should be empty for identical empty structs")
	})

	t.Run("OnlyPointerChanges", func(t *testing.T) {
		str1 := "original"
		str2 := "changed"

		oldStruct := DiverseStruct{
			StringPtr: &str1,
			IntPtr:    nil,
		}

		newStruct := DiverseStruct{
			StringPtr: &str2,
			IntPtr:    nil,
		}

		patch, err := Diff(oldStruct, newStruct)
		require.NoError(t, err, "Should handle pointer-only changes")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		t.Logf("Pointer changes patch: %s", patch)
		t.Logf("Operations: %+v", patch.Operations())

		operations := patch.Operations()
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")

		setMap := setOps.(map[string]interface{})
		assert.Equal(t, "changed", setMap["string_ptr"], "StringPtr should be changed")
	})
}

// Test structures for deep nesting scenarios (4 levels deep)
type Level4Node struct {
	Value    string            `bson:"value"`
	Data     map[string]string `bson:"data"`
	Counter  int               `bson:"counter"`
	Active   bool              `bson:"active"`
}

type Level3Node struct {
	Name     string      `bson:"name"`
	Level4   Level4Node  `bson:"level4"`
	Level4Ptr *Level4Node `bson:"level4_ptr,omitempty"`
	Items    []string    `bson:"items"`
}

type Level2Node struct {
	ID       int         `bson:"id"`
	Level3   Level3Node  `bson:"level3"`
	Level3Ptr *Level3Node `bson:"level3_ptr,omitempty"`
	Tags     []Level4Node `bson:"tags"`
}

type Level1Node struct {
	Type     string      `bson:"type"`
	Level2   Level2Node  `bson:"level2"`
	Level2Ptr *Level2Node `bson:"level2_ptr,omitempty"`
	Metadata map[string]interface{} `bson:"metadata"`
}

type DeepNestedStruct struct {
	RootField string      `bson:"root_field"`
	Level1    Level1Node  `bson:"level1"`
	Level1Ptr *Level1Node `bson:"level1_ptr,omitempty"`
	Timestamp time.Time   `bson:"timestamp"`
}

func TestDiff_Structs_DeepNesting4Levels(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	
	// Create a deeply nested structure (4 levels)
	oldStruct := DeepNestedStruct{
		RootField: "root_original",
		Timestamp: now,
		Level1: Level1Node{
			Type: "level1_original",
			Metadata: map[string]interface{}{
				"version": "1.0",
				"author":  "original",
			},
			Level2: Level2Node{
				ID: 100,
				Tags: []Level4Node{
					{Value: "tag1", Data: map[string]string{"key": "value1"}, Counter: 1, Active: true},
					{Value: "tag2", Data: map[string]string{"key": "value2"}, Counter: 2, Active: false},
				},
				Level3: Level3Node{
					Name:  "level3_original",
					Items: []string{"item1", "item2"},
					Level4: Level4Node{
						Value:   "level4_original",
						Data:    map[string]string{"deep": "original", "nested": "value"},
						Counter: 42,
						Active:  true,
					},
					Level4Ptr: &Level4Node{
						Value:   "level4_ptr_original", 
						Data:    map[string]string{"pointer": "original"},
						Counter: 99,
						Active:  false,
					},
				},
			},
		},
	}

	// Create completely independent new structure to avoid pointer sharing
	newStruct := DeepNestedStruct{
		RootField: "root_changed", // Changed
		Timestamp: later,          // Changed
		Level1: Level1Node{
			Type: "level1_changed", // Changed
			Metadata: map[string]interface{}{
				"version": "2.0",        // Changed
				"author":  "original",   // Same
				"status":  "updated",    // Added
			},
			Level2: Level2Node{
				ID: 200, // Changed
				Tags: []Level4Node{
					{Value: "tag1", Data: map[string]string{"key": "value1"}, Counter: 1, Active: true}, // Same
					{Value: "tag2_modified", Data: map[string]string{"key": "value2_new"}, Counter: 3, Active: true}, // Modified
					{Value: "tag3", Data: map[string]string{"key": "value3"}, Counter: 5, Active: true}, // Added
				},
				Level3: Level3Node{
					Name:  "level3_changed", // Changed
					Items: []string{"item1", "item2_modified", "item3"}, // Modified and added
					Level4: Level4Node{
						Value:   "level4_changed", // Changed
						Data:    map[string]string{
							"deep":    "changed",     // Changed
							"nested":  "value",       // Same
							"new":     "field",       // Added
						},
						Counter: 84,    // Changed (42 * 2)
						Active:  false, // Changed
					},
					Level4Ptr: &Level4Node{
						Value:   "level4_ptr_changed", // Changed
						Data:    map[string]string{
							"pointer": "changed", // Changed
							"extra":   "data",    // Added
						},
						Counter: 150,  // Changed
						Active:  true, // Changed
					},
				},
			},
		},
	}

	t.Logf("Testing 4-level deep nested structure changes")
	
	patch, err := Diff(oldStruct, newStruct)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	t.Logf("Deep nesting (4 levels) patch: %s", patch)
	
	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap := setOps.(map[string]interface{})

	// Root level changes
	assert.Equal(t, "root_changed", setMap["root_field"], "Root field should be changed")
	assert.Equal(t, later, setMap["timestamp"], "Timestamp should be changed")

	// Level 1 changes
	assert.Equal(t, "level1_changed", setMap["level1.type"], "Level1 type should be changed")
	assert.Equal(t, "2.0", setMap["level1.metadata.version"], "Level1 metadata version should be changed")
	assert.Equal(t, "updated", setMap["level1.metadata.status"], "Level1 metadata status should be added")

	// Level 2 changes
	assert.Equal(t, 200, setMap["level1.level2.id"], "Level2 ID should be changed")
	assert.Contains(t, setMap, "level1.level2.tags", "Level2 tags array should be updated")

	// Level 3 changes
	assert.Equal(t, "level3_changed", setMap["level1.level2.level3.name"], "Level3 name should be changed")
	assert.Contains(t, setMap, "level1.level2.level3.items", "Level3 items array should be updated")

	// Level 4 changes (deepest level)
	assert.Equal(t, "level4_changed", setMap["level1.level2.level3.level4.value"], "Level4 value should be changed")
	assert.Equal(t, "changed", setMap["level1.level2.level3.level4.data.deep"], "Level4 deep data should be changed")
	assert.Equal(t, "field", setMap["level1.level2.level3.level4.data.new"], "Level4 new data should be added")
	assert.Equal(t, 84, setMap["level1.level2.level3.level4.counter"], "Level4 counter should be changed")
	assert.Equal(t, false, setMap["level1.level2.level3.level4.active"], "Level4 active should be changed")

	// Level 4 pointer changes
	assert.Equal(t, "level4_ptr_changed", setMap["level1.level2.level3.level4_ptr.value"], "Level4 ptr value should be changed")
	assert.Equal(t, "changed", setMap["level1.level2.level3.level4_ptr.data.pointer"], "Level4 ptr data should be changed")
	assert.Equal(t, "data", setMap["level1.level2.level3.level4_ptr.data.extra"], "Level4 ptr extra data should be added")
	assert.Equal(t, 150, setMap["level1.level2.level3.level4_ptr.counter"], "Level4 ptr counter should be changed")
	assert.Equal(t, true, setMap["level1.level2.level3.level4_ptr.active"], "Level4 ptr active should be changed")

	// Verify the depth of field paths
	deepestFieldPath := "level1.level2.level3.level4_ptr.data.extra"
	assert.Contains(t, setMap, deepestFieldPath, "Should handle 6-level deep field path")
	
	// Count field path depths for analysis
	depthCounts := map[int]int{}
	for fieldPath := range setMap {
		depth := len(strings.Split(fieldPath, "."))
		depthCounts[depth]++
	}
	
	t.Logf("Field path depth analysis:")
	for depth := 1; depth <= 6; depth++ {
		if count := depthCounts[depth]; count > 0 {
			t.Logf("  Depth %d: %d fields", depth, count)
		}
	}

	// Verify we have changes at multiple depth levels
	assert.Greater(t, depthCounts[2], 0, "Should have level 2 depth changes")
	assert.Greater(t, depthCounts[3], 0, "Should have level 3 depth changes") 
	assert.Greater(t, depthCounts[4], 0, "Should have level 4 depth changes")
	assert.Greater(t, depthCounts[5], 0, "Should have level 5 depth changes")
	assert.Greater(t, depthCounts[6], 0, "Should have level 6 depth changes")

	// Performance check - ensure reasonable patch size
	patchString := patch.String()
	t.Logf("Patch size: %d characters", len(patchString))
	t.Logf("Total changes detected: %d", len(setMap))
	
	// Should handle deep nesting efficiently without exponential growth
	assert.Less(t, len(patchString), 10000, "Patch should remain reasonably sized even with deep nesting")
}