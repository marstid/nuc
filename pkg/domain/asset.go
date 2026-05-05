package domain

// AssetType represents the type of an asset.
type AssetType string

// Asset type constants matching the Nucleus Security API.
const (
	AssetTypeHost       AssetType = "Host"
	AssetTypeWebApp     AssetType = "Web Application"
	AssetTypeDatabase   AssetType = "Database"
	AssetTypeContainer  AssetType = "Container"
	AssetTypeCloudAsset AssetType = "Cloud Asset"
	AssetTypeRepository AssetType = "Repository"
	AssetTypeMobileApp  AssetType = "Mobile Application"
	AssetTypeNetwork    AssetType = "Network"
	AssetTypeOther      AssetType = "Other"
)

// Asset represents a Nucleus Security asset within a project.
// Note: the Nucleus API returns most numeric fields as strings.
type Asset struct {
	ID                     string    `json:"asset_id"`
	Name                   string    `json:"asset_name"`
	Type                   AssetType `json:"asset_type"`
	IPAddress              string    `json:"ip_address"`
	DomainName             string    `json:"domain_name"`
	OperatingSystem        string    `json:"operating_system_name"`
	OperatingSystemVersion string    `json:"operating_system_version"`
	Criticality            string    `json:"asset_criticality"`
	Groups                 []string  `json:"asset_groups"`
	Location               string    `json:"asset_location"`
	Notes                  string    `json:"asset_notes"`
	DataSensitivityScore   string    `json:"asset_data_sensitivity_score"`
	Public                 string    `json:"asset_public"` // "0" or "1"
	Active                 bool      `json:"active"`
	ScanDate               string    `json:"scan_date"`
	RiskScore              string    `json:"asset_base_risk_score"`
	VulnCritical           string    `json:"finding_count_critical"`
	VulnHigh               string    `json:"finding_count_high"`
	VulnMedium             string    `json:"finding_count_medium"`
	VulnLow                string    `json:"finding_count_low"`
}

// AssetGroup represents a logical grouping of assets.
type AssetGroup struct {
	Name       string `json:"asset_group"`
	AssetCount int    `json:"asset_count"`
}

// AssetListOptions defines filters for listing assets.
type AssetListOptions struct {
	Start              *int
	Limit              *int
	IPAddress          string
	AssetName          string
	AssetGroups        string
	AssetType          AssetType
	InactiveAssets     *bool
	UnscannedAssets    *bool
	AssetsWithFindings *bool
}
