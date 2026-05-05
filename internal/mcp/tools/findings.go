package tools

import (
	"context"

	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listFindingsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	Severity  string `json:"severity,omitempty" jsonschema:"Filter by severity: Critical,High,Medium,Low,Info"`
	Status    string `json:"status,omitempty" jsonschema:"Filter by status: Active,Fixed,Accepted Risk,False Positive,etc."`
	Start     *int   `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit     *int   `json:"limit,omitempty" jsonschema:"Max results to return"`
}

type getFindingInput struct {
	ProjectID     string `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	FindingNumber string `json:"finding_number" jsonschema:"required,The finding number"`
}

type searchFindingsInput struct {
	ProjectID          string   `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	AssetName          string   `json:"asset_name,omitempty" jsonschema:"Filter by asset name (supports wildcards)"`
	IPAddress          string   `json:"ip_address,omitempty" jsonschema:"Filter by IP address"`
	AssetGroups        []string `json:"asset_groups,omitempty" jsonschema:"Filter by asset groups"`
	ScanType           string   `json:"scan_type,omitempty" jsonschema:"Filter by scan type"`
	FindingCVE         string   `json:"finding_cve,omitempty" jsonschema:"Filter by CVE identifier"`
	FindingName        string   `json:"finding_name,omitempty" jsonschema:"Filter by finding name (supports wildcards)"`
	FindingSeverity    string   `json:"finding_severity,omitempty" jsonschema:"Filter by severity: Critical,High,Medium,Low,Info"`
	FindingExploitable string   `json:"finding_exploitable,omitempty" jsonschema:"Filter by exploitability"`
	Start              int      `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit              int      `json:"limit,omitempty" jsonschema:"Max results (default 100, max 1000)"`
}

type updateFindingInput struct {
	ProjectID     string `json:"project_id" jsonschema:"required,The Nucleus project ID"`
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
	ProjectID string                  `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	Updates   []bulkUpdateFindingItem `json:"updates" jsonschema:"required,List of finding updates"`
}

type getMitigatedFindingsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	Start     *int   `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit     *int   `json:"limit,omitempty" jsonschema:"Max results"`
	StartDate string `json:"start_date,omitempty" jsonschema:"Filter mitigated after date (YYYY-MM-DD)"`
}

type getFindingTrendInput struct {
	ProjectID   string   `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	StartDate   string   `json:"start_date,omitempty" jsonschema:"Trend start date (YYYY-MM-DD)"`
	EndDate     string   `json:"end_date,omitempty" jsonschema:"Trend end date (YYYY-MM-DD)"`
	AssetGroups []string `json:"asset_groups,omitempty" jsonschema:"Filter by asset groups"`
}

func registerFindings(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_findings",
		Description: "List findings in a project with optional severity and status filters",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listFindingsInput) (*mcp.CallToolResult, any, error) {
		opts := &domain.FindingListOptions{
			Start:    input.Start,
			Limit:    input.Limit,
			Severity: domain.Severity(input.Severity),
			Status:   domain.FindingStatus(input.Status),
		}
		findings, err := svc.Client.ListFindings(ctx, input.ProjectID, opts)
		if err != nil {
			return errorResult("listing findings", err), nil, nil
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding",
		Description: "Get detailed information about a specific finding by its finding number",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getFindingInput) (*mcp.CallToolResult, any, error) {
		finding, err := svc.Client.GetFinding(ctx, input.ProjectID, input.FindingNumber)
		if err != nil {
			return errorResult("getting finding", err), nil, nil
		}
		return jsonResult(finding), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_findings",
		Description: "Search findings using advanced filters including asset name, CVE, severity, and exploitability",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchFindingsInput) (*mcp.CallToolResult, any, error) {
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
		findings, err := svc.Client.SearchFindings(ctx, input.ProjectID, search, input.Start, input.Limit)
		if err != nil {
			return errorResult("searching findings", err), nil, nil
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_finding",
		Description: "Update a finding's status, severity, or add a comment",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input updateFindingInput) (*mcp.CallToolResult, any, error) {
		update := &service.UpdateFindingInput{
			FindingNumber: input.FindingNumber,
			Status:        domain.FindingStatus(input.Status),
			Severity:      domain.Severity(input.Severity),
			Comment:       input.Comment,
			DueDate:       input.DueDate,
		}
		if err := svc.Client.UpdateFinding(ctx, input.ProjectID, update); err != nil {
			return errorResult("updating finding", err), nil, nil
		}
		return jsonResult(map[string]string{"status": "updated", "finding_number": input.FindingNumber}), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bulk_update_findings",
		Description: "Update multiple findings at once",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input bulkUpdateFindingsInput) (*mcp.CallToolResult, any, error) {
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
		if err := svc.Client.BulkUpdateFindings(ctx, input.ProjectID, updates); err != nil {
			return errorResult("bulk updating findings", err), nil, nil
		}
		return jsonResult(map[string]any{"status": "updated", "count": len(updates)}), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_mitigated_findings",
		Description: "Get findings that have been mitigated, optionally filtered by date",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getMitigatedFindingsInput) (*mcp.CallToolResult, any, error) {
		opts := &domain.MitigatedOptions{
			Start:     input.Start,
			Limit:     input.Limit,
			StartDate: input.StartDate,
		}
		findings, err := svc.Client.GetMitigatedFindings(ctx, input.ProjectID, opts)
		if err != nil {
			return errorResult("getting mitigated findings", err), nil, nil
		}
		return jsonResult(findings), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_trend",
		Description: "Get vulnerability trend data over time for a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getFindingTrendInput) (*mcp.CallToolResult, any, error) {
		opts := &domain.TrendOptions{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			AssetGroups: input.AssetGroups,
		}
		trend, err := svc.Client.GetFindingTrend(ctx, input.ProjectID, opts)
		if err != nil {
			return errorResult("getting finding trend", err), nil, nil
		}
		return jsonResult(trend), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_overview",
		Description: "Get a summary overview of findings including severity distribution and vulnerability score",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		overview, err := svc.Client.GetFindingOverview(ctx, input.ProjectID)
		if err != nil {
			return errorResult("getting finding overview", err), nil, nil
		}
		return jsonResult(overview), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_frameworks",
		Description: "Get compliance frameworks associated with findings in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		frameworks, err := svc.Client.GetFrameworks(ctx, input.ProjectID)
		if err != nil {
			return errorResult("getting finding frameworks", err), nil, nil
		}
		return jsonResult(frameworks), nil, nil
	})
}
