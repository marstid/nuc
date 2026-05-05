package nucleus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"
)

// ListAssets returns assets matching the given options.
func (c *Client) ListAssets(ctx context.Context, projectID string, opts *domain.AssetListOptions) ([]domain.Asset, error) {
	params := buildAssetListParams(opts)

	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/assets", projectID), params)
	if err != nil {
		return nil, fmt.Errorf("listing assets for project %s: %w", projectID, err)
	}

	var assets []domain.Asset
	if err := json.Unmarshal(body, &assets); err != nil {
		return nil, fmt.Errorf("decoding assets response: %w", err)
	}

	return assets, nil
}

// GetAsset returns a specific asset by ID.
func (c *Client) GetAsset(ctx context.Context, projectID, assetID string) (*domain.Asset, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/assets/%s", projectID, assetID), nil)
	if err != nil {
		return nil, fmt.Errorf("getting asset %s in project %s: %w", assetID, projectID, err)
	}

	var asset domain.Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		return nil, fmt.Errorf("decoding asset response: %w", err)
	}

	return &asset, nil
}

// CreateAsset creates a new asset in the project.
func (c *Client) CreateAsset(ctx context.Context, projectID string, input *service.CreateAssetInput) (*domain.Asset, error) {
	body, err := c.post(ctx, fmt.Sprintf("/projects/%s/assets", projectID), input)
	if err != nil {
		return nil, fmt.Errorf("creating asset in project %s: %w", projectID, err)
	}

	var asset domain.Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		return nil, fmt.Errorf("decoding created asset response: %w", err)
	}

	return &asset, nil
}

// UpdateAsset modifies an existing asset.
func (c *Client) UpdateAsset(ctx context.Context, projectID, assetID string, input *service.UpdateAssetInput) (*domain.Asset, error) {
	body, err := c.put(ctx, fmt.Sprintf("/projects/%s/assets/%s", projectID, assetID), input)
	if err != nil {
		return nil, fmt.Errorf("updating asset %s in project %s: %w", assetID, projectID, err)
	}

	var asset domain.Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		return nil, fmt.Errorf("decoding updated asset response: %w", err)
	}

	return &asset, nil
}

// DeleteAsset removes an asset from the project.
func (c *Client) DeleteAsset(ctx context.Context, projectID, assetID string) error {
	if err := c.delete(ctx, fmt.Sprintf("/projects/%s/assets/%s", projectID, assetID), nil); err != nil {
		return fmt.Errorf("deleting asset %s in project %s: %w", assetID, projectID, err)
	}
	return nil
}

// ListAssetGroups returns all asset groups in a project.
func (c *Client) ListAssetGroups(ctx context.Context, projectID string) ([]domain.AssetGroup, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/assets/groups", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("listing asset groups for project %s: %w", projectID, err)
	}

	var groups []domain.AssetGroup
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("decoding asset groups response: %w", err)
	}

	return groups, nil
}

// CreateAssetGroup creates a new asset group.
func (c *Client) CreateAssetGroup(ctx context.Context, projectID, name string) error {
	payload := map[string]string{"asset_group": name}

	if _, err := c.post(ctx, fmt.Sprintf("/projects/%s/assets/groups", projectID), payload); err != nil {
		return fmt.Errorf("creating asset group %q in project %s: %w", name, projectID, err)
	}
	return nil
}

// DeleteAssetGroup removes an asset group by name.
// The API expects the group name as a query parameter: DELETE /assets/groups?asset_group=<name>
func (c *Client) DeleteAssetGroup(ctx context.Context, projectID, name string) error {
	params := url.Values{}
	params.Set("asset_group", name)
	if err := c.delete(ctx, fmt.Sprintf("/projects/%s/assets/groups", projectID), params); err != nil {
		return fmt.Errorf("deleting asset group %q in project %s: %w", name, projectID, err)
	}
	return nil
}

// buildAssetListParams converts AssetListOptions to URL query parameters.
func buildAssetListParams(opts *domain.AssetListOptions) url.Values {
	if opts == nil {
		return nil
	}

	params := url.Values{}

	if opts.Start != nil {
		params.Set("start", strconv.Itoa(*opts.Start))
	}
	if opts.Limit != nil {
		params.Set("limit", strconv.Itoa(*opts.Limit))
	}
	if opts.IPAddress != "" {
		params.Set("ip_address", opts.IPAddress)
	}
	if opts.AssetName != "" {
		params.Set("asset_name", opts.AssetName)
	}
	if opts.AssetGroups != "" {
		params.Set("asset_groups", opts.AssetGroups)
	}
	if opts.AssetType != "" {
		params.Set("asset_type", string(opts.AssetType))
	}
	if opts.InactiveAssets != nil {
		params.Set("inactive_assets", strconv.FormatBool(*opts.InactiveAssets))
	}
	if opts.UnscannedAssets != nil {
		params.Set("unscanned_assets", strconv.FormatBool(*opts.UnscannedAssets))
	}
	if opts.AssetsWithFindings != nil {
		params.Set("assets_with_findings", strconv.FormatBool(*opts.AssetsWithFindings))
	}

	if len(params) == 0 {
		return nil
	}
	return params
}
