package nucleus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/marstid/nuc/pkg/domain"
)

// loadFixture loads a test fixture file from testdata/.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture: %s", name)
	return data
}

// newTestClient creates a Client pointing at a test server.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, "test-api-key",
		WithRetryConfig(&RetryConfig{
			MaxRetries:     0,
			InitialBackoff: 0,
			MaxBackoff:     0,
			Multiplier:     1,
			Jitter:         0,
		}),
	)
}

func TestClient_ListProjects(t *testing.T) {
	fixture := loadFixture(t, "projects_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-apikey"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	projects, err := client.List(context.Background())

	require.NoError(t, err)
	require.Len(t, projects, 2)
	assert.Equal(t, "1", projects[0].ID)
	assert.Equal(t, "Production", projects[0].Name)
	assert.Equal(t, "2", projects[1].ID)
	assert.Equal(t, "Development", projects[1].Name)
}

func TestClient_ListProjects_Unauthorized(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.List(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestClient_ListProjects_ServerError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal server error"}`))
	})

	_, err := client.List(context.Background())

	require.Error(t, err)
	var apiErr *domain.APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}

func TestClient_GetProject(t *testing.T) {
	fixture := loadFixture(t, "projects_get.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	project, err := client.Get(context.Background(), "42")

	require.NoError(t, err)
	assert.Equal(t, "42", project.ID)
	assert.Equal(t, "Production", project.Name)
	assert.Equal(t, "Production environment vulnerabilities", project.Description)
}

func TestClient_GetProject_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Project not found"}`))
	})

	_, err := client.Get(context.Background(), "99999")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestClient_GetRiskScore(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/riskscore", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"project_id": "42", "score": 78}`))
	})

	riskScore, err := client.GetRiskScore(context.Background(), "42")

	require.NoError(t, err)
	assert.Equal(t, 78, riskScore.Score)
}

func TestClient_GetRiskScore_RateLimited(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"Rate limit exceeded"}`))
	})

	_, err := client.GetRiskScore(context.Background(), "42")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrRateLimited)
}
