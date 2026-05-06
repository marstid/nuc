package tools

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/marstid/nuc/pkg/nucleus"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Services holds the Nucleus API client used by all tool handlers.
type Services struct {
	Client         *nucleus.Client
	DefaultProject string
}

// RegisterAll registers all MCP tools on the given server.
func RegisterAll(svc *Services, server *mcp.Server) {
	registerProjects(svc, server)
	registerFindings(svc, server)
	registerAssets(svc, server)
	registerScans(svc, server)
	registerMetrics(svc, server)
}

func jsonResult(v any) *mcp.CallToolResult {
	data, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}
}

func errorResult(op string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Error %s: %v", op, err)},
		},
	}
}

func (s *Services) resolveProjectID(projectID string) (string, error) {
	if projectID != "" {
		return projectID, nil
	}
	if s.DefaultProject != "" {
		return s.DefaultProject, nil
	}
	return "", errors.New("project_id is required: provide project_id, set NUC_PROJECT/default_project, or use list_projects to choose one")
}
