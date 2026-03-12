package openapi

import "gopkg.in/yaml.v3"

// ParseSpec parses raw YAML bytes into a minimal OpenAPI Spec.
func ParseSpec(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	if spec.Paths == nil {
		spec.Paths = make(map[string]PathItem)
	}
	if spec.Components.Schemas == nil {
		spec.Components.Schemas = make(map[string]Schema)
	}
	return &spec, nil
}
