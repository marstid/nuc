package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marstid/nuc/pkg/config"
	"github.com/marstid/nuc/pkg/nucleus"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerInitialization(t *testing.T) {
	cfg := &Config{
		APIKey:         "test-key",
		BaseURL:        "https://example.com/api",
		DefaultProject: "42",
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
		"list_assets", "get_asset", "update_asset", "list_asset_groups", "get_asset_group_metrics", "list_teams", "list_services",
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

func TestResolveDefaultProject(t *testing.T) {
	t.Run("configured default wins", func(t *testing.T) {
		client := nucleus.NewClient("http://127.0.0.1:1", "test-key")

		projectID, err := resolveDefaultProject(context.Background(), client, "configured")

		if err != nil {
			t.Fatalf("resolveDefaultProject() error: %v", err)
		}
		if projectID != "configured" {
			t.Fatalf("expected configured, got %q", projectID)
		}
	})

	t.Run("single listed project becomes default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/projects" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"project_id":"42","project_name":"Only"}]`))
		}))
		defer server.Close()

		client := nucleus.NewClient(server.URL, "test-key")

		projectID, err := resolveDefaultProject(context.Background(), client, "")

		if err != nil {
			t.Fatalf("resolveDefaultProject() error: %v", err)
		}
		if projectID != "42" {
			t.Fatalf("expected 42, got %q", projectID)
		}
	})

	t.Run("multiple listed projects returns no default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`[{"project_id":"1"},{"project_id":"2"}]`))
		}))
		defer server.Close()

		client := nucleus.NewClient(server.URL, "test-key")

		projectID, err := resolveDefaultProject(context.Background(), client, "")

		if err != nil {
			t.Fatalf("resolveDefaultProject() error: %v", err)
		}
		if projectID != "" {
			t.Fatalf("expected empty default, got %q", projectID)
		}
	})

	t.Run("zero listed projects returns no default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`[]`))
		}))
		defer server.Close()

		client := nucleus.NewClient(server.URL, "test-key")

		projectID, err := resolveDefaultProject(context.Background(), client, "")

		if err != nil {
			t.Fatalf("resolveDefaultProject() error: %v", err)
		}
		if projectID != "" {
			t.Fatalf("expected empty default, got %q", projectID)
		}
	})

	t.Run("list error returns no default and error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := nucleus.NewClient(server.URL, "test-key")

		projectID, err := resolveDefaultProject(context.Background(), client, "")

		if err == nil {
			t.Fatal("expected error")
		}
		if projectID != "" {
			t.Fatalf("expected empty default, got %q", projectID)
		}
	})
}

func TestResolveIncludesDefaultProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NUC_API_KEY", "")
	t.Setenv("NUC_BASE_URL", "")
	t.Setenv("NUC_PROJECT", "")

	if err := config.Save(&config.Config{
		APIKey:         "file-key",
		BaseURL:        "https://example.com/api",
		DefaultProject: "file-project",
	}); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if cfg.DefaultProject != "file-project" {
		t.Fatalf("expected file-project, got %q", cfg.DefaultProject)
	}
}

func TestResolveDefaultProjectEnvOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NUC_API_KEY", "")
	t.Setenv("NUC_BASE_URL", "")
	t.Setenv("NUC_PROJECT", "env-project")

	if err := config.Save(&config.Config{
		APIKey:         "file-key",
		BaseURL:        "https://example.com/api",
		DefaultProject: "file-project",
	}); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if cfg.DefaultProject != "env-project" {
		t.Fatalf("expected env-project, got %q", cfg.DefaultProject)
	}
}

func TestNewIgnoresDefaultProjectAutoDetectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := New(&Config{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
}
