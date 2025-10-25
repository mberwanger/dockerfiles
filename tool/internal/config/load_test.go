package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_FromStdin(t *testing.T) {
	// Save original stdin and restore after test
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("Failed to close pipe reader: %v", err)
		}
	}()

	os.Stdin = r

	// Write test config to pipe
	testConfig := `version: 1
defaults:
  registry: test.io
images:
  test:
    path: test/path
    versions:
      v1:
        base_image:
          name: ubuntu
`
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				t.Errorf("Failed to close pipe writer: %v", err)
			}
		}()
		if _, err := w.Write([]byte(testConfig)); err != nil {
			t.Errorf("Failed to write to pipe: %v", err)
		}
	}()

	config, err := Load("-")
	if err != nil {
		t.Fatalf("Load(\"-\") error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}
	if config.Defaults.Registry != "test.io" {
		t.Errorf("Registry = %s, want test.io", config.Defaults.Registry)
	}
	if config.Defaults.BasePath == "" {
		t.Error("BasePath should be set to current working directory")
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "manifest.yaml")

	testConfig := `version: 1
defaults:
  registry: test.io
images:
  test-image:
    path: images/test
    versions:
      v1:
        base_image:
          name: ubuntu
          source: docker.io
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}
	if config.Defaults.Registry != "test.io" {
		t.Errorf("Registry = %s, want test.io", config.Defaults.Registry)
	}
	if config.Defaults.BasePath != tmpDir {
		t.Errorf("BasePath = %s, want %s", config.Defaults.BasePath, tmpDir)
	}
	if _, exists := config.Images["test-image"]; !exists {
		t.Error("Image 'test-image' not found in config")
	}
}

func TestLoad_DefaultLocations(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tests := []struct {
		name     string
		filename string
	}{
		{"images/manifest.yml", "images/manifest.yml"},
		{"images/manifest.yaml", "images/manifest.yaml"},
		{"manifest.yml", "manifest.yml"},
		{"manifest.yaml", "manifest.yaml"},
		{".manifest.yml", ".manifest.yml"},
		{".manifest.yaml", ".manifest.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing files
			for _, f := range []string{
				"images/manifest.yml",
				"images/manifest.yaml",
				"manifest.yml",
				"manifest.yaml",
				".manifest.yml",
				".manifest.yaml",
			} {
				_ = os.Remove(f) // Ignore errors, file may not exist
			}
			_ = os.RemoveAll("images") // Ignore errors, dir may not exist

			// Create directory if needed
			if strings.Contains(tt.filename, "/") {
				dir := filepath.Dir(tt.filename)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
			}

			testConfig := `version: 1
images:
  test:
    versions:
      v1: {}
`
			if err := os.WriteFile(tt.filename, []byte(testConfig), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			config, err := Load("")
			if err != nil {
				t.Fatalf("Load(\"\") error = %v", err)
			}

			if config.Version != 1 {
				t.Errorf("Version = %d, want 1", config.Version)
			}
		})
	}
}

func TestLoad_NoConfigFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err = Load("")
	if err == nil {
		t.Error("Load(\"\") should return error when no config found")
	}
	if !strings.Contains(err.Error(), "no config file found") {
		t.Errorf("Error message should mention no config found, got: %v", err)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/manifest.yaml")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-manifest.yaml")

	testConfig := `version: 1
defaults:
  registry: registry.example.com
images:
  app:
    path: images/app
    defaults:
      base_image:
        name: alpine
    versions:
      "1.0":
        custom_value: test
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}
	if config.Defaults.Registry != "registry.example.com" {
		t.Errorf("Registry = %s, want registry.example.com", config.Defaults.Registry)
	}
	if config.Defaults.BasePath != tmpDir {
		t.Errorf("BasePath = %s, want %s", config.Defaults.BasePath, tmpDir)
	}

	app, exists := config.Images["app"]
	if !exists {
		t.Fatal("Image 'app' not found")
	}
	if app.Path != "images/app" {
		t.Errorf("Image path = %s, want images/app", app.Path)
	}
	if app.Defaults == nil || app.Defaults.BaseImage == nil {
		t.Fatal("Image defaults or base image is nil")
	}
	if app.Defaults.BaseImage.Name != "alpine" {
		t.Errorf("Base image name = %s, want alpine", app.Defaults.BaseImage.Name)
	}
}

func TestLoadReader_Version1(t *testing.T) {
	testConfig := `version: 1
defaults:
  registry: test.io
images:
  myapp:
    versions:
      v1:
        base_image:
          name: ubuntu
        custom_key: custom_value
`
	reader := strings.NewReader(testConfig)
	config, err := loadReader(reader)
	if err != nil {
		t.Fatalf("loadReader() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}
	if config.Defaults.Registry != "test.io" {
		t.Errorf("Registry = %s, want test.io", config.Defaults.Registry)
	}

	myapp, exists := config.Images["myapp"]
	if !exists {
		t.Fatal("Image 'myapp' not found")
	}
	v1, exists := myapp.Versions["v1"]
	if !exists {
		t.Fatal("Version 'v1' not found")
	}
	if v1.BaseImage == nil || v1.BaseImage.Name != "ubuntu" {
		t.Error("Base image should be ubuntu")
	}
	if v1.Values["custom_key"] != "custom_value" {
		t.Errorf("custom_key = %v, want custom_value", v1.Values["custom_key"])
	}
}

func TestLoadReader_MissingVersion(t *testing.T) {
	testConfig := `defaults:
  registry: test.io
images:
  test: {}
`
	reader := strings.NewReader(testConfig)
	_, err := loadReader(reader)
	if err == nil {
		t.Error("loadReader() should return error for missing version")
	}
	if !strings.Contains(err.Error(), "version is required") {
		t.Errorf("Error should mention version is required, got: %v", err)
	}
}

func TestLoadReader_UnsupportedVersion(t *testing.T) {
	testConfig := `version: 99
images:
  test: {}
`
	reader := strings.NewReader(testConfig)
	_, err := loadReader(reader)
	if err == nil {
		t.Error("loadReader() should return error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported config version") {
		t.Errorf("Error should mention unsupported version, got: %v", err)
	}
}

func TestLoadReader_InvalidYAML(t *testing.T) {
	testConfig := `version: 1
images:
  test:
    - invalid
	  yaml: structure
`
	reader := strings.NewReader(testConfig)
	_, err := loadReader(reader)
	if err == nil {
		t.Error("loadReader() should return error for invalid YAML")
	}
}

func TestLoadReader_EmptyConfig(t *testing.T) {
	testConfig := ``
	reader := strings.NewReader(testConfig)
	_, err := loadReader(reader)
	if err == nil {
		t.Error("loadReader() should return error for empty config")
	}
}
