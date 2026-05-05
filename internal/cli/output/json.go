package output

import (
	"encoding/json"
	"io"
)

// jsonFormatter renders data as machine-readable JSON.
type jsonFormatter struct{}

// Format renders tabular data as a JSON array of objects.
// Each row becomes an object with headers as keys.
func (f *jsonFormatter) Format(w io.Writer, headers []string, rows [][]string) error {
	objects := make([]map[string]string, 0, len(rows))

	for _, row := range rows {
		obj := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			}
		}
		objects = append(objects, obj)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

// FormatSingle renders a single object as a JSON object with field labels as keys.
func (f *jsonFormatter) FormatSingle(w io.Writer, fields []Field) error {
	obj := make(map[string]string, len(fields))
	for _, field := range fields {
		obj[field.Label] = field.Value
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}
