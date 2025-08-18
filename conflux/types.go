package conflux

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// 낙관적 동시성 제어 - 연산별 결과 타입들
// ============================================================================

// InsertResult FindOneAndInsert 연산의 결과
type InsertResult[T any] struct {
	Entity    T         // 생성된 엔터티
	Version   int64     // 엔터티의 버전 (항상 1)
	CreatedAt time.Time // 생성 시각
	Duplicate bool      // 중복으로 인한 실패 여부
}

// NewInsertResult 새 InsertResult 생성
func NewInsertResult[T any](entity T, version int64) *InsertResult[T] {
	return &InsertResult[T]{
		Entity:    entity,
		Version:   version,
		CreatedAt: time.Now(),
		Duplicate: false,
	}
}

// NewDuplicateInsertResult 중복 삽입 실패 결과
func NewDuplicateInsertResult[T any](existingEntity T, version int64) *InsertResult[T] {
	return &InsertResult[T]{
		Entity:    existingEntity,
		Version:   version,
		CreatedAt: time.Now(),
		Duplicate: true,
	}
}

// IsSuccess 삽입이 성공했는지 확인
func (r *InsertResult[T]) IsSuccess() bool {
	return !r.Duplicate
}

// IsDuplicate 중복으로 실패했는지 확인
func (r *InsertResult[T]) IsDuplicate() bool {
	return r.Duplicate
}

// UpsertResult FindOneAndUpsert 연산의 결과
type UpsertResult[T any] struct {
	Entity     T         // 결과 엔터티
	Created    bool      // 새로 생성되었는지 여부
	Version    int64     // 엔터티의 현재 버전
	ModifiedAt time.Time // 최종 수정 시각
}

// NewUpsertResult 새 UpsertResult 생성
func NewUpsertResult[T any](entity T, created bool, version int64) *UpsertResult[T] {
	return &UpsertResult[T]{
		Entity:     entity,
		Created:    created,
		Version:    version,
		ModifiedAt: time.Now(),
	}
}

// WasCreated 새로 생성되었는지 확인
func (r *UpsertResult[T]) WasCreated() bool {
	return r.Created
}

// WasUpdated 기존 엔터티가 업데이트되었는지 확인
func (r *UpsertResult[T]) WasUpdated() bool {
	return !r.Created
}

// UpdateResult FindOneAndUpdate 연산의 결과
type UpdateResult[T any] struct {
	Entity        T         // 업데이트된 엔터티
	Version       int64     // 엔터티의 새 버전
	UpdatedAt     time.Time // 업데이트 시각
	VersionConflict bool    // 버전 충돌 발생 여부
	NotFound      bool      // 엔터티를 찾지 못한 경우
}

// NewUpdateResult 새 UpdateResult 생성
func NewUpdateResult[T any](entity T, version int64) *UpdateResult[T] {
	return &UpdateResult[T]{
		Entity:          entity,
		Version:         version,
		UpdatedAt:       time.Now(),
		VersionConflict: false,
		NotFound:        false,
	}
}

// NewVersionConflictResult 버전 충돌 결과
func NewVersionConflictResult[T any](currentEntity T, currentVersion int64) *UpdateResult[T] {
	return &UpdateResult[T]{
		Entity:          currentEntity,
		Version:         currentVersion,
		UpdatedAt:       time.Now(),
		VersionConflict: true,
		NotFound:        false,
	}
}

// NewNotFoundResult 엔터티를 찾지 못한 경우의 결과
func NewNotFoundResult[T any]() *UpdateResult[T] {
	var empty T
	return &UpdateResult[T]{
		Entity:          empty,
		Version:         0,
		UpdatedAt:       time.Now(),
		VersionConflict: false,
		NotFound:        true,
	}
}

// IsSuccess 업데이트가 성공했는지 확인
func (r *UpdateResult[T]) IsSuccess() bool {
	return !r.VersionConflict && !r.NotFound
}

// HasVersionConflict 버전 충돌이 발생했는지 확인
func (r *UpdateResult[T]) HasVersionConflict() bool {
	return r.VersionConflict
}

