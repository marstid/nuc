// Package output provides output formatting for the nuc CLI.
// It supports multiple output formats (table, JSON) with a common interface.
package output

import "io"

// Formatter renders structured data to a writer.
type Formatter interface {
	// Format renders tabular data with headers and rows.
	Format(w io.Writer, headers []string, rows [][]string) error

	// FormatSingle renders a single object as key-value pairs.
	FormatSingle(w io.Writer, fields []Field) error
}

// Field represents a labeled value for single-object display.
type Field struct {
	Label string
	Value string
}

// Format is the output format type.
type Format string

const (
	// FormatTable renders human-readable tables (default for TTY).
	FormatTable Format = "table"

	// FormatJSON renders machine-readable JSON (default for pipes).
	FormatJSON Format = "json"

	// FormatYAML renders YAML output.
	FormatYAML Format = "yaml"
)

// New creates a Formatter for the given format string.
// Returns a table formatter for unknown/empty format strings.
func New(format string) Formatter {
	switch Format(format) {
	case FormatJSON:
		return &jsonFormatter{}
	case FormatYAML:
		return &yamlFormatter{}
	default:
		return &tableFormatter{}
	}
}
