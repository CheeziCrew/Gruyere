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

	diffEndpoints(oldSpec, newSpec, result)
	diffSchemas(oldSpec, newSpec, result)

	result.HasChanges = len(result.AddedEndpoints) > 0 ||
		len(result.RemovedEndpoints) > 0 ||
		len(result.SchemaChanges) > 0

	return result
}

func diffEndpoints(oldSpec, newSpec *Spec, result *ChangelogResult) {
	oldEps := extractEndpoints(oldSpec)
	newEps := extractEndpoints(newSpec)

	for key, tags := range newEps {
		if _, exists := oldEps[key]; !exists {
			addEndpointByTags(result.AddedEndpoints, key, tags)
		}
	}
	for key, tags := range oldEps {
		if _, exists := newEps[key]; !exists {
			addEndpointByTags(result.RemovedEndpoints, key, tags)
		}
	}

	sortEndpointMap(result.AddedEndpoints)
	sortEndpointMap(result.RemovedEndpoints)
}

func addEndpointByTags(m map[string][]Endpoint, key [2]string, tags []string) {
	ep := Endpoint{Path: key[0], Method: key[1], Tags: tags}
	for _, tag := range tags {
		m[tag] = append(m[tag], ep)
	}
}

func sortEndpointMap(m map[string][]Endpoint) {
	for _, eps := range m {
		sort.Slice(eps, func(i, j int) bool {
			if eps[i].Path != eps[j].Path {
				return eps[i].Path < eps[j].Path
			}
			return eps[i].Method < eps[j].Method
		})
	}
}

func diffSchemas(oldSpec, newSpec *Spec, result *ChangelogResult) {
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
		if diff := diffOneSchema(name, oldSchema, newSchema); diff != nil {
			result.SchemaChanges = append(result.SchemaChanges, *diff)
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
}

func diffOneSchema(name string, oldSchema, newSchema Schema) *SchemaDiff {
	oldFields := fieldsToMap(getComponentFields(oldSchema))
	newFields := fieldsToMap(getComponentFields(newSchema))

	added := findAddedFields(newFields, oldFields)
	removed := findAddedFields(oldFields, newFields) // reversed args = removed
	changed := findChangedFields(oldFields, newFields)

	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		return nil
	}

	return &SchemaDiff{
		Name:              name,
		Status:            "modified",
		Fields:            getComponentFields(newSchema),
		AddedFields:       added,
		RemovedFields:     removed,
		ChangedTypeFields: changed,
	}
}

func fieldsToMap(fields []SchemaField) map[string]string {
	m := make(map[string]string)
	for _, f := range fields {
		m[f.Name] = f.Type
	}
	return m
}

func findAddedFields(source, baseline map[string]string) []SchemaField {
	var result []SchemaField
	for name, typ := range source {
		if _, ok := baseline[name]; !ok {
			result = append(result, SchemaField{Name: name, Type: typ})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func findChangedFields(oldFields, newFields map[string]string) []FieldTypeChange {
	var result []FieldTypeChange
	for name, newType := range newFields {
		if oldType, ok := oldFields[name]; ok && oldType != newType {
			result = append(result, FieldTypeChange{Name: name, OldType: oldType, NewType: newType})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
