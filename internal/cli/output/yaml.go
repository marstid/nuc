package output

import (
	"io"

	"gopkg.in/yaml.v3"
)

// yamlFormatter renders data as YAML.
type yamlFormatter struct{}

// Format renders tabular data as a YAML sequence of mappings.
// Each row becomes a mapping with headers as keys.
func (f *yamlFormatter) Format(w io.Writer, headers []string, rows [][]string) error {
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

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(objects)
}

// FormatSingle renders a single object as YAML key-value pairs.
func (f *yamlFormatter) FormatSingle(w io.Writer, fields []Field) error {
	obj := make(map[string]string, len(fields))
	for _, field := range fields {
		obj[field.Label] = field.Value
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(obj)
}
