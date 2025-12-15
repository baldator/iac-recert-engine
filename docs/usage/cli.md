# Command Line Interface

ICE provides a comprehensive command-line interface for running recertification processes, managing configurations, and integrating with various environments. This guide covers practical usage patterns and advanced scenarios.

## Installation

### Pre-built Binaries

Download the latest release from GitHub:

```bash
# Linux/macOS
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/

# Windows
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-windows-amd64.exe -o ice.exe
```

### From Source

```bash
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine
go build -o ice ./cmd/ice
sudo mv ice /usr/local/bin/
```

### Verify Installation

```bash
ice version
# ICE v1.0.0

ice --help
# Display help information
```

## Basic Usage

### Configuration Setup

Create a basic configuration file:

```yaml
# config.yaml
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
assignment:
  strategy: "static"
  fallback_assignees: ["devops-team"]
```

Set your authentication token:

```bash
export GITHUB_TOKEN=your_github_token_here
```

### First Run

Run ICE with dry-run to preview changes:

```bash
ice run --dry-run --config config.yaml
```

Expected output:
```
2025-12-14T10:30:00Z INFO Starting recertification run (DRY RUN)
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:05Z INFO Found 25 files requiring recertification
2025-12-14T10:30:06Z INFO Would create 3 file groups
2025-12-14T10:30:07Z INFO Would create PR for terraform-prod (15 files)
2025-12-14T10:30:08Z INFO Would create PR for k8s-manifests (8 files)
2025-12-14T10:30:09Z INFO Would create PR for policies (2 files)
2025-12-14T10:30:10Z INFO Recertification run completed successfully (DRY RUN)
```

### Production Run

Execute the actual recertification:

```bash
ice run --config config.yaml
```

## Configuration Management

### Configuration File Locations

ICE searches for configuration files in this order:

1. Explicit path: `--config /path/to/config.yaml`
2. Environment variable: `ICE_CONFIG_PATH=/path/to/config.yaml`
3. Current directory: `./ice.yaml`, `./ice.yml`
4. Home directory: `~/.ice.yaml`, `~/.ice.yml`
5. Default location: `~/.config/ice/config.yaml`

### Environment Variables

Use environment variables for sensitive data:

```bash
export GITHUB_TOKEN=ghp_your_token
export ICE_CONFIG_PATH=/etc/ice/config.yaml

ice run
```

### Multiple Configurations

Maintain separate configs for different environments:

```bash
# Development
ice run --config config.dev.yaml

# Staging
ice run --config config.staging.yaml

# Production
ice run --config config.prod.yaml
```

## Advanced Usage Patterns

### Repository Override

Override repository settings without modifying config:

```bash
ice run --repo-url https://github.com/other-org/other-repo --config config.yaml
```

### Verbose Logging

Enable detailed logging for troubleshooting:

```bash
ice run --verbose --config config.yaml
```

Sample verbose output:
```
2025-12-14T10:30:00Z INFO Starting recertification run
2025-12-14T10:30:00Z DEBUG Loading configuration from config.yaml
2025-12-14T10:30:00Z DEBUG Configuration validated successfully
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:01Z DEBUG Scanning pattern: terraform-prod
2025-12-14T10:30:02Z DEBUG Found 15 matching files for pattern terraform-prod
2025-12-14T10:30:02Z DEBUG Checking recertification status for terraform/prod/main.tf
2025-12-14T10:30:03Z DEBUG File needs recertification (last: 2025-06-01, required: 2025-09-01)
```

### Selective Processing

Use dry-run to test specific scenarios:

```bash
# Test configuration without creating PRs
ice run --dry-run --config config.yaml

# Override settings for testing
ice run --dry-run --repo-url https://github.com/test/repo --config config.yaml
```

## Integration Patterns

### Shell Scripts

Create reusable scripts:

```bash
#!/bin/bash
# recertify.sh

set -e

echo "Starting recertification for $REPO_URL"

export GITHUB_TOKEN="$ICE_GITHUB_TOKEN"

ice run --config /etc/ice/config.yaml --repo-url "$REPO_URL"

echo "Recertification completed"
```

Make executable and run:

```bash
chmod +x recertify.sh
./recertify.sh
```

### Cron Jobs

Schedule regular recertification:

```bash
# Daily at 2 AM
0 2 * * * /usr/local/bin/ice run --config /etc/ice/config.yaml >> /var/log/ice.log 2>&1

# Weekly on Sunday
0 2 * * 0 /usr/local/bin/ice run --config /etc/ice/config.yaml
```

