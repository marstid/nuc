// Package cli implements the nuc command-line interface using cobra.
// It consumes service interfaces and formats output for terminal display.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/marstid/nuc/internal/cli/output"
	"github.com/marstid/nuc/pkg/config"
	"github.com/marstid/nuc/pkg/domain"
	"github.com/marstid/nuc/pkg/nucleus"
)

// Exit codes for the CLI.
const (
	ExitSuccess  = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitAuth     = 3
	ExitNotFound = 4
)

// globalFlags holds the CLI's persistent flag values.
type globalFlags struct {
	apiKey  string
	baseURL string
	project string
	output  string
	quiet   bool
}

var flags globalFlags

// NewRootCmd creates and returns the root cobra command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "nuc",
		Short: "Nucleus Security CLI",
		Long:  "A command-line interface for the Nucleus Security vulnerability management platform.",
		// Silence usage on errors — we handle it ourselves.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Persistent flags available to all subcommands.
	root.PersistentFlags().StringVar(&flags.apiKey, "api-key", "", "Nucleus API key (env: NUC_API_KEY)")
	root.PersistentFlags().StringVar(&flags.baseURL, "base-url", "", "Nucleus API base URL (env: NUC_BASE_URL)")
	root.PersistentFlags().StringVarP(&flags.project, "project", "p", "", "Default project ID (env: NUC_PROJECT)")
	root.PersistentFlags().StringVarP(&flags.output, "output", "o", "", "Output format: table, json (default: table for TTY, json for pipe)")
	root.PersistentFlags().BoolVarP(&flags.quiet, "quiet", "q", false, "Only print IDs (for scripting)")

	// Register subcommands.
	root.AddCommand(newVersionCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newProjectsCmd())
	root.AddCommand(newAssetsCmd())
	root.AddCommand(newFindingsCmd())
	root.AddCommand(newScansCmd())
	root.AddCommand(newMetricsCmd())

	return root
}

// Execute runs the root command and exits with the appropriate code.
func Execute() {
	cmd := NewRootCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		code := exitCodeForError(err)
		os.Exit(code)
	}
}

// requireClient resolves configuration and creates a Nucleus API client.
// The returned client satisfies all service interfaces (ProjectService, AssetService, etc.).
// Call this from commands that require API access.
func requireClient() (*nucleus.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	apiKey := resolveAPIKey(cfg)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: no API key configured. Set via --api-key flag, NUC_API_KEY env, or 'nuc config set api_key <key>'", domain.ErrUnauthorized)
	}

	baseURL := resolveBaseURL(cfg)
	if baseURL == "" {
		return nil, fmt.Errorf(
			"no base URL configured. Set via --base-url flag, NUC_BASE_URL env, or 'nuc config set base_url <url>'\n" +
				"  Example: nuc config set base_url https://nucleus-eu6.nucleussec.com/nucleus/api",
		)
	}

	client := nucleus.NewClient(baseURL, apiKey)
	return client, nil
}

// getFormatter returns the appropriate output formatter based on flags and config.
func getFormatter() output.Formatter {
	format := flags.output

	if format == "" {
		// Load from config if not specified by flag.
		cfg, err := config.Load()
		if err == nil && cfg.OutputFormat != "" {
			format = cfg.OutputFormat
		}
	}

	if format == "" {
		// Auto-detect: table for TTY, json for pipe.
		if isTerminal() {
			format = string(output.FormatTable)
		} else {
			format = string(output.FormatJSON)
		}
	}

	return output.New(format)
}

// resolveAPIKey returns the API key using priority: flag > env > config.
func resolveAPIKey(cfg *config.Config) string {
	if flags.apiKey != "" {
		return flags.apiKey
	}
	if env := os.Getenv("NUC_API_KEY"); env != "" {
		return env
	}
	return cfg.APIKey
}

// resolveBaseURL returns the base URL using priority: flag > env > config > default.
func resolveBaseURL(cfg *config.Config) string {
	if flags.baseURL != "" {
		return flags.baseURL
	}
	if env := os.Getenv("NUC_BASE_URL"); env != "" {
		return env
	}
	return cfg.BaseURLOrDefault()
}

// resolveProjectID returns the project ID using priority: flag > env > config.
func resolveProjectID() (string, error) {
	if flags.project != "" {
		return flags.project, nil
	}
	if env := os.Getenv("NUC_PROJECT"); env != "" {
		return env, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}
	if cfg.DefaultProject != "" {
		return cfg.DefaultProject, nil
	}

	return "", fmt.Errorf("no project specified. Use --project flag, NUC_PROJECT env, or 'nuc config set default_project <id>'")
}

// exitCodeForError maps domain errors to CLI exit codes.
func exitCodeForError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return ExitAuth
	case errors.Is(err, domain.ErrForbidden):
		return ExitAuth
	case errors.Is(err, domain.ErrNotFound):
		return ExitNotFound
	default:
		return ExitError
	}
}

// isTerminal returns true if stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
