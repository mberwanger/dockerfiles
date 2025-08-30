# Docker Base Images

Docker base images built from templates using a Go-based generation system.

## Quick Start

```bash
# Generate all Dockerfiles from templates
go run ./tool generate image --all

# Generate specific image
go run ./tool generate image core

# Generate GitHub Actions workflow
go run ./tool generate workflow -o .github/workflows/dockerfiles.yaml

# Clean generated files
go run ./tool clean
```

## Directory Structure

```
images/
├── manifest.yaml          # Central configuration for all images
├── base/                  # Foundational base images (Debian, tini)
├── lang/                  # Programming languages (Python, Go, Java)
├── app/                   # Applications (mailcatcher, minio)
├── runtime/               # Web servers (node-passenger, node-pm2, php-fpm)
└── util/                  # Development utilities (k8s-toolbox, yq)

tool/                      # Go build system
├── cmd/                   # CLI commands
└── internal/              # Generator, template engine, workflow builder
```

## Development Workflow

1. **Edit templates**: Modify source files in `images/{category}/{image}/source/`
   - Template files use `.tmpl` extension with Go template syntax
   - Non-template files (like certificates) are copied as-is

2. **Generate Dockerfiles**: Run `go run ./tool generate image --all`

3. **Generate workflow** (optional): Run `go run ./tool generate workflow -o .github/workflows/dockerfiles.yaml`

4. **Test locally**: Build and test your changes

5. **Commit changes**: Commit both template and generated files

6. **Create PR**: CI validates that generated files are up to date

7. **Merge**: Images are built and published automatically

## Template System

Templates use Go's `text/template` syntax:

```dockerfile
{{generation_message}}
FROM {{from_image .base_image}}

RUN apt-get update && apt-get install -y \
    python{{.python_version}} \
    && rm -rf /var/lib/apt/lists/*
```

### Template Functions

- `generation_message`: Adds "GENERATED FILE, DO NOT MODIFY" header
- `from_image`: Generates FROM statements with proper registry paths
- Standard Go template functions: `index`, `range`, `if`, etc.

## Manifest Configuration

The `images/manifest.yaml` defines all images and their versions:

```yaml
version: 1

defaults:
  registry: ghcr.io/mberwanger

images:
  python:
    path: lang/python
    defaults:
      base_image:
        name: core:bullseye
    versions:
      "3.13":
        python_version: "3.13"
```

## Important Notes

- **Never edit generated Dockerfiles directly** - always modify templates
- **Generated files must be committed** - CI validates they're up to date
- Source directories contain both `.tmpl` templates and static files

## Requirements

- Go 1.21 or later
- Docker (for building images locally)
