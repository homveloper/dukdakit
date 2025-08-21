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

// SimpleFrom creates a new Query from a slice data source.
// This is a simplified version for integration testing.
func SimpleFrom[T any](source []T) *Query[T] {
	return &Query[T]{
		source:   NewSliceIterator(source),
		pipeline: []Operation{},
		ctx:      context.Background(),
		err:      nil,
	}
}

// SimpleToSlice executes the pipeline and collects all results into a slice.
// This is a simplified version for integration testing.
func (q *Query[T]) SimpleToSlice(ctx context.Context) ([]T, error) {
	if q.err != nil {
		return nil, q.err
	}

	var result []T
	
	// Get the source iterator
	sourceIter := q.source
	if sourceIter == nil {
		return nil, fmt.Errorf("source iterator is nil")
	}
	defer sourceIter.Close()

	// Process each item through the pipeline
	for {
		// Get next item from source
		sourceItem, hasNext, err := sourceIter.Next(ctx)
		if err != nil {
			return result, err
		}
		if !hasNext {
			break
		}

		// Apply all operations in pipeline
		currentItem := any(sourceItem)
		var pipelineErr error

		for _, op := range q.pipeline {
			currentItem, pipelineErr = op.Apply(ctx, currentItem)
			if pipelineErr != nil {
				return result, pipelineErr
			}
			// If filtered out (nil), skip this item
			if currentItem == nil {
				break
			}
		}

		// If item passed all filters, add to results
		if currentItem != nil {
			if typedItem, ok := currentItem.(T); ok {
				result = append(result, typedItem)
			} else {
				return result, fmt.Errorf("type assertion failed: expected %T, got %T", 
					*new(T), currentItem)
			}
		}
	}

	return result, nil
}

// TestFilterChaining_IntegrationBasic tests basic Filter chain functionality
func TestFilterChaining_IntegrationBasic(t *testing.T) {
	t.Run("Single Filter operation", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5, 6}

		// Filter even numbers
		result, err := SimpleFrom(data).
			Filter(func(x int) bool { return x%2 == 0 }).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		expected := []int{2, 4, 6}
		assert.Equal(t, expected, result)
	})

	t.Run("Multiple Filter operations", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		// Filter > 3 and < 8
		result, err := SimpleFrom(data).
			Filter(func(x int) bool { return x > 3 }).
			Filter(func(x int) bool { return x < 8 }).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		expected := []int{4, 5, 6, 7}
		assert.Equal(t, expected, result)
	})

	t.Run("FilterE with error handling", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, -1, 4, 5}

		expectedErr := errors.New("negative number not allowed")
		
		result, err := SimpleFrom(data).
			FilterE(func(ctx context.Context, x int) (bool, error) {
				if x < 0 {
					return false, expectedErr
				}
				return x%2 == 0, nil
			}).
			SimpleToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		// Should have partial results before error
		assert.Equal(t, []int{2}, result)
	})

	t.Run("Mixed Filter and FilterE operations", func(t *testing.T) {
		ctx := context.Background()
		data := []string{"hello", "world", "test", "go", "lang"}

		result, err := SimpleFrom(data).
			Filter(func(s string) bool { return len(s) > 2 }).
			FilterE(func(ctx context.Context, s string) (bool, error) {
				if s == "test" {
					return false, errors.New("test word not allowed")
				}
				return len(s) <= 5, nil
			}).
			SimpleToSlice(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test word not allowed")
		// Should have processed "hello" and "world" before hitting "test"
		expected := []string{"hello", "world"}
		assert.Equal(t, expected, result)
	})
}

// TestFilterChaining_IntegrationContext tests context handling
func TestFilterChaining_IntegrationContext(t *testing.T) {
	t.Run("Context cancellation during FilterE", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		data := []int{1, 2, 3, 4, 5}

		result, err := SimpleFrom(data).
			FilterE(func(ctx context.Context, x int) (bool, error) {
				// Cancel context when we hit 3
				if x == 3 {
					cancel()
				}
				// Check cancellation
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				default:
					return x%2 == 0, nil
				}
			}).
			SimpleToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		// Should have processed some items before cancellation
		assert.True(t, len(result) <= 2)
	})

	t.Run("Context timeout during FilterE", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		data := []int{1, 2, 3, 4, 5}

		result, err := SimpleFrom(data).
			FilterE(func(ctx context.Context, x int) (bool, error) {
				// Simulate slow processing
				select {
				case <-time.After(30 * time.Millisecond):
					return x > 0, nil
				case <-ctx.Done():
					return false, ctx.Err()
				}
			}).
			SimpleToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		// Should have partial results before timeout
		assert.True(t, len(result) <= len(data))
	})

	t.Run("Context value propagation", func(t *testing.T) {
		type contextKey string
		const key = contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")

		data := []string{"a", "b", "c"}

		result, err := SimpleFrom(data).
			FilterE(func(ctx context.Context, s string) (bool, error) {
				value := ctx.Value(key)
				if value != "test-value" {
					return false, errors.New("context value not found")
				}
				return true, nil
			}).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		expected := []string{"a", "b", "c"}
		assert.Equal(t, expected, result)
	})
}

