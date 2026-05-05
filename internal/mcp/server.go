package mcp

import (
	"github.com/marstid/nuc/internal/mcp/prompts"
	"github.com/marstid/nuc/internal/mcp/resources"
	"github.com/marstid/nuc/internal/mcp/tools"
	"github.com/marstid/nuc/pkg/nucleus"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server wired with all tools, resources, and prompts.
func New(cfg *Config) (*mcp.Server, error) {
	client := nucleus.NewClient(cfg.BaseURL, cfg.APIKey)

	svc := &tools.Services{
		Client: client,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nucleus-mcp",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "Nucleus Security MCP server. Always specify a project_id when calling tools. " +
			"Use list_projects first if you don't know the project ID. " +
			"Use search_findings for targeted queries; use get_finding_overview for summaries.",
	})

	tools.RegisterAll(svc, server)
	resources.RegisterAll(svc, server)
	prompts.RegisterAll(server)

	return server, nil
}
