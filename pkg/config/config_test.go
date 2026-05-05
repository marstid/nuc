package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "config.yaml")

	cfg, err := LoadFrom(path)

	require.NoError(t, err)
	assert.Equal(t, &Config{}, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(path, []byte("{{invalid yaml"), 0o600)
	require.NoError(t, err)

	_, err = LoadFrom(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nuc", "config.yaml")

	original := &Config{
		APIKey:         "sk-test-key-12345",
		BaseURL:        "https://custom.nucleussec.com/api",
		DefaultProject: "42",
		OutputFormat:   "json",
	}

	err := SaveTo(original, path)
	require.NoError(t, err)

	loaded, err := LoadFrom(path)
	require.NoError(t, err)

	assert.Equal(t, original.APIKey, loaded.APIKey)
	assert.Equal(t, original.BaseURL, loaded.BaseURL)
	assert.Equal(t, original.DefaultProject, loaded.DefaultProject)
	assert.Equal(t, original.OutputFormat, loaded.OutputFormat)
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "nested", "dir")
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{APIKey: "test"}
	err := SaveTo(cfg, path)

	require.NoError(t, err)

	// Verify file was created.
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Verify directory permissions.
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}

func TestSave_FilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg := &Config{APIKey: "secret"}
	err := SaveTo(cfg, path)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestGet_AllKeys(t *testing.T) {
	cfg := &Config{
		APIKey:         "my-key",
		BaseURL:        "https://example.com",
		DefaultProject: "99",
		OutputFormat:   "yaml",
	}

	assert.Equal(t, "my-key", cfg.Get("api_key"))
	assert.Equal(t, "https://example.com", cfg.Get("base_url"))
	assert.Equal(t, "99", cfg.Get("default_project"))
	assert.Equal(t, "yaml", cfg.Get("output_format"))
}

func TestGet_UnknownKey(t *testing.T) {
	cfg := &Config{APIKey: "test"}

	assert.Equal(t, "", cfg.Get("unknown_key"))
}

func TestSet_ValidKeys(t *testing.T) {
	cfg := &Config{}

	require.NoError(t, cfg.Set("api_key", "new-key"))
	require.NoError(t, cfg.Set("base_url", "https://new.example.com"))
	require.NoError(t, cfg.Set("default_project", "123"))
	require.NoError(t, cfg.Set("output_format", "json"))

	assert.Equal(t, "new-key", cfg.APIKey)
	assert.Equal(t, "https://new.example.com", cfg.BaseURL)
	assert.Equal(t, "123", cfg.DefaultProject)
	assert.Equal(t, "json", cfg.OutputFormat)
}

func TestSet_InvalidKey(t *testing.T) {
	cfg := &Config{}

	err := cfg.Set("invalid_key", "value")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
	assert.Contains(t, err.Error(), "invalid_key")
}

func TestKeys(t *testing.T) {
	keys := Keys()

	assert.Contains(t, keys, "api_key")
	assert.Contains(t, keys, "base_url")
	assert.Contains(t, keys, "default_project")
	assert.Contains(t, keys, "output_format")
	assert.Len(t, keys, 4)
}

func TestBaseURLOrDefault(t *testing.T) {
	t.Run("returns configured URL", func(t *testing.T) {
		cfg := &Config{BaseURL: "https://custom.example.com"}
		assert.Equal(t, "https://custom.example.com", cfg.BaseURLOrDefault())
	})

	t.Run("returns default when empty", func(t *testing.T) {
		cfg := &Config{}
		assert.Equal(t, DefaultBaseURL, cfg.BaseURLOrDefault())
	})
}

func TestOutputFormatOrDefault(t *testing.T) {
	t.Run("returns configured format", func(t *testing.T) {
		cfg := &Config{OutputFormat: "json"}
		assert.Equal(t, "json", cfg.OutputFormatOrDefault())
	})

	t.Run("returns default when empty", func(t *testing.T) {
		cfg := &Config{}
		assert.Equal(t, DefaultOutputFormat, cfg.OutputFormatOrDefault())
	})
}
