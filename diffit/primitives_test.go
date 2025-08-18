package diffit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures for primitive type testing
type SimpleUser struct {
	Name  string `bson:"name"`
	Age   int    `bson:"age"`
	Email string `bson:"email"`
}

type NumericTypes struct {
	Int     int     `bson:"int"`
	Int32   int32   `bson:"int32"`
	Int64   int64   `bson:"int64"`
	Float32 float32 `bson:"float32"`
	Float64 float64 `bson:"float64"`
	Uint    uint    `bson:"uint"`
}

type StringFields struct {
	Name        string `bson:"name"`
	Description string `bson:"description"`
	Category    string `bson:"category"`
	EmptyField  string `bson:"empty_field"`
}

type BooleanFields struct {
	IsActive    bool `bson:"is_active"`
	IsVerified  bool `bson:"is_verified"`
	IsPublic    bool `bson:"is_public"`
	HasAccess   bool `bson:"has_access"`
}

func TestDiff_PrimitiveTypes_BasicFieldChanges(t *testing.T) {
	oldUser := SimpleUser{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	newUser := SimpleUser{
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

func TestDiff_PrimitiveTypes_NumericFields(t *testing.T) {
	oldData := NumericTypes{
		Int:     100,
		Int32:   200,
		Int64:   1000,
		Float32: 10.5,
		Float64: 20.75,
		Uint:    50,
	}

	newData := NumericTypes{
		Int:     150, // Changed to 150
		Int32:   250, // Changed to 250
		Int64:   1100, // Changed to 1100
		Float32: 15.5,  // Changed to 15.5
		Float64: 25.75, // Changed to 25.75
		Uint:    75,    // Changed to 75
	}

	// Numeric optimization is disabled - all changes use $set
	patch, err := Diff(oldData, newData, WithNumericOptimization(true))
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()

	// Check if $set operation exists for numeric fields
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation for numeric fields")

	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")
	
	assert.Equal(t, 150, setMap["int"], "Int should be set to 150")
	assert.Equal(t, int32(250), setMap["int32"], "Int32 should be set to 250")
	assert.Equal(t, int64(1100), setMap["int64"], "Int64 should be set to 1100")
	assert.Equal(t, float32(15.5), setMap["float32"], "Float32 should be set to 15.5")
	assert.Equal(t, 25.75, setMap["float64"], "Float64 should be set to 25.75")
	assert.Equal(t, uint(75), setMap["uint"], "Uint should be set to 75")

	t.Logf("Numeric field patch: %s", patch)
}

func TestDiff_PrimitiveTypes_StringChanges(t *testing.T) {
	oldData := StringFields{
		Name:        "Original",
		Description: "Old description",
		Category:    "category1",
		EmptyField:  "",
	}

	newData := StringFields{
		Name:        "Updated",        // Changed
		Description: "New description", // Changed
		Category:    "category1",      // Same
		EmptyField:  "now has value",  // Changed from empty
	}

	patch, err := Diff(oldData, newData)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")
	
	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")

	// Check string changes
	assert.Equal(t, "Updated", setMap["name"], "Name should be updated")
	assert.Equal(t, "New description", setMap["description"], "Description should be updated")
	assert.Equal(t, "now has value", setMap["empty_field"], "EmptyField should be set")

	// Category should not be in patch since it didn't change
	assert.NotContains(t, setMap, "category", "Category should not be in patch")

	t.Logf("String changes patch: %s", patch)
}

func TestDiff_PrimitiveTypes_BooleanChanges(t *testing.T) {
	oldData := BooleanFields{
		IsActive:   true,
		IsVerified: false,
		IsPublic:   true,
		HasAccess:  false,
	}

	newData := BooleanFields{
		IsActive:   false, // Changed
		IsVerified: true,  // Changed
		IsPublic:   true,  // Same
		HasAccess:  false, // Same
	}

	patch, err := Diff(oldData, newData)
	require.NoError(t, err, "Diff should not return error")
	assert.False(t, patch.IsEmpty(), "Patch should not be empty")

	operations := patch.Operations()
	setOps, ok := operations["$set"]
	require.True(t, ok, "Should have $set operation")
	
	setMap, ok := setOps.(map[string]interface{})
	require.True(t, ok, "$set should be a map")

	// Check boolean changes
	assert.Equal(t, false, setMap["is_active"], "IsActive should be changed to false")
	assert.Equal(t, true, setMap["is_verified"], "IsVerified should be changed to true")

	// Unchanged fields should not be in patch
	assert.NotContains(t, setMap, "is_public", "IsPublic should not be in patch")
	assert.NotContains(t, setMap, "has_access", "HasAccess should not be in patch")

	t.Logf("Boolean changes patch: %s", patch)
}

func TestDiff_PrimitiveTypes_NoChanges(t *testing.T) {
	user := SimpleUser{
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

func TestDiff_PrimitiveTypes_ZeroValueHandling(t *testing.T) {
	t.Run("ZeroAsSet", func(t *testing.T) {
		oldData := SimpleUser{Name: "John", Age: 25, Email: "john@example.com"}
		newData := SimpleUser{Name: "", Age: 0, Email: ""} // All zero values

		patch, err := Diff(oldData, newData, WithZeroValueHandling(ZeroAsSet))
		require.NoError(t, err, "Diff should not return error")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")
		
		setMap := setOps.(map[string]interface{})
		assert.Equal(t, "", setMap["name"], "Name should be set to empty string")
		assert.Equal(t, 0, setMap["age"], "Age should be set to 0")
		assert.Equal(t, "", setMap["email"], "Email should be set to empty string")
	})

	t.Run("ZeroAsUnset", func(t *testing.T) {
		oldData := SimpleUser{Name: "John", Age: 25, Email: "john@example.com"}
		newData := SimpleUser{Name: "", Age: 0, Email: ""} // All zero values

		patch, err := Diff(oldData, newData, WithZeroValueHandling(ZeroAsUnset))
		require.NoError(t, err, "Diff should not return error")
		assert.False(t, patch.IsEmpty(), "Patch should not be empty")

		operations := patch.Operations()
		unsetOps, ok := operations["$unset"]
		require.True(t, ok, "Should have $unset operation")
		
		unsetMap := unsetOps.(map[string]interface{})
		assert.Contains(t, unsetMap, "name", "Name should be unset")
		assert.Contains(t, unsetMap, "age", "Age should be unset")
		assert.Contains(t, unsetMap, "email", "Email should be unset")
	})

	t.Run("ZeroIgnore", func(t *testing.T) {
		oldData := SimpleUser{Name: "John", Age: 25, Email: "john@example.com"}
		newData := SimpleUser{Name: "", Age: 0, Email: ""} // All zero values

		patch, err := Diff(oldData, newData, WithZeroValueHandling(ZeroIgnore))
		require.NoError(t, err, "Diff should not return error")
		assert.True(t, patch.IsEmpty(), "Patch should be empty when ignoring zero values")
	})
}

func TestDiff_PrimitiveTypes_JSONMarshaling(t *testing.T) {
	oldUser := SimpleUser{Name: "John", Age: 25}
	newUser := SimpleUser{Name: "Jane", Age: 26}

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

func TestDiff_PrimitiveTypes_IgnoreFields(t *testing.T) {
	oldUser := SimpleUser{
		Name:  "John",
		Age:   25,
		Email: "john@example.com",
	}

	newUser := SimpleUser{
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

func TestDiff_PrimitiveTypes_EdgeCases(t *testing.T) {
	t.Run("TypeMismatch", func(t *testing.T) {
		user := SimpleUser{Name: "John"}
		numeric := NumericTypes{Int: 123}

		_, err := Diff(user, numeric)
		assert.Error(t, err, "Should return error for type mismatch")
		assert.Contains(t, err.Error(), "incompatible types", "Error should mention incompatible types")
	})

	t.Run("BothNil", func(t *testing.T) {
		_, err := Diff(nil, nil)
		assert.Error(t, err, "Should return error for both nil values")
		assert.Contains(t, err.Error(), "both values are nil", "Error should mention nil values")
	})

	t.Run("OneNilStruct", func(t *testing.T) {
		newUser := SimpleUser{Name: "John", Age: 25}
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

	t.Run("StructToNil", func(t *testing.T) {
		oldUser := SimpleUser{Name: "John", Age: 25}
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