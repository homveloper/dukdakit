package pagit

import "errors"

var (
	ErrProviderNotImplemented = errors.New("provider method not implemented")
	ErrTotalCountUnavailable  = errors.New("total count is not available")
)

// buildOffsetResult creates OffsetResult with total count
func buildOffsetResult[T any](
	data []T,
	totalCount int64,
	config OffsetConfig,
	offset int,
) OffsetResult[T] {
	totalPages := calculateTotalPages(totalCount, config.PageSize)

	return OffsetResult[T]{
		Data:       data,
		TotalCount: totalCount,
		Page:       config.Page,
		PageSize:   config.PageSize,
		TotalPages: totalPages,
		HasNext:    config.Page < totalPages,
		HasPrev:    config.Page > 1,
		Offset:     offset,
		Count:      len(data),
	}
}

// buildOffsetResultWithoutTotal creates OffsetResultWithoutTotal
func buildOffsetResultWithoutTotal[T any](
	data []T,
	config OffsetConfig,
	offset int,
) OffsetResultWithoutTotal[T] {
	// Heuristic: if we got less than requested, likely no more data
	var hasNext *bool
	if len(data) < config.PageSize {
		hasNextVal := false
		hasNext = &hasNextVal
	} else {
		// We got a full page, there might be more
		hasNextVal := true
		hasNext = &hasNextVal
	}

	return OffsetResultWithoutTotal[T]{
		Data:           data,
		Page:           config.Page,
		PageSize:       config.PageSize,
		HasNext:        hasNext,
		HasPrev:        config.Page > 1,
		Offset:         offset,
		Count:          len(data),
		EstimatedTotal: -1, // Unknown
	}
}