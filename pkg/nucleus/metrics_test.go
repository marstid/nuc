package nucleus

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marstid/nuc/pkg/domain"
)

func TestClient_GetFindingMetrics(t *testing.T) {
	fixture := loadFixture(t, "metrics_finding.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/metrics", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	metrics, err := client.GetFindingMetrics(context.Background(), "42")

	require.NoError(t, err)
	assert.Equal(t, "2024-10-17", metrics.MetricDate)
	assert.Equal(t, 45, metrics.Discovered30)
	assert.Equal(t, 38, metrics.Remediated30)
	assert.Equal(t, 12, metrics.RemediationDays30)
	assert.Equal(t, 142, metrics.Discovered90)
	assert.True(t, metrics.Success)
}

func TestClient_GetAssetGroupMetrics(t *testing.T) {
	fixture := loadFixture(t, "metrics_groups.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/assets/groups/metrics", r.URL.Path)

		// Verify asset_groups is JSON-encoded array in query param.
		var groups []string
		require.NoError(t, json.Unmarshal([]byte(r.URL.Query().Get("asset_groups")), &groups))
		assert.Equal(t, []string{"production", "staging"}, groups)

		// Verify optional metrics param.
		var metrics []string
		require.NoError(t, json.Unmarshal([]byte(r.URL.Query().Get("metrics")), &metrics))
		assert.Equal(t, []string{"risk_score", "vuln_count_critical"}, metrics)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	opts := &domain.AssetGroupMetricsOptions{
		AssetGroups: []string{"production", "staging"},
		Metrics:     []string{"risk_score", "vuln_count_critical"},
	}

	result, err := client.GetAssetGroupMetrics(context.Background(), "42", opts)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, 75, result[0].RiskScore)
	assert.Equal(t, 12, result[0].VulnCountCritical)
	assert.Equal(t, 60, result[1].RiskScore)
}

func TestClient_GetAssetGroupMetrics_RequiresGroups(t *testing.T) {
	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		// Should never be called.
		t.Error("unexpected HTTP request")
	})

	_, err := client.GetAssetGroupMetrics(context.Background(), "42", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one asset group is required")
}
