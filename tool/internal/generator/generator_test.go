package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

func TestGenerateAll(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			BasePath: tmpDir,
			Registry: "test.io",
		},
		Images: map[string]config.Image{
			"app1": {
				Path: "images/app1",
				Versions: map[string]*config.ImageConfig{
					"v1": {
						Values: map[string]interface{}{},
					},
				},
			},
			"app2": {
				Path: "images/app2",
				Versions: map[string]*config.ImageConfig{
					"v1": {
						Values: map[string]interface{}{},
					},
				},
			},
		},
	}

	// Create source directories
	for imageName := range cfg.Images {
		sourceDir := filepath.Join(tmpDir, "images", imageName, "source")
		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("Failed to create source directory: %v", err)
		}
		// Create a simple template file
		tmplPath := filepath.Join(sourceDir, "Dockerfile.tmpl")
		if err := os.WriteFile(tmplPath, []byte("FROM alpine\n"), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}
	}

	if err := GenerateAll(cfg); err != nil {
		t.Fatalf("GenerateAll() error = %v", err)
	}

	// Verify both images were generated
	for imageName := range cfg.Images {
		outputPath := filepath.Join(tmpDir, "images", imageName, "v1", "Dockerfile")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s does not exist", outputPath)
		}
	}
}

func TestGenerateAll_ImageNotFound(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			BasePath: "/tmp",
		},
		Images: map[string]config.Image{
			"nonexistent": {
				Path: "does/not/exist",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
		},
	}

	err := GenerateAll(cfg)
	if err == nil {
		t.Error("GenerateAll() should return error for nonexistent source directory")
	}
}

func TestGenerateImage(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Version: 1,
		Defaults: config.Defaults{
			BasePath: tmpDir,
			Registry: "registry.test.io",
		},
		Images: map[string]config.Image{
			"myapp": {
				Path: "images/myapp",
				Defaults: &config.ImageConfig{
					Values: map[string]interface{}{
						"default_key": "default_value",
					},
				},
				Versions: map[string]*config.ImageConfig{
					"v1.0": {
						Values: map[string]interface{}{
							"custom_key": "custom_value",
						},
					},
					"v2.0": {
						Values: map[string]interface{}{
							"custom_key": "custom_value_v2",
						},
					},
				},
			},
		},
	}

	// Create source directory and template
	sourceDir := filepath.Join(tmpDir, "images/myapp/source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	tmplContent := "FROM alpine\n# Version: {{version}}\n"
	tmplPath := filepath.Join(sourceDir, "Dockerfile.tmpl")
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	if err := GenerateImage(cfg, "myapp"); err != nil {
		t.Fatalf("GenerateImage() error = %v", err)
	}

	// Verify both versions were generated
	for version := range cfg.Images["myapp"].Versions {
		outputPath := filepath.Join(tmpDir, "images/myapp", version, "Dockerfile")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s does not exist", outputPath)
		}
	}
}

func TestGenerateImage_ImageNotFound(t *testing.T) {
	cfg := &config.Config{
		Images: map[string]config.Image{},
	}

	err := GenerateImage(cfg, "nonexistent")
	if err == nil {
		t.Error("GenerateImage() should return error for nonexistent image")
	}
	if err != nil && err.Error() != "image nonexistent not found in config" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestGenerateImage_NoBasePath(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{
			BasePath: "",
		},
		Images: map[string]config.Image{
			"myapp": {
				Path: "images/myapp",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
		},
	}

	err := GenerateImage(cfg, "myapp")
	if err == nil {
		t.Error("GenerateImage() should return error when base path is not set")
	}
}

func TestGenerateImage_SourceDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Defaults: config.Defaults{
			BasePath: tmpDir,
		},
		Images: map[string]config.Image{
			"myapp": {
				Path: "images/myapp",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
		},
	}

	err := GenerateImage(cfg, "myapp")
	if err == nil {
		t.Error("GenerateImage() should return error when source directory doesn't exist")
	}
}

func TestGenerateImage_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	absoluteImagePath := filepath.Join(tmpDir, "absolute/path/myapp")

	cfg := &config.Config{
		Defaults: config.Defaults{
			BasePath: tmpDir,
			Registry: "test.io",
		},
		Images: map[string]config.Image{
			"myapp": {
				Path: absoluteImagePath,
				Versions: map[string]*config.ImageConfig{
					"v1": {
						Values: map[string]interface{}{},
					},
				},
			},
		},
	}

	// Create source directory
	sourceDir := filepath.Join(absoluteImagePath, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	tmplPath := filepath.Join(sourceDir, "Dockerfile.tmpl")
	if err := os.WriteFile(tmplPath, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	if err := GenerateImage(cfg, "myapp"); err != nil {
		t.Fatalf("GenerateImage() error = %v", err)
	}

	outputPath := filepath.Join(absoluteImagePath, "v1", "Dockerfile")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s does not exist", outputPath)
	}
}

func TestDiscoverTemplateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file structure
	files := map[string]string{
		"Dockerfile.tmpl":             "FROM alpine",
		"config.yaml.tmpl":            "key: value",
		"nested/deep/file.tmpl":       "content",
		"regular.txt":                 "not a template",
		"nested/another-regular.json": "{}",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(tmpDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	templateFiles, err := discoverTemplateFiles(tmpDir)
	if err != nil {
		t.Fatalf("discoverTemplateFiles() error = %v", err)
	}

	expectedTemplates := []string{
		"Dockerfile.tmpl",
		"config.yaml.tmpl",
		"nested/deep/file.tmpl",
	}

	if len(templateFiles) != len(expectedTemplates) {
		t.Errorf("Expected %d template files, got %d", len(expectedTemplates), len(templateFiles))
	}

	// Convert to map for easier comparison
	foundMap := make(map[string]bool)
	for _, f := range templateFiles {
		foundMap[filepath.ToSlash(f)] = true
	}

	for _, expected := range expectedTemplates {
		if !foundMap[expected] {
			t.Errorf("Expected template file %s not found", expected)
		}
	}
}

func TestDiscoverTemplateFiles_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	templateFiles, err := discoverTemplateFiles(tmpDir)
	if err != nil {
		t.Fatalf("discoverTemplateFiles() error = %v", err)
	}

	if len(templateFiles) != 0 {
		t.Errorf("Expected no template files, got %d", len(templateFiles))
	}
}

