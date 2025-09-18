package builder

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

// DockerContext represents a Docker build context with its dependencies
type DockerContext struct {
	Path         string   `json:"path"`
	Dependencies []string `json:"dependencies"`
}

// DependencyBatch represents a batch of docker contexts that can be built in parallel
type DependencyBatch struct {
	Batch       int              `json:"batch"`
	Dockerfiles []string         `json:"dockerfiles"`
	Contexts    []*DockerContext `json:"contexts,omitempty"`
}

// GenerateDependencyIndex creates a dependency index for CI builds
func GenerateDependencyIndex(projectDir string) ([]*DependencyBatch, error) {
	// Find all Dockerfiles
	dockerfiles, err := findDockerfiles(projectDir)
	if err != nil {
		return nil, fmt.Errorf("finding dockerfiles: %w", err)
	}

	// Parse dependencies for each dockerfile
	contexts := make([]*DockerContext, 0, len(dockerfiles))
	for _, dockerfilePath := range dockerfiles {
		deps, err := extractDependencies(dockerfilePath)
		if err != nil {
			return nil, fmt.Errorf("extracting dependencies from %s: %w", dockerfilePath, err)
		}

		// Convert absolute path to relative path from project directory
		relPath, err := filepath.Rel(projectDir, filepath.Dir(dockerfilePath))
		if err != nil {
			return nil, fmt.Errorf("getting relative path for %s: %w", dockerfilePath, err)
		}

		contexts = append(contexts, &DockerContext{
			Path:         relPath,
			Dependencies: deps,
		})
	}

	// Group contexts by dependency resolution
	batches := resolveDependencies(contexts)

	return batches, nil
}

// findDockerfiles finds all Dockerfiles in the project, excluding templates and CI files
func findDockerfiles(projectDir string) ([]string, error) {
	var dockerfiles []string

	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a Dockerfile
		if info.Name() != "Dockerfile" {
			return nil
		}

		// Skip source directories
		if strings.Contains(path, "source") {
			return nil
		}

		// Skip backup directory
		if strings.Contains(path, "backup") {
			return nil
		}

		// Skip CI directory
		if strings.Contains(path, filepath.Join(projectDir, "ci")) {
			return nil
		}

		// Skip deprecated directories
		if strings.Contains(path, "deprecated") {
			return nil
		}

		dockerfiles = append(dockerfiles, path)
		return nil
	})

	return dockerfiles, err
}

