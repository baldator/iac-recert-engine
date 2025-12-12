# 🔨 Building

This guide covers building IaC Recertification Engine from source code.

## Prerequisites

### Required Tools
- **Go**: Version 1.24.0 or later
- **Git**: For version control
- **Make**: Build automation (optional but recommended)

### Optional Tools
- **Docker**: For containerized builds
- **Cross-compilation tools**: For building multiple architectures

## Quick Start

### Basic Build
```bash
# Clone the repository
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine

# Build using Make (recommended)
make build

# Or build directly with Go
go build -o ice ./cmd/ice

# Verify build
./ice version
```

## Build Process

### Go Build Basics

ICE is built using standard Go tooling. The main entry point is in `cmd/ice/main.go`.

#### Build Command
```bash
go build -o ice ./cmd/ice
```

#### Build Flags
```bash
# Build with version information
go build -ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" -o ice ./cmd/ice

# Build with optimization
go build -ldflags="-s -w" -o ice ./cmd/ice

# Build with debug information
go build -gcflags="all=-N -l" -o ice ./cmd/ice
```

### Makefile Targets

The project includes a Makefile with common build targets:

```bash
# Build the binary
make build

# Run tests
make test

# Run linting
make lint

# Clean build artifacts
make clean

# Install dependencies
make deps

# Cross-platform build
make build-all

# Docker build
make docker-build
```

### Build Configuration

#### Go Modules
The project uses Go modules for dependency management:

```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
```

#### Build Tags
Use build tags to include/exclude code:

```bash
# Build with debug features
go build -tags debug -o ice ./cmd/ice

# Build without certain features
go build -tags '!feature' -o ice ./cmd/ice
```

## Cross-Platform Building

### Supported Platforms
ICE supports multiple operating systems and architectures:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64
- **FreeBSD**: amd64

### Cross-Compilation
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ice-linux-amd64 ./cmd/ice

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o ice-linux-arm64 ./cmd/ice

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o ice-darwin-amd64 ./cmd/ice

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o ice-darwin-arm64 ./cmd/ice

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ice-windows-amd64.exe ./cmd/ice

# FreeBSD AMD64
GOOS=freebsd GOARCH=amd64 go build -o ice-freebsd-amd64 ./cmd/ice
```

### Automated Cross-Compilation
```bash
# Build for all platforms
make build-all

# Or use a script
#!/bin/bash
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    output="ice-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$output" ./cmd/ice
done
```

## Docker Building

### Dockerfile
The project includes a multi-stage Dockerfile for optimized builds:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ice ./cmd/ice

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ice .
CMD ["./ice"]
```

### Building Docker Image
```bash
# Build locally
docker build -t ice .

# Build with specific tag
docker build -t ice:v1.0.0 .

# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t ice .

# Run the container
docker run --rm ice version
```

### Docker Compose
For development with Docker Compose:

```yaml
version: '3.8'
services:
  ice:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - .:/app
    working_dir: /app
    command: go build -o ice ./cmd/ice
```

## Build Optimization

### Binary Size Optimization
```bash
# Strip debug information
go build -ldflags="-s -w" -o ice ./cmd/ice

# Use UPX compression
upx --best --lzma ice

# Check binary size
ls -lh ice
```

### Performance Optimization
```bash
# Build with optimizations
go build -ldflags="-s -w" -gcflags="all=-l -B" -o ice ./cmd/ice

# Profile-guided optimization (Go 1.20+)
go build -ldflags="-s -w" -pgo=default.pgo -o ice ./cmd/ice
```

### Static Linking
```bash
# Fully static binary
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o ice ./cmd/ice

# Check dynamic dependencies
ldd ice  # Should show "not a dynamic executable"
```

## Release Builds

### Version Information
Embed version information in the binary:

```bash
# Get version from git
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse HEAD)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build with version info
go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME" -o ice ./cmd/ice
```

### Release Script
```bash
#!/bin/bash
set -e

VERSION=${1:-"v$(date +%Y%m%d-%H%M%S)"}
OUTPUT_DIR="dist"

# Clean
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build for all platforms
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output="$OUTPUT_DIR/ice-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo "Building $output..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w -X main.version=$VERSION" \
        -o "$output" \
        ./cmd/ice
done

# Create archives
cd "$OUTPUT_DIR"
for file in ice-*; do
    if [[ "$file" == *.exe ]]; then
        zip "${file%.exe}.zip" "$file"
    else
        tar -czf "${file}.tar.gz" "$file"
    fi
done

echo "Release builds completed in $OUTPUT_DIR"
```

## CI/CD Integration

### GitHub Actions
```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Build
        run: go build -v ./cmd/ice
      
      - name: Test
        run: go test -v ./...
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: ice
          path: ice
```

### Release Automation
```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Build releases
        run: make release
      
      - name: Create release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false
```

## Troubleshooting Build Issues

### Common Problems

**"go: command not found"**
- Ensure Go is installed and in PATH
- Check installation: `go version`

**"module not found" errors**
- Run `go mod tidy`
- Check network connectivity
- Clear module cache: `go clean -modcache`

**"build constraints exclude all Go files"**
- Check build tags match
- Ensure correct GOOS/GOARCH for cross-compilation

**Large binary size**
- Use `-ldflags="-s -w"` to strip debug info
- Consider UPX compression
- Check for unnecessary dependencies

**Cross-compilation failures**
- Ensure CGO_ENABLED=0 for cross-compilation
- Check target platform support
- Use correct GOOS/GOARCH values

**Docker build issues**
- Ensure Dockerfile is in correct location
- Check build context
- Verify base image availability

### Build Debugging
```bash
# Verbose build output
go build -v -x ./cmd/ice

# Show build environment
go env

# Check dependencies
go list -m all

# Analyze binary
file ice
nm ice | head -20
```

## Performance Considerations

### Build Time Optimization
- Use build caching with `go build -cache`
- Parallel builds with `make -j$(nproc)`
- Incremental builds with proper dependencies

### Runtime Performance
- Profile builds with `go build -pgo`
- Use appropriate optimization flags
- Consider static linking for deployment

## Contributing Builds

When contributing build-related changes:

1. **Test builds** on multiple platforms
2. **Update documentation** for build changes
3. **Maintain compatibility** with existing build scripts
4. **Follow conventions** for version embedding
5. **Ensure reproducibility** of builds

### Build Checklist
- [ ] Builds successfully on Linux/macOS/Windows
- [ ] Cross-compilation works for all targets
- [ ] Docker build succeeds
- [ ] Version information is embedded
- [ ] Binary size is reasonable
- [ ] No build warnings or errors
- [ ] CI/CD pipelines pass

## Resources

- [Go Build Documentation](https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies)
- [Go Modules](https://golang.org/ref/mod)
- [Cross-Compilation Guide](https://golang.org/doc/install/source#environment)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [GitHub Actions](https://docs.github.com/en/actions)
