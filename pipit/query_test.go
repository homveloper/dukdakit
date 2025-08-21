package pipit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuery_TypeSystemBasics tests the core type system components
func TestQuery_TypeSystemBasics(t *testing.T) {
	t.Run("Query struct initialization", func(t *testing.T) {
		query := &Query[int]{
			source:   nil,
			pipeline: make([]Operation, 0),
			ctx:      context.Background(),
			err:      nil,
		}
		
		assert.NotNil(t, query)
		assert.NotNil(t, query.ctx)
		assert.Nil(t, query.err)
		assert.Equal(t, 0, len(query.pipeline))
	})
	
	t.Run("Generic type safety", func(t *testing.T) {
		// Test different generic types
		stringQuery := &Query[string]{}
		intQuery := &Query[int]{}
		structQuery := &Query[struct{ Name string }]{}
		
		assert.IsType(t, &Query[string]{}, stringQuery)
		assert.IsType(t, &Query[int]{}, intQuery)
		assert.IsType(t, &Query[struct{ Name string }]{}, structQuery)
	})
}

// TestOperationType_String tests the string representation of operation types
func TestOperationType_String(t *testing.T) {
	tests := []struct {
		opType   OperationType
		expected string
	}{
		{FilterOp, "Filter"},
		{MapOp, "Map"},
		{TakeOp, "Take"},
		{DropOp, "Drop"},
		{DistinctOp, "Distinct"},
		{OperationType(999), "Unknown(999)"},
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("OperationType_%s", tt.expected), func(t *testing.T) {
			result := tt.opType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOperationType_Constants tests that operation type constants are properly defined
func TestOperationType_Constants(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, OperationType(0), FilterOp)
	assert.Equal(t, OperationType(1), MapOp)
	assert.Equal(t, OperationType(2), TakeOp)
	assert.Equal(t, OperationType(3), DropOp)
	assert.Equal(t, OperationType(4), DistinctOp)
	
	// Test that they are distinct
	allOps := []OperationType{FilterOp, MapOp, TakeOp, DropOp, DistinctOp}
	for i := 0; i < len(allOps); i++ {
		for j := i + 1; j < len(allOps); j++ {
			assert.NotEqual(t, allOps[i], allOps[j], 
				"Operation types %s and %s should be different", 
				allOps[i].String(), allOps[j].String())
		}
	}
}

// MockIterator is a test implementation of Iterator interface
type MockIterator[T any] struct {
	data    []T
	index   int
	closed  bool
	failAt  int // Fail at this index (-1 for no failure)
}

func NewMockIterator[T any](data []T) *MockIterator[T] {
	return &MockIterator[T]{
		data:   data,
		index:  0,
		failAt: -1,
	}
}

func (m *MockIterator[T]) SetFailAt(index int) {
	m.failAt = index
}

func (m *MockIterator[T]) Next(ctx context.Context) (T, bool, error) {
	var zero T
	
	// Check context cancellation
	select {
	case <-ctx.Done():
		return zero, false, ctx.Err()
	default:
	}
	
	// Check if closed
	if m.closed {
		return zero, false, fmt.Errorf("iterator is closed")
	}
	
	// Check for artificial failure
	if m.failAt >= 0 && m.index == m.failAt {
		return zero, false, fmt.Errorf("artificial failure at index %d", m.index)
	}
	
	// Check bounds
	if m.index >= len(m.data) {
		return zero, false, nil
	}
	
	item := m.data[m.index]
	m.index++
	return item, true, nil
}

func (m *MockIterator[T]) HasNext() bool {
	return !m.closed && m.index < len(m.data)
}

func (m *MockIterator[T]) Close() error {
	m.closed = true
	return nil
}

// TestIterator_Interface tests the Iterator interface implementation
func TestIterator_Interface(t *testing.T) {
	t.Run("MockIterator basic functionality", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		iter := NewMockIterator(data)
		
		ctx := context.Background()
		
		// Test iteration
		for i, expected := range data {
			assert.True(t, iter.HasNext(), "HasNext should return true at index %d", i)
			
			item, hasNext, err := iter.Next(ctx)
			require.NoError(t, err)
			assert.Equal(t, expected, item)
			assert.True(t, hasNext)
		}
		
		// Test end of iteration
		assert.False(t, iter.HasNext())
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, item) // zero value for int
		assert.False(t, hasNext)
	})
	
	t.Run("Iterator context cancellation", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		iter := NewMockIterator(data)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		item, hasNext, err := iter.Next(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, 0, item)
		assert.False(t, hasNext)
	})
	
	t.Run("Iterator timeout", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewMockIterator(data)
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		// First call should work
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, item)
		assert.True(t, hasNext)
		
		// Wait for timeout
		time.Sleep(2 * time.Millisecond)
		
		// Second call should timeout
		item, hasNext, err = iter.Next(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Equal(t, 0, item)
		assert.False(t, hasNext)
	})
	
	t.Run("Iterator error handling", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewMockIterator(data)
		iter.SetFailAt(1) // Fail at second element
		
		ctx := context.Background()
		
		// First element should work
		item, hasNext, err := iter.Next(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, item)
		assert.True(t, hasNext)
		
		// Second element should fail
		item, hasNext, err = iter.Next(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "artificial failure at index 1")
		assert.Equal(t, 0, item)
		assert.False(t, hasNext)
	})
	
	t.Run("Iterator close functionality", func(t *testing.T) {
		data := []int{1, 2, 3}
		iter := NewMockIterator(data)
		
		ctx := context.Background()
		
		// Should work before close
		assert.True(t, iter.HasNext())
		
		// Close iterator
		err := iter.Close()
		require.NoError(t, err)
		
		// Should not work after close
		assert.False(t, iter.HasNext())
		
		item, hasNext, err := iter.Next(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator is closed")
		assert.Equal(t, 0, item)
		assert.False(t, hasNext)
	})
}

