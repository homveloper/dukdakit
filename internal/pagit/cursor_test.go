package pagit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestItem struct {
	ID   int64
	Name string
}

func TestPaginateCursor_FromStart(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)

	config := CursorConfig[int64]{
		PageSize:  2,
		Cursor:    nil, // Start from beginning
		Direction: CursorForward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, int64(1), result.Data[0].ID)
	assert.Equal(t, int64(2), result.Data[1].ID)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, int64(2), *result.NextCursor)
	assert.Equal(t, 2, result.Count)
}

func TestPaginateCursor_ForwardFromCursor(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)
	cursor := int64(2)

	config := CursorConfig[int64]{
		PageSize:  2,
		Cursor:    &cursor,
		Direction: CursorForward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, int64(3), result.Data[0].ID)
	assert.Equal(t, int64(4), result.Data[1].ID)
	assert.True(t, result.HasNext)
	assert.True(t, result.HasPrev)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, int64(4), *result.NextCursor)
	assert.Equal(t, 2, result.Count)
}

func TestPaginateCursor_BackwardFromCursor(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)
	cursor := int64(4)

	config := CursorConfig[int64]{
		PageSize:  2,
		Cursor:    &cursor,
		Direction: CursorBackward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, int64(2), result.Data[0].ID)
	assert.Equal(t, int64(3), result.Data[1].ID)
	assert.True(t, result.HasNext)
	assert.True(t, result.HasPrev)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, int64(3), *result.NextCursor)
	assert.Equal(t, 2, result.Count)
}

func TestPaginateCursor_EmptyData(t *testing.T) {
	var data []TestItem
	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)

	config := CursorConfig[int64]{
		PageSize:  10,
		Cursor:    nil,
		Direction: CursorForward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Empty(t, result.Data)
	assert.False(t, result.HasNext)
	assert.False(t, result.HasPrev)
	assert.Nil(t, result.NextCursor)
	assert.Equal(t, 0, result.Count)
}

func TestPaginateCursor_InvalidPageSize(t *testing.T) {
	data := []TestItem{{ID: 1, Name: "Item1"}}
	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)

	config := CursorConfig[int64]{
		PageSize:  0,
		Cursor:    nil,
		Direction: CursorForward,
	}

	_, err := PaginateCursor(context.Background(), provider, config, extractor)
	assert.Equal(t, ErrInvalidPageSize, err)
}

func TestPaginateCursor_InvalidCursor(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
	}
	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)
	invalidCursor := int64(999)

	config := CursorConfig[int64]{
		PageSize:  2,
		Cursor:    &invalidCursor,
		Direction: CursorForward,
	}

	_, err := PaginateCursor(context.Background(), provider, config, extractor)
	assert.Equal(t, ErrInvalidCursor, err)
}

func TestPaginateCursor_StringCursor(t *testing.T) {
	type StringItem struct {
		Name  string
		Value int
	}

	data := []StringItem{
		{Name: "alpha", Value: 1},
		{Name: "beta", Value: 2},
		{Name: "gamma", Value: 3},
	}

	extractor := func(item StringItem) string { return item.Name }
	provider := NewSliceCursorProvider(data, extractor)

	config := CursorConfig[string]{
		PageSize:  1,
		Cursor:    nil,
		Direction: CursorForward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "alpha", result.Data[0].Name)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, "alpha", *result.NextCursor)
}

func TestPaginateCursor_LastPage(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
	}

	extractor := func(item TestItem) int64 { return item.ID }
	provider := NewSliceCursorProvider(data, extractor)
	cursor := int64(2)

	config := CursorConfig[int64]{
		PageSize:  2,
		Cursor:    &cursor,
		Direction: CursorForward,
	}

	result, err := PaginateCursor(context.Background(), provider, config, extractor)

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, int64(3), result.Data[0].ID)
	assert.False(t, result.HasNext)
	assert.True(t, result.HasPrev)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, 1, result.Count)
}
