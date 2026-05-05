package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/pkg/domain"
)

func newScansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scans",
		Short: "Manage vulnerability scans",
		Long:  "List and upload vulnerability scans for a Nucleus Security project.",
	}

	cmd.AddCommand(newScansListCmd())
	cmd.AddCommand(newScansUploadCmd())

	return cmd
}

func newScansListCmd() *cobra.Command {
	var (
		start int
		limit int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List scans for a project",
		Long:  "List vulnerability scans associated with a project.",
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
			scans, err := client.ListScans(ctx, projectID, start, limit)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, s := range scans {
					fmt.Fprintln(os.Stdout, s.ID)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"ID", "Name", "Type", "Scan Date", "Assets"}
			rows := make([][]string, 0, len(scans))

			for _, s := range scans {
				rows = append(rows, []string{
					s.ID,
					s.Name,
					s.Type,
					s.ScanDate,
					s.AssetCount,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().IntVar(&start, "start", 0, "Pagination start offset")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (API default: 100, max: 100)")

	return cmd
}

func newScansUploadCmd() *cobra.Command {
	var (
		description string
		scanType    string
	)

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a scan file",
		Long:  "Upload a vulnerability scan file to a Nucleus Security project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			filePath := args[0]

			projectID, err := resolveProjectID()
			if err != nil {
				return err
			}

			client, err := requireClient()
			if err != nil {
				return err
			}

			opts := &domain.ScanUploadOptions{
				Description: description,
				Type:        scanType,
			}

			ctx := context.Background()
			result, err := client.UploadScan(ctx, projectID, filePath, opts)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, result.ScanID)
				return nil
			}

			fmt.Fprintf(os.Stdout, "Scan uploaded successfully (ID: %s): %s\n", result.ScanID, result.Message)
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Description for the scan")
	cmd.Flags().StringVar(&scanType, "type", "", "Scan type (e.g. nessus, qualys)")

	return cmd
}
