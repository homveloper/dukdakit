package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Policy defines retry behavior
type Policy interface {
	ShouldRetry(attempt int, err error) bool
	GetDelay(attempt int) time.Duration
	GetMaxAttempts() int
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	Jitter          bool
	RetryableErrors []string
	CircuitBreaker  *CircuitBreakerConfig
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	HalfOpenTimeout  time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
		RetryableErrors: []string{
			"NetworkError",
			"TimeoutError",
			"ServiceUnavailable",
			"InternalServerError",
		},
	}
}

// ExponentialBackoffPolicy implements exponential backoff retry policy
type ExponentialBackoffPolicy struct {
	config RetryConfig
	rng    *rand.Rand
	mu     sync.Mutex
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(config RetryConfig) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldRetry determines if the operation should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= p.config.MaxAttempts {
		return false
	}

	if len(p.config.RetryableErrors) == 0 {
		return true // Retry all errors if no specific errors are configured
	}

	errorType := fmt.Sprintf("%T", err)
	for _, retryableError := range p.config.RetryableErrors {
		if errorType == retryableError || err.Error() == retryableError {
			return true
		}
	}

	return false
}

// GetDelay calculates the delay for the next retry attempt
func (p *ExponentialBackoffPolicy) GetDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return p.config.BaseDelay
	}

	delay := float64(p.config.BaseDelay) * math.Pow(p.config.Multiplier, float64(attempt-1))
	
	if p.config.Jitter {
		p.mu.Lock()
		jitter := p.rng.Float64() * 0.1 * delay // 10% jitter
		p.mu.Unlock()
		delay += jitter
	}

	maxDelay := float64(p.config.MaxDelay)
	if delay > maxDelay {
		delay = maxDelay
	}

	return time.Duration(delay)
}

// GetMaxAttempts returns the maximum number of attempts
func (p *ExponentialBackoffPolicy) GetMaxAttempts() int {
	return p.config.MaxAttempts
}

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	config       CircuitBreakerConfig
	failures     int
	lastFailTime time.Time
	state        CircuitState
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Since(cb.lastFailTime) > cb.config.ResetTimeout
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failures = 0
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.config.FailureThreshold {
		cb.state = CircuitOpen
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Retrier provides retry functionality
type Retrier struct {
	policy         Policy
	circuitBreaker *CircuitBreaker
	metrics        *RetryMetrics
	mu             sync.RWMutex
}

// RetryMetrics tracks retry statistics
type RetryMetrics struct {
	TotalAttempts    int64
	SuccessfulCalls  int64
	FailedCalls      int64
	CircuitBreaks    int64
	AverageAttempts  float64
}

// New creates a new retrier with the given configuration
func New(config ...RetryConfig) *Retrier {
	cfg := DefaultRetryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	retrier := &Retrier{
		policy:  NewExponentialBackoffPolicy(cfg),
		metrics: &RetryMetrics{},
	}

	if cfg.CircuitBreaker != nil {
		retrier.circuitBreaker = NewCircuitBreaker(*cfg.CircuitBreaker)
	}

	return retrier
}

// Execute executes a function with retry logic
func (r *Retrier) Execute(ctx context.Context, fn func() error) error {
	return r.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, fn()
	})
}

// ExecuteWithResult executes a function with retry logic and returns a result
func (r *Retrier) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) error {
	r.mu.Lock()
	r.metrics.TotalAttempts++
	r.mu.Unlock()

	var lastErr error
	maxAttempts := r.policy.GetMaxAttempts()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check circuit breaker if configured
		if r.circuitBreaker != nil {
			if err := r.circuitBreaker.Execute(ctx, func() error {
				_, lastErr = fn()
				return lastErr
			}); err != nil {
				if err.Error() == "circuit breaker is open" {
					r.incrementCircuitBreaks()
					r.incrementFailedCalls()
					return err
				}
			}
		} else {
			_, lastErr = fn()
		}

		// Success
		if lastErr == nil {
			r.incrementSuccessfulCalls()
			r.updateAverageAttempts(float64(attempt))
			return nil
		}

		// Check if we should retry
		if !r.policy.ShouldRetry(attempt, lastErr) {
			break
		}

		// Last attempt, don't wait
		if attempt == maxAttempts {
			break
		}

		// Wait before next attempt
		delay := r.policy.GetDelay(attempt)
		select {
		case <-ctx.Done():
			r.incrementFailedCalls()
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	r.incrementFailedCalls()
	r.updateAverageAttempts(float64(maxAttempts))
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}

// GetMetrics returns current retry metrics
func (r *Retrier) GetMetrics() RetryMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.metrics
}

// ResetMetrics resets all metrics to zero
func (r *Retrier) ResetMetrics() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = &RetryMetrics{}
}

// Helper methods for metrics
func (r *Retrier) incrementSuccessfulCalls() {
	r.mu.Lock()
	r.metrics.SuccessfulCalls++
	r.mu.Unlock()
}

func (r *Retrier) incrementFailedCalls() {
	r.mu.Lock()
	r.metrics.FailedCalls++
	r.mu.Unlock()
}

func (r *Retrier) incrementCircuitBreaks() {
	r.mu.Lock()
	r.metrics.CircuitBreaks++
	r.mu.Unlock()
}

func (r *Retrier) updateAverageAttempts(attempts float64) {
	r.mu.Lock()
	totalCalls := r.metrics.SuccessfulCalls + r.metrics.FailedCalls
	if totalCalls > 0 {
		r.metrics.AverageAttempts = (r.metrics.AverageAttempts*float64(totalCalls-1) + attempts) / float64(totalCalls)
	}
	r.mu.Unlock()
}