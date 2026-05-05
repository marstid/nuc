package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/internal/cli/output"
	"github.com/marstid/nuc/pkg/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage nuc configuration",
		Long:  "View and modify the nuc CLI configuration file.",
	}

	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: fmt.Sprintf("Set a configuration value. Valid keys: %s",
			strings.Join(config.Keys(), ", ")),
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := cfg.Set(key, value); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Set %s successfully.\n", key)
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: fmt.Sprintf("Get a configuration value. Valid keys: %s",
			strings.Join(config.Keys(), ", ")),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			value := cfg.Get(key)
			if value == "" {
				fmt.Fprintf(os.Stdout, "%s: (not set)\n", key)
				return nil
			}

			// Mask the API key for security.
			if key == "api_key" {
				value = maskAPIKey(value)
			}

			fmt.Fprintf(os.Stdout, "%s: %s\n", key, value)
			return nil
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			formatter := output.New(string(output.FormatTable))
			var fields []output.Field

			for _, key := range config.Keys() {
				value := cfg.Get(key)
				if value == "" {
					value = "(not set)"
				} else if key == "api_key" {
					value = maskAPIKey(value)
				}
				fields = append(fields, output.Field{
					Label: key,
					Value: value,
				})
			}

			return formatter.FormatSingle(os.Stdout, fields)
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the configuration file path",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, path)
			return nil
		},
	}
}

// maskAPIKey shows only the last 4 characters of an API key for security.
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}
