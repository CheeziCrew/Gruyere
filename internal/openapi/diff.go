package openapi

import (
	"sort"
	"strings"
)

var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// extractEndpoints pulls all endpoints from a spec, keyed by (path, method).
func extractEndpoints(spec *Spec) map[[2]string][]string {
	endpoints := make(map[[2]string][]string)
	for path, pathItem := range spec.Paths {
		for method, op := range pathItem {
			if !httpMethods[strings.ToLower(method)] {
				continue
			}
			tags := op.Tags
			if len(tags) == 0 {
				tags = []string{"No Tag"}
			}
			key := [2]string{path, strings.ToUpper(method)}
			endpoints[key] = tags
		}
	}
	return endpoints
}

// getFieldType returns the effective type string for a property.
func getFieldType(p Property) string {
	if p.Ref != "" {
		return "$ref"
	}
	if p.Type == "" {
		return "object"
	}
	return p.Type
}

// getComponentFields extracts sorted fields from a schema.
func getComponentFields(s Schema) []SchemaField {
	var fields []SchemaField
	for name, prop := range s.Properties {
		fields = append(fields, SchemaField{Name: name, Type: getFieldType(prop)})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

// Diff computes the full changelog between two OpenAPI specs.
func Diff(oldSpec, newSpec *Spec, baseBranch, featureBranch, openAPIPath string) *ChangelogResult {
	result := &ChangelogResult{
		BaseBranch:       baseBranch,
		FeatureBranch:    featureBranch,
		OpenAPIPath:      openAPIPath,
		AddedEndpoints:   make(map[string][]Endpoint),
		RemovedEndpoints: make(map[string][]Endpoint),
	}

	// Diff endpoints
	oldEps := extractEndpoints(oldSpec)
	newEps := extractEndpoints(newSpec)

	for key, tags := range newEps {
		if _, exists := oldEps[key]; !exists {
			ep := Endpoint{Path: key[0], Method: key[1], Tags: tags}
			for _, tag := range tags {
				result.AddedEndpoints[tag] = append(result.AddedEndpoints[tag], ep)
			}
		}
	}
	for key, tags := range oldEps {
		if _, exists := newEps[key]; !exists {
			ep := Endpoint{Path: key[0], Method: key[1], Tags: tags}
			for _, tag := range tags {
				result.RemovedEndpoints[tag] = append(result.RemovedEndpoints[tag], ep)
			}
		}
	}

	// Sort endpoints within each tag
	sortEps := func(m map[string][]Endpoint) {
		for _, eps := range m {
			sort.Slice(eps, func(i, j int) bool {
				if eps[i].Path != eps[j].Path {
					return eps[i].Path < eps[j].Path
				}
				return eps[i].Method < eps[j].Method
			})
		}
	}
	sortEps(result.AddedEndpoints)
	sortEps(result.RemovedEndpoints)

	// Diff schemas
	oldSchemas := oldSpec.Components.Schemas
	newSchemas := newSpec.Components.Schemas

	for name, newSchema := range newSchemas {
		oldSchema, exists := oldSchemas[name]
		if !exists {
			result.SchemaChanges = append(result.SchemaChanges, SchemaDiff{
				Name:   name,
				Status: "added",
				Fields: getComponentFields(newSchema),
			})
			continue
		}

		oldFields := make(map[string]string)
		for _, f := range getComponentFields(oldSchema) {
			oldFields[f.Name] = f.Type
		}
		newFields := make(map[string]string)
		for _, f := range getComponentFields(newSchema) {
			newFields[f.Name] = f.Type
		}

		var added, removed []SchemaField
		var changed []FieldTypeChange

		for name, typ := range newFields {
			if _, ok := oldFields[name]; !ok {
				added = append(added, SchemaField{Name: name, Type: typ})
			}
		}
		for name, typ := range oldFields {
			if _, ok := newFields[name]; !ok {
				removed = append(removed, SchemaField{Name: name, Type: typ})
			}
		}
		for name, newType := range newFields {
			if oldType, ok := oldFields[name]; ok && oldType != newType {
				changed = append(changed, FieldTypeChange{Name: name, OldType: oldType, NewType: newType})
			}
		}

		sort.Slice(added, func(i, j int) bool { return added[i].Name < added[j].Name })
		sort.Slice(removed, func(i, j int) bool { return removed[i].Name < removed[j].Name })
		sort.Slice(changed, func(i, j int) bool { return changed[i].Name < changed[j].Name })

		if len(added) > 0 || len(removed) > 0 || len(changed) > 0 {
			result.SchemaChanges = append(result.SchemaChanges, SchemaDiff{
				Name:              name,
				Status:            "modified",
				Fields:            getComponentFields(newSchema),
				AddedFields:       added,
				RemovedFields:     removed,
				ChangedTypeFields: changed,
			})
		}
	}

	for name, oldSchema := range oldSchemas {
		if _, exists := newSchemas[name]; !exists {
			result.SchemaChanges = append(result.SchemaChanges, SchemaDiff{
				Name:   name,
				Status: "removed",
				Fields: getComponentFields(oldSchema),
			})
		}
	}

	sort.Slice(result.SchemaChanges, func(i, j int) bool {
		return result.SchemaChanges[i].Name < result.SchemaChanges[j].Name
	})

	result.HasChanges = len(result.AddedEndpoints) > 0 ||
		len(result.RemovedEndpoints) > 0 ||
		len(result.SchemaChanges) > 0

	return result
}
