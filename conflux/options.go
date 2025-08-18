package conflux

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Repository 연산 옵션들
// ============================================================================

// InsertOption FindOneAndInsert 연산의 옵션
type InsertOption func(*InsertConfig)

// InsertConfig FindOneAndInsert 연산 설정
type InsertConfig struct {
	// OnDuplicate 중복 발견 시 처리 전략
	OnDuplicate DuplicateStrategy
	
	// ConflictResolver 충돌 해결 전략 (커스텀)
	ConflictResolver any
	
	// Timeout 연산 타임아웃
	Timeout time.Duration
	
	// Metadata 연산에 추가할 메타데이터
	Metadata map[string]any
}

// NewInsertConfig 기본 InsertConfig 생성
func NewInsertConfig() *InsertConfig {
	return &InsertConfig{
		OnDuplicate: FailOnDuplicate,
		Timeout:     30 * time.Second,
		Metadata:    make(map[string]any),
	}
}

// UpsertOption FindOneAndUpsert 연산의 옵션
type UpsertOption func(*UpsertConfig)

// UpsertConfig FindOneAndUpsert 연산 설정
type UpsertConfig struct {
	// ConflictResolver 버전 충돌 해결 전략
	ConflictResolver any
	
	// OnConflict 충돌 발견 시 기본 처리 전략
	OnConflict ConflictStrategy
	
	// MaxRetries 최대 재시도 횟수 (외부 재시도 라이브러리와 조합 시 힌트)
	MaxRetries int
	
	// Timeout 연산 타임아웃
	Timeout time.Duration
	
	// Metadata 연산에 추가할 메타데이터
	Metadata map[string]any
}

// NewUpsertConfig 기본 UpsertConfig 생성
func NewUpsertConfig() *UpsertConfig {
	return &UpsertConfig{
		OnConflict: RetryOnConflict,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
		Metadata:   make(map[string]any),
	}
}

// UpdateOption FindOneAndUpdate 연산의 옵션
type UpdateOption func(*UpdateConfig)

// UpdateConfig FindOneAndUpdate 연산 설정
type UpdateConfig struct {
	// ConflictResolver 버전 충돌 해결 전략
	ConflictResolver any
	
	// OnConflict 충돌 발견 시 기본 처리 전략
	OnConflict ConflictStrategy
	
	// OnNotFound 엔터티를 찾지 못한 경우 처리 전략
	OnNotFound NotFoundStrategy
	
	// MaxRetries 최대 재시도 횟수
	MaxRetries int
	
	// Timeout 연산 타임아웃
	Timeout time.Duration
	
	// Metadata 연산에 추가할 메타데이터
	Metadata map[string]any
}

// NewUpdateConfig 기본 UpdateConfig 생성
func NewUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		OnConflict: FailOnConflict,
		OnNotFound: FailOnNotFound,
		MaxRetries: 0, // 기본적으로 재시도 없음
		Timeout:    30 * time.Second,
		Metadata:   make(map[string]any),
	}
}

// ============================================================================
// 중복 처리 전략
// ============================================================================

// DuplicateStrategy 중복 엔터티 발견 시 처리 전략
type DuplicateStrategy int

const (
	// FailOnDuplicate 중복 시 실패 (기본값)
	FailOnDuplicate DuplicateStrategy = iota
	
	// ReturnExistingOnDuplicate 중복 시 기존 엔터티 반환
	ReturnExistingOnDuplicate
	
	// UpdateOnDuplicate 중복 시 업데이트 시도
	UpdateOnDuplicate
)

// ============================================================================
// Not Found 처리 전략
// ============================================================================

// NotFoundStrategy 엔터티를 찾지 못한 경우 처리 전략
type NotFoundStrategy int

const (
	// FailOnNotFound 못 찾으면 실패 (기본값)
	FailOnNotFound NotFoundStrategy = iota
	
	// CreateOnNotFound 못 찾으면 새로 생성
	CreateOnNotFound
	
	// IgnoreOnNotFound 못 찾아도 무시 (no-op)
	IgnoreOnNotFound
)

// ============================================================================
// 충돌 해결자 인터페이스 (기존 확장)
// ============================================================================

