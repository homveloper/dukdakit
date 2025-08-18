package conflux

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// 기본 팩토리 구현체들
// ============================================================================

// SimpleEntityFactory 간단한 엔터티 팩토리 구현체
// 함수형 방식으로 생성/업데이트 로직을 주입받습니다
type SimpleEntityFactory[T any] struct {
	createFn func(ctx context.Context) (T, error)
	updateFn func(ctx context.Context, existing T) (T, error)
}

// NewSimpleEntityFactory 새 SimpleEntityFactory 생성
func NewSimpleEntityFactory[T any](
	createFn func(ctx context.Context) (T, error),
	updateFn func(ctx context.Context, existing T) (T, error),
) UpsertFunc[T] {
	return &SimpleEntityFactory[T]{
		createFn: createFn,
		updateFn: updateFn,
	}
}

// CreateFn EntityFactory 인터페이스 구현
func (f *SimpleEntityFactory[T]) CreateFn(ctx context.Context) (T, error) {
	if f.createFn == nil {
		var empty T
		return empty, fmt.Errorf("create function not provided")
	}
	return f.createFn(ctx)
}

// UpdateFn EntityFactory 인터페이스 구현
func (f *SimpleEntityFactory[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	if f.updateFn == nil {
		// 업데이트 함수가 없으면 기존 엔터티 그대로 반환 (no-op)
		return existing, nil
	}
	return f.updateFn(ctx, existing)
}

// ============================================================================
// 조건부 팩토리 구현체
// ============================================================================

// ConditionalEntityFactory 조건부 실행이 가능한 팩토리 구현체
type ConditionalEntityFactory[T any] struct {
	createFn      func(ctx context.Context) (T, error)
	updateFn      func(ctx context.Context, existing T) (T, error)
	shouldCreateFn func(ctx context.Context) (bool, error)
	shouldUpdateFn func(ctx context.Context, existing T) (bool, error)
}

// NewConditionalEntityFactory 새 ConditionalEntityFactory 생성
func NewConditionalEntityFactory[T any](
	createFn func(ctx context.Context) (T, error),
	updateFn func(ctx context.Context, existing T) (T, error),
	shouldCreateFn func(ctx context.Context) (bool, error),
	shouldUpdateFn func(ctx context.Context, existing T) (bool, error),
) ConditionalUpsertFunc[T] {
	return &ConditionalEntityFactory[T]{
		createFn:       createFn,
		updateFn:       updateFn,
		shouldCreateFn: shouldCreateFn,
		shouldUpdateFn: shouldUpdateFn,
	}
}

// CreateFn EntityFactory 인터페이스 구현
func (f *ConditionalEntityFactory[T]) CreateFn(ctx context.Context) (T, error) {
	if f.createFn == nil {
		var empty T
		return empty, fmt.Errorf("create function not provided")
	}
	return f.createFn(ctx)
}

// UpdateFn EntityFactory 인터페이스 구현
func (f *ConditionalEntityFactory[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	if f.updateFn == nil {
		return existing, nil
	}
	return f.updateFn(ctx, existing)
}

// ShouldCreate ConditionalFactory 인터페이스 구현
func (f *ConditionalEntityFactory[T]) ShouldCreate(ctx context.Context) (bool, error) {
	if f.shouldCreateFn == nil {
		return true, nil // 기본값은 생성 허용
	}
	return f.shouldCreateFn(ctx)
}

// ShouldUpdate ConditionalFactory 인터페이스 구현
func (f *ConditionalEntityFactory[T]) ShouldUpdate(ctx context.Context, existing T) (bool, error) {
	if f.shouldUpdateFn == nil {
		return true, nil // 기본값은 업데이트 허용
	}
	return f.shouldUpdateFn(ctx, existing)
}

// ============================================================================
// 버전 인식 팩토리
// ============================================================================

// VersionAwareFactory 버전을 인식하고 관리하는 팩토리
type VersionAwareFactory[T Versioned] struct {
	createFn func(ctx context.Context) (T, error)
	updateFn func(ctx context.Context, existing T) (T, error)
}

// NewVersionAwareFactory 새 VersionAwareFactory 생성
func NewVersionAwareFactory[T Versioned](
	createFn func(ctx context.Context) (T, error),
	updateFn func(ctx context.Context, existing T) (T, error),
) UpsertFunc[T] {
	return &VersionAwareFactory[T]{
		createFn: createFn,
		updateFn: updateFn,
	}
}

// CreateFn EntityFactory 인터페이스 구현
func (f *VersionAwareFactory[T]) CreateFn(ctx context.Context) (T, error) {
	if f.createFn == nil {
		var empty T
		return empty, fmt.Errorf("create function not provided")
	}
	
	entity, err := f.createFn(ctx)
	if err != nil {
		return entity, err
	}
	
	// 새 엔터티의 버전을 1로 설정
	entity.SetVersion(1)
	
	// Timestamped 인터페이스도 구현하고 있다면 타임스탬프 설정
	if timestamped, ok := any(entity).(Timestamped); ok {
		now := time.Now()
		timestamped.SetCreatedAt(now)
		timestamped.SetUpdatedAt(now)
	}
	
	return entity, nil
}

// UpdateFn EntityFactory 인터페이스 구현
func (f *VersionAwareFactory[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	if f.updateFn == nil {
		return existing, nil
	}
	
	updated, err := f.updateFn(ctx, existing)
	if err != nil {
		return updated, err
	}
	
	// 버전 증가
	updated.SetVersion(existing.GetVersion() + 1)
	
	// 업데이트 타임스탬프 설정
	if timestamped, ok := any(updated).(Timestamped); ok {
		timestamped.SetUpdatedAt(time.Now())
	}
	
	return updated, nil
}

// ============================================================================
// 빌더 패턴 팩토리
// ============================================================================

// FactoryBuilder 팩토리를 빌드하는 빌더
type FactoryBuilder[T any] struct {
	createFn       func(ctx context.Context) (T, error)
	updateFn       func(ctx context.Context, existing T) (T, error)
	shouldCreateFn func(ctx context.Context) (bool, error)
	shouldUpdateFn func(ctx context.Context, existing T) (bool, error)
	validationFn   func(ctx context.Context, entity T) error
}

// NewFactoryBuilder 새 FactoryBuilder 생성
func NewFactoryBuilder[T any]() *FactoryBuilder[T] {
	return &FactoryBuilder[T]{}
}

// WithCreate 생성 함수 설정
func (b *FactoryBuilder[T]) WithCreate(fn func(ctx context.Context) (T, error)) *FactoryBuilder[T] {
	b.createFn = fn
	return b
}

// WithUpdate 업데이트 함수 설정
func (b *FactoryBuilder[T]) WithUpdate(fn func(ctx context.Context, existing T) (T, error)) *FactoryBuilder[T] {
	b.updateFn = fn
	return b
}

// WithCreateCondition 생성 조건 함수 설정
func (b *FactoryBuilder[T]) WithCreateCondition(fn func(ctx context.Context) (bool, error)) *FactoryBuilder[T] {
	b.shouldCreateFn = fn
	return b
}

// WithUpdateCondition 업데이트 조건 함수 설정
func (b *FactoryBuilder[T]) WithUpdateCondition(fn func(ctx context.Context, existing T) (bool, error)) *FactoryBuilder[T] {
	b.shouldUpdateFn = fn
	return b
}

// WithValidation 검증 함수 설정
func (b *FactoryBuilder[T]) WithValidation(fn func(ctx context.Context, entity T) error) *FactoryBuilder[T] {
	b.validationFn = fn
	return b
}

// Build 팩토리를 생성합니다
func (b *FactoryBuilder[T]) Build() UpsertFunc[T] {
	if b.shouldCreateFn != nil || b.shouldUpdateFn != nil {
		return &ConditionalEntityFactory[T]{
			createFn:       b.createFn,
			updateFn:       b.updateFn,
			shouldCreateFn: b.shouldCreateFn,
			shouldUpdateFn: b.shouldUpdateFn,
		}
	}
	
	return &SimpleEntityFactory[T]{
		createFn: b.createFn,
		updateFn: b.updateFn,
	}
}

// BuildValidated 검증 기능이 포함된 팩토리를 생성합니다
func (b *FactoryBuilder[T]) BuildValidated() UpsertFunc[T] {
	factory := b.Build()
	if b.validationFn == nil {
		return factory
	}
	
	return &ValidatedFactory[T]{
		factory:      factory,
		validationFn: b.validationFn,
	}
}

// ============================================================================
// 검증 기능이 포함된 팩토리
// ============================================================================

// ValidatedFactory 검증 기능이 포함된 팩토리 래퍼
type ValidatedFactory[T any] struct {
	factory      UpsertFunc[T]
	validationFn func(ctx context.Context, entity T) error
}

// CreateFn EntityFactory 인터페이스 구현
func (f *ValidatedFactory[T]) CreateFn(ctx context.Context) (T, error) {
	entity, err := f.factory.CreateFn(ctx)
	if err != nil {
		return entity, err
	}
	
	if f.validationFn != nil {
		if err := f.validationFn(ctx, entity); err != nil {
			var empty T
			return empty, fmt.Errorf("validation failed: %w", err)
		}
	}
	
	return entity, nil
}

// UpdateFn EntityFactory 인터페이스 구현
func (f *ValidatedFactory[T]) UpdateFn(ctx context.Context, existing T) (T, error) {
	entity, err := f.factory.UpdateFn(ctx, existing)
	if err != nil {
		return entity, err
	}
	
	if f.validationFn != nil {
		if err := f.validationFn(ctx, entity); err != nil {
			return existing, fmt.Errorf("validation failed: %w", err)
		}
	}
	
	return entity, nil
}