package pipit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapOperation_Type tests the Type method implementation
func TestMapOperation_Type(t *testing.T) {
	t.Run("MapOperation returns correct type", func(t *testing.T) {
		op := &MapOperation[int, string]{
			mapper: func(x int) string { return fmt.Sprintf("%d", x) },
		}
		
		assert.Equal(t, MapOp, op.Type())
		assert.Equal(t, "Map", op.Type().String())
	})
}

// TestMapOperation_Apply tests the Apply method with different scenarios
func TestMapOperation_Apply(t *testing.T) {
	ctx := context.Background()
	
	t.Run("Apply with simple mapper", func(t *testing.T) {
		op := &MapOperation[int, string]{
			mapper: func(x int) string { return fmt.Sprintf("num_%d", x) },
		}
		
		result, err := op.Apply(ctx, 42)
		require.NoError(t, err)
		assert.Equal(t, "num_42", result)
	})
	
	t.Run("Apply with context-aware mapper - success", func(t *testing.T) {
		op := &MapOperation[string, int]{
			safeMapperWithContext: func(ctx context.Context, s string) (int, error) {
				return strconv.Atoi(s)
			},
		}
		
		result, err := op.Apply(ctx, "123")
		require.NoError(t, err)
		assert.Equal(t, 123, result)
	})
	
	t.Run("Apply with context-aware mapper - error", func(t *testing.T) {
		op := &MapOperation[string, int]{
			safeMapperWithContext: func(ctx context.Context, s string) (int, error) {
				return strconv.Atoi(s)
			},
		}
		
		result, err := op.Apply(ctx, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
		assert.Equal(t, 0, result)
	})
	
	t.Run("Apply with type assertion failure", func(t *testing.T) {
		op := &MapOperation[int, string]{
			mapper: func(x int) string { return fmt.Sprintf("%d", x) },
		}
		
		result, err := op.Apply(ctx, "not an int")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type assertion failed")
		assert.Nil(t, result)
	})
	
	t.Run("Apply with context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		op := &MapOperation[int, string]{
			mapper: func(x int) string { return fmt.Sprintf("%d", x) },
		}
		
		result, err := op.Apply(cancelCtx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, "", result) // Zero value for string
	})
	
	t.Run("Apply with context-aware mapper and cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		op := &MapOperation[int, string]{
			safeMapperWithContext: func(ctx context.Context, x int) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				default:
					return fmt.Sprintf("%d", x), nil
				}
			},
		}
		
		result, err := op.Apply(cancelCtx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, "", result)
	})
}

// TestMap_Function tests the Map standalone function
func TestMap_Function(t *testing.T) {
	t.Run("Map function creates correct Query[U]", func(t *testing.T) {
		data := []int{1, 2, 3}
		source := NewSliceIterator(data)
		
		queryT := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		queryU := Map(queryT, func(x int) string { return fmt.Sprintf("num_%d", x) })
		
		// Should return new Query[string] with map operation added to pipeline
		assert.Equal(t, 1, len(queryU.pipeline))
		assert.Equal(t, MapOp, queryU.pipeline[0].Type())
		assert.Equal(t, queryT.ctx, queryU.ctx)
		assert.Nil(t, queryU.err)
	})
	
	t.Run("Map with existing error returns error Query", func(t *testing.T) {
		existingErr := errors.New("existing error")
		queryT := &Query[int]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		queryU := Map(queryT, func(x int) string { return fmt.Sprintf("%d", x) })
		
		assert.Equal(t, existingErr, queryU.err)
		assert.Equal(t, 0, len(queryU.pipeline)) // No operation added
	})
	
	t.Run("Map preserves pipeline order", func(t *testing.T) {
		data := []int{1, 2, 3}
		source := NewSliceIterator(data)
		
		queryT := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Create a filter operation first
		filteredQuery := queryT.Filter(func(x int) bool { return x > 1 })
		
		// Then map
		mappedQuery := Map(filteredQuery, func(x int) string { return fmt.Sprintf("num_%d", x) })
		
		assert.Equal(t, 2, len(mappedQuery.pipeline))
		assert.Equal(t, FilterOp, mappedQuery.pipeline[0].Type())
		assert.Equal(t, MapOp, mappedQuery.pipeline[1].Type())
	})
}

