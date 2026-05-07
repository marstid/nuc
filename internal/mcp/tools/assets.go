package tools

import (
	"context"
	"strings"

	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listAssetsInput struct {
	ProjectID          string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	IPAddress          string `json:"ip_address,omitempty" jsonschema:"Filter by IP address"`
	AssetName          string `json:"asset_name,omitempty" jsonschema:"Filter by asset name"`
	AssetGroups        string `json:"asset_groups,omitempty" jsonschema:"Filter by asset groups"`
	AssetType          string `json:"asset_type,omitempty" jsonschema:"Filter by type: Host,Web Application,Database,Container,Cloud Asset,Repository,Mobile Application,Network,Other"`
	InactiveAssets     *bool  `json:"inactive_assets,omitempty" jsonschema:"Include inactive assets"`
	UnscannedAssets    *bool  `json:"unscanned_assets,omitempty" jsonschema:"Include unscanned assets"`
	AssetsWithFindings *bool  `json:"assets_with_findings,omitempty" jsonschema:"Only assets with findings"`
	Start              *int   `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit              *int   `json:"limit,omitempty" jsonschema:"Max results to return"`
}

type getAssetInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	AssetID   string `json:"asset_id" jsonschema:"required,The asset ID"`
}

type updateAssetInput struct {
	ProjectID            string   `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	AssetID              string   `json:"asset_id" jsonschema:"required,The asset ID"`
	Name                 *string  `json:"asset_name,omitempty" jsonschema:"Asset name"`
	Criticality          *string  `json:"asset_criticality,omitempty" jsonschema:"Criticality level"`
	Groups               []string `json:"asset_groups,omitempty" jsonschema:"Asset group memberships"`
	Notes                *string  `json:"asset_notes,omitempty" jsonschema:"Notes"`
	Location             *string  `json:"asset_location,omitempty" jsonschema:"Asset location"`
	DataSensitivityScore *int     `json:"asset_data_sensitivity_score,omitempty" jsonschema:"Data sensitivity score"`
	Public               *bool    `json:"asset_public,omitempty" jsonschema:"Whether the asset is public-facing"`
	Active               *bool    `json:"active,omitempty" jsonschema:"Whether the asset is active"`
}

type getAssetGroupMetricsInput struct {
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	AssetGroups []string `json:"asset_groups" jsonschema:"required,Asset group names (up to 50)"`
	Metrics     []string `json:"metrics,omitempty" jsonschema:"Specific metric names to include"`
}

type listTeamsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	InGroup   string `json:"in_group,omitempty" jsonschema:"Optional asset group prefix to filter teams (e.g. /service returns only teams with service sub-groups)"`
}

type listServicesInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	Team      string `json:"team,omitempty" jsonschema:"Optional team name (e.g. team-euc) to filter services associated with that team"`
}

