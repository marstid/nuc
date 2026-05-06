package mcp

import (
	"fmt"
	"os"

	nucconfig "github.com/marstid/nuc/pkg/config"
)

// Config holds the MCP server configuration resolved from file, flags, and environment.
type Config struct {
	APIKey         string
	BaseURL        string
	DefaultProject string
}

// Resolve builds a Config from the nuc config file, environment variables, and CLI flags,
// in order of increasing priority.
func Resolve(apiKey, baseURL string) (*Config, error) {
	cfg := &Config{}

	fileCfg, err := nucconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("loading nuc config: %w", err)
	}
	cfg.APIKey = fileCfg.APIKey
	cfg.BaseURL = fileCfg.BaseURL
	cfg.DefaultProject = fileCfg.DefaultProject

	if env := os.Getenv("NUC_API_KEY"); env != "" {
		cfg.APIKey = env
	}
	if env := os.Getenv("NUC_BASE_URL"); env != "" {
		cfg.BaseURL = env
	}
	if env := os.Getenv("NUC_PROJECT"); env != "" {
		cfg.DefaultProject = env
	}

	if apiKey != "" {
		cfg.APIKey = apiKey
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key configured: set via 'nuc config set api_key <key>', NUC_API_KEY env var, or --api-key flag")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("no base URL configured: set via 'nuc config set base_url <url>', NUC_BASE_URL env var, or --base-url flag")
	}

	return cfg, nil
}