func TestDiscoverTemplateFiles_NonexistentDir(t *testing.T) {
	_, err := discoverTemplateFiles("/nonexistent/directory")
	if err == nil {
		t.Error("discoverTemplateFiles() should return error for nonexistent directory")
	}
}

func TestCopyNonTemplateFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"regular.txt":          "content",
		"data.json":            "{}",
		"script.sh":            "#!/bin/bash",
		"template.tmpl":        "should be excluded",
		"nested/deep/file.txt": "nested content",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(sourceDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	exclude := []string{"template.tmpl"}
	if err := copyNonTemplateFiles(sourceDir, outputDir, exclude); err != nil {
		t.Fatalf("copyNonTemplateFiles() error = %v", err)
	}

	// Verify copied files
	copiedFiles := []string{
		"regular.txt",
		"data.json",
		"script.sh",
		"nested/deep/file.txt",
	}

	for _, file := range copiedFiles {
		destPath := filepath.Join(outputDir, file)
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not copied", file)
		} else {
			// Verify content
			content, err := os.ReadFile(destPath)
			if err != nil {
				t.Errorf("Failed to read copied file %s: %v", file, err)
			}
			if string(content) != files[file] {
				t.Errorf("File %s content mismatch: got %s, want %s", file, content, files[file])
			}
		}
	}

	// Verify excluded file was not copied
	excludedPath := filepath.Join(outputDir, "template.tmpl")
	if _, err := os.Stat(excludedPath); err == nil {
		t.Error("Excluded file template.tmpl should not be copied")
	}
}

func TestCopyNonTemplateFiles_EmptyDirs(t *testing.T) {
	err := copyNonTemplateFiles("", "/tmp/output", nil)
	if err == nil {
		t.Error("copyNonTemplateFiles() should return error for empty source dir")
	}

	err = copyNonTemplateFiles("/tmp/source", "", nil)
	if err == nil {
		t.Error("copyNonTemplateFiles() should return error for empty output dir")
	}
}

func TestCopyNonTemplateFiles_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create a file with specific permissions
	executablePath := filepath.Join(sourceDir, "script.sh")
	if err := os.WriteFile(executablePath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to write executable: %v", err)
	}

	if err := copyNonTemplateFiles(sourceDir, outputDir, nil); err != nil {
		t.Fatalf("copyNonTemplateFiles() error = %v", err)
	}

	destPath := filepath.Join(outputDir, "script.sh")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("Failed to stat copied file: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("File permissions not preserved: got %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestCleanupOrphanedVersions(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "myapp")

	// Create version directories
	dirs := []string{
		"source",      // Should be skipped
		"v1.0",        // Valid version
		"v2.0",        // Valid version
		"v3.0-orphan", // Orphaned version
		"old-version", // Orphaned version
	}

	for _, dir := range dirs {
		path := filepath.Join(imagePath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		// Create a dummy file to ensure directory is not empty
		dummyFile := filepath.Join(path, "dummy.txt")
		if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to write dummy file: %v", err)
		}
	}

	// Create a regular file (not a directory)
	regularFile := filepath.Join(imagePath, "README.md")
	if err := os.WriteFile(regularFile, []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to write regular file: %v", err)
	}

	versions := map[string]*config.ImageConfig{
		"v1.0": {Values: map[string]interface{}{}},
		"v2.0": {Values: map[string]interface{}{}},
	}

	if err := cleanupOrphanedVersions(imagePath, versions); err != nil {
		t.Fatalf("cleanupOrphanedVersions() error = %v", err)
	}

	// Verify valid versions still exist
	for version := range versions {
		path := filepath.Join(imagePath, version)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Valid version directory %s was incorrectly removed", version)
		}
	}

	// Verify source directory still exists
	sourcePath := filepath.Join(imagePath, "source")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Error("Source directory was incorrectly removed")
	}

	// Verify orphaned versions were removed
	orphanedVersions := []string{"v3.0-orphan", "old-version"}
	for _, orphan := range orphanedVersions {
		path := filepath.Join(imagePath, orphan)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("Orphaned version directory %s was not removed", orphan)
		}
	}

	// Verify regular file still exists
	if _, err := os.Stat(regularFile); os.IsNotExist(err) {
		t.Error("Regular file was incorrectly removed")
	}
}

func TestCleanupOrphanedVersions_NonexistentDir(t *testing.T) {
	versions := map[string]*config.ImageConfig{
		"v1": {},
	}

	err := cleanupOrphanedVersions("/nonexistent/path", versions)
	if err == nil {
		t.Error("cleanupOrphanedVersions() should return error for nonexistent directory")
	}
}

func TestCleanupOrphanedVersions_EmptyImagePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty image path
	imagePath := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	versions := map[string]*config.ImageConfig{
		"v1": {},
	}

	// Should not error on empty directory
	if err := cleanupOrphanedVersions(imagePath, versions); err != nil {
		t.Fatalf("cleanupOrphanedVersions() error = %v", err)
	}
}
