package dukdakit

import (
	"context"

	"github.com/homveloper/dukdakit/internal/pagit"
)

// PagitCategory provides pagination utilities for game development
type PagitCategory struct{}

// Pagit is the global instance for pagination features
var Pagit = &PagitCategory{}

// PaginateCursor performs cursor-based pagination using a cursor provider
//
// This function provides cursor-based pagination which is ideal for:
// - Real-time data where items can be added/removed (game logs, leaderboards)
// - Large datasets where offset-based pagination becomes inefficient
// - APIs that need consistent pagination even when data changes
// - Social features like friend lists, guild member lists
//
// Example usage:
//
//	// In-memory data
//	users := []User{{ID: 1, Name: "Player1"}, {ID: 2, Name: "Player2"}}
//	provider := dukdakit.Pagit.NewSliceCursorProvider(users, func(u User) int64 { return u.ID })
//	
//	config := dukdakit.CursorConfig[int64]{
//	    PageSize: 20,
//	    Cursor: nil, // Start from beginning
//	    Direction: dukdakit.CursorForward,
//	}
//	
//	result, err := dukdakit.PaginateCursor(ctx, provider, config, 
//	    func(u User) int64 { return u.ID })
//
//	// Database provider example
//	dbProvider := &UserCursorDBProvider{db: gameDB}
//	result, err = dukdakit.PaginateCursor(ctx, dbProvider, config, extractor)
func PaginateCursor[T any, C comparable](
	ctx context.Context,
	provider pagit.CursorDataProvider[T, C],
	config pagit.CursorConfig[C],
	extractor pagit.CursorExtractor[T, C],
) (pagit.CursorResult[T, C], error) {
	return pagit.PaginateCursor(ctx, provider, config, extractor)
}

// PaginateOffset performs offset-based pagination using a data provider
//
// This function provides traditional offset-based pagination which is ideal for:
// - Admin interfaces and management tools
// - Reports and analytics dashboards
// - Static datasets like game items, achievements
// - Use cases requiring page numbers and total counts
//
// Returns either OffsetResult (with total count) or OffsetResultWithoutTotal
// depending on whether the provider supports counting.
//
// Example usage:
//
//	// In-memory data
//	products := []Product{{ID: 1, Name: "Sword"}, {ID: 2, Name: "Shield"}}
//	provider := dukdakit.Pagit.NewSliceProvider(products)
//	
//	config := dukdakit.OffsetConfig{
//	    Page:     1,
//	    PageSize: 10,
//	}
//	
//	result, err := dukdakit.PaginateOffset(ctx, provider, config)
//
//	// Database provider example
//	dbProvider := &ItemDBProvider{db: gameDB}
//	result, err = dukdakit.PaginateOffset(ctx, dbProvider, config)
func PaginateOffset[T any](
	ctx context.Context,
	provider pagit.DataProvider[T],
	config pagit.OffsetConfig,
) (any, error) {
	return pagit.PaginateOffset(ctx, provider, config)
}

// NewSliceProvider creates a provider for in-memory slice data (offset-based)
func NewSliceProvider[T any](data []T) *pagit.SliceProvider[T] {
	return pagit.NewSliceProvider(data)
}

// NewSliceCursorProvider creates a provider for in-memory slice data (cursor-based)
func NewSliceCursorProvider[T any, C comparable](
	data []T,
	extractor pagit.CursorExtractor[T, C],
) *pagit.SliceCursorProvider[T, C] {
	return pagit.NewSliceCursorProvider(data, extractor)
}

// Type aliases for easier usage (non-generic types only for Go 1.21 compatibility)
type (
	// OffsetConfig holds configuration for offset-based pagination
	OffsetConfig = pagit.OffsetConfig
	
	// CountProvider provides total count for offset-based pagination
	CountProvider = pagit.CountProvider
	
	// CursorDirection specifies the direction of cursor pagination
	CursorDirection = pagit.CursorDirection
)

// Constants for cursor direction
const (
	CursorForward  = pagit.CursorForward
	CursorBackward = pagit.CursorBackward
)

// Default constants
const (
	DefaultCursorPageSize = pagit.DefaultCursorPageSize
	DefaultOffsetPageSize = pagit.DefaultOffsetPageSize
	MaxPageSize          = pagit.MaxPageSize
)

// Configuration builders for common patterns

// OffsetFromPage creates an OffsetConfig with default page size
func OffsetFromPage(page int) OffsetConfig {
	return pagit.OffsetFromPage(page)
}

// OffsetFromPageSize creates an OffsetConfig with default page 1
func OffsetFromPageSize(pageSize int) OffsetConfig {
	return pagit.OffsetFromPageSize(pageSize)
}

// NewOffsetConfig creates an OffsetConfig with validation
func NewOffsetConfig(page, pageSize int) OffsetConfig {
	return pagit.NewOffsetConfig(page, pageSize)
}