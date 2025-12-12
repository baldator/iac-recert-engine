# Installation

This guide covers all available installation methods for IaC Recertification Engine.

## System Requirements

### Minimum Requirements
- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.24 or later (for building from source)
- **Memory**: 256 MB RAM minimum, 512 MB recommended
- **Disk Space**: 50 MB for binary, additional space for audit logs

### Recommended Requirements
- **CPU**: 1+ core
- **Memory**: 1 GB RAM
- **Network**: Stable internet connection for Git provider APIs

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

Download pre-built binaries for your platform from the [GitHub Releases](https://github.com/baldator/iac-recert-engine/releases) page.

#### Linux (amd64)
```bash
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/
```

#### Linux (arm64)
```bash
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-arm64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-darwin-amd64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-darwin-arm64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/
```

#### Windows (amd64)
```powershell
# PowerShell
Invoke-WebRequest -Uri "https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-windows-amd64.exe" -OutFile "ice.exe"
# Add to PATH or move to a directory in PATH
```

#### Verify Installation
```bash
ice version
```

### Method 2: Docker (Containerized)

Use the official Docker image from GitHub Container Registry.

#### Pull the Image
```bash
docker pull ghcr.io/baldator/iac-recert-engine:latest
```

#### Verify Installation
```bash
docker run --rm ghcr.io/baldator/iac-recert-engine:latest version
```

#### Using Docker Compose
Create a `docker-compose.yml`:
```yaml
version: '3.8'
services:
  ice:
    image: ghcr.io/baldator/iac-recert-engine:latest
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./audit:/app/audit
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    command: run --config /app/config.yaml
```

### Method 3: Build from Source

Build ICE from the source code.

#### Prerequisites
- Go 1.24 or later
- Git
- Make (optional, for using Makefile)

#### Clone and Build
```bash
# Clone the repository
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine

# Build the binary
go build -o ice ./cmd/ice

# Optional: Use Makefile for additional targets
make build
```

#### Cross-Platform Building
```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o ice-linux-amd64 ./cmd/ice

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o ice-darwin-amd64 ./cmd/ice

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o ice-windows-amd64.exe ./cmd/ice
```

### Method 4: Kubernetes Deployment

Deploy ICE as a Kubernetes CronJob for scheduled execution.

#### Using Helm (Recommended)
```bash
# Add the repository (when available)
helm repo add ice https://charts.example.com/
helm install ice ice/iac-recert-engine

# Or use the provided helm chart
helm install ice ./helm/iac-recert-engine
```

#### Manual Kubernetes Manifest
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: iac-recert-engine
spec:
  schedule: "0 2 * * 0"  # Weekly on Sunday at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: ice
            image: ghcr.io/baldator/iac-recert-engine:latest
            env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: ice-secrets
                  key: github-token
            volumeMounts:
            - name: config
              mountPath: /app/config.yaml
              subPath: config.yaml
          volumes:
          - name: config
            configMap:
              name: ice-config
          restartPolicy: OnFailure
```

### Method 5: CI/CD Integration

Integrate ICE directly into your CI/CD pipelines.

#### GitHub Actions
```yaml
name: IaC Recertification
on:
  schedule:
    - cron: '0 2 * * 0'
  workflow_dispatch:

jobs:
  recertify:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Build ICE
        run: go build -o ice ./cmd/ice

      - name: Run Recertification
        run: ./ice run --config config.yaml
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### GitLab CI
```yaml
stages:
  - recertify

iac_recertification:
  stage: recertify
  image: golang:1.24
  only:
    - schedules
  script:
    - go build -o ice ./cmd/ice
    - ./ice run --config config.yaml
  dependencies: []
```

#### Azure DevOps Pipeline
```yaml
trigger: none
schedules:
- cron: "0 2 * * 0"
  displayName: Weekly Recertification
  branches:
    include:
    - main

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: GoTool@0
  inputs:
    version: '1.24'

- script: go build -o ice ./cmd/ice
  displayName: Build ICE

- script: ./ice run --config config.yaml
  displayName: Run Recertification
  env:
    AZURE_DEVOPS_TOKEN: $(System.AccessToken)
```

## Configuration

After installation, create a configuration file. See the [Configuration Overview](configuration/overview.md) for detailed options.

### Basic Configuration
```yaml
version: "1.0"

repository:
  url: "https://github.com/your-org/your-repo"
  provider: "github"

auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"

patterns:
  - name: "terraform"
    paths: ["**/*.tf"]
    recertification_days: 90

pr_strategy:
  type: "per_pattern"

assignment:
  strategy: "static"
  fallback_assignees: ["your-team"]
```

## Authentication Setup

Configure authentication for your Git provider.

### GitHub
1. Create a [Personal Access Token](https://github.com/settings/tokens)
2. Required scopes: `repo`, `workflow`
3. Set environment variable: `export GITHUB_TOKEN=your_token`

### Azure DevOps
1. Create a [Personal Access Token](https://dev.azure.com/your-org/_usersSettings/tokens)
2. Required scopes: `Code (read, write)`, `Pull Request Threads (read, write)`
3. Set environment variable: `export AZURE_DEVOPS_TOKEN=your_token`

### GitLab
1. Create a [Personal Access Token](https://gitlab.com/-/profile/personal_access_tokens)
2. Required scopes: `api`, `read_repository`, `write_repository`
3. Set environment variable: `export GITLAB_TOKEN=your_token`

## Verification

Verify your installation is working correctly:

```bash
# Check version
ice version

# Validate configuration
ice config validate --config config.yaml

# Run a dry-run test
ice run --config config.yaml --dry-run
```

## Troubleshooting Installation

### Common Issues

**"command not found" error**
- Ensure the binary is in your PATH
- On Windows, you may need to restart your terminal
- On macOS/Linux, check your shell profile (.bashrc, .zshrc)

**"permission denied" error**
- Ensure the binary has execute permissions: `chmod +x ice`
- On macOS, you may need to allow the binary in Security & Privacy settings

**Docker permission issues**
- Ensure your user is in the docker group: `sudo usermod -aG docker $USER`
- Restart your session after adding to the docker group

**Build failures**
- Ensure Go 1.24+ is installed: `go version`
- Clear module cache: `go clean -modcache`
- Reinitialize modules: `go mod tidy`

### Getting Help

- Check the [Troubleshooting](troubleshooting/common-issues.md) section
- Review GitHub Issues for similar problems
- Join community discussions for installation help

## Next Steps

With ICE installed, you're ready to:

- [Create your first configuration](quick-start.md)
- [Explore configuration options](configuration/overview.md)
- [Set up automated scheduling](usage/scheduling.md)
- [Integrate with your CI/CD](usage/ci-cd.md)
