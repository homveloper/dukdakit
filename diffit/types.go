package diffit

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// BsonPatch represents a MongoDB update document patch
type BsonPatch struct {
	operations   map[string]interface{}
	arrayFilters []interface{}
	metadata     PatchMetadata
}

// PatchMetadata tracks information about the patch
type PatchMetadata struct {
	FieldsChanged  []string          `json:"fieldsChanged"`
	OperationTypes map[string]string `json:"operationTypes"`
	TotalChanges   int               `json:"totalChanges"`
}

// PatchInfo provides JSON-serializable information about the patch
type PatchInfo struct {
	Operations   map[string]interface{} `json:"operations"`
	ArrayFilters []interface{}          `json:"arrayFilters,omitempty"`
	Metadata     PatchMetadata          `json:"metadata"`
}

// ArrayChangeType represents the type of array change
type ArrayChangeType int

const (
	ArrayAdded ArrayChangeType = iota
	ArrayRemoved
	ArrayModified
	ArrayReplaced
)

// DiffConfig holds configuration options for diff operation
type DiffConfig struct {
	IgnoreFields         []string
	ArrayStrategy        ArrayStrategy
	DeepCompare          bool
	CustomComparers      map[string]FieldComparer
	ZeroValueHandling    ZeroValueStrategy
	DetectPointerSharing bool // When true, detects and reports pointer sharing between old and new values
	pointerTracker       *PointerTracker // internal tracker for pointer sharing detection
}

// ArrayStrategy defines how arrays should be compared
type ArrayStrategy int

const (
	ArrayReplace ArrayStrategy = iota
	ArraySmart
	ArrayAppend
	ArrayMerge
)

// ZeroValueStrategy defines how zero values should be handled
type ZeroValueStrategy int

const (
	ZeroAsUnset ZeroValueStrategy = iota
	ZeroAsSet
	ZeroIgnore
)

// Option is a function that configures DiffConfig
type Option func(*DiffConfig)

// FieldComparer defines interface for custom field comparison
type FieldComparer interface {
	Compare(oldValue, newValue interface{}) (FieldDiff, error)
}

// FieldDiff represents a single field change
type FieldDiff struct {
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
	Path      string      `json:"path"`
}

// ArrayFilterIdentifier generates a unique identifier for array filters
type ArrayFilterIdentifier struct {
	counter int
}

// BsonPatch Methods

// MarshalBSON implements bson.Marshaler interface for direct MongoDB usage
func (p *BsonPatch) MarshalBSON() ([]byte, error) {
	if len(p.operations) == 0 {
		return nil, fmt.Errorf("empty patch")
	}
	return bson.Marshal(p.operations)
}

// Operations returns the underlying MongoDB update operations
func (p *BsonPatch) Operations() map[string]interface{} {
	return p.operations
}

// ArrayFilters returns the array filters for positional updates
func (p *BsonPatch) ArrayFilters() []interface{} {
	return p.arrayFilters
}

// IsEmpty returns true if the patch contains no operations
func (p *BsonPatch) IsEmpty() bool {
	return len(p.operations) == 0
}

// MarshalJSON implements json.Marshaler interface for JSON serialization
func (p *BsonPatch) MarshalJSON() ([]byte, error) {
	info := PatchInfo{
		Operations:   p.operations,
		ArrayFilters: p.arrayFilters,
		Metadata:     p.metadata,
	}
	return json.Marshal(info)
}

// String implements fmt.Stringer interface for pretty printing
func (p *BsonPatch) String() string {
	var parts []string

	if len(p.operations) > 0 {
		operationsJSON, _ := json.MarshalIndent(p.operations, "", "  ")
		parts = append(parts, fmt.Sprintf("Operations:\n%s", string(operationsJSON)))
	}

	if len(p.arrayFilters) > 0 {
		filtersJSON, _ := json.MarshalIndent(p.arrayFilters, "", "  ")
		parts = append(parts, fmt.Sprintf("ArrayFilters:\n%s", string(filtersJSON)))
	}

	if p.metadata.TotalChanges > 0 {
		parts = append(parts, fmt.Sprintf("Changes: %d fields modified", p.metadata.TotalChanges))
	}

	if len(parts) == 0 {
		return "BsonPatch: <empty>"
	}

	return fmt.Sprintf("BsonPatch:\n%s", strings.Join(parts, "\n"))
}

// GoString implements fmt.GoStringer interface for debugging
func (p *BsonPatch) GoString() string {
	return fmt.Sprintf("&diffit.BsonPatch{operations: %#v, arrayFilters: %#v, metadata: %#v}",
		p.operations, p.arrayFilters, p.metadata)
}

// Info returns structured information about the patch for inspection
func (p *BsonPatch) Info() PatchInfo {
	return PatchInfo{
		Operations:   p.operations,
		ArrayFilters: p.arrayFilters,
		Metadata:     p.metadata,
	}
}

// Configuration option functions

// WithIgnoreFields configures fields to ignore during comparison
func WithIgnoreFields(fields ...string) Option {
	return func(config *DiffConfig) {
		config.IgnoreFields = fields
	}
}

// WithArrayStrategy configures how arrays should be compared
func WithArrayStrategy(strategy ArrayStrategy) Option {
	return func(config *DiffConfig) {
		config.ArrayStrategy = strategy
	}
}

// WithDeepCompare configures whether to perform deep comparison of nested structures
func WithDeepCompare(enabled bool) Option {
	return func(config *DiffConfig) {
		config.DeepCompare = enabled
	}
}

// WithNumericOptimization is deprecated - numeric optimization has been removed for clarity
// This function is kept for backward compatibility but does nothing
func WithNumericOptimization(enabled bool) Option {
	return func(config *DiffConfig) {
		// No-op: numeric optimization feature has been removed
	}
}

// WithCustomComparer configures a custom comparer for specific fields
func WithCustomComparer(field string, comparer FieldComparer) Option {
	return func(config *DiffConfig) {
		config.CustomComparers[field] = comparer
	}
}

// WithZeroValueHandling configures how zero values should be treated
func WithZeroValueHandling(strategy ZeroValueStrategy) Option {
	return func(config *DiffConfig) {
		config.ZeroValueHandling = strategy
	}
}