// TestFilterChaining_IntegrationComplexTypes tests with complex data types
func TestFilterChaining_IntegrationComplexTypes(t *testing.T) {
	type User struct {
		ID       int
		Name     string
		IsActive bool
		Score    float64
	}

	t.Run("Complex type filtering", func(t *testing.T) {
		ctx := context.Background()
		users := []User{
			{ID: 1, Name: "Alice", IsActive: true, Score: 85.5},
			{ID: 2, Name: "Bob", IsActive: false, Score: 72.0},
			{ID: 3, Name: "Charlie", IsActive: true, Score: 91.2},
			{ID: 4, Name: "David", IsActive: true, Score: 68.8},
			{ID: 5, Name: "Eve", IsActive: false, Score: 95.1},
		}

		// Filter active users with score > 80
		result, err := SimpleFrom(users).
			Filter(func(u User) bool { return u.IsActive }).
			Filter(func(u User) bool { return u.Score > 80.0 }).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2) // Alice and Charlie
		
		// Verify Alice
		assert.Equal(t, "Alice", result[0].Name)
		assert.Equal(t, 85.5, result[0].Score)
		
		// Verify Charlie
		assert.Equal(t, "Charlie", result[1].Name)
		assert.Equal(t, 91.2, result[1].Score)
	})

	t.Run("Error handling with complex types", func(t *testing.T) {
		ctx := context.Background()
		users := []User{
			{ID: 1, Name: "Alice", IsActive: true, Score: 85.5},
			{ID: 2, Name: "", IsActive: true, Score: 90.0}, // Invalid name
			{ID: 3, Name: "Charlie", IsActive: true, Score: 75.0},
		}

		result, err := SimpleFrom(users).
			Filter(func(u User) bool { return u.IsActive }).
			FilterE(func(ctx context.Context, u User) (bool, error) {
				if u.Name == "" {
					return false, fmt.Errorf("user %d has empty name", u.ID)
				}
				return u.Score > 80.0, nil
			}).
			SimpleToSlice(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty name")
		// Should have processed Alice before hitting the error
		assert.Len(t, result, 1)
		assert.Equal(t, "Alice", result[0].Name)
	})
}

// TestFilterChaining_IntegrationEdgeCases tests edge cases
func TestFilterChaining_IntegrationEdgeCases(t *testing.T) {
	t.Run("Empty input slice", func(t *testing.T) {
		ctx := context.Background()
		data := []int{}

		result, err := SimpleFrom(data).
			Filter(func(x int) bool { return x > 0 }).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("All items filtered out", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 3, 5, 7, 9} // All odd numbers

		result, err := SimpleFrom(data).
			Filter(func(x int) bool { return x%2 == 0 }). // Filter evens
			SimpleToSlice(ctx)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Pipeline with nil pointer checks", func(t *testing.T) {
		ctx := context.Background()
		data := []*string{
			simpleStringPtr("hello"),
			nil,
			simpleStringPtr("world"),
			nil,
			simpleStringPtr("test"),
		}

		result, err := SimpleFrom(data).
			Filter(func(s *string) bool { return s != nil }).
			Filter(func(s *string) bool { return len(*s) > 3 }).
			SimpleToSlice(ctx)

		require.NoError(t, err)
		expected := []*string{simpleStringPtr("hello"), simpleStringPtr("world"), simpleStringPtr("test")}
		assert.Len(t, result, 3)
		for i, item := range result {
			assert.Equal(t, *expected[i], *item)
		}
	})

	t.Run("No operations in pipeline", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3}

		result, err := SimpleFrom(data).SimpleToSlice(ctx)

		require.NoError(t, err)
		assert.Equal(t, data, result)
	})
}

// Helper function for creating string pointers
func simpleStringPtr(s string) *string {
	return &s
}