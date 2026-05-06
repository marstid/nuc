package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerMetrics(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_finding_metrics",
		Description: "Get aggregated vulnerability discovery and remediation metrics over 30, 90, and 180-day windows",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, any, error) {
		projectID, err := svc.resolveProjectID(input.ProjectID)
		if err != nil {
			return errorResult("resolving project", err), nil, nil
		}
		metrics, err := svc.Client.GetFindingMetrics(ctx, projectID)
		if err != nil {
			return errorResult("getting finding metrics", err), nil, nil
		}
		return jsonResult(metrics), nil, nil
	})
}
