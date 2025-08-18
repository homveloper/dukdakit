package diffit

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Diff compares two values and returns a BsonPatch for MongoDB updates
func Diff(oldValue, newValue interface{}, options ...Option) (*BsonPatch, error) {
	// Apply configuration options
	config := newDiffConfig()

	for _, opt := range options {
		opt(config)
	}

	// Initialize pointer tracking if enabled
	if config.DetectPointerSharing {
		config.pointerTracker = NewPointerTracker()
	}

	// Validate input types
	if err := validateInputs(oldValue, newValue); err != nil {
		return nil, err
	}

	// Track root level pointers if tracking is enabled
	if config.DetectPointerSharing {
		trackRootPointers(config.pointerTracker, oldValue, newValue)
	}

	// Create new patch
	patch := newBsonPatch()

	// Perform the comparison
	if err := compareValues(patch, config, "", oldValue, newValue); err != nil {
		return nil, err
	}

	// Check for pointer sharing after comparison if enabled
	if config.DetectPointerSharing {
		if sharingErr := config.pointerTracker.CheckForSharing(); sharingErr != nil {
			return nil, sharingErr
		}
	}

	return patch, nil
}

// validateInputs ensures the input values are compatible for comparison
func validateInputs(oldValue, newValue interface{}) error {
	if oldValue == nil && newValue == nil {
		return errors.New("both values are nil")
	}

	// Allow nil to non-nil comparison
	if oldValue == nil || newValue == nil {
		return nil
	}

	oldType := reflect.TypeOf(oldValue)
	newType := reflect.TypeOf(newValue)

	// Types must be the same for comparison
	if oldType != newType {
		return errors.New("incompatible types for comparison")
	}

	return nil
}

// compareValues compares two values and generates appropriate patch operations
func compareValues(patch *BsonPatch, config *DiffConfig, fieldPath string, oldValue, newValue interface{}) error {
	// Handle nil values
	if oldValue == nil && newValue == nil {
		return nil // No change
	}

	if oldValue == nil && newValue != nil {
		// Field was added - handle based on type
		newVal := reflect.ValueOf(newValue)
		switch newVal.Kind() {
		case reflect.Struct:
			// For structs, compare each field individually
			zeroVal := reflect.Zero(newVal.Type())
			return compareStructs(patch, config, fieldPath, zeroVal, newVal)
		default:
			return handleFieldSet(patch, config, fieldPath, newValue)
		}
	}

	if oldValue != nil && newValue == nil {
		// Field was removed - handle based on type
		oldVal := reflect.ValueOf(oldValue)
		switch oldVal.Kind() {
		case reflect.Struct:
			// For structs, unset each non-zero field individually
			return unsetStructFields(patch, config, fieldPath, oldVal)
		default:
			return handleFieldUnset(patch, config, fieldPath)
		}
	}

	// Both values are non-nil, compare based on type
	oldVal := reflect.ValueOf(oldValue)
	newVal := reflect.ValueOf(newValue)

	if oldVal.Type() != newVal.Type() {
		return errors.New("type mismatch during comparison")
	}

	// Special handling for time.Time which has Kind() == Struct but should be treated as scalar
	if _, isTime := oldValue.(time.Time); isTime {
		return compareScalars(patch, config, fieldPath, oldValue, newValue)
	}

	switch oldVal.Kind() {
	case reflect.Struct:
		return compareStructs(patch, config, fieldPath, oldVal, newVal)
	case reflect.Slice, reflect.Array:
		return compareArrays(patch, config, fieldPath, oldVal, newVal)
	case reflect.Map:
		return compareMaps(patch, config, fieldPath, oldVal, newVal)
	case reflect.Ptr:
		return comparePointers(patch, config, fieldPath, oldVal, newVal)
	default:
		return compareScalars(patch, config, fieldPath, oldValue, newValue)
	}
}

