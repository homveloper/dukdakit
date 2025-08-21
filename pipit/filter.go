package pipit

import (
	"context"
	"fmt"
)

// FilterOperation represents a filtering operation in the pipeline.
// It supports both simple predicates and context-aware predicates with error handling.
//
// FilterOperation implements the Operation interface and can be applied to items
// flowing through the pipeline. Items that don't match the predicate are filtered out.
//
// Example usage:
//   data := []int{1, 2, 3, 4, 5}
//   result := From(data).Filter(func(x int) bool { return x%2 == 0 }).ToSlice(ctx)
//   // result: [2, 4]
type FilterOperation[T any] struct {
	// predicate is a simple boolean function for basic filtering
	predicate func(T) bool
	
	// safePredicateWithContext provides context-aware filtering with error handling
	// This allows for operations that might fail or need to be cancelled
	safePredicateWithContext func(context.Context, T) (bool, error)
}

// Type returns the operation type identifier for FilterOperation.
// This implements the Operation interface.
func (op *FilterOperation[T]) Type() OperationType {
	return FilterOp
}

// Apply executes the filter operation on a single item.
// It returns the item unchanged if it passes the predicate, or nil if filtered out.
//
// The method handles both simple predicates and context-aware predicates with error handling.
// For context-aware predicates, cancellation and timeouts are properly respected.
//
// Returns:
//   - item: the original item if it passes the filter, nil if filtered out
//   - error: nil on success, predicate error, or type assertion error
func (op *FilterOperation[T]) Apply(ctx context.Context, item any) (any, error) {
	// Type assertion to ensure the item is of the expected type
	typedItem, ok := item.(T)
	if !ok {
		return nil, fmt.Errorf("type assertion failed: expected %T, got %T", 
			*new(T), item)
	}
	
	// Handle context-aware predicate with error handling
	if op.safePredicateWithContext != nil {
		match, err := op.safePredicateWithContext(ctx, typedItem)
		if err != nil {
			return nil, err
		}
		if match {
			return typedItem, nil
		}
		return nil, nil // Filtered out (nil means item is removed from pipeline)
	}
	
	// Handle simple predicate
	if op.predicate != nil && op.predicate(typedItem) {
		return typedItem, nil
	}
	
	return nil, nil // Filtered out
}