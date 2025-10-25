# Docker Base Images

Docker base images built from templates using a Go-based generation system.

## Quick Start

```bash
# Generate all Dockerfiles and workflow (default)
make

# Generate all Dockerfiles
make generate-all

# Generate specific image
make generate IMAGE=core

# Generate GitHub Actions workflow
make generate-workflow

# Run tests
make test

# Clean generated files
make clean
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

1. **Edit configuration or templates**:
   - Update `images/manifest.yaml` to add/modify image versions
   - Or modify template files in `images/{category}/{image}/source/`
   - Template files use `.tmpl` extension with Go template syntax
   - Non-template files (like certificates) are copied as-is

2. **Regenerate files**: Run `make` at the root of the repo
   - This generates all Dockerfiles from templates
   - And updates the GitHub Actions workflow

3. **Commit changes**: Commit both template/manifest and generated files
   ```bash
   git add images/ .github/workflows/
   git commit -m "Update Python to 3.13"
   ```

4. **Create PR**: Open a pull request
   - CI validates that generated files are up to date
   - Docker images are built and tested (but not pushed)

5. **Merge to main**: Once approved and merged
   - Images are automatically built and pushed to the registry

**Note**: A daily scheduled job (1 PM UTC) automatically rebuilds and pushes all images to keep them up to date with the latest base image updates and security patches.

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
- Make
- Docker (for building images locally)
