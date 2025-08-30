package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
	"github.com/mberwanger/dockerfiles/tool/internal/template"
)

func GenerateAll(cfg *config.Config) error {
	for imageName := range cfg.Images {
		log.Debugf("generating image '%s'", imageName)
		if err := GenerateImage(cfg, imageName); err != nil {
			return fmt.Errorf("generating %s: %w", imageName, err)
		}
	}
	return nil
}

func GenerateImage(cfg *config.Config, imageName string) error {
	image, exists := cfg.Images[imageName]
	if !exists {
		return fmt.Errorf("image %s not found in config", imageName)
	}

	var imagePath string
	if filepath.IsAbs(image.Path) {
		imagePath = image.Path
	} else {
		basePath := cfg.Defaults.BasePath
		if basePath == "" {
			return fmt.Errorf("base path not set in config")
		}
		imagePath = filepath.Join(basePath, image.Path)
	}

	sourceDir := filepath.Join(imagePath, "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", sourceDir)
	}

	if err := cleanupOrphanedVersions(imagePath, image.Versions); err != nil {
		return fmt.Errorf("cleaning up orphaned versions: %w", err)
	}

	imageDefaults := image.Defaults
	if imageDefaults == nil {
		imageDefaults = &config.ImageConfig{
			Values: make(map[string]interface{}),
		}
	}

	for versionName, versionConfig := range image.Versions {
		log.Debugf("  → version %s", versionName)

		if versionConfig == nil {
			versionConfig = &config.ImageConfig{
				Values: make(map[string]interface{}),
			}
		}

		mergedConfig := versionConfig.Merge(imageDefaults)
		mergedConfig.Values["version"] = versionName

		if _, hasRegistry := mergedConfig.Values["registry"]; !hasRegistry {
			mergedConfig.Values["registry"] = cfg.Defaults.Registry
		}

		outputDir := filepath.Join(imagePath, versionName)
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("removing output directory %s: %w", outputDir, err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", outputDir, err)
		}

		templateData := template.NewData(mergedConfig, imageName)

		templateFiles, err := discoverTemplateFiles(sourceDir)
		if err != nil {
			return fmt.Errorf("discovering template files: %w", err)
		}

		// Process template files
		for _, templateFile := range templateFiles {
			templatePath := filepath.Join(sourceDir, templateFile)
			outputFilename := strings.TrimSuffix(templateFile, ".tmpl")

			outputPath := filepath.Join(outputDir, outputFilename)
			outputFileDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputFileDir, 0755); err != nil {
				return fmt.Errorf("creating output directory %s: %w", outputFileDir, err)
			}

			if err := template.WriteFile(templatePath, outputPath, templateData); err != nil {
				return fmt.Errorf("processing template %s: %w", templateFile, err)
			}
		}

		if err := copyNonTemplateFiles(sourceDir, outputDir, templateFiles); err != nil {
			return fmt.Errorf("copying non-template files: %w", err)
		}
	}

	return nil
}

func discoverTemplateFiles(sourceDir string) ([]string, error) {
	var templateFiles []string

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".tmpl") {
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return fmt.Errorf("getting relative path for %s: %w", path, err)
			}
			templateFiles = append(templateFiles, relPath)
		}

		return nil
	})

	return templateFiles, err
}

func copyNonTemplateFiles(sourceDir, outputDir string, exclude []string) error {
	if sourceDir == "" || outputDir == "" {
		return fmt.Errorf("source and output directories cannot be empty")
	}

	excludeSet := make(map[string]bool)
	for _, file := range exclude {
		excludeSet[file] = true
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == sourceDir {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		if excludeSet[relPath] || excludeSet[filepath.Base(path)] {
			return nil
		}

		destPath := filepath.Join(outputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", path, err)
		}

		if err := os.WriteFile(destPath, content, info.Mode()); err != nil {
			return fmt.Errorf("writing file %s: %w", destPath, err)
		}

		return nil
	})
}

func cleanupOrphanedVersions(imagePath string, versions map[string]*config.ImageConfig) error {
	entries, err := os.ReadDir(imagePath)
	if err != nil {
		return fmt.Errorf("reading image directory: %w", err)
	}

	for _, entry := range entries {
		// Skip non-directories and the source directory
		if !entry.IsDir() || entry.Name() == "source" {
			continue
		}

		if _, exists := versions[entry.Name()]; !exists {
			orphanedPath := filepath.Join(imagePath, entry.Name())
			log.Infof("removing orphaned version directory: %s", orphanedPath)
			if err := os.RemoveAll(orphanedPath); err != nil {
				return fmt.Errorf("removing orphaned directory %s: %w", orphanedPath, err)
			}
		}
	}

	return nil
}
