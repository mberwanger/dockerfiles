package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		templateData string
		data         *Data
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "simple template",
			templateData: "FROM {{base_image.Name}}\nVERSION {{version}}\n",
			data: NewData(&config.ImageConfig{
				BaseImage: &config.BaseImage{
					Name:   "ubuntu:20.04",
					Source: "dockerhub",
				},
				Values: map[string]interface{}{
					"version": "1.0",
				},
			}, "testapp"),
			wantContains: []string{"ubuntu:20.04", "1.0"},
			wantErr:      false,
		},
		{
			name:         "template with functions",
			templateData: "{{generation_message}}\n{{from_image \"base_image\"}}\n",
			data: NewData(&config.ImageConfig{
				BaseImage: &config.BaseImage{
					Name:   "alpine",
					Source: "dockerhub",
				},
				Values: map[string]interface{}{},
			}, "myapp"),
			wantContains: []string{"GENERATED FILE", "FROM alpine"},
			wantErr:      false,
		},
		{
			name:         "template with get function",
			templateData: "Port: {{get \"port\"}}\n",
			data: NewData(&config.ImageConfig{
				Values: map[string]interface{}{
					"port": 8080,
				},
			}, "testapp"),
			wantContains: []string{"Port: 8080"},
			wantErr:      false,
		},
		{
			name:         "template with custom value functions",
			templateData: "Version: {{version}}\nName: {{image_name}}\n",
			data: NewData(&config.ImageConfig{
				Values: map[string]interface{}{
					"version": "2.0",
				},
			}, "customapp"),
			wantContains: []string{"Version: 2.0", "Name: customapp"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create template file
			templatePath := filepath.Join(tmpDir, tt.name+".tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.templateData), 0644); err != nil {
				t.Fatalf("Failed to write template file: %v", err)
			}

			// Create output path
			outputPath := filepath.Join(tmpDir, tt.name+".out")

			// Execute WriteFile
			err := WriteFile(templatePath, outputPath, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Read and verify output
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			output := string(content)
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestWriteFile_TemplateNotFound(t *testing.T) {
	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	err := WriteFile("/nonexistent/template.tmpl", "/tmp/output", data)
	if err == nil {
		t.Error("WriteFile() should return error for nonexistent template")
	}
}

func TestWriteFile_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid template
	templatePath := filepath.Join(tmpDir, "invalid.tmpl")
	invalidTemplate := "{{range .missing}}\n"
	if err := os.WriteFile(templatePath, []byte(invalidTemplate), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output")
	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	err := WriteFile(templatePath, outputPath, data)
	if err == nil {
		t.Error("WriteFile() should return error for invalid template")
	}
}

func TestRender(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		templateData string
		data         *Data
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "basic rendering",
			templateData: "Name: {{image_name}}\n",
			data: NewData(&config.ImageConfig{
				Values: map[string]interface{}{},
			}, "myapp"),
			wantContains: []string{"Name: myapp"},
			wantErr:      false,
		},
		{
			name:         "complex template",
			templateData: "{{generation_message}}\n{{from_image \"base_image\"}}\nVERSION {{version}}\n",
			data: NewData(&config.ImageConfig{
				BaseImage: &config.BaseImage{
					Name:   "ubuntu",
					Source: "dockerhub",
				},
				Values: map[string]interface{}{
					"version": "3.0",
				},
			}, "complexapp"),
			wantContains: []string{"GENERATED FILE", "FROM ubuntu", "VERSION 3.0"},
			wantErr:      false,
		},
		{
			name:         "template with conditionals",
			templateData: "{{if version}}Version: {{version}}{{end}}\n",
			data: NewData(&config.ImageConfig{
				Values: map[string]interface{}{
					"version": "1.5",
				},
			}, "app"),
			wantContains: []string{"Version: 1.5"},
			wantErr:      false,
		},
		{
			name:         "template with map values",
			templateData: "Config: {{get \"config\"}}\n",
			data: NewData(&config.ImageConfig{
				Values: map[string]interface{}{
					"config": map[string]interface{}{
						"key": "value",
					},
				},
			}, "app"),
			wantContains: []string{"Config: map[key:value]"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create template file
			templatePath := filepath.Join(tmpDir, tt.name+".tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.templateData), 0644); err != nil {
				t.Fatalf("Failed to write template file: %v", err)
			}

			// Execute render
			output, err := render(templatePath, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestRender_FileNotFound(t *testing.T) {
	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	_, err := render("/nonexistent/template.tmpl", data)
	if err == nil {
		t.Error("render() should return error for nonexistent template file")
	}
}

func TestRender_ParseError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create template with syntax error
	templatePath := filepath.Join(tmpDir, "bad.tmpl")
	badTemplate := "{{if .test}\n" // Missing end
	if err := os.WriteFile(templatePath, []byte(badTemplate), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	_, err := render(templatePath, data)
	if err == nil {
		t.Error("render() should return error for template with parse error")
	}
}

func TestRender_ExecutionError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create template that will fail during execution
	templatePath := filepath.Join(tmpDir, "exec_error.tmpl")
	errorTemplate := "{{.NonexistentField.Method}}\n"
	if err := os.WriteFile(templatePath, []byte(errorTemplate), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	_, err := render(templatePath, data)
	if err == nil {
		t.Error("render() should return error for template execution error")
	}
}

func TestRender_ValueFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		templateData string
		values       map[string]interface{}
		want         string
	}{
		{
			name:         "string value function",
			templateData: "{{mystring}}\n",
			values: map[string]interface{}{
				"mystring": "hello world",
			},
			want: "hello world",
		},
		{
			name:         "int value function",
			templateData: "{{myint}}\n",
			values: map[string]interface{}{
				"myint": 42,
			},
			want: "42",
		},
		{
			name:         "float value function",
			templateData: "{{myfloat}}\n",
			values: map[string]interface{}{
				"myfloat": 3.14,
			},
			want: "3.14",
		},
		{
			name:         "map value function",
			templateData: "{{range $k, $v := mymap}}{{$k}}:{{$v}} {{end}}\n",
			values: map[string]interface{}{
				"mymap": map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
			},
			want: "key1:val1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create template file
			templatePath := filepath.Join(tmpDir, tt.name+".tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.templateData), 0644); err != nil {
				t.Fatalf("Failed to write template file: %v", err)
			}

			data := NewData(&config.ImageConfig{
				Values: tt.values,
			}, "testapp")

			output, err := render(templatePath, data)
			if err != nil {
				t.Fatalf("render() error = %v", err)
			}

			if !strings.Contains(output, tt.want) {
				t.Errorf("Output should contain %q, got: %s", tt.want, output)
			}
		})
	}
}

func TestRender_ComplexDockerfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a realistic Dockerfile template
	templatePath := filepath.Join(tmpDir, "Dockerfile.tmpl")
	dockerfileTemplate := `{{generation_message}}
{{from_image "base_image"}}

WORKDIR /app

COPY . .

{{if build_command}}
RUN {{build_command}}
{{end}}

EXPOSE {{port}}

CMD ["./{{image_name}}"]
`
	if err := os.WriteFile(templatePath, []byte(dockerfileTemplate), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	data := NewData(&config.ImageConfig{
		BaseImage: &config.BaseImage{
			Name:   "golang:1.21",
			Source: "dockerhub",
		},
		Values: map[string]interface{}{
			"port":          8080,
			"build_command": "go build -o app",
		},
	}, "myservice")

	output, err := render(templatePath, data)
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}

	expectedParts := []string{
		"GENERATED FILE",
		"FROM golang:1.21",
		"WORKDIR /app",
		"COPY . .",
		"RUN go build -o app",
		"EXPOSE 8080",
		`CMD ["./myservice"]`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q, got:\n%s", part, output)
		}
	}
}

func TestWriteFile_Permissions(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := filepath.Join(tmpDir, "template.tmpl")
	templateData := "FROM alpine\n"
	if err := os.WriteFile(templatePath, []byte(templateData), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output")
	data := NewData(&config.ImageConfig{
		Values: map[string]interface{}{},
	}, "testapp")

	if err := WriteFile(templatePath, outputPath, data); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Check file permissions
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Mode().Perm() != 0644 {
		t.Errorf("File permissions = %o, want 0644", info.Mode().Perm())
	}
}

func TestRender_WithRegistryArg(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := filepath.Join(tmpDir, "Dockerfile.tmpl")
	templateData := "{{from_image \"base_image\"}}\nRUN echo test\n"
	if err := os.WriteFile(templatePath, []byte(templateData), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	data := NewData(&config.ImageConfig{
		BaseImage: &config.BaseImage{
			Name:   "myimage",
			Source: "custom",
		},
		Values: map[string]interface{}{
			"registry": "my-registry.io",
		},
	}, "testapp")

	output, err := render(templatePath, data)
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}

	if !strings.Contains(output, "ARG REGISTRY=my-registry.io") {
		t.Error("Output should contain ARG REGISTRY")
	}
	if !strings.Contains(output, "FROM ${REGISTRY}/myimage") {
		t.Error("Output should contain FROM with registry variable")
	}
}
