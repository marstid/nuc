package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerInitialization(t *testing.T) {
	cfg := &Config{
		APIKey:  "test-key",
		BaseURL: "https://example.com/api",
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.1.0"}, nil)

	clientTrans, serverTrans := mcp.NewInMemoryTransports()

	go func() {
		if err := srv.Run(context.Background(), serverTrans); err != nil {
			t.Errorf("server run: %v", err)
		}
	}()

	session, err := client.Connect(context.Background(), clientTrans, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	expectedTools := []string{
		"list_projects", "get_project", "get_project_risk_score",
		"list_findings", "get_finding", "search_findings", "update_finding",
		"bulk_update_findings", "get_mitigated_findings", "get_finding_trend",
		"get_finding_overview", "get_finding_frameworks",
		"list_assets", "get_asset", "update_asset", "list_asset_groups", "get_asset_group_metrics", "list_teams",
		"list_scans",
		"get_finding_metrics",
	}

	if len(tools.Tools) != len(expectedTools) {
		t.Fatalf("expected %d tools, got %d", len(expectedTools), len(tools.Tools))
	}

	gotNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		gotNames[tool.Name] = true
	}

	for _, name := range expectedTools {
		if !gotNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}

	t.Logf("All %d tools registered correctly", len(tools.Tools))

	resources, err := session.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}

	if len(resources.Resources) < 1 {
		t.Fatalf("expected at least 1 resource, got %d", len(resources.Resources))
	}
	t.Logf("Resources: %d registered", len(resources.Resources))

	templates, err := session.ListResourceTemplates(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResourceTemplates: %v", err)
	}
	t.Logf("Resource templates: %d registered", len(templates.ResourceTemplates))

	prompts, err := session.ListPrompts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	expectedPrompts := []string{"analyze_finding", "security_report", "risk_assessment", "triage_findings", "nucleus_report"}
	if len(prompts.Prompts) != len(expectedPrompts) {
		t.Fatalf("expected %d prompts, got %d", len(expectedPrompts), len(prompts.Prompts))
	}

	gotPromptNames := make(map[string]bool)
	for _, p := range prompts.Prompts {
		gotPromptNames[p.Name] = true
	}
	for _, name := range expectedPrompts {
		if !gotPromptNames[name] {
			t.Errorf("missing prompt: %s", name)
		}
	}
	t.Logf("All %d prompts registered correctly", len(prompts.Prompts))
}
