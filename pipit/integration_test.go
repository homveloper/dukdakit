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

// TestFilterMapChaining_IntegrationBasic tests basic Filter+Map chain functionality
func TestFilterMapChaining_IntegrationBasic(t *testing.T) {
	t.Run("Filter then Map with integers", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5, 6}

		// Filter evens, then map to strings
		result, err := Map(
			From(data).Filter(func(x int) bool { return x%2 == 0 }),
			func(x int) string { return fmt.Sprintf("num_%d", x) },
		).ToSlice(ctx)

		require.NoError(t, err)
		expected := []string{"num_2", "num_4", "num_6"}
		assert.Equal(t, expected, result)
	})

	t.Run("Map then Filter with strings", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5}

		// Map to strings, then filter by length
		result, err := Map(
			From(data),
			func(x int) string { return fmt.Sprintf("item_%d", x) },
		).Filter(func(s string) bool { return len(s) > 6 }).ToSlice(ctx)

		require.NoError(t, err)
		// "item_1" through "item_5" all have length > 6
		expected := []string{"item_1", "item_2", "item_3", "item_4", "item_5"}
		assert.Equal(t, expected, result)
	})

	t.Run("Multiple Filter and Map operations", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		// Complex pipeline: Filter > 3, Map to string, Filter length > 2, Map to uppercase
		result, err := Map(
			Map(
				From(data).Filter(func(x int) bool { return x > 3 }).
					Filter(func(x int) bool { return x < 8 }),
				func(x int) string { return fmt.Sprintf("%d", x) },
			).Filter(func(s string) bool { return len(s) > 0 }),
			func(s string) string { return "VAL_" + s },
		).ToSlice(ctx)

		require.NoError(t, err)
		expected := []string{"VAL_4", "VAL_5", "VAL_6", "VAL_7"}
		assert.Equal(t, expected, result)
	})
}

// TestFilterMapChaining_IntegrationError tests error handling in chained operations
func TestFilterMapChaining_IntegrationError(t *testing.T) {
	t.Run("Filter error propagation", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, -1, 4, 5}

		expectedErr := errors.New("negative number not allowed")

		result, err := Map(
			From(data).FilterE(func(ctx context.Context, x int) (bool, error) {
				if x < 0 {
					return false, expectedErr
				}
				return x%2 == 0, nil
			}),
			func(x int) string { return fmt.Sprintf("num_%d", x) },
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})

	t.Run("Map error propagation", func(t *testing.T) {
		ctx := context.Background()
		data := []string{"1", "2", "invalid", "4"}

		result, err := MapE(
			From(data).Filter(func(s string) bool { return len(s) > 0 }),
			func(ctx context.Context, s string) (int, error) {
				return strconv.Atoi(s)
			},
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
		// Should have partial results up to the error
		assert.Len(t, result, 2) // [1, 2] before hitting "invalid"
	})

	t.Run("Mixed error handling in complex pipeline", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5}

		// Create a pipeline where Map operation fails on specific values
		result, err := MapE(
			From(data).Filter(func(x int) bool { return x > 2 }),
			func(ctx context.Context, x int) (string, error) {
				if x == 4 {
					return "", errors.New("value 4 is forbidden")
				}
				return fmt.Sprintf("processed_%d", x), nil
			},
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value 4 is forbidden")
		// Should have processed 3 before failing on 4
		assert.Len(t, result, 1)
		assert.Equal(t, "processed_3", result[0])
	})
}

