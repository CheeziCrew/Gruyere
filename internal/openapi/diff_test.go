package openapi

import (
	"testing"
)

func TestGetFieldType(t *testing.T) {
	tests := []struct {
		name string
		prop Property
		want string
	}{
		{"ref set", Property{Ref: "#/components/schemas/Foo"}, "$ref"},
		{"type empty", Property{}, "object"},
		{"type set", Property{Type: "string"}, "string"},
		{"ref takes precedence over type", Property{Type: "string", Ref: "#/ref"}, "$ref"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFieldType(tt.prop); got != tt.want {
				t.Errorf("getFieldType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetComponentFields(t *testing.T) {
	tests := []struct {
		name   string
		schema Schema
		want   []SchemaField
	}{
		{
			name: "multiple properties sorted by name",
			schema: Schema{Properties: map[string]Property{
				"zebra": {Type: "string"},
				"alpha": {Type: "integer"},
				"middle": {Ref: "#/ref"},
			}},
			want: []SchemaField{
				{Name: "alpha", Type: "integer"},
				{Name: "middle", Type: "$ref"},
				{Name: "zebra", Type: "string"},
			},
		},
		{
			name:   "empty properties",
			schema: Schema{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getComponentFields(tt.schema)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFieldsToMap(t *testing.T) {
	fields := []SchemaField{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
	}
	m := fieldsToMap(fields)
	if len(m) != 2 {
		t.Fatalf("len = %d, want 2", len(m))
	}
	if m["id"] != "integer" {
		t.Errorf("id = %q, want %q", m["id"], "integer")
	}
	if m["name"] != "string" {
		t.Errorf("name = %q, want %q", m["name"], "string")
	}
}

func TestFindAddedFields(t *testing.T) {
	tests := []struct {
		name     string
		source   map[string]string
		baseline map[string]string
		want     int
	}{
		{
			name:     "fields in source not in baseline",
			source:   map[string]string{"a": "string", "b": "integer", "c": "boolean"},
			baseline: map[string]string{"a": "string"},
			want:     2,
		},
		{
			name:     "empty source",
			source:   map[string]string{},
			baseline: map[string]string{"a": "string"},
			want:     0,
		},
		{
			name:     "empty baseline returns all source",
			source:   map[string]string{"x": "string", "y": "integer"},
			baseline: map[string]string{},
			want:     2,
		},
		{
			name:     "no difference",
			source:   map[string]string{"a": "string"},
			baseline: map[string]string{"a": "string"},
			want:     0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAddedFields(tt.source, tt.baseline)
			if len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestFindAddedFields_Sorted(t *testing.T) {
	source := map[string]string{"z": "string", "a": "integer", "m": "boolean"}
	baseline := map[string]string{}
	got := findAddedFields(source, baseline)
	for i := 1; i < len(got); i++ {
		if got[i-1].Name >= got[i].Name {
			t.Errorf("not sorted: %q >= %q", got[i-1].Name, got[i].Name)
		}
	}
}

func TestFindChangedFields(t *testing.T) {
	tests := []struct {
		name      string
		oldFields map[string]string
		newFields map[string]string
		want      int
	}{
		{
			name:      "different types detected",
			oldFields: map[string]string{"a": "string", "b": "integer"},
			newFields: map[string]string{"a": "integer", "b": "integer"},
			want:      1,
		},
		{
			name:      "same types not included",
			oldFields: map[string]string{"a": "string"},
			newFields: map[string]string{"a": "string"},
			want:      0,
		},
		{
			name:      "empty maps",
			oldFields: map[string]string{},
			newFields: map[string]string{},
			want:      0,
		},
		{
			name:      "field only in new not counted as changed",
			oldFields: map[string]string{},
			newFields: map[string]string{"a": "string"},
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findChangedFields(tt.oldFields, tt.newFields)
			if len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestFindChangedFields_Values(t *testing.T) {
	old := map[string]string{"status": "string"}
	new := map[string]string{"status": "integer"}
	got := findChangedFields(old, new)
	if len(got) != 1 {
		t.Fatalf("expected 1 change, got %d", len(got))
	}
	if got[0].Name != "status" || got[0].OldType != "string" || got[0].NewType != "integer" {
		t.Errorf("got %+v", got[0])
	}
}

func TestExtractEndpoints(t *testing.T) {
	spec := &Spec{
		Paths: map[string]PathItem{
			"/users": {
				"get":        {Tags: []string{"Users"}},
				"post":       {Tags: []string{"Users"}},
				"parameters": {Tags: []string{"Ignored"}},
			},
			"/items": {
				"delete": {Tags: []string{"Items"}},
			},
		},
	}

	eps := extractEndpoints(spec)

	// Should have 3 endpoints (get, post, delete), not 4 (parameters skipped)
	if len(eps) != 3 {
		t.Errorf("endpoint count = %d, want 3", len(eps))
	}

	// Verify parameters key is not present
	for key := range eps {
		if key[1] == "PARAMETERS" {
			t.Error("parameters should be filtered out")
		}
	}

	// Verify methods are uppercased
	if _, ok := eps[[2]string{"/users", "GET"}]; !ok {
		t.Error("expected /users GET")
	}
	if _, ok := eps[[2]string{"/users", "POST"}]; !ok {
		t.Error("expected /users POST")
	}
}

func TestExtractEndpoints_EmptyTags(t *testing.T) {
	spec := &Spec{
		Paths: map[string]PathItem{
			"/no-tags": {
				"get": {Tags: nil},
			},
		},
	}

	eps := extractEndpoints(spec)
	key := [2]string{"/no-tags", "GET"}
	tags, ok := eps[key]
	if !ok {
		t.Fatal("expected endpoint")
	}
	if len(tags) != 1 || tags[0] != "No Tag" {
		t.Errorf("tags = %v, want [No Tag]", tags)
	}
}

func TestAddEndpointByTags(t *testing.T) {
	m := make(map[string][]Endpoint)
	key := [2]string{"/users", "GET"}
	tags := []string{"Users", "Admin"}

	addEndpointByTags(m, key, tags)

	if len(m) != 2 {
		t.Fatalf("tag count = %d, want 2", len(m))
	}
	if len(m["Users"]) != 1 {
		t.Errorf("Users endpoints = %d, want 1", len(m["Users"]))
	}
	if len(m["Admin"]) != 1 {
		t.Errorf("Admin endpoints = %d, want 1", len(m["Admin"]))
	}
	if m["Users"][0].Path != "/users" || m["Users"][0].Method != "GET" {
		t.Errorf("endpoint = %+v", m["Users"][0])
	}
}

func TestDiffOneSchema(t *testing.T) {
	tests := []struct {
		name      string
		old       Schema
		new       Schema
		wantNil   bool
		wantAdded int
		wantRmvd  int
		wantChgd  int
	}{
		{
			name: "identical schemas returns nil",
			old:  Schema{Properties: map[string]Property{"id": {Type: "integer"}}},
			new:  Schema{Properties: map[string]Property{"id": {Type: "integer"}}},
			wantNil: true,
		},
		{
			name:      "added field",
			old:       Schema{Properties: map[string]Property{"id": {Type: "integer"}}},
			new:       Schema{Properties: map[string]Property{"id": {Type: "integer"}, "name": {Type: "string"}}},
			wantAdded: 1,
		},
		{
			name:     "removed field",
			old:      Schema{Properties: map[string]Property{"id": {Type: "integer"}, "name": {Type: "string"}}},
			new:      Schema{Properties: map[string]Property{"id": {Type: "integer"}}},
			wantRmvd: 1,
		},
		{
			name:     "changed type",
			old:      Schema{Properties: map[string]Property{"status": {Type: "string"}}},
			new:      Schema{Properties: map[string]Property{"status": {Type: "integer"}}},
			wantChgd: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := diffOneSchema("TestSchema", tt.old, tt.new)
			if tt.wantNil {
				if diff != nil {
					t.Fatal("expected nil diff for identical schemas")
				}
				return
			}
			if diff == nil {
				t.Fatal("expected non-nil diff")
			}
			if diff.Status != "modified" {
				t.Errorf("status = %q, want %q", diff.Status, "modified")
			}
			if len(diff.AddedFields) != tt.wantAdded {
				t.Errorf("added = %d, want %d", len(diff.AddedFields), tt.wantAdded)
			}
			if len(diff.RemovedFields) != tt.wantRmvd {
				t.Errorf("removed = %d, want %d", len(diff.RemovedFields), tt.wantRmvd)
			}
			if len(diff.ChangedTypeFields) != tt.wantChgd {
				t.Errorf("changed = %d, want %d", len(diff.ChangedTypeFields), tt.wantChgd)
			}
		})
	}
}

func TestDiff_FullEndToEnd(t *testing.T) {
	oldSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {
				"get": {Tags: []string{"Users"}},
			},
			"/old-endpoint": {
				"delete": {Tags: []string{"Legacy"}},
			},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{
				"User": {Properties: map[string]Property{
					"id":   {Type: "integer"},
					"name": {Type: "string"},
				}},
				"OldModel": {Properties: map[string]Property{
					"x": {Type: "string"},
				}},
			},
		},
	}

	newSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {
				"get": {Tags: []string{"Users"}},
			},
			"/new-endpoint": {
				"post": {Tags: []string{"New"}},
			},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{
				"User": {Properties: map[string]Property{
					"id":    {Type: "integer"},
					"name":  {Type: "integer"}, // changed type
					"email": {Type: "string"},  // added field
				}},
				"NewModel": {Properties: map[string]Property{
					"y": {Type: "boolean"},
				}},
			},
		},
	}

	result := Diff(oldSpec, newSpec, "main", "feature", "api.yaml")

	if !result.HasChanges {
		t.Fatal("expected HasChanges = true")
	}
	if result.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q", result.BaseBranch)
	}
	if result.FeatureBranch != "feature" {
		t.Errorf("FeatureBranch = %q", result.FeatureBranch)
	}

	// Added endpoint
	if len(result.AddedEndpoints) == 0 {
		t.Error("expected added endpoints")
	}
	if eps, ok := result.AddedEndpoints["New"]; !ok || len(eps) != 1 {
		t.Errorf("expected 1 added endpoint under 'New' tag, got %v", result.AddedEndpoints)
	}

	// Removed endpoint
	if len(result.RemovedEndpoints) == 0 {
		t.Error("expected removed endpoints")
	}
	if eps, ok := result.RemovedEndpoints["Legacy"]; !ok || len(eps) != 1 {
		t.Errorf("expected 1 removed endpoint under 'Legacy' tag, got %v", result.RemovedEndpoints)
	}

	// Schema changes: NewModel added, OldModel removed, User modified
	if len(result.SchemaChanges) != 3 {
		t.Fatalf("schema changes = %d, want 3", len(result.SchemaChanges))
	}

	// Sorted by name: NewModel, OldModel, User
	schemaMap := make(map[string]SchemaDiff)
	for _, sc := range result.SchemaChanges {
		schemaMap[sc.Name] = sc
	}

	if sc, ok := schemaMap["NewModel"]; !ok || sc.Status != "added" {
		t.Error("expected NewModel as added")
	}
	if sc, ok := schemaMap["OldModel"]; !ok || sc.Status != "removed" {
		t.Error("expected OldModel as removed")
	}
	if sc, ok := schemaMap["User"]; !ok || sc.Status != "modified" {
		t.Error("expected User as modified")
	} else {
		if len(sc.AddedFields) != 1 {
			t.Errorf("User added fields = %d, want 1", len(sc.AddedFields))
		}
		if len(sc.ChangedTypeFields) != 1 {
			t.Errorf("User changed fields = %d, want 1", len(sc.ChangedTypeFields))
		}
	}
}

func TestDiff_IdenticalSpecs(t *testing.T) {
	spec := &Spec{
		Paths: map[string]PathItem{
			"/users": {"get": {Tags: []string{"Users"}}},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{
				"User": {Properties: map[string]Property{"id": {Type: "integer"}}},
			},
		},
	}

	result := Diff(spec, spec, "main", "main", "api.yaml")
	if result.HasChanges {
		t.Error("expected HasChanges = false for identical specs")
	}
}

func TestSortEndpointMap(t *testing.T) {
	m := map[string][]Endpoint{
		"Users": {
			{Path: "/users", Method: "POST"},
			{Path: "/users", Method: "GET"},
			{Path: "/accounts", Method: "GET"},
		},
	}

	sortEndpointMap(m)

	eps := m["Users"]
	if len(eps) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(eps))
	}
	// Should be sorted by path then method
	if eps[0].Path != "/accounts" || eps[0].Method != "GET" {
		t.Errorf("[0] = %v, want /accounts GET", eps[0])
	}
	if eps[1].Path != "/users" || eps[1].Method != "GET" {
		t.Errorf("[1] = %v, want /users GET", eps[1])
	}
	if eps[2].Path != "/users" || eps[2].Method != "POST" {
		t.Errorf("[2] = %v, want /users POST", eps[2])
	}
}

func TestSortEndpointMap_SamePath(t *testing.T) {
	m := map[string][]Endpoint{
		"Tag": {
			{Path: "/api", Method: "DELETE"},
			{Path: "/api", Method: "GET"},
			{Path: "/api", Method: "POST"},
		},
	}

	sortEndpointMap(m)

	eps := m["Tag"]
	if eps[0].Method != "DELETE" || eps[1].Method != "GET" || eps[2].Method != "POST" {
		t.Errorf("expected DELETE, GET, POST order, got %v", eps)
	}
}

func TestSortEndpointMap_Empty(t *testing.T) {
	m := map[string][]Endpoint{}
	sortEndpointMap(m) // should not panic
}

func TestSortEndpointMap_MultipleTags(t *testing.T) {
	m := map[string][]Endpoint{
		"Users": {
			{Path: "/users", Method: "POST"},
			{Path: "/users", Method: "GET"},
		},
		"Admin": {
			{Path: "/admin/z", Method: "GET"},
			{Path: "/admin/a", Method: "GET"},
		},
	}

	sortEndpointMap(m)

	if m["Users"][0].Method != "GET" {
		t.Errorf("Users[0] = %s, want GET", m["Users"][0].Method)
	}
	if m["Admin"][0].Path != "/admin/a" {
		t.Errorf("Admin[0] = %s, want /admin/a", m["Admin"][0].Path)
	}
}

func TestFindChangedFields_Sorted(t *testing.T) {
	old := map[string]string{"z": "string", "a": "string", "m": "string"}
	new := map[string]string{"z": "integer", "a": "boolean", "m": "number"}
	got := findChangedFields(old, new)
	if len(got) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Name >= got[i].Name {
			t.Errorf("not sorted: %q >= %q", got[i-1].Name, got[i].Name)
		}
	}
}

func TestFindChangedFields_FieldOnlyInOld(t *testing.T) {
	old := map[string]string{"removed": "string"}
	new := map[string]string{}
	got := findChangedFields(old, new)
	if len(got) != 0 {
		t.Errorf("expected 0 changes for field only in old, got %d", len(got))
	}
}

func TestDiffOneSchema_AllChangeTypes(t *testing.T) {
	old := Schema{Properties: map[string]Property{
		"kept":    {Type: "string"},
		"changed": {Type: "string"},
		"removed": {Type: "integer"},
	}}
	new := Schema{Properties: map[string]Property{
		"kept":    {Type: "string"},
		"changed": {Type: "boolean"},
		"added":   {Type: "number"},
	}}

	diff := diffOneSchema("All", old, new)
	if diff == nil {
		t.Fatal("expected non-nil diff")
	}
	if len(diff.AddedFields) != 1 {
		t.Errorf("added = %d, want 1", len(diff.AddedFields))
	}
	if len(diff.RemovedFields) != 1 {
		t.Errorf("removed = %d, want 1", len(diff.RemovedFields))
	}
	if len(diff.ChangedTypeFields) != 1 {
		t.Errorf("changed = %d, want 1", len(diff.ChangedTypeFields))
	}
}

func TestDiff_OnlySchemaChanges(t *testing.T) {
	oldSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {"get": {Tags: []string{"Users"}}},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{
				"User": {Properties: map[string]Property{"id": {Type: "integer"}}},
			},
		},
	}

	newSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {"get": {Tags: []string{"Users"}}},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{
				"User": {Properties: map[string]Property{"id": {Type: "integer"}, "name": {Type: "string"}}},
			},
		},
	}

	result := Diff(oldSpec, newSpec, "main", "feature", "api.yaml")
	if !result.HasChanges {
		t.Error("expected HasChanges")
	}
	if len(result.AddedEndpoints) != 0 {
		t.Error("expected no added endpoints")
	}
	if len(result.SchemaChanges) != 1 {
		t.Errorf("expected 1 schema change, got %d", len(result.SchemaChanges))
	}
}

func TestDiff_OnlyEndpointChanges(t *testing.T) {
	oldSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {"get": {Tags: []string{"Users"}}},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{},
		},
	}

	newSpec := &Spec{
		Paths: map[string]PathItem{
			"/users": {"get": {Tags: []string{"Users"}}},
			"/items": {"post": {Tags: []string{"Items"}}},
		},
		Components: struct {
			Schemas map[string]Schema `yaml:"schemas"`
		}{
			Schemas: map[string]Schema{},
		},
	}

	result := Diff(oldSpec, newSpec, "main", "feature", "api.yaml")
	if !result.HasChanges {
		t.Error("expected HasChanges")
	}
	if len(result.SchemaChanges) != 0 {
		t.Error("expected no schema changes")
	}
}
