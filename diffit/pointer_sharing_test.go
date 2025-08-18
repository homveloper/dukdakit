package diffit

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestData struct {
	Field  string      `bson:"field"`
	PtrField *string   `bson:"ptr_field"`
	Nested  *NestedData `bson:"nested"`
}

type NestedData struct {
	Value string `bson:"value"`
}

func TestPointerSharingDetection_Success(t *testing.T) {
	// Test case where pointer sharing should NOT be detected (independent structures)
	str1 := "original"
	str2 := "changed"
	
	oldData := TestData{
		Field:    "test",
		PtrField: &str1,
		Nested:   &NestedData{Value: "nested_original"},
	}
	
	newData := TestData{
		Field:    "test",
		PtrField: &str2, // Different pointer
		Nested:   &NestedData{Value: "nested_changed"}, // Different pointer
	}

	// Should work fine with pointer sharing detection enabled
	patch, err := Diff(oldData, newData, WithDetectPointerSharing(true))
	require.NoError(t, err, "Should not detect pointer sharing in independent structures")
	assert.False(t, patch.IsEmpty(), "Should detect changes")

	t.Logf("Independent structures patch: %s", patch)
}

func TestPointerSharingDetection_Error(t *testing.T) {
	// Test case where pointer sharing should be detected
	sharedStr := "shared"
	sharedNested := &NestedData{Value: "shared_nested"}
	
	oldData := TestData{
		Field:    "test1",
		PtrField: &sharedStr,   // Same pointer
		Nested:   sharedNested, // Same pointer
	}
	
	newData := TestData{
		Field:    "test2",
		PtrField: &sharedStr,   // Same pointer - SHARING!
		Nested:   sharedNested, // Same pointer - SHARING!
	}

	// Should detect pointer sharing and return error
	patch, err := Diff(oldData, newData, WithDetectPointerSharing(true))
	require.Error(t, err, "Should detect pointer sharing")
	assert.Nil(t, patch, "Should not return patch when pointer sharing detected")

	// Check if it's the right type of error
	var sharingErr *PointerSharingError
	assert.ErrorAs(t, err, &sharingErr, "Should be PointerSharingError")
	
	t.Logf("Detected pointer sharing: %s", err.Error())
	assert.Contains(t, err.Error(), "pointer sharing detected", "Error should mention pointer sharing")
}

func TestPointerSharingDetection_Disabled(t *testing.T) {
	// Test case where pointer sharing exists but detection is disabled
	sharedStr := "shared"
	
	oldData := TestData{
		Field:    "test1",
		PtrField: &sharedStr, // Same pointer
	}
	
	newData := TestData{
		Field:    "test2", 
		PtrField: &sharedStr, // Same pointer - but detection disabled
	}

	// Should work fine when detection is disabled (default behavior)
	patch, err := Diff(oldData, newData) // Detection disabled by default
	require.NoError(t, err, "Should not check for pointer sharing when disabled")
	
	// However, the diff result might be incorrect due to pointer sharing
	// This test just verifies that the feature can be disabled
	t.Logf("Pointer sharing ignored (disabled): %s", patch)
}

func TestPointerSharingDetection_ComplexNesting(t *testing.T) {
	// Test case with deep nesting and shared pointers
	type Level2 struct {
		Data *string `bson:"data"`
	}
	type Level1 struct {
		Level2 *Level2 `bson:"level2"`
	}
	type Root struct {
		Level1 *Level1 `bson:"level1"`
	}

	sharedData := "deep_shared"
	sharedLevel2 := &Level2{Data: &sharedData}
	sharedLevel1 := &Level1{Level2: sharedLevel2}
	
	oldRoot := Root{Level1: sharedLevel1}
	newRoot := Root{Level1: sharedLevel1} // Same nested structure - SHARING!

	// Should detect sharing in nested structures
	_, err := Diff(oldRoot, newRoot, WithDetectPointerSharing(true))
	require.Error(t, err, "Should detect pointer sharing in nested structures")
	
	var sharingErr *PointerSharingError
	if assert.ErrorAs(t, err, &sharingErr, "Should be PointerSharingError") {
		t.Logf("Deep sharing detected: %s", sharingErr.Error())
		t.Logf("Field path: %s", sharingErr.FieldPath)
		t.Logf("Address: %p", unsafe.Pointer(sharingErr.Address))
	}
}

func TestPointerSharingDetection_Arrays(t *testing.T) {
	// Test case with arrays containing shared pointers
	type Container struct {
		Items []*string `bson:"items"`
	}

	sharedStr1 := "shared1"
	sharedStr2 := "shared2"
	
	oldData := Container{
		Items: []*string{&sharedStr1, &sharedStr2},
	}
	
	newData := Container{
		Items: []*string{&sharedStr1, &sharedStr2}, // Same pointers in array
	}

	// Should detect sharing in array elements
	_, err := Diff(oldData, newData, WithDetectPointerSharing(true))
	require.Error(t, err, "Should detect pointer sharing in arrays")
	
	t.Logf("Array sharing detected: %s", err.Error())
	assert.Contains(t, strings.ToLower(err.Error()), "pointer sharing", "Error should mention pointer sharing")
}