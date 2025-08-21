package pipit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipitError_BasicFunctionality tests the core PipitError functionality
func TestPipitError_BasicFunctionality(t *testing.T) {
	t.Run("PipitError creation and fields", func(t *testing.T) {
		cause := fmt.Errorf("underlying error")
		ctx := context.Background()
		
		err := &PipitError{
			Op:      "Filter",
			Stage:   2,
			Cause:   cause,
			Context: ctx,
			Item:    "test-item",
		}
		
		assert.Equal(t, "Filter", err.Op)
		assert.Equal(t, 2, err.Stage)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, ctx, err.Context)
		assert.Equal(t, "test-item", err.Item)
	})
	
	t.Run("PipitError implements error interface", func(t *testing.T) {
		err := &PipitError{
			Op:    "Map",
			Stage: 1,
			Cause: fmt.Errorf("type assertion failed"),
		}
		
		// Should implement error interface
		var _ error = err
		
		expectedMsg := "pipit: Map at stage 1: type assertion failed"
		assert.Equal(t, expectedMsg, err.Error())
	})
	
	t.Run("PipitError without cause", func(t *testing.T) {
		err := &PipitError{
			Op:    "ToSlice",
			Stage: 3,
			Cause: nil,
		}
		
		expectedMsg := "pipit: ToSlice at stage 3"
		assert.Equal(t, expectedMsg, err.Error())
	})
}

// TestPipitError_Unwrap tests the error unwrapping functionality
func TestPipitError_Unwrap(t *testing.T) {
	t.Run("Unwrap returns underlying cause", func(t *testing.T) {
		cause := fmt.Errorf("original error")
		err := &PipitError{
			Op:    "Filter",
			Stage: 0,
			Cause: cause,
		}
		
		assert.Equal(t, cause, err.Unwrap())
		assert.True(t, errors.Is(err, cause))
	})
	
	t.Run("Unwrap with nil cause", func(t *testing.T) {
		err := &PipitError{
			Op:    "Map",
			Stage: 1,
			Cause: nil,
		}
		
		assert.Nil(t, err.Unwrap())
	})
	
	t.Run("Error wrapping with context errors", func(t *testing.T) {
		// Test with context.Canceled
		err := &PipitError{
			Op:    "Map",
			Stage: 1,
			Cause: context.Canceled,
		}
		
		assert.True(t, errors.Is(err, context.Canceled))
		
		// Test with context.DeadlineExceeded
		err2 := &PipitError{
			Op:    "Filter",
			Stage: 0,
			Cause: context.DeadlineExceeded,
		}
		
		assert.True(t, errors.Is(err2, context.DeadlineExceeded))
	})
}

// TestPipitError_Is tests the Is method for error comparison
func TestPipitError_Is(t *testing.T) {
	t.Run("Is with same PipitError", func(t *testing.T) {
		err1 := &PipitError{Op: "Filter", Stage: 1}
		err2 := &PipitError{Op: "Filter", Stage: 1}
		
		assert.True(t, err1.Is(err2))
	})
	
	t.Run("Is with different PipitError", func(t *testing.T) {
		err1 := &PipitError{Op: "Filter", Stage: 1}
		err2 := &PipitError{Op: "Map", Stage: 1}
		err3 := &PipitError{Op: "Filter", Stage: 2}
		
		assert.False(t, err1.Is(err2))
		assert.False(t, err1.Is(err3))
	})
	
	t.Run("Is with wrapped error", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		pipitErr := &PipitError{
			Op:    "Filter",
			Stage: 0,
			Cause: originalErr,
		}
		
		assert.True(t, pipitErr.Is(originalErr))
		assert.False(t, pipitErr.Is(fmt.Errorf("different error")))
	})
	
	t.Run("Is with nil", func(t *testing.T) {
		err := &PipitError{Op: "Filter", Stage: 0}
		assert.False(t, err.Is(nil))
	})
}

