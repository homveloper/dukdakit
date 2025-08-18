package conflux

import (
	"context"
)

// ============================================================================
// 낙관적 동시성 제어 레포지토리 인터페이스
// ============================================================================

// Repository 낙관적 동시성 제어를 지원하는 레포지토리 인터페이스
// T: 엔터티 타입, F: 필터 타입 (각 인프라에서 정의)
type Repository[T any, F any] interface {
	// FindOneAndInsert 새 엔터티를 원자적으로 삽입합니다
	// duplicateCheckFilter: 중복 검사용 조건 (인프라별 타입)
	// createFunc: 엔터티 생성 로직
	// options: 선택적 설정 (충돌 해결 전략, 타임아웃 등)
	FindOneAndInsert(ctx context.Context, duplicateCheckFilter F, createFunc CreateFunc[T], options ...InsertOption) (*InsertResult[T], error)
	
	// FindOneAndUpsert 엔터티를 원자적으로 생성하거나 업데이트합니다
	// 없으면 생성, 있으면 업데이트합니다
	// lookupFilter: 엔터티 조회용 조건 (ID나 복합 키 등)  
	// options: 선택적 설정 (충돌 해결 전략, 재시도 횟수 등)
	FindOneAndUpsert(ctx context.Context, lookupFilter F, upsertFunc UpsertFunc[T], options ...UpsertOption) (*UpsertResult[T], error)
	
	// FindOneAndUpdate 기존 엔터티를 원자적으로 업데이트합니다
	// 엔터티가 없거나 버전 충돌 시 적절한 결과를 반환합니다
	// options: 선택적 설정 (충돌 해결 전략, NotFound 처리 등)
	FindOneAndUpdate(ctx context.Context, lookupFilter F, expectedVersion int64, updateFunc UpdateFunc[T], options ...UpdateOption) (*UpdateResult[T], error)
	
	// NewEntity 새로운 엔터티 인스턴스를 생성합니다
	// 사용자가 올바른 타입의 엔터티를 생성할 수 있도록 도와줍니다
	NewEntity() T
}

// ============================================================================
// 선택적 기능을 위한 확장 인터페이스들
// ============================================================================

// ReadRepository 읽기 전용 연산을 제공하는 인터페이스
type ReadRepository[T any, F any] interface {
	// FindOne 필터 조건으로 엔터티를 조회합니다
	FindOne(ctx context.Context, filter F) (T, error)
	
	// Exists 엔터티가 존재하는지 확인합니다
	Exists(ctx context.Context, filter F) (bool, error)
	
	// GetVersion 엔터티의 현재 버전을 조회합니다
	GetVersion(ctx context.Context, filter F) (int64, error)
}

// BatchRepository 배치 연산을 지원하는 인터페이스
type BatchRepository[T any, F any] interface {
	// FindMany 필터 조건으로 여러 엔터티들을 조회합니다
	FindMany(ctx context.Context, filter F, limit int) ([]T, error)
	
	// InsertMany 여러 엔터티를 배치 삽입합니다
	// 중복 충돌 시 성공한 것과 실패한 것을 구분하여 반환합니다
	InsertMany(ctx context.Context, duplicateCheckFilters []F, createFuncs []CreateFunc[T]) ([]*InsertResult[T], error)
}

// QueryRepository 고급 쿼리 기능을 지원하는 인터페이스
type QueryRepository[T any, F any] interface {
	// FindByFilter 사용자 정의 필터로 엔터티들을 조회합니다
	FindByFilter(ctx context.Context, filter F, limit int, offset int) ([]T, error)
	
	// CountByFilter 필터 조건에 맞는 엔터티 개수를 반환합니다
	CountByFilter(ctx context.Context, filter F) (int64, error)
	
	// FindWithSort 정렬이 포함된 조회
	FindWithSort(ctx context.Context, filter F, sortBy string, ascending bool, limit int) ([]T, error)
}

// FullRepository 모든 기능을 제공하는 완전한 레포지토리 인터페이스
type FullRepository[T any, F any] interface {
	Repository[T, F]
	ReadRepository[T, F]
	BatchRepository[T, F]
	QueryRepository[T, F]
}

