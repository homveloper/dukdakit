package pipit

import (
	"context"
	"fmt"
)

// PipitError represents a structured error that occurred during pipeline processing.
// It implements the error interface and provides additional context about where
// and why the error occurred in the pipeline.
//
// PipitError supports error wrapping (Go 1.13+) and can be used with errors.Is()
// and errors.As() for proper error handling and inspection.
type PipitError struct {
	// Op identifies the operation that failed (e.g., "Filter", "Map", "ToSlice")
	Op string

	// Stage indicates which stage in the pipeline the error occurred
	// (0 = first operation, 1 = second operation, etc.)
	Stage int

	// Cause is the underlying error that caused this pipeline error
	Cause error

	// Context holds the context that was active when the error occurred
	// This can provide additional information about timeouts, cancellations, etc.
	Context context.Context

	// Item represents the data item being processed when the error occurred
	// This is optional and may be nil for some error types
	Item interface{}
}

// Error returns a human-readable description of the error.
// It implements the error interface and follows Go error formatting conventions.
func (e *PipitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("pipit: %s at stage %d: %v", e.Op, e.Stage, e.Cause)
	}
	return fmt.Sprintf("pipit: %s at stage %d", e.Op, e.Stage)
}

// Unwrap returns the underlying cause of the error.
// This enables error wrapping support introduced in Go 1.13.
// It allows errors.Is() and errors.As() to work properly with PipitError.
//
// Example:
//
//	if errors.Is(err, context.Canceled) {
//	    // Handle cancellation
//	}
func (e *PipitError) Unwrap() error {
	return e.Cause
}

// Is reports whether this error matches the target error.
// It implements the Is method for enhanced error checking.
func (e *PipitError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check if target is also a PipitError with same operation
	if pipitErr, ok := target.(*PipitError); ok {
		return e.Op == pipitErr.Op && e.Stage == pipitErr.Stage
	}

	// Delegate to the wrapped error
	return e.Cause != nil && e.Cause == target
}

// withError creates a new PipitError or updates an existing Query's error state.
// This is a helper function used throughout the pipeline to maintain error context.
//
// If the Query already has an error, this function does nothing (fail-fast principle).
// Otherwise, it creates a new PipitError with the provided operation and underlying error.
func (q *Query[T]) withError(op string, err error) *Query[T] {
	if err == nil || q.err != nil {
		return q // No new error or already has an error
	}

	q.err = &PipitError{
		Op:      op,
		Stage:   len(q.pipeline), // Current pipeline length indicates the stage
		Cause:   err,
		Context: q.ctx,
		Item:    nil, // Will be set by specific operations if needed
	}
	return q
}

// withErrorAndItem creates a PipitError with additional item context.
// This is useful when we know which specific item caused the error.
func (q *Query[T]) withErrorAndItem(op string, err error, item interface{}) *Query[T] {
	if err == nil || q.err != nil {
		return q
	}

	q.err = &PipitError{
		Op:      op,
		Stage:   len(q.pipeline),
		Cause:   err,
		Context: q.ctx,
		Item:    item,
	}
	return q
}

// HasError returns true if the Query has encountered an error.
// This is a convenience method for checking error state.
func (q *Query[T]) HasError() bool {
	return q.err != nil
}

// Error returns the current error state of the Query.
// Returns nil if no error has occurred.
func (q *Query[T]) Error() error {
	return q.err
}

// Common error variables for frequent error conditions
var (
	// ErrEmptyPipeline indicates an operation was attempted on an empty pipeline
	ErrEmptyPipeline = &PipitError{
		Op:    "EmptyPipeline",
		Stage: -1,
		Cause: fmt.Errorf("pipeline has no operations"),
	}

	// ErrNilSource indicates the data source is nil
	ErrNilSource = &PipitError{
		Op:    "NilSource",
		Stage: -1,
		Cause: fmt.Errorf("data source is nil"),
	}

	// ErrClosedIterator indicates an operation was attempted on a closed iterator
	ErrClosedIterator = &PipitError{
		Op:    "ClosedIterator",
		Stage: -1,
		Cause: fmt.Errorf("iterator is closed"),
	}
)

// NewPipitError creates a new PipitError with the specified parameters.
// This is useful for creating custom errors in user code.
func NewPipitError(op string, stage int, cause error) *PipitError {
	return &PipitError{
		Op:    op,
		Stage: stage,
		Cause: cause,
	}
}

// NewPipitErrorWithContext creates a new PipitError with context information.
func NewPipitErrorWithContext(op string, stage int, cause error, ctx context.Context) *PipitError {
	return &PipitError{
		Op:      op,
		Stage:   stage,
		Cause:   cause,
		Context: ctx,
	}
}

// errorIterator implements Iterator for error cases
type errorIterator[T any] struct {
	err error
}

func (ei *errorIterator[T]) Next(ctx context.Context) (T, bool, error) {
	var zero T
	return zero, false, ei.err
}

func (ei *errorIterator[T]) HasNext() bool {
	return false
}

func (ei *errorIterator[T]) Close() error {
	return nil
}
