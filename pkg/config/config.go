// Package config handles configuration loading, saving, and access for the nuc CLI.
// Configuration is stored as YAML in an XDG-compliant location.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// appName is the directory name for the config file.
	appName = "nuc"

	// configFileName is the name of the config file.
	configFileName = "config.yaml"

	// DefaultBaseURL is empty — Nucleus instances are per-org (e.g. nucleus-eu6.nucleussec.com).
	// Users must set base_url via config, --base-url flag, or NUC_BASE_URL env var.
	DefaultBaseURL = ""

	// DefaultOutputFormat is the default output format.
	DefaultOutputFormat = "table"
)

// validKeys is the allowlist of valid configuration keys.
var validKeys = []string{"api_key", "base_url", "default_project", "output_format"}

// Config represents the persisted CLI configuration.
type Config struct {
	APIKey         string `yaml:"api_key,omitempty"`
	BaseURL        string `yaml:"base_url,omitempty"`
	DefaultProject string `yaml:"default_project,omitempty"`
	OutputFormat   string `yaml:"output_format,omitempty"`
}

// Dir returns the configuration directory path.
// Uses os.UserConfigDir() which respects XDG_CONFIG_HOME on Linux,
// ~/Library/Application Support on macOS, and %AppData% on Windows.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determining config directory: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// Path returns the full path to the configuration file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads the configuration from disk.
// Returns a zero-value Config (not an error) if the file does not exist.
// Returns an error if the file exists but contains invalid YAML.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is derived from os.UserConfigDir(), not user input
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// LoadFrom reads the configuration from a specific path.
// Returns a zero-value Config (not an error) if the file does not exist.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is supplied by the caller (e.g., --config flag)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk, creating the directory if needed.
// Directory permissions: 0700. File permissions: 0600.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(cfg, path)
}

// SaveTo writes the configuration to a specific path.
func SaveTo(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Get returns the value for a configuration key.
// Returns an empty string for unknown keys.
func (c *Config) Get(key string) string {
	switch key {
	case "api_key":
		return c.APIKey
	case "base_url":
		return c.BaseURL
	case "default_project":
		return c.DefaultProject
	case "output_format":
		return c.OutputFormat
	default:
		return ""
	}
}

// Set sets a configuration value by key.
// Returns an error if the key is not in the valid allowlist.
func (c *Config) Set(key, value string) error {
	switch key {
	case "api_key":
		c.APIKey = value
	case "base_url":
		c.BaseURL = value
	case "default_project":
		c.DefaultProject = value
	case "output_format":
		c.OutputFormat = value
	default:
		return fmt.Errorf("unknown config key %q (valid keys: %v)", key, validKeys)
	}
	return nil
}

// Keys returns all valid configuration key names.
func Keys() []string {
	result := make([]string, len(validKeys))
	copy(result, validKeys)
	return result
}

// BaseURLOrDefault returns the configured base URL or the default if not set.
func (c *Config) BaseURLOrDefault() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return DefaultBaseURL
}

// OutputFormatOrDefault returns the configured output format or the default if not set.
func (c *Config) OutputFormatOrDefault() string {
	if c.OutputFormat != "" {
		return c.OutputFormat
	}
	return DefaultOutputFormat
}
