package pagit

import "context"

// DataProvider provides data for pagination without requiring all data in memory
// This interface allows integration with databases, APIs, or other data sources
type DataProvider[T any] interface {
	// GetData retrieves a page of data starting from offset with specified limit
	GetData(ctx context.Context, offset, limit int) ([]T, error)
}

// CountProvider provides total count for offset-based pagination
// Separated from DataProvider to handle cases where counting is expensive or impossible
type CountProvider interface {
	// GetTotalCount returns the total number of items
	// Returns -1 if count is unknown or too expensive to calculate
	GetTotalCount(ctx context.Context) (int64, error)
}

// CursorDataProvider provides data for cursor-based pagination
type CursorDataProvider[T any, C comparable] interface {
	// GetDataAfter retrieves items after the given cursor
	GetDataAfter(ctx context.Context, cursor *C, limit int) ([]T, error)
	
	// GetDataBefore retrieves items before the given cursor
	GetDataBefore(ctx context.Context, cursor *C, limit int) ([]T, error)
}

// CursorCheckProvider checks if cursors exist for navigation
type CursorCheckProvider[C comparable] interface {
	// HasDataAfter checks if there are items after the given cursor
	HasDataAfter(ctx context.Context, cursor C) (bool, error)
	
	// HasDataBefore checks if there are items before the given cursor
	HasDataBefore(ctx context.Context, cursor C) (bool, error)
}

// OptionalCountProvider is a combined interface for providers that may support counting
type OptionalCountProvider[T any] interface {
	DataProvider[T]
	CountProvider
}

// FullCursorProvider is a combined interface for complete cursor-based pagination
type FullCursorProvider[T any, C comparable] interface {
	CursorDataProvider[T, C]
	CursorCheckProvider[C]
}

// OffsetResultWithoutTotal represents offset pagination result when total count is unknown
type OffsetResultWithoutTotal[T any] struct {
	// Data contains the paginated items
	Data []T
	
	// Page is the current page number (1-based)
	Page int
	
	// PageSize is the number of items per page
	PageSize int
	
	// HasNext indicates if there are more pages (if determinable)
	// nil means unknown
	HasNext *bool
	
	// HasPrev indicates if there are previous pages
	HasPrev bool
	
	// Offset is the starting index of items in this page (0-based)
	Offset int
	
	// Count is the number of items in this page
	Count int
	
	// EstimatedTotal is an estimated total count if available
	// -1 means unknown
	EstimatedTotal int64
}