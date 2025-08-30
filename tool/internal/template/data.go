package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

type Data struct {
	Values            map[string]interface{}
	rootPathIncluded  bool
	generationMessage string
}

func NewData(mergedConfig *config.ImageConfig, imageName string) *Data {
	data := make(map[string]interface{})

	if mergedConfig.BaseImage != nil {
		data["base_image"] = mergedConfig.BaseImage
	}

	for k, v := range mergedConfig.Values {
		data[k] = v
	}

	data["image_name"] = imageName

	return &Data{
		Values:            data,
		rootPathIncluded:  false,
		generationMessage: generateMessage(imageName),
	}
}

func (d *Data) functions() template.FuncMap {
	return template.FuncMap{
		"generation_message": func() string { return d.generationMessage },
		"from_image": func(arg interface{}) string {
			if argStr, ok := arg.(string); ok {
				if val, exists := d.Values[argStr]; exists {
					return d.fromImage(val)
				}
			}
			return d.fromImage(arg)
		},
		"get": d.get,
	}
}

func (d *Data) get(key string) interface{} {
	return d.Values[key]
}

func (d *Data) fromImage(baseImage interface{}) string {
	var imageName, imageSource string

	switch v := baseImage.(type) {
	case string:
		if val, exists := d.Values[v]; exists {
			return d.fromImage(val)
		}
		imageName = v
		imageSource = ""
	case *config.BaseImage:
		imageName = v.Name
		imageSource = v.Source
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			imageName = name
		}
		if source, ok := v["source"].(string); ok {
			imageSource = source
		}
	default:
		imageName = fmt.Sprintf("%v", baseImage)
	}

	if imageSource == "dockerhub" {
		return fmt.Sprintf("FROM %s", imageName)
	}

	needsRegistryArg := imageSource != "dockerhub" && !d.rootPathIncluded
	needsRegistryPath := imageSource != "dockerhub"

	var imagePath string
	if needsRegistryPath {
		imagePath = fmt.Sprintf("${REGISTRY}/%s", imageName)
	} else {
		imagePath = imageName
	}

	var result strings.Builder
	if needsRegistryArg {
		registryVal, exists := d.Values["registry"]
		if !exists {
			return fmt.Sprintf("# ERROR: registry not set in config\nFROM %s", imagePath)
		}

		registry, ok := registryVal.(string)
		if !ok {
			return fmt.Sprintf("# ERROR: registry is not a string\nFROM %s", imagePath)
		}
		result.WriteString(fmt.Sprintf("ARG REGISTRY=%s\n", registry))

		d.rootPathIncluded = true
	}
	result.WriteString(fmt.Sprintf("FROM %s", imagePath))

	return result.String()
}

func generateMessage(imageName string) string {
	return fmt.Sprintf(`# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image %s
#
# Or regenerate all images with:
#   go run tool/main.go generate all`, imageName)
}
