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

// TestFilterOperation_Type tests the Type method implementation
func TestFilterOperation_Type(t *testing.T) {
	t.Run("FilterOperation returns correct type", func(t *testing.T) {
		op := &FilterOperation[int]{
			predicate: func(x int) bool { return x > 0 },
		}
		
		assert.Equal(t, FilterOp, op.Type())
		assert.Equal(t, "Filter", op.Type().String())
	})
}

// TestFilterOperation_Apply tests the Apply method with different scenarios
func TestFilterOperation_Apply(t *testing.T) {
	ctx := context.Background()
	
	t.Run("Apply with simple predicate - match", func(t *testing.T) {
		op := &FilterOperation[int]{
			predicate: func(x int) bool { return x%2 == 0 },
		}
		
		result, err := op.Apply(ctx, 4)
		require.NoError(t, err)
		assert.Equal(t, 4, result)
	})
	
	t.Run("Apply with simple predicate - no match", func(t *testing.T) {
		op := &FilterOperation[int]{
			predicate: func(x int) bool { return x%2 == 0 },
		}
		
		result, err := op.Apply(ctx, 3)
		require.NoError(t, err)
		assert.Nil(t, result) // Filtered out
	})
	
	t.Run("Apply with context-aware predicate - match", func(t *testing.T) {
		op := &FilterOperation[string]{
			safePredicateWithContext: func(ctx context.Context, s string) (bool, error) {
				return len(s) > 3, nil
			},
		}
		
		result, err := op.Apply(ctx, "hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})
	
	t.Run("Apply with context-aware predicate - no match", func(t *testing.T) {
		op := &FilterOperation[string]{
			safePredicateWithContext: func(ctx context.Context, s string) (bool, error) {
				return len(s) > 5, nil
			},
		}
		
		result, err := op.Apply(ctx, "hi")
		require.NoError(t, err)
		assert.Nil(t, result) // Filtered out
	})
	
	t.Run("Apply with context-aware predicate - error", func(t *testing.T) {
		expectedErr := errors.New("predicate failed")
		op := &FilterOperation[int]{
			safePredicateWithContext: func(ctx context.Context, x int) (bool, error) {
				if x < 0 {
					return false, expectedErr
				}
				return true, nil
			},
		}
		
		result, err := op.Apply(ctx, -1)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
	
	t.Run("Apply with type assertion failure", func(t *testing.T) {
		op := &FilterOperation[int]{
			predicate: func(x int) bool { return x > 0 },
		}
		
		result, err := op.Apply(ctx, "not an int")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type assertion failed")
		assert.Nil(t, result)
	})
	
	t.Run("Apply with context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		op := &FilterOperation[int]{
			safePredicateWithContext: func(ctx context.Context, x int) (bool, error) {
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				default:
					return true, nil
				}
			},
		}
		
		result, err := op.Apply(cancelCtx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, result)
	})
}

// TestQuery_Filter tests the Filter method on Query
func TestQuery_Filter(t *testing.T) {
	t.Run("Filter method chains correctly", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		source := NewSliceIterator(data)
		
		query := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		filteredQuery := query.Filter(func(x int) bool { return x%2 == 0 })
		
		// Should return new Query with filter added to pipeline
		assert.True(t, query != filteredQuery, "Filter should return a new Query instance")
		assert.Equal(t, 1, len(filteredQuery.pipeline))
		assert.Equal(t, FilterOp, filteredQuery.pipeline[0].Type())
		assert.Equal(t, source, filteredQuery.source)
		assert.Equal(t, query.ctx, filteredQuery.ctx)
		assert.Nil(t, filteredQuery.err)
	})
	
	t.Run("Filter with existing error returns same query", func(t *testing.T) {
		existingErr := errors.New("existing error")
		query := &Query[int]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		result := query.Filter(func(x int) bool { return x > 0 })
		
		assert.Same(t, query, result)
		assert.Equal(t, existingErr, result.err)
		assert.Equal(t, 0, len(result.pipeline)) // No operation added
	})
	
	t.Run("Filter preserves pipeline order", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		source := NewSliceIterator(data)
		
		query := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		result := query.
			Filter(func(x int) bool { return x > 2 }).
			Filter(func(x int) bool { return x < 5 })
		
		assert.Equal(t, 2, len(result.pipeline))
		assert.Equal(t, FilterOp, result.pipeline[0].Type())
		assert.Equal(t, FilterOp, result.pipeline[1].Type())
	})
}

