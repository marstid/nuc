package tools

import (
	"context"

	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listFindingsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	Severity  string `json:"severity,omitempty" jsonschema:"Filter by severity: Critical,High,Medium,Low,Info"`
	Status    string `json:"status,omitempty" jsonschema:"Filter by status: Active,Fixed,Accepted Risk,False Positive,etc."`
	Start     *int   `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit     *int   `json:"limit,omitempty" jsonschema:"Max results to return"`
}

type getFindingInput struct {
	ProjectID     string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	FindingNumber string `json:"finding_number" jsonschema:"required,The finding number"`
}

type searchFindingsInput struct {
	ProjectID          string   `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	AssetName          string   `json:"asset_name,omitempty" jsonschema:"Filter by asset name (supports wildcards)"`
	IPAddress          string   `json:"ip_address,omitempty" jsonschema:"Filter by IP address"`
	AssetGroups        []string `json:"asset_groups,omitempty" jsonschema:"Filter by asset groups"`
	ScanType           string   `json:"scan_type,omitempty" jsonschema:"Filter by scan type"`
	FindingCVE         string   `json:"finding_cve,omitempty" jsonschema:"Filter by CVE identifier"`
	FindingName        string   `json:"finding_name,omitempty" jsonschema:"Filter by finding name (supports wildcards)"`
	FindingSeverity    string   `json:"finding_severity,omitempty" jsonschema:"Filter by severity: Critical,High,Medium,Low,Info"`
	FindingExploitable string   `json:"finding_exploitable,omitempty" jsonschema:"Filter by exploitability"`
	Start              int      `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit              *int     `json:"limit,omitempty" jsonschema:"Max results to return. Omit to fetch all results automatically."`
}

type updateFindingInput struct {
	ProjectID     string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	FindingNumber string `json:"finding_number" jsonschema:"required,The finding number to update"`
	Status        string `json:"finding_status,omitempty" jsonschema:"New status: Active,Fixed,Accepted Risk,False Positive,etc."`
	Severity      string `json:"finding_severity,omitempty" jsonschema:"New severity: Critical,High,Medium,Low,Info"`
	Comment       string `json:"comment,omitempty" jsonschema:"Comment explaining the change"`
	DueDate       string `json:"due_date,omitempty" jsonschema:"New due date (YYYY-MM-DD)"`
}

type bulkUpdateFindingItem struct {
	FindingNumber string `json:"finding_number" jsonschema:"required"`
	Status        string `json:"finding_status,omitempty"`
	Severity      string `json:"finding_severity,omitempty"`
	Comment       string `json:"comment,omitempty"`
	DueDate       string `json:"due_date,omitempty"`
}

type bulkUpdateFindingsInput struct {
	ProjectID string                  `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	Updates   []bulkUpdateFindingItem `json:"updates" jsonschema:"required,List of finding updates"`
}

type getMitigatedFindingsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	Start     *int   `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit     *int   `json:"limit,omitempty" jsonschema:"Max results"`
	StartDate string `json:"start_date,omitempty" jsonschema:"Filter mitigated after date (YYYY-MM-DD)"`
}

type getFindingTrendInput struct {
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	StartDate   string   `json:"start_date,omitempty" jsonschema:"Trend start date (YYYY-MM-DD)"`
	EndDate     string   `json:"end_date,omitempty" jsonschema:"Trend end date (YYYY-MM-DD)"`
	AssetGroups []string `json:"asset_groups,omitempty" jsonschema:"Filter by asset groups"`
}

func paginateAll[T any](ctx context.Context, fetch func(offset, limit int) ([]T, error)) ([]T, error) {
	const pageSize = 1000
	var all []T
	offset := 0
	for {
		page, err := fetch(offset, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) < pageSize {
			break
		}
		offset += pageSize
	}
	return all, nil
}

func registerFindings(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_findings",
		Description: "List findings in a project with optional severity and status filters",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listFindingsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		var findings []domain.Finding
		if input.Limit != nil {
			opts := &domain.FindingListOptions{
				Start:    input.Start,
				Limit:    input.Limit,
				Severity: domain.Severity(input.Severity),
				Status:   domain.FindingStatus(input.Status),
			}
			findings, err = svc.Client.ListFindings(ctx, projectID, opts)
			if err != nil {
				return errorResult("listing findings", err), nil, nil
			}
		} else {
			findings, err = paginateAll(ctx, func(offset, limit int) ([]domain.Finding, error) {
				o := offset
				l := limit
				opts := &domain.FindingListOptions{
					Start:    &o,
					Limit:    &l,
					Severity: domain.Severity(input.Severity),
					Status:   domain.FindingStatus(input.Status),
				}
				return svc.Client.ListFindings(ctx, projectID, opts)
			})
			if err != nil {
				return errorResult("listing findings", err), nil, nil
			}
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding",
		Description: "Get detailed information about a specific finding by its finding number",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getFindingInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		finding, err := svc.Client.GetFinding(ctx, projectID, input.FindingNumber)
		if err != nil {
			return errorResult("getting finding", err), nil, nil
		}
		return jsonResult(finding), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_findings",
		Description: "Search findings using advanced filters including asset name, CVE, severity, and exploitability",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchFindingsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		search := &domain.FindingSearch{
			AssetName:          input.AssetName,
			IPAddress:          input.IPAddress,
			AssetGroups:        input.AssetGroups,
			ScanType:           input.ScanType,
			FindingCVE:         input.FindingCVE,
			FindingName:        input.FindingName,
			FindingSeverity:    input.FindingSeverity,
			FindingExploitable: input.FindingExploitable,
		}
		var findings []domain.Finding
		if input.Limit != nil {
			findings, err = svc.Client.SearchFindings(ctx, projectID, search, input.Start, *input.Limit)
			if err != nil {
				return errorResult("searching findings", err), nil, nil
			}
		} else {
			findings, err = paginateAll(ctx, func(offset, limit int) ([]domain.Finding, error) {
				return svc.Client.SearchFindings(ctx, projectID, search, offset, limit)
			})
			if err != nil {
				return errorResult("searching findings", err), nil, nil
			}
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_finding",
		Description: "Update a finding's status, severity, or add a comment",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input updateFindingInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		update := &service.UpdateFindingInput{
			FindingNumber: input.FindingNumber,
			Status:        domain.FindingStatus(input.Status),
			Severity:      domain.Severity(input.Severity),
			Comment:       input.Comment,
			DueDate:       input.DueDate,
		}
		if err := svc.Client.UpdateFinding(ctx, projectID, update); err != nil {
			return errorResult("updating finding", err), nil, nil
		}
		return jsonResult(map[string]string{"status": "updated", "finding_number": input.FindingNumber}), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bulk_update_findings",
		Description: "Update multiple findings at once",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input bulkUpdateFindingsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		updates := make([]service.UpdateFindingInput, len(input.Updates))
		for i, item := range input.Updates {
			updates[i] = service.UpdateFindingInput{
				FindingNumber: item.FindingNumber,
				Status:        domain.FindingStatus(item.Status),
				Severity:      domain.Severity(item.Severity),
				Comment:       item.Comment,
				DueDate:       item.DueDate,
			}
		}
		if err := svc.Client.BulkUpdateFindings(ctx, projectID, updates); err != nil {
			return errorResult("bulk updating findings", err), nil, nil
		}
		return jsonResult(map[string]any{"status": "updated", "count": len(updates)}), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_mitigated_findings",
		Description: "Get findings that have been mitigated, optionally filtered by date",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getMitigatedFindingsInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		var findings []domain.MitigatedFinding
		if input.Limit != nil {
			opts := &domain.MitigatedOptions{
				Start:     input.Start,
				Limit:     input.Limit,
				StartDate: input.StartDate,
			}
			findings, err = svc.Client.GetMitigatedFindings(ctx, projectID, opts)
			if err != nil {
				return errorResult("getting mitigated findings", err), nil, nil
			}
		} else {
			findings, err = paginateAll(ctx, func(offset, limit int) ([]domain.MitigatedFinding, error) {
				o := offset
				l := limit
				opts := &domain.MitigatedOptions{
					Start:     &o,
					Limit:     &l,
					StartDate: input.StartDate,
				}
				return svc.Client.GetMitigatedFindings(ctx, projectID, opts)
			})
			if err != nil {
				return errorResult("getting mitigated findings", err), nil, nil
			}
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_trend",
		Description: "Get vulnerability trend data over time for a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getFindingTrendInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		opts := &domain.TrendOptions{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			AssetGroups: input.AssetGroups,
		}
		trend, err := svc.Client.GetFindingTrend(ctx, projectID, opts)
		if err != nil {
			return errorResult("getting finding trend", err), nil, nil
		}
		return jsonResult(trend), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_overview",
		Description: "Get a summary overview of findings including severity distribution and vulnerability score",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		overview, err := svc.Client.GetFindingOverview(ctx, projectID)
		if err != nil {
			return errorResult("getting finding overview", err), nil, nil
		}
		return jsonResult(overview), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_frameworks",
		Description: "Get compliance frameworks associated with findings in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		frameworks, err := svc.Client.GetFrameworks(ctx, projectID)
		if err != nil {
			return errorResult("getting finding frameworks", err), nil, nil
		}
		return jsonResult(frameworks), nil, nil
	})
}
