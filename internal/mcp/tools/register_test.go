package tools

import (
	"strings"
	"testing"
)

func TestResolveProjectID(t *testing.T) {
	t.Run("explicit wins over default", func(t *testing.T) {
		svc := &Services{DefaultProject: "default"}

		projectID, err := svc.resolveProjectID("explicit")

		if err != nil {
			t.Fatalf("resolveProjectID() error: %v", err)
		}
		if projectID != "explicit" {
			t.Fatalf("expected explicit, got %q", projectID)
		}
	})

	t.Run("default used when explicit is empty", func(t *testing.T) {
		svc := &Services{DefaultProject: "default"}

		projectID, err := svc.resolveProjectID("")

		if err != nil {
			t.Fatalf("resolveProjectID() error: %v", err)
		}
		if projectID != "default" {
			t.Fatalf("expected default, got %q", projectID)
		}
	})

	t.Run("error when no project is available", func(t *testing.T) {
		svc := &Services{}

		projectID, err := svc.resolveProjectID("")

		if err == nil {
			t.Fatal("expected error")
		}
		if projectID != "" {
			t.Fatalf("expected empty project ID, got %q", projectID)
		}
		if !strings.Contains(err.Error(), "project_id is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
