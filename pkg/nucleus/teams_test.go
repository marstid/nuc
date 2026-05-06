package nucleus

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListTeams(t *testing.T) {
	fixture := []byte(`[
		{"team_id": "1", "team_name": "team-euc", "project_id": "42", "asset_groups": ["10", "20"]},
		{"team_id": "2", "team_name": "team-mdd", "project_id": "42", "asset_groups": ["30"]}
	]`)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/teams", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-apikey"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	teams, err := client.ListTeams(context.Background(), "42")

	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, "1", teams[0].TeamID)
	assert.Equal(t, "team-euc", teams[0].TeamName)
	assert.Equal(t, "42", teams[0].ProjectID)
	assert.Equal(t, []string{"10", "20"}, teams[0].AssetGroups)
	assert.Equal(t, "2", teams[1].TeamID)
	assert.Equal(t, "team-mdd", teams[1].TeamName)
}

func TestClient_ListTeams_Empty(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects/42/teams", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	teams, err := client.ListTeams(context.Background(), "42")

	require.NoError(t, err)
	assert.Empty(t, teams)
}
