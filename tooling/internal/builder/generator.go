package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateAll generates all Dockerfiles from the manifest
func GenerateAll(manifest Manifest) error {
	for imageName := range manifest {
		if err := GenerateImage(manifest, imageName); err != nil {
			return fmt.Errorf("generating %s: %w", imageName, err)
		}
	}
	return nil
}

// GenerateImage generates Dockerfiles for a specific image
func GenerateImage(manifest Manifest, imageName string) error {
	config, exists := manifest[imageName]
	if !exists {
		return fmt.Errorf("image %s not found in manifest", imageName)
	}

	fmt.Printf("Generating %s Dockerfiles\n", imageName)

	sourceDir := GetSourceDir(imageName)
	templateFiles := config.GetTemplateFiles()
	defaults := config.Defaults
	if defaults == nil {
		defaults = make(map[string]interface{})
	}

	for version, values := range config.Versions {
		fmt.Printf("- %s ", version)

		outputDir := GetOutputDir(imageName, version)

		// Clean and create the output directory
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("removing output directory %s: %w", outputDir, err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", outputDir, err)
		}

		// Merge values: defaults + version-specific values + version
		templateValues := make(map[string]interface{})
		for k, v := range defaults {
			templateValues[k] = v
		}
		if values != nil {
			for k, v := range values {
				templateValues[k] = v
			}
		}

		// Handle version override
		actualVersion := version
		if versionOverride, ok := templateValues["version_override"].(string); ok {
			actualVersion = versionOverride
		}
		templateValues["version"] = actualVersion

		// Create template data
		templateData := NewTemplateData(templateValues, imageName)
		templateData.Version = actualVersion

		// Process template files
		for _, templateFile := range templateFiles {
			templatePath := ResolveTemplateFile(sourceDir, templateFile)
			outputFilename := templateFile

			// If using .tmpl file, output without .tmpl extension
			if filepath.Ext(templatePath) == ".tmpl" {
				outputFilename = strings.TrimSuffix(templateFile, ".tmpl")
			}

			if err := WriteTemplateWithOutputName(templatePath, outputDir, outputFilename, templateData); err != nil {
				return fmt.Errorf("processing template %s: %w", templatePath, err)
			}
		}

		// Copy non-template files
		if err := CopyNonTemplateFiles(sourceDir, outputDir, templateFiles); err != nil {
			return fmt.Errorf("copying non-template files: %w", err)
		}

		fmt.Println("Done!")
	}

	return nil
}