// compareStructs compares two struct values field by field
func compareStructs(patch *BsonPatch, config *DiffConfig, basePath string, oldVal, newVal reflect.Value) error {
	structType := oldVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get BSON field name from tag
		bsonTag := field.Tag.Get("bson")
		bsonFieldName := getBsonFieldName(fieldName, bsonTag)

		// Check if field should be ignored (check both struct field name and bson field name)
		if isIgnoredField(fieldName, config.IgnoreFields) || isIgnoredField(bsonFieldName, config.IgnoreFields) {
			continue
		}
		if bsonFieldName == "-" {
			continue // Skip fields marked with bson:"-"
		}

		// Build field path
		var fieldPath string
		if basePath == "" {
			fieldPath = bsonFieldName
		} else {
			fieldPath = basePath + "." + bsonFieldName
		}

		// Check if the full field path should be ignored
		if isIgnoredField(fieldPath, config.IgnoreFields) {
			continue
		}

		// Get field values
		oldFieldVal := oldVal.Field(i)
		newFieldVal := newVal.Field(i)

		// Compare field values
		if err := compareValues(patch, config, fieldPath, oldFieldVal.Interface(), newFieldVal.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// compareArrays compares two array/slice values
func compareArrays(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	switch config.ArrayStrategy {
	case ArrayReplace:
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	case ArraySmart:
		return handleArraySmart(patch, config, fieldPath, oldVal, newVal)
	case ArrayAppend:
		return handleArrayAppend(patch, config, fieldPath, oldVal, newVal)
	case ArrayMerge:
		return handleArrayMerge(patch, config, fieldPath, oldVal, newVal)
	default:
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}
}

// compareMaps compares two map values
func compareMaps(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Get all keys from both maps
	allKeys := make(map[interface{}]bool)

	for _, key := range oldVal.MapKeys() {
		allKeys[key.Interface()] = true
	}
	for _, key := range newVal.MapKeys() {
		allKeys[key.Interface()] = true
	}

	// Compare each key
	for keyInterface := range allKeys {
		key := reflect.ValueOf(keyInterface)
		keyStr := key.String() // Assume string keys for now

		var subFieldPath string
		if fieldPath == "" {
			subFieldPath = keyStr
		} else {
			subFieldPath = fieldPath + "." + keyStr
		}

		oldMapVal := oldVal.MapIndex(key)
		newMapVal := newVal.MapIndex(key)

		var oldInterface, newInterface interface{}
		if oldMapVal.IsValid() {
			oldInterface = oldMapVal.Interface()
		}
		if newMapVal.IsValid() {
			newInterface = newMapVal.Interface()
		}

		if err := compareValues(patch, config, subFieldPath, oldInterface, newInterface); err != nil {
			return err
		}
	}

	return nil
}

// isEqual compares two values for equality, handling special types like time.Time
func isEqual(oldValue, newValue interface{}) bool {
	// Handle time.Time specially using Equal method
	if oldTime, ok := oldValue.(time.Time); ok {
		if newTime, ok := newValue.(time.Time); ok {
			return oldTime.Equal(newTime)
		}
		return false
	}

	// Handle *time.Time pointers specially
	if oldTimePtr, ok := oldValue.(*time.Time); ok {
		if newTimePtr, ok := newValue.(*time.Time); ok {
			if oldTimePtr == nil && newTimePtr == nil {
				return true
			}
			if oldTimePtr == nil || newTimePtr == nil {
				return false
			}
			return oldTimePtr.Equal(*newTimePtr)
		}
		return false
	}

	// For all other types, use reflect.DeepEqual
	return reflect.DeepEqual(oldValue, newValue)
}

// compareScalars compares scalar values (primitives)
func compareScalars(patch *BsonPatch, config *DiffConfig, fieldPath string, oldValue, newValue interface{}) error {
	// Check if values are equal using custom equal function
	if isEqual(oldValue, newValue) {
		return nil // No change
	}

	// Removed numeric optimization - all changes use $set for clarity and consistency

	// Handle zero values based on strategy
	if isZeroValue(newValue) {
		switch config.ZeroValueHandling {
		case ZeroAsUnset:
			return handleFieldUnset(patch, config, fieldPath)
		case ZeroIgnore:
			return nil // Ignore zero values
		case ZeroAsSet:
			fallthrough
		default:
			return handleFieldSet(patch, config, fieldPath, newValue)
		}
	}

	// Regular field set
	return handleFieldSet(patch, config, fieldPath, newValue)
}

// Field operation handlers

// handleFieldSet handles setting a field value
func handleFieldSet(patch *BsonPatch, config *DiffConfig, fieldPath string, value interface{}) error {
	patch.addOperation("$set", fieldPath, value)
	return nil
}

// handleFieldUnset handles unsetting a field
func handleFieldUnset(patch *BsonPatch, config *DiffConfig, fieldPath string) error {
	patch.addOperation("$unset", fieldPath, "")
	return nil
}

// Array operation handlers (simplified implementations for now)

// handleArrayReplace replaces the entire array
func handleArrayReplace(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Check if arrays are equal
	if reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
		return nil // No change needed
	}

	patch.addOperation("$set", fieldPath, newVal.Interface())
	return nil
}

// handleArraySmart performs intelligent array comparison
func handleArraySmart(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Check if arrays are equal first
	if reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
		return nil // No change needed
	}

	oldLen := oldVal.Len()
	newLen := newVal.Len()

	// For primitive types or when lengths are very different, use replace strategy
	if oldLen == 0 || newLen == 0 || abs(oldLen-newLen) > max(oldLen, newLen)/2 {
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// Check array element type
	if oldLen > 0 {
		elemType := oldVal.Index(0).Kind()
		// For primitive types, use replace for simplicity
		if isPrimitiveKind(elemType) {
			return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
		}
	}

	// For struct arrays, try element-wise comparison with array filters
	return handleArraySmartStructs(patch, config, fieldPath, oldVal, newVal)
}

// handleArrayAppend adds new elements to array (append-only strategy)
func handleArrayAppend(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Check if arrays are equal first
	if reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
		return nil // No change needed
	}

	oldLen := oldVal.Len()
	newLen := newVal.Len()

	// If new array is shorter, fall back to replace (can't append-only)
	if newLen < oldLen {
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// Check if old array is a prefix of new array
	for i := 0; i < oldLen; i++ {
		if !reflect.DeepEqual(oldVal.Index(i).Interface(), newVal.Index(i).Interface()) {
			// Elements were modified, can't use append-only - fall back to replace
			return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
		}
	}

	// Old array is a prefix, we can append new elements
	if newLen > oldLen {
		// Use $push to add new elements
		// If multiple elements, use $each to push them all at once
		if newLen-oldLen == 1 {
			// Single element append
			newElem := newVal.Index(oldLen)
			patch.addOperation("$push", fieldPath, newElem.Interface())
		} else {
			// Multiple elements append - use $each
			elementsToAdd := make([]interface{}, newLen-oldLen)
			for i := oldLen; i < newLen; i++ {
				elementsToAdd[i-oldLen] = newVal.Index(i).Interface()
			}
			patch.addOperation("$push", fieldPath, map[string]interface{}{
				"$each": elementsToAdd,
			})
		}
	}

	return nil
}

// handleArrayMerge merges arrays intelligently using content-based matching
func handleArrayMerge(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Check if arrays are equal first
	if reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
		return nil // No change needed
	}

	oldLen := oldVal.Len()
	newLen := newVal.Len()

	// For empty arrays, use simple logic
	if oldLen == 0 && newLen == 0 {
		return nil // Both empty
	}
	if oldLen == 0 {
		// Old empty, new has items - set the entire array
		return handleFieldSet(patch, config, fieldPath, newVal.Interface())
	}
	if newLen == 0 {
		// New empty, old has items - unset the array
		return handleFieldUnset(patch, config, fieldPath)
	}

	// Check array element type - merge strategy works best with structs
	if oldLen > 0 {
		elemType := oldVal.Index(0).Kind()
		// For primitive types, merge doesn't provide much benefit over replace
		if isPrimitiveKind(elemType) {
			return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
		}
	}

	// Try to implement intelligent merging for structs
	return handleArrayMergeStructs(patch, config, fieldPath, oldVal, newVal)
}

