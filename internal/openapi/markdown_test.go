package openapi

import (
	"strings"
	"testing"
)

func emptyResult() *ChangelogResult {
	return &ChangelogResult{
		FeatureBranch:    "v1.0",
		AddedEndpoints:   make(map[string][]Endpoint),
		RemovedEndpoints: make(map[string][]Endpoint),
	}
}

func TestRenderMarkdown_NoChanges(t *testing.T) {
	r := emptyResult()
	md := RenderMarkdown(r)

	if !strings.Contains(md, "# API-Changelog: version v1.0") {
		t.Error("expected header line")
	}
	// No endpoint or schema sections
	if strings.Contains(md, "## API-endpoints") {
		t.Error("should not contain endpoint section with no changes")
	}
	if strings.Contains(md, "## API-Model updates") {
		t.Error("should not contain schema section with no changes")
	}
}

func TestRenderMarkdown_AddedEndpoints(t *testing.T) {
	r := emptyResult()
	r.AddedEndpoints["Users"] = []Endpoint{
		{Path: "/users", Method: "GET", Tags: []string{"Users"}},
		{Path: "/users", Method: "POST", Tags: []string{"Users"}},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "New endpoints") {
		t.Error("expected 'New endpoints' label")
	}
	if !strings.Contains(md, "[GET] /users") {
		t.Error("expected GET /users")
	}
	if !strings.Contains(md, "[POST] /users") {
		t.Error("expected POST /users")
	}
}

func TestRenderMarkdown_RemovedEndpoints(t *testing.T) {
	r := emptyResult()
	r.RemovedEndpoints["Legacy"] = []Endpoint{
		{Path: "/old", Method: "DELETE", Tags: []string{"Legacy"}},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "Removed endpoints") {
		t.Error("expected 'Removed endpoints' label")
	}
	if !strings.Contains(md, "[DELETE] /old") {
		t.Error("expected DELETE /old")
	}
}

func TestRenderMarkdown_SchemaAdded(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "NewModel",
			Status: "added",
			Fields: []SchemaField{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "string"},
			},
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "Added") {
		t.Error("expected 'Added' status")
	}
	if !strings.Contains(md, "NewModel") {
		t.Error("expected schema name")
	}
	if !strings.Contains(md, "id") || !strings.Contains(md, "name") {
		t.Error("expected field names")
	}
}

func TestRenderMarkdown_SchemaRemoved(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "OldModel",
			Status: "removed",
			Fields: []SchemaField{
				{Name: "x", Type: "string"},
			},
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "Removed") {
		t.Error("expected 'Removed' status")
	}
	if !strings.Contains(md, "OldModel") {
		t.Error("expected schema name")
	}
}

func TestRenderMarkdown_SchemaModified(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "User",
			Status: "modified",
			Fields: []SchemaField{
				{Name: "id", Type: "integer"},
				{Name: "email", Type: "string"},
			},
			AddedFields: []SchemaField{
				{Name: "email", Type: "string"},
			},
			RemovedFields: []SchemaField{
				{Name: "name", Type: "string"},
			},
			ChangedTypeFields: []FieldTypeChange{
				{Name: "status", OldType: "string", NewType: "integer"},
			},
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "User") {
		t.Error("expected schema name")
	}
	if !strings.Contains(md, "Added Fields") {
		t.Error("expected 'Added Fields' section")
	}
	if !strings.Contains(md, "email") {
		t.Error("expected added field name")
	}
	if !strings.Contains(md, "Removed Fields") {
		t.Error("expected 'Removed Fields' section")
	}
	if !strings.Contains(md, "Changed Types") {
		t.Error("expected 'Changed Types' section")
	}
	if !strings.Contains(md, "`string` -> `integer`") {
		t.Error("expected type change notation")
	}
}

func TestRenderMarkdown_SchemaAddedNoFields(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "EmptyModel",
			Status: "added",
			Fields: nil,
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "EmptyModel") {
		t.Error("expected schema name")
	}
	if !strings.Contains(md, "No fields") {
		t.Error("expected 'No fields' for added schema with no fields")
	}
}

func TestRenderMarkdown_SchemaRemovedNoFields(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "GoneModel",
			Status: "removed",
			Fields: nil,
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "GoneModel") {
		t.Error("expected schema name")
	}
	// Removed with no fields should not have a "Fields:" section
	if strings.Contains(md, "Fields:") {
		t.Error("should not contain 'Fields:' section for removed schema with no fields")
	}
}