// MockOperation is a test implementation of Operation interface
type MockOperation struct {
	opType      OperationType
	transformer func(any) any
	shouldError bool
}

func NewMockOperation(opType OperationType, transformer func(any) any) *MockOperation {
	return &MockOperation{
		opType:      opType,
		transformer: transformer,
		shouldError: false,
	}
}

func (m *MockOperation) SetShouldError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockOperation) Apply(ctx context.Context, item any) (any, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	if m.shouldError {
		return nil, fmt.Errorf("mock operation error")
	}
	
	if m.transformer != nil {
		return m.transformer(item), nil
	}
	
	return item, nil
}

func (m *MockOperation) Type() OperationType {
	return m.opType
}

// TestOperation_Interface tests the Operation interface implementation
func TestOperation_Interface(t *testing.T) {
	t.Run("MockOperation basic functionality", func(t *testing.T) {
		// Create a doubling operation
		op := NewMockOperation(MapOp, func(item any) any {
			if i, ok := item.(int); ok {
				return i * 2
			}
			return item
		})
		
		ctx := context.Background()
		
		// Test transformation
		result, err := op.Apply(ctx, 5)
		require.NoError(t, err)
		assert.Equal(t, 10, result)
		
		// Test operation type
		assert.Equal(t, MapOp, op.Type())
	})
	
	t.Run("Operation context cancellation", func(t *testing.T) {
		op := NewMockOperation(FilterOp, nil)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		result, err := op.Apply(ctx, 42)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, result)
	})
	
	t.Run("Operation error handling", func(t *testing.T) {
		op := NewMockOperation(FilterOp, nil)
		op.SetShouldError(true)
		
		ctx := context.Background()
		
		result, err := op.Apply(ctx, 42)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock operation error")
		assert.Nil(t, result)
	})
}

// BenchmarkQuery_TypeSystemOverhead benchmarks the overhead of the type system
func BenchmarkQuery_TypeSystemOverhead(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	
	b.Run("Query struct creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &Query[int]{
				source:   NewMockIterator(data),
				pipeline: make([]Operation, 0),
				ctx:      context.Background(),
				err:      nil,
			}
		}
	})
	
	b.Run("OperationType string conversion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FilterOp.String()
		}
	})
}