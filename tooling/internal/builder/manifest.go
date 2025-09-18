package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GlobalConfig represents global configuration
type GlobalConfig struct {
	Registry string `yaml:"registry"`
}

// ManifestFile represents the complete manifest file structure
type ManifestFile struct {
	Config      GlobalConfig              `yaml:"config"`
	Dockerfiles map[string]ImageConfig   `yaml:"dockerfiles"`
}

// Manifest represents the structure of manifest.yaml
type Manifest map[string]ImageConfig

// ImageConfig represents configuration for a single image
type ImageConfig struct {
	Defaults      map[string]interface{} `yaml:"defaults"`
	Versions      map[string]Version     `yaml:"versions"`
	TemplateFiles []string               `yaml:"template_files"`
}

// Version represents a specific version configuration
type Version map[string]interface{}

// BaseImage represents either a string or object base image reference
type BaseImage struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// Global config instance
var globalConfig *GlobalConfig

// LoadManifest loads and parses the manifest.yaml file
func LoadManifest(path string) (Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var manifestFile ManifestFile
	if err := yaml.Unmarshal(data, &manifestFile); err != nil {
		return nil, fmt.Errorf("parsing manifest YAML: %w", err)
	}

	// Store global config for access by other functions
	globalConfig = &manifestFile.Config

	return manifestFile.Dockerfiles, nil
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *GlobalConfig {
	return globalConfig
}

// GetNamespace extracts the namespace from the registry
// Examples:
//   "ghcr.io/mberwanger" -> "mberwanger"
//   "docker.io/biograph" -> "biograph"
//   "registry.com/org/team" -> "team" (last component)
func GetNamespace() string {
	config := GetGlobalConfig()
	if config == nil || config.Registry == "" {
		return "biograph" // fallback for backwards compatibility
	}

	// Split by "/" and take the last component
	parts := strings.Split(config.Registry, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "biograph" // fallback
}

// GetTemplateFiles returns the list of template files for an image
func (ic *ImageConfig) GetTemplateFiles() []string {
	files := ic.TemplateFiles
	if files == nil {
		files = []string{}
	}

	// Add Dockerfile if not explicitly included
	hasDockerfile := false
	for _, file := range files {
		if file == "Dockerfile" || file == "Dockerfile.tmpl" || file == "Dockerfile.gotmpl" {
			hasDockerfile = true
			break
		}
	}
	if !hasDockerfile {
		files = append(files, "Dockerfile.tmpl")
	}

	return files
}

// GetSourceDir returns the source directory path for an image
func GetSourceDir(imageName string) string {
	// Check each subdirectory for the image
	subdirs := []string{"base", "lang", "app", "runtime", "util"}
	for _, subdir := range subdirs {
		path := filepath.Join("dockerfiles", subdir, imageName)
		if _, err := os.Stat(path); err == nil {
			return filepath.Join(path, "source")
		}
	}
	// Default to base for backwards compatibility
	return filepath.Join("dockerfiles/base", imageName, "source")
}

// GetOutputDir returns the output directory path for an image version
func GetOutputDir(imageName, version string) string {
	// Check each subdirectory for the image
	subdirs := []string{"base", "lang", "app", "runtime", "util"}
	for _, subdir := range subdirs {
		path := filepath.Join("dockerfiles", subdir, imageName)
		if _, err := os.Stat(path); err == nil {
			return filepath.Join(path, version)
		}
	}
	// Default to base for backwards compatibility
	return filepath.Join("dockerfiles/base", imageName, version)
}

// ResolveTemplateFile resolves the actual template file path, preferring .tmpl versions
func ResolveTemplateFile(sourceDir, filename string) string {
	// If the filename is explicitly .tmpl, use it directly
	if filepath.Ext(filename) == ".tmpl" {
		return filepath.Join(sourceDir, filename)
	}

	// Try .tmpl version first
	tmplPath := filepath.Join(sourceDir, filename+".tmpl")
	if _, err := os.Stat(tmplPath); err == nil {
		return tmplPath
	}

	// Fall back to original filename
	return filepath.Join(sourceDir, filename)
}