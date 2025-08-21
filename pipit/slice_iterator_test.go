package pipit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSliceIterator_BasicFunctionality tests core SliceIterator functionality
func TestSliceIterator_BasicFunctionality(t *testing.T) {
	t.Run("NewSliceIterator creates iterator correctly", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		assert.NotNil(t, iter)
		assert.False(t, iter.IsClosed())
		assert.Equal(t, 5, iter.Len())
		assert.Equal(t, 5, iter.Remaining())
		assert.True(t, iter.HasNext())
	})
	
	t.Run("Iterator copies slice data", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		// Modify original slice
		data[0] = 999
		
		ctx := context.Background()
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.True(t, hasNext)
		assert.Equal(t, 1, item) // Should be original value, not modified
	})
	
	t.Run("Empty slice iterator", func(t *testing.T) {
		iter := NewSliceIterator([]int{})
		defer iter.Close()
		
		assert.False(t, iter.HasNext())
		assert.Equal(t, 0, iter.Len())
		assert.Equal(t, 0, iter.Remaining())
		
		ctx := context.Background()
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.False(t, hasNext)
		assert.Equal(t, 0, item) // zero value
	})
}

// TestSliceIterator_Next tests the Next() method thoroughly
func TestSliceIterator_Next(t *testing.T) {
	t.Run("Successful iteration through all items", func(t *testing.T) {
		data := []string{"a", "b", "c"}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		
		// First item
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.True(t, hasNext)
		assert.Equal(t, "a", item)
		assert.Equal(t, 2, iter.Remaining())
		
		// Second item
		item, hasNext, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.True(t, hasNext)
		assert.Equal(t, "b", item)
		assert.Equal(t, 1, iter.Remaining())
		
		// Third item
		item, hasNext, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.True(t, hasNext)
		assert.Equal(t, "c", item)
		assert.Equal(t, 0, iter.Remaining())
		
		// End of iteration
		item, hasNext, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.False(t, hasNext)
		assert.Equal(t, "", item) // zero value for string
		assert.False(t, iter.HasNext())
	})
	
	t.Run("Context cancellation during Next", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		item, hasNext, err := iter.Next(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.False(t, hasNext)
		assert.Equal(t, 0, item) // zero value
	})
	
	t.Run("Context timeout during Next", func(t *testing.T) {
		data := []float64{1.1, 2.2, 3.3}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(1 * time.Millisecond) // Ensure timeout
		
		item, hasNext, err := iter.Next(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.False(t, hasNext)
		assert.Equal(t, 0.0, item) // zero value for float64
	})
	
	t.Run("Next on closed iterator", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		
		// Close the iterator
		err := iter.Close()
		require.NoError(t, err)
		
		ctx := context.Background()
		item, hasNext, err := iter.Next(ctx)
		assert.Error(t, err)
		assert.False(t, hasNext)
		assert.Equal(t, 0, item)
		
		// Check that it's a PipitError with correct cause
		var pipitErr *PipitError
		require.True(t, errors.As(err, &pipitErr))
		assert.Equal(t, "SliceIterator", pipitErr.Op)
	})
}

// TestSliceIterator_HasNext tests the HasNext() method
func TestSliceIterator_HasNext(t *testing.T) {
	t.Run("HasNext changes as iteration progresses", func(t *testing.T) {
		data := []int{10, 20}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		
		// Initially has items
		assert.True(t, iter.HasNext())
		
		// After first item
		_, _, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.True(t, iter.HasNext())
		
		// After second item
		_, _, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.False(t, iter.HasNext())
	})
	
	t.Run("HasNext on closed iterator", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		
		assert.True(t, iter.HasNext())
		
		iter.Close()
		assert.False(t, iter.HasNext())
	})
}

