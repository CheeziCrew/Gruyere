package openapi

import (
	"fmt"
	"sort"
	"strings"
)

// RenderMarkdown generates a markdown changelog from a ChangelogResult.
func RenderMarkdown(r *ChangelogResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# API-Changelog: version %s\n\n", r.FeatureBranch))

	// Endpoints section
	allTags := make(map[string]bool)
	for tag := range r.AddedEndpoints {
		allTags[tag] = true
	}
	for tag := range r.RemovedEndpoints {
		allTags[tag] = true
	}

	if len(allTags) > 0 {
		b.WriteString("## API-endpoints\n\n")

		tags := make([]string, 0, len(allTags))
		for tag := range allTags {
			tags = append(tags, tag)
		}
		sort.Strings(tags)

		for _, tag := range tags {
			b.WriteString(fmt.Sprintf("### %s:\n\n", tag))

			if eps, ok := r.AddedEndpoints[tag]; ok && len(eps) > 0 {
				b.WriteString("#### New endpoints:\n\n```\n")
				for _, ep := range eps {
					b.WriteString(fmt.Sprintf("- [%s] %s\n", ep.Method, ep.Path))
				}
				b.WriteString("```\n\n")
			}

			if eps, ok := r.RemovedEndpoints[tag]; ok && len(eps) > 0 {
				b.WriteString("#### Removed endpoints:\n\n```\n")
				for _, ep := range eps {
					b.WriteString(fmt.Sprintf("- [%s] %s\n", ep.Method, ep.Path))
				}
				b.WriteString("```\n\n")
			}
		}
	}

	// Schema section
	if len(r.SchemaChanges) > 0 {
		b.WriteString("## API-Model updates\n\n")

		for _, sc := range r.SchemaChanges {
			switch sc.Status {
			case "added":
				b.WriteString(fmt.Sprintf("- **%s** *(Added)*\n", sc.Name))
				if len(sc.Fields) > 0 {
					b.WriteString("   - **Fields:**\n")
					for _, f := range sc.Fields {
						b.WriteString(fmt.Sprintf("      - %s: `%s`\n", f.Name, f.Type))
					}
				} else {
					b.WriteString("   - No fields\n")
				}
				b.WriteString("\n")

			case "removed":
				b.WriteString(fmt.Sprintf("- **%s** *(Removed)*\n", sc.Name))
				if len(sc.Fields) > 0 {
					b.WriteString("   - **Fields:**\n")
					for _, f := range sc.Fields {
						b.WriteString(fmt.Sprintf("      - %s: `%s`\n", f.Name, f.Type))
					}
				}
				b.WriteString("\n")

			case "modified":
				b.WriteString(fmt.Sprintf("- **%s**\n", sc.Name))
				if len(sc.AddedFields) > 0 {
					b.WriteString("   - **Added Fields:**\n")
					for _, f := range sc.AddedFields {
						b.WriteString(fmt.Sprintf("      - %s: `%s`\n", f.Name, f.Type))
					}
				}
				if len(sc.RemovedFields) > 0 {
					b.WriteString("   - **Removed Fields:**\n")
					for _, f := range sc.RemovedFields {
						b.WriteString(fmt.Sprintf("      - %s: `%s`\n", f.Name, f.Type))
					}
				}
				if len(sc.ChangedTypeFields) > 0 {
					b.WriteString("   - **Fields with Changed Types:**\n")
					for _, f := range sc.ChangedTypeFields {
						b.WriteString(fmt.Sprintf("      - %s: `%s` -> `%s`\n", f.Name, f.OldType, f.NewType))
					}
				}
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}
