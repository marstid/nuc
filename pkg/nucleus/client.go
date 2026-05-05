// Package nucleus implements the Nucleus Security API client.
// It provides HTTP-based implementations of the service interfaces
// with built-in retry logic and circuit breaker support.
package nucleus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the Nucleus Security API client.
// It implements all service interfaces (ProjectService, AssetService, etc.)
// and handles HTTP communication, authentication, and resilience.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Nucleus Security API client.
func NewClient(baseURL, apiKey string, opts ...TransportOption) *Client {
	// Ensure baseURL doesn't have trailing slash.
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient := newHTTPClient(apiKey, opts...)

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// get performs a GET request to the given path and returns the response body.
func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u := c.buildURL(path, params)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req)
}

// post performs a POST request with a JSON body.
func (c *Client) post(ctx context.Context, path string, body any) ([]byte, error) {
	return c.postWithParams(ctx, path, nil, body)
}

// postWithParams performs a POST request with query parameters and a JSON body.
func (c *Client) postWithParams(ctx context.Context, path string, params url.Values, body any) ([]byte, error) {
	u := c.buildURL(path, params)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBytes))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req)
}

// put performs a PUT request with a JSON body.
func (c *Client) put(ctx context.Context, path string, body any) ([]byte, error) {
	u := c.buildURL(path, nil)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBytes))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, path string, params url.Values) error {
	u := c.buildURL(path, params)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	_, err = c.doRequest(req)
	return err
}

// doRequest executes the request and handles the response.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on response body

	if resp.StatusCode >= 400 {
		return nil, handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, nil
}

// buildURL constructs the full URL for an API endpoint.
func (c *Client) buildURL(path string, params url.Values) string {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}