// unsetStructFields unsets all non-zero fields in a struct
func unsetStructFields(patch *BsonPatch, config *DiffConfig, basePath string, oldVal reflect.Value) error {
	structType := oldVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get BSON field name from tag
		bsonTag := field.Tag.Get("bson")
		bsonFieldName := getBsonFieldName(fieldName, bsonTag)

		// Check if field should be ignored
		if isIgnoredField(fieldName, config.IgnoreFields) || isIgnoredField(bsonFieldName, config.IgnoreFields) {
			continue
		}
		if bsonFieldName == "-" {
			continue // Skip fields marked with bson:"-"
		}

		// Build field path
		var fieldPath string
		if basePath == "" {
			fieldPath = bsonFieldName
		} else {
			fieldPath = basePath + "." + bsonFieldName
		}

		// Check if the full field path should be ignored
		if isIgnoredField(fieldPath, config.IgnoreFields) {
			continue
		}

		// Get field value
		oldFieldVal := oldVal.Field(i)

		// Check if field has non-zero value
		if !oldFieldVal.IsZero() {
			// For nested structs, recursively unset fields
			if oldFieldVal.Kind() == reflect.Struct {
				if err := unsetStructFields(patch, config, fieldPath, oldFieldVal); err != nil {
					return err
				}
			} else {
				// Unset this field
				patch.addOperation("$unset", fieldPath, "")
			}
		}
	}

	return nil
}

