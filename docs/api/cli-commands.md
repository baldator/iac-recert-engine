# CLI Commands

ICE provides a command-line interface for running recertification processes and managing configurations. This document details all available commands, flags, and usage examples.

## Command Structure

```bash
ice [global-flags] <command> [command-flags] [arguments]
```

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--config`, `-c` | string | Path to configuration file (default: search in home directory) |
| `--help`, `-h` | - | Display help information |
| `--version`, `-v` | - | Display version information |

## Commands

### `ice run`

Run the recertification process against a repository.

```bash
ice run [flags]
```

**Description:**
Scans the repository for files requiring recertification, groups them according to the configured strategy, and creates pull requests for review.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | boolean | `false` | Run without creating PRs (preview mode) |
| `--repo-url` | string | - | Override repository URL from config |
| `--verbose`, `-v` | boolean | `false` | Enable verbose logging |
| `--config` | string | - | Path to configuration file |

**Examples:**

**Basic run:**
```bash
ice run
```

**Dry run (preview changes):**
```bash
ice run --dry-run
```

**Override repository:**
```bash
ice run --repo-url https://github.com/org/repo
```

**Verbose logging:**
```bash
ice run --verbose
```

**Custom config file:**
```bash
ice run --config /path/to/config.yaml
```

**Combined flags:**
```bash
ice run --dry-run --verbose --config staging-config.yaml
```

### `ice help`

Display help information for commands.

```bash
ice help [command]
```

**Examples:**

**General help:**
```bash
ice help
```

**Command-specific help:**
```bash
ice help run
```

### `ice version`

Display version information.

```bash
ice version
```

**Output:**
```
ICE v1.0.0
```

## Configuration File Resolution

ICE resolves configuration files in the following order:

1. **Explicit path:** `--config /path/to/file.yaml`
2. **Environment variable:** `ICE_CONFIG_PATH=/path/to/file.yaml`
3. **Current directory:** `./ice.yaml`, `./ice.yml`
4. **Home directory:** `~/.ice.yaml`, `~/.ice.yml`
5. **Default location:** `~/.config/ice/config.yaml`

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ICE_CONFIG_PATH` | Path to configuration file | `/etc/ice/config.yaml` |
| `GITHUB_TOKEN` | GitHub personal access token | `ghp_...` |
| `AZURE_DEVOPS_TOKEN` | Azure DevOps PAT | `abc123...` |
| `GITLAB_TOKEN` | GitLab PAT | `glpat-...` |

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `3` | Authentication error |
| `4` | Network error |
| `5` | Validation error |

## Command Examples

### Development Workflow

**Test configuration:**
```bash
ice run --dry-run --verbose
```

**Run against staging:**
```bash
ICE_CONFIG_PATH=config.staging.yaml ice run
```

**Debug specific repository:**
```bash
ice run --repo-url https://github.com/test/repo --dry-run --verbose
```

### Production Deployment

**Scheduled run:**
```bash
#!/bin/bash
cd /opt/ice
export GITHUB_TOKEN=$(cat /secrets/github-token)
ice run --config production.yaml
```

**Docker container:**
```bash
docker run --rm \
  -e GITHUB_TOKEN \
  -v $(pwd)/config.yaml:/config.yaml \
  ice:latest run --config /config.yaml
```

**Kubernetes CronJob:**
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ice-recert
spec:
  schedule: "0 2 * * 0"  # Weekly
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: ice
            image: ice:latest
            env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: ice-secrets
                  key: github-token
            command: ["ice", "run"]
```

### CI/CD Integration

**GitHub Actions:**
```yaml
name: Recertification
on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  recert:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Run ICE
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        ice run --config .ice/config.yaml
```

**GitLab CI:**
```yaml
recertification:
  image: ice:latest
  script:
    - ice run --config ice-config.yaml
  only:
    - schedules
  variables:
    GITLAB_TOKEN: $GITLAB_TOKEN
```

**Azure Pipelines:**
```yaml
schedules:
- cron: "0 2 * * 0"
  displayName: Weekly recertification
  branches:
    include:
    - main

steps:
- script: |
    ice run --config config.yaml
  env:
    AZURE_DEVOPS_TOKEN: $(AZURE_DEVOPS_TOKEN)
