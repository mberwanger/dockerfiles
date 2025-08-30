package workflow

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

//go:embed templates/workflow.tmpl
var workflowTemplate string

type Workflow struct {
	Jobs []Job
}

type Job struct {
	ID             string
	Name           string
	ImageName      string
	Version        string
	DockerfilePath string
	Needs          []string
}

func Generate(cfg *config.Config, outputPath string) error {
	jobs, err := buildJobsFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("building jobs from config: %w", err)
	}

	orderedJobs, err := orderJobsByDependencies(jobs)
	if err != nil {
		return fmt.Errorf("ordering jobs by dependencies: %w", err)
	}

	if err := writeWorkflow(orderedJobs, outputPath); err != nil {
		return fmt.Errorf("writing workflow: %w", err)
	}

	return nil
}

func GenerateToWriter(cfg *config.Config, w io.Writer) error {
	jobs, err := buildJobsFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("building jobs from config: %w", err)
	}

	orderedJobs, err := orderJobsByDependencies(jobs)
	if err != nil {
		return fmt.Errorf("ordering jobs by dependencies: %w", err)
	}

	if err := writeWorkflowToWriter(orderedJobs, w); err != nil {
		return fmt.Errorf("writing workflow: %w", err)
	}

	return nil
}

func buildJobsFromConfig(cfg *config.Config) ([]Job, error) {
	var jobs []Job

	// Sort image names for deterministic ordering
	imageNames := make([]string, 0, len(cfg.Images))
	for imageName := range cfg.Images {
		imageNames = append(imageNames, imageName)
	}
	sort.Strings(imageNames)

	for _, imageName := range imageNames {
		image := cfg.Images[imageName]

		// Sort versions for deterministic ordering
		versions := make([]string, 0, len(image.Versions))
		for version := range image.Versions {
			versions = append(versions, version)
		}
		sort.Strings(versions)

		for _, version := range versions {
			dockerfilePath := filepath.Join("images", image.Path, version, "Dockerfile")

			job := Job{
				ID:             generateJobID(imageName, version),
				Name:           fmt.Sprintf("Build %s:%s", imageName, version),
				ImageName:      imageName,
				Version:        version,
				DockerfilePath: dockerfilePath,
			}

			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func orderJobsByDependencies(jobs []Job) ([]Job, error) {
	jobMap := make(map[string]*Job)
	for i := range jobs {
		key := fmt.Sprintf("%s:%s", jobs[i].ImageName, jobs[i].Version)
		jobMap[key] = &jobs[i]
	}

	for i := range jobs {
		deps, err := parseDockerfileDependencies(jobs[i].DockerfilePath)
		if err != nil {
			return nil, fmt.Errorf("parsing dependencies for %s: %w", jobs[i].Name, err)
		}

		var needs []string
		for _, dep := range deps {
			if depJob, exists := jobMap[dep]; exists {
				needs = append(needs, depJob.ID)
			}
		}
		jobs[i].Needs = needs
	}

	sorted, err := topologicalSort(jobs)
	if err != nil {
		return nil, fmt.Errorf("sorting jobs: %w", err)
	}

	return sorted, nil
}

func parseDockerfileDependencies(dockerfilePath string) ([]string, error) {
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("reading Dockerfile: %w", err)
	}

	depsMap := make(map[string]bool)
	lines := strings.Split(string(content), "\n")

	fromPattern := regexp.MustCompile(`^\s*FROM\s+\$\{REGISTRY\}/([^:\s]+):([^\s]+)`)
	copyFromPattern := regexp.MustCompile(`^\s*COPY\s+.*--from=([^\s]+)`)

	// Track internal stage names
	stageNames := make(map[string]bool)
	for _, line := range lines {
		if match := regexp.MustCompile(`^\s*FROM\s+.*\s+AS\s+([^\s]+)`).FindStringSubmatch(line); match != nil {
			stageNames[match[1]] = true
		}
	}

	for _, line := range lines {
		if match := fromPattern.FindStringSubmatch(line); match != nil {
			imageName := match[1]
			version := match[2]
			dep := fmt.Sprintf("%s:%s", imageName, version)
			depsMap[dep] = true
		}

		if match := copyFromPattern.FindStringSubmatch(line); match != nil {
			fromRef := match[1]
			// Skip if it's an internal stage reference
			if !stageNames[fromRef] {
				// Try to parse as ${REGISTRY}/image:version
				if registryMatch := regexp.MustCompile(`\$\{REGISTRY\}/([^:\s]+):([^\s]+)`).FindStringSubmatch(fromRef); registryMatch != nil {
					imageName := registryMatch[1]
					version := registryMatch[2]
					dep := fmt.Sprintf("%s:%s", imageName, version)
					depsMap[dep] = true
				}
			}
		}
	}

	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	return deps, nil
}

func topologicalSort(jobs []Job) ([]Job, error) {
	var sorted []Job
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	// Create job map for quick lookup
	jobMap := make(map[string]*Job)
	for i := range jobs {
		jobMap[jobs[i].ID] = &jobs[i]
	}

	var visit func(string) error
	visit = func(jobID string) error {
		if visited[jobID] {
			return nil
		}
		if visiting[jobID] {
			return fmt.Errorf("circular dependency detected involving job %s", jobID)
		}

		visiting[jobID] = true
		job := jobMap[jobID]

		// Visit dependencies first
		for _, dep := range job.Needs {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[jobID] = false
		visited[jobID] = true
		sorted = append(sorted, *job)
		return nil
	}

	// Visit all jobs
	for _, job := range jobs {
		if err := visit(job.ID); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

func generateJobID(imageName, version string) string {
	id := fmt.Sprintf("%s-%s", imageName, version)
	id = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(id, "-")
	id = regexp.MustCompile(`-+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")

	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(id) {
		id = "build-" + id
	}

	return id
}

func writeWorkflow(jobs []Job, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return writeWorkflowToWriter(jobs, file)
}

func writeWorkflowToWriter(jobs []Job, w io.Writer) error {
	tmpl, err := template.New("workflow").Parse(workflowTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	data := Workflow{Jobs: jobs}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