// comparePointers compares two pointer values
func comparePointers(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Handle nil pointers
	if oldVal.IsNil() && newVal.IsNil() {
		return nil // No change
	}

	if oldVal.IsNil() && !newVal.IsNil() {
		// nil to value - handle based on the pointed type
		derefVal := newVal.Elem()
		if derefVal.Kind() == reflect.Struct {
			// For struct pointers, compare field by field
			zeroVal := reflect.Zero(derefVal.Type())
			return compareStructs(patch, config, fieldPath, zeroVal, derefVal)
		} else {
			// For non-struct pointers, set the dereferenced value
			return handleFieldSet(patch, config, fieldPath, derefVal.Interface())
		}
	}

	if !oldVal.IsNil() && newVal.IsNil() {
		// value to nil - handle based on the pointed type
		derefVal := oldVal.Elem()
		if derefVal.Kind() == reflect.Struct {
			// For struct pointers, unset field by field
			return unsetStructFields(patch, config, fieldPath, derefVal)
		} else {
			// For non-struct pointers, unset the field
			return handleFieldUnset(patch, config, fieldPath)
		}
	}

	// Both pointers are non-nil, compare their values
	oldElem := oldVal.Elem()
	newElem := newVal.Elem()

	if oldElem.Kind() == reflect.Struct && newElem.Kind() == reflect.Struct {
		// For struct pointers, compare field by field
		return compareStructs(patch, config, fieldPath, oldElem, newElem)
	} else {
		// For non-struct pointers, compare as scalars
		return compareScalars(patch, config, fieldPath, oldElem.Interface(), newElem.Interface())
	}
}

// Helper functions for array processing

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// isPrimitiveKind checks if a reflect.Kind is a primitive type
func isPrimitiveKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

