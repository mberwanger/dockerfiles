package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestImageConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *ImageConfig
		wantErr bool
	}{
		{
			name: "with base image and custom values",
			yaml: `
base_image:
  name: ubuntu
  source: docker.io
custom_key: custom_value
another_key: 123
`,
			want: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"custom_key":  "custom_value",
					"another_key": 123,
				},
			},
			wantErr: false,
		},
		{
			name: "with base image only name",
			yaml: `
base_image:
  name: alpine
`,
			want: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "alpine",
					Source: "",
				},
				Values: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "without base image",
			yaml: `
env_var: value
port: 8080
`,
			want: &ImageConfig{
				BaseImage: nil,
				Values: map[string]interface{}{
					"env_var": "value",
					"port":    8080,
				},
			},
			wantErr: false,
		},
		{
			name: "empty config",
			yaml: `{}`,
			want: &ImageConfig{
				BaseImage: nil,
				Values:    map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ImageConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !imageConfigEqual(&got, tt.want) {
				t.Errorf("UnmarshalYAML() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestImageConfig_MarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		config  *ImageConfig
		wantErr bool
	}{
		{
			name: "with base image and values",
			config: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"custom_key": "custom_value",
					"port":       8080,
				},
			},
			wantErr: false,
		},
		{
			name: "without base image",
			config: &ImageConfig{
				Values: map[string]interface{}{
					"env_var": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "empty config",
			config: &ImageConfig{
				Values: map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Unmarshal and verify round-trip
			var got ImageConfig
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Errorf("Failed to unmarshal marshaled data: %v", err)
				return
			}
			if !imageConfigEqual(&got, tt.config) {
				t.Errorf("Round-trip failed: got = %+v, want %+v", got, tt.config)
			}
		})
	}
}

func TestImageConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		config   *ImageConfig
		defaults *ImageConfig
		want     *ImageConfig
	}{
		{
			name: "merge with defaults - override base image",
			config: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"key1": "value1",
				},
			},
			defaults: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "alpine",
					Source: "default.io",
				},
				Values: map[string]interface{}{
					"key2": "value2",
				},
			},
			want: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "merge without base image - use default",
			config: &ImageConfig{
				Values: map[string]interface{}{
					"key1": "value1",
				},
			},
			defaults: &ImageConfig{
				BaseImage: &BaseImage{
					Name: "alpine",
				},
				Values: map[string]interface{}{
					"key2": "value2",
				},
			},
			want: &ImageConfig{
				BaseImage: &BaseImage{
					Name: "alpine",
				},
				Values: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "merge nested maps",
			config: &ImageConfig{
				Values: map[string]interface{}{
					"nested": map[string]interface{}{
						"key1": "override",
						"key3": "new",
					},
				},
			},
			defaults: &ImageConfig{
				Values: map[string]interface{}{
					"nested": map[string]interface{}{
						"key1": "default",
						"key2": "value2",
					},
				},
			},
			want: &ImageConfig{
				Values: map[string]interface{}{
					"nested": map[string]interface{}{
						"key1": "override",
						"key2": "value2",
						"key3": "new",
					},
				},
			},
		},
		{
			name: "nil defaults",
			config: &ImageConfig{
				BaseImage: &BaseImage{Name: "ubuntu"},
				Values:    map[string]interface{}{"key": "value"},
			},
			defaults: nil,
			want: &ImageConfig{
				BaseImage: &BaseImage{Name: "ubuntu"},
				Values:    map[string]interface{}{"key": "value"},
			},
		},
		{
			name:   "nil config",
			config: nil,
			defaults: &ImageConfig{
				BaseImage: &BaseImage{Name: "alpine"},
				Values:    map[string]interface{}{"key": "value"},
			},
			want: &ImageConfig{
				BaseImage: &BaseImage{Name: "alpine"},
				Values:    map[string]interface{}{"key": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.Merge(tt.defaults)
			if !imageConfigEqual(got, tt.want) {
				t.Errorf("Merge() got = %+v, want %+v", got, tt.want)
			}

			// Verify that merge doesn't mutate original configs
			if tt.config != nil {
				tt.config.Values["test_key"] = "test_value"
				if _, exists := got.Values["test_key"]; exists {
					t.Errorf("Merge() result was mutated by changes to original config")
				}
			}
		})
	}
}

func TestImageConfig_deepCopy(t *testing.T) {
	tests := []struct {
		name   string
		config *ImageConfig
	}{
		{
			name: "full config",
			config: &ImageConfig{
				BaseImage: &BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"key1": "value1",
					"nested": map[string]interface{}{
						"key2": "value2",
					},
					"array": []interface{}{1, 2, 3},
				},
			},
		},
		{
			name: "nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.deepCopy()
			if !imageConfigEqual(got, tt.config) {
				t.Errorf("deepCopy() got = %+v, want %+v", got, tt.config)
			}

			// Verify deep copy - mutations shouldn't affect original
			if got != nil && tt.config != nil {
				got.Values["new_key"] = "new_value"
				if _, exists := tt.config.Values["new_key"]; exists {
					t.Errorf("deepCopy() result affected original config")
				}

				if nested, ok := got.Values["nested"].(map[string]interface{}); ok {
					nested["new_nested"] = "value"
					if orig, ok := tt.config.Values["nested"].(map[string]interface{}); ok {
						if _, exists := orig["new_nested"]; exists {
							t.Errorf("deepCopy() nested map affected original config")
						}
					}
				}
			}
		})
	}
}

func TestDeepCopyValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "string",
			value: "test",
		},
		{
			name:  "int",
			value: 42,
		},
		{
			name: "map",
			value: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		},
		{
			name: "nested map",
			value: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name:  "array",
			value: []interface{}{1, "two", 3.0},
		},
		{
			name: "nested array",
			value: []interface{}{
				[]interface{}{1, 2},
				[]interface{}{3, 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepCopyValue(tt.value)
			if !deepEqual(got, tt.value) {
				t.Errorf("deepCopyValue() got = %+v, want %+v", got, tt.value)
			}
		})
	}
}

// Helper functions for testing

func imageConfigEqual(a, b *ImageConfig) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare BaseImage
	if (a.BaseImage == nil) != (b.BaseImage == nil) {
		return false
	}
	if a.BaseImage != nil {
		if a.BaseImage.Name != b.BaseImage.Name || a.BaseImage.Source != b.BaseImage.Source {
			return false
		}
	}

	// Compare Values
	return deepEqual(a.Values, b.Values)
}

func deepEqual(a, b interface{}) bool {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !deepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for i, v := range av {
			if !deepEqual(v, bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
