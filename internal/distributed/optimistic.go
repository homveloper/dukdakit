package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OptimisticConfig holds configuration for optimistic concurrency control
type OptimisticConfig struct {
	MaxRetries    int
	RetryDelay    time.Duration
	VersionField  string
	EnableMetrics bool
}

// DefaultOptimisticConfig returns default configuration
func DefaultOptimisticConfig() OptimisticConfig {
	return OptimisticConfig{
		MaxRetries:    3,
		RetryDelay:    100 * time.Millisecond,
		VersionField:  "version",
		EnableMetrics: true,
	}
}

// OptimisticController handles optimistic concurrency control
type OptimisticController struct {
	config  OptimisticConfig
	metrics *OptimisticMetrics
	mu      sync.RWMutex
}

// OptimisticMetrics tracks performance metrics
type OptimisticMetrics struct {
	TotalOperations   int64
	SuccessfulUpdates int64
	ConflictRetries   int64
	FailedOperations  int64
}

// VersionedEntity represents an entity with version control
type VersionedEntity interface {
	GetID() string
	GetVersion() int64
	SetVersion(version int64)
}

// ConflictError represents a version conflict
type ConflictError struct {
	EntityID        string
	ExpectedVersion int64
	ActualVersion   int64
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("version conflict for entity %s: expected %d, got %d",
		e.EntityID, e.ExpectedVersion, e.ActualVersion)
}

// NewOptimistic creates a new optimistic concurrency controller
func NewOptimistic(config ...OptimisticConfig) *OptimisticController {
	cfg := DefaultOptimisticConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &OptimisticController{
		config: cfg,
		metrics: &OptimisticMetrics{},
	}
}

// UpdateWithOptimisticLock performs an optimistic update operation
func (oc *OptimisticController) UpdateWithOptimisticLock(
	ctx context.Context,
	entity VersionedEntity,
	updateFn func(VersionedEntity) error,
	validateFn func(VersionedEntity) error,
) error {
	oc.mu.Lock()
	oc.metrics.TotalOperations++
	oc.mu.Unlock()

	originalVersion := entity.GetVersion()
	
	for attempt := 0; attempt <= oc.config.MaxRetries; attempt++ {
		// Validate entity state before update
		if validateFn != nil {
			if err := validateFn(entity); err != nil {
				oc.incrementFailedOperations()
				return fmt.Errorf("validation failed: %w", err)
			}
		}

		// Increment version for optimistic locking
		entity.SetVersion(originalVersion + 1)

		// Apply the update
		if err := updateFn(entity); err != nil {
			// Check if it's a version conflict
			if conflictErr, ok := err.(*ConflictError); ok {
				if attempt < oc.config.MaxRetries {
					oc.incrementConflictRetries()
					
					// Wait before retry
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(oc.config.RetryDelay):
						// Update to latest version and retry
						originalVersion = conflictErr.ActualVersion
						continue
					}
				}
			}
			
			oc.incrementFailedOperations()
			return fmt.Errorf("update failed after %d attempts: %w", attempt+1, err)
		}

		// Success
		oc.incrementSuccessfulUpdates()
		return nil
	}

	oc.incrementFailedOperations()
	return fmt.Errorf("max retry attempts (%d) exceeded for entity %s", 
		oc.config.MaxRetries, entity.GetID())
}

// CompareAndSwap performs atomic compare-and-swap operation
func (oc *OptimisticController) CompareAndSwap(
	ctx context.Context,
	entity VersionedEntity,
	expectedVersion int64,
	updateFn func(VersionedEntity) error,
) error {
	if entity.GetVersion() != expectedVersion {
		return &ConflictError{
			EntityID:        entity.GetID(),
			ExpectedVersion: expectedVersion,
			ActualVersion:   entity.GetVersion(),
		}
	}

	return oc.UpdateWithOptimisticLock(ctx, entity, updateFn, nil)
}

// GetMetrics returns current metrics
func (oc *OptimisticController) GetMetrics() OptimisticMetrics {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return *oc.metrics
}

// ResetMetrics resets all metrics to zero
func (oc *OptimisticController) ResetMetrics() {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.metrics = &OptimisticMetrics{}
}

// Helper methods for metrics
func (oc *OptimisticController) incrementSuccessfulUpdates() {
	oc.mu.Lock()
	oc.metrics.SuccessfulUpdates++
	oc.mu.Unlock()
}

func (oc *OptimisticController) incrementConflictRetries() {
	oc.mu.Lock()
	oc.metrics.ConflictRetries++
	oc.mu.Unlock()
}

func (oc *OptimisticController) incrementFailedOperations() {
	oc.mu.Lock()
	oc.metrics.FailedOperations++
	oc.mu.Unlock()
}