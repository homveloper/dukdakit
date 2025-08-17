package pagit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginateOffset_BasicPagination(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     1,
		PageSize: 2,
	}

	result, err := PaginateOffset(context.Background(), provider, config)

	require.NoError(t, err)
	
	// Should return OffsetResult since SliceProvider supports counting
	offsetResult, ok := result.(OffsetResult[TestItem])
	require.True(t, ok, "Should return OffsetResult")
	
	assert.Len(t, offsetResult.Data, 2)
	assert.Equal(t, int64(1), offsetResult.Data[0].ID)
	assert.Equal(t, int64(2), offsetResult.Data[1].ID)
	assert.Equal(t, int64(5), offsetResult.TotalCount)
	assert.Equal(t, 1, offsetResult.Page)
	assert.Equal(t, 2, offsetResult.PageSize)
	assert.Equal(t, 3, offsetResult.TotalPages)
	assert.True(t, offsetResult.HasNext)
	assert.False(t, offsetResult.HasPrev)
	assert.Equal(t, 0, offsetResult.Offset)
	assert.Equal(t, 2, offsetResult.Count)
}

func TestPaginateOffset_SecondPage(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     2,
		PageSize: 2,
	}

	result, err := PaginateOffset(context.Background(), provider, config)

	require.NoError(t, err)
	
	offsetResult, ok := result.(OffsetResult[TestItem])
	require.True(t, ok)
	
	assert.Len(t, offsetResult.Data, 2)
	assert.Equal(t, int64(3), offsetResult.Data[0].ID)
	assert.Equal(t, int64(4), offsetResult.Data[1].ID)
	assert.Equal(t, int64(5), offsetResult.TotalCount)
	assert.Equal(t, 2, offsetResult.Page)
	assert.Equal(t, 3, offsetResult.TotalPages)
	assert.True(t, offsetResult.HasNext)
	assert.True(t, offsetResult.HasPrev)
	assert.Equal(t, 2, offsetResult.Offset)
}

func TestPaginateOffset_LastPage(t *testing.T) {
	data := []TestItem{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
		{ID: 3, Name: "Item3"},
		{ID: 4, Name: "Item4"},
		{ID: 5, Name: "Item5"},
	}

	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     3,
		PageSize: 2,
	}

	result, err := PaginateOffset(context.Background(), provider, config)

	require.NoError(t, err)
	
	offsetResult, ok := result.(OffsetResult[TestItem])
	require.True(t, ok)
	
	assert.Len(t, offsetResult.Data, 1)
	assert.Equal(t, int64(5), offsetResult.Data[0].ID)
	assert.Equal(t, int64(5), offsetResult.TotalCount)
	assert.Equal(t, 3, offsetResult.Page)
	assert.Equal(t, 3, offsetResult.TotalPages)
	assert.False(t, offsetResult.HasNext)
	assert.True(t, offsetResult.HasPrev)
	assert.Equal(t, 4, offsetResult.Offset)
	assert.Equal(t, 1, offsetResult.Count)
}

func TestPaginateOffset_EmptyData(t *testing.T) {
	var data []TestItem
	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     1,
		PageSize: 10,
	}

	result, err := PaginateOffset(context.Background(), provider, config)

	require.NoError(t, err)
	
	offsetResult, ok := result.(OffsetResult[TestItem])
	require.True(t, ok)
	
	assert.Empty(t, offsetResult.Data)
	assert.Equal(t, int64(0), offsetResult.TotalCount)
	assert.Equal(t, 1, offsetResult.Page)
	assert.Equal(t, 0, offsetResult.TotalPages)
	assert.False(t, offsetResult.HasNext)
	assert.False(t, offsetResult.HasPrev)
	assert.Equal(t, 0, offsetResult.Count)
}

func TestPaginateOffset_InvalidPage(t *testing.T) {
	data := []TestItem{{ID: 1, Name: "Item1"}}
	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     0,
		PageSize: 10,
	}

	_, err := PaginateOffset(context.Background(), provider, config)
	assert.Equal(t, ErrInvalidPage, err)
}

func TestPaginateOffset_InvalidPageSize(t *testing.T) {
	data := []TestItem{{ID: 1, Name: "Item1"}}
	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     1,
		PageSize: 0,
	}

	_, err := PaginateOffset(context.Background(), provider, config)
	assert.Equal(t, ErrInvalidPageSize, err)
}

func TestPaginateOffset_BeyondData(t *testing.T) {
	data := []TestItem{{ID: 1, Name: "Item1"}}
	provider := NewSliceProvider(data)

	config := OffsetConfig{
		Page:     10, // Way beyond available data
		PageSize: 10,
	}

	result, err := PaginateOffset(context.Background(), provider, config)

	require.NoError(t, err)
	
	offsetResult, ok := result.(OffsetResult[TestItem])
	require.True(t, ok)
	
	assert.Empty(t, offsetResult.Data)
	assert.Equal(t, int64(1), offsetResult.TotalCount)
	assert.Equal(t, 10, offsetResult.Page)
	assert.Equal(t, 1, offsetResult.TotalPages)
	assert.False(t, offsetResult.HasNext)
	assert.True(t, offsetResult.HasPrev)
	assert.Equal(t, 90, offsetResult.Offset)
	assert.Equal(t, 0, offsetResult.Count)
}