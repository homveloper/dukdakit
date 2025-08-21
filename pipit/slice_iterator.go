package pipit

import (
	"context"
	"sync"
)

// SliceIterator provides a thread-safe, context-aware iterator over a slice.
// It implements the Iterator[T] interface and supports lazy evaluation with
// proper resource management and cancellation handling.
//
// SliceIterator is designed to be the primary data source for Query pipelines,
// converting regular Go slices into the iterator pattern required for lazy evaluation.
//
// Example usage:
//   data := []int{1, 2, 3, 4, 5}
//   iter := NewSliceIterator(data)
//   defer iter.Close()
//   
//   for {
//       item, hasNext, err := iter.Next(ctx)
//       if err != nil || !hasNext {
//           break
//       }
//       // Process item
//   }
type SliceIterator[T any] struct {
	// data holds the underlying slice data
	// This is a copy of the original slice to prevent external modifications
	data []T
	
	// index tracks the current position in the slice
	index int
	
	// closed indicates whether the iterator has been closed
	closed bool
	
	// mutex provides thread-safe access to the iterator state
	// This allows multiple goroutines to safely interact with the same iterator
	mutex sync.RWMutex
}

// NewSliceIterator creates a new SliceIterator from the provided slice.
// It makes a copy of the slice to ensure data integrity and prevent
// external modifications from affecting the iteration.
//
// The iterator starts at index 0 and is ready for immediate use.
// Remember to call Close() when done to release any resources.
//
// Example:
//   numbers := []int{10, 20, 30}
//   iter := NewSliceIterator(numbers)
//   defer iter.Close()
func NewSliceIterator[T any](data []T) *SliceIterator[T] {
	// Create a copy of the slice to prevent external modifications
	// This ensures the iterator behavior is predictable and safe
	dataCopy := make([]T, len(data))
	copy(dataCopy, data)
	
	return &SliceIterator[T]{
		data:   dataCopy,
		index:  0,
		closed: false,
	}
}

// Next returns the next item from the slice, along with a boolean indicating
// whether more items are available and any error that occurred.
//
// This method respects context cancellation and will return immediately if
// the context is cancelled or times out.
//
// Returns:
//   - item: the next item in the sequence (zero value if no more items or error)
//   - hasNext: true if this item is valid and more items may be available
//   - error: nil on success, context error on cancellation, or iterator error
//
// Thread-safe: Multiple goroutines can safely call Next() concurrently.
func (si *SliceIterator[T]) Next(ctx context.Context) (T, bool, error) {
	var zero T
	
	// Check context cancellation first for immediate response
	select {
	case <-ctx.Done():
		return zero, false, ctx.Err()
	default:
	}
	
	// Acquire read lock for thread-safe access
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	// Check if iterator is closed
	if si.closed {
		return zero, false, NewPipitError("SliceIterator", si.index, ErrClosedIterator.Cause)
	}
	
	// Check if we've reached the end of the slice
	if si.index >= len(si.data) {
		return zero, false, nil // End of iteration, not an error
	}
	
	// Get the current item and advance the index
	item := si.data[si.index]
	si.index++
	
	// Return the item with hasNext=true (there might be more items)
	return item, true, nil
}

// HasNext reports whether there are more items available in the iterator.
// This is a hint method and may not be entirely accurate for all iterator types,
// but for SliceIterator it provides an accurate count.
//
// Returns false if the iterator is closed or has reached the end.
//
// Thread-safe: Multiple goroutines can safely call HasNext() concurrently.
func (si *SliceIterator[T]) HasNext() bool {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	return !si.closed && si.index < len(si.data)
}

// Close releases any resources held by the iterator and marks it as closed.
// After calling Close(), all subsequent calls to Next() will return an error.
//
// It is safe to call Close() multiple times. Subsequent calls will be no-ops.
//
// Thread-safe: Multiple goroutines can safely call Close() concurrently.
func (si *SliceIterator[T]) Close() error {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	
	// Mark as closed (idempotent operation)
	si.closed = true
	
	// Clear the data slice to help with garbage collection
	// This is especially important for slices containing pointers
	si.data = nil
	
	return nil
}

// Reset resets the iterator back to the beginning.
// This allows reusing the same iterator multiple times.
// 
// If the iterator is closed, Reset will return an error.
// Thread-safe: Can be called concurrently with other methods.
func (si *SliceIterator[T]) Reset() error {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	
	if si.closed {
		return NewPipitError("SliceIterator", si.index, ErrClosedIterator.Cause)
	}
	
	si.index = 0
	return nil
}

// Len returns the total number of items in the iterator.
// This is useful for progress tracking and allocation hints.
//
// Thread-safe: Can be called concurrently with other methods.
func (si *SliceIterator[T]) Len() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	if si.closed {
		return 0
	}
	
	return len(si.data)
}

// Remaining returns the number of items remaining in the iterator.
// This is useful for progress tracking and optimization decisions.
//
// Thread-safe: Can be called concurrently with other methods.
func (si *SliceIterator[T]) Remaining() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	if si.closed {
		return 0
	}
	
	remaining := len(si.data) - si.index
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsClosed returns true if the iterator has been closed.
// This is useful for checking iterator state in concurrent scenarios.
//
// Thread-safe: Can be called concurrently with other methods.
func (si *SliceIterator[T]) IsClosed() bool {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	return si.closed
}

// ToSlice returns all remaining items as a slice.
// This is a convenience method for collecting all remaining items at once.
// After calling ToSlice(), the iterator will be at the end.
//
// If context is cancelled during iteration, returns partial results and the context error.
//
// Thread-safe: Can be called concurrently with other methods.
func (si *SliceIterator[T]) ToSlice(ctx context.Context) ([]T, error) {
	var result []T
	
	for {
		item, hasNext, err := si.Next(ctx)
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