package pagit

import (
	"context"
	"slices"
)

// SliceProvider provides pagination for in-memory slice data
type SliceProvider[T any] struct {
	data []T
}

// NewSliceProvider creates a new slice provider for offset-based pagination
func NewSliceProvider[T any](data []T) *SliceProvider[T] {
	return &SliceProvider[T]{data: data}
}

// GetData implements DataProvider interface
func (p *SliceProvider[T]) GetData(ctx context.Context, offset, limit int) ([]T, error) {
	if offset >= len(p.data) {
		return []T{}, nil
	}

	endIndex := min(offset+limit, len(p.data))
	return p.data[offset:endIndex], nil
}

// GetTotalCount implements CountProvider interface
func (p *SliceProvider[T]) GetTotalCount(ctx context.Context) (int64, error) {
	return int64(len(p.data)), nil
}

// SliceCursorProvider provides cursor-based pagination for in-memory slice data
type SliceCursorProvider[T any, C comparable] struct {
	data      []T
	extractor CursorExtractor[T, C]
}

// NewSliceCursorProvider creates a new slice cursor provider
func NewSliceCursorProvider[T any, C comparable](
	data []T,
	extractor CursorExtractor[T, C],
) *SliceCursorProvider[T, C] {
	return &SliceCursorProvider[T, C]{
		data:      data,
		extractor: extractor,
	}
}

// GetDataAfter implements CursorDataProvider interface
func (p *SliceCursorProvider[T, C]) GetDataAfter(
	ctx context.Context,
	cursor *C,
	limit int,
) ([]T, error) {
	if cursor == nil {
		// Start from beginning
		endIndex := min(limit, len(p.data))
		return p.data[:endIndex], nil
	}

	// Find cursor position
	cursorIndex := p.findCursorIndex(*cursor)
	if cursorIndex == -1 {
		return []T{}, ErrInvalidCursor
	}

	// Return items after cursor
	startIndex := cursorIndex + 1
	if startIndex >= len(p.data) {
		return []T{}, nil
	}

	endIndex := min(startIndex+limit, len(p.data))
	return p.data[startIndex:endIndex], nil
}

// GetDataBefore implements CursorDataProvider interface
func (p *SliceCursorProvider[T, C]) GetDataBefore(
	ctx context.Context,
	cursor *C,
	limit int,
) ([]T, error) {
	if cursor == nil {
		// No data before the beginning
		return []T{}, nil
	}

	// Find cursor position
	cursorIndex := p.findCursorIndex(*cursor)
	if cursorIndex == -1 {
		return []T{}, ErrInvalidCursor
	}

	// Return items before cursor
	endIndex := cursorIndex
	if endIndex <= 0 {
		return []T{}, nil
	}

	startIndex := max(0, endIndex-limit)
	return p.data[startIndex:endIndex], nil
}

// HasDataAfter implements CursorCheckProvider interface
func (p *SliceCursorProvider[T, C]) HasDataAfter(
	ctx context.Context,
	cursor C,
) (bool, error) {
	cursorIndex := p.findCursorIndex(cursor)
	if cursorIndex == -1 {
		return false, ErrInvalidCursor
	}

	return cursorIndex+1 < len(p.data), nil
}

// HasDataBefore implements CursorCheckProvider interface
func (p *SliceCursorProvider[T, C]) HasDataBefore(
	ctx context.Context,
	cursor C,
) (bool, error) {
	cursorIndex := p.findCursorIndex(cursor)
	if cursorIndex == -1 {
		return false, ErrInvalidCursor
	}

	return cursorIndex > 0, nil
}

// findCursorIndex finds the index of item with matching cursor value
func (p *SliceCursorProvider[T, C]) findCursorIndex(cursor C) int {
	return slices.IndexFunc(p.data, func(item T) bool {
		return p.extractor(item) == cursor
	})
}