### Systemd Service

Create a systemd service for reliable execution:

```ini
# /etc/systemd/system/ice-recert.service
[Unit]
Description=ICE Recertification Service
After=network.target

[Service]
Type=oneshot
User=ice
Environment=GITHUB_TOKEN=your_token_here
Environment=ICE_CONFIG_PATH=/etc/ice/config.yaml
ExecStart=/usr/local/bin/ice run
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable ice-recert
sudo systemctl start ice-recert
```

Create a timer for scheduled execution:

```ini
# /etc/systemd/system/ice-recert.timer
[Unit]
Description=Run ICE recertification weekly

[Timer]
OnCalendar=weekly
Persistent=true

[Install]
WantedBy=timers.target
```

```bash
sudo systemctl enable ice-recert.timer
sudo systemctl start ice-recert.timer
```

## Troubleshooting

### Common Issues

**Configuration not found:**
```bash
# Check current directory
ls -la *.yaml

# Specify explicit path
ice run --config ./config.yaml

# Check environment
echo $ICE_CONFIG_PATH
```

**Authentication failed:**
```bash
# Verify token
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check token permissions
# Visit https://github.com/settings/tokens
```

**Permission denied:**
```bash
# Check file permissions
ls -la $(which ice)

# Add to PATH
export PATH=$PATH:/usr/local/bin
```

### Debug Mode

Enable maximum logging:

```bash
ice run --verbose --config config.yaml 2>&1 | tee ice.log
```

### Health Checks

Verify ICE can access the repository:

```bash
# Test repository access
ice run --dry-run --config config.yaml

# Check for configuration errors
ice run --dry-run --verbose --config config.yaml 2>&1 | head -20
```

## Performance Tuning

### Large Repositories

For repositories with many files:

```yaml
global:
  max_concurrent_prs: 3  # Reduce API rate limiting
```

```bash
# Adjust Go garbage collection
GOGC=50 ice run --config config.yaml
```

### Memory Optimization

Monitor memory usage:

```bash
# Run with memory profiling
ice run --config config.yaml &
ICE_PID=$!
sleep 10
go tool pprof http://localhost:6060/debug/pprof/heap
kill $ICE_PID
```

### Timeout Handling

Prevent hanging processes:

```bash
# Timeout after 1 hour
timeout 3600 ice run --config config.yaml
```

## Security Best Practices

### Token Management

Store tokens securely:

```bash
# Use environment variables
export GITHUB_TOKEN="$(aws secretsmanager get-secret-value --secret-id ice-tokens --query SecretString --output text | jq -r .github_token)"

# Or use a secrets manager
export GITHUB_TOKEN="$(vault kv get -field=token secret/ice/github)"
```

### Configuration Security

Protect configuration files:

```bash
# Set proper permissions
chmod 600 config.yaml

# Use secure directories
sudo mkdir -p /etc/ice
sudo chown ice:ice /etc/ice
sudo chmod 700 /etc/ice
```

### Audit Logging

Enable audit logging for compliance:

```yaml
audit:
  enabled: true
  storage: "file"
  config:
    directory: "/var/log/ice/audit"
```

## Advanced Scenarios

### Multi-Repository Processing

Process multiple repositories:

```bash
#!/bin/bash
repos=(
    "https://github.com/org/repo1"
    "https://github.com/org/repo2"
    "https://github.com/org/repo3"
)

for repo in "${repos[@]}"; do
    echo "Processing $repo"
    ice run --repo-url "$repo" --config config.yaml
    sleep 60  # Rate limiting
done
```

### Conditional Execution

Run only when changes are needed:

```bash
#!/bin/bash
if ice run --dry-run --config config.yaml | grep -q "Found.*requiring recertification"; then
    echo "Changes needed, running recertification"
    ice run --config config.yaml
else
    echo "No changes needed"
fi
```

### Error Recovery

Implement retry logic:

```bash
#!/bin/bash
max_attempts=3
attempt=1

while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt of $max_attempts"
    if ice run --config config.yaml; then
        echo "Success!"
        exit 0
    else
        echo "Failed, retrying in 5 minutes..."
        sleep 300
        ((attempt++))
    fi
done

echo "All attempts failed"
exit 1
```

## Next Steps

- [Docker Usage](docker.md) - Containerized deployment
- [CI/CD Integration](ci-cd.md) - Automated workflows
- [Configuration Overview](../configuration/overview.md) - Configuration reference
