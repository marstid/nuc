package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/internal/cli/output"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage Nucleus Security projects",
		Long:  "List, view, and inspect Nucleus Security projects.",
	}

	cmd.AddCommand(newProjectsListCmd())
	cmd.AddCommand(newProjectsGetCmd())
	cmd.AddCommand(newProjectsRiskScoreCmd())

	return cmd
}

func newProjectsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long:  "List all projects accessible to the authenticated user.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			client, err := requireClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			projects, err := client.List(ctx)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, p := range projects {
					fmt.Fprintln(os.Stdout, p.ID)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"ID", "Name", "Description", "Updated"}
			rows := make([][]string, 0, len(projects))

			for _, p := range projects {
				updated := ""
				if !p.UpdatedAt.IsZero() {
					updated = p.UpdatedAt.Format("2006-01-02")
				}
				rows = append(rows, []string{
					p.ID,
					p.Name,
					p.Description,
					updated,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}
}

func newProjectsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <project-id>",
		Short: "Get project details",
		Long:  "Show detailed information about a specific project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			projectID := args[0]

			client, err := requireClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			project, err := client.Get(ctx, projectID)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, project.ID)
				return nil
			}

			formatter := getFormatter()
			fields := []output.Field{
				{Label: "ID", Value: project.ID},
				{Label: "Name", Value: project.Name},
				{Label: "Description", Value: project.Description},
			}
			if !project.CreatedAt.IsZero() {
				fields = append(fields, output.Field{Label: "Created", Value: project.CreatedAt.Format("2006-01-02 15:04:05")})
			}
			if !project.UpdatedAt.IsZero() {
				fields = append(fields, output.Field{Label: "Updated", Value: project.UpdatedAt.Format("2006-01-02 15:04:05")})
			}

			return formatter.FormatSingle(os.Stdout, fields)
		},
	}
}

func newProjectsRiskScoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "riskscore [project-id]",
		Short: "Get project risk score",
		Long:  "Show the risk score for a project. Uses the default project if no ID is provided.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var projectID string
			if len(args) > 0 {
				projectID = args[0]
			} else {
				var err error
				projectID, err = resolveProjectID()
				if err != nil {
					return err
				}
			}

			client, err := requireClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			riskScore, err := client.GetRiskScore(ctx, projectID)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, riskScore.Score)
				return nil
			}

			formatter := getFormatter()
			fields := []output.Field{
				{Label: "Project ID", Value: projectID},
				{Label: "Risk Score", Value: strconv.Itoa(riskScore.Score)},
			}

			return formatter.FormatSingle(os.Stdout, fields)
		},
	}
}
