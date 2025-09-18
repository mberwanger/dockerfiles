package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// TemplateData holds all data passed to templates
type TemplateData struct {
	Values            map[string]interface{}
	rootPathIncluded  bool
	GenerationMessage string
	Version           string
}

// NewTemplateData creates a new template data instance
func NewTemplateData(values map[string]interface{}, imageName string) *TemplateData {
	return &TemplateData{
		Values:            values,
		rootPathIncluded:  false,
		GenerationMessage: generateMessage(imageName),
	}
}

// Get retrieves a value from the template data
func (td *TemplateData) Get(key string) interface{} {
	return td.Values[key]
}

// FromImage generates the FROM instruction with proper ROOT_PATH handling
func (td *TemplateData) FromImage(baseImage interface{}) string {
	var imageName, imageSource string

	switch v := baseImage.(type) {
	case string:
		imageName = v
		imageSource = ""
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			imageName = name
		}
		if source, ok := v["source"].(string); ok {
			imageSource = source
		}
	case map[interface{}]interface{}:
		// Handle YAML's default map type
		if name, ok := v["name"].(string); ok {
			imageName = name
		}
		if source, ok := v["source"].(string); ok {
			imageSource = source
		}
	case Version:
		// Handle builder.Version type
		if name, ok := v["name"].(string); ok {
			imageName = name
		}
		if source, ok := v["source"].(string); ok {
			imageSource = source
		}
	default:
		// If baseImage is a variable reference, look it up in Values
		if varName, ok := baseImage.(string); ok {
			if val, exists := td.Values[varName]; exists {
				return td.FromImage(val)
			}
		}
		imageName = fmt.Sprintf("%v", baseImage)
	}

	useRootPath := imageSource != "dockerhub" && !td.rootPathIncluded
	var imagePath string
	if useRootPath {
		imagePath = fmt.Sprintf("${REGISTRY}/%s", imageName)
	} else {
		imagePath = imageName
	}

	var result strings.Builder
	if useRootPath && !td.rootPathIncluded {
		config := GetGlobalConfig()
		registry := "us-docker.pkg.dev/biograph-artifact-registry/docker" // default fallback
		if config != nil && config.Registry != "" {
			registry = config.Registry
		}
		result.WriteString(fmt.Sprintf("ARG REGISTRY=%s\n", registry))
		td.rootPathIncluded = true
	}
	result.WriteString(fmt.Sprintf("FROM %s", imagePath))

	return result.String()
}

// generateMessage creates the generation warning message
func generateMessage(imageName string) string {
	return fmt.Sprintf(`# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate %s'`, imageName)
}

// TemplateFuncs returns the function map for templates
func (td *TemplateData) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"generation_message": func() string { return td.GenerationMessage },
		"from_image": func(arg interface{}) string {
			// If arg is a string that matches a key in Values, resolve it
			if argStr, ok := arg.(string); ok {
				if val, exists := td.Values[argStr]; exists {
					return td.FromImage(val)
				}
			}
			return td.FromImage(arg)
		},
		// Allow direct access to all values
		"get": td.Get,
	}
}

