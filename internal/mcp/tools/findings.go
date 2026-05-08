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

type findingSummaryFilterInput struct {
	Property   string   `json:"property" jsonschema:"required,The property to filter on. Common properties: finding_severity (Critical/High/Medium/Low/Informational), finding_name, finding_status (Active/Fixed/Accepted Risk/False Positive/In Progress), finding_exploitable (0=no/1=yes/2=manual), scan_type (Container/DAST/SAST/Infrastructure), asset_groups, team_names, finding_cve"`
	Value      string   `json:"value,omitempty" jsonschema:"Single string value to filter by. Use this for single-value filters. Mutually exclusive with values."`
	Values     []string `json:"values,omitempty" jsonschema:"Array of string values for multi-value filters (supported for finding_severity, team_names, asset_groups). Mutually exclusive with value."`
	ExactMatch bool     `json:"exact_match,omitempty" jsonschema:"Set to true for exact match filtering. Default is false which uses substring/contains matching."`
}

type findingSummarySortInput struct {
	Property  string `json:"property" jsonschema:"required,The property to sort by (e.g. finding_severities)"`
	Direction string `json:"direction" jsonschema:"required,Sort direction: ASC or DESC"`
}

type getFindingsSummaryInput struct {
	ProjectID string                     `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
	Filter    []findingSummaryFilterInput `json:"filter,omitempty" jsonschema:"Array of filter conditions. Each filter specifies a property and value (string) or values (array). Multiple filters are ANDed together."`
	Sort      []findingSummarySortInput   `json:"sort,omitempty" jsonschema:"Array of sort rules with property and direction (ASC/DESC)."`
	Start     int                        `json:"start,omitempty" jsonschema:"Pagination offset (default: 0)"`
	Limit     *int                       `json:"limit,omitempty" jsonschema:"Max results per page (API max: 100). Omit to auto-paginate and return all results."`
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

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_findings_summary",
		Description: `Get findings grouped by finding_number (deduplicated across assets) with flexible filtering and sorting.

Unlike search_findings which returns one result per finding-per-asset, this returns unique findings with aggregate counts showing how many assets are affected (asset_count), total instances (finding_count), how many are fixed (asset_fixed_count), and how many are mitigated (asset_mitigated_count).

FILTERING:
Each filter is an object with: property, value (string) OR values (string array), and optional exact_match (bool).
- String filter example: {"property": "finding_name", "value": "SQL Injection"}
- Array filter example: {"property": "finding_severity", "values": ["Critical", "High"]}
- Exact match example: {"property": "scan_type", "value": "Container", "exact_match": true}
- Multiple filters are ANDed together.

Common filterable properties:
- finding_severity: Critical, High, Medium, Low, Informational (supports array via values)
- finding_name: vulnerability name (substring match by default, use exact_match for exact)
- finding_status: Active, Fixed, Accepted Risk, False Positive, In Progress, etc.
- finding_exploitable: "0" (not exploitable), "1" (exploitable), "2" (manually set)
- scan_type: Container, DAST, SAST, Infrastructure, etc. (use exact_match recommended)
- asset_groups: filter by asset group names (supports array via values)
- team_names: filter by assigned team names (supports array via values)
- finding_cve: filter by CVE identifier

SORTING:
- property: "finding_severities" (commonly used)
- direction: "ASC" (ascending) or "DESC" (descending)

EXAMPLES:
1. All critical and high findings:
   filter: [{"property": "finding_severity", "values": ["Critical", "High"]}]

2. Exploitable findings for a specific team:
   filter: [{"property": "finding_exploitable", "value": "1"}, {"property": "team_names", "values": ["team-platform"]}]

3. Container findings sorted by severity descending:
   filter: [{"property": "scan_type", "value": "Container", "exact_match": true}]
   sort: [{"property": "finding_severities", "direction": "DESC"}]

4. Active findings in specific asset groups:
   filter: [{"property": "finding_status", "value": "Active"}, {"property": "asset_groups", "values": ["/service/my-service"]}]`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input getFindingsSummaryInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}

		// Convert MCP input filters to domain filters.
		var filters []domain.FindingSummaryFilter
		for _, f := range input.Filter {
			filters = append(filters, domain.FindingSummaryFilter{
				Property:   f.Property,
				Value:      f.Value,
				Values:     f.Values,
				ExactMatch: f.ExactMatch,
			})
		}

		// Convert sort rules.
		var sorts []domain.FindingSummarySort
		for _, s := range input.Sort {
			sorts = append(sorts, domain.FindingSummarySort{
				Property:  s.Property,
				Direction: s.Direction,
			})
		}

		summaryReq := &domain.FindingSummaryRequest{
			Filter: filters,
			Sort:   sorts,
		}

		const findingsSummaryPageSize = 100

		var findings []domain.FindingSummary
		if input.Limit != nil {
			findings, err = svc.Client.GetFindingsSummary(ctx, projectID, summaryReq, input.Start, *input.Limit)
			if err != nil {
				return errorResult("getting findings summary", err), nil, nil
			}
		} else {
			// Auto-paginate with page size 100 (API max).
			offset := input.Start
			for {
				page, err := svc.Client.GetFindingsSummary(ctx, projectID, summaryReq, offset, findingsSummaryPageSize)
				if err != nil {
					return errorResult("getting findings summary", err), nil, nil
				}
				findings = append(findings, page...)
				if len(page) < findingsSummaryPageSize {
					break
				}
				offset += findingsSummaryPageSize
			}
		}

		return jsonResult(findings), nil, nil
	})
}
