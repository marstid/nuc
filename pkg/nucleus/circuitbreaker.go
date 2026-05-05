package nucleus

import (
	"sync"
	"time"

	"github.com/marstid/nuc/pkg/domain"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// CircuitClosed is the normal operating state — requests pass through.
	CircuitClosed CircuitState = iota

	// CircuitOpen is the tripped state — requests fail immediately.
	CircuitOpen

	// CircuitHalfOpen is the probing state — one request allowed to test recovery.
	CircuitHalfOpen
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig defines the configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before the circuit opens.
	FailureThreshold int

	// SuccessThreshold is the number of consecutive successes in half-open state
	// needed to close the circuit.
	SuccessThreshold int

	// OpenTimeout is the duration the circuit stays open before transitioning to half-open.
	OpenTimeout time.Duration
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenTimeout:      30 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
type CircuitBreaker struct {
	mu     sync.Mutex
	config *CircuitBreakerConfig

	state        CircuitState
	failureCount int
	successCount int
	lastFailure  time.Time

	// now is a function returning the current time. Overridable for testing.
	now func() time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(cfg *CircuitBreakerConfig) *CircuitBreaker {
	if cfg == nil {
		cfg = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		config: cfg,
		state:  CircuitClosed,
		now:    time.Now,
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// Allow checks if a request is allowed through the circuit breaker.
// Returns nil if allowed, or domain.ErrCircuitOpen if the circuit is open.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.currentState()
	switch state {
	case CircuitClosed:
		return nil
	case CircuitHalfOpen:
		return nil
	case CircuitOpen:
		return domain.ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.currentState() {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.setState(CircuitClosed)
		}
	case CircuitClosed:
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.currentState() {
	case CircuitClosed:
		cb.failureCount++
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.setState(CircuitOpen)
			cb.lastFailure = cb.now()
		}
	case CircuitHalfOpen:
		// Any failure in half-open state re-opens the circuit.
		cb.setState(CircuitOpen)
		cb.lastFailure = cb.now()
	}
}

// currentState returns the actual state, including time-based transitions.
// Must be called with the mutex held.
func (cb *CircuitBreaker) currentState() CircuitState {
	if cb.state == CircuitOpen {
		// Check if we should transition to half-open.
		if cb.now().Sub(cb.lastFailure) >= cb.config.OpenTimeout {
			cb.setState(CircuitHalfOpen)
			return CircuitHalfOpen
		}
	}
	return cb.state
}

// setState transitions the circuit breaker to a new state, resetting counters.
// Must be called with the mutex held.
func (cb *CircuitBreaker) setState(state CircuitState) {
	cb.state = state
	cb.failureCount = 0
	cb.successCount = 0
}
