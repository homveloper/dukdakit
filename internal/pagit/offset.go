package pagit

import (
	"context"
	"errors"
)

var (
	ErrInvalidPage = errors.New("page must be greater than 0")
)

// PaginateOffset performs offset-based pagination using a data provider
//
// This function provides traditional offset-based pagination which is ideal for:
// - Static datasets where items rarely change
// - Use cases requiring page numbers and total counts
// - Admin interfaces and reports
// - Simple pagination with jump-to-page functionality
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - provider: Data provider that fetches paginated data
//   - config: Configuration for offset pagination
//
// Returns:
//   - OffsetResult if total count is available
//   - OffsetResultWithoutTotal if total count is unavailable
//
// Example usage:
//
//	// With slice data (in-memory)
//	sliceProvider := NewSliceProvider(products)
//	result := PaginateOffset(ctx, sliceProvider, OffsetConfig{Page: 1, PageSize: 10})
//
//	// With database provider
//	dbProvider := &UserDBProvider{db: myDB}
//	result := PaginateOffset(ctx, dbProvider, OffsetConfig{Page: 1, PageSize: 10})
func PaginateOffset[T any](
	ctx context.Context,
	provider DataProvider[T],
	config OffsetConfig,
) (any, error) {
	// Validate input
	if config.Page <= 0 {
		return nil, ErrInvalidPage
	}

	if config.PageSize <= 0 || config.PageSize > MaxPageSize {
		return nil, ErrInvalidPageSize
	}

	// Calculate offset
	offset := (config.Page - 1) * config.PageSize

	// Get page data
	data, err := provider.GetData(ctx, offset, config.PageSize)
	if err != nil {
		return nil, err
	}

	// Try to get total count if provider supports it
	if countProvider, ok := provider.(CountProvider); ok {
		totalCount, err := countProvider.GetTotalCount(ctx)
		if err != nil && !errors.Is(err, ErrTotalCountUnavailable) {
			return nil, err
		}

		// If total count is available, return full result
		if err == nil && totalCount >= 0 {
			return buildOffsetResult(data, totalCount, config, offset), nil
		}
	}

	// Return result without total count
	return buildOffsetResultWithoutTotal(data, config, offset), nil
}

// OffsetFromPageSize creates an OffsetConfig with default page 1
func OffsetFromPageSize(pageSize int) OffsetConfig {
	if pageSize <= 0 {
		pageSize = DefaultOffsetPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return OffsetConfig{
		Page:     1,
		PageSize: pageSize,
	}
}

// OffsetFromPage creates an OffsetConfig with default page size
func OffsetFromPage(page int) OffsetConfig {
	if page <= 0 {
		page = 1
	}

	return OffsetConfig{
		Page:     page,
		PageSize: DefaultOffsetPageSize,
	}
}

// NewOffsetConfig creates an OffsetConfig with validation
func NewOffsetConfig(page, pageSize int) OffsetConfig {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultOffsetPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return OffsetConfig{
		Page:     page,
		PageSize: pageSize,
	}
}

// calculateTotalPages calculates the total number of pages
func calculateTotalPages(totalCount int64, pageSize int) int {
	if totalCount == 0 || pageSize <= 0 {
		return 0
	}

	pages := int(totalCount) / pageSize
	if int(totalCount)%pageSize != 0 {
		pages++
	}

	return pages
}

// GetPageInfo returns pagination metadata without data
func GetPageInfo(totalCount int64, config OffsetConfig) (OffsetResult[any], error) {
	if config.Page <= 0 {
		return OffsetResult[any]{}, ErrInvalidPage
	}

	if config.PageSize <= 0 || config.PageSize > MaxPageSize {
		return OffsetResult[any]{}, ErrInvalidPageSize
	}

	totalPages := calculateTotalPages(totalCount, config.PageSize)
	offset := (config.Page - 1) * config.PageSize

	return OffsetResult[any]{
		Data:       nil, // No data, just metadata
		TotalCount: totalCount,
		Page:       config.Page,
		PageSize:   config.PageSize,
		TotalPages: totalPages,
		HasNext:    config.Page < totalPages,
		HasPrev:    config.Page > 1,
		Offset:     offset,
		Count:      0, // No data provided
	}, nil
}