// ConflictContext 충돌 컨텍스트 정보
type ConflictContext[T any] struct {
	// Operation 수행 중인 연산 타입
	Operation string // "insert", "update", "upsert"
	
	// ExpectedVersion 기대했던 버전
	ExpectedVersion int64
	
	// CurrentVersion 현재 실제 버전
	CurrentVersion int64
	
	// AttemptNumber 시도 횟수 (1부터 시작)
	AttemptNumber int
	
	// Metadata 추가 메타데이터
	Metadata map[string]any
}

// EnhancedConflictResolver 향상된 충돌 해결자
type EnhancedConflictResolver[T any] interface {
	ConflictResolver[T]
	
	// ResolveConflictWithContext 컨텍스트 정보를 포함한 충돌 해결
	ResolveConflictWithContext(
		ctx context.Context, 
		conflictCtx *ConflictContext[T],
		current T, 
		incoming T,
	) (T, ConflictResolution, error)
}

// ConflictResolution 충돌 해결 결과
type ConflictResolution int

const (
	// ResolveWithMerged 병합된 엔터티로 해결
	ResolveWithMerged ConflictResolution = iota
	
	// ResolveWithCurrent 현재 엔터티로 해결 (충돌 무시)
	ResolveWithCurrent
	
	// ResolveWithIncoming 새로운 엔터티로 해결 (덮어쓰기)
	ResolveWithIncoming
	
	// ResolveWithRetry 재시도 요청
	ResolveWithRetry
	
	// ResolveWithFail 해결 불가, 실패
	ResolveWithFail
)

// ============================================================================
// 옵션 빌더 함수들
// ============================================================================

// WithDuplicateStrategy 중복 처리 전략 설정
func WithDuplicateStrategy(strategy DuplicateStrategy) InsertOption {
	return func(config *InsertConfig) {
		config.OnDuplicate = strategy
	}
}

// WithUpsertConflictStrategy Upsert 연산의 충돌 처리 전략 설정
func WithUpsertConflictStrategy(strategy ConflictStrategy) UpsertOption {
	return func(config *UpsertConfig) {
		config.OnConflict = strategy
	}
}

// WithUpdateConflictStrategy Update 연산의 충돌 처리 전략 설정  
func WithUpdateConflictStrategy(strategy ConflictStrategy) UpdateOption {
	return func(config *UpdateConfig) {
		config.OnConflict = strategy
	}
}

// WithInsertConflictResolver Insert 연산의 커스텀 충돌 해결자 설정
func WithInsertConflictResolver(resolver any) InsertOption {
	return func(config *InsertConfig) {
		config.ConflictResolver = resolver
	}
}

// WithUpsertConflictResolver Upsert 연산의 커스텀 충돌 해결자 설정
func WithUpsertConflictResolver(resolver any) UpsertOption {
	return func(config *UpsertConfig) {
		config.ConflictResolver = resolver
	}
}

// WithUpdateConflictResolver Update 연산의 커스텀 충돌 해결자 설정
func WithUpdateConflictResolver(resolver any) UpdateOption {
	return func(config *UpdateConfig) {
		config.ConflictResolver = resolver
	}
}

// WithNotFoundStrategy Not Found 처리 전략 설정
func WithNotFoundStrategy(strategy NotFoundStrategy) UpdateOption {
	return func(config *UpdateConfig) {
		config.OnNotFound = strategy
	}
}

// WithUpsertMaxRetries Upsert 연산의 최대 재시도 횟수 설정
func WithUpsertMaxRetries(maxRetries int) UpsertOption {
	return func(config *UpsertConfig) {
		config.MaxRetries = maxRetries
	}
}

// WithUpdateMaxRetries Update 연산의 최대 재시도 횟수 설정
func WithUpdateMaxRetries(maxRetries int) UpdateOption {
	return func(config *UpdateConfig) {
		config.MaxRetries = maxRetries
	}
}

// WithInsertTimeout Insert 연산의 타임아웃 설정
func WithInsertTimeout(timeout time.Duration) InsertOption {
	return func(config *InsertConfig) {
		config.Timeout = timeout
	}
}

