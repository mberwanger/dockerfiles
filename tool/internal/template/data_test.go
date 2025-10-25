package template

import (
	"strings"
	"testing"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

func TestNewData(t *testing.T) {
	tests := []struct {
		name         string
		mergedConfig *config.ImageConfig
		imageName    string
		wantValues   map[string]interface{}
	}{
		{
			name: "with base image and values",
			mergedConfig: &config.ImageConfig{
				BaseImage: &config.BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				Values: map[string]interface{}{
					"version": "1.0",
					"port":    8080,
				},
			},
			imageName: "myapp",
			wantValues: map[string]interface{}{
				"base_image": &config.BaseImage{
					Name:   "ubuntu",
					Source: "docker.io",
				},
				"version":    "1.0",
				"port":       8080,
				"image_name": "myapp",
			},
		},
		{
			name: "without base image",
			mergedConfig: &config.ImageConfig{
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			imageName: "testapp",
			wantValues: map[string]interface{}{
				"key":        "value",
				"image_name": "testapp",
			},
		},
		{
			name: "empty config",
			mergedConfig: &config.ImageConfig{
				Values: map[string]interface{}{},
			},
			imageName: "emptyapp",
			wantValues: map[string]interface{}{
				"image_name": "emptyapp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewData(tt.mergedConfig, tt.imageName)

			if data == nil {
				t.Fatal("NewData() returned nil")
			}

			// Check if all expected values are present
			for key, want := range tt.wantValues {
				got, exists := data.Values[key]
				if !exists {
					t.Errorf("Expected key %s not found in Values", key)
					continue
				}

				// Special handling for BaseImage comparison
				if key == "base_image" {
					wantBase, wantOk := want.(*config.BaseImage)
					gotBase, gotOk := got.(*config.BaseImage)
					if !wantOk || !gotOk {
						t.Errorf("base_image type mismatch")
						continue
					}
					if wantBase.Name != gotBase.Name || wantBase.Source != gotBase.Source {
						t.Errorf("base_image = %+v, want %+v", gotBase, wantBase)
					}
				} else if got != want {
					t.Errorf("Values[%s] = %v, want %v", key, got, want)
				}
			}

			// Check generation message
			if data.generationMessage == "" {
				t.Error("generationMessage should not be empty")
			}
			if !strings.Contains(data.generationMessage, tt.imageName) {
				t.Errorf("generationMessage should contain image name %s", tt.imageName)
			}

			// Check rootPathIncluded is initialized to false
			if data.rootPathIncluded {
				t.Error("rootPathIncluded should be false initially")
			}
		})
	}
}

func TestData_get(t *testing.T) {
	data := &Data{
		Values: map[string]interface{}{
			"string_key": "string_value",
			"int_key":    42,
			"bool_key":   true,
		},
	}

	tests := []struct {
		name string
		key  string
		want interface{}
	}{
		{
			name: "string value",
			key:  "string_key",
			want: "string_value",
		},
		{
			name: "int value",
			key:  "int_key",
			want: 42,
		},
		{
			name: "bool value",
			key:  "bool_key",
			want: true,
		},
		{
			name: "nonexistent key",
			key:  "nonexistent",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.get(tt.key)
			if got != tt.want {
				t.Errorf("get(%s) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestData_fromImage(t *testing.T) {
	tests := []struct {
		name      string
		data      *Data
		baseImage interface{}
		want      string
	}{
		{
			name: "BaseImage with dockerhub source",
			data: &Data{
				Values:           map[string]interface{}{},
				rootPathIncluded: false,
			},
			baseImage: &config.BaseImage{
				Name:   "ubuntu:20.04",
				Source: "dockerhub",
			},
			want: "FROM ubuntu:20.04",
		},
		{
			name: "BaseImage with custom registry - first call",
			data: &Data{
				Values: map[string]interface{}{
					"registry": "my-registry.io",
				},
				rootPathIncluded: false,
			},
			baseImage: &config.BaseImage{
				Name:   "myimage",
				Source: "custom",
			},
			want: "ARG REGISTRY=my-registry.io\nFROM ${REGISTRY}/myimage",
		},
		{
			name: "BaseImage with custom registry - second call (ARG already included)",
			data: &Data{
				Values: map[string]interface{}{
					"registry": "my-registry.io",
				},
				rootPathIncluded: true,
			},
			baseImage: &config.BaseImage{
				Name:   "anotherimage",
				Source: "custom",
			},
			want: "FROM ${REGISTRY}/anotherimage",
		},
		{
			name: "string base image without registry",
			data: &Data{
				Values: map[string]interface{}{
					"registry": "test.io",
				},
				rootPathIncluded: false,
			},
			baseImage: "alpine:latest",
			want:      "ARG REGISTRY=test.io\nFROM ${REGISTRY}/alpine:latest",
		},
		{
			name: "map base image",
			data: &Data{
				Values: map[string]interface{}{
					"registry": "test.io",
				},
				rootPathIncluded: false,
			},
			baseImage: map[string]interface{}{
				"name":   "testimage",
				"source": "custom",
			},
			want: "ARG REGISTRY=test.io\nFROM ${REGISTRY}/testimage",
		},
		{
			name: "string reference to base_image value",
			data: &Data{
				Values: map[string]interface{}{
					"my_base": &config.BaseImage{
						Name:   "ubuntu",
						Source: "dockerhub",
					},
				},
				rootPathIncluded: false,
			},
			baseImage: "my_base",
			want:      "FROM ubuntu",
		},
		{
			name: "no registry set when needed",
			data: &Data{
				Values:           map[string]interface{}{},
				rootPathIncluded: false,
			},
			baseImage: &config.BaseImage{
				Name:   "myimage",
				Source: "custom",
			},
			want: "# ERROR: registry not set in config\nFROM ${REGISTRY}/myimage",
		},
		{
			name: "registry is not a string",
			data: &Data{
				Values: map[string]interface{}{
					"registry": 12345,
				},
				rootPathIncluded: false,
			},
			baseImage: &config.BaseImage{
				Name:   "myimage",
				Source: "custom",
			},
			want: "# ERROR: registry is not a string\nFROM ${REGISTRY}/myimage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.fromImage(tt.baseImage)
			if got != tt.want {
				t.Errorf("fromImage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestData_functions(t *testing.T) {
	mergedConfig := &config.ImageConfig{
		BaseImage: &config.BaseImage{
			Name:   "ubuntu",
			Source: "docker.io",
		},
		Values: map[string]interface{}{
			"version": "1.0",
			"port":    8080,
		},
	}

	data := NewData(mergedConfig, "testapp")
	funcMap := data.functions()

	// Test generation_message function
	if genMsgFunc, ok := funcMap["generation_message"]; ok {
		if fn, ok := genMsgFunc.(func() string); ok {
			msg := fn()
			if msg == "" {
				t.Error("generation_message() should not return empty string")
			}
			if !strings.Contains(msg, "testapp") {
				t.Error("generation_message() should contain image name")
			}
		} else {
			t.Error("generation_message is not a func() string")
		}
	} else {
		t.Error("generation_message function not found")
	}

	// Test from_image function
	if fromImageFunc, ok := funcMap["from_image"]; ok {
		if fn, ok := fromImageFunc.(func(interface{}) string); ok {
			// Test with base_image key
			result := fn("base_image")
			if !strings.Contains(result, "FROM") {
				t.Errorf("from_image(\"base_image\") should contain FROM, got: %s", result)
			}

			// Test with direct BaseImage
			result2 := fn(&config.BaseImage{Name: "alpine", Source: "dockerhub"})
			if result2 != "FROM alpine" {
				t.Errorf("from_image(BaseImage) = %s, want FROM alpine", result2)
			}
		} else {
			t.Error("from_image is not a func(interface{}) string")
		}
	} else {
		t.Error("from_image function not found")
	}

	// Test get function
	if _, ok := funcMap["get"]; ok {
		// Type assertion is complex due to the function signature, just verify it exists
		t.Log("get function exists")
	} else {
		t.Error("get function not found")
	}
}

func TestGenerateMessage(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		wantParts []string
	}{
		{
			name:      "normal image name",
			imageName: "myapp",
			wantParts: []string{
				"GENERATED FILE, DO NOT MODIFY",
				"myapp",
				"go run tool/main.go generate image",
				"go run tool/main.go generate all",
			},
		},
		{
			name:      "image with special chars",
			imageName: "my-app-123",
			wantParts: []string{
				"GENERATED FILE, DO NOT MODIFY",
				"my-app-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateMessage(tt.imageName)

			if got == "" {
				t.Error("generateMessage() returned empty string")
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("generateMessage() should contain %q, got: %s", part, got)
				}
			}

			// Should be formatted as comments
			lines := strings.Split(got, "\n")
			for _, line := range lines {
				if line != "" && !strings.HasPrefix(line, "#") {
					t.Errorf("All non-empty lines should start with #, got: %s", line)
				}
			}
		})
	}
}

func TestData_fromImage_RootPathIncluded(t *testing.T) {
	data := &Data{
		Values: map[string]interface{}{
			"registry": "my-registry.io",
		},
		rootPathIncluded: false,
	}

	// First call should include ARG
	result1 := data.fromImage(&config.BaseImage{
		Name:   "image1",
		Source: "custom",
	})

	if !strings.Contains(result1, "ARG REGISTRY=") {
		t.Error("First call should include ARG REGISTRY")
	}

	if !data.rootPathIncluded {
		t.Error("rootPathIncluded should be true after first call")
	}

	// Second call should not include ARG
	result2 := data.fromImage(&config.BaseImage{
		Name:   "image2",
		Source: "custom",
	})

	if strings.Contains(result2, "ARG REGISTRY=") {
		t.Error("Second call should not include ARG REGISTRY")
	}
}

func TestData_fromImage_RecursiveReference(t *testing.T) {
	data := &Data{
		Values: map[string]interface{}{
			"my_base":  "ubuntu:20.04",
			"registry": "test.io",
		},
		rootPathIncluded: false,
	}

	// Test referencing a string value - it resolves to a string which still needs registry
	result := data.fromImage("my_base")
	expected := "ARG REGISTRY=test.io\nFROM ${REGISTRY}/ubuntu:20.04"
	if result != expected {
		t.Errorf("fromImage(\"my_base\") = %s, want %s", result, expected)
	}
}

func TestData_fromImage_NestedReference(t *testing.T) {
	data := &Data{
		Values: map[string]interface{}{
			"level1": "level2",
			"level2": &config.BaseImage{
				Name:   "alpine",
				Source: "dockerhub",
			},
		},
		rootPathIncluded: false,
	}

	// Test nested reference
	result := data.fromImage("level1")
	expected := "FROM alpine"
	if result != expected {
		t.Errorf("fromImage(\"level1\") = %s, want %s", result, expected)
	}
}
