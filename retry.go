package dukdakit

import (
	"github.com/homveloper/dukdakit/internal/retry"
)

// RetryCategory provides retry and resilience features
type RetryCategory struct{}

// Retry is the global instance for retry features
var Retry = &RetryCategory{}

// New creates a new retrier with the given configuration
//
// Example usage:
//   retrier := dukdakit.Retry.New()
//   
//   // With custom config:
//   config := retry.RetryConfig{
//       MaxAttempts: 5,
//       BaseDelay:   200 * time.Millisecond,
//       Multiplier:  1.5,
//   }
//   retrier := dukdakit.Retry.New(config)
//   
//   // Execute with retry:
//   err := retrier.Execute(ctx, func() error {
//       return someOperation()
//   })
func (r *RetryCategory) New(config ...retry.RetryConfig) *retry.Retrier {
	return retry.New(config...)
}

// Config returns the default retry configuration
func (r *RetryCategory) Config() retry.RetryConfig {
	return retry.DefaultRetryConfig()
}

// NewExponentialBackoff creates a new exponential backoff policy
func (r *RetryCategory) NewExponentialBackoff(config retry.RetryConfig) *retry.ExponentialBackoffPolicy {
	return retry.NewExponentialBackoffPolicy(config)
}

// NewCircuitBreaker creates a new circuit breaker
func (r *RetryCategory) NewCircuitBreaker(config retry.CircuitBreakerConfig) *retry.CircuitBreaker {
	return retry.NewCircuitBreaker(config)
}