// TestSliceIterator_Close tests the Close() method
func TestSliceIterator_Close(t *testing.T) {
	t.Run("Close marks iterator as closed", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		
		assert.False(t, iter.IsClosed())
		
		err := iter.Close()
		require.NoError(t, err)
		assert.True(t, iter.IsClosed())
		assert.False(t, iter.HasNext())
		assert.Equal(t, 0, iter.Len())
		assert.Equal(t, 0, iter.Remaining())
	})
	
	t.Run("Multiple Close calls are safe", func(t *testing.T) {
		data := []string{"a", "b"}
		iter := NewSliceIterator(data)
		
		// First close
		err := iter.Close()
		require.NoError(t, err)
		assert.True(t, iter.IsClosed())
		
		// Second close should be a no-op
		err = iter.Close()
		require.NoError(t, err)
		assert.True(t, iter.IsClosed())
	})
}

// TestSliceIterator_Reset tests the Reset() method
func TestSliceIterator_Reset(t *testing.T) {
	t.Run("Reset allows reiterating", func(t *testing.T) {
		data := []int{100, 200, 300}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		
		// Consume some items
		item, _, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 100, item)
		
		item, _, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 200, item)
		
		assert.Equal(t, 1, iter.Remaining())
		
		// Reset
		err = iter.Reset()
		require.NoError(t, err)
		
		// Should be back at the beginning
		assert.Equal(t, 3, iter.Remaining())
		assert.True(t, iter.HasNext())
		
		// First item should be available again
		item, _, err = iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 100, item)
	})
	
	t.Run("Reset on closed iterator fails", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		iter.Close()
		
		err := iter.Reset()
		assert.Error(t, err)
		
		var pipitErr *PipitError
		require.True(t, errors.As(err, &pipitErr))
		assert.Equal(t, "SliceIterator", pipitErr.Op)
	})
}

// TestSliceIterator_ToSlice tests the ToSlice() convenience method
func TestSliceIterator_ToSlice(t *testing.T) {
	t.Run("ToSlice collects all remaining items", func(t *testing.T) {
		data := []string{"x", "y", "z"}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		
		// Consume first item
		item, _, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, "x", item)
		
		// ToSlice should get remaining items
		remaining, err := iter.ToSlice(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{"y", "z"}, remaining)
		
		// Iterator should be exhausted
		assert.False(t, iter.HasNext())
		assert.Equal(t, 0, iter.Remaining())
	})
	
	t.Run("ToSlice with context cancellation", func(t *testing.T) {
		data := make([]int, 10) // Smaller dataset for reliable timing
		for i := range data {
			data[i] = i
		}
		
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately before calling ToSlice
		
		// ToSlice should return error due to cancelled context
		partial, err := iter.ToSlice(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		// May have no partial results since context is cancelled immediately
		assert.True(t, len(partial) <= len(data))
	})
	
	t.Run("ToSlice on empty iterator", func(t *testing.T) {
		iter := NewSliceIterator([]int{})
		defer iter.Close()
		
		ctx := context.Background()
		result, err := iter.ToSlice(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// TestSliceIterator_ThreadSafety tests concurrent access to the iterator
func TestSliceIterator_ThreadSafety(t *testing.T) {
	t.Run("Concurrent Next calls", func(t *testing.T) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make([]int, 0)
		errors := make([]error, 0)
		
		// Start multiple goroutines calling Next()
		numGoroutines := 10
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for {
					item, hasNext, err := iter.Next(ctx)
					
					mu.Lock()
					if err != nil {
						errors = append(errors, err)
					} else if hasNext {
						results = append(results, item)
					}
					mu.Unlock()
					
					if !hasNext || err != nil {
						break
					}
				}
			}()
		}
		
		wg.Wait()
		
		// Should have no errors and all items (though order may vary)
		assert.Empty(t, errors)
		assert.Len(t, results, 100)
		
		// Verify all numbers are present (though order may be different)
		resultMap := make(map[int]bool)
		for _, r := range results {
			resultMap[r] = true
		}
		assert.Len(t, resultMap, 100) // No duplicates
		
		for i := 0; i < 100; i++ {
			assert.True(t, resultMap[i], "Missing number %d", i)
		}
	})
	
	t.Run("Concurrent HasNext and Next calls", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		var wg sync.WaitGroup
		
		// Goroutine calling HasNext repeatedly
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				iter.HasNext() // Should not panic or race
				time.Sleep(1 * time.Microsecond)
			}
		}()
		
		// Goroutine calling Next
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, hasNext, err := iter.Next(ctx)
				if err != nil || !hasNext {
					break
				}
				time.Sleep(1 * time.Microsecond)
			}
		}()
		
		wg.Wait() // Should complete without deadlock or panic
	})
	
	t.Run("Concurrent Close calls", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewSliceIterator(data)
		
		var wg sync.WaitGroup
		
		// Multiple goroutines calling Close()
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := iter.Close()
				assert.NoError(t, err) // Should always succeed
			}()
		}
		
		wg.Wait()
		assert.True(t, iter.IsClosed())
	})
}

