package workflow

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Images: map[string]config.Image{
			"app1": {
				Path: "app1",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
			"app2": {
				Path: "app2",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
		},
	}

	// Create mock Dockerfiles
	for _, image := range cfg.Images {
		for version := range image.Versions {
			dockerfilePath := filepath.Join(tmpDir, "images", image.Path, version, "Dockerfile")
			if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
				t.Fatalf("Failed to write Dockerfile: %v", err)
			}
		}
	}

	outputPath := filepath.Join(tmpDir, "workflow.yaml")

	// Change to tmpDir so relative paths work
	oldWd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	if err := Generate(cfg, outputPath); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "app1") {
		t.Error("Output should contain app1")
	}
	if !strings.Contains(output, "app2") {
		t.Error("Output should contain app2")
	}
}

func TestGenerateToWriter(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Images: map[string]config.Image{
			"myapp": {
				Path: "myapp",
				Versions: map[string]*config.ImageConfig{
					"v1.0": {},
				},
			},
		},
	}

	// Create mock Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "images/myapp/v1.0/Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(dockerfilePath, []byte("FROM ubuntu\n"), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Change to tmpDir
	oldWd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	var buf bytes.Buffer
	if err := GenerateToWriter(cfg, &buf); err != nil {
		t.Fatalf("GenerateToWriter() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "myapp") {
		t.Error("Output should contain myapp")
	}
	if !strings.Contains(output, "v1.0") {
		t.Error("Output should contain v1.0")
	}
}

func TestBuildJobsFromConfig(t *testing.T) {
	cfg := &config.Config{
		Images: map[string]config.Image{
			"app1": {
				Path: "path/to/app1",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
					"v2": {},
				},
			},
			"app2": {
				Path: "path/to/app2",
				Versions: map[string]*config.ImageConfig{
					"v1": {},
				},
			},
		},
	}

	jobs, err := buildJobsFromConfig(cfg)
	if err != nil {
		t.Fatalf("buildJobsFromConfig() error = %v", err)
	}

	// Should have 3 jobs total (app1:v1, app1:v2, app2:v1)
	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}

	// Verify job properties
	foundApp1V1 := false
	for _, job := range jobs {
		if job.ImageName == "app1" && job.Version == "v1" {
			foundApp1V1 = true
			if job.Name != "Build app1:v1" {
				t.Errorf("Job name = %s, want Build app1:v1", job.Name)
			}
			expectedPath := filepath.Join("images", "path/to/app1", "v1", "Dockerfile")
			if job.DockerfilePath != expectedPath {
				t.Errorf("DockerfilePath = %s, want %s", job.DockerfilePath, expectedPath)
			}
		}
	}

	if !foundApp1V1 {
		t.Error("Job for app1:v1 not found")
	}

	// Verify deterministic ordering (sorted by image name then version)
	if jobs[0].ImageName != "app1" {
		t.Errorf("First job should be for app1, got %s", jobs[0].ImageName)
	}
}

func TestOrderJobsByDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Dockerfiles with dependencies
	// base:v1 has no dependencies
	basePath := filepath.Join(tmpDir, "images/base/v1/Dockerfile")
	if err := os.MkdirAll(filepath.Dir(basePath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(basePath, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// app:v1 depends on base:v1
	appPath := filepath.Join(tmpDir, "images/app/v1/Dockerfile")
	if err := os.MkdirAll(filepath.Dir(appPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	appDockerfile := "ARG REGISTRY=test.io\nFROM ${REGISTRY}/base:v1\n"
	if err := os.WriteFile(appPath, []byte(appDockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	jobs := []Job{
		{
			ID:             "app-v1",
			Name:           "Build app:v1",
			ImageName:      "app",
			Version:        "v1",
			DockerfilePath: appPath,
		},
		{
			ID:             "base-v1",
			Name:           "Build base:v1",
			ImageName:      "base",
			Version:        "v1",
			DockerfilePath: basePath,
		},
	}

	ordered, err := orderJobsByDependencies(jobs)
	if err != nil {
		t.Fatalf("orderJobsByDependencies() error = %v", err)
	}

	if len(ordered) != 2 {
		t.Errorf("Expected 2 jobs, got %d", len(ordered))
	}

	// base:v1 should come before app:v1
	baseIndex := -1
	appIndex := -1
	for i, job := range ordered {
		if job.ID == "base-v1" {
			baseIndex = i
		}
		if job.ID == "app-v1" {
			appIndex = i
		}
	}

	if baseIndex == -1 || appIndex == -1 {
		t.Fatal("Jobs not found in ordered list")
	}

	if baseIndex >= appIndex {
		t.Error("base-v1 should come before app-v1")
	}

	// Verify app:v1 has base-v1 in its needs
	var appJob *Job
	for i := range ordered {
		if ordered[i].ID == "app-v1" {
			appJob = &ordered[i]
			break
		}
	}

	if appJob == nil {
		t.Fatal("app-v1 job not found")
	}

	foundDep := false
	for _, need := range appJob.Needs {
		if need == "base-v1" {
			foundDep = true
			break
		}
	}
	if !foundDep {
		t.Error("app-v1 should have base-v1 as a dependency")
	}
}

func TestParseDockerfileDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		dockerfile string
		wantDeps   []string
		wantErr    bool
	}{
		{
			name:       "no dependencies",
			dockerfile: "FROM alpine\nRUN echo hello\n",
			wantDeps:   []string{},
			wantErr:    false,
		},
		{
			name:       "single dependency",
			dockerfile: "ARG REGISTRY=test.io\nFROM ${REGISTRY}/base:v1\n",
			wantDeps:   []string{"base:v1"},
			wantErr:    false,
		},
		{
			name: "multiple dependencies",
			dockerfile: `ARG REGISTRY=test.io
FROM ${REGISTRY}/base:v1
COPY --from=${REGISTRY}/builder:v2 /app /app
`,
			wantDeps: []string{"base:v1", "builder:v2"},
			wantErr:  false,
		},
		{
			name: "with internal stage",
			dockerfile: `FROM alpine AS builder
RUN make
FROM ${REGISTRY}/base:v1
COPY --from=builder /output /output
`,
			wantDeps: []string{"base:v1"},
			wantErr:  false,
		},
		{
			name: "duplicate dependencies",
			dockerfile: `ARG REGISTRY=test.io
FROM ${REGISTRY}/base:v1
COPY --from=${REGISTRY}/base:v1 /lib /lib
`,
			wantDeps: []string{"base:v1"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerfilePath := filepath.Join(tmpDir, tt.name, "Dockerfile")
			if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(dockerfilePath, []byte(tt.dockerfile), 0644); err != nil {
				t.Fatalf("Failed to write Dockerfile: %v", err)
			}

			deps, err := parseDockerfileDependencies(dockerfilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDockerfileDependencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(deps) != len(tt.wantDeps) {
				t.Errorf("Got %d dependencies, want %d", len(deps), len(tt.wantDeps))
				t.Errorf("Got: %v, want: %v", deps, tt.wantDeps)
				return
			}

			for i, dep := range deps {
				if dep != tt.wantDeps[i] {
					t.Errorf("Dependency %d = %s, want %s", i, dep, tt.wantDeps[i])
				}
			}
		})
	}
}

func TestParseDockerfileDependencies_FileNotFound(t *testing.T) {
	_, err := parseDockerfileDependencies("/nonexistent/Dockerfile")
	if err == nil {
		t.Error("parseDockerfileDependencies() should return error for nonexistent file")
	}
}

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []Job
		wantErr bool
	}{
		{
			name: "simple chain",
			jobs: []Job{
				{ID: "c", Needs: []string{"b"}},
				{ID: "b", Needs: []string{"a"}},
				{ID: "a", Needs: []string{}},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			jobs: []Job{
				{ID: "a", Needs: []string{"b"}},
				{ID: "b", Needs: []string{"a"}},
			},
			wantErr: true,
		},
		{
			name: "no dependencies",
			jobs: []Job{
				{ID: "a", Needs: []string{}},
				{ID: "b", Needs: []string{}},
			},
			wantErr: false,
		},
		{
			name: "diamond dependency",
			jobs: []Job{
				{ID: "d", Needs: []string{"b", "c"}},
				{ID: "b", Needs: []string{"a"}},
				{ID: "c", Needs: []string{"a"}},
				{ID: "a", Needs: []string{}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted, err := topologicalSort(tt.jobs)
			if (err != nil) != tt.wantErr {
				t.Errorf("topologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify all jobs are present
			if len(sorted) != len(tt.jobs) {
				t.Errorf("Sorted has %d jobs, want %d", len(sorted), len(tt.jobs))
			}

			// Verify dependencies are satisfied (dependencies come before dependents)
			jobIndex := make(map[string]int)
			for i, job := range sorted {
				jobIndex[job.ID] = i
			}

			for _, job := range sorted {
				for _, dep := range job.Needs {
					depIdx, depExists := jobIndex[dep]
					jobIdx := jobIndex[job.ID]
					if !depExists {
						t.Errorf("Dependency %s not found for job %s", dep, job.ID)
					}
					if depIdx >= jobIdx {
						t.Errorf("Dependency %s should come before %s", dep, job.ID)
					}
				}
			}
		})
	}
}

func TestGenerateJobID(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		version   string
		want      string
	}{
		{
			name:      "simple",
			imageName: "app",
			version:   "v1",
			want:      "app-v1",
		},
		{
			name:      "with dots",
			imageName: "my.app",
			version:   "v1.0",
			want:      "my-app-v1-0",
		},
		{
			name:      "with special chars",
			imageName: "my_app",
			version:   "v1-beta",
			want:      "my_app-v1-beta",
		},
		{
			name:      "starts with number",
			imageName: "123app",
			version:   "v1",
			want:      "build-123app-v1",
		},
		{
			name:      "multiple consecutive special chars",
			imageName: "my:::app",
			version:   "v1",
			want:      "my-app-v1",
		},
		{
			name:      "leading/trailing dashes",
			imageName: "-app-",
			version:   "-v1-",
			want:      "app-v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateJobID(tt.imageName, tt.version)
			if got != tt.want {
				t.Errorf("generateJobID() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestWriteWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output", "workflow.yaml")

	jobs := []Job{
		{
			ID:             "test-v1",
			Name:           "Build test:v1",
			ImageName:      "test",
			Version:        "v1",
			DockerfilePath: "images/test/v1/Dockerfile",
			Needs:          []string{},
		},
	}

	if err := writeWorkflow(jobs, outputPath); err != nil {
		t.Fatalf("writeWorkflow() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "test-v1") {
		t.Error("Output should contain job ID")
	}
	if !strings.Contains(output, "Build test:v1") {
		t.Error("Output should contain job name")
	}
}

func TestWriteWorkflowToWriter(t *testing.T) {
	jobs := []Job{
		{
			ID:             "myapp-v1",
			Name:           "Build myapp:v1",
			ImageName:      "myapp",
			Version:        "v1",
			DockerfilePath: "images/myapp/v1/Dockerfile",
			Needs:          []string{},
		},
		{
			ID:             "myapp-v2",
			Name:           "Build myapp:v2",
			ImageName:      "myapp",
			Version:        "v2",
			DockerfilePath: "images/myapp/v2/Dockerfile",
			Needs:          []string{"myapp-v1"},
		},
	}

	var buf bytes.Buffer
	if err := writeWorkflowToWriter(jobs, &buf); err != nil {
		t.Fatalf("writeWorkflowToWriter() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Output should not be empty")
	}
	if !strings.Contains(output, "myapp-v1") {
		t.Error("Output should contain first job")
	}
	if !strings.Contains(output, "myapp-v2") {
		t.Error("Output should contain second job")
	}
}

func TestBuildJobsFromConfig_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Images: map[string]config.Image{},
	}

	jobs, err := buildJobsFromConfig(cfg)
	if err != nil {
		t.Fatalf("buildJobsFromConfig() error = %v", err)
	}

	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs for empty config, got %d", len(jobs))
	}
}

func TestOrderJobsByDependencies_NoDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Create simple Dockerfiles with no dependencies
	for i := 1; i <= 3; i++ {
		path := filepath.Join(tmpDir, "Dockerfile", string(rune('0'+i)))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("FROM alpine\n"), 0644); err != nil {
			t.Fatalf("Failed to write Dockerfile: %v", err)
		}
	}

	jobs := []Job{
		{ID: "job1", DockerfilePath: filepath.Join(tmpDir, "Dockerfile/1")},
		{ID: "job2", DockerfilePath: filepath.Join(tmpDir, "Dockerfile/2")},
		{ID: "job3", DockerfilePath: filepath.Join(tmpDir, "Dockerfile/3")},
	}

	ordered, err := orderJobsByDependencies(jobs)
	if err != nil {
		t.Fatalf("orderJobsByDependencies() error = %v", err)
	}

	if len(ordered) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(ordered))
	}

	// All jobs should have no dependencies
	for _, job := range ordered {
		if len(job.Needs) != 0 {
			t.Errorf("Job %s should have no dependencies, got %d", job.ID, len(job.Needs))
		}
	}
}
