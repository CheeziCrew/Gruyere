package openapi

import (
	"fmt"
	"sort"
	"strings"
)

// fieldLineFmt is the format string for rendering a schema field line.
const fieldLineFmt = "      - %s: `%s`\n"

// RenderMarkdown generates a markdown changelog from a ChangelogResult.
func RenderMarkdown(r *ChangelogResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# API-Changelog: version %s\n\n", r.FeatureBranch))
	renderEndpointSection(&b, r)
	renderSchemaSection(&b, r)

	return b.String()
}

func renderEndpointSection(b *strings.Builder, r *ChangelogResult) {
	allTags := collectTags(r)
	if len(allTags) == 0 {
		return
	}

	b.WriteString("## API-endpoints\n\n")

	tags := make([]string, 0, len(allTags))
	for tag := range allTags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		b.WriteString(fmt.Sprintf("### %s:\n\n", tag))
		renderEndpointList(b, "New", r.AddedEndpoints[tag])
		renderEndpointList(b, "Removed", r.RemovedEndpoints[tag])
	}
}

func collectTags(r *ChangelogResult) map[string]bool {
	allTags := make(map[string]bool)
	for tag := range r.AddedEndpoints {
		allTags[tag] = true
	}
	for tag := range r.RemovedEndpoints {
		allTags[tag] = true
	}
	return allTags
}

func renderEndpointList(b *strings.Builder, label string, eps []Endpoint) {
	if len(eps) == 0 {
		return
	}
	b.WriteString(fmt.Sprintf("#### %s endpoints:\n\n```\n", label))
	for _, ep := range eps {
		b.WriteString(fmt.Sprintf("- [%s] %s\n", ep.Method, ep.Path))
	}
	b.WriteString("```\n\n")
}

func renderSchemaSection(b *strings.Builder, r *ChangelogResult) {
	if len(r.SchemaChanges) == 0 {
		return
	}

	b.WriteString("## API-Model updates\n\n")
	for _, sc := range r.SchemaChanges {
		renderSchemaDiff(b, sc)
	}
}

func renderSchemaDiff(b *strings.Builder, sc SchemaDiff) {
	switch sc.Status {
	case "added":
		renderAddedSchema(b, sc)
	case "removed":
		renderRemovedSchema(b, sc)
	case "modified":
		renderModifiedSchema(b, sc)
	}
}

func renderAddedSchema(b *strings.Builder, sc SchemaDiff) {
	b.WriteString(fmt.Sprintf("- **%s** *(Added)*\n", sc.Name))
	if len(sc.Fields) > 0 {
		b.WriteString("   - **Fields:**\n")
		writeSchemaFields(b, sc.Fields)
	} else {
		b.WriteString("   - No fields\n")
	}
	b.WriteString("\n")
}

func renderRemovedSchema(b *strings.Builder, sc SchemaDiff) {
	b.WriteString(fmt.Sprintf("- **%s** *(Removed)*\n", sc.Name))
	if len(sc.Fields) > 0 {
		b.WriteString("   - **Fields:**\n")
		writeSchemaFields(b, sc.Fields)
	}
	b.WriteString("\n")
}

func renderModifiedSchema(b *strings.Builder, sc SchemaDiff) {
	b.WriteString(fmt.Sprintf("- **%s**\n", sc.Name))
	writeFieldSection(b, "Added Fields", sc.AddedFields)
	writeFieldSection(b, "Removed Fields", sc.RemovedFields)
	if len(sc.ChangedTypeFields) > 0 {
		b.WriteString("   - **Fields with Changed Types:**\n")
		for _, f := range sc.ChangedTypeFields {
			b.WriteString(fmt.Sprintf("      - %s: `%s` -> `%s`\n", f.Name, f.OldType, f.NewType))
		}
	}
	b.WriteString("\n")
}

func writeFieldSection(b *strings.Builder, label string, fields []SchemaField) {
	if len(fields) == 0 {
		return
	}
	b.WriteString(fmt.Sprintf("   - **%s:**\n", label))
	writeSchemaFields(b, fields)
}

func writeSchemaFields(b *strings.Builder, fields []SchemaField) {
	for _, f := range fields {
		b.WriteString(fmt.Sprintf(fieldLineFmt, f.Name, f.Type))
	}
}
