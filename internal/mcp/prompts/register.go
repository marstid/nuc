package prompts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all MCP prompts on the given server.
func RegisterAll(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "analyze_finding",
		Description: "Deep-dive analysis of a specific vulnerability finding with remediation guidance",
		Arguments: []*mcp.PromptArgument{
			{Name: "project_id", Description: "The Nucleus project ID. Optional when a default project is configured or auto-detected."},
			{Name: "finding_number", Description: "The finding number to analyze", Required: true},
		},
	}, analyzeFindingPrompt)

	server.AddPrompt(&mcp.Prompt{
		Name:        "security_report",
		Description: "Generate a security posture report with trend analysis for a project",
		Arguments: []*mcp.PromptArgument{
			{Name: "project_id", Description: "The Nucleus project ID. Optional when a default project is configured or auto-detected."},
			{Name: "timeframe", Description: "Timeframe in days: 30, 90, or 180 (default: 90)"},
		},
	}, securityReportPrompt)

	server.AddPrompt(&mcp.Prompt{
		Name:        "risk_assessment",
		Description: "Perform a comprehensive risk assessment for a project with action items",
		Arguments: []*mcp.PromptArgument{
			{Name: "project_id", Description: "The Nucleus project ID. Optional when a default project is configured or auto-detected."},
		},
	}, riskAssessmentPrompt)

	server.AddPrompt(&mcp.Prompt{
		Name:        "triage_findings",
		Description: "Guided finding triage workflow to systematically review and assign status to findings",
		Arguments: []*mcp.PromptArgument{
			{Name: "project_id", Description: "The Nucleus project ID. Optional when a default project is configured or auto-detected."},
			{Name: "severity", Description: "Filter by severity: Critical, High, Medium, Low"},
		},
	}, triageFindingsPrompt)

	server.AddPrompt(&mcp.Prompt{
		Name:        "nucleus_report",
		Description: "Generate a Nucleus Security status summary report, optionally focused on a specific team",
		Arguments: []*mcp.PromptArgument{
			{Name: "project_id", Description: "The Nucleus project ID. Optional when a default project is configured or auto-detected."},
		},
	}, nucleusReportPrompt)
}

func analyzeFindingPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectID := projectRef(req.Params.Arguments["project_id"])
	findingNumber := req.Params.Arguments["finding_number"]

	return &mcp.GetPromptResult{
		Description: "Analyze a specific vulnerability finding in detail",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Analyze finding %s in %s.

1. Call get_finding to retrieve the full finding details.
2. Call get_asset to retrieve the affected asset (use the asset_id from the finding).
3. Provide:
   - Vulnerability summary (name, CVE, severity, description)
   - Exploitability assessment (EPSS score, Mandiant data if available)
   - Affected asset context (type, criticality, public-facing status)
   - Remediation recommendation from the finding
   - Suggested next action (e.g., update status, assign due date)`, findingNumber, projectID)},
			},
		},
	}, nil
}

func securityReportPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectID := projectRef(req.Params.Arguments["project_id"])
	timeframe := req.Params.Arguments["timeframe"]
	if timeframe == "" {
		timeframe = "90"
	}

	return &mcp.GetPromptResult{
		Description: "Generate a security posture report",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Generate a security report for %s covering the last %s days.

1. Call get_finding_overview for vulnerability distribution.
2. Call get_finding_metrics for discovery/remediation rates.
3. Call get_finding_trend for the vulnerability trajectory over the past %s days.
4. Call get_project_risk_score for overall risk.
5. Provide a structured report:
   - Executive summary (risk score, critical/high counts)
   - Vulnerability distribution by severity
   - Remediation velocity (MTTR, discovered vs. remediated)
   - Trend analysis (improving/stable/declining)
   - Top 5 recommendations`, projectID, timeframe, timeframe)},
			},
		},
	}, nil
}

func riskAssessmentPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectID := projectRef(req.Params.Arguments["project_id"])

	return &mcp.GetPromptResult{
		Description: "Perform a comprehensive risk assessment",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Perform a comprehensive risk assessment for %s.

1. Call get_project_risk_score for the overall score.
2. Call get_finding_overview for vulnerability counts.
3. Call get_finding_metrics for remediation velocity.
4. Call search_findings with finding_severity=Critical and finding_exploitable=Yes for active critical exploitable findings.
5. Call list_asset_groups then get_asset_group_metrics for group-level risk breakdown.
6. Provide:
   - Overall risk rating with justification
   - Critical exploitable findings requiring immediate attention
   - Asset group risk comparison
   - Remediation gap analysis
   - Prioritized action items`, projectID)},
			},
		},
	}, nil
}

func triageFindingsPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectID := projectRef(req.Params.Arguments["project_id"])
	severity := req.Params.Arguments["severity"]
	if severity == "" {
		severity = "Critical"
	}

	return &mcp.GetPromptResult{
		Description: "Guided finding triage workflow",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Guide me through triaging findings in %s.

1. Call get_finding_overview to understand the scope.
2. Call search_findings with finding_severity=%s to get the relevant findings.
3. For each finding, help me decide:
   - Is this a true positive or false positive?
   - What severity does it deserve?
   - Should it be accepted as risk, fixed, or exception requested?
4. When I confirm a decision, use update_finding to apply the status change with a comment.
5. Continue until all findings are triaged or I say stop.`, projectID, severity)},
			},
		},
	}, nil
}

func nucleusReportPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectID := projectRef(req.Params.Arguments["project_id"])
	projectArg := projectToolArgument(req.Params.Arguments["project_id"])

	return &mcp.GetPromptResult{
		Description: "Generate a Nucleus Security status summary report",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Generate a Nucleus Security status summary report for %s.

Step 1: Ask the user what kind of report they want using the question tool:
- "General overview" — full project status
- "Team-specific report" — focused on a single team

Step 2: If the user chose team-specific:
1. Call list_teams%s to get available teams.
2. Present the team names as options using the question tool so the user can pick one.
3. Use the selected team name as the asset_groups filter in subsequent calls.

Step 3: Gather data and generate the report.

For a GENERAL OVERVIEW:
1. Call get_project_risk_score for overall risk.
2. Call get_finding_overview for vulnerability distribution.
3. Call get_finding_metrics for 30/90/180-day discovery and remediation rates.
4. Call get_finding_trend for the vulnerability trajectory (use start_date 90 days ago).

For a TEAM-SPECIFIC REPORT:
1. Call get_asset_group_metrics with asset_groups=[selected_team] for team-level risk and metrics.
2. Call search_findings with asset_groups=[selected_team] and finding_severity=Critical for the team's critical findings.
3. Call search_findings with asset_groups=[selected_team] and finding_exploitable=Yes for the team's exploitable findings.
4. Call get_finding_trend with asset_groups=[selected_team] for the team's vulnerability trajectory.

IMPORTANT: The Nucleus API silently returns empty results if asset_groups has more than ~12 entries. Only pass one team at a time.

Step 4: Present a structured report:
- Executive summary (risk score, critical/high counts)
- Vulnerability distribution by severity
- Remediation velocity (MTTR, discovered vs. remediated)
- Trend analysis (improving/stable/declining)
- Top findings or risk areas
- Prioritized recommendations`, projectID, projectArg)},
			},
		},
	}, nil
}

func projectRef(projectID string) string {
	if projectID != "" {
		return projectID
	}
	return "the default Nucleus project"
}

func projectToolArgument(projectID string) string {
	if projectID != "" {
		return fmt.Sprintf(" with project_id=%s", projectID)
	}
	return " without project_id so the MCP default project is used"
}