// IsNotFound 엔터티를 찾지 못했는지 확인
func (r *UpdateResult[T]) IsNotFound() bool {
	return r.NotFound
}

// ============================================================================
// 공통 인터페이스
// ============================================================================

// Result 모든 결과 타입이 구현해야 하는 공통 인터페이스
type Result[T any] interface {
	GetEntity() T
	GetVersion() int64
}

// GetEntity InsertResult의 엔터티 반환
func (r *InsertResult[T]) GetEntity() T {
	return r.Entity
}

// GetVersion InsertResult의 버전 반환
func (r *InsertResult[T]) GetVersion() int64 {
	return r.Version
}

// GetEntity UpsertResult의 엔터티 반환
func (r *UpsertResult[T]) GetEntity() T {
	return r.Entity
}

// GetVersion UpsertResult의 버전 반환
func (r *UpsertResult[T]) GetVersion() int64 {
	return r.Version
}

// GetEntity UpdateResult의 엔터티 반환
func (r *UpdateResult[T]) GetEntity() T {
	return r.Entity
}

// GetVersion UpdateResult의 버전 반환
func (r *UpdateResult[T]) GetVersion() int64 {
	return r.Version
}

// ============================================================================
// 엔터티 로직 인터페이스 - IoC 패턴의 핵심
// ============================================================================

// CreateFunc 새 엔터티 생성을 위한 로직 (FindOneAndInsert 전용)
type CreateFunc[T any] interface {
	// CreateFn 새 엔터티를 생성합니다
	CreateFn(ctx context.Context) (T, error)
}

// UpdateFunc 기존 엔터티 업데이트를 위한 로직 (FindOneAndUpdate 전용)
type UpdateFunc[T any] interface {
	// UpdateFn 기존 엔터티를 업데이트합니다
	UpdateFn(ctx context.Context, existing T) (T, error)
}

// UpsertFunc 생성/업데이트 모두 지원하는 로직 (FindOneAndUpsert 전용)
type UpsertFunc[T any] interface {
	CreateFunc[T]
	UpdateFunc[T]
}

// ============================================================================
// 함수형 타입 별칭 - 간편한 함수형 프로그래밍 지원
// ============================================================================

// CreateFn 함수형 생성 로직
type CreateFn[T any] func(ctx context.Context) (T, error)

// UpdateFn 함수형 업데이트 로직
type UpdateFn[T any] func(ctx context.Context, existing T) (T, error)

// ============================================================================
// 함수형 래퍼 - 함수를 인터페이스로 변환
// ============================================================================

// FuncCreateFunc 함수를 CreateFunc 인터페이스로 래핑
type FuncCreateFunc[T any] struct {
	Fn CreateFn[T]
}

// CreateFn CreateFunc 인터페이스 구현
func (f FuncCreateFunc[T]) CreateFn(ctx context.Context) (T, error) {
	return f.Fn(ctx)
}

// NewCreateFunc 함수를 CreateFunc로 변환
func NewCreateFunc[T any](fn CreateFn[T]) CreateFunc[T] {
	return FuncCreateFunc[T]{Fn: fn}
}

// FuncUpdateFunc 함수를 UpdateFunc 인터페이스로 래핑
type FuncUpdateFunc[T any] struct {
	Fn UpdateFn[T]
}

