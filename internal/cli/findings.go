package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/internal/cli/output"
	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"
)

func newFindingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "findings",
		Short: "Manage vulnerability findings",
		Long:  "List, search, view, and update vulnerability findings within a Nucleus Security project.",
	}

	cmd.AddCommand(newFindingsSearchCmd())
	cmd.AddCommand(newFindingsUpdateCmd())
	cmd.AddCommand(newFindingsBulkUpdateCmd())
	cmd.AddCommand(newFindingsOverviewCmd())
	cmd.AddCommand(newFindingsMitigatedCmd())
	cmd.AddCommand(newFindingsTrendCmd())
	cmd.AddCommand(newFindingsFrameworksCmd())

	return cmd
}

func newFindingsSearchCmd() *cobra.Command {
	var (
		assetName   string
		assetGroups []string
		findingName string
		cve         string
		severity    string
		status      string
		scanType    string
		exploitable string
		start       int
		limit       int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search findings with filters",
		Long:  "Search findings using flexible criteria. Supports wildcards (* or %) in string fields.",
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

			if len(assetGroups) > 0 {
				assetGroups, err = expandGroupGlobs(client, projectID, assetGroups)
				if err != nil {
					return err
				}
			}

			search := &domain.FindingSearch{
				AssetName:          assetName,
				AssetGroups:        assetGroups,
				FindingName:        findingName,
				FindingCVE:         cve,
				FindingSeverity:    severity,
				ScanType:           scanType,
				FindingExploitable: exploitable,
			}
			if status != "" {
				search.JustificationStatus = []string{status}
			}

			ctx := context.Background()
			findings, err := client.SearchFindings(ctx, projectID, search, start, limit)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, f := range findings {
					fmt.Fprintln(os.Stdout, f.FindingNumber)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"ID", "Number", "Name", "Severity", "Status", "Asset", "CVE", "Scan Type", "Discovered"}
			rows := make([][]string, 0, len(findings))

			for _, f := range findings {
				rows = append(rows, []string{
					f.FindingID,
					f.FindingNumber,
					f.Name,
					f.Severity,
					f.Status,
					f.AssetName,
					f.CVE,
					f.ScanType,
					f.Discovered,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().StringVar(&assetName, "asset-name", "", "Filter by asset name (supports wildcards * and %)")
	cmd.Flags().StringSliceVar(&assetGroups, "groups", nil, "Filter by asset groups (comma-separated, supports glob patterns e.g. '*team-euc*')")
	cmd.Flags().StringVar(&findingName, "name", "", "Filter by finding name (supports wildcards * and %)")
	cmd.Flags().StringVar(&cve, "cve", "", "Filter by CVE identifier (supports wildcards)")
	cmd.Flags().StringVar(&severity, "severity", "", "Filter by severity (Critical, High, Medium, Low, Informational)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by justification/status (e.g. Active, Fixed, Accepted Risk)")
	cmd.Flags().StringVar(&scanType, "scan-type", "", "Filter by scan type")
	cmd.Flags().StringVar(&exploitable, "exploitable", "", "Filter by exploitable (1=yes, 0=no)")
	cmd.Flags().IntVar(&start, "start", 0, "Pagination start offset")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (API default: 100, max: 1000)")

	return cmd
}

func newFindingsUpdateCmd() *cobra.Command {
	var (
		status   string
		severity string
		comment  string
		dueDate  string
	)

	cmd := &cobra.Command{
		Use:   "update <finding-number>",
		Short: "Update a finding's status, severity, or comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			projectID, err := resolveProjectID()
			if err != nil {
				return err
			}

			client, err := requireClient()
			if err != nil {
				return err
			}

			input := &service.UpdateFindingInput{
				FindingNumber: args[0],
			}
			if status != "" {
				input.Status = domain.FindingStatus(status)
			}
			if severity != "" {
				input.Severity = domain.Severity(severity)
			}
			if comment != "" {
				input.Comment = comment
			}
			if dueDate != "" {
				input.DueDate = dueDate
			}

			ctx := context.Background()
			if err := client.UpdateFinding(ctx, projectID, input); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Updated finding %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "New status")
	cmd.Flags().StringVar(&severity, "severity", "", "New severity")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD)")

	return cmd
}

func newFindingsBulkUpdateCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "bulk-update",
		Short: "Bulk update findings",
		Long:  "Bulk update multiple findings. Reads JSON array of updates from --file or stdin.",
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

			var reader *os.File
			if file != "" {
				reader, err = os.Open(file) //nolint:gosec // file path is supplied by the user intentionally
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer reader.Close() //nolint:errcheck // best-effort close on input file
			} else {
				reader = os.Stdin
			}

			var updates []service.UpdateFindingInput
			if err := json.NewDecoder(reader).Decode(&updates); err != nil {
				return fmt.Errorf("parsing updates JSON: %w", err)
			}

			ctx := context.Background()
			if err := client.BulkUpdateFindings(ctx, projectID, updates); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Updated %d findings\n", len(updates))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file with updates (reads from stdin if not set)")

	return cmd
}

func newFindingsOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show findings overview/summary",
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
			overview, err := client.GetFindingOverview(ctx, projectID)
			if err != nil {
				return err
			}

			formatter := getFormatter()
			fields := []output.Field{
				{Label: "Critical", Value: overview.Critical},
				{Label: "High", Value: overview.High},
				{Label: "Medium", Value: overview.Medium},
				{Label: "Low", Value: overview.Low},
				{Label: "Crit+High", Value: overview.CritHigh},
				{Label: "Exploitable", Value: overview.ExploitableCount},
				{Label: "CVE Count", Value: overview.CVECount},
				{Label: "IAVA Count", Value: overview.IAVACount},
				{Label: "Vuln Score", Value: fmt.Sprintf("%d", overview.VulnerabilityScore)},
			}

			return formatter.FormatSingle(os.Stdout, fields)
		},
	}
}

