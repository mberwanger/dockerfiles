package config

type Config struct {
	Version  int              `yaml:"version" json:"version"`
	Defaults Defaults         `yaml:"defaults" json:"defaults"`
	Images   map[string]Image `yaml:"images" json:"images"`
}

type Defaults struct {
	BasePath string `yaml:"-" json:"-"`
	Registry string `yaml:"registry,omitempty" json:"registry,omitempty"`
}

type Image struct {
	Path     string                  `yaml:"path,omitempty" json:"path,omitempty"`
	Defaults *ImageConfig            `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Versions map[string]*ImageConfig `yaml:"versions" json:"versions"`
}

type ImageConfig struct {
	BaseImage *BaseImage             `yaml:"base_image,omitempty" json:"base_image,omitempty"`
	Values    map[string]interface{} `yaml:"-" json:"-"`
}

type BaseImage struct {
	Name   string `yaml:"name" json:"name"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
}

func (ic *ImageConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First unmarshal into a raw map
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	ic.Values = make(map[string]interface{})

	// Extract base_image if present
	if baseImageRaw, ok := raw["base_image"]; ok {
		if baseImageMap, ok := baseImageRaw.(map[string]interface{}); ok {
			ic.BaseImage = &BaseImage{}
			if name, ok := baseImageMap["name"].(string); ok {
				ic.BaseImage.Name = name
			}
			if source, ok := baseImageMap["source"].(string); ok {
				ic.BaseImage.Source = source
			}
		}
		delete(raw, "base_image")
	}

	for k, v := range raw {
		ic.Values[k] = v
	}

	return nil
}

func (ic *ImageConfig) MarshalYAML() (interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range ic.Values {
		result[k] = v
	}

	if ic.BaseImage != nil {
		result["base_image"] = ic.BaseImage
	}

	return result, nil
}

func (ic *ImageConfig) Merge(defaults *ImageConfig) *ImageConfig {
	if defaults == nil {
		return ic.deepCopy()
	}
	if ic == nil {
		return defaults.deepCopy()
	}

	result := &ImageConfig{
		Values: make(map[string]interface{}),
	}

	if ic.BaseImage != nil {
		result.BaseImage = &BaseImage{
			Name:   ic.BaseImage.Name,
			Source: ic.BaseImage.Source,
		}
	} else if defaults.BaseImage != nil {
		result.BaseImage = &BaseImage{
			Name:   defaults.BaseImage.Name,
			Source: defaults.BaseImage.Source,
		}
	}

	for k, val := range defaults.Values {
		result.Values[k] = deepCopyValue(val)
	}

	mergeInto(result.Values, ic.Values)

	return result
}

func mergeInto(dest, source map[string]interface{}) {
	for k, srcVal := range source {
		if destVal, exists := dest[k]; exists {
			if destMap, destOk := destVal.(map[string]interface{}); destOk {
				if srcMap, srcOk := srcVal.(map[string]interface{}); srcOk {
					mergeInto(destMap, srcMap)
					continue
				}
			}
		}
		dest[k] = deepCopyValue(srcVal)
	}
}

func (ic *ImageConfig) deepCopy() *ImageConfig {
	if ic == nil {
		return nil
	}

	result := &ImageConfig{
		Values: make(map[string]interface{}),
	}

	if ic.BaseImage != nil {
		result.BaseImage = &BaseImage{
			Name:   ic.BaseImage.Name,
			Source: ic.BaseImage.Source,
		}
	}

	for k, v := range ic.Values {
		result.Values[k] = deepCopyValue(v)
	}

	return result
}

func deepCopyValue(val interface{}) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range v {
			result[k] = deepCopyValue(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = deepCopyValue(item)
		}
		return result
	default:
		return v
	}
}