// UpdateFn UpdateFunc 인터페이스 구현
func (f FuncUpdateFunc[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	return f.Fn(ctx, existing)
}

// NewUpdateFunc 함수를 UpdateFunc로 변환
func NewUpdateFunc[T any](fn UpdateFn[T]) UpdateFunc[T] {
	return FuncUpdateFunc[T]{Fn: fn}
}

// FuncUpsertFunc 함수들을 UpsertFunc 인터페이스로 래핑
type FuncUpsertFunc[T any] struct {
	CreateFunc CreateFn[T]
	UpdateFunc UpdateFn[T]
}

// CreateFn CreateFunc 인터페이스 구현
func (f FuncUpsertFunc[T]) CreateFn(ctx context.Context) (T, error) {
	return f.CreateFunc(ctx)
}

// UpdateFn UpdateFunc 인터페이스 구현  
func (f FuncUpsertFunc[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	if f.UpdateFunc == nil {
		return existing, nil
	}
	return f.UpdateFunc(ctx, existing)
}

// NewUpsertFunc 함수들을 UpsertFunc로 변환
func NewUpsertFunc[T any](createFn CreateFn[T], updateFn UpdateFn[T]) UpsertFunc[T] {
	return FuncUpsertFunc[T]{
		CreateFunc: createFn,
		UpdateFunc: updateFn,
	}
}

// ConditionalUpsertFunc 조건부 실행이 가능한 Upsert 로직
type ConditionalUpsertFunc[T any] interface {
	UpsertFunc[T]
	
	// ShouldCreate 엔터티를 생성해야 하는지 판단합니다
	ShouldCreate(ctx context.Context) (bool, error)
	
	// ShouldUpdate 기존 엔터티를 업데이트해야 하는지 판단합니다
	ShouldUpdate(ctx context.Context, existing T) (bool, error)
}

// ============================================================================
// 충돌 해결 전략
// ============================================================================

// ConflictResolver 버전 충돌 해결 전략을 정의합니다
type ConflictResolver[T any] interface {
	// ResolveConflict 버전 충돌 시 해결 방법을 결정합니다
	// current: 현재 저장된 엔터티 (최신 버전)
	// incoming: 수정하려던 엔터티 (구버전 기반)
	// 반환값: 최종 적용할 엔터티, 적용 여부, 에러
	ResolveConflict(ctx context.Context, current T, incoming T) (T, bool, error)
}

// ConflictStrategy 기본 제공되는 충돌 해결 전략들
type ConflictStrategy int

const (
	// FailOnConflict 충돌 시 실패 (기본값)
	FailOnConflict ConflictStrategy = iota
	
	// OverwriteOnConflict 충돌 시 덮어쓰기 (Last Writer Wins)
	OverwriteOnConflict
	
	// MergeOnConflict 충돌 시 병합 (사용자 정의 병합 로직 필요)
	MergeOnConflict
	
	// RetryOnConflict 충돌 시 재시도
	RetryOnConflict
)

// ============================================================================
// 버전 관리 인터페이스
// ============================================================================

// Versioned 버전 관리가 가능한 엔터티가 구현해야 하는 인터페이스
type Versioned interface {
	GetVersion() int64
	SetVersion(version int64)
}

// Timestamped 타임스탬프 관리가 가능한 엔터티 인터페이스
type Timestamped interface {
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
}

// Entity 완전한 엔터티가 구현해야 하는 인터페이스 조합
type Entity interface {
	Versioned
	Timestamped
}

// ============================================================================
// 기본 엔터티 구현체
// ============================================================================

// BaseEntity 기본적인 버전 및 타임스탬프 관리 기능을 제공하는 구조체
// 사용자 정의 엔터티에 임베드하여 사용할 수 있습니다
type BaseEntity struct {
	Version   int64     `json:"version" bson:"version"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// GetVersion 버전 반환
func (e *BaseEntity) GetVersion() int64 {
	return e.Version
}

// SetVersion 버전 설정
func (e *BaseEntity) SetVersion(version int64) {
	e.Version = version
}

// GetCreatedAt 생성 시간 반환
func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// SetCreatedAt 생성 시간 설정
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// GetUpdatedAt 업데이트 시간 반환
func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// SetUpdatedAt 업데이트 시간 설정
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// ============================================================================
// 에러 타입들
// ============================================================================

// ConflictError 버전 충돌 에러
type ConflictError struct {
	ExpectedVersion int64
	ActualVersion   int64
	Message         string
}

func (e *ConflictError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("version conflict: expected %d, got %d", e.ExpectedVersion, e.ActualVersion)
}

// NewConflictError 새 충돌 에러 생성
func NewConflictError(expected, actual int64, message string) *ConflictError {
	return &ConflictError{
		ExpectedVersion: expected,
		ActualVersion:   actual,
		Message:         message,
	}
}