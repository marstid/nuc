package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListProjectsTool(t *testing.T) {
	cfg, err := Resolve("", "")
	if err != nil {
		t.Skipf("skipping live test: %v", err)
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

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "list_projects",
	})
	if err != nil {
		t.Fatalf("CallTool list_projects: %v", err)
	}

	if result.IsError {
		t.Fatalf("list_projects returned error: %v", result.Content)
	}

	t.Logf("list_projects returned %d content items", len(result.Content))
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			t.Logf("Response (first 500 chars): %s", truncate(tc.Text, 500))
		}
	}
}

func TestGetFindingOverviewTool(t *testing.T) {
	cfg, err := Resolve("", "")
	if err != nil {
		t.Skipf("skipping live test: %v", err)
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

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_finding_overview",
		Arguments: map[string]any{"project_id": "3"},
	})
	if err != nil {
		t.Fatalf("CallTool get_finding_overview: %v", err)
	}

	if result.IsError {
		t.Fatalf("get_finding_overview returned error: %v", result.Content)
	}

	t.Logf("get_finding_overview returned %d content items", len(result.Content))
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			t.Logf("Response (first 500 chars): %s", truncate(tc.Text, 500))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func TestListTeamsTool(t *testing.T) {
	cfg, err := Resolve("", "")
	if err != nil {
		t.Skipf("skipping live test: %v", err)
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

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_teams",
		Arguments: map[string]any{"project_id": "3"},
	})
	if err != nil {
		t.Fatalf("CallTool list_teams: %v", err)
	}

	if result.IsError {
		t.Fatalf("list_teams returned error: %v", result.Content)
	}

	t.Logf("list_teams returned %d content items", len(result.Content))
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			t.Logf("Response: %s", tc.Text)
		}
	}
}
