package openapi

import "testing"

func TestParseSpec(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantErr       bool
		checkPaths    int
		checkSchemas  int
		pathsNotNil   bool
		schemasNotNil bool
	}{
		{
			name: "valid YAML with paths and components",
			input: `
paths:
  /users:
    get:
      tags: [Users]
    post:
      tags: [Users]
  /items:
    get:
      tags: [Items]
components:
  schemas:
    User:
      properties:
        name:
          type: string
        age:
          type: integer
`,
			checkPaths:    2,
			checkSchemas:  1,
			pathsNotNil:   true,
			schemasNotNil: true,
		},
		{
			name:          "empty YAML initializes maps",
			input:         "",
			checkPaths:    0,
			checkSchemas:  0,
			pathsNotNil:   true,
			schemasNotNil: true,
		},
		{
			name:    "invalid YAML returns error",
			input:   ":\n  :\n    - ][",
			wantErr: true,
		},
		{
			name: "paths but no components initializes schemas map",
			input: `
paths:
  /health:
    get:
      tags: [Health]
`,
			checkPaths:    1,
			checkSchemas:  0,
			pathsNotNil:   true,
			schemasNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseSpec([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.pathsNotNil && spec.Paths == nil {
				t.Error("expected Paths to be initialized, got nil")
			}
			if tt.schemasNotNil && spec.Components.Schemas == nil {
				t.Error("expected Components.Schemas to be initialized, got nil")
			}
			if got := len(spec.Paths); got != tt.checkPaths {
				t.Errorf("Paths count = %d, want %d", got, tt.checkPaths)
			}
			if got := len(spec.Components.Schemas); got != tt.checkSchemas {
				t.Errorf("Schemas count = %d, want %d", got, tt.checkSchemas)
			}
		})
	}
}

func TestParseSpec_FieldValues(t *testing.T) {
	input := `
paths:
  /pets:
    get:
      tags: [Pets, Animals]
components:
  schemas:
    Pet:
      properties:
        name:
          type: string
        owner:
          $ref: "#/components/schemas/User"
`
	spec, err := ParseSpec([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem, ok := spec.Paths["/pets"]
	if !ok {
		t.Fatal("expected /pets path")
	}
	op, ok := pathItem["get"]
	if !ok {
		t.Fatal("expected get operation")
	}
	if len(op.Tags) != 2 || op.Tags[0] != "Pets" || op.Tags[1] != "Animals" {
		t.Errorf("tags = %v, want [Pets Animals]", op.Tags)
	}

	schema, ok := spec.Components.Schemas["Pet"]
	if !ok {
		t.Fatal("expected Pet schema")
	}
	if schema.Properties["name"].Type != "string" {
		t.Errorf("name type = %q, want %q", schema.Properties["name"].Type, "string")
	}
	if schema.Properties["owner"].Ref != "#/components/schemas/User" {
		t.Errorf("owner ref = %q, want %q", schema.Properties["owner"].Ref, "#/components/schemas/User")
	}
}