// WithUpsertTimeout Upsert 연산의 타임아웃 설정
func WithUpsertTimeout(timeout time.Duration) UpsertOption {
	return func(config *UpsertConfig) {
		config.Timeout = timeout
	}
}

// WithUpdateTimeout Update 연산의 타임아웃 설정
func WithUpdateTimeout(timeout time.Duration) UpdateOption {
	return func(config *UpdateConfig) {
		config.Timeout = timeout
	}
}

// WithInsertMetadata Insert 연산의 메타데이터 설정
func WithInsertMetadata(metadata map[string]any) InsertOption {
	return func(config *InsertConfig) {
		for k, v := range metadata {
			config.Metadata[k] = v
		}
	}
}

// WithUpsertMetadata Upsert 연산의 메타데이터 설정
func WithUpsertMetadata(metadata map[string]any) UpsertOption {
	return func(config *UpsertConfig) {
		for k, v := range metadata {
			config.Metadata[k] = v
		}
	}
}

// WithUpdateMetadata Update 연산의 메타데이터 설정
func WithUpdateMetadata(metadata map[string]any) UpdateOption {
	return func(config *UpdateConfig) {
		for k, v := range metadata {
			config.Metadata[k] = v
		}
	}
}

// ============================================================================
// 미리 정의된 충돌 해결자들
// ============================================================================

// OverwriteResolver 덮어쓰기 충돌 해결자
type OverwriteResolver[T any] struct{}

func (r *OverwriteResolver[T]) ResolveConflict(ctx context.Context, current T, incoming T) (T, bool, error) {
	return incoming, true, nil // 항상 새로운 값으로 덮어쓰기
}

func (r *OverwriteResolver[T]) ResolveConflictWithContext(
	ctx context.Context, 
	conflictCtx *ConflictContext[T], 
	current T, 
	incoming T,
) (T, ConflictResolution, error) {
	return incoming, ResolveWithIncoming, nil
}

// IgnoreConflictResolver 충돌 무시 해결자
type IgnoreConflictResolver[T any] struct{}

func (r *IgnoreConflictResolver[T]) ResolveConflict(ctx context.Context, current T, incoming T) (T, bool, error) {
	return current, true, nil // 현재 값 유지
}

func (r *IgnoreConflictResolver[T]) ResolveConflictWithContext(
	ctx context.Context, 
	conflictCtx *ConflictContext[T], 
	current T, 
	incoming T,
) (T, ConflictResolution, error) {
	return current, ResolveWithCurrent, nil
}

// RetryRequestResolver 재시도 요청 해결자
type RetryRequestResolver[T any] struct {
	MaxAttempts int
}

func (r *RetryRequestResolver[T]) ResolveConflict(ctx context.Context, current T, incoming T) (T, bool, error) {
	// 기존 인터페이스에서는 재시도를 에러로 표현
	return incoming, false, fmt.Errorf("retry requested")
}

func (r *RetryRequestResolver[T]) ResolveConflictWithContext(
	ctx context.Context, 
	conflictCtx *ConflictContext[T], 
	current T, 
	incoming T,
) (T, ConflictResolution, error) {
	if conflictCtx.AttemptNumber >= r.MaxAttempts {
		return current, ResolveWithFail, fmt.Errorf("max attempts exceeded")
	}
	return incoming, ResolveWithRetry, nil
}

// ============================================================================
// 편의 함수들
// ============================================================================

// NewOverwriteResolver 덮어쓰기 해결자 생성
func NewOverwriteResolver[T any]() EnhancedConflictResolver[T] {
	return &OverwriteResolver[T]{}
}

// NewIgnoreConflictResolver 충돌 무시 해결자 생성
func NewIgnoreConflictResolver[T any]() EnhancedConflictResolver[T] {
	return &IgnoreConflictResolver[T]{}
}

// NewRetryRequestResolver 재시도 요청 해결자 생성
func NewRetryRequestResolver[T any](maxAttempts int) EnhancedConflictResolver[T] {
	return &RetryRequestResolver[T]{MaxAttempts: maxAttempts}
}