package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func WriteFile(templatePath, outputPath string, data *Data) error {
	content, err := render(templatePath, data)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func render(templatePath string, data *Data) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	tmpl := template.New(filepath.Base(templatePath))

	fn := data.functions()
	for key, value := range data.Values {
		switch v := value.(type) {
		case string:
			fn[key] = func(val string) func() string {
				return func() string { return val }
			}(v)
		case int:
			fn[key] = func(val int) func() int {
				return func() int { return val }
			}(v)
		case float64:
			fn[key] = func(val float64) func() float64 {
				return func() float64 { return val }
			}(v)
		case map[string]interface{}:
			fn[key] = func(val map[string]interface{}) func() map[string]interface{} {
				return func() map[string]interface{} { return val }
			}(v)
		default:
			fn[key] = func(val interface{}) func() interface{} {
				return func() interface{} { return val }
			}(v)
		}
	}

	tmpl = tmpl.Funcs(fn)
	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	templateContext := struct {
		*Data
		Values map[string]interface{}
	}{
		Data:   data,
		Values: data.Values,
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, templateContext); err != nil {
		return "", fmt.Errorf("executing template %s: %w", templatePath, err)
	}

	return result.String(), nil
}
