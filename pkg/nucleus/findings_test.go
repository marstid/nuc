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

func TestClient_ListFindings(t *testing.T) {
	fixture := loadFixture(t, "findings_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-apikey"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	findings, err := client.ListFindings(context.Background(), "42", nil)

	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "9997101", findings[0].FindingID)
	assert.Equal(t, "VULN-001", findings[0].FindingNumber)
	assert.Equal(t, "SQL Injection", findings[0].Name)
	assert.Equal(t, "Critical", findings[0].Severity)
	assert.Equal(t, "Active", findings[0].Status)
	assert.Equal(t, "VULN-002", findings[1].FindingNumber)
	assert.Equal(t, "XSS Reflected", findings[1].Name)
	assert.Equal(t, "High", findings[1].Severity)
}

func TestClient_ListFindings_WithFilters(t *testing.T) {
	fixture := loadFixture(t, "findings_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("start"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "Critical", r.URL.Query().Get("finding_severity"))
		assert.Equal(t, "Active", r.URL.Query().Get("finding_status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	start := 0
	limit := 10
	opts := &domain.FindingListOptions{
		Start:    &start,
		Limit:    &limit,
		Severity: domain.SeverityCritical,
		Status:   domain.FindingStatusActive,
	}

	findings, err := client.ListFindings(context.Background(), "42", opts)

	require.NoError(t, err)
	require.Len(t, findings, 2)
}

func TestClient_GetFinding(t *testing.T) {
	fixture := loadFixture(t, "findings_get.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/VULN-001", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	finding, err := client.GetFinding(context.Background(), "42", "VULN-001")

	require.NoError(t, err)
	assert.Equal(t, "9997101", finding.FindingID)
	assert.Equal(t, "VULN-001", finding.FindingNumber)
	assert.Equal(t, "SQL Injection", finding.Name)
	assert.Equal(t, "Critical", finding.Severity)
	assert.Equal(t, "Active", finding.Status)
	assert.Equal(t, "CVE-2024-1001", finding.CVE)
	assert.Equal(t, "1", finding.Exploitable)
}

func TestClient_GetFinding_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Finding not found"}`))
	})

	_, err := client.GetFinding(context.Background(), "42", "VULN-999")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestClient_SearchFindings(t *testing.T) {
	fixture := loadFixture(t, "findings_search.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/findings/search", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload domain.FindingSearch
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "Critical", payload.FindingSeverity)
		assert.Equal(t, "web-*", payload.AssetName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	search := &domain.FindingSearch{
		FindingSeverity: "Critical",
		AssetName:       "web-*",
	}

	findings, err := client.SearchFindings(context.Background(), "42", search, 0, 0)

	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "VULN-001", findings[0].FindingNumber)
	assert.Equal(t, "Critical", findings[0].Severity)
	assert.Equal(t, 0, findings[0].JustificationIsMitigating.Value)
	assert.True(t, findings[0].JustificationIsMitigating.Set)
	assert.Equal(t, 1, findings[1].JustificationIsMitigating.Value)
	assert.True(t, findings[1].JustificationIsMitigating.Set)
}

func TestClient_SearchFindings_WithPagination(t *testing.T) {
	fixture := loadFixture(t, "findings_search.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/findings/search", r.URL.Path)
		assert.Equal(t, "50", r.URL.Query().Get("start"))
		assert.Equal(t, "500", r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	search := &domain.FindingSearch{
		FindingSeverity: "Critical",
	}

	findings, err := client.SearchFindings(context.Background(), "42", search, 50, 500)

	require.NoError(t, err)
	require.Len(t, findings, 2)
}

func TestClient_UpdateFinding(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/projects/42/findings", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var input service.UpdateFindingInput
		require.NoError(t, json.Unmarshal(body, &input))
		assert.Equal(t, "VULN-001", input.FindingNumber)
		assert.Equal(t, domain.FindingStatusFixed, input.Status)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	input := &service.UpdateFindingInput{
		FindingNumber: "VULN-001",
		Status:        domain.FindingStatusFixed,
	}

	err := client.UpdateFinding(context.Background(), "42", input)

	require.NoError(t, err)
}

func TestClient_BulkUpdateFindings(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/projects/42/findings/bulk", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload struct {
			Updates []service.UpdateFindingInput `json:"updates"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		require.Len(t, payload.Updates, 2)
		assert.Equal(t, "VULN-001", payload.Updates[0].FindingNumber)
		assert.Equal(t, "VULN-002", payload.Updates[1].FindingNumber)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	updates := []service.UpdateFindingInput{
		{FindingNumber: "VULN-001", Status: domain.FindingStatusFixed},
		{FindingNumber: "VULN-002", Status: domain.FindingStatusInProgress},
	}

	err := client.BulkUpdateFindings(context.Background(), "42", updates)

	require.NoError(t, err)
}

func TestClient_GetMitigatedFindings(t *testing.T) {
	fixture := loadFixture(t, "findings_mitigated.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/mitigated", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("start"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("start_date"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	start := 0
	limit := 50
	opts := &domain.MitigatedOptions{
		Start:     &start,
		Limit:     &limit,
		StartDate: "2024-01-01",
	}

	findings, err := client.GetMitigatedFindings(context.Background(), "42", opts)

	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "ALAS2-2023-2316", findings[0].FindingNumber)
	assert.Equal(t, "High", findings[0].Severity)
	assert.Equal(t, "854", findings[0].RemediationDays)
	assert.Equal(t, "20", findings[0].TotalMitigated)
}

func TestClient_GetFindingTrend(t *testing.T) {
	fixture := loadFixture(t, "findings_trend.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/trend", r.URL.Path)
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("start_date"))
		assert.Equal(t, "2024-03-01", r.URL.Query().Get("end_date"))
		assert.Equal(t, "group1,group2", r.URL.Query().Get("asset_groups"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	opts := &domain.TrendOptions{
		StartDate:   "2024-01-01",
		EndDate:     "2024-03-01",
		AssetGroups: []string{"group1", "group2"},
	}

	trend, err := client.GetFindingTrend(context.Background(), "42", opts)

	require.NoError(t, err)
	require.Len(t, trend.DataPoints, 1)
	assert.Equal(t, "2024-01-15", trend.DataPoints[0].Date)
	assert.Equal(t, "3", trend.DataPoints[0].Critical)
	assert.Equal(t, "10", trend.DataPoints[0].High)
	assert.Equal(t, "20", trend.DataPoints[0].Medium)
	assert.Equal(t, "5", trend.DataPoints[0].Low)
	assert.Equal(t, "1", trend.DataPoints[0].Informational)
}

func TestClient_GetFindingOverview(t *testing.T) {
	fixture := loadFixture(t, "findings_overview.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/overview", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	overview, err := client.GetFindingOverview(context.Background(), "42")

	require.NoError(t, err)
	assert.Equal(t, "5", overview.Critical)
	assert.Equal(t, "12", overview.High)
	assert.Equal(t, "15", overview.Medium)
	assert.Equal(t, "8", overview.Low)
	assert.Equal(t, "17", overview.CritHigh)
	assert.Equal(t, "12", overview.ExploitableCount)
	assert.Equal(t, 42000, overview.VulnerabilityScore)
}

func TestClient_GetFrameworks(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/findings/frameworks", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["NIST CSF", "ISO 27001", "PCI DSS"]`))
	})

	frameworks, err := client.GetFrameworks(context.Background(), "42")

	require.NoError(t, err)
	require.Len(t, frameworks, 3)
	assert.Equal(t, "NIST CSF", frameworks[0])
	assert.Equal(t, "ISO 27001", frameworks[1])
	assert.Equal(t, "PCI DSS", frameworks[2])
}