// extractDependencies extracts FROM dependencies from a Dockerfile
func extractDependencies(dockerfilePath string) ([]string, error) {
	content, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return nil, err
	}

	var dependencies []string
	fromRegex := regexp.MustCompile(`^FROM\s+(.+)`)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := fromRegex.FindStringSubmatch(line); matches != nil {
			dep := cleanDependency(matches[1])
			if dep != "" {
				dependencies = append(dependencies, dep)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Filter to only internal dependencies (using namespace derived from root_path)
	namespace := GetNamespace()

	var filtered []string
	for _, dep := range dependencies {
		if strings.HasPrefix(dep, namespace) {
			filtered = append(filtered, dep)
		}
	}

	return filtered, nil
}

// cleanDependency cleans up a FROM dependency string
func cleanDependency(dependency string) string {
	// Replace ${REGISTRY} and ${ROOT_PATH} (legacy) with the namespace derived from registry
	namespace := GetNamespace()
	dependency = strings.Replace(dependency, "${REGISTRY}", namespace, -1)
	dependency = strings.Replace(dependency, "${ROOT_PATH}", namespace, -1) // legacy support

	// Remove AS clause
	dependency = regexp.MustCompile(`\s+AS\s+.+`).ReplaceAllString(dependency, "")

	return strings.TrimSpace(dependency)
}

// resolveDependencies groups docker contexts into batches based on dependency resolution
func resolveDependencies(contexts []*DockerContext) []*DependencyBatch {
	var batches []*DependencyBatch
	satisfiedDependencies := make(map[string]bool)
	remainingContexts := make([]*DockerContext, len(contexts))
	copy(remainingContexts, contexts)

	batchNum := 0
	for len(remainingContexts) > 0 {
		var currentBatch []*DockerContext

		// Find contexts that can be built (all dependencies satisfied)
		var newRemaining []*DockerContext
		for _, ctx := range remainingContexts {
			canBuild := true
			for _, dep := range ctx.Dependencies {
				if !satisfiedDependencies[dep] {
					canBuild = false
					break
				}
			}

			if canBuild {
				currentBatch = append(currentBatch, ctx)
			} else {
				newRemaining = append(newRemaining, ctx)
			}
		}

		// If no progress was made, we have a circular dependency
		if len(currentBatch) == 0 && len(newRemaining) > 0 {
			return nil // Could return error here
		}

		// Create batch
		if len(currentBatch) > 0 {
			dockerfiles := make([]string, len(currentBatch))
			for i, ctx := range currentBatch {
				dockerfiles[i] = ctx.Path
				pathParts := strings.Split(ctx.Path, "/")
				if len(pathParts) >= 3 {
					namespace := GetNamespace()
					imageName := fmt.Sprintf("%s/%s:%s", namespace, pathParts[len(pathParts)-2], pathParts[len(pathParts)-1])
					satisfiedDependencies[imageName] = true
				}
			}

			// Sort dockerfiles for consistency
			sort.Strings(dockerfiles)

			batches = append(batches, &DependencyBatch{
				Batch:       batchNum,
				Dockerfiles: dockerfiles,
				Contexts:    currentBatch,
			})
			batchNum++
		}

		remainingContexts = newRemaining
	}

	return batches
}

// WorkflowJob represents a single job in the GitHub Actions workflow
type WorkflowJob struct {
	Name           string   `yaml:"name"`
	JobID          string   `yaml:"-"`
	RunsOn         string   `yaml:"runs-on"`
	Needs          []string `yaml:"needs,omitempty"`
	DockerfilePath string   `yaml:"-"`
	ImageName      string   `yaml:"-"`
	ImageTag       string   `yaml:"-"`
	BatchNumber    int      `yaml:"-"`
}

// WorkflowData represents the complete workflow structure
type WorkflowData struct {
	Jobs []WorkflowJob
}

// GenerateWorkflow creates a complete GitHub Actions workflow from dependency batches
func GenerateWorkflow(projectDir string, outputPath string) error {
	// Generate dependency batches
	batches, err := GenerateDependencyIndex(projectDir)
	if err != nil {
		return fmt.Errorf("generating dependency index: %w", err)
	}

	// Convert batches to workflow jobs
	var jobs []WorkflowJob
	var allJobIDs [][]string // Track job IDs per batch for dependencies

	for _, batch := range batches {
		var batchJobIDs []string

		for _, dockerfilePath := range batch.Dockerfiles {
			// Generate job ID from dockerfile path
			jobID := pathToJobID(dockerfilePath)

			// Generate image info
			imageName, imageTag := pathToImageInfo(dockerfilePath)

			// Determine dependencies (needs all jobs from previous batches)
			var needs []string
			if batch.Batch > 0 {
				// Depend on all jobs from previous batch
				for i := 0; i < batch.Batch; i++ {
					needs = append(needs, allJobIDs[i]...)
				}
			}

			job := WorkflowJob{
				Name:           fmt.Sprintf("Build %s:%s", imageName, imageTag),
				JobID:          jobID,
				RunsOn:         "ubuntu-latest",
				Needs:          needs,
				DockerfilePath: dockerfilePath,
				ImageName:      imageName,
				ImageTag:       imageTag,
				BatchNumber:    batch.Batch,
			}

			jobs = append(jobs, job)
			batchJobIDs = append(batchJobIDs, jobID)
		}

		allJobIDs = append(allJobIDs, batchJobIDs)
	}

	// Generate workflow file
	workflowData := WorkflowData{Jobs: jobs}
	return writeWorkflowFile(workflowData, outputPath)
}

// pathToJobID converts a dockerfile path to a valid GitHub Actions job ID
func pathToJobID(dockerfilePath string) string {
	// Remove dockerfiles/ prefix and convert to job ID format
	path := strings.TrimPrefix(dockerfilePath, "dockerfiles/")

	// Replace slashes and dots with hyphens, ensure valid job ID
	jobID := strings.ReplaceAll(path, "/", "-")
	jobID = strings.ReplaceAll(jobID, ".", "-")
	jobID = strings.ToLower(jobID)

	// Ensure it starts with letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(jobID) {
		jobID = "build-" + jobID
	}

	return jobID
}

