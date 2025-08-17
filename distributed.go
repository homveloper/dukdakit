package dukdakit

import (
	"github.com/homveloper/dukdakit/internal/distributed"
)

// DistributedCategory provides distributed computing features
type DistributedCategory struct{}

// Distributed is the global instance for distributed features
var Distributed = &DistributedCategory{}

// NewOptimistic creates a new optimistic concurrency controller
//
// Example usage:
//
//	controller := dukdakit.Distributed.NewOptimistic()
//
//	// With custom config:
//	config := distributed.OptimisticConfig{
//	    MaxRetries: 5,
//	    RetryDelay: 200 * time.Millisecond,
//	}
//	controller := dukdakit.Distributed.NewOptimistic(config)
func (d *DistributedCategory) NewOptimistic(config ...distributed.OptimisticConfig) *distributed.OptimisticController {
	return distributed.NewOptimistic(config...)
}

// OptimisticConfig returns the default optimistic configuration
func (d *DistributedCategory) OptimisticConfig() distributed.OptimisticConfig {
	return distributed.DefaultOptimisticConfig()
}
