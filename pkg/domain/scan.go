package domain

// Scan represents a vulnerability scan in Nucleus Security.
// Note: the API returns scan_id as a string (e.g. "2020927").
type Scan struct {
	ID          string `json:"scan_id"`
	Name        string `json:"scan_name"`
	Description string `json:"scan_description"`
	Type        string `json:"scan_type"`
	Status      string `json:"scan_status"`
	ScanDate    string `json:"scan_date"`
	AssetCount  string `json:"asset_count"`
}

// ScanResult represents the result of uploading a scan.
type ScanResult struct {
	ScanID  string `json:"scan_id"`
	Message string `json:"message"`
}

// ScanUploadOptions defines optional parameters for uploading a scan.
type ScanUploadOptions struct {
	Description string
	Type        string
}