// handleArraySmartStructs handles smart comparison of struct arrays using array filters
func handleArraySmartStructs(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// For now, implement a basic version that generates array filters for matching elements
	// This is a simplified implementation - a full version would need more sophisticated matching

	oldLen := oldVal.Len()
	newLen := newVal.Len()
	minLen := min(oldLen, newLen)

	arrayFilterId := &ArrayFilterIdentifier{}
	hasElementChanges := false
	hasLengthChanges := newLen != oldLen

	// Check if any elements at existing positions changed
	for i := 0; i < minLen; i++ {
		oldElem := oldVal.Index(i)
		newElem := newVal.Index(i)

		// For struct elements, compare field by field
		if oldElem.Kind() == reflect.Struct && newElem.Kind() == reflect.Struct {
			if !reflect.DeepEqual(oldElem.Interface(), newElem.Interface()) {
				hasElementChanges = true
				break
			}
		} else {
			// For non-struct elements (like time.Time), use isEqual
			if !isEqual(oldElem.Interface(), newElem.Interface()) {
				hasElementChanges = true
				break
			}
		}
	}

	// If we have both element changes and length changes, use replace strategy
	// to avoid MongoDB mixed operation conflicts
	if hasElementChanges && hasLengthChanges {
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// If only elements changed (same length), use array filters
	if hasElementChanges && !hasLengthChanges {
		for i := 0; i < minLen; i++ {
			oldElem := oldVal.Index(i)
			newElem := newVal.Index(i)

			// Check if this specific element changed
			var elemChanged bool
			if oldElem.Kind() == reflect.Struct && newElem.Kind() == reflect.Struct {
				elemChanged = !reflect.DeepEqual(oldElem.Interface(), newElem.Interface())
			} else {
				elemChanged = !isEqual(oldElem.Interface(), newElem.Interface())
			}

			if elemChanged {
				// Generate array filter for this position
				filterKey := arrayFilterId.Next()

				// Create array filter for position-based update
				patch.addArrayFilter(map[string]interface{}{
					filterKey + "._index": i,
				})

				// Add update operation using array filter
				updatePath := fieldPath + ".$[" + filterKey + "]"
				patch.addOperation("$set", updatePath, newElem.Interface())
			}
		}
		return nil
	}

	// If only length changed (elements appended), use push operations
	if !hasElementChanges && hasLengthChanges {
		if newLen > oldLen {
			// Add new elements
			if newLen-oldLen == 1 {
				// Single element append
				newElem := newVal.Index(oldLen)
				patch.addOperation("$push", fieldPath, newElem.Interface())
			} else {
				// Multiple elements append - use $each
				elementsToAdd := make([]interface{}, newLen-oldLen)
				for i := oldLen; i < newLen; i++ {
					elementsToAdd[i-oldLen] = newVal.Index(i).Interface()
				}
				patch.addOperation("$push", fieldPath, map[string]interface{}{
					"$each": elementsToAdd,
				})
			}
			return nil
		} else if newLen < oldLen {
			// For now, fall back to replace when removing elements
			// A full implementation might use $pull or other operations
			return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
		}
	}

	// No changes detected
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleArrayMergeStructs implements intelligent merging for struct arrays
func handleArrayMergeStructs(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	oldLen := oldVal.Len()
	newLen := newVal.Len()

	// Create maps to track which elements have been matched
	oldMatched := make([]bool, oldLen)
	newMatched := make([]bool, newLen)

	arrayFilterId := &ArrayFilterIdentifier{}
	hasChanges := false

	// Phase 1: Try to match elements by content similarity
	// This is a simplified matching - in production, you might want more sophisticated matching
	for i := 0; i < newLen; i++ {
		newElem := newVal.Index(i)
		bestMatch := -1

		// Look for the best match in old array
		for j := 0; j < oldLen; j++ {
			if oldMatched[j] {
				continue // Already matched
			}

			oldElem := oldVal.Index(j)
			if elementsAreSimilar(oldElem, newElem) {
				bestMatch = j
				break // Take first similar match
			}
		}

		if bestMatch >= 0 {
			// Found a match - check if update is needed
			oldElem := oldVal.Index(bestMatch)
			if !reflect.DeepEqual(oldElem.Interface(), newElem.Interface()) {
				// Generate array filter for this match
				filterKey := arrayFilterId.Next()

				// Create array filter based on element position in old array
				patch.addArrayFilter(map[string]interface{}{
					filterKey + "._index": bestMatch,
				})

				// Add update operation using array filter
				updatePath := fieldPath + ".$[" + filterKey + "]"
				patch.addOperation("$set", updatePath, newElem.Interface())
				hasChanges = true
			}

			oldMatched[bestMatch] = true
			newMatched[i] = true
		}
	}

	// Phase 2: Count unmatched elements (we'll handle them in Phase 4)
	// This phase is now just for counting, actual handling moved to Phase 4

	// Phase 3: Check for any complex scenarios that require fallback
	unMatchedOldCount := 0
	unMatchedNewCount := 0

	for j := 0; j < oldLen; j++ {
		if !oldMatched[j] {
			unMatchedOldCount++
		}
	}

	for i := 0; i < newLen; i++ {
		if !newMatched[i] {
			unMatchedNewCount++
		}
	}

	// If we have both updates, additions, and removals, it's too complex for merge strategy
	// Also, MongoDB doesn't allow mixed operations on the same field path
	if hasChanges && unMatchedNewCount > 0 {
		// We have both ArrayFilter updates and new additions - fallback to replace
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	if unMatchedOldCount > 0 {
		// We have removals, which are complex to handle - fallback to replace
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// Phase 4: Handle pure additions (no updates, no removals)
	if !hasChanges && unMatchedNewCount > 0 {
		// Only new elements to add
		if unMatchedNewCount == 1 {
			// Single element
			for i := 0; i < newLen; i++ {
				if !newMatched[i] {
					newElem := newVal.Index(i)
					patch.addOperation("$push", fieldPath, newElem.Interface())
					break
				}
			}
		} else {
			// Multiple elements - use $each
			elementsToAdd := make([]interface{}, 0, unMatchedNewCount)
			for i := 0; i < newLen; i++ {
				if !newMatched[i] {
					newElem := newVal.Index(i)
					elementsToAdd = append(elementsToAdd, newElem.Interface())
				}
			}
			patch.addOperation("$push", fieldPath, map[string]interface{}{
				"$each": elementsToAdd,
			})
		}
		return nil
	}

	// If no changes were detected or applied
	if !hasChanges {
		return nil
	}

	return nil
}

// elementsAreSimilar determines if two array elements are similar enough to be considered a match
func elementsAreSimilar(oldElem, newElem reflect.Value) bool {
	// For structs, we can implement more sophisticated matching
	if oldElem.Kind() == reflect.Struct && newElem.Kind() == reflect.Struct {
		return structsAreSimilar(oldElem, newElem)
	}

	// For non-structs, they must be equal to be similar
	return reflect.DeepEqual(oldElem.Interface(), newElem.Interface())
}

// structsAreSimilar determines if two structs are similar based on key fields
func structsAreSimilar(oldStruct, newStruct reflect.Value) bool {
	structType := oldStruct.Type()

	// Look for common identifier fields
	identifierFields := []string{"ID", "Id", "Key", "Name", "UUID", "Uuid"}

	for _, fieldName := range identifierFields {
		if field, found := structType.FieldByName(fieldName); found && field.IsExported() {
			oldFieldVal := oldStruct.FieldByName(fieldName)
			newFieldVal := newStruct.FieldByName(fieldName)

			if oldFieldVal.IsValid() && newFieldVal.IsValid() {
				// If both have the same identifier value, they're similar
				if isEqual(oldFieldVal.Interface(), newFieldVal.Interface()) {
					return true
				}
			}
		}
	}

	// If no identifier field found or they don't match,
	// consider them similar only if they're identical
	return reflect.DeepEqual(oldStruct.Interface(), newStruct.Interface())
}

// Enhanced handleArraySmart with better heuristics
func handleArraySmartEnhanced(patch *BsonPatch, config *DiffConfig, fieldPath string, oldVal, newVal reflect.Value) error {
	// Check if arrays are equal first
	if reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
		return nil // No change needed
	}

	oldLen := oldVal.Len()
	newLen := newVal.Len()

	// Smart heuristics for when to use different strategies
	lengthDiff := abs(oldLen - newLen)
	maxLen := max(oldLen, newLen)

	// If length difference is too large, use replace
	if maxLen > 0 && lengthDiff > maxLen/2 {
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// If arrays are small, replace is often more efficient
	if maxLen <= 3 {
		return handleArrayReplace(patch, config, fieldPath, oldVal, newVal)
	}

	// Use existing smart logic for complex cases
	return handleArraySmartStructs(patch, config, fieldPath, oldVal, newVal)
}

// trackRootPointers tracks pointers at the root level of old and new values
func trackRootPointers(tracker *PointerTracker, oldValue, newValue interface{}) {
	trackPointersInValue(tracker, true, "", reflect.ValueOf(oldValue))
	trackPointersInValue(tracker, false, "", reflect.ValueOf(newValue))
}

// trackPointersInValue recursively tracks all pointers in a reflect.Value
func trackPointersInValue(tracker *PointerTracker, isOld bool, basePath string, val reflect.Value) {
	if !val.IsValid() {
		return
	}

	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			addr := val.Pointer()
			if addr != 0 {
				tracker.TrackPointer(isOld, basePath, addr)
				// Track the pointed-to value as well
				trackPointersInValue(tracker, isOld, basePath, val.Elem())
			}
		}
	case reflect.Struct:
		structType := val.Type()
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			bsonTag := field.Tag.Get("bson")
			bsonFieldName := getBsonFieldName(fieldName, bsonTag)

			if bsonFieldName == "-" {
				continue
			}

			var fieldPath string
			if basePath == "" {
				fieldPath = bsonFieldName
			} else {
				fieldPath = basePath + "." + bsonFieldName
			}

			trackPointersInValue(tracker, isOld, fieldPath, val.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			indexPath := fmt.Sprintf("%s[%d]", basePath, i)
			trackPointersInValue(tracker, isOld, indexPath, val.Index(i))
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			keyStr := key.String()
			var keyPath string
			if basePath == "" {
				keyPath = keyStr
			} else {
				keyPath = basePath + "." + keyStr
			}
			trackPointersInValue(tracker, isOld, keyPath, val.MapIndex(key))
		}
	}
}
