package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type projectIDInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"The Nucleus project ID. Optional when a default project is configured or auto-detected."`
}

func registerProjects(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all Nucleus Security projects accessible to the authenticated user",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		projects, err := svc.Client.List(ctx)
		if err != nil {
			return errorResult("listing projects", err), nil, nil
		}
		return jsonResult(projects), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project",
		Description: "Get details of a specific Nucleus Security project by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		project, err := svc.Client.Get(ctx, projectID)
		if err != nil {
			return errorResult("getting project", err), nil, nil
		}
		return jsonResult(project), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project_risk_score",
		Description: "Get the risk score for a specific Nucleus Security project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		score, err := svc.Client.GetRiskScore(ctx, projectID)
		if err != nil {
			return errorResult("getting project risk score", err), nil, nil
		}
		return jsonResult(score), nil, nil
	})
}
