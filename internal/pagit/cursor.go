package pagit

import (
	"context"
	"errors"
)

var (
	ErrInvalidPageSize = errors.New("page size must be between 1 and MaxPageSize")
	ErrInvalidCursor   = errors.New("invalid cursor value")
)

// PaginateCursor performs cursor-based pagination using a cursor provider
//
// This function provides cursor-based pagination which is ideal for:
// - Real-time data where items can be added/removed
// - Large datasets where offset-based pagination becomes inefficient
// - APIs that need consistent pagination even when data changes
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - provider: Cursor data provider that fetches paginated data
//   - config: Configuration for cursor pagination
//   - extractor: Function to extract cursor value from each item
//
// Returns:
//   - CursorResult with paginated data and navigation info
//
// Example usage:
//
//	// With slice data (in-memory)
//	sliceProvider := NewSliceCursorProvider(users, func(u User) int64 { return u.ID })
//	result := PaginateCursor(ctx, sliceProvider, config, extractor)
//
//	// With database provider
//	dbProvider := &UserCursorDBProvider{db: myDB}
//	result := PaginateCursor(ctx, dbProvider, config, extractor)
func PaginateCursor[T any, C comparable](
	ctx context.Context,
	provider CursorDataProvider[T, C],
	config CursorConfig[C],
	extractor CursorExtractor[T, C],
) (CursorResult[T, C], error) {
	// Validate page size
	if config.PageSize <= 0 || config.PageSize > MaxPageSize {
		return CursorResult[T, C]{}, ErrInvalidPageSize
	}

	var data []T
	var err error

	// Get data based on direction
	switch config.Direction {
	case CursorBackward:
		data, err = provider.GetDataBefore(ctx, config.Cursor, config.PageSize)
	default: // CursorForward or unspecified
		data, err = provider.GetDataAfter(ctx, config.Cursor, config.PageSize)
	}

	if err != nil {
		return CursorResult[T, C]{}, err
	}

	result := CursorResult[T, C]{
		Data:  data,
		Count: len(data),
	}

	// Set cursors from data if available
	if len(data) > 0 {
		firstCursor := extractor(data[0])
		lastCursor := extractor(data[len(data)-1])

		// Set navigation cursors
		if config.Direction == CursorBackward {
			result.NextCursor = &lastCursor
			if len(data) == config.PageSize {
				result.PrevCursor = &firstCursor
			}
		} else {
			result.NextCursor = &lastCursor
			if config.Cursor != nil {
				result.PrevCursor = &firstCursor
			}
		}
	}

	// Check for more data if provider supports it
	if checkProvider, ok := provider.(CursorCheckProvider[C]); ok {
		if len(data) > 0 {
			lastCursor := extractor(data[len(data)-1])
			firstCursor := extractor(data[0])

			hasNext, err := checkProvider.HasDataAfter(ctx, lastCursor)
			if err == nil {
				result.HasNext = hasNext
			}

			if config.Cursor != nil {
				hasPrev, err := checkProvider.HasDataBefore(ctx, firstCursor)
				if err == nil {
					result.HasPrev = hasPrev
				}
			}
		}
	} else {
		// Fallback: assume there might be more data if we got a full page
		result.HasNext = len(data) == config.PageSize
		result.HasPrev = config.Cursor != nil
	}

	return result, nil
}

