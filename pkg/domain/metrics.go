package domain

// FindingMetrics represents aggregated vulnerability discovery/remediation metrics
// over 30, 90, and 180-day rolling windows from GET /projects/{id}/findings/metrics.
type FindingMetrics struct {
	MetricDate         string `json:"metric_date"`
	Discovered30       int    `json:"projectmetricsdiscovered30"`
	Remediated30       int    `json:"projectmetricsremediated30"`
	RemediationDays30  int    `json:"projectmetricsremdays30"`
	Discovered90       int    `json:"projectmetricsdiscovered90"`
	Remediated90       int    `json:"projectmetricsremediated90"`
	RemediationDays90  int    `json:"projectmetricsremdays90"`
	Discovered180      int    `json:"projectmetricsdiscovered180"`
	Remediated180      int    `json:"projectmetricsremediated180"`
	RemediationDays180 int    `json:"projectmetricsremdays180"`
	Success            bool   `json:"success"`
}

// AssetGroupMetrics represents security metrics for a single asset group
// from GET /projects/{id}/assets/groups/metrics.
type AssetGroupMetrics struct {
	// Group identifier — the API returns results keyed per-group but the
	// group name is passed as a query parameter, so we track it locally.
	GroupName string `json:"-"`

	RiskScore                      int `json:"risk_score"`
	AssetCount                     int `json:"asset_count"`
	VulnCount                      int `json:"vuln_count"`
	VulnCountCritical              int `json:"vuln_count_critical"`
	VulnCountHigh                  int `json:"vuln_count_high"`
	AvgAgeCritical                 int `json:"avg_age_critical"`
	AvgAgeHigh                     int `json:"avg_age_high"`
	ChurnPct7d                     int `json:"churn_pct_7d"`
	ChurnPctCritical7d             int `json:"churn_pct_critical_7d"`
	ChurnPctHigh7d                 int `json:"churn_pct_high_7d"`
	PastDuePctCritical             int `json:"past_due_pct_critical"`
	PastDuePctHigh                 int `json:"past_due_pct_high"`
	MTTR7d                         int `json:"mttr_7d"`
	MTTRCritical7d                 int `json:"mttr_critical_7d"`
	MTTRHigh7d                     int `json:"mttr_high_7d"`
	MTTRPublishDate7d              int `json:"mttr_publish_date_7d"`
	MTTRPublishDateCritical7d      int `json:"mttr_publish_date_critical_7d"`
	MTTRPublishDateHigh7d          int `json:"mttr_publish_date_high_7d"`
	NucleusExploitedCount          int `json:"nucleus_exploited_count"`
	NucleusZeroDayCount            int `json:"nucleus_zero_day_count"`
	NucleusThreatCount             int `json:"nucleus_threat_count"`
	NucleusExploitedByMalwareCount int `json:"nucleus_exploited_by_malware_count"`
	ComplianceFailPct              int `json:"compliance_fail_pct"`
	CompliancePassPct              int `json:"compliance_pass_pct"`
	AssetExternalPct               int `json:"asset_external_pct"`
}

// AssetGroupMetricsOptions defines parameters for GET /assets/groups/metrics.
type AssetGroupMetricsOptions struct {
	// AssetGroups is required — up to 50 group names.
	AssetGroups []string
	// Metrics is optional — specific metric names to include.
	Metrics []string
}