// convertERBToGoTemplate converts ERB-style templates to Go template syntax
func convertERBToGoTemplate(content string) string {
	// Convert <%= expression %> to {{ expression }}
	content = regexp.MustCompile(`<%=\s*([^%]+?)\s*%>`).ReplaceAllString(content, "{{ $1 }}")

	// Handle function calls with string arguments - replace function_name('arg') with (function_name "arg")
	content = regexp.MustCompile(`\{\{\s*(\w+)\('([^']*)'\)\s*\}\}`).ReplaceAllString(content, `{{ $1 "$2" }}`)

	// Handle function calls with variable arguments - replace function_name(var) with (function_name (index .Values "var"))
	content = regexp.MustCompile(`\{\{\s*(\w+)\((\w+)\)\s*\}\}`).ReplaceAllString(content, `{{ $1 (index .Values "$2") }}`)

	// Handle direct variable access - replace var with (index .Values "var")
	content = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`).ReplaceAllString(content, `{{ index .Values "$1" }}`)

	return content
}

// isGoTemplate determines if a file uses Go template syntax instead of ERB
func isGoTemplate(content string) bool {
	// Check for Go template syntax patterns
	goPatterns := []string{
		`\{\{\s*\.`,        // {{ .Variable }}
		`\{\{\s*range\s+`,  // {{ range }}
		`\{\{\s*if\s+`,     // {{ if }}
		`\{\{\s*with\s+`,   // {{ with }}
		`\{\{\s*end\s*\}\}`, // {{ end }}
		`\{\{\s*\$`,        // {{ $var }}
	}

	for _, pattern := range goPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}

	// Check for ERB patterns
	erbPatterns := []string{
		`<%=`,  // <%= ... %>
		`<%\s+`, // <% ... %>
	}

	for _, pattern := range erbPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return false
		}
	}

	// If neither Go nor ERB patterns found, assume Go template
	return true
}

// RenderTemplate renders a template file with the given data
func RenderTemplate(templatePath string, data *TemplateData) (string, error) {
	// Read template content
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template file %s: %w", templatePath, err)
	}

	contentStr := string(content)
	var goTemplateContent string

	// Determine if this is already a Go template or needs conversion from ERB
	if isGoTemplate(contentStr) {
		goTemplateContent = contentStr
	} else {
		// Convert ERB syntax to Go template syntax
		goTemplateContent = convertERBToGoTemplate(contentStr)
	}

	// Create template
	tmpl := template.New(filepath.Base(templatePath))

	// Add all values as template variables and functions for direct access
	funcs := data.TemplateFuncs()
	for key, value := range data.Values {
		switch v := value.(type) {
		case string:
			funcs[key] = func(val string) func() string {
				return func() string { return val }
			}(v)
		case int:
			funcs[key] = func(val int) func() int {
				return func() int { return val }
			}(v)
		case float64:
			funcs[key] = func(val float64) func() float64 {
				return func() float64 { return val }
			}(v)
		case map[string]interface{}:
			funcs[key] = func(val map[string]interface{}) func() map[string]interface{} {
				return func() map[string]interface{} { return val }
			}(v)
		default:
			funcs[key] = func(val interface{}) func() interface{} {
				return func() interface{} { return val }
			}(v)
		}
	}

	tmpl = tmpl.Funcs(funcs)

	// Parse template
	tmpl, err = tmpl.Parse(goTemplateContent)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	// Render template - pass both data and values as context
	templateContext := struct {
		*TemplateData
		Values map[string]interface{}
	}{
		TemplateData: data,
		Values:       data.Values,
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, templateContext); err != nil {
		return "", fmt.Errorf("executing template %s: %w", templatePath, err)
	}

	return result.String(), nil
}

// WriteTemplate renders and writes a template to the output directory
func WriteTemplate(templatePath, outputDir string, data *TemplateData) error {
	content, err := RenderTemplate(templatePath, data)
	if err != nil {
		return err
	}

	filename := filepath.Base(templatePath)
	outputPath := filepath.Join(outputDir, filename)

	return ioutil.WriteFile(outputPath, []byte(content), 0644)
}

// WriteTemplateWithOutputName renders and writes a template with a custom output filename
func WriteTemplateWithOutputName(templatePath, outputDir, outputFilename string, data *TemplateData) error {
	content, err := RenderTemplate(templatePath, data)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, outputFilename)
	return ioutil.WriteFile(outputPath, []byte(content), 0644)
}

// CopyNonTemplateFiles copies non-template files from source directory to output
func CopyNonTemplateFiles(sourceDir, outputDir string, templateFiles []string) error {
	// Create a map of template files for quick lookup
	templateSet := make(map[string]bool)
	for _, file := range templateFiles {
		templateSet[file] = true
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's the source directory itself
		if path == sourceDir {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip template files
		if templateSet[filepath.Base(path)] {
			return nil
		}

		// Create destination path
		destPath := filepath.Join(outputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(destPath, content, info.Mode())
	})
}