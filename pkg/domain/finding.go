package domain

import (
	"encoding/json"
	"strconv"
)

// FlexFloat is a custom JSON type for the EPSS score field.
// The Nucleus API returns it inconsistently as float (0.01), int (0), or empty string ("").
// FlexFloat unmarshals a JSON value that may be a float, int, or empty string.
// This is needed because the Nucleus API returns epss_score inconsistently.
type FlexFloat struct {
	Value float64
	Set   bool
}

// FlexInt is a custom JSON type for integer fields that the Nucleus API
// returns inconsistently as numbers, numeric strings, or empty strings.
type FlexInt struct {
	Value int
	Set   bool
}

// UnmarshalJSON implements json.Unmarshaler for FlexInt.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		f.Value = 0
		f.Set = false
		return nil
	}

	if len(s) >= 2 && s[0] == '"' {
		inner := s[1 : len(s)-1]
		if inner == "" {
			f.Value = 0
			f.Set = false
			return nil
		}

		v, err := strconv.Atoi(inner)
		if err != nil {
			return err
		}
		f.Value = v
		f.Set = true
		return nil
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	f.Value = v
	f.Set = true
	return nil
}

// MarshalJSON implements json.Marshaler for FlexInt.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

// UnmarshalJSON implements json.Unmarshaler for FlexFloat.
func (f *FlexFloat) UnmarshalJSON(data []byte) error {
	s := string(data)
	// null → zero value
	if s == "null" {
		f.Value = 0
		f.Set = false
		return nil
	}
	// quoted string (e.g. "" or "0.85") — strip quotes and parse
	if len(s) >= 2 && s[0] == '"' {
		inner := s[1 : len(s)-1]
		if inner == "" {
			f.Value = 0
			f.Set = false
			return nil
		}
		v, err := strconv.ParseFloat(inner, 64)
		if err != nil {
			return err
		}
		f.Value = v
		f.Set = v != 0
		return nil
	}
	// bare numeric literal (int or float)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	f.Value = v
	f.Set = v != 0
	return nil
}

