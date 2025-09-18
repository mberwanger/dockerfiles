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