func newFindingsMitigatedCmd() *cobra.Command {
	var (
		startDate string
		start     int
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "mitigated",
		Short: "List mitigated findings",
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

			opts := &domain.MitigatedOptions{}
			if startDate != "" {
				opts.StartDate = startDate
			}
			if start > 0 {
				opts.Start = &start
			}
			if limit > 0 {
				opts.Limit = &limit
			}

			ctx := context.Background()
			findings, err := client.GetMitigatedFindings(ctx, projectID, opts)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, f := range findings {
					fmt.Fprintln(os.Stdout, f.FindingNumber)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"Number", "Name", "Severity", "Scan Type", "Remediated Date", "Remediation Days", "Total Mitigated"}
			rows := make([][]string, 0, len(findings))

			for _, f := range findings {
				rows = append(rows, []string{
					f.FindingNumber,
					f.Name,
					f.Severity,
					f.ScanType,
					f.RemediatedDate,
					f.RemediationDays,
					f.TotalMitigated,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&start, "start", 0, "Pagination start offset")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")

	return cmd
}

func newFindingsTrendCmd() *cobra.Command {
	var (
		startDate string
		endDate   string
		groups    string
	)

	cmd := &cobra.Command{
		Use:   "trend",
		Short: "Show finding discovery trend data",
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

			opts := &domain.TrendOptions{}
			if startDate != "" {
				opts.StartDate = startDate
			}
			if endDate != "" {
				opts.EndDate = endDate
			}
			if groups != "" {
				parts := strings.Split(groups, ",")
				expanded, expandErr := expandGroupGlobs(client, projectID, parts)
				if expandErr != nil {
					return expandErr
				}
				opts.AssetGroups = expanded
			}

			ctx := context.Background()
			trend, err := client.GetFindingTrend(ctx, projectID, opts)
			if err != nil {
				return err
			}

			formatter := getFormatter()
			headers := []string{"Date", "Critical", "High", "Medium", "Low", "Informational"}
			rows := make([][]string, 0, len(trend.DataPoints))

			for _, dp := range trend.DataPoints {
				rows = append(rows, []string{
					dp.Date,
					dp.Critical,
					dp.High,
					dp.Medium,
					dp.Low,
					dp.Informational,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&groups, "groups", "", "Asset groups (comma-separated, supports glob patterns e.g. '*team-euc*')")

	return cmd
}

func newFindingsFrameworksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "frameworks",
		Short: "List compliance frameworks",
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
			frameworks, err := client.GetFrameworks(ctx, projectID)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, f := range frameworks {
					fmt.Fprintln(os.Stdout, f)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"Framework"}
			rows := make([][]string, 0, len(frameworks))

			for _, f := range frameworks {
				rows = append(rows, []string{f})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}
}
