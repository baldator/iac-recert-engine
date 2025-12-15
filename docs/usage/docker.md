# Docker Usage

ICE can be run in Docker containers for easy deployment, isolation, and portability. This guide covers Docker usage patterns, configuration, and integration scenarios.

## Quick Start

### Basic Docker Run

Run ICE with a configuration file:

```bash
docker run --rm \
  -e GITHUB_TOKEN \
  -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml
```

### Dry Run (Safe Testing)

Test configuration without creating PRs:

```bash
docker run --rm \
  -e GITHUB_TOKEN \
  -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --dry-run --config /config.yaml
```

## Docker Images

### Official Images

ICE provides official Docker images:

- `ghcr.io/baldator/iac-recert-engine:latest` - Latest stable release
- `ghcr.io/baldator/iac-recert-engine:v1.0.0` - Specific version
- `ghcr.io/baldator/iac-recert-engine:main` - Development build

### Image Details

The Docker image is based on Alpine Linux and includes:

- Go runtime for plugin support
- Git for repository operations
- CA certificates for HTTPS
- Minimal attack surface

### Building Custom Images

Build your own image with custom plugins:

```dockerfile
FROM ghcr.io/baldator/iac-recert-engine:latest

# Add custom plugins
COPY plugins/ /app/plugins/

# Install additional tools if needed
RUN apk add --no-cache git curl

# Set working directory
WORKDIR /app

# Default command
CMD ["run"]
```

Build and tag:

```bash
docker build -t my-org/ice:latest .
```

## Configuration

### Volume Mounting

Mount configuration files:

```bash
docker run --rm \
  -v /path/to/config:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml
```

### Environment Variables

Pass configuration via environment:

```bash
docker run --rm \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -e ICE_CONFIG_PATH=/config.yaml \
  -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest
```

### Multiple Configurations

Use different configs for environments:

```bash
# Development
docker run --rm \
  -v config.dev.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml

# Production
docker run --rm \
  -v config.prod.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml
```

## Security

### Token Management

Never pass tokens as command arguments:

```bash
# ❌ Bad - token visible in process list
docker run --rm ghcr.io/baldator/iac-recert-engine:latest run --token $GITHUB_TOKEN

# ✅ Good - use environment variables
docker run --rm -e GITHUB_TOKEN ghcr.io/baldator/iac-recert-engine:latest run
```

### Secret Management

Use Docker secrets or external secret managers:

```bash
# Docker secrets
echo $GITHUB_TOKEN | docker secret create github_token -
docker service create --secret github_token --env GITHUB_TOKEN_FILE=/run/secrets/github_token ...

# AWS Secrets Manager
export GITHUB_TOKEN=$(aws secretsmanager get-secret-value --secret-id ice-tokens --query SecretString --output text | jq -r .github_token)

# HashiCorp Vault
export GITHUB_TOKEN=$(vault kv get -field=token secret/ice/github)
```

### User Permissions

Run as non-root user:

```bash
docker run --rm \
  --user $(id -u):$(id -g) \
  -v $(pwd):/work \
  -w /work \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config config.yaml
```

### Network Security

Limit network access:

```bash
# Allow only GitHub API
docker run --rm \
  --network none \
  --add-host api.github.com:140.82.121.6 \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config config.yaml
```

## Integration Patterns

### Docker Compose

Use Docker Compose for complex setups:

```yaml
# docker-compose.yml
version: '3.8'
services:
  ice:
    image: ghcr.io/baldator/iac-recert-engine:latest
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    volumes:
      - ./config.yaml:/config.yaml:ro
      - ./audit:/audit
    command: ["run", "--config", "/config.yaml"]
    restart: "no"
```

Run with compose:

```bash
docker-compose up ice
```

### Kubernetes

Deploy ICE as a Kubernetes Job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: ice-recert
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
          mountPath: /config
          readOnly: true
        command: ["ice", "run", "--config", "/config/config.yaml"]
      volumes:
      - name: config
        configMap:
          name: ice-config
      restartPolicy: Never
```

Apply and run:

```bash
kubectl apply -f ice-job.yaml
kubectl logs -f job/ice-recert
```

### Cron Jobs

Schedule regular runs with Docker:

```bash
# Add to crontab
0 2 * * 0 docker run --rm -e GITHUB_TOKEN -v /etc/ice:/config ghcr.io/baldator/iac-recert-engine:latest run --config /config/config.yaml
```

### Systemd with Docker

Create systemd service for Docker-based deployment:

```ini
# /etc/systemd/system/ice-docker.service
[Unit]
Description=ICE Recertification (Docker)
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
User=ice
Environment=GITHUB_TOKEN=your_token_here
ExecStart=/usr/bin/docker run --rm -e GITHUB_TOKEN -v /etc/ice:/config ghcr.io/baldator/iac-recert-engine:latest run --config /config/config.yaml
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### Docker Swarm

Deploy as a Swarm service:

