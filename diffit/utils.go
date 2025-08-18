package diffit

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

// Helper functions for working with BsonPatch

// addArrayFilter adds an array filter to the patch
func (p *BsonPatch) addArrayFilter(filter interface{}) {
	if p.arrayFilters == nil {
		p.arrayFilters = make([]interface{}, 0)
	}
	p.arrayFilters = append(p.arrayFilters, filter)
}

// HasArrayFilters returns true if the patch has array filters
func (p *BsonPatch) HasArrayFilters() bool {
	return len(p.arrayFilters) > 0
}

// addOperation adds a MongoDB operation to the patch
func (p *BsonPatch) addOperation(operator string, field string, value interface{}) {
	if p.operations == nil {
		p.operations = make(map[string]interface{})
	}

	if p.operations[operator] == nil {
		p.operations[operator] = make(map[string]interface{})
	}

	operatorMap := p.operations[operator].(map[string]interface{})
	operatorMap[field] = value

	// Update metadata
	p.updateMetadata(field, operator)
}

// updateMetadata updates the patch metadata with new field change
func (p *BsonPatch) updateMetadata(field, operation string) {
	if p.metadata.OperationTypes == nil {
		p.metadata.OperationTypes = make(map[string]string)
	}

	// Track field change
	found := false
	for _, existingField := range p.metadata.FieldsChanged {
		if existingField == field {
			found = true
			break
		}
	}
	if !found {
		p.metadata.FieldsChanged = append(p.metadata.FieldsChanged, field)
	}

	// Track operation type
	p.metadata.OperationTypes[field] = operation
	p.metadata.TotalChanges++
}

// Next returns the next available array filter identifier
func (a *ArrayFilterIdentifier) Next() string {
	identifier := fmt.Sprintf("elem%d", a.counter)
	a.counter++
	return identifier
}

// newBsonPatch creates a new empty BsonPatch
func newBsonPatch() *BsonPatch {
	return &BsonPatch{
		operations:   make(map[string]interface{}),
		arrayFilters: make([]interface{}, 0),
		metadata: PatchMetadata{
			FieldsChanged:  make([]string, 0),
			OperationTypes: make(map[string]string),
			TotalChanges:   0,
		},
	}
}

// newDiffConfig creates a default DiffConfig
func newDiffConfig() *DiffConfig {
	return &DiffConfig{
		ArrayStrategy:     ArraySmart,
		DeepCompare:       true,
		ZeroValueHandling: ZeroAsSet,
		CustomComparers:   make(map[string]FieldComparer),
	}
}

// isIgnoredField checks if a field should be ignored based on configuration
func isIgnoredField(fieldName string, ignoreFields []string) bool {
	for _, ignored := range ignoreFields {
		if fieldName == ignored {
			return true
		}
	}
	return false
}

// Removed isNumericType and calculateNumericDiff functions 
// as numeric optimization has been removed for clarity

// isZeroValue checks if a value is considered a zero value
func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return !v
	case time.Time:
		return v.IsZero()
	default:
		return false
	}
}

// getBsonFieldName extracts the BSON field name from struct tag or uses field name
func getBsonFieldName(fieldName string, tag string) string {
	// Parse bson tag to get field name
	if tag != "" {
		// Extract field name from tag like `bson:"field_name,omitempty"`
		if len(tag) > 0 && tag != "-" {
			// Split by comma to handle omitempty and other options
			parts := strings.Split(tag, ",")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}
	
	// Convert field name to snake_case by default
	return toSnakeCase(fieldName)
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	
	return result.String()
}

// PointerSharingError represents an error when pointer sharing is detected
type PointerSharingError struct {
	FieldPath string
	Address   uintptr
	Message   string
}

func (e *PointerSharingError) Error() string {
	return e.Message
}

// WithDetectPointerSharing enables detection of pointer sharing between old and new values
// When enabled, the diff will return an error if it detects that the same pointer is used
// in both old and new structures, which can lead to incorrect diff results
func WithDetectPointerSharing(detect bool) Option {
	return func(config *DiffConfig) {
		config.DetectPointerSharing = detect
	}
}

// PointerTracker tracks pointer addresses to detect sharing
type PointerTracker struct {
	oldPointers map[uintptr]string // address -> field path
	newPointers map[uintptr]string // address -> field path
}

// NewPointerTracker creates a new pointer tracker
func NewPointerTracker() *PointerTracker {
	return &PointerTracker{
		oldPointers: make(map[uintptr]string),
		newPointers: make(map[uintptr]string),
	}
}

// TrackPointer tracks a pointer address for the given field path
func (pt *PointerTracker) TrackPointer(isOld bool, fieldPath string, ptr uintptr) {
	if isOld {
		pt.oldPointers[ptr] = fieldPath
	} else {
		pt.newPointers[ptr] = fieldPath
	}
}

// CheckForSharing checks if there are any shared pointers between old and new values
func (pt *PointerTracker) CheckForSharing() *PointerSharingError {
	for addr, newPath := range pt.newPointers {
		if oldPath, exists := pt.oldPointers[addr]; exists {
			return &PointerSharingError{
				FieldPath: newPath,
				Address:   addr,
				Message: fmt.Sprintf("pointer sharing detected at address %p: old field '%s' and new field '%s' point to the same memory location", 
					unsafe.Pointer(addr), oldPath, newPath),
			}
		}
	}
	return nil
}