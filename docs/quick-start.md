# Quick Start

This guide will get you up and running with IaC Recertification Engine in under 15 minutes. We'll set up a basic configuration and run your first recertification scan.

## Prerequisites

Before starting, ensure you have:

- **Go 1.24+** installed
- **Git** installed and configured
- Access to a Git repository (GitHub, Azure DevOps, or GitLab)
- A personal access token for your Git provider

## Step 1: Install ICE

### Option A: Download Pre-built Binary

```bash
# Download the latest release from GitHub
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
chmod +x ice
sudo mv ice /usr/local/bin/
```

### Option B: Build from Source

```bash
# Clone the repository
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine

# Build the binary
go build -o ice ./cmd/ice
```

### Option C: Use Docker

```bash
# Pull the Docker image
docker pull ghcr.io/baldator/iac-recert-engine:latest
```

## Step 2: Create Configuration

Create a `config.yaml` file in your project directory:

```yaml
version: "1.0"

repository:
  url: "https://github.com/your-org/your-repo"
  provider: "github"

auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"

global:
  dry_run: true  # Set to false when ready to create real PRs
  verbose_logging: true

patterns:
  - name: "terraform"
    description: "Terraform configurations"
    paths:
      - "**/*.tf"
    recertification_days: 90
    enabled: true

pr_strategy:
  type: "per_pattern"

assignment:
  strategy: "static"
  fallback_assignees:
    - "your-username"

pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true
```

## Step 3: Set Environment Variables

Export your Git provider token:

```bash
# For GitHub
export GITHUB_TOKEN="your_personal_access_token_here"

# For Azure DevOps
export AZURE_DEVOPS_TOKEN="your_pat_here"

# For GitLab
export GITLAB_TOKEN="your_token_here"
```

## Step 4: Run Your First Scan

### Using the Binary

```bash
# Run in dry-run mode (no PRs created)
./ice run --config config.yaml

# Or specify config inline
./ice run --repo-url https://github.com/your-org/your-repo --token-env GITHUB_TOKEN
```

### Using Docker

```bash
docker run --rm -v $(pwd):/app -e GITHUB_TOKEN \
  ghcr.io/baldator/iac-recert-engine:latest \
  run --config /app/config.yaml
```

## Step 5: Review the Output

ICE will output detailed information about the scan:

```
2025/01/15 10:30:15 INFO Scanning repository: https://github.com/your-org/your-repo
2025/01/15 10:30:16 INFO Found 15 files matching pattern 'terraform'
2025/01/15 10:30:17 INFO 3 files require recertification
2025/01/15 10:30:18 INFO DRY RUN: Would create PR '🔄 Recertification: terraform' with 3 files
2025/01/15 10:30:18 INFO Scan completed successfully
```

## Step 6: Create Your First PR

Once you're satisfied with the dry run:

1. **Remove dry_run: true** from your config
2. **Re-run the command**
3. **Check your repository** for the new pull request

The PR will include:
- A clear title and description
- List of files requiring review
- Certification checklist
- Links to relevant documentation

## Next Steps

### Customize Your Configuration

- **Add more patterns** for different file types (CloudFormation, Kubernetes, etc.)
- **Adjust recertification periods** based on your compliance requirements
- **Configure assignment strategies** to automatically assign reviewers
- **Set up plugins** for advanced assignment logic

### Integrate with CI/CD

Add ICE to your CI/CD pipeline for automated recertification:

```yaml
# GitHub Actions example
name: IaC Recertification
on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday
  workflow_dispatch:

jobs:
  recertify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run IaC Recertification
        run: |
          docker run --rm -v ${{ github.workspace }}:/app \
            -e GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }} \
            ghcr.io/baldator/iac-recert-engine:latest \
            run --config /app/config.yaml
```

### Explore Advanced Features

- **Plugin system** for custom assignment and filtering logic
- **Audit logging** for compliance reporting
- **Multi-repository support** for organization-wide governance
- **Integration with external systems** via webhooks

## Troubleshooting

### Common Issues

**"Repository not found" error**
- Verify the repository URL is correct
- Ensure your token has appropriate permissions
- Check if the repository is private and your token has access

**"No files found" message**
- Review your pattern configurations
- Ensure file paths match your repository structure
- Check exclude patterns aren't filtering out all files

**"Authentication failed" error**
- Verify your token is valid and not expired
- Ensure the token has the required scopes
- Check the token environment variable name matches your config

### Getting Help

- Check the [Troubleshooting](troubleshooting/common-issues.md) section
- Review the [Configuration](configuration/overview.md) documentation
- Open an issue on GitHub for bugs or feature requests

## What's Next?

Now that you have ICE running, explore:

- [Full Configuration Guide](configuration/overview.md) for advanced options
- [CI/CD Integration](usage/ci-cd.md) for automated workflows
- [Plugin Development](api/plugin-interface.md) for custom extensions
- [Contributing](development/contributing.md) to the project
