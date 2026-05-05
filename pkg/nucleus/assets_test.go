package nucleus

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"
)

func TestClient_ListAssets(t *testing.T) {
	fixture := loadFixture(t, "assets_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/assets", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-apikey"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	assets, err := client.ListAssets(context.Background(), "42", nil)

	require.NoError(t, err)
	require.Len(t, assets, 2)
	assert.Equal(t, "1", assets[0].ID)
	assert.Equal(t, "web-01", assets[0].Name)
	assert.Equal(t, domain.AssetTypeHost, assets[0].Type)
	assert.Equal(t, "10.0.0.1", assets[0].IPAddress)
	assert.Equal(t, "2", assets[1].ID)
	assert.Equal(t, "api-server", assets[1].Name)
	assert.Equal(t, domain.AssetTypeWebApp, assets[1].Type)
}

func TestClient_ListAssets_WithFilters(t *testing.T) {
	fixture := loadFixture(t, "assets_list.json")

	start := 10
	limit := 25
	inactive := true

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/assets", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "10", q.Get("start"))
		assert.Equal(t, "25", q.Get("limit"))
		assert.Equal(t, "192.168.1.1", q.Get("ip_address"))
		assert.Equal(t, "web", q.Get("asset_name"))
		assert.Equal(t, "production", q.Get("asset_groups"))
		assert.Equal(t, "Host", q.Get("asset_type"))
		assert.Equal(t, "true", q.Get("inactive_assets"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	opts := &domain.AssetListOptions{
		Start:          &start,
		Limit:          &limit,
		IPAddress:      "192.168.1.1",
		AssetName:      "web",
		AssetGroups:    "production",
		AssetType:      domain.AssetTypeHost,
		InactiveAssets: &inactive,
	}

	assets, err := client.ListAssets(context.Background(), "42", opts)

	require.NoError(t, err)
	require.Len(t, assets, 2)
}

func TestClient_GetAsset(t *testing.T) {
	fixture := loadFixture(t, "assets_get.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/assets/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	asset, err := client.GetAsset(context.Background(), "42", "123")

	require.NoError(t, err)
	assert.Equal(t, "123", asset.ID)
	assert.Equal(t, "db-primary", asset.Name)
	assert.Equal(t, domain.AssetTypeDatabase, asset.Type)
	assert.Equal(t, "10.0.1.50", asset.IPAddress)
	assert.Equal(t, "Critical", asset.Criticality)
	assert.Contains(t, asset.Groups, "production")
	assert.Contains(t, asset.Groups, "database-tier")
}

func TestClient_GetAsset_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Asset not found"}`))
	})

	_, err := client.GetAsset(context.Background(), "42", "99999")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestClient_CreateAsset(t *testing.T) {
	fixture := loadFixture(t, "assets_get.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/assets", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody service.CreateAssetInput
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "db-primary", reqBody.Name)
		assert.Equal(t, domain.AssetTypeDatabase, reqBody.Type)
		assert.Equal(t, "10.0.1.50", reqBody.IPAddress)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(fixture)
	})

	input := &service.CreateAssetInput{
		Name:      "db-primary",
		Type:      domain.AssetTypeDatabase,
		IPAddress: "10.0.1.50",
		Groups:    []string{"production", "database-tier"},
	}

	asset, err := client.CreateAsset(context.Background(), "42", input)

	require.NoError(t, err)
	assert.Equal(t, "123", asset.ID)
	assert.Equal(t, "db-primary", asset.Name)
	assert.Equal(t, domain.AssetTypeDatabase, asset.Type)
}

func TestClient_UpdateAsset(t *testing.T) {
	newName := "db-primary-updated"
	fixture := []byte(`{
		"asset_id": "123",
		"asset_name": "db-primary-updated",
		"asset_type": "Database",
		"ip_address": "10.0.1.50",
		"active": true
	}`)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/projects/42/assets/123", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody service.UpdateAssetInput
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, &newName, reqBody.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	input := &service.UpdateAssetInput{
		Name: &newName,
	}

	asset, err := client.UpdateAsset(context.Background(), "42", "123", input)

	require.NoError(t, err)
	assert.Equal(t, "123", asset.ID)
	assert.Equal(t, "db-primary-updated", asset.Name)
}

func TestClient_DeleteAsset(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/42/assets/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteAsset(context.Background(), "42", "123")

	require.NoError(t, err)
}

func TestClient_ListAssetGroups(t *testing.T) {
	fixture := []byte(`[
		{"asset_group": "production", "asset_count": 15},
		{"asset_group": "staging", "asset_count": 8}
	]`)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/assets/groups", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	groups, err := client.ListAssetGroups(context.Background(), "42")

	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "production", groups[0].Name)
	assert.Equal(t, 15, groups[0].AssetCount)
	assert.Equal(t, "staging", groups[1].Name)
	assert.Equal(t, 8, groups[1].AssetCount)
}

func TestClient_CreateAssetGroup(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/assets/groups", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody map[string]string
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "new-group", reqBody["asset_group"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{}`))
	})

	err := client.CreateAssetGroup(context.Background(), "42", "new-group")

	require.NoError(t, err)
}

func TestClient_DeleteAssetGroup(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/projects/42/assets/groups", r.URL.Path)
		assert.Equal(t, "old-group", r.URL.Query().Get("asset_group"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	})

	err := client.DeleteAssetGroup(context.Background(), "42", "old-group")

	require.NoError(t, err)
}