// TestQuery_WithError tests the withError helper method
func TestQuery_WithError(t *testing.T) {
	t.Run("withError creates PipitError", func(t *testing.T) {
		query := &Query[int]{
			source:   nil,
			pipeline: make([]Operation, 2), // Stage should be 2
			ctx:      context.Background(),
			err:      nil,
		}
		
		originalErr := fmt.Errorf("test error")
		result := query.withError("TestOp", originalErr)
		
		// Should return the same query
		assert.Equal(t, query, result)
		
		// Should have created a PipitError
		require.NotNil(t, query.err)
		
		pipitErr, ok := query.err.(*PipitError)
		require.True(t, ok)
		assert.Equal(t, "TestOp", pipitErr.Op)
		assert.Equal(t, 2, pipitErr.Stage)
		assert.Equal(t, originalErr, pipitErr.Cause)
		assert.Equal(t, query.ctx, pipitErr.Context)
	})
	
	t.Run("withError does nothing when error already exists", func(t *testing.T) {
		existingErr := fmt.Errorf("existing error")
		query := &Query[int]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		newErr := fmt.Errorf("new error")
		result := query.withError("NewOp", newErr)
		
		assert.Equal(t, query, result)
		assert.Equal(t, existingErr, query.err) // Should keep existing error
	})
	
	t.Run("withError does nothing when error is nil", func(t *testing.T) {
		query := &Query[int]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		result := query.withError("TestOp", nil)
		
		assert.Equal(t, query, result)
		assert.Nil(t, query.err)
	})
}

// TestQuery_WithErrorAndItem tests the withErrorAndItem helper method
func TestQuery_WithErrorAndItem(t *testing.T) {
	t.Run("withErrorAndItem includes item context", func(t *testing.T) {
		query := &Query[string]{
			source:   nil,
			pipeline: make([]Operation, 1),
			ctx:      context.Background(),
			err:      nil,
		}
		
		originalErr := fmt.Errorf("processing failed")
		testItem := "failed-item"
		
		result := query.withErrorAndItem("Map", originalErr, testItem)
		
		assert.Equal(t, query, result)
		
		pipitErr, ok := query.err.(*PipitError)
		require.True(t, ok)
		assert.Equal(t, "Map", pipitErr.Op)
		assert.Equal(t, 1, pipitErr.Stage)
		assert.Equal(t, originalErr, pipitErr.Cause)
		assert.Equal(t, testItem, pipitErr.Item)
	})
}

// TestQuery_ErrorHelperMethods tests HasError and Error methods
func TestQuery_ErrorHelperMethods(t *testing.T) {
	t.Run("HasError with no error", func(t *testing.T) {
		query := &Query[int]{err: nil}
		assert.False(t, query.HasError())
		assert.Nil(t, query.Error())
	})
	
	t.Run("HasError with error", func(t *testing.T) {
		err := fmt.Errorf("test error")
		query := &Query[int]{err: err}
		assert.True(t, query.HasError())
		assert.Equal(t, err, query.Error())
	})
}

// TestCommonErrors tests predefined error variables
func TestCommonErrors(t *testing.T) {
	t.Run("ErrEmptyPipeline", func(t *testing.T) {
		assert.Equal(t, "EmptyPipeline", ErrEmptyPipeline.Op)
		assert.Equal(t, -1, ErrEmptyPipeline.Stage)
		assert.NotNil(t, ErrEmptyPipeline.Cause)
		assert.Contains(t, ErrEmptyPipeline.Error(), "pipeline has no operations")
	})
	
	t.Run("ErrNilSource", func(t *testing.T) {
		assert.Equal(t, "NilSource", ErrNilSource.Op)
		assert.Equal(t, -1, ErrNilSource.Stage)
		assert.Contains(t, ErrNilSource.Error(), "data source is nil")
	})
	
	t.Run("ErrClosedIterator", func(t *testing.T) {
		assert.Equal(t, "ClosedIterator", ErrClosedIterator.Op)
		assert.Contains(t, ErrClosedIterator.Error(), "iterator is closed")
	})
}

