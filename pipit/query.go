package pipit

import (
	"context"
	"fmt"
)

// Query represents a lazy-evaluated data pipeline with generic type safety.
// It supports method chaining and deferred execution until a terminal operation is called.
//
// Example usage:
//
//	result := pipit.From([]int{1, 2, 3, 4, 5}).
//	    Filter(func(x int) bool { return x > 2 }).
//	    Map(func(x int) string { return fmt.Sprintf("num_%d", x) }).
//	    ToSlice()
type Query[T any] struct {
	// source holds the original data iterator
	source Iterator[T]

	// pipeline contains the sequence of operations to be applied
	pipeline []Operation

	// ctx holds the context for cancellation and timeout support
	ctx context.Context

	// err tracks any error that occurred during pipeline building
	err error
}

// Iterator represents a lazy data source that can produce items on demand.
// It supports context-aware iteration for cancellation and timeout handling.
type Iterator[T any] interface {
	// Next returns the next item, whether more items exist, and any error.
	// It respects context cancellation and should return ctx.Err() when cancelled.
	Next(ctx context.Context) (T, bool, error)

	// HasNext indicates if there are more items available.
	// This is a hint and may not be accurate for infinite iterators.
	HasNext() bool

	// Close releases any resources held by the iterator.
	// It should be called when the iterator is no longer needed.
	Close() error
}

// Operation represents a single transformation step in the pipeline.
// Each operation is applied to items as they flow through the pipeline.
type Operation interface {
	// Apply transforms an input item and returns the result.
	// It may filter out items by returning (nil, nil) or transform types.
	Apply(ctx context.Context, item any) (any, error)

	// Type returns the operation type for debugging and optimization.
	Type() OperationType
}

// OperationType identifies the kind of operation for optimization and debugging.
type OperationType int

const (
	// FilterOp represents filtering operations (Where, Filter)
	FilterOp OperationType = iota

	// MapOp represents transformation operations (Map, Select)
	MapOp

	// TakeOp represents limiting operations (Take, Limit)
	TakeOp

	// DropOp represents skipping operations (Drop, Skip)
	DropOp

	// DistinctOp represents deduplication operations
	DistinctOp
)

// String returns a human-readable name for the operation type.
func (ot OperationType) String() string {
	switch ot {
	case FilterOp:
		return "Filter"
	case MapOp:
		return "Map"
	case TakeOp:
		return "Take"
	case DropOp:
		return "Drop"
	case DistinctOp:
		return "Distinct"
	default:
		return fmt.Sprintf("Unknown(%d)", int(ot))
	}
}

// From creates a new Query from a slice data source.
// This is the primary entry point for creating data pipelines.
func From[T any](source []T) *Query[T] {
	return &Query[T]{
		source:   NewSliceIterator(source),
		pipeline: []Operation{},
		ctx:      context.Background(),
		err:      nil,
	}
}

// ToSlice executes the pipeline and collects all results into a slice.
// This is a terminal operation that triggers pipeline execution.
func (q *Query[T]) ToSlice(ctx context.Context) ([]T, error) {
	if q.err != nil {
		return nil, q.err
	}

	var result []T
	iter := q.execute(ctx)
	defer iter.Close()

	for {
		item, hasNext, err := iter.Next(ctx)
		if err != nil {
			return result, err // Return partial results with error
		}
		if !hasNext {
			break
		}
		result = append(result, item)
	}

	return result, nil
}

