package openapi

// Spec is a minimal OpenAPI structure — only the parts we need for diffing.
type Spec struct {
	Paths      map[string]PathItem `yaml:"paths"`
	Components struct {
		Schemas map[string]Schema `yaml:"schemas"`
	} `yaml:"components"`
}

// PathItem maps HTTP methods to operations.
type PathItem map[string]Operation

// Operation holds the tags for an endpoint.
type Operation struct {
	Tags []string `yaml:"tags"`
}

// Schema holds the properties of a component.
type Schema struct {
	Properties map[string]Property `yaml:"properties"`
}

// Property holds the type info for a schema field.
type Property struct {
	Type string `yaml:"type"`
	Ref  string `yaml:"$ref"`
}

// Endpoint represents a single API endpoint.
type Endpoint struct {
	Path   string
	Method string
	Tags   []string
}

// SchemaField represents a single property in a component schema.
type SchemaField struct {
	Name string
	Type string
}

// FieldTypeChange tracks a field whose type changed between versions.
type FieldTypeChange struct {
	Name    string
	OldType string
	NewType string
}

// SchemaDiff captures changes to a single component schema.
type SchemaDiff struct {
	Name              string
	Status            string // "added", "removed", "modified"
	Fields            []SchemaField
	AddedFields       []SchemaField
	RemovedFields     []SchemaField
	ChangedTypeFields []FieldTypeChange
}

// ChangelogResult holds the full diff between two OpenAPI specs.
type ChangelogResult struct {
	BaseBranch    string
	FeatureBranch string
	OpenAPIPath   string

	AddedEndpoints   map[string][]Endpoint // tag -> endpoints
	RemovedEndpoints map[string][]Endpoint

	SchemaChanges []SchemaDiff

	HasChanges bool
}
