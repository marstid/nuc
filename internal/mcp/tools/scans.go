package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listScansInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,The Nucleus project ID"`
	Start     int    `json:"start,omitempty" jsonschema:"Pagination offset"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Max results (default 1, max 100)"`
}

func registerScans(svc *Services, server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scans",
		Description: "List vulnerability scans in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input listScansInput) (*mcp.CallToolResult, any, error) {
		scans, err := svc.Client.ListScans(ctx, input.ProjectID, input.Start, input.Limit)
		if err != nil {
			return errorResult("listing scans", err), nil, nil
		}
		return jsonResult(scans), nil, nil
	})
}