// execute runs the pipeline and returns all results by processing the entire pipeline.
// This is a simplified approach that processes all items through the pipeline.
func (q *Query[T]) execute(ctx context.Context) Iterator[T] {
	if q.err != nil {
		return &errorIterator[T]{err: q.err}
	}

	// For Map operations, we need to trace back to find the original source
	sourceQuery := q
	for sourceQuery.source == nil && len(sourceQuery.pipeline) > 0 {
		// This is a hack - we need to find a better architecture
		// For now, return error for unsupported pipeline structure
		return &errorIterator[T]{err: fmt.Errorf("cannot execute pipeline with nil source - Map operations need better integration")}
	}

	// Process all items through the pipeline and collect results
	var results []T
	var processingErr error

	// Start with source iterator
	sourceIter := sourceQuery.source
	if sourceIter == nil {
		return &errorIterator[T]{err: fmt.Errorf("source iterator is nil")}
	}
	defer sourceIter.Close()

	for {
		// Get next item from source
		sourceItem, hasNext, err := sourceIter.Next(ctx)
		if err != nil {
			processingErr = err
			break
		}
		if !hasNext {
			break
		}

		// Process through the pipeline
		currentItem := any(sourceItem)
		var pipelineErr error

		for _, op := range q.pipeline {
			currentItem, pipelineErr = op.Apply(ctx, currentItem)
			if pipelineErr != nil {
				processingErr = pipelineErr
				break
			}
			// If filtered out (nil), skip this item
			if currentItem == nil {
				break
			}
		}

		if pipelineErr != nil {
			break
		}

		// If item wasn't filtered out, add to results
		if currentItem != nil {
			if typedItem, ok := currentItem.(T); ok {
				results = append(results, typedItem)
			} else {
				processingErr = fmt.Errorf("type assertion failed: expected %T, got %T",
					*new(T), currentItem)
				break
			}
		}
	}

	// Return iterator over results
	if processingErr != nil {
		return &errorIterator[T]{err: processingErr}
	}

	return NewSliceIterator(results)
}

// Filter applies a predicate function to each element in the pipeline.
// Elements that don't satisfy the predicate (return false) are removed from the pipeline.
//
// This is a lazy operation - the predicate is not executed until a terminal operation
// is called. The filtering happens as items flow through the pipeline.
//
// Example:
//
//	data := []int{1, 2, 3, 4, 5}
//	query := From(data).Filter(func(x int) bool { return x%2 == 0 })
//	result := query.ToSlice(ctx) // [2, 4]
//
// The predicate function should be pure (no side effects) and fast, as it will
// be called for each item during pipeline execution.
//
// If the Query already has an error, Filter returns the same Query unchanged.
func (q *Query[T]) Filter(predicate func(T) bool) *Query[T] {
	if q.err != nil {
		return q // Fail-fast: propagate existing errors
	}

	// Create new FilterOperation and add to pipeline
	op := &FilterOperation[T]{predicate: predicate}

	// Return new Query with the operation added to the pipeline
	return &Query[T]{
		source:   q.source,
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}
}

// FilterE applies a context-aware predicate function with error handling to each element.
// This is the error-safe version of Filter that supports context cancellation and
// predicates that can fail during execution.
//
// The predicate function receives the current context and can return an error.
// If the predicate returns an error, the entire pipeline execution will stop and
// return that error. Context cancellation is properly handled.
//
// This is a lazy operation - the predicate is not executed until a terminal operation
// is called. Any errors are captured and propagated through the pipeline.
//
// Example:
//
//	data := []string{"http://example.com", "invalid-url"}
//	query := From(data).FilterE(func(ctx context.Context, url string) (bool, error) {
//	    _, err := http.NewRequestWithContext(ctx, "GET", url, nil)
//	    if err != nil {
//	        return false, err // Invalid URL causes error
//	    }
//	    return true, nil
//	})
//	result := query.ToSlice(ctx) // Error if invalid URL encountered
//
// If the Query already has an error, FilterE returns the same Query unchanged.
func (q *Query[T]) FilterE(predicate func(context.Context, T) (bool, error)) *Query[T] {
	if q.err != nil {
		return q // Fail-fast: propagate existing errors
	}

	// Create new FilterOperation with context-aware predicate
	op := &FilterOperation[T]{safePredicateWithContext: predicate}

	// Return new Query with the operation added to the pipeline
	return &Query[T]{
		source:   q.source,
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}
}