// TestMapE_Function tests the MapE function with error handling
func TestMapE_Function(t *testing.T) {
	t.Run("MapE function creates correct Query[U]", func(t *testing.T) {
		data := []string{"1", "2", "3"}
		source := NewSliceIterator(data)
		
		queryT := &Query[string]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		queryU := MapE(queryT, func(ctx context.Context, s string) (int, error) {
			return strconv.Atoi(s)
		})
		
		// Should return new Query[int] with map operation added to pipeline
		assert.Equal(t, 1, len(queryU.pipeline))
		assert.Equal(t, MapOp, queryU.pipeline[0].Type())
		assert.Equal(t, queryT.ctx, queryU.ctx)
		assert.Nil(t, queryU.err)
	})
	
	t.Run("MapE with existing error returns error Query", func(t *testing.T) {
		existingErr := errors.New("existing error")
		queryT := &Query[string]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		queryU := MapE(queryT, func(ctx context.Context, s string) (int, error) {
			return strconv.Atoi(s)
		})
		
		assert.Equal(t, existingErr, queryU.err)
		assert.Equal(t, 0, len(queryU.pipeline)) // No operation added
	})
	
	t.Run("MapE with mixed operations", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		source := NewSliceIterator(data)
		
		queryT := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		result := queryT.
			Filter(func(x int) bool { return x > 2 }).
			FilterE(func(ctx context.Context, x int) (bool, error) {
				return x < 5, nil
			})
		
		mappedQuery := Map(result, func(x int) string { return fmt.Sprintf("num_%d", x) })
		
		assert.Equal(t, 3, len(mappedQuery.pipeline))
		assert.Equal(t, FilterOp, mappedQuery.pipeline[0].Type())
		assert.Equal(t, FilterOp, mappedQuery.pipeline[1].Type())
		assert.Equal(t, MapOp, mappedQuery.pipeline[2].Type())
	})
}

// TestMapOperation_ContextIntegration tests context handling
func TestMapOperation_ContextIntegration(t *testing.T) {
	t.Run("Context timeout during mapper execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		
		op := &MapOperation[int, string]{
			safeMapperWithContext: func(ctx context.Context, x int) (string, error) {
				select {
				case <-time.After(20 * time.Millisecond): // Longer than timeout
					return fmt.Sprintf("%d", x), nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			},
		}
		
		result, err := op.Apply(ctx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, "", result)
	})
	
	t.Run("Context value propagation", func(t *testing.T) {
		type contextKey string
		const key = contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")
		
		op := &MapOperation[string, string]{
			safeMapperWithContext: func(ctx context.Context, s string) (string, error) {
				value := ctx.Value(key)
				if value == "test-value" {
					return s + "_processed", nil
				}
				return "", errors.New("context value not found")
			},
		}
		
		result, err := op.Apply(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "test_processed", result)
	})
}

// BenchmarkMapOperation_Apply benchmarks the Apply method performance
func BenchmarkMapOperation_Apply(b *testing.B) {
	ctx := context.Background()
	
	b.Run("Simple mapper", func(b *testing.B) {
		op := &MapOperation[int, string]{
			mapper: func(x int) string { return fmt.Sprintf("%d", x) },
		}
		
		for i := 0; i < b.N; i++ {
			_, _ = op.Apply(ctx, i)
		}
	})
	
	b.Run("Context-aware mapper", func(b *testing.B) {
		op := &MapOperation[int, string]{
			safeMapperWithContext: func(ctx context.Context, x int) (string, error) {
				return fmt.Sprintf("%d", x), nil
			},
		}
		
		for i := 0; i < b.N; i++ {
			_, _ = op.Apply(ctx, i)
		}
	})
}

// TestMapOperation_RealWorldScenarios tests realistic usage scenarios
func TestMapOperation_RealWorldScenarios(t *testing.T) {
	t.Run("Transform user data with validation", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}
		
		type UserSummary struct {
			ID          int
			DisplayName string
		}
		
		users := []User{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 3, Name: "Charlie"},
		}
		
		op := &MapOperation[User, UserSummary]{
			safeMapperWithContext: func(ctx context.Context, u User) (UserSummary, error) {
				if u.Name == "" {
					return UserSummary{}, fmt.Errorf("user %d has empty name", u.ID)
				}
				return UserSummary{
					ID:          u.ID,
					DisplayName: fmt.Sprintf("User_%s", u.Name),
				}, nil
			},
		}
		
		ctx := context.Background()
		
		// Valid user
		result, err := op.Apply(ctx, users[0])
		require.NoError(t, err)
		summary := result.(UserSummary)
		assert.Equal(t, 1, summary.ID)
		assert.Equal(t, "User_Alice", summary.DisplayName)
	})
	
	t.Run("Type conversion with error handling", func(t *testing.T) {
		op := &MapOperation[string, int]{
			safeMapperWithContext: func(ctx context.Context, s string) (int, error) {
				// Simulate network request or complex processing
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				default:
					return strconv.Atoi(s)
				}
			},
		}
		
		ctx := context.Background()
		
		// Valid conversion
		result, err := op.Apply(ctx, "123")
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		
		// Invalid conversion
		result, err = op.Apply(ctx, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
		assert.Equal(t, 0, result)
	})
}