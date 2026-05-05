package nucleus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/marstid/nuc/pkg/domain"
)

// GetFindingMetrics retrieves aggregated discovery/remediation metrics for a project.
// Corresponds to GET /projects/{project_id}/findings/metrics.
func (c *Client) GetFindingMetrics(ctx context.Context, projectID string) (*domain.FindingMetrics, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/findings/metrics", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("getting finding metrics for project %s: %w", projectID, err)
	}

	var metrics domain.FindingMetrics
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("decoding finding metrics response: %w", err)
	}

	return &metrics, nil
}

// GetAssetGroupMetrics retrieves security metrics for one or more asset groups.
// Corresponds to GET /projects/{project_id}/assets/groups/metrics.
// opts.AssetGroups is required (up to 50 groups); opts.Metrics is optional.
func (c *Client) GetAssetGroupMetrics(ctx context.Context, projectID string, opts *domain.AssetGroupMetricsOptions) ([]domain.AssetGroupMetrics, error) {
	if opts == nil || len(opts.AssetGroups) == 0 {
		return nil, fmt.Errorf("at least one asset group is required")
	}

	// The API expects JSON-encoded arrays as query parameters.
	groupsJSON, err := json.Marshal(opts.AssetGroups)
	if err != nil {
		return nil, fmt.Errorf("encoding asset_groups: %w", err)
	}

	params := url.Values{}
	params.Set("asset_groups", string(groupsJSON))

	if len(opts.Metrics) > 0 {
		metricsJSON, err := json.Marshal(opts.Metrics)
		if err != nil {
			return nil, fmt.Errorf("encoding metrics: %w", err)
		}
		params.Set("metrics", string(metricsJSON))
	}

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/assets/groups/metrics", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("getting asset group metrics for project %s: %w", projectID, err)
	}

	// The API returns a map keyed by group name: {"/<group>": { risk_score: 654, ... }, ...}
	var raw map[string]domain.AssetGroupMetrics
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding asset group metrics response: %w", err)
	}

	// Preserve the requested ordering.
	result := make([]domain.AssetGroupMetrics, 0, len(raw))
	for _, group := range opts.AssetGroups {
		if m, ok := raw[group]; ok {
			m.GroupName = group
			result = append(result, m)
		}
	}
	// Append any groups returned that weren't in the request (defensive).
	requested := make(map[string]struct{}, len(opts.AssetGroups))
	for _, g := range opts.AssetGroups {
		requested[g] = struct{}{}
	}
	for name, m := range raw {
		if _, ok := requested[name]; !ok {
			m.GroupName = name
			result = append(result, m)
		}
	}

	return result, nil
}
