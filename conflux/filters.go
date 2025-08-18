package conflux

import (
	"fmt"
	"time"
)

// ============================================================================
// 공통 필터 인터페이스
// ============================================================================

// Filter 모든 필터가 구현해야 하는 기본 인터페이스
type Filter interface {
	// Validate 필터가 유효한지 검증
	Validate() error
}

// ============================================================================
// 범용 맵 기반 필터 (NoSQL에 적합)
// ============================================================================

// MapFilter 키-값 쌍 기반 필터 (MongoDB, Redis 등에 적합)
type MapFilter map[string]any

// Validate MapFilter 유효성 검증
func (f MapFilter) Validate() error {
	if len(f) == 0 {
		return fmt.Errorf("filter cannot be empty")
	}
	return nil
}

// And 조건 추가 (AND 연산)
func (f MapFilter) And(key string, value any) MapFilter {
	f[key] = value
	return f
}

// NewMapFilter 새 MapFilter 생성
func NewMapFilter() MapFilter {
	return make(MapFilter)
}

// ============================================================================
// SQL 기반 필터 (PostgreSQL, MySQL 등에 적합)
// ============================================================================

// SQLFilter SQL WHERE 절 기반 필터
type SQLFilter struct {
	Query string        // WHERE 절 쿼리 (예: "email = ? AND status = ?")
	Args  []interface{} // 쿼리 파라미터
}

// Validate SQLFilter 유효성 검증
func (f *SQLFilter) Validate() error {
	if f.Query == "" {
		return fmt.Errorf("SQL query cannot be empty")
	}
	return nil
}

// NewSQLFilter 새 SQLFilter 생성
func NewSQLFilter(query string, args ...interface{}) *SQLFilter {
	return &SQLFilter{
		Query: query,
		Args:  args,
	}
}

// And WHERE 조건 추가
func (f *SQLFilter) And(condition string, args ...interface{}) *SQLFilter {
	if f.Query != "" {
		f.Query += " AND "
	}
	f.Query += condition
	f.Args = append(f.Args, args...)
	return f
}

// Or WHERE 조건 추가 (OR 연산)
func (f *SQLFilter) Or(condition string, args ...interface{}) *SQLFilter {
	if f.Query != "" {
		f.Query += " OR "
	}
	f.Query += condition
	f.Args = append(f.Args, args...)
	return f
}

// ============================================================================
// Redis 키 패턴 필터
// ============================================================================

// RedisFilter Redis 키 패턴 기반 필터
type RedisFilter struct {
	Pattern   string // 키 패턴 (예: "user:*", "session:123:*")
	KeyPrefix string // 키 접두사
	Suffix    string // 키 접미사
}

// Validate RedisFilter 유효성 검증
func (f *RedisFilter) Validate() error {
	if f.Pattern == "" && f.KeyPrefix == "" {
		return fmt.Errorf("Redis filter must have pattern or key prefix")
	}
	return nil
}

// NewRedisFilter 새 RedisFilter 생성
func NewRedisFilter(pattern string) *RedisFilter {
	return &RedisFilter{Pattern: pattern}
}

// NewRedisFilterWithPrefix 접두사 기반 RedisFilter 생성
func NewRedisFilterWithPrefix(prefix string) *RedisFilter {
	return &RedisFilter{KeyPrefix: prefix}
}

// ============================================================================
// 복합 필터 (여러 조건 조합)
// ============================================================================

// CompositeFilter 여러 필터를 조합한 복합 필터
type CompositeFilter struct {
	Filters   []Filter // 하위 필터들
	Operator  string   // 연산자 ("AND", "OR")
	Negated   bool     // NOT 연산 여부
}

// Validate CompositeFilter 유효성 검증
func (f *CompositeFilter) Validate() error {
	if len(f.Filters) == 0 {
		return fmt.Errorf("composite filter must have at least one filter")
	}
	
	for _, filter := range f.Filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("invalid sub-filter: %w", err)
		}
	}
	
	return nil
}

// NewCompositeFilter 새 CompositeFilter 생성
func NewCompositeFilter(operator string) *CompositeFilter {
	return &CompositeFilter{
		Operator: operator,
		Filters:  make([]Filter, 0),
	}
}

// Add 필터 추가
func (f *CompositeFilter) Add(filter Filter) *CompositeFilter {
	f.Filters = append(f.Filters, filter)
	return f
}

// Not NOT 연산 설정
func (f *CompositeFilter) Not() *CompositeFilter {
	f.Negated = true
	return f
}

// ============================================================================
// 시간 범위 필터
// ============================================================================

// TimeRangeFilter 시간 범위 기반 필터
type TimeRangeFilter struct {
	Field string     // 시간 필드명
	Start *time.Time // 시작 시간 (nil이면 제한 없음)
	End   *time.Time // 종료 시간 (nil이면 제한 없음)
}

// Validate TimeRangeFilter 유효성 검증
func (f *TimeRangeFilter) Validate() error {
	if f.Field == "" {
		return fmt.Errorf("time field name cannot be empty")
	}
	
	if f.Start != nil && f.End != nil && f.Start.After(*f.End) {
		return fmt.Errorf("start time cannot be after end time")
	}
	
	return nil
}

// NewTimeRangeFilter 새 TimeRangeFilter 생성
func NewTimeRangeFilter(field string) *TimeRangeFilter {
	return &TimeRangeFilter{Field: field}
}

// After 시작 시간 설정
func (f *TimeRangeFilter) After(start time.Time) *TimeRangeFilter {
	f.Start = &start
	return f
}

// Before 종료 시간 설정
func (f *TimeRangeFilter) Before(end time.Time) *TimeRangeFilter {
	f.End = &end
	return f
}

// Between 시간 범위 설정
func (f *TimeRangeFilter) Between(start, end time.Time) *TimeRangeFilter {
	f.Start = &start
	f.End = &end
	return f
}

// ============================================================================
// 페이지네이션 필터
// ============================================================================

// PaginationFilter 페이지네이션을 포함한 필터
type PaginationFilter struct {
	BaseFilter Filter // 기본 필터
	Limit      int    // 결과 개수 제한
	Offset     int    // 건너뛸 개수 (offset 기반)
	Cursor     string // 커서 (cursor 기반)
	SortBy     string // 정렬 기준 필드
	Ascending  bool   // 오름차순 여부
}

// Validate PaginationFilter 유효성 검증
func (f *PaginationFilter) Validate() error {
	if f.BaseFilter != nil {
		if err := f.BaseFilter.Validate(); err != nil {
			return fmt.Errorf("invalid base filter: %w", err)
		}
	}
	
	if f.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	
	if f.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	
	return nil
}

// NewPaginationFilter 새 PaginationFilter 생성
func NewPaginationFilter(baseFilter Filter) *PaginationFilter {
	return &PaginationFilter{
		BaseFilter: baseFilter,
		Limit:      50, // 기본값
		Ascending:  true,
	}
}

// WithLimit 결과 개수 제한 설정
func (f *PaginationFilter) WithLimit(limit int) *PaginationFilter {
	f.Limit = limit
	return f
}

// WithOffset 오프셋 설정
func (f *PaginationFilter) WithOffset(offset int) *PaginationFilter {
	f.Offset = offset
	return f
}

// WithCursor 커서 설정
func (f *PaginationFilter) WithCursor(cursor string) *PaginationFilter {
	f.Cursor = cursor
	return f
}

// WithSort 정렬 설정
func (f *PaginationFilter) WithSort(field string, ascending bool) *PaginationFilter {
	f.SortBy = field
	f.Ascending = ascending
	return f
}

// ============================================================================
// 필터 빌더 유틸리티
// ============================================================================

// FilterBuilder 다양한 필터 생성을 도와주는 빌더
type FilterBuilder struct{}

// NewFilterBuilder 새 FilterBuilder 생성
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

// Map MapFilter 생성
func (b *FilterBuilder) Map() MapFilter {
	return NewMapFilter()
}

// SQL SQLFilter 생성
func (b *FilterBuilder) SQL(query string, args ...interface{}) *SQLFilter {
	return NewSQLFilter(query, args...)
}

// Redis RedisFilter 생성
func (b *FilterBuilder) Redis(pattern string) *RedisFilter {
	return NewRedisFilter(pattern)
}

// TimeRange TimeRangeFilter 생성
func (b *FilterBuilder) TimeRange(field string) *TimeRangeFilter {
	return NewTimeRangeFilter(field)
}

// And AND 복합 필터 생성
func (b *FilterBuilder) And() *CompositeFilter {
	return NewCompositeFilter("AND")
}

// Or OR 복합 필터 생성
func (b *FilterBuilder) Or() *CompositeFilter {
	return NewCompositeFilter("OR")
}

// Paginated 페이지네이션 필터 생성
func (b *FilterBuilder) Paginated(baseFilter Filter) *PaginationFilter {
	return NewPaginationFilter(baseFilter)
}