func registerAssets(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Description: "List assets in a project with optional filters",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listAssetsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		opts := &domain.AssetListOptions{
			Start:              input.Start,
			Limit:              input.Limit,
			IPAddress:          input.IPAddress,
			AssetName:          input.AssetName,
			AssetGroups:        input.AssetGroups,
			AssetType:          domain.AssetType(input.AssetType),
			InactiveAssets:     input.InactiveAssets,
			UnscannedAssets:    input.UnscannedAssets,
			AssetsWithFindings: input.AssetsWithFindings,
		}
		assets, err := svc.Client.ListAssets(ctx, projectID, opts)
		if err != nil {
			return errorResult("listing assets", err), nil, nil
		}
		return jsonResult(assets), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset",
		Description: "Get details of a specific asset by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getAssetInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		asset, err := svc.Client.GetAsset(ctx, projectID, input.AssetID)
		if err != nil {
			return errorResult("getting asset", err), nil, nil
		}
		return jsonResult(asset), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_asset",
		Description: "Update an asset's properties such as name, criticality, groups, and notes",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input updateAssetInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		update := &service.UpdateAssetInput{
			Name:                 input.Name,
			Criticality:          input.Criticality,
			Groups:               input.Groups,
			Notes:                input.Notes,
			Location:             input.Location,
			DataSensitivityScore: input.DataSensitivityScore,
			Public:               input.Public,
			Active:               input.Active,
		}
		asset, err := svc.Client.UpdateAsset(ctx, projectID, input.AssetID, update)
		if err != nil {
			return errorResult("updating asset", err), nil, nil
		}
		return jsonResult(asset), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_asset_groups",
		Description: "List all asset groups in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		groups, err := svc.Client.ListAssetGroups(ctx, projectID)
		if err != nil {
			return errorResult("listing asset groups", err), nil, nil
		}
		return jsonResult(groups), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset_group_metrics",
		Description: "Get security metrics for one or more asset groups in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getAssetGroupMetricsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		opts := &domain.AssetGroupMetricsOptions{
			AssetGroups: input.AssetGroups,
			Metrics:     input.Metrics,
		}
		metrics, err := svc.Client.GetAssetGroupMetrics(ctx, projectID, opts)
		if err != nil {
			return errorResult("getting asset group metrics", err), nil, nil
		}
		return jsonResult(metrics), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_teams",
		Description: "List all teams in a project. Teams are asset groups with names starting with /teams/. Returns top-level teams only (e.g. /teams/team-euc), excluding sub-groups like /teams/team-euc/container. Optionally pass in_group to return only teams that have sub-groups matching that prefix (e.g. in_group=/service to return only teams with service assets).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listTeamsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}

		if input.InGroup == "" {
			teams, err := svc.Client.ListTeams(ctx, projectID)
			if err != nil {
				return errorResult("listing teams", err), nil, nil
			}
			return jsonResult(teams), nil, nil
		}

		inGroup := input.InGroup
		if !strings.HasPrefix(inGroup, "/") {
			inGroup = "/" + inGroup
		}

		groups, err := svc.Client.ListAssetGroups(ctx, projectID)
		if err != nil {
			return errorResult("listing asset groups for team filter", err), nil, nil
		}

		matchingTeams := teamsWithGroupPrefix(groups, inGroup)

		teams, err := svc.Client.ListTeams(ctx, projectID)
		if err != nil {
			return errorResult("listing teams", err), nil, nil
		}

		filtered := make([]domain.Team, 0, len(teams))
		for _, t := range teams {
			if matchingTeams[t.TeamName] {
				filtered = append(filtered, t)
			}
		}
		return jsonResult(filtered), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_services",
		Description: "List all services in a project. Services are asset groups with names starting with /service/. Returns top-level service names only (e.g. my-service). Optionally filter by team to return only services associated with that team.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listServicesInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}

		if input.Team != "" {
			teamGroup := "/teams/" + input.Team
			assets, err := svc.Client.ListAssets(ctx, projectID, &domain.AssetListOptions{AssetGroups: teamGroup})
			if err != nil {
				return errorResult("listing services for team", err), nil, nil
			}
			seen := make(map[string]bool)
			var services []string
			for _, a := range assets {
				for _, g := range a.Groups {
					if !strings.HasPrefix(g, "/service/") {
						continue
					}
					rest := strings.TrimPrefix(g, "/service/")
					if strings.Contains(rest, "/") {
						continue
					}
					if !seen[rest] {
						seen[rest] = true
						services = append(services, rest)
					}
				}
			}
			return jsonResult(services), nil, nil
		}

		groups, err := svc.Client.ListAssetGroups(ctx, projectID)
		if err != nil {
			return errorResult("listing services", err), nil, nil
		}
		var services []string
		for _, g := range groups {
			if !strings.HasPrefix(g.Name, "/service/") {
				continue
			}
			rest := strings.TrimPrefix(g.Name, "/service/")
			if strings.Contains(rest, "/") {
				continue
			}
			services = append(services, rest)
		}
		return jsonResult(services), nil, nil
	})
}

func teamsWithGroupPrefix(groups []domain.AssetGroup, inGroup string) map[string]bool {
	matching := make(map[string]bool)
	for _, g := range groups {
		if !strings.HasPrefix(g.Name, "/teams/") {
			continue
		}
		afterTeams := g.Name[len("/teams/"):]
		slashIdx := strings.Index(afterTeams, "/")
		if slashIdx < 0 {
			continue
		}
		teamName := afterTeams[:slashIdx]
		remainder := afterTeams[slashIdx:]
		if strings.HasPrefix(remainder, inGroup) {
			matching[teamName] = true
		}
	}
	return matching
}
