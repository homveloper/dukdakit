package pagit

// CursorExtractor extracts cursor value from an item
// The domain defines how to extract the cursor from their data
type CursorExtractor[T any, C comparable] func(item T) C

// CursorResult represents the result of cursor-based pagination
type CursorResult[T any, C comparable] struct {
	// Data contains the paginated items
	Data []T
	
	// HasNext indicates if there are more items after this page
	HasNext bool
	
	// HasPrev indicates if there are items before this page  
	HasPrev bool
	
	// NextCursor is the cursor to use for the next page
	// Only set if HasNext is true
	NextCursor *C
	
	// PrevCursor is the cursor to use for the previous page
	// Only set if HasPrev is true
	PrevCursor *C
	
	// Count is the number of items in this page
	Count int
}

// OffsetResult represents the result of offset-based pagination
type OffsetResult[T any] struct {
	// Data contains the paginated items
	Data []T
	
	// TotalCount is the total number of items available
	TotalCount int64
	
	// Page is the current page number (1-based)
	Page int
	
	// PageSize is the number of items per page
	PageSize int
	
	// TotalPages is the total number of pages
	TotalPages int
	
	// HasNext indicates if there are more pages
	HasNext bool
	
	// HasPrev indicates if there are previous pages
	HasPrev bool
	
	// Offset is the starting index of items in this page (0-based)
	Offset int
	
	// Count is the number of items in this page
	Count int
}

// CursorConfig holds configuration for cursor-based pagination
type CursorConfig[C comparable] struct {
	// PageSize is the number of items per page
	PageSize int
	
	// Cursor is the starting cursor position
	// nil means start from the beginning
	Cursor *C
	
	// Direction specifies pagination direction
	Direction CursorDirection
}

// OffsetConfig holds configuration for offset-based pagination
type OffsetConfig struct {
	// Page is the page number (1-based)
	Page int
	
	// PageSize is the number of items per page
	PageSize int
}

// CursorDirection specifies the direction of cursor pagination
type CursorDirection int

const (
	// CursorForward paginates forward from the cursor
	CursorForward CursorDirection = iota
	// CursorBackward paginates backward from the cursor
	CursorBackward
)

// DefaultCursorPageSize is the default page size for cursor pagination
const DefaultCursorPageSize = 20

// DefaultOffsetPageSize is the default page size for offset pagination
const DefaultOffsetPageSize = 20

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 1000