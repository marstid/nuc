package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/pkg/domain"
)

func newMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "View vulnerability metrics",
		Long:  "Retrieve vulnerability management metrics for a project.",
	}

	cmd.AddCommand(newMetricsFindingsCmd())
	cmd.AddCommand(newMetricsGroupsCmd())

	return cmd
}

// newMetricsFindingsCmd implements: nuc metrics findings
func newMetricsFindingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "findings",
		Short: "Get finding metrics (30/90/180-day windows)",
		Long:  "Retrieve aggregated vulnerability discovery and remediation metrics over 30, 90, and 180-day rolling windows.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			projectID, err := resolveProjectID()
			if err != nil {
				return err
			}

			client, err := requireClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			m, err := client.GetFindingMetrics(ctx, projectID)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, m.MetricDate)
				return nil
			}

			formatter := getFormatter()
			fields := []struct{ label, value string }{
				{"Metric Date", m.MetricDate},
				{"Discovered (30d)", strconv.Itoa(m.Discovered30)},
				{"Remediated (30d)", strconv.Itoa(m.Remediated30)},
				{"Avg Remediation Days (30d)", strconv.Itoa(m.RemediationDays30)},
				{"Discovered (90d)", strconv.Itoa(m.Discovered90)},
				{"Remediated (90d)", strconv.Itoa(m.Remediated90)},
				{"Avg Remediation Days (90d)", strconv.Itoa(m.RemediationDays90)},
				{"Discovered (180d)", strconv.Itoa(m.Discovered180)},
				{"Remediated (180d)", strconv.Itoa(m.Remediated180)},
				{"Avg Remediation Days (180d)", strconv.Itoa(m.RemediationDays180)},
			}

			headers := []string{"Metric", "Value"}
			rows := make([][]string, 0, len(fields))
			for _, f := range fields {
				rows = append(rows, []string{f.label, f.value})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}
}

// newMetricsGroupsCmd implements: nuc metrics groups --groups=<names> [--metrics=<names>]
func newMetricsGroupsCmd() *cobra.Command {
	var (
		groups  []string
		metrics []string
	)

	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Get metrics for asset groups",
		Long:  "Retrieve security metrics for one or more asset groups (up to 50 per request).",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if len(groups) == 0 {
				return fmt.Errorf("at least one --groups value is required")
			}

			projectID, err := resolveProjectID()
			if err != nil {
				return err
			}

			client, err := requireClient()
			if err != nil {
				return err
			}

			if len(groups) > 0 {
				groups, err = expandGroupGlobs(client, projectID, groups)
				if err != nil {
					return err
				}
			}

			opts := &domain.AssetGroupMetricsOptions{
				AssetGroups: groups,
				Metrics:     metrics,
			}

			ctx := context.Background()
			result, err := client.GetAssetGroupMetrics(ctx, projectID, opts)
			if err != nil {
				return err
			}

			if len(result) == 0 {
				fmt.Fprintln(os.Stdout, "No metrics found.")
				return nil
			}

			if flags.quiet {
				for i, g := range result {
					// In quiet mode, print the group index (name not in response body).
					fmt.Fprintf(os.Stdout, "%d\n", i)
					_ = g
				}
				return nil
			}

			formatter := getFormatter()

			// Build column headers from the requested (or all) metric names.
			// If specific metrics were requested, use those; otherwise derive from struct fields.
			var colKeys []string
			if len(metrics) > 0 {
				colKeys = metrics
			} else {
				// Default display columns.
				colKeys = []string{
					"risk_score", "asset_count", "vuln_count",
					"vuln_count_critical", "vuln_count_high",
					"mttr_7d", "nucleus_exploited_count",
				}
			}

			// Sort for consistent ordering.
			sort.Strings(colKeys)

			headers := append([]string{"Group"}, formatMetricHeaders(colKeys)...)
			rows := make([][]string, 0, len(result))

			for i, g := range result {
				groupLabel := groups[i] // correlate by position (API returns in same order)
				row := []string{groupLabel}
				for _, k := range colKeys {
					row = append(row, metricValue(g, k))
				}
				rows = append(rows, row)
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().StringSliceVar(&groups, "groups", nil, "Asset group names (comma-separated, supports glob patterns e.g. '*team-euc*', max 50)")
	cmd.Flags().StringSliceVar(&metrics, "metrics", nil, "Comma-separated metric names to include (optional)")
	_ = cmd.MarkFlagRequired("groups")

	return cmd
}

// metricValue extracts a named metric from an AssetGroupMetrics struct.
func metricValue(g domain.AssetGroupMetrics, key string) string {
	switch key {
	case "risk_score":
		return strconv.Itoa(g.RiskScore)
	case "asset_count":
		return strconv.Itoa(g.AssetCount)
	case "vuln_count":
		return strconv.Itoa(g.VulnCount)
	case "vuln_count_critical":
		return strconv.Itoa(g.VulnCountCritical)
	case "vuln_count_high":
		return strconv.Itoa(g.VulnCountHigh)
	case "avg_age_critical":
		return strconv.Itoa(g.AvgAgeCritical)
	case "avg_age_high":
		return strconv.Itoa(g.AvgAgeHigh)
	case "churn_pct_7d":
		return strconv.Itoa(g.ChurnPct7d)
	case "mttr_7d":
		return strconv.Itoa(g.MTTR7d)
	case "mttr_critical_7d":
		return strconv.Itoa(g.MTTRCritical7d)
	case "mttr_high_7d":
		return strconv.Itoa(g.MTTRHigh7d)
	case "nucleus_exploited_count":
		return strconv.Itoa(g.NucleusExploitedCount)
	case "nucleus_zero_day_count":
		return strconv.Itoa(g.NucleusZeroDayCount)
	case "nucleus_threat_count":
		return strconv.Itoa(g.NucleusThreatCount)
	case "compliance_fail_pct":
		return strconv.Itoa(g.ComplianceFailPct)
	case "compliance_pass_pct":
		return strconv.Itoa(g.CompliancePassPct)
	case "asset_external_pct":
		return strconv.Itoa(g.AssetExternalPct)
	default:
		return ""
	}
}

// formatMetricHeaders converts snake_case metric names to Title Case display strings.
func formatMetricHeaders(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		parts := strings.Split(k, "_")
		for j, p := range parts {
			if len(p) > 0 {
				parts[j] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		out[i] = strings.Join(parts, " ")
	}
	return out
}
