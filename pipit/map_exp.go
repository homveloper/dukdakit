package pipit

import (
	"context"
)

// UnsafeMapOperation handles runtime type conversion without compile-time safety
// This is an experimental operation that bypasses Go's type system
type UnsafeMapOperation struct {
	unsafeMapper                func(any) any
	unsafeMapperWithContext     func(context.Context, any) (any, error)
}

// Type returns the operation type identifier for UnsafeMapOperation
func (op *UnsafeMapOperation) Type() OperationType {
	return MapOp // Same as regular map operation
}

// Apply executes the unsafe mapping operation on a single item
// WARNING: This bypasses all type checking and can cause runtime panics
func (op *UnsafeMapOperation) Apply(ctx context.Context, item any) (any, error) {
	// Check context cancellation first for immediate response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	// Handle context-aware unsafe mapper with error handling
	if op.unsafeMapperWithContext != nil {
		// Use the context-aware mapper
		return op.unsafeMapperWithContext(ctx, item)
	}
	
	// Apply simple unsafe mapping - no type checking whatsoever
	// This can panic if the mapper function expects specific types
	result := op.unsafeMapper(item)
	return result, nil
}

// MapUnsafe applies an unsafe transformation function that operates on any type
// This method completely bypasses Go's type system and should be used with extreme caution
//
// WARNING: This is an experimental feature that sacrifices type safety for runtime flexibility
// Use only when:
// - Prototyping with dynamic data
// - Migrating from untyped systems
// - Dealing with JSON/interface{} data where types are known at runtime
//
// Example (dangerous but possible):
//   data := []int{1, 2, 3}
//   result := From(data).MapUnsafe(func(item any) any {
//       // Runtime type assertion required - can panic!
//       return fmt.Sprintf("num_%d", item.(int))
//   }).ToSlice(ctx)
//
// If the Query already has an error, MapUnsafe returns the same Query unchanged.
func (q *Query[T]) MapUnsafe(mapper func(any) any) *Query[any] {
	if q.err != nil {
		return &Query[any]{err: q.err, ctx: q.ctx}
	}
	
	// Create unsafe operation
	op := &UnsafeMapOperation{unsafeMapper: mapper}
	
	// Return new Query[any] - we lose all type information
	return &Query[any]{
		source:   nil, // Will be resolved during execution
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}
}

// MapUnsafeE applies a context-aware unsafe transformation function with error handling
// This is the error-safe version of MapUnsafe that supports context cancellation and 
// mappers that can fail during execution, while still bypassing compile-time type safety
//
// The mapper function receives the current context and can return an error.
// If the mapper returns an error, the entire pipeline execution will stop and
// return that error. Context cancellation is properly handled.
//
// WARNING: This is an experimental feature that sacrifices type safety for runtime flexibility
// Use only when:
// - Processing dynamic data that might fail (JSON parsing, etc.)
// - Need timeout/cancellation support in unsafe operations
// - Migrating from systems where errors are expected
//
// Example (with error handling):
//   data := []any{"123", "invalid", "456"}
//   result := From(data).MapUnsafeE(func(ctx context.Context, item any) (any, error) {
//       str, ok := item.(string)
//       if !ok {
//           return nil, fmt.Errorf("expected string, got %T", item)
//       }
//       
//       // This can fail and return error instead of panicking
//       num, err := strconv.Atoi(str)
//       if err != nil {
//           return nil, fmt.Errorf("invalid number: %w", err)
//       }
//       return num, nil
//   }).ToSlice(ctx)
//
// If the Query already has an error, MapUnsafeE returns the same Query unchanged.
func (q *Query[T]) MapUnsafeE(mapper func(context.Context, any) (any, error)) *Query[any] {
	if q.err != nil {
		return &Query[any]{err: q.err, ctx: q.ctx}
	}
	
	// Create unsafe operation with context-aware mapper
	op := &UnsafeMapOperation{unsafeMapperWithContext: mapper}
	
	// Return new Query[any] - we lose all type information but gain error handling
	return &Query[any]{
		source:   nil, // Will be resolved during execution
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}
}