```

## Command Output

### Normal Execution

```
2025-12-14T10:30:00Z INFO Starting recertification run
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:05Z INFO Found 25 files requiring recertification
2025-12-14T10:30:06Z INFO Grouping files by pattern
2025-12-14T10:30:07Z INFO Created 3 file groups
2025-12-14T10:30:08Z INFO Creating pull request for terraform-prod (15 files)
2025-12-14T10:30:10Z INFO Pull request created: https://github.com/org/repo/pull/123
2025-12-14T10:30:11Z INFO Creating pull request for k8s-manifests (8 files)
2025-12-14T10:30:13Z INFO Pull request created: https://github.com/org/repo/pull/124
2025-12-14T10:30:14Z INFO Creating pull request for policies (2 files)
2025-12-14T10:30:16Z INFO Pull request created: https://github.com/org/repo/pull/125
2025-12-14T10:30:17Z INFO Recertification run completed successfully
```

### Dry Run Output

```
2025-12-14T10:30:00Z INFO Starting recertification run (DRY RUN)
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:05Z INFO Found 25 files requiring recertification
2025-12-14T10:30:06Z INFO Grouping files by pattern
2025-12-14T10:30:07Z INFO Would create 3 file groups
2025-12-14T10:30:08Z INFO Would create PR for terraform-prod (15 files):
  - terraform/prod/main.tf
  - terraform/prod/variables.tf
  - ...
2025-12-14T10:30:09Z INFO Would create PR for k8s-manifests (8 files):
  - k8s/prod/deployment.yaml
  - k8s/prod/service.yaml
  - ...
2025-12-14T10:30:10Z INFO Would create PR for policies (2 files):
  - policies/security.rego
  - policies/network.rego
2025-12-14T10:30:11Z INFO Recertification run completed successfully (DRY RUN)
```

### Verbose Logging

```
2025-12-14T10:30:00Z INFO Starting recertification run
2025-12-14T10:30:00Z DEBUG Loading configuration from config.yaml
2025-12-14T10:30:00Z DEBUG Configuration validated successfully
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:01Z DEBUG Scanning pattern: terraform-prod
2025-12-14T10:30:02Z DEBUG Found 15 matching files for pattern terraform-prod
2025-12-14T10:30:02Z DEBUG Checking recertification status for terraform/prod/main.tf
2025-12-14T10:30:03Z DEBUG File terraform/prod/main.tf needs recertification (last: 2025-06-01, required: 2025-09-01)
2025-12-14T10:30:05Z INFO Found 25 files requiring recertification
2025-12-14T10:30:06Z INFO Grouping files by pattern
2025-12-14T10:30:07Z DEBUG Created group terraform-prod with 15 files
2025-12-14T10:30:07Z DEBUG Created group k8s-manifests with 8 files
2025-12-14T10:30:07Z DEBUG Created group policies with 2 files
2025-12-14T10:30:07Z INFO Created 3 file groups
```

## Error Handling

### Configuration Errors

```
Error: config validation failed: patterns[0]: recertification_days must be >= 1
```

### Authentication Errors

```
Error: authentication failed: 401 Unauthorized
Check that your token is valid and has the required permissions.
```

### Network Errors

```
Error: failed to connect to repository: dial tcp: lookup github.com: no such host
Check your network connection and DNS resolution.
```

### Rate Limiting

```
Error: API rate limit exceeded
Retry after: 2025-12-14T11:30:00Z
```

## Performance Tuning

### Memory Usage

For large repositories, adjust Go garbage collection:

```bash
GOGC=50 ice run  # Reduce GC frequency
```

### Concurrent Operations

Control concurrency with configuration:

```yaml
global:
  max_concurrent_prs: 3  # Reduce for API rate limits
```

### Timeout Handling

ICE handles timeouts automatically but can be configured:

```bash
timeout 3600 ice run  # 1 hour timeout
```

## Troubleshooting

### Command Not Found

```bash
# Check if ICE is installed
which ice

# Add to PATH
export PATH=$PATH:/usr/local/bin

# Or use full path
/usr/local/bin/ice run
```

### Permission Denied

```bash
# Check file permissions
ls -la $(which ice)

# Run with sudo if installed system-wide
sudo ice run
```

### Configuration Not Found

```bash
# Check current directory
ls -la *.yaml

# Specify explicit path
ice run --config ./config.yaml

# Check environment
echo $ICE_CONFIG_PATH
```

### Token Issues

```bash
# Test token validity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check token scopes
# Visit https://github.com/settings/tokens
```

## Next Steps

- [Configuration Schema](configuration-schema.md) - Complete configuration reference
- [Plugin Interface](plugin-interface.md) - Plugin development API
- [Quick Start](../../quick-start.md) - Getting started guide
