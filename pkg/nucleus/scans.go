package nucleus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/marstid/nuc/pkg/domain"
)

// ListScans returns scans for a given project with optional pagination.
// Pagination is controlled via start/limit query parameters (API default: 1, max: 100).
func (c *Client) ListScans(ctx context.Context, projectID string, start, limit int) ([]domain.Scan, error) {
	params := url.Values{}
	if start > 0 {
		params.Set("start", fmt.Sprintf("%d", start))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/scans", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("listing scans for project %s: %w", projectID, err)
	}

	var scans []domain.Scan
	if err := json.Unmarshal(body, &scans); err != nil {
		return nil, fmt.Errorf("decoding scans response: %w", err)
	}

	return scans, nil
}

// UploadScan uploads a scan file to a project via multipart/form-data.
func (c *Client) UploadScan(ctx context.Context, projectID, filePath string, opts *domain.ScanUploadOptions) (*domain.ScanResult, error) {
	file, err := os.Open(filePath) //nolint:gosec // G304: filePath is supplied by the CLI caller, not external user input
	if err != nil {
		return nil, fmt.Errorf("opening scan file: %w", err)
	}
	defer file.Close() //nolint:errcheck // best-effort close on read-only file

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copying file content: %w", err)
	}

	if opts != nil {
		if opts.Description != "" {
			if err := writer.WriteField("description", opts.Description); err != nil {
				return nil, fmt.Errorf("writing description field: %w", err)
			}
		}
		if opts.Type != "" {
			if err := writer.WriteField("scan_type", opts.Type); err != nil {
				return nil, fmt.Errorf("writing scan_type field: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	u := c.buildURL(fmt.Sprintf("/projects/%s/scans", projectID), nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("uploading scan for project %s: %w", projectID, err)
	}

	var result domain.ScanResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding scan upload response: %w", err)
	}

	return &result, nil
}