// TestNewPipitError tests the constructor functions
func TestNewPipitError(t *testing.T) {
	t.Run("NewPipitError basic", func(t *testing.T) {
		cause := fmt.Errorf("test cause")
		err := NewPipitError("TestOp", 5, cause)
		
		assert.Equal(t, "TestOp", err.Op)
		assert.Equal(t, 5, err.Stage)
		assert.Equal(t, cause, err.Cause)
		assert.Nil(t, err.Context)
	})
	
	t.Run("NewPipitErrorWithContext", func(t *testing.T) {
		cause := fmt.Errorf("test cause")
		ctx := context.WithValue(context.Background(), "key", "value")
		
		err := NewPipitErrorWithContext("TestOp", 3, cause, ctx)
		
		assert.Equal(t, "TestOp", err.Op)
		assert.Equal(t, 3, err.Stage)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, ctx, err.Context)
	})
}

// TestPipitError_ContextIntegration tests error handling with context
func TestPipitError_ContextIntegration(t *testing.T) {
	t.Run("Error with canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		err := &PipitError{
			Op:      "Filter",
			Stage:   0,
			Cause:   ctx.Err(),
			Context: ctx,
		}
		
		assert.True(t, errors.Is(err, context.Canceled))
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("Error with deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(1 * time.Millisecond) // Ensure timeout
		
		err := &PipitError{
			Op:      "Map",
			Stage:   1,
			Cause:   ctx.Err(),
			Context: ctx,
		}
		
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
		assert.Contains(t, err.Error(), "deadline exceeded")
	})
}

// BenchmarkPipitError_Creation benchmarks error creation performance
func BenchmarkPipitError_Creation(b *testing.B) {
	cause := fmt.Errorf("benchmark error")
	ctx := context.Background()
	
	b.Run("PipitError creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &PipitError{
				Op:      "BenchOp",
				Stage:   i % 10,
				Cause:   cause,
				Context: ctx,
			}
		}
	})
	
	b.Run("PipitError Error() method", func(b *testing.B) {
		err := &PipitError{
			Op:    "BenchOp",
			Stage: 1,
			Cause: cause,
		}
		
		for i := 0; i < b.N; i++ {
			_ = err.Error()
		}
	})
	
	b.Run("Error unwrapping", func(b *testing.B) {
		err := &PipitError{
			Op:    "BenchOp",
			Stage: 1,
			Cause: cause,
		}
		
		for i := 0; i < b.N; i++ {
			_ = errors.Is(err, cause)
		}
	})
}

// TestPipitError_RealWorldScenarios tests realistic error scenarios
func TestPipitError_RealWorldScenarios(t *testing.T) {
	t.Run("Type assertion failure", func(t *testing.T) {
		err := &PipitError{
			Op:    "Map",
			Stage: 2,
			Cause: fmt.Errorf("type assertion failed: expected int, got string"),
			Item:  "invalid-item",
		}
		
		assert.Contains(t, err.Error(), "Map at stage 2")
		assert.Contains(t, err.Error(), "type assertion failed")
		assert.Equal(t, "invalid-item", err.Item)
	})
	
	t.Run("Network timeout during processing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		time.Sleep(2 * time.Millisecond) // Simulate timeout
		
		err := &PipitError{
			Op:      "Filter",
			Stage:   0,
			Cause:   ctx.Err(),
			Context: ctx,
		}
		
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
		assert.Contains(t, err.Error(), "Filter at stage 0")
	})
	
	t.Run("Chained error propagation", func(t *testing.T) {
		// Simulate nested error from database layer
		dbErr := fmt.Errorf("database connection failed")
		serviceErr := fmt.Errorf("user service unavailable: %w", dbErr)
		
		pipitErr := &PipitError{
			Op:    "Filter",
			Stage: 1,
			Cause: serviceErr,
		}
		
		// Should be able to unwrap through multiple layers
		assert.True(t, errors.Is(pipitErr, dbErr))
		assert.True(t, errors.Is(pipitErr, serviceErr))
	})
}