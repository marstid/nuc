package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/internal/cli/output"
	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/service"
)

func newAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Manage assets in a project",
		Long:  "List, view, create, update, and delete assets within a Nucleus Security project.",
	}

	cmd.AddCommand(newAssetsListCmd())
	cmd.AddCommand(newAssetsGetCmd())
	cmd.AddCommand(newAssetsCreateCmd())
	cmd.AddCommand(newAssetsUpdateCmd())
	cmd.AddCommand(newAssetsDeleteCmd())
	cmd.AddCommand(newAssetsGroupsCmd())

	return cmd
}

func newAssetsListCmd() *cobra.Command {
	var (
		assetType string
		group     string
		name      string
		ip        string
		start     int
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List assets in a project",
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

			opts := &domain.AssetListOptions{}
			if assetType != "" {
				opts.AssetType = domain.AssetType(assetType)
			}
			if group != "" {
				opts.AssetGroups = group
			}
			if name != "" {
				opts.AssetName = name
			}
			if ip != "" {
				opts.IPAddress = ip
			}
			if start > 0 {
				opts.Start = &start
			}
			if limit > 0 {
				opts.Limit = &limit
			}

			ctx := context.Background()
			assets, err := client.ListAssets(ctx, projectID, opts)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, a := range assets {
					fmt.Fprintln(os.Stdout, a.ID)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"ID", "Name", "Type", "IP", "Groups", "Criticality"}
			rows := make([][]string, 0, len(assets))

			for _, a := range assets {
				rows = append(rows, []string{
					a.ID,
					a.Name,
					string(a.Type),
					a.IPAddress,
					strings.Join(a.Groups, ", "),
					a.Criticality,
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}

	cmd.Flags().StringVar(&assetType, "type", "", "Filter by asset type (Host, Web Application, etc.)")
	cmd.Flags().StringVar(&group, "group", "", "Filter by asset group")
	cmd.Flags().StringVar(&name, "name", "", "Filter by asset name")
	cmd.Flags().StringVar(&ip, "ip", "", "Filter by IP address")
	cmd.Flags().IntVar(&start, "start", 0, "Pagination start offset")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")

	return cmd
}

func newAssetsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <asset-id>",
		Short: "Get asset details",
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

			ctx := context.Background()
			asset, err := client.GetAsset(ctx, projectID, args[0])
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, asset.ID)
				return nil
			}

			formatter := getFormatter()
			fields := []output.Field{
				{Label: "ID", Value: asset.ID},
				{Label: "Name", Value: asset.Name},
				{Label: "Type", Value: string(asset.Type)},
				{Label: "IP Address", Value: asset.IPAddress},
				{Label: "Domain", Value: asset.DomainName},
				{Label: "OS", Value: asset.OperatingSystem},
				{Label: "Criticality", Value: asset.Criticality},
				{Label: "Groups", Value: strings.Join(asset.Groups, ", ")},
				{Label: "Active", Value: strconv.FormatBool(asset.Active)},
			}

			return formatter.FormatSingle(os.Stdout, fields)
		},
	}
}

func newAssetsCreateCmd() *cobra.Command {
	var (
		name        string
		assetType   string
		ip          string
		domainName  string
		osName      string
		group       string
		criticality string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new asset",
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

			input := &service.CreateAssetInput{
				Name: name,
				Type: domain.AssetType(assetType),
			}
			if ip != "" {
				input.IPAddress = ip
			}
			if domainName != "" {
				input.DomainName = domainName
			}
			if osName != "" {
				input.OperatingSystem = osName
			}
			if group != "" {
				input.Groups = strings.Split(group, ",")
			}
			if criticality != "" {
				input.Criticality = criticality
			}

			ctx := context.Background()
			asset, err := client.CreateAsset(ctx, projectID, input)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, asset.ID)
				return nil
			}

			fmt.Fprintf(os.Stdout, "Created asset %s (%s)\n", asset.ID, asset.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Asset name (required)")
	cmd.Flags().StringVar(&assetType, "type", "", "Asset type (required): Host, Web Application, Database, etc.")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address")
	cmd.Flags().StringVar(&domainName, "domain", "", "Domain name")
	cmd.Flags().StringVar(&osName, "os", "", "Operating system")
	cmd.Flags().StringVar(&group, "group", "", "Asset groups (comma-separated)")
	cmd.Flags().StringVar(&criticality, "criticality", "", "Asset criticality")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func newAssetsUpdateCmd() *cobra.Command {
	var (
		name        string
		assetType   string
		ip          string
		domainName  string
		osName      string
		group       string
		criticality string
	)

	cmd := &cobra.Command{
		Use:   "update <asset-id>",
		Short: "Update an existing asset",
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

			input := &service.UpdateAssetInput{}
			if name != "" {
				input.Name = &name
			}
			if assetType != "" {
				input.Type = domain.AssetType(assetType)
			}
			if ip != "" {
				input.IPAddress = &ip
			}
			if domainName != "" {
				input.DomainName = &domainName
			}
			if osName != "" {
				input.OperatingSystem = &osName
			}
			if group != "" {
				groups := strings.Split(group, ",")
				input.Groups = groups
			}
			if criticality != "" {
				input.Criticality = &criticality
			}

			ctx := context.Background()
			asset, err := client.UpdateAsset(ctx, projectID, args[0], input)
			if err != nil {
				return err
			}

			if flags.quiet {
				fmt.Fprintln(os.Stdout, asset.ID)
				return nil
			}

			fmt.Fprintf(os.Stdout, "Updated asset %s (%s)\n", asset.ID, asset.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Asset name")
	cmd.Flags().StringVar(&assetType, "type", "", "Asset type")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address")
	cmd.Flags().StringVar(&domainName, "domain", "", "Domain name")
	cmd.Flags().StringVar(&osName, "os", "", "Operating system")
	cmd.Flags().StringVar(&group, "group", "", "Asset groups (comma-separated)")
	cmd.Flags().StringVar(&criticality, "criticality", "", "Asset criticality")

	return cmd
}

func newAssetsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <asset-id>",
		Short: "Delete an asset",
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

			ctx := context.Background()
			if err := client.DeleteAsset(ctx, projectID, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Deleted asset %s\n", args[0])
			return nil
		},
	}
}

func newAssetsGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage asset groups",
	}

	cmd.AddCommand(newAssetsGroupsListCmd())
	cmd.AddCommand(newAssetsGroupsCreateCmd())
	cmd.AddCommand(newAssetsGroupsDeleteCmd())

	return cmd
}

func newAssetsGroupsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List asset groups",
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
			groups, err := client.ListAssetGroups(ctx, projectID)
			if err != nil {
				return err
			}

			if flags.quiet {
				for _, g := range groups {
					fmt.Fprintln(os.Stdout, g.Name)
				}
				return nil
			}

			formatter := getFormatter()
			headers := []string{"Group", "Asset Count"}
			rows := make([][]string, 0, len(groups))

			for _, g := range groups {
				rows = append(rows, []string{
					g.Name,
					strconv.Itoa(g.AssetCount),
				})
			}

			return formatter.Format(os.Stdout, headers, rows)
		},
	}
}

func newAssetsGroupsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <group-name>",
		Short: "Create an asset group",
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

			ctx := context.Background()
			if err := client.CreateAssetGroup(ctx, projectID, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Created asset group %q\n", args[0])
			return nil
		},
	}
}

func newAssetsGroupsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <group-name>",
		Short: "Delete an asset group",
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

			ctx := context.Background()
			if err := client.DeleteAssetGroup(ctx, projectID, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Deleted asset group %q\n", args[0])
			return nil
		},
	}
}