// MarshalJSON implements json.Marshaler for FlexFloat.
func (f FlexFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func (f FlexFloat) String() string {
	if !f.Set {
		return ""
	}
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

// Severity represents the severity level of a finding.
type Severity string

// Severity constants for finding severity levels.
const (
	SeverityCritical Severity = "Critical"
	SeverityHigh     Severity = "High"
	SeverityMedium   Severity = "Medium"
	SeverityLow      Severity = "Low"
	SeverityInfo     Severity = "Info"
)

// FindingStatus represents the current status of a finding.
type FindingStatus string

// FindingStatus constants for finding status values.
const (
	FindingStatusActive                 FindingStatus = "Active"
	FindingStatusInProgress             FindingStatus = "In Progress"
	FindingStatusFixed                  FindingStatus = "Fixed"
	FindingStatusAcceptedRisk           FindingStatus = "Accepted Risk"
	FindingStatusFalsePositive          FindingStatus = "False Positive"
	FindingStatusExceptionRequested     FindingStatus = "Exception Requested"
	FindingStatusExceptionGranted       FindingStatus = "Exception Granted"
	FindingStatusDuplicate              FindingStatus = "Duplicate"
	FindingStatusManuallyMitigated      FindingStatus = "Manually Mitigated"
	FindingStatusScanMitigated          FindingStatus = "Scan Mitigated"
	FindingStatusWaitingForVerification FindingStatus = "Waiting For Verification"
	FindingStatusWaitingForThirdParty   FindingStatus = "Waiting For Third Party"
)

// Finding represents a vulnerability finding returned by the search endpoint.
// All IDs and numeric values are returned as strings by the API.
type Finding struct {
	ScanID            string `json:"scan_id"`
	FindingID         string `json:"finding_id"`
	FindingNumber     string `json:"finding_number"`
	Name              string `json:"finding_name"`
	Severity          string `json:"finding_severity"`
	SeverityAdjusted  string `json:"finding_severity_adjusted"`
	ScanDate          string `json:"scan_date"`
	ScanType          string `json:"scan_type"`
	CVE               string `json:"finding_cve"`
	Result            string `json:"finding_result"`
	AssetID           string `json:"asset_id"`
	AssetName         string `json:"asset_name"`
	IPAddress         string `json:"ip_address"`
	Discovered        string `json:"finding_discovered"`
	Exploitable       string `json:"finding_exploitable"`
	Status            string `json:"justification_status_name"`
	JustificationText string `json:"justification_text"`
	DueDate           string `json:"due_date"`
	Description       string `json:"finding_description"`
	Recommendation    string `json:"finding_recommendation"`
	Port              string `json:"finding_port"`
	Path              string `json:"finding_path"`
	Package           string `json:"finding_package"`
	PackageVersion    string `json:"finding_package_version"`
	// EPSSScore uses FlexFloat because the API returns it as float (0.01), int (0), or empty string ("").
	EPSSScore                 FlexFloat `json:"epss_score"`
	JustificationIsMitigating FlexInt   `json:"justification_is_mitigating"`
	MandiantExploited         string    `json:"mandiant_exploit_in_the_wild"`
	MandiantRiskRating        string    `json:"mandiant_risk_rating"`
}

// MitigatedFinding represents a mitigated finding returned by the mitigated endpoint.
type MitigatedFinding struct {
	FindingNumber   string   `json:"finding_number"`
	Name            string   `json:"finding_name"`
	Severity        string   `json:"finding_severity"`
	Severities      []string `json:"finding_severities"`
	Discovered      string   `json:"finding_discovered"`
	Exploitable     string   `json:"finding_exploitable"`
	ScanType        string   `json:"scan_type"`
	ManualMitigated string   `json:"manual_mitigated"`
	RemediatedDate  string   `json:"finding_remediated_date"`
	RemediationDays string   `json:"finding_remediation_days"`
	TotalOpen       string   `json:"total_open"`
	TotalManual     string   `json:"total_manual"`
	TotalMitigated  string   `json:"total_mitigated"`
}

// FindingSearch defines the search criteria for POST /projects/{id}/findings/search.
// All string fields accept wildcards (* or %) and most can be arrays.
type FindingSearch struct {
	AssetID             *int     `json:"asset_id,omitempty"`
	AssetName           string   `json:"asset_name,omitempty"`
	IPAddress           string   `json:"ip_address,omitempty"`
	AssetGroups         []string `json:"asset_groups,omitempty"`
	ScanType            string   `json:"scan_type,omitempty"`
	FindingCVE          string   `json:"finding_cve,omitempty"`
	FindingName         string   `json:"finding_name,omitempty"`
	FindingNumber       string   `json:"finding_number,omitempty"`
	FindingSeverity     string   `json:"finding_severity,omitempty"`
	FindingExploitable  string   `json:"finding_exploitable,omitempty"`
	JustificationStatus []string `json:"justification_status,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
	Team                string   `json:"team,omitempty"`
	IncludeTotalCount   bool     `json:"include_total_count,omitempty"`
}

// FindingListOptions defines filters for listing findings.
type FindingListOptions struct {
	Start    *int
	Limit    *int
	Severity Severity
	Status   FindingStatus
}

// FindingOverview represents the summary returned by GET /projects/{id}/findings/overview.
// Most counts are returned as strings by the API.
type FindingOverview struct {
	CritHigh           string `json:"finding_count_crithigh"`
	CVECount           string `json:"finding_count_cve"`
	IAVACount          string `json:"finding_count_iava"`
	ExploitableCount   string `json:"finding_count_exploitable"`
	Critical           string `json:"finding_count_critical"`
	High               string `json:"finding_count_high"`
	Medium             string `json:"finding_count_medium"`
	Low                string `json:"finding_count_low"`
	VulnerabilityScore int    `json:"finding_vulnerability_score"`
}

// FindingTrendPoint represents a single data point in the vulnDiscoveredBar trend.
// All severity counts are returned as strings by the API.
type FindingTrendPoint struct {
	Date          string `json:"vuln_date_short"`
	Critical      string `json:"Critical"`
	High          string `json:"High"`
	Medium        string `json:"Medium"`
	Low           string `json:"Low"`
	Informational string `json:"Informational"`
}

// FindingTrend represents trend data returned by GET /projects/{id}/findings/trend.
type FindingTrend struct {
	DataPoints []FindingTrendPoint `json:"vulnDiscoveredBar"`
}

// MitigatedOptions defines filters for getting mitigated findings.
type MitigatedOptions struct {
	Start     *int
	Limit     *int
	StartDate string
}

// TrendOptions defines filters for getting finding trends.
type TrendOptions struct {
	StartDate   string
	EndDate     string
	AssetGroups []string
}

// FindingSummaryFilter represents a single filter condition for the findings summary endpoint.
// The API accepts "value" as either a string or array of strings depending on the property.
// Use Value for single-value filters, or Values for multi-value filters
// (finding_severity, team_names, asset_groups).
// A custom MarshalJSON ensures the correct shape is sent to the API.
type FindingSummaryFilter struct {
	Property   string
	Value      string
	Values     []string
	ExactMatch bool
}

// MarshalJSON serializes the filter, outputting "value" as either a string or []string
// depending on which field is populated.
func (f FindingSummaryFilter) MarshalJSON() ([]byte, error) {
	type wire struct {
		Property   string `json:"property"`
		Value      any    `json:"value"`
		ExactMatch bool   `json:"exactMatch"`
	}

	var value any
	if len(f.Values) > 0 {
		value = f.Values
	} else {
		value = f.Value
	}

	return json.Marshal(wire{
		Property:   f.Property,
		Value:      value,
		ExactMatch: f.ExactMatch,
	})
}

// FindingSummarySort represents a sort rule for the findings summary endpoint.
type FindingSummarySort struct {
	Property  string `json:"property"`
	Direction string `json:"direction"`
}

// FindingSummaryRequest is the POST body for /projects/{id}/findings/summary.
type FindingSummaryRequest struct {
	Filter []FindingSummaryFilter `json:"filter,omitempty"`
	Sort   []FindingSummarySort   `json:"sort,omitempty"`
}

// FindingSummary represents a finding grouped by finding_number from the summary endpoint.
// Unlike Finding which represents an individual instance per asset, FindingSummary aggregates
// across assets showing counts and combined statuses/severities.
type FindingSummary struct {
	FindingNumber        string    `json:"finding_number"`
	FindingName          string    `json:"finding_name"`
	FindingSeverity      string    `json:"finding_severity"`
	FindingSeverities    []string  `json:"finding_severities"`
	FindingStatuses      []string  `json:"finding_statuses"`
	FindingStatus        string    `json:"finding_status"`
	FindingDiscovered    string    `json:"finding_discovered"`
	FindingExploitable   FlexInt   `json:"finding_exploitable"`
	FindingPinned        bool      `json:"finding_pinned"`
	FindingSeverityScore FlexInt   `json:"finding_severity_score"`
	ScanType             string    `json:"scan_type"`
	ScanDate             string    `json:"scan_date"`
	AssetCount           FlexInt   `json:"asset_count"`
	FindingCount         FlexInt   `json:"finding_count"`
	AssetFixedCount      FlexInt   `json:"asset_fixed_count"`
	AssetMitigatedCount  FlexInt   `json:"asset_mitigated_count"`
	FindingCVE           string    `json:"finding_cve"`
	FindingIAVA          string    `json:"finding_iava"`
	IssueOpenCount       FlexInt   `json:"issue_open_count"`
	IssueClosedCount     FlexInt   `json:"issue_closed_count"`
	EPSSScore            FlexFloat `json:"epss_score"`
	CISAVulnName         string    `json:"cisa_vulnerability_name"`
}
