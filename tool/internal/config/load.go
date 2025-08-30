package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	if path == "-" {
		config, err := loadReader(os.Stdin)
		if err != nil {
			return nil, err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		config.Defaults.BasePath = cwd
		return config, nil
	}
	if path != "" {
		return loadFile(path)
	}
	for _, f := range [6]string{
		"images/manifest.yml",
		"images/manifest.yaml",
		"manifest.yml",
		"manifest.yaml",
		".manifest.yml",
		".manifest.yaml",
	} {
		m, err := loadFile(f)
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return m, err
	}

	return nil, fmt.Errorf("no config file found in any of the default locations")
}

func loadFile(file string) (*Config, error) {
	f, err := os.Open(file) // #nosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	config, err := loadReader(f)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of config file: %w", err)
	}
	config.Defaults.BasePath = filepath.Dir(absPath)

	return config, nil
}

func loadReader(fd io.Reader) (*Config, error) {
	data, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	var versioned struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &versioned); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if versioned.Version == 0 {
		return nil, fmt.Errorf("config version is required")
	}

	switch versioned.Version {
	case 1:
		var config Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse v1 config: %w", err)
		}
		return &config, nil
	default:
		return nil, fmt.Errorf("unsupported config version %d (only version 1 is supported)", versioned.Version)
	}
}
