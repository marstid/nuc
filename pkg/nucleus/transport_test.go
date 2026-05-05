package nucleus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/marstid/nuc/pkg/domain"
)

func TestTransport_AddsAuthHeader(t *testing.T) {
	var capturedAPIKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAPIKey = r.Header.Get("x-apikey")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	transport := NewResilientTransport(http.DefaultTransport, "test-api-key", nil, nil)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/test", nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "test-api-key", capturedAPIKey)
}

func TestTransport_SetsContentTypeForBody(t *testing.T) {
	var capturedContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	transport := NewResilientTransport(http.DefaultTransport, "key", nil, nil)
	client := &http.Client{Transport: transport}

	body := strings.NewReader(`{"foo":"bar"}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/test", body)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "application/json", capturedContentType)
}

func TestTransport_CombinesCircuitBreakerAndRetry(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	retryCfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0,
	}
	cbCfg := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenTimeout:      1 * time.Hour,
	}

	transport := NewResilientTransport(http.DefaultTransport, "key", retryCfg, cbCfg)
	client := &http.Client{Transport: transport}

	// First request: 3 attempts (1 + 2 retries), all 500 → circuit records failure.
	req1, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp1, err := client.Do(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp1.StatusCode)

	// Second request: another set of 500s → circuit opens.
	req2, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	resp2.Body.Close()

	// Third request: circuit should be open, fail immediately.
	req3, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	_, err = client.Do(req3)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCircuitOpen)
}

func TestTransport_RespectsTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newHTTPClient("key",
		WithTimeout(100*time.Millisecond),
		WithRetryConfig(&RetryConfig{
			MaxRetries:     0, // No retries.
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
			Multiplier:     1,
			Jitter:         0,
		}),
	)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	_, err := client.Do(req)

	require.Error(t, err)
	assert.True(t, os.IsTimeout(err) || strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "Timeout exceeded"),
		"expected timeout error, got: %v", err)
}

func TestTransport_RecordsSuccessForCircuitBreaker(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cbCfg := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenTimeout:      30 * time.Second,
	}
	transport := NewResilientTransport(http.DefaultTransport, "key", nil, cbCfg)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, CircuitClosed, transport.circuitBreaker.State())
}
