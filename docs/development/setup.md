# 🛠️ Development Setup

This guide covers setting up a development environment for contributing to IaC Recertification Engine.

## Prerequisites

### Required Software
- **Go**: Version 1.24.0 or later
- **Git**: Version control system
- **Make**: Build automation tool (optional but recommended)

### Optional Tools
- **golangci-lint**: Go linter and formatter
- **Docker**: For containerized development and testing
- **GitHub CLI**: For enhanced GitHub integration

## Installation

### 1. Install Go

#### Linux (Ubuntu/Debian)
```bash
# Add the official Go repository
sudo apt update
sudo apt install software-properties-common
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# Download and install Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

#### macOS
```bash
# Using Homebrew
brew install go

# Or download from go.dev
# https://go.dev/dl/
```

#### Windows
```powershell
# Download from go.dev
# https://go.dev/dl/

# Or using Chocolatey
choco install golang
```

#### Verify Installation
```bash
go version
# Should show: go version go1.24.0 or later
```

### 2. Install Git
```bash
# Linux
sudo apt install git

# macOS
brew install git

# Windows: Download from https://git-scm.com/
```

### 3. Install Make (Optional)
```bash
# Linux
sudo apt install make

# macOS
brew install make

# Windows: Install via Chocolatey or use WSL
```

### 4. Install golangci-lint (Recommended)
```bash
# Using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or download binary from https://github.com/golangci/golangci-lint/releases
```

### 5. Install GitHub CLI (Optional)
```bash
# Linux
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

# macOS
brew install gh

# Windows
winget install --id GitHub.cli
```

## Clone and Setup Project

### 1. Clone the Repository
```bash
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine
```

### 2. Initialize Go Modules
```bash
# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy

# Verify modules
go mod verify
```

### 3. Build the Project
```bash
# Using Make (recommended)
make build

# Or using Go directly
go build -o ice ./cmd/ice
```

### 4. Verify Build
```bash
# Check if binary was created
ls -la ice

# Run version command
./ice version
```

## Development Environment Configuration

### Environment Variables
Set up environment variables for development:

```bash
# Add Go bin to PATH (if not already done)
export PATH=$PATH:$(go env GOPATH)/bin

# Set up Git (replace with your info)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Optional: Set up GitHub CLI
gh auth login
```

### IDE Setup

#### Visual Studio Code
1. Install Go extension: `ms-vscode.go`
2. Configure settings in `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v"],
  "go.testTimeout": "300s"
}
```

#### GoLand
1. Import project as Go module
2. Configure Go SDK (1.24+)
3. Enable golangci-lint integration

#### Vim/Neovim
Install vim-go plugin and configure:
```vim
let g:go_fmt_command = "gofmt"
let g:go_metalinter_command = "golangci-lint"
```

### Pre-commit Hooks (Optional)
Set up pre-commit hooks to ensure code quality:

```bash
# Install pre-commit (if not already installed)
pip install pre-commit

# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Testing Setup

### Run Tests
```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/scan/...
```

### Test Configuration
- Tests use the `testify` framework
- Mock external dependencies as needed
- Integration tests may require test repositories

## Docker Development (Optional)

### Build Docker Image
```bash
# Build development image
docker build -t ice-dev -f Dockerfile .

# Or use docker-compose
docker-compose build
```

### Run in Docker
```bash
# Run tests in container
docker run --rm -v $(pwd):/app ice-dev make test

# Run development shell
docker run --rm -it -v $(pwd):/app ice-dev bash
```

## Troubleshooting Setup Issues

### Common Issues

**"go: command not found"**
- Ensure Go is installed and in PATH
- Restart your terminal/shell
- Check installation location: `which go`

**"go mod download: module not found"**
- Check internet connection
- Clear module cache: `go clean -modcache`
- Reinitialize: `go mod tidy`

**Build failures**
- Ensure Go 1.24+ is installed: `go version`
- Clean build: `go clean -cache && go clean -modcache`
- Rebuild: `make clean && make build`

**Permission issues**
- On Linux/macOS: `chmod +x ice`
- On Windows: Check execution permissions

**golangci-lint issues**
- Ensure golangci-lint is installed: `golangci-lint --version`
- Update config: Check `.golangci.yml`
- Run manually: `golangci-lint run`

### Getting Help
- Check [Go installation docs](https://go.dev/doc/install)
- Review [golangci-lint docs](https://golangci-lint.run/)
- Search existing GitHub issues
- Ask in community discussions

## Next Steps

With your development environment set up:

1. [Read the contributing guide](contributing.md)
2. [Learn about testing](testing.md)
3. [Understand the build process](building.md)
4. Start exploring the codebase and making contributions!

## Development Tips

- Use `go mod tidy` regularly to keep dependencies clean
- Run `make lint` before committing
- Use `go test -race` to detect race conditions
- Keep your Go version updated
- Use branches for feature development
- Write tests for new code
- Follow the existing code patterns and style
