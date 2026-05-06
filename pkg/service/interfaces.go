// Package service defines the application-layer interfaces consumed by delivery adapters
// (CLI, MCP server). These interfaces are delivery-agnostic and backed by the nucleus/ package.
package service

import (
	"context"

	"github.com/marstid/nuc/pkg/domain"
)

// ProjectService defines operations on Nucleus Security projects.
type ProjectService interface {
	// List returns all projects accessible to the authenticated user.
	List(ctx context.Context) ([]domain.Project, error)

	// Get returns a specific project by ID.
	Get(ctx context.Context, projectID string) (*domain.Project, error)

	// GetRiskScore returns the risk score for a project.
	GetRiskScore(ctx context.Context, projectID string) (*domain.RiskScore, error)
}

// AssetService defines operations on assets within a project.
type AssetService interface {
	// List returns assets matching the given options.
	List(ctx context.Context, projectID string, opts *domain.AssetListOptions) ([]domain.Asset, error)

	// Get returns a specific asset by ID.
	Get(ctx context.Context, projectID, assetID string) (*domain.Asset, error)

	// Create creates a new asset in the project.
	Create(ctx context.Context, projectID string, input *CreateAssetInput) (*domain.Asset, error)

	// Update modifies an existing asset.
	Update(ctx context.Context, projectID, assetID string, input *UpdateAssetInput) (*domain.Asset, error)

	// Delete removes an asset from the project.
	Delete(ctx context.Context, projectID, assetID string) error

	// ListGroups returns all asset groups in a project.
	ListGroups(ctx context.Context, projectID string) ([]domain.AssetGroup, error)

	// CreateGroup creates a new asset group.
	CreateGroup(ctx context.Context, projectID, name string) error

	// DeleteGroup removes an asset group.
	DeleteGroup(ctx context.Context, projectID, name string) error

	// GetGroupMetrics returns metrics for specified asset groups.
	GetGroupMetrics(ctx context.Context, projectID string, opts *domain.AssetGroupMetricsOptions) ([]domain.AssetGroupMetrics, error)

	// ListTeams returns all teams in a project.
	ListTeams(ctx context.Context, projectID string) ([]domain.Team, error)
}

// FindingService defines operations on vulnerability findings.
type FindingService interface {
	// List returns findings matching the given options.
	List(ctx context.Context, projectID string, opts *domain.FindingListOptions) ([]domain.Finding, error)

	// Get returns detailed information about a specific finding.
	Get(ctx context.Context, projectID, findingNumber string) (*domain.Finding, error)

	// Search performs a filtered search for findings using FindingSearch criteria.
	// Pagination is controlled via start/limit (API default: 100, max: 1000).
	Search(ctx context.Context, projectID string, search *domain.FindingSearch, start, limit int) ([]domain.Finding, error)

	// Update modifies a finding's status, severity, or other attributes.
	Update(ctx context.Context, projectID string, input *UpdateFindingInput) error

	// BulkUpdate modifies multiple findings at once.
	BulkUpdate(ctx context.Context, projectID string, updates []UpdateFindingInput) error

	// GetMitigated returns findings that have been mitigated.
	GetMitigated(ctx context.Context, projectID string, opts *domain.MitigatedOptions) ([]domain.MitigatedFinding, error)

	// GetTrend returns trend data for findings over time.
	GetTrend(ctx context.Context, projectID string, opts *domain.TrendOptions) (*domain.FindingTrend, error)

	// GetOverview returns a summary overview of findings.
	GetOverview(ctx context.Context, projectID string) (*domain.FindingOverview, error)

	// GetFrameworks returns compliance frameworks associated with findings.
	GetFrameworks(ctx context.Context, projectID string) ([]string, error)
}

// ScanService defines operations on vulnerability scans.
type ScanService interface {
	// List returns scans in a project with optional pagination.
	// Pagination is controlled via start/limit (API default: 1, max: 100).
	List(ctx context.Context, projectID string, start, limit int) ([]domain.Scan, error)

	// Upload uploads a scan file to a project.
	Upload(ctx context.Context, projectID, filePath string, opts *domain.ScanUploadOptions) (*domain.ScanResult, error)
}

// MetricsService defines operations for fetching vulnerability metrics.
type MetricsService interface {
	// GetFindingMetrics returns aggregated discovery/remediation metrics over 30/90/180-day windows.
	GetFindingMetrics(ctx context.Context, projectID string) (*domain.FindingMetrics, error)

	// GetAssetGroupMetrics returns security metrics for one or more asset groups.
	GetAssetGroupMetrics(ctx context.Context, projectID string, opts *domain.AssetGroupMetricsOptions) ([]domain.AssetGroupMetrics, error)
}

// CreateAssetInput contains the data needed to create a new asset.
type CreateAssetInput struct {
	Name                   string           `json:"asset_name"`
	Type                   domain.AssetType `json:"asset_type"`
	IPAddress              string           `json:"ip_address,omitempty"`
	DomainName             string           `json:"domain_name,omitempty"`
	OperatingSystem        string           `json:"operating_system_name,omitempty"`
	OperatingSystemVersion string           `json:"operating_system_version,omitempty"`
	Criticality            string           `json:"asset_criticality,omitempty"`
	Groups                 []string         `json:"asset_groups,omitempty"`
	Location               string           `json:"asset_location,omitempty"`
	Notes                  string           `json:"asset_notes,omitempty"`
	DataSensitivityScore   *int             `json:"asset_data_sensitivity_score,omitempty"`
	Public                 *bool            `json:"asset_public,omitempty"`
}

// UpdateAssetInput contains the data to update an existing asset.
type UpdateAssetInput struct {
	Name                   *string          `json:"asset_name,omitempty"`
	Type                   domain.AssetType `json:"asset_type,omitempty"`
	IPAddress              *string          `json:"ip_address,omitempty"`
	DomainName             *string          `json:"domain_name,omitempty"`
	OperatingSystem        *string          `json:"operating_system_name,omitempty"`
	OperatingSystemVersion *string          `json:"operating_system_version,omitempty"`
	Criticality            *string          `json:"asset_criticality,omitempty"`
	Groups                 []string         `json:"asset_groups,omitempty"`
	Location               *string          `json:"asset_location,omitempty"`
	Notes                  *string          `json:"asset_notes,omitempty"`
	DataSensitivityScore   *int             `json:"asset_data_sensitivity_score,omitempty"`
	Public                 *bool            `json:"asset_public,omitempty"`
	Active                 *bool            `json:"active,omitempty"`
}

// UpdateFindingInput contains the data to update a finding.
type UpdateFindingInput struct {
	FindingNumber string               `json:"finding_number"`
	Status        domain.FindingStatus `json:"finding_status,omitempty"`
	Severity      domain.Severity      `json:"finding_severity,omitempty"`
	Comment       string               `json:"comment,omitempty"`
	DueDate       string               `json:"due_date,omitempty"`
	TeamID        *int                 `json:"team_id,omitempty"`
	AssetID       *int                 `json:"asset_id,omitempty"`
}