func TestRenderMarkdown_SchemaModifiedOnlyAdded(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "GrowingModel",
			Status: "modified",
			AddedFields: []SchemaField{
				{Name: "newField", Type: "string"},
			},
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "Added Fields") {
		t.Error("expected 'Added Fields' section")
	}
	if strings.Contains(md, "Removed Fields") {
		t.Error("should not contain 'Removed Fields' when none removed")
	}
	if strings.Contains(md, "Changed Types") {
		t.Error("should not contain 'Changed Types' when none changed")
	}
}

func TestRenderMarkdown_SchemaModifiedOnlyRemoved(t *testing.T) {
	r := emptyResult()
	r.SchemaChanges = []SchemaDiff{
		{
			Name:   "ShrinkingModel",
			Status: "modified",
			RemovedFields: []SchemaField{
				{Name: "oldField", Type: "integer"},
			},
		},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "Removed Fields") {
		t.Error("expected 'Removed Fields' section")
	}
	if strings.Contains(md, "Added Fields") {
		t.Error("should not contain 'Added Fields' when none added")
	}
}

func TestRenderMarkdown_MultipleTags(t *testing.T) {
	r := emptyResult()
	r.AddedEndpoints["Beta"] = []Endpoint{
		{Path: "/beta", Method: "GET"},
	}
	r.AddedEndpoints["Alpha"] = []Endpoint{
		{Path: "/alpha", Method: "POST"},
	}

	md := RenderMarkdown(r)

	// Tags should appear in sorted order
	alphaIdx := strings.Index(md, "### Alpha:")
	betaIdx := strings.Index(md, "### Beta:")
	if alphaIdx < 0 || betaIdx < 0 {
		t.Fatal("expected both tags in output")
	}
	if alphaIdx > betaIdx {
		t.Error("tags should be sorted alphabetically")
	}
}

func TestRenderMarkdown_TagWithBothAddedAndRemoved(t *testing.T) {
	r := emptyResult()
	r.AddedEndpoints["Users"] = []Endpoint{
		{Path: "/users/new", Method: "POST"},
	}
	r.RemovedEndpoints["Users"] = []Endpoint{
		{Path: "/users/old", Method: "DELETE"},
	}

	md := RenderMarkdown(r)

	if !strings.Contains(md, "New endpoints") {
		t.Error("expected 'New endpoints'")
	}
	if !strings.Contains(md, "Removed endpoints") {
		t.Error("expected 'Removed endpoints'")
	}
}

func TestRenderEndpointList_Empty(t *testing.T) {
	var b strings.Builder
	renderEndpointList(&b, "New", nil)
	if b.Len() != 0 {
		t.Error("expected empty output for nil endpoints")
	}
}

func TestWriteFieldSection_Empty(t *testing.T) {
	var b strings.Builder
	writeFieldSection(&b, "Test", nil)
	if b.Len() != 0 {
		t.Error("expected empty output for nil fields")
	}
}

func TestWriteFieldSection_WithFields(t *testing.T) {
	var b strings.Builder
	writeFieldSection(&b, "Added Fields", []SchemaField{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
	})
	result := b.String()
	if !strings.Contains(result, "Added Fields") {
		t.Error("expected label")
	}
	if !strings.Contains(result, "id: `integer`") {
		t.Error("expected field")
	}
}

func TestRenderSchemaDiff_UnknownStatus(t *testing.T) {
	var b strings.Builder
	renderSchemaDiff(&b, SchemaDiff{Name: "X", Status: "unknown"})
	if b.Len() != 0 {
		t.Error("expected empty output for unknown status")
	}
}

func TestCollectTags(t *testing.T) {
	tests := []struct {
		name     string
		result   *ChangelogResult
		wantTags []string
	}{
		{
			name: "merges tags from added and removed",
			result: &ChangelogResult{
				AddedEndpoints: map[string][]Endpoint{
					"Users":  {{Path: "/u", Method: "GET"}},
					"Shared": {{Path: "/s", Method: "GET"}},
				},
				RemovedEndpoints: map[string][]Endpoint{
					"Legacy": {{Path: "/old", Method: "DELETE"}},
					"Shared": {{Path: "/s2", Method: "POST"}},
				},
			},
			wantTags: []string{"Legacy", "Shared", "Users"},
		},
		{
			name: "empty endpoints",
			result: &ChangelogResult{
				AddedEndpoints:   make(map[string][]Endpoint),
				RemovedEndpoints: make(map[string][]Endpoint),
			},
			wantTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := collectTags(tt.result)
			if tt.wantTags == nil {
				if len(tags) != 0 {
					t.Errorf("expected empty tags, got %v", tags)
				}
				return
			}
			if len(tags) != len(tt.wantTags) {
				t.Fatalf("tag count = %d, want %d", len(tags), len(tt.wantTags))
			}
			for _, tag := range tt.wantTags {
				if !tags[tag] {
					t.Errorf("missing tag %q", tag)
				}
			}
		})
	}
}