// TestQuery_FilterE tests the FilterE method with error handling
func TestQuery_FilterE(t *testing.T) {
	t.Run("FilterE method chains correctly", func(t *testing.T) {
		data := []string{"hello", "world", "test"}
		source := NewSliceIterator(data)
		
		query := &Query[string]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		filteredQuery := query.FilterE(func(ctx context.Context, s string) (bool, error) {
			return len(s) > 4, nil
		})
		
		// Should return new Query with filter added to pipeline
		assert.True(t, query != filteredQuery, "Filter should return a new Query instance")
		assert.Equal(t, 1, len(filteredQuery.pipeline))
		assert.Equal(t, FilterOp, filteredQuery.pipeline[0].Type())
		assert.Equal(t, source, filteredQuery.source)
		assert.Equal(t, query.ctx, filteredQuery.ctx)
		assert.Nil(t, filteredQuery.err)
	})
	
	t.Run("FilterE with existing error returns same query", func(t *testing.T) {
		existingErr := errors.New("existing error")
		query := &Query[string]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		result := query.FilterE(func(ctx context.Context, s string) (bool, error) {
			return len(s) > 0, nil
		})
		
		assert.Same(t, query, result)
		assert.Equal(t, existingErr, result.err)
		assert.Equal(t, 0, len(result.pipeline)) // No operation added
	})
	
	t.Run("FilterE with mixed Filter operations", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		source := NewSliceIterator(data)
		
		query := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		result := query.
			Filter(func(x int) bool { return x > 1 }).
			FilterE(func(ctx context.Context, x int) (bool, error) {
				if x > 10 {
					return false, errors.New("too large")
				}
				return x < 5, nil
			}).
			Filter(func(x int) bool { return x%2 == 0 })
		
		assert.Equal(t, 3, len(result.pipeline))
		assert.Equal(t, FilterOp, result.pipeline[0].Type())
		assert.Equal(t, FilterOp, result.pipeline[1].Type())
		assert.Equal(t, FilterOp, result.pipeline[2].Type())
	})
}

// TestFilterOperation_ContextIntegration tests context handling
func TestFilterOperation_ContextIntegration(t *testing.T) {
	t.Run("Context timeout during predicate execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		
		op := &FilterOperation[int]{
			safePredicateWithContext: func(ctx context.Context, x int) (bool, error) {
				select {
				case <-time.After(20 * time.Millisecond): // Longer than timeout
					return true, nil
				case <-ctx.Done():
					return false, ctx.Err()
				}
			},
		}
		
		result, err := op.Apply(ctx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, result)
	})
	
	t.Run("Context value propagation", func(t *testing.T) {
		type contextKey string
		const key = contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")
		
		op := &FilterOperation[string]{
			safePredicateWithContext: func(ctx context.Context, s string) (bool, error) {
				value := ctx.Value(key)
				if value == "test-value" {
					return true, nil
				}
				return false, errors.New("context value not found")
			},
		}
		
		result, err := op.Apply(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// BenchmarkFilterOperation_Apply benchmarks the Apply method performance
func BenchmarkFilterOperation_Apply(b *testing.B) {
	ctx := context.Background()
	
	b.Run("Simple predicate", func(b *testing.B) {
		op := &FilterOperation[int]{
			predicate: func(x int) bool { return x%2 == 0 },
		}
		
		for i := 0; i < b.N; i++ {
			_, _ = op.Apply(ctx, i)
		}
	})
	
	b.Run("Context-aware predicate", func(b *testing.B) {
		op := &FilterOperation[int]{
			safePredicateWithContext: func(ctx context.Context, x int) (bool, error) {
				return x%2 == 0, nil
			},
		}
		
		for i := 0; i < b.N; i++ {
			_, _ = op.Apply(ctx, i)
		}
	})
}

// TestFilterOperation_RealWorldScenarios tests realistic usage scenarios
func TestFilterOperation_RealWorldScenarios(t *testing.T) {
	t.Run("Filtering user data with validation", func(t *testing.T) {
		type User struct {
			ID       int
			Name     string
			IsActive bool
		}
		
		users := []User{
			{ID: 1, Name: "Alice", IsActive: true},
			{ID: 2, Name: "Bob", IsActive: false},
			{ID: 3, Name: "Charlie", IsActive: true},
			{ID: 4, Name: "", IsActive: true}, // Invalid name
		}
		
		op := &FilterOperation[User]{
			safePredicateWithContext: func(ctx context.Context, u User) (bool, error) {
				if u.Name == "" {
					return false, fmt.Errorf("user %d has empty name", u.ID)
				}
				return u.IsActive, nil
			},
		}
		
		ctx := context.Background()
		
		// Valid active user
		result, err := op.Apply(ctx, users[0])
		require.NoError(t, err)
		assert.Equal(t, users[0], result)
		
		// Valid inactive user (filtered out)
		result, err = op.Apply(ctx, users[1])
		require.NoError(t, err)
		assert.Nil(t, result)
		
		// Invalid user (error)
		result, err = op.Apply(ctx, users[3])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty name")
		assert.Nil(t, result)
	})
}