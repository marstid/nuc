package nucleus

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/marstid/nuc/pkg/domain"
)

func TestClient_ListScans(t *testing.T) {
	fixture := loadFixture(t, "scans_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/scans", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-apikey"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	scans, err := client.ListScans(context.Background(), "42", 0, 0)

	require.NoError(t, err)
	require.Len(t, scans, 2)
	assert.Equal(t, "1", scans[0].ID)
	assert.Equal(t, "Nessus Weekly", scans[0].Name)
	assert.Equal(t, "Weekly vulnerability scan", scans[0].Description)
	assert.Equal(t, "Nessus", scans[0].Type)
	assert.Equal(t, "Completed", scans[0].Status)
	assert.Equal(t, "2", scans[1].ID)
	assert.Equal(t, "Qualys Monthly", scans[1].Name)
	assert.Equal(t, "Processing", scans[1].Status)
}

func TestClient_ListScans_WithPagination(t *testing.T) {
	fixture := loadFixture(t, "scans_list.json")

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/projects/42/scans", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("start"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	scans, err := client.ListScans(context.Background(), "42", 10, 50)

	require.NoError(t, err)
	require.Len(t, scans, 2)
}

func TestClient_UploadScan(t *testing.T) {
	fixture := loadFixture(t, "scans_upload.json")

	// Create a temporary file to upload.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "scan_results.csv")
	fileContent := []byte("host,vuln,severity\n192.168.1.1,CVE-2024-001,High\n")
	require.NoError(t, os.WriteFile(tmpFile, fileContent, 0o644))

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/scans", r.URL.Path)

		// Verify multipart content type.
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		require.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediaType)

		// Parse multipart form to verify file content.
		reader := multipart.NewReader(r.Body, params["boundary"])
		part, err := reader.NextPart()
		require.NoError(t, err)
		assert.Equal(t, "file", part.FormName())
		assert.Equal(t, "scan_results.csv", part.FileName())

		body, err := io.ReadAll(part)
		require.NoError(t, err)
		assert.Equal(t, fileContent, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	result, err := client.UploadScan(context.Background(), "42", tmpFile, nil)

	require.NoError(t, err)
	assert.Equal(t, "3", result.ScanID)
	assert.Equal(t, "Scan uploaded successfully", result.Message)
}

func TestClient_UploadScan_WithOptions(t *testing.T) {
	fixture := loadFixture(t, "scans_upload.json")

	// Create a temporary file to upload.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nessus_scan.xml")
	fileContent := []byte("<NessusClientData_v2><Report></Report></NessusClientData_v2>")
	require.NoError(t, os.WriteFile(tmpFile, fileContent, 0o644))

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/42/scans", r.URL.Path)

		// Parse multipart form.
		contentType := r.Header.Get("Content-Type")
		_, params, err := mime.ParseMediaType(contentType)
		require.NoError(t, err)

		reader := multipart.NewReader(r.Body, params["boundary"])

		fields := make(map[string]string)
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			if part.FileName() != "" {
				// Skip file content but verify file name.
				assert.Equal(t, "nessus_scan.xml", part.FileName())
				io.ReadAll(part)
				continue
			}

			value, err := io.ReadAll(part)
			require.NoError(t, err)
			fields[part.FormName()] = string(value)
		}

		assert.Equal(t, "Weekly Nessus scan", fields["description"])
		assert.Equal(t, "Nessus", fields["scan_type"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	})

	opts := &domain.ScanUploadOptions{
		Description: "Weekly Nessus scan",
		Type:        "Nessus",
	}
	result, err := client.UploadScan(context.Background(), "42", tmpFile, opts)

	require.NoError(t, err)
	assert.Equal(t, "3", result.ScanID)
	assert.Equal(t, "Scan uploaded successfully", result.Message)
}