// pathToImageInfo extracts image name and tag from dockerfile path
func pathToImageInfo(dockerfilePath string) (string, string) {
	// Remove dockerfiles/ prefix
	path := strings.TrimPrefix(dockerfilePath, "dockerfiles/")

	// Split path parts
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "unknown", "unknown"
	}

	// Extract image name and version
	// e.g., base/tini/v0.19.0 -> tini, v0.19.0
	imageName := parts[len(parts)-2]
	version := parts[len(parts)-1]

	return imageName, version
}

// writeWorkflowFile generates and writes the GitHub Actions workflow file
func writeWorkflowFile(data WorkflowData, outputPath string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create workflow template
	tmpl := template.Must(template.New("workflow").Parse(workflowTemplate))

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating workflow file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

// workflowTemplate is the GitHub Actions workflow template
const workflowTemplate = `name: Build Docker Images (Generated)

on:
  pull_request:
    branches: [ master ]
  push:
    branches: [ master ]
  schedule:
    # Run daily at 2 AM UTC to check for updates
    - cron: '0 2 * * *'
  workflow_dispatch:
    inputs:
      force_push:
        description: 'Force push images even on PR'
        required: false
        default: false
        type: boolean

env:
  REGISTRY: ghcr.io

permissions:
  contents: read
  packages: write

jobs:{{range .Jobs}}
  {{.JobID}}:
    name: "{{.Name}}"
    runs-on: {{.RunsOn}}{{if .Needs}}
    needs: [{{range $i, $need := .Needs}}{{if $i}}, {{end}}{{$need}}{{end}}]{{end}}
    steps:
      - name: Checkout
        uses: actions/checkout@v5

      - name: Build {{.ImageName}}:{{.ImageTag}}
        uses: ./.github/actions/build-dockerfile
        with:
          dockerfile_path: {{.DockerfilePath}}
          image_name: {{.ImageName}}
          image_tag: {{.ImageTag}}
          force_push: ${{ "{{" }} inputs.force_push || 'false' {{ "}}" }}
          registry: ${{ "{{" }} env.REGISTRY {{ "}}" }}
          registry_username: ${{ "{{" }} github.actor {{ "}}" }}
          registry_password: ${{ "{{" }} secrets.GITHUB_TOKEN {{ "}}" }}
          root_path: ${{ "{{" }} env.REGISTRY {{ "}}" }}/${{ "{{" }} github.repository_owner {{ "}}" }}
          is_main_branch: ${{ "{{" }} github.ref == 'refs/heads/master' {{ "}}" }}
{{end}}

  # Notify about build results
  notify:
    needs: [{{range $i, $job := .Jobs}}{{if $i}}, {{end}}{{$job.JobID}}{{end}}]
    if: always() && github.event_name == 'push' && github.ref == 'refs/heads/master'
    runs-on: ubuntu-latest
    steps:
      - name: Notify on success
        if: ${{ "{{" }} !contains(needs.*.result, 'failure') {{ "}}" }}
        run: |
          echo "✅ Docker image build completed successfully"
          # Add Slack notification here if needed

      - name: Notify on failure
        if: ${{ "{{" }} contains(needs.*.result, 'failure') {{ "}}" }}
        run: |
          echo "❌ Docker image build failed"
          # Add Slack notification here if needed
          exit 1
`
