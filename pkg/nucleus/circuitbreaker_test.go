package nucleus

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/marstid/nuc/pkg/domain"
)

func newTestCB(failureThreshold, successThreshold int, openTimeout time.Duration) *CircuitBreaker {
	return NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		OpenTimeout:      openTimeout,
	})
}

func TestCircuitBreaker_StartsInClosedState(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	assert.Equal(t, CircuitClosed, cb.State())
	assert.NoError(t, cb.Allow())
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := newTestCB(3, 2, 30*time.Second)

	// Record failures below threshold.
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())
	assert.NoError(t, cb.Allow())

	// Third failure should open the circuit.
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	cb := newTestCB(2, 1, 1*time.Hour)

	// Open the circuit.
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())

	// Requests should be rejected.
	err := cb.Allow()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCircuitOpen)
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := newTestCB(2, 1, 100*time.Millisecond)

	// Use controllable time.
	now := time.Now()
	cb.now = func() time.Time { return now }

	// Open the circuit.
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())

	// Advance time past the open timeout.
	now = now.Add(150 * time.Millisecond)
	assert.Equal(t, CircuitHalfOpen, cb.State())

	// Should allow one request through.
	assert.NoError(t, cb.Allow())
}

func TestCircuitBreaker_ClosesOnSuccessInHalfOpen(t *testing.T) {
	cb := newTestCB(2, 2, 100*time.Millisecond)

	now := time.Now()
	cb.now = func() time.Time { return now }

	// Open the circuit.
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())

	// Transition to half-open.
	now = now.Add(150 * time.Millisecond)
	assert.Equal(t, CircuitHalfOpen, cb.State())

	// First success in half-open — still half-open (need 2).
	cb.RecordSuccess()
	assert.Equal(t, CircuitHalfOpen, cb.State())

	// Second success — should close.
	cb.RecordSuccess()
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	cb := newTestCB(2, 2, 100*time.Millisecond)

	now := time.Now()
	cb.now = func() time.Time { return now }

	// Open the circuit.
	cb.RecordFailure()
	cb.RecordFailure()

	// Transition to half-open.
	now = now.Add(150 * time.Millisecond)
	assert.Equal(t, CircuitHalfOpen, cb.State())

	// Failure in half-open should re-open.
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_SuccessResetFailureCount(t *testing.T) {
	cb := newTestCB(3, 1, 30*time.Second)

	// Record some failures.
	cb.RecordFailure()
	cb.RecordFailure()

	// Success should reset failure count.
	cb.RecordSuccess()
	assert.Equal(t, CircuitClosed, cb.State())

	// Need 3 fresh failures to open again.
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())

	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := newTestCB(100, 2, 30*time.Second)

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent failures.
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cb.RecordFailure()
		}()
	}
	wg.Wait()

	// Concurrent Allow calls.
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = cb.Allow()
		}()
	}
	wg.Wait()

	// Concurrent successes.
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cb.RecordSuccess()
		}()
	}
	wg.Wait()

	// Should not panic — state may vary but must be consistent.
	state := cb.State()
	assert.Contains(t, []CircuitState{CircuitClosed, CircuitOpen, CircuitHalfOpen}, state)
}
