package nucleus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry_SucceedsOnFirstAttempt(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := DefaultRetryConfig()
	resp, err := doWithRetry(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetry_SucceedsAfterTransientFailure(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0,
	}

	resp, err := doWithRetry(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetry_RespectsRetryAfterHeader(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		Jitter:         0,
	}

	start := time.Now()
	resp, err := doWithRetry(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	// Should have waited at least ~1 second due to Retry-After.
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond)
}

func TestRetry_ExhaustsMaxRetries(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0,
	}

	resp, err := doWithRetry(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})

	// When max retries exhausted with server errors, the last response is returned.
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	// Initial attempt + 2 retries = 3 total.
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetry_DoesNotRetryClientErrors(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := DefaultRetryConfig()
	resp, err := doWithRetry(context.Background(), cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := &RetryConfig{
		MaxRetries:     10,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
		Jitter:         0,
	}

	_, err := doWithRetry(ctx, cfg, func() (*http.Response, error) {
		return http.Get(srv.URL)
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	// Should have attempted at most a couple of times before context deadline.
	assert.LessOrEqual(t, atomic.LoadInt32(&attempts), int32(3))
}

func TestRetry_ExponentialBackoffTiming(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
		Jitter:         0, // No jitter for deterministic test.
	}

	// Verify backoff calculations.
	assert.Equal(t, 100*time.Millisecond, cfg.backoff(0))
	assert.Equal(t, 200*time.Millisecond, cfg.backoff(1))
	assert.Equal(t, 400*time.Millisecond, cfg.backoff(2))
	assert.Equal(t, 800*time.Millisecond, cfg.backoff(3))
	assert.Equal(t, 1600*time.Millisecond, cfg.backoff(4))
	// Should be capped at MaxBackoff.
	assert.Equal(t, 2*time.Second, cfg.backoff(5))
	assert.Equal(t, 2*time.Second, cfg.backoff(10))
}

func TestRetryAfterDuration_Seconds(t *testing.T) {
	d := retryAfterDuration("5")
	assert.Equal(t, 5*time.Second, d)
}

func TestRetryAfterDuration_Empty(t *testing.T) {
	d := retryAfterDuration("")
	assert.Equal(t, time.Duration(0), d)
}

func TestRetryAfterDuration_InvalidString(t *testing.T) {
	d := retryAfterDuration("not-a-number")
	assert.Equal(t, time.Duration(0), d)
}
