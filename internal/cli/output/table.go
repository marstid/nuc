package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// tableFormatter renders data as human-readable ASCII tables.
type tableFormatter struct{}

// Format renders tabular data as an ASCII table.
func (f *tableFormatter) Format(w io.Writer, headers []string, rows [][]string) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
	return nil
}

// FormatSingle renders a single object as a key-value table.
func (f *tableFormatter) FormatSingle(w io.Writer, fields []Field) error {
	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator(":")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	for _, field := range fields {
		table.Append([]string{field.Label, field.Value})
	}

	table.Render()
	return nil
}
