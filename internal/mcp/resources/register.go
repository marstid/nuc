package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/marstid/nuc/internal/mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all MCP resources and resource templates on the given server.
func RegisterAll(svc *tools.Services, server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:     "All Projects",
		URI:      "nucleus://projects",
		MIMEType: "application/json",
	}, projectsResource(svc))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Finding Overview",
		URITemplate: "nucleus://projects/{project_id}/overview",
		MIMEType:    "application/json",
	}, overviewResource(svc))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Project Risk Score",
		URITemplate: "nucleus://projects/{project_id}/risk-score",
		MIMEType:    "application/json",
	}, riskScoreResource(svc))

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "Finding Metrics",
		URITemplate: "nucleus://projects/{project_id}/metrics",
		MIMEType:    "application/json",
	}, metricsResource(svc))
}

func projectsResource(svc *tools.Services) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		projects, err := svc.Client.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing projects: %w", err)
		}
		return resourceJSON(req.Params.URI, projects)
	}
}

func overviewResource(svc *tools.Services) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		projectID, err := extractProjectID(req.Params.URI)
		if err != nil {
			return nil, err
		}
		overview, err := svc.Client.GetFindingOverview(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("getting finding overview: %w", err)
		}
		return resourceJSON(req.Params.URI, overview)
	}
}

func riskScoreResource(svc *tools.Services) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		projectID, err := extractProjectID(req.Params.URI)
		if err != nil {
			return nil, err
		}
		score, err := svc.Client.GetRiskScore(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("getting risk score: %w", err)
		}
		return resourceJSON(req.Params.URI, score)
	}
}

func metricsResource(svc *tools.Services) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		projectID, err := extractProjectID(req.Params.URI)
		if err != nil {
			return nil, err
		}
		metrics, err := svc.Client.GetFindingMetrics(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("getting finding metrics: %w", err)
		}
		return resourceJSON(req.Params.URI, metrics)
	}
}

func extractProjectID(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parsing URI %q: %w", uri, err)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "projects" {
		return "", fmt.Errorf("invalid resource URI format: %q", uri)
	}
	return parts[1], nil
}

func resourceJSON(uri string, v any) (*mcp.ReadResourceResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling resource: %w", err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: uri, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}