// TestSliceIterator_GenericTypes tests the iterator with different types
func TestSliceIterator_GenericTypes(t *testing.T) {
	t.Run("String iterator", func(t *testing.T) {
		data := []string{"hello", "world", "pipit"}
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		var results []string
		
		for {
			item, hasNext, err := iter.Next(ctx)
			require.NoError(t, err)
			if !hasNext {
				break
			}
			results = append(results, item)
		}
		
		assert.Equal(t, data, results)
	})
	
	t.Run("Struct iterator", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		data := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		var results []Person
		
		for {
			item, hasNext, err := iter.Next(ctx)
			require.NoError(t, err)
			if !hasNext {
				break
			}
			results = append(results, item)
		}
		
		assert.Equal(t, data, results)
	})
	
	t.Run("Pointer iterator", func(t *testing.T) {
		val1, val2 := 42, 84
		data := []*int{&val1, &val2}
		
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		ctx := context.Background()
		var results []*int
		
		for {
			item, hasNext, err := iter.Next(ctx)
			require.NoError(t, err)
			if !hasNext {
				break
			}
			results = append(results, item)
		}
		
		require.Len(t, results, 2)
		assert.Equal(t, 42, *results[0])
		assert.Equal(t, 84, *results[1])
	})
}

// BenchmarkSliceIterator_Performance benchmarks iterator performance
func BenchmarkSliceIterator_Performance(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	
	b.Run("NewSliceIterator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewSliceIterator(data)
			iter.Close()
		}
	})
	
	b.Run("Next iteration", func(b *testing.B) {
		ctx := context.Background()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			iter := NewSliceIterator(data)
			
			for {
				_, hasNext, err := iter.Next(ctx)
				if err != nil || !hasNext {
					break
				}
			}
			
			iter.Close()
		}
	})
	
	b.Run("HasNext calls", func(b *testing.B) {
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			iter.HasNext()
		}
	})
	
	b.Run("ToSlice conversion", func(b *testing.B) {
		ctx := context.Background()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			iter := NewSliceIterator(data)
			_, err := iter.ToSlice(ctx)
			if err != nil {
				b.Fatal(err)
			}
			iter.Close()
		}
	})
}

// TestSliceIterator_MemoryUsage tests memory efficiency
func TestSliceIterator_MemoryUsage(t *testing.T) {
	t.Run("Large slice handling", func(t *testing.T) {
		// Create a reasonably large slice
		size := 10000
		data := make([]int, size)
		for i := range data {
			data[i] = i
		}
		
		iter := NewSliceIterator(data)
		defer iter.Close()
		
		assert.Equal(t, size, iter.Len())
		assert.Equal(t, size, iter.Remaining())
		
		// Consume half
		ctx := context.Background()
		for i := 0; i < size/2; i++ {
			_, hasNext, err := iter.Next(ctx)
			require.NoError(t, err)
			require.True(t, hasNext)
		}
		
		assert.Equal(t, size/2, iter.Remaining())
	})
	
	t.Run("Close clears data for GC", func(t *testing.T) {
		data := []string{"test1", "test2", "test3"}
		iter := NewSliceIterator(data)
		
		// Before close
		assert.Equal(t, 3, iter.Len())
		
		// After close
		iter.Close()
		assert.Equal(t, 0, iter.Len()) // Data should be cleared
		assert.True(t, iter.IsClosed())
	})
}