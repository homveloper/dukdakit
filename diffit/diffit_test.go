package diffit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures
type User struct {
	Name  string `bson:"name"`
	Age   int    `bson:"age"`
	Email string `bson:"email"`
}

type Player struct {
	ID    int64  `bson:"id"`
	Name  string `bson:"name"`
	Score int    `bson:"score"`
	Level int    `bson:"level"`
}

func TestDiff_BasicFieldChanges(t *testing.T) {
	oldUser := User{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	newUser := User{
		Name:  "John",
		Age:   26, // Changed
		Email: "john.doe@example.com", // Changed
	}

	// Disable numeric optimization for this test to get $set operations
	patch, err := Diff(oldUser, newUser, WithNumericOptimization(false))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	require.NotNil(t, operations, "Operations should be present")

	// Check if $set operation exists
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")
	
	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")
	
	// Check age change
	assert.Equal(t, 26, setMap["age"], "Age should be updated to 26")
	
	// Check email change
	assert.Equal(t, "john.doe@example.com", setMap["email"], "Email should be updated")

	t.Logf("Patch: %s", patch)
}

func TestDiff_NumericFields(t *testing.T) {
	oldPlayer := Player{
		ID:    1,
		Name:  "Player1",
		Score: 1000,
		Level: 5,
	}

	newPlayer := Player{
		ID:    1,
		Name:  "Player1",
		Score: 1100, // Changed to 1100
		Level: 6,    // Changed to 6
	}

	// Numeric optimization is disabled - all changes use $set
	patch, err := Diff(oldPlayer, newPlayer, WithNumericOptimization(true))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()

	// Check if $set operation exists for numeric fields
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")

	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")
	
	assert.Equal(t, 1100, setMap["score"], "Score should be set to 1100")
	assert.Equal(t, 6, setMap["level"], "Level should be set to 6")

	t.Logf("Numeric field patch: %s", patch)
}

func TestDiff_NoChanges(t *testing.T) {
	user := User{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	patch, err := Diff(user, user)
	require.NoError(t, err, "Diff should not return error")
	assert.True(t, patch.IsEmpty(), "Patch should be empty for identical values")

	operations := patch.Operations()
	assert.Empty(t, operations, "Operations should be empty")
	
	assert.Equal(t, 0, patch.metadata.TotalChanges, "Should have no changes")

	t.Logf("Empty patch: %s", patch)
}

func TestDiff_JSONMarshaling(t *testing.T) {
	oldUser := User{Name: "John", Age: 25}
	newUser := User{Name: "Jane", Age: 26}

	patch, err := Diff(oldUser, newUser)
	require.NoError(t, err, "Diff should not return error")

	// Test JSON marshaling
	jsonData, err := json.Marshal(patch)
	require.NoError(t, err, "JSON marshaling should not fail")
	assert.NotEmpty(t, jsonData, "JSON data should not be empty")

	t.Logf("JSON patch: %s", string(jsonData))

	// Verify JSON structure
	var info PatchInfo
	err = json.Unmarshal(jsonData, &info)
	require.NoError(t, err, "JSON unmarshaling should not fail")

	assert.NotNil(t, info.Operations, "Operations should be present in JSON")
	assert.Greater(t, info.Metadata.TotalChanges, 0, "Metadata should show changes")

	// Verify specific fields in metadata
	assert.Contains(t, info.Metadata.FieldsChanged, "name", "Name field should be tracked")
	assert.Contains(t, info.Metadata.FieldsChanged, "age", "Age field should be tracked")
}

func TestDiff_IgnoreFields(t *testing.T) {
	oldUser := User{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	newUser := User{
		Name:  "Jane", // Changed
		Age:   26,     // Changed (but should be ignored)
		Email: "jane@example.com", // Changed
	}

	patch, err := Diff(oldUser, newUser, WithIgnoreFields("age"))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	require.NotNil(t, operations, "Operations should be present")

	// Should have $set operations for name and email
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")
	
	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")

	// Age should not be in the patch
	assert.NotContains(t, setMap, "age", "Age field should be ignored")

	// Name and email should be present
	assert.Equal(t, "Jane", setMap["name"], "Name change should be present")
	assert.Equal(t, "jane@example.com", setMap["email"], "Email change should be present")

	// Verify metadata doesn't include ignored field
	assert.NotContains(t, patch.metadata.FieldsChanged, "age", "Ignored field should not be in metadata")

	t.Logf("Patch with ignored fields: %s", patch)
}

func TestDiff_NilValues(t *testing.T) {
	t.Run("BothNil", func(t *testing.T) {
		_, err := Diff(nil, nil)
		assert.Error(t, err, "Should return error for both nil values")
		assert.Contains(t, err.Error(), "both values are nil", "Error should mention nil values")
	})

	t.Run("OldNil", func(t *testing.T) {
		newUser := User{Name: "John", Age: 25}
		patch, err := Diff(nil, newUser)
		require.NoError(t, err, "Should handle nil to struct comparison")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")
		
		setMap := setOps.(map[string]interface{})
		assert.Equal(t, "John", setMap["name"], "Should set name field")
		assert.Equal(t, 25, setMap["age"], "Should set age field")
	})

	t.Run("NewNil", func(t *testing.T) {
		oldUser := User{Name: "John", Age: 25}
		patch, err := Diff(oldUser, nil)
		require.NoError(t, err, "Should handle struct to nil comparison")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		unsetOps, ok := operations["$unset"]
		require.True(t, ok, "Should have $unset operation")
		
		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "name", "Should unset name field")
		assert.Contains(t, unsetMap, "age", "Should unset age field")
	})
}

func TestDiff_TypeMismatch(t *testing.T) {
	user := User{Name: "John"}
	player := Player{Name: "John"}

	_, err := Diff(user, player)
	assert.Error(t, err, "Should return error for type mismatch")
	assert.Contains(t, err.Error(), "incompatible types", "Error should mention incompatible types")
}

func TestBsonPatch_StringFormatting(t *testing.T) {
	oldUser := User{Name: "John", Age: 25}
	newUser := User{Name: "Jane", Age: 26}

	patch, err := Diff(oldUser, newUser)
	require.NoError(t, err)

	// Test String() method
	str := patch.String()
	assert.Contains(t, str, "BsonPatch:", "String should contain BsonPatch prefix")
	assert.Contains(t, str, "Operations:", "String should contain operations")
	assert.Contains(t, str, "Changes:", "String should contain change count")

	// Test GoString() method  
	goStr := patch.GoString()
	assert.Contains(t, goStr, "&diffit.BsonPatch{", "GoString should contain type info")
	assert.Contains(t, goStr, "operations:", "GoString should contain field names")
}

func TestBsonPatch_Info(t *testing.T) {
	oldUser := User{Name: "John", Age: 25}
	newUser := User{Name: "Jane", Age: 30}

	patch, err := Diff(oldUser, newUser)
	require.NoError(t, err)

	info := patch.Info()
	assert.NotNil(t, info.Operations, "Info should contain operations")
	assert.Equal(t, patch.metadata, info.Metadata, "Info should contain metadata")
	assert.Equal(t, patch.arrayFilters, info.ArrayFilters, "Info should contain array filters")
}