// ============================================================================
// 트랜잭션 지원을 위한 인터페이스들
// ============================================================================

// TransactionContext 트랜잭션 컨텍스트를 나타내는 인터페이스
type TransactionContext interface {
	// Commit 트랜잭션을 커밋합니다
	Commit(ctx context.Context) error
	
	// Rollback 트랜잭션을 롤백합니다
	Rollback(ctx context.Context) error
	
	// IsActive 트랜잭션이 활성 상태인지 확인합니다
	IsActive() bool
}

// TransactionalRepository 트랜잭션을 지원하는 레포지토리
type TransactionalRepository[T any, ID comparable] interface {
	Repository[T, ID]
	
	// BeginTransaction 새 트랜잭션을 시작합니다
	BeginTransaction(ctx context.Context) (TransactionContext, error)
	
	// WithTransaction 트랜잭션 컨텍스트 내에서 작동하는 레포지토리를 반환합니다
	WithTransaction(tx TransactionContext) Repository[T, ID]
}

// ============================================================================
// 캐싱을 지원하는 인터페이스들
// ============================================================================

// CacheOptions 캐시 설정 옵션
type CacheOptions struct {
	TTL       int64  // 캐시 유효시간(초)
	Namespace string // 캐시 네임스페이스
	Tags      []string // 캐시 태그들
}

// CachedRepository 캐싱을 지원하는 레포지토리
type CachedRepository[T any, ID comparable] interface {
	Repository[T, ID]
	
	// FindByIDCached 캐시를 사용하여 엔터티를 조회합니다
	FindByIDCached(ctx context.Context, id ID, opts *CacheOptions) (T, error)
	
	// InvalidateCache 특정 엔터티의 캐시를 무효화합니다
	InvalidateCache(ctx context.Context, id ID) error
	
	// InvalidateCacheByTags 태그로 관련 캐시들을 일괄 무효화합니다
	InvalidateCacheByTags(ctx context.Context, tags []string) error
}

// ============================================================================
// 이벤트 지원을 위한 인터페이스들
// ============================================================================

// EntityEvent 엔터티 관련 이벤트
type EntityEvent[T any] struct {
	Type      string    // 이벤트 타입 (created, updated, deleted)
	Entity    T         // 관련 엔터티
	OldEntity *T        // 이전 엔터티 (업데이트 시)
	Version   int64     // 엔터티 버전
	Metadata  map[string]any // 추가 메타데이터
}

// EventHandler 엔터티 이벤트를 처리하는 핸들러
type EventHandler[T any] interface {
	// HandleEvent 엔터티 이벤트를 처리합니다
	HandleEvent(ctx context.Context, event *EntityEvent[T]) error
}

// EventAwareRepository 이벤트를 발생시키는 레포지토리
type EventAwareRepository[T any, ID comparable] interface {
	Repository[T, ID]
	
	// RegisterEventHandler 이벤트 핸들러를 등록합니다
	RegisterEventHandler(handler EventHandler[T]) error
	
	// UnregisterEventHandler 이벤트 핸들러를 해제합니다
	UnregisterEventHandler(handler EventHandler[T]) error
}

// ============================================================================
// 메트릭스 및 모니터링을 위한 인터페이스들
// ============================================================================

// RepositoryMetrics 레포지토리 성능 메트릭
type RepositoryMetrics struct {
	TotalOperations   int64   // 총 연산 수
	SuccessfulOps     int64   // 성공한 연산 수
	ConflictCount     int64   // 충돌 발생 수
	AverageLatencyMs  float64 // 평균 응답시간(밀리초)
	ErrorRate         float64 // 에러율 (0.0 ~ 1.0)
}

// MetricsRepository 메트릭 수집을 지원하는 레포지토리
type MetricsRepository[T any, ID comparable] interface {
	Repository[T, ID]
	
	// GetMetrics 현재까지의 메트릭을 반환합니다
	GetMetrics(ctx context.Context) (*RepositoryMetrics, error)
	
	// ResetMetrics 메트릭을 초기화합니다
	ResetMetrics(ctx context.Context) error
}