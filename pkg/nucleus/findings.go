package nucleus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"
)

// ListFindings returns findings matching the given options.
func (c *Client) ListFindings(ctx context.Context, projectID string, opts *domain.FindingListOptions) ([]domain.Finding, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Start != nil {
			params.Set("start", fmt.Sprintf("%d", *opts.Start))
		}
		if opts.Limit != nil {
			params.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
		if opts.Severity != "" {
			params.Set("finding_severity", string(opts.Severity))
		}
		if opts.Status != "" {
			params.Set("finding_status", string(opts.Status))
		}
	}

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("listing findings for project %s: %w", projectID, err)
	}

	var findings []domain.Finding
	if err := json.Unmarshal(body, &findings); err != nil {
		return nil, fmt.Errorf("decoding findings response: %w", err)
	}

	return findings, nil
}

// GetFinding returns detailed information about a specific finding.
func (c *Client) GetFinding(ctx context.Context, projectID, findingNumber string) (*domain.Finding, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/%s", projectID, findingNumber), nil)
	if err != nil {
		return nil, fmt.Errorf("getting finding %s for project %s: %w", findingNumber, projectID, err)
	}

	var finding domain.Finding
	if err := json.Unmarshal(body, &finding); err != nil {
		return nil, fmt.Errorf("decoding finding response: %w", err)
	}

	return &finding, nil
}

// SearchFindings performs a filtered search for findings using the FindingsSearch criteria.
// The search body is posted directly as a flat JSON object (no wrapper).
// Pagination is controlled via start/limit query parameters (API default: 100, max: 1000).
func (c *Client) SearchFindings(ctx context.Context, projectID string, search *domain.FindingSearch, start, limit int) ([]domain.Finding, error) {
	params := url.Values{}
	if start > 0 {
		params.Set("start", fmt.Sprintf("%d", start))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := c.postWithParams(ctx, fmt.Sprintf("/projects/%s/findings/search", projectID), params, search)
	if err != nil {
		return nil, fmt.Errorf("searching findings for project %s: %w", projectID, err)
	}

	var findings []domain.Finding
	if err := json.Unmarshal(body, &findings); err != nil {
		return nil, fmt.Errorf("decoding search findings response: %w", err)
	}

	return findings, nil
}

// UpdateFinding modifies a finding's status, severity, or other attributes.
func (c *Client) UpdateFinding(ctx context.Context, projectID string, input *service.UpdateFindingInput) error {
	_, err := c.put(ctx, fmt.Sprintf("/projects/%s/findings", projectID), input)
	if err != nil {
		return fmt.Errorf("updating finding for project %s: %w", projectID, err)
	}

	return nil
}

// BulkUpdateFindings modifies multiple findings at once.
func (c *Client) BulkUpdateFindings(ctx context.Context, projectID string, updates []service.UpdateFindingInput) error {
	payload := struct {
		Updates []service.UpdateFindingInput `json:"updates"`
	}{
		Updates: updates,
	}

	_, err := c.put(ctx, fmt.Sprintf("/projects/%s/findings/bulk", projectID), payload)
	if err != nil {
		return fmt.Errorf("bulk updating findings for project %s: %w", projectID, err)
	}

	return nil
}

// GetMitigatedFindings returns findings that have been mitigated.
func (c *Client) GetMitigatedFindings(ctx context.Context, projectID string, opts *domain.MitigatedOptions) ([]domain.MitigatedFinding, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Start != nil {
			params.Set("start", fmt.Sprintf("%d", *opts.Start))
		}
		if opts.Limit != nil {
			params.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
		if opts.StartDate != "" {
			params.Set("start_date", opts.StartDate)
		}
	}

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/mitigated", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("getting mitigated findings for project %s: %w", projectID, err)
	}

	var findings []domain.MitigatedFinding
	if err := json.Unmarshal(body, &findings); err != nil {
		return nil, fmt.Errorf("decoding mitigated findings response: %w", err)
	}

	return findings, nil
}

// GetFindingTrend returns trend data for findings over time.
func (c *Client) GetFindingTrend(ctx context.Context, projectID string, opts *domain.TrendOptions) (*domain.FindingTrend, error) {
	params := url.Values{}
	if opts != nil {
		if opts.StartDate != "" {
			params.Set("start_date", opts.StartDate)
		}
		if opts.EndDate != "" {
			params.Set("end_date", opts.EndDate)
		}
		if len(opts.AssetGroups) > 0 {
			params.Set("asset_groups", strings.Join(opts.AssetGroups, ","))
		}
	}

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/trend", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("getting finding trend for project %s: %w", projectID, err)
	}

	var trend domain.FindingTrend
	if err := json.Unmarshal(body, &trend); err != nil {
		return nil, fmt.Errorf("decoding finding trend response: %w", err)
	}

	return &trend, nil
}

// GetFindingOverview returns a summary overview of findings.
func (c *Client) GetFindingOverview(ctx context.Context, projectID string) (*domain.FindingOverview, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/overview", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("getting finding overview for project %s: %w", projectID, err)
	}

	var overview domain.FindingOverview
	if err := json.Unmarshal(body, &overview); err != nil {
		return nil, fmt.Errorf("decoding finding overview response: %w", err)
	}

	return &overview, nil
}

// GetFindingsSummary returns findings grouped by finding_number with optional filters and sorting.
// Pagination uses query params start/limit (API max: 100 per page).
func (c *Client) GetFindingsSummary(ctx context.Context, projectID string, req *domain.FindingSummaryRequest, start, limit int) ([]domain.FindingSummary, error) {
	params := url.Values{}
	if start > 0 {
		params.Set("start", fmt.Sprintf("%d", start))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := c.postWithParams(ctx, fmt.Sprintf("/projects/%s/findings/summary", projectID), params, req)
	if err != nil {
		return nil, fmt.Errorf("getting findings summary for project %s: %w", projectID, err)
	}

	var findings []domain.FindingSummary
	if err := json.Unmarshal(body, &findings); err != nil {
		return nil, fmt.Errorf("decoding findings summary response: %w", err)
	}

	return findings, nil
}

// GetFrameworks returns compliance frameworks associated with findings.
func (c *Client) GetFrameworks(ctx context.Context, projectID string) ([]string, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/frameworks", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("getting frameworks for project %s: %w", projectID, err)
	}

	var frameworks []string
	if err := json.Unmarshal(body, &frameworks); err != nil {
		return nil, fmt.Errorf("decoding frameworks response: %w", err)
	}

	return frameworks, nil
}
