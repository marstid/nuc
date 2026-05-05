package nucleus

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// ResilientTransport wraps an http.RoundTripper with authentication,
// retry logic, and circuit breaker support.
type ResilientTransport struct {
	base           http.RoundTripper
	apiKey         string
	retry          *RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewResilientTransport creates a transport with auth, retry, and circuit breaker.
func NewResilientTransport(base http.RoundTripper, apiKey string, retryCfg *RetryConfig, cbCfg *CircuitBreakerConfig) *ResilientTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if retryCfg == nil {
		retryCfg = DefaultRetryConfig()
	}

	return &ResilientTransport{
		base:           base,
		apiKey:         apiKey,
		retry:          retryCfg,
		circuitBreaker: NewCircuitBreaker(cbCfg),
	}
}

// RoundTrip implements http.RoundTripper with resilience patterns.
func (t *ResilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check circuit breaker before making any request.
	if err := t.circuitBreaker.Allow(); err != nil {
		return nil, err
	}

	// Add authorization header — Nucleus uses x-apikey, not Bearer.
	req.Header.Set("x-apikey", t.apiKey)
	req.Header.Set("Accept", "application/json")
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Buffer the request body for replay on retry.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close() //nolint:errcheck // best-effort close when buffering request body
	}

	ctx := req.Context()

	resp, err := doWithRetry(ctx, t.retry, func() (*http.Response, error) {
		// Recreate the request body for each attempt.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		return t.base.RoundTrip(req)
	})

	if err != nil {
		t.circuitBreaker.RecordFailure()
		return nil, err
	}

	// Record result for circuit breaker.
	if resp.StatusCode >= 500 {
		t.circuitBreaker.RecordFailure()
	} else {
		t.circuitBreaker.RecordSuccess()
	}

	return resp, nil
}

// TransportOption is a functional option for configuring the resilient transport.
type TransportOption func(*transportOptions)

type transportOptions struct {
	timeout  time.Duration
	retryCfg *RetryConfig
	cbCfg    *CircuitBreakerConfig
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) TransportOption {
	return func(o *transportOptions) {
		o.timeout = d
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(cfg *RetryConfig) TransportOption {
	return func(o *transportOptions) {
		o.retryCfg = cfg
	}
}

// WithCircuitBreakerConfig sets the circuit breaker configuration.
func WithCircuitBreakerConfig(cfg *CircuitBreakerConfig) TransportOption {
	return func(o *transportOptions) {
		o.cbCfg = cfg
	}
}

// newHTTPClient creates an http.Client with the resilient transport.
func newHTTPClient(apiKey string, opts ...TransportOption) *http.Client {
	options := &transportOptions{
		timeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	transport := NewResilientTransport(
		http.DefaultTransport,
		apiKey,
		options.retryCfg,
		options.cbCfg,
	)

	return &http.Client{
		Transport: transport,
		Timeout:   options.timeout,
	}
}
