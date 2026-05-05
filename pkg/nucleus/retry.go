package nucleus

import (
	"context"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/marstid/nuc/pkg/domain"
)

// RetryConfig defines the retry policy for HTTP requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (excluding the initial request).
	MaxRetries int

	// InitialBackoff is the delay before the first retry.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum delay between retries.
	MaxBackoff time.Duration

	// Multiplier is the factor by which the backoff increases each attempt.
	Multiplier float64

	// Jitter is the randomization factor (0.0 to 1.0) applied to backoff.
	Jitter float64
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.2,
	}
}

// backoff calculates the backoff duration for the given attempt (0-indexed).
func (c *RetryConfig) backoff(attempt int) time.Duration {
	backoff := float64(c.InitialBackoff) * math.Pow(c.Multiplier, float64(attempt))
	if backoff > float64(c.MaxBackoff) {
		backoff = float64(c.MaxBackoff)
	}

	// Apply jitter: ±Jitter%
	if c.Jitter > 0 {
		jitterRange := backoff * c.Jitter
		backoff = backoff - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	return time.Duration(backoff)
}

// isRetryableStatusCode returns true if the HTTP status code is retryable.
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// retryAfterDuration parses the Retry-After header value.
// It supports both seconds (integer) and HTTP-date formats.
func retryAfterDuration(header string) time.Duration {
	if header == "" {
		return 0
	}

	// Try parsing as seconds first.
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date.
	if t, err := http.ParseTime(header); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return 0
}

// doWithRetry executes the given function with retry logic.
// It respects context cancellation and the Retry-After header.
func doWithRetry(ctx context.Context, cfg *RetryConfig, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before each attempt.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := fn()
		if err == nil && !isRetryableStatusCode(resp.StatusCode) {
			return resp, nil
		}

		// Determine if we should retry.
		if attempt == cfg.MaxRetries {
			if err != nil {
				lastErr = err
			} else {
				return resp, nil
			}
			break
		}

		// For HTTP errors (connection failures, timeouts), always retry.
		if err != nil {
			lastErr = err
		} else if isRetryableStatusCode(resp.StatusCode) {
			// Check for Retry-After header on 429.
			if resp.StatusCode == http.StatusTooManyRequests {
				if retryAfter := retryAfterDuration(resp.Header.Get("Retry-After")); retryAfter > 0 {
					if err := sleepWithContext(ctx, retryAfter); err != nil {
						return nil, err
					}
					continue
				}
			}
			// Close body before retry to prevent resource leaks.
			resp.Body.Close() //nolint:errcheck // best-effort close on retried response body
			lastErr = &domain.APIError{
				StatusCode: resp.StatusCode,
				Message:    http.StatusText(resp.StatusCode),
			}
		} else {
			// Non-retryable response — return it immediately.
			return resp, nil
		}

		// Sleep with exponential backoff.
		backoff := cfg.backoff(attempt)
		if err := sleepWithContext(ctx, backoff); err != nil {
			return nil, err
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, domain.ErrRetryExhausted
}

// sleepWithContext sleeps for the specified duration, returning early if the context is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
