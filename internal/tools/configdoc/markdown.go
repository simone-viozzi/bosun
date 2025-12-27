package configdoc

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// FieldRow represents a single row in the configuration table.
type FieldRow struct {
	Key         string
	Type        string
	Default     string
	Required    string
	EnumValues  string
	Description string
	Deprecated  bool
}

// ScopeSection represents a section of the Markdown document for one scope.
type ScopeSection struct {
	Scope       string
	Description string
	Fields      []FieldRow
}

// FormatDoc represents documentation for special value formats.
type FormatDoc struct {
	Name        string
	Description string
	Examples    []string
}

// MarkdownDoc is the internal representation for Markdown generation.
type MarkdownDoc struct {
	Title      string
	Sections   []ScopeSection
	FormatDocs []FormatDoc
}

// markdownTemplate is the template for generating Markdown documentation.
const markdownTemplate = `# {{.Title}}

> Auto-generated from ConfigV1 schema. Do not edit manually.

{{range .Sections}}{{if .Fields}}## {{.Scope}} Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
{{range .Fields}}| ` + "`{{.Key}}`" + ` | {{.Type}} | {{.Default}} | {{.Required}} | {{.EnumValues}} | {{if .Deprecated}}~~{{.Description}}~~ *(deprecated)*{{else}}{{.Description}}{{end}} |
{{end}}
{{end}}{{end}}## Value Formats

{{range .FormatDocs}}### {{.Name}}

{{.Description}}

**Examples:** {{range $i, $e := .Examples}}{{if $i}}, {{end}}` + "`{{$e}}`" + `{{end}}

{{end}}`

// fieldSpecToRow converts a FieldSpec to a FieldRow for Markdown rendering.
func fieldSpecToRow(f schema.FieldSpec) FieldRow {
	defaultVal := f.Default
	if defaultVal == "" {
		defaultVal = "-"
	}

	required := "No"
	if f.Required {
		required = "Yes"
	}

	enumValues := "-"
	if len(f.Enum) > 0 {
		enumValues = strings.Join(f.Enum, " \\| ")
	}

	return FieldRow{
		Key:         f.Key,
		Type:        typeDisplayName(f.Type),
		Default:     defaultVal,
		Required:    required,
		EnumValues:  enumValues,
		Description: f.Doc,
		Deprecated:  f.Deprecated,
	}
}

// generateMarkdown produces Markdown documentation from the spec.
func generateMarkdown(spec schema.Spec, title string) ([]byte, error) {
	// Group fields by scope
	scopeFields := spec.Scopes()

	// Build sections in canonical order
	sections := make([]ScopeSection, 0, len(scopeOrder))
	for _, scope := range scopeOrder {
		fields, ok := scopeFields[scope]
		if !ok || len(fields) == 0 {
			continue
		}

		// Sort fields by key within each scope
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Key < fields[j].Key
		})

		var rows []FieldRow
		for _, f := range fields {
			rows = append(rows, fieldSpecToRow(f))
		}

		sections = append(sections, ScopeSection{
			Scope:  scopeDisplayName(scope),
			Fields: rows,
		})
	}

	// Build format documentation
	formatDocs := []FormatDoc{
		{
			Name:        "Duration",
			Description: "Go duration syntax. Supported units: `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`.",
			Examples:    []string{"30s", "5m", "1h30m", "100ms"},
		},
		{
			Name:        "Byte Size",
			Description: "Docker/go-units byte size syntax. Supports base-10 (KB, MB, GB, TB) and base-2 (KiB, MiB, GiB, TiB) units.",
			Examples:    []string{"100MB", "1GiB", "500KB", "10GB"},
		},
		{
			Name:        "List",
			Description: "List values can be specified as CSV (comma-separated) or JSON array syntax.",
			Examples:    []string{"value1,value2,value3", `["value1", "value2"]`},
		},
	}

	doc := MarkdownDoc{
		Title:      title,
		Sections:   sections,
		FormatDocs: formatDocs,
	}

	// Execute template
	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		return nil, fmt.Errorf("configdoc: failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, doc); err != nil {
		return nil, fmt.Errorf("configdoc: failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