```bash
docker service create \
  --name ice-recert \
  --env GITHUB_TOKEN \
  --mount type=bind,source=/etc/ice/config.yaml,target=/config.yaml \
  --restart-condition none \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml
```

## Advanced Usage

### Custom Entry Points

Create custom entry point scripts:

```dockerfile
FROM ghcr.io/baldator/iac-recert-engine:latest

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

```bash
#!/bin/sh
# entrypoint.sh

# Pre-run setup
echo "Starting ICE recertification..."

# Validate configuration
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Run ICE
exec ice run --config "$CONFIG_FILE" "$@"
```

### Multi-Stage Builds

Optimize image size:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o ice ./cmd/ice

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /app
COPY --from=builder /app/ice .
USER nobody
ENTRYPOINT ["./ice"]
CMD ["run"]
```

### Plugin Support

Include custom plugins in Docker images:

```dockerfile
FROM ghcr.io/baldator/iac-recert-engine:latest

# Copy custom plugins
COPY plugins/myplugin /app/plugins/myplugin/

# Ensure plugin permissions
RUN chmod +x /app/plugins/myplugin/myplugin

# Verify plugin loading
RUN ice run --help | grep -q "plugin" || exit 1
```

### Volume Management

Handle persistent data:

```bash
# Audit logs
docker run --rm \
  -v ice-audit:/audit \
  -e GITHUB_TOKEN \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml

# Cache directory
docker run --rm \
  -v ice-cache:/cache \
  -e XDG_CACHE_HOME=/cache \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /config.yaml
```

## Troubleshooting

### Image Pull Issues

```bash
# Check image availability
docker pull ghcr.io/baldator/iac-recert-engine:latest

# Use specific version
docker run ghcr.io/baldator/iac-recert-engine:v1.0.0 --version

# Build from source
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine
docker build -t ice-local .
```

### Permission Issues

```bash
# Fix volume permissions
docker run --rm -v $(pwd):/work alpine chmod -R 755 /work

# Run as current user
docker run --rm --user $(id -u):$(id -g) -v $(pwd):/work -w /work ...

# Use privileged mode (not recommended)
docker run --privileged ...
```

### Network Issues

```bash
# Test connectivity
docker run --rm alpine ping -c 3 api.github.com

# Use host network
docker run --rm --network host ...

# Configure DNS
docker run --rm --dns 8.8.8.8 ...
```

### Memory Issues

```bash
# Increase memory limit
docker run --memory 1g --memory-swap 2g ...

# Monitor resource usage
docker stats

# Debug memory issues
docker run -e GODEBUG=gctrace=1 ...
```

### Configuration Issues

```bash
# Validate config in container
docker run --rm -v $(pwd)/config.yaml:/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --dry-run --config /config.yaml

# Debug config loading
docker run --rm -v $(pwd)/config.yaml:/config.yaml \
  -e ICE_CONFIG_PATH=/config.yaml \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --verbose --dry-run
```

## Performance Optimization

### Resource Limits

Set appropriate resource limits:

```bash
docker run --rm \
  --memory 512m \
  --cpus 0.5 \
  --memory-swap 1g \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config config.yaml
```

### Layer Caching

Optimize Docker layer caching:

```dockerfile
FROM ghcr.io/baldator/iac-recert-engine:latest

# Copy configuration first (changes infrequently)
COPY config.yaml /config.yaml

# Copy scripts (changes more frequently)
COPY scripts/ /scripts/

# Install dependencies
RUN apk add --no-cache curl jq

CMD ["run", "--config", "/config.yaml"]
```

### Multi-Architecture

Build for multiple architectures:

```bash
# Build multi-arch image
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/myorg/ice:latest \
  --push .
```

## Production Deployment

### Health Checks

Implement health checks:

```yaml
# docker-compose.yml
services:
  ice:
    image: ghcr.io/baldator/iac-recert-engine:latest
    healthcheck:
      test: ["CMD", "ice", "run", "--dry-run", "--config", "/config.yaml"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Logging

Configure structured logging:

```bash
docker run --rm \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config config.yaml
```

### Monitoring

Monitor container metrics:

```bash
# Container stats
docker stats ice-container

# Logs
docker logs -f ice-container

# Events
docker events --filter container=ice-container
```

## Best Practices

### Image Management

- Use specific version tags instead of `latest`
- Regularly update base images for security
- Scan images for vulnerabilities
- Use multi-stage builds to reduce size

### Configuration Management

- Mount configs as read-only volumes
- Use ConfigMaps/Secrets in Kubernetes
- Validate configurations before deployment
- Keep configs version controlled

### Security

- Run as non-root user when possible
- Use read-only filesystems where applicable
- Limit network access with custom networks
- Regularly rotate authentication tokens

### Reliability

- Implement proper error handling and retries
- Use health checks and monitoring
- Set appropriate resource limits
- Test deployments in staging environments

## Next Steps

- [Command Line Interface](cli.md) - Direct binary usage
- [CI/CD Integration](ci-cd.md) - Automated workflows
- [Configuration Overview](../configuration/overview.md) - Configuration reference
