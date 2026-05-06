package mcp

import (
	"context"
	"time"

	"github.com/marstid/nuc/internal/mcp/prompts"
	"github.com/marstid/nuc/internal/mcp/resources"
	"github.com/marstid/nuc/internal/mcp/tools"
	"github.com/marstid/nuc/pkg/nucleus"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server wired with all tools, resources, and prompts.
func New(cfg *Config) (*mcp.Server, error) {
	client := nucleus.NewClient(cfg.BaseURL, cfg.APIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defaultProject, _ := resolveDefaultProject(ctx, client, cfg.DefaultProject)

	svc := &tools.Services{
		Client:         client,
		DefaultProject: defaultProject,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nucleus-mcp",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "Nucleus Security MCP server. project_id is optional when a default project is configured or auto-detected. " +
			"Use list_projects first if you need a different project ID. " +
			"Use search_findings for targeted queries; use get_finding_overview for summaries.",
	})

	tools.RegisterAll(svc, server)
	resources.RegisterAll(svc, server)
	prompts.RegisterAll(server)

	return server, nil
}

func resolveDefaultProject(ctx context.Context, client *nucleus.Client, configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}

	projects, err := client.List(ctx)
	if err != nil {
		return "", err
	}
	if len(projects) != 1 {
		return "", nil
	}
	if projects[0].ID == "" {
		return "", nil
	}
	return projects[0].ID, nil
}