// TestFilterMapChaining_IntegrationContext tests context handling in chained operations
func TestFilterMapChaining_IntegrationContext(t *testing.T) {
	t.Run("Context cancellation during Filter", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		data := []int{1, 2, 3, 4, 5}

		result, err := Map(
			From(data).FilterE(func(ctx context.Context, x int) (bool, error) {
				// Cancel context during processing
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
			}),
			func(x int) string { return fmt.Sprintf("num_%d", x) },
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		// Should have some partial results before cancellation
		assert.True(t, len(result) <= 2) // At most items 1,2 before cancellation
	})

	t.Run("Context timeout during Map", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		data := []int{1, 2, 3, 4, 5}

		result, err := MapE(
			From(data).Filter(func(x int) bool { return x > 0 }),
			func(ctx context.Context, x int) (string, error) {
				// Simulate slow processing
				select {
				case <-time.After(30 * time.Millisecond):
					return fmt.Sprintf("slow_%d", x), nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			},
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		// Should have some partial results before timeout
		assert.True(t, len(result) <= len(data))
	})

	t.Run("Context value propagation", func(t *testing.T) {
		type contextKey string
		const key = contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")

		data := []string{"a", "b", "c"}

		result, err := MapE(
			From(data).FilterE(func(ctx context.Context, s string) (bool, error) {
				value := ctx.Value(key)
				if value != "test-value" {
					return false, errors.New("context value not found in filter")
				}
				return true, nil
			}),
			func(ctx context.Context, s string) (string, error) {
				value := ctx.Value(key)
				if value != "test-value" {
					return "", errors.New("context value not found in map")
				}
				return s + "_processed", nil
			},
		).ToSlice(ctx)

		require.NoError(t, err)
		expected := []string{"a_processed", "b_processed", "c_processed"}
		assert.Equal(t, expected, result)
	})
}

// TestFilterMapChaining_IntegrationComplexTypes tests with complex data types
func TestFilterMapChaining_IntegrationComplexTypes(t *testing.T) {
	type User struct {
		ID       int
		Name     string
		IsActive bool
		Score    float64
	}

	type UserSummary struct {
		ID          int
		DisplayName string
		Rating      string
	}

	t.Run("Complex type transformation pipeline", func(t *testing.T) {
		ctx := context.Background()
		users := []User{
			{ID: 1, Name: "Alice", IsActive: true, Score: 85.5},
			{ID: 2, Name: "Bob", IsActive: false, Score: 72.0},
			{ID: 3, Name: "Charlie", IsActive: true, Score: 91.2},
			{ID: 4, Name: "David", IsActive: true, Score: 68.8},
			{ID: 5, Name: "Eve", IsActive: false, Score: 95.1},
		}

		// Filter active users with score > 80, then transform to summary
		result, err := MapE(
			From(users).
				Filter(func(u User) bool { return u.IsActive }).
				Filter(func(u User) bool { return u.Score > 80.0 }),
			func(ctx context.Context, u User) (UserSummary, error) {
				var rating string
				switch {
				case u.Score >= 90:
					rating = "Excellent"
				case u.Score >= 80:
					rating = "Good"
				default:
					rating = "Average"
				}

				return UserSummary{
					ID:          u.ID,
					DisplayName: fmt.Sprintf("User_%s", u.Name),
					Rating:      rating,
				}, nil
			},
		).ToSlice(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2) // Alice and Charlie

		// Verify Alice (ID: 1)
		alice := result[0]
		assert.Equal(t, 1, alice.ID)
		assert.Equal(t, "User_Alice", alice.DisplayName)
		assert.Equal(t, "Good", alice.Rating)

		// Verify Charlie (ID: 3)
		charlie := result[1]
		assert.Equal(t, 3, charlie.ID)
		assert.Equal(t, "User_Charlie", charlie.DisplayName)
		assert.Equal(t, "Excellent", charlie.Rating)
	})

	t.Run("Error handling with complex types", func(t *testing.T) {
		ctx := context.Background()
		users := []User{
			{ID: 1, Name: "Alice", IsActive: true, Score: 85.5},
			{ID: 2, Name: "", IsActive: true, Score: 90.0}, // Invalid name
			{ID: 3, Name: "Charlie", IsActive: true, Score: 75.0},
		}

		result, err := MapE(
			From(users).Filter(func(u User) bool { return u.IsActive }),
			func(ctx context.Context, u User) (UserSummary, error) {
				if u.Name == "" {
					return UserSummary{}, fmt.Errorf("user %d has empty name", u.ID)
				}
				return UserSummary{
					ID:          u.ID,
					DisplayName: u.Name,
					Rating:      "Valid",
				}, nil
			},
		).ToSlice(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty name")
		// Should have processed Alice before hitting the error
		assert.Len(t, result, 1)
		assert.Equal(t, "Alice", result[0].DisplayName)
	})
}

// TestFilterMapChaining_IntegrationEmptyAndNil tests edge cases
func TestFilterMapChaining_IntegrationEmptyAndNil(t *testing.T) {
	t.Run("Empty input slice", func(t *testing.T) {
		ctx := context.Background()
		data := []int{}

		result, err := Map(
			From(data).Filter(func(x int) bool { return x > 0 }),
			func(x int) string { return fmt.Sprintf("%d", x) },
		).ToSlice(ctx)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("All items filtered out", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 3, 5, 7, 9} // All odd numbers

		result, err := Map(
			From(data).Filter(func(x int) bool { return x%2 == 0 }), // Filter evens
			func(x int) string { return fmt.Sprintf("even_%d", x) },
		).ToSlice(ctx)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Pipeline with nil checks", func(t *testing.T) {
		ctx := context.Background()
		data := []*string{
			stringPtr("hello"),
			nil,
			stringPtr("world"),
			nil,
			stringPtr("test"),
		}

		result, err := Map(
			From(data).Filter(func(s *string) bool { return s != nil }),
			func(s *string) string { return *s + "_processed" },
		).ToSlice(ctx)

		require.NoError(t, err)
		expected := []string{"hello_processed", "world_processed", "test_processed"}
		assert.Equal(t, expected, result)
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
