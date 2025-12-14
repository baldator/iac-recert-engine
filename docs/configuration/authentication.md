# Authentication

ICE requires authentication to access Git repositories and create pull requests. Authentication is configured through personal access tokens (PATs) stored in environment variables.

## Configuration Schema

```yaml
auth:
  provider: string    # Required: github, azure, or gitlab
  token_env: string   # Required: Environment variable name
```

## Supported Providers

### GitHub

**Token Type**: Personal Access Token (PAT) or GitHub App token

**Required Scopes**:
- `repo`: Full repository access
- `workflow`: Access to GitHub Actions (if using CI/CD integration)
- `read:org`: Organization access (for organization-wide features)

**Configuration Example**:
```yaml
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"
```

**Token Creation**:
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with required scopes
3. Set environment variable: `export GITHUB_TOKEN=your_token_here`

### Azure DevOps

**Token Type**: Personal Access Token (PAT)

**Required Scopes**:
- `Code (read, write)`: Repository access
- `Pull Request Threads (read, write)`: PR comment access
- `Project and Team (read)`: Project metadata access

**Configuration Example**:
```yaml
auth:
  provider: "azure"
  token_env: "AZURE_DEVOPS_TOKEN"
```

**Token Creation**:
1. Go to Azure DevOps → User settings → Personal access tokens
2. Create token with Code scope (Read & Write)
3. Set environment variable: `export AZURE_DEVOPS_TOKEN=your_token_here`

### GitLab

**Token Type**: Personal Access Token or Project Access Token

**Required Scopes**:
- `api`: Full API access
- `read_repository`: Repository read access
- `write_repository`: Repository write access

**Configuration Example**:
```yaml
auth:
  provider: "gitlab"
  token_env: "GITLAB_TOKEN"
```

**Token Creation**:
1. Go to GitLab → User Settings → Access Tokens
2. Create personal access token with required scopes
3. Set environment variable: `export GITLAB_TOKEN=your_token_here`

## Environment Variable Setup

### Linux/macOS
```bash
export GITHUB_TOKEN=ghp_your_token_here
export AZURE_DEVOPS_TOKEN=your_azure_token
export GITLAB_TOKEN=your_gitlab_token
```

### Windows (Command Prompt)
```cmd
set GITHUB_TOKEN=ghp_your_token_here
set AZURE_DEVOPS_TOKEN=your_azure_token
set GITLAB_TOKEN=your_gitlab_token
```

### Windows (PowerShell)
```powershell
$env:GITHUB_TOKEN="ghp_your_token_here"
$env:AZURE_DEVOPS_TOKEN="your_azure_token"
$env:GITLAB_TOKEN="your_gitlab_token"
```

### Docker
```bash
docker run --env GITHUB_TOKEN=ghp_your_token_here ...
```

### Kubernetes
```yaml
env:
- name: GITHUB_TOKEN
  valueFrom:
    secretKeyRef:
      name: ice-secrets
      key: github-token
```

## Token Security Best Practices

### Token Storage
- **Never commit tokens to version control**
- Use secret management systems (Vault, AWS Secrets Manager, etc.)
- Rotate tokens regularly (GitHub recommends every 30 days)
- Use tokens with minimal required scopes

### Token Scope Minimization
- Create separate tokens for different environments
- Use fine-grained permissions when available
- Regularly audit token usage and permissions

### Environment Separation
```yaml
# Production
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN_PROD"

# Staging
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN_STAGING"
```

## Authentication Validation

ICE validates authentication on startup:

```bash
# Test authentication
ice run --config config.yaml --dry-run

# Check repository access
ice run --config config.yaml --verbose
```

Common authentication errors:
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: Insufficient token permissions
- `404 Not Found`: Repository not accessible or token lacks org access

## Troubleshooting

### "Authentication failed"
- Verify token is not expired
- Check token has correct scopes
- Ensure environment variable is set correctly
- Test token manually with API calls

### "Repository not found"
- Verify repository URL is correct
- Check token has access to the repository
- For private repositories, ensure token has `repo` scope

### "Rate limit exceeded"
- GitHub: 5,000 requests/hour for PATs
- Azure: Based on organization settings
- GitLab: Configurable per instance
- Implement request throttling or use multiple tokens

### Token Rotation
```bash
# Update token in environment
export GITHUB_TOKEN=new_token_here

# Restart ICE processes
# For long-running processes, implement hot-reload if supported
```

## Advanced Configuration

### Multiple Repositories
For organization-wide scanning with different tokens:

```yaml
# config-repo1.yaml
repository:
  url: "https://github.com/org/repo1"
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN_REPO1"

# config-repo2.yaml
repository:
  url: "https://github.com/org/repo2"
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN_REPO2"
```

### CI/CD Integration
```yaml
# GitHub Actions
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

# GitLab CI
variables:
  GITLAB_TOKEN: $GITLAB_TOKEN

# Azure Pipelines
variables:
  AZURE_DEVOPS_TOKEN: $(AZURE_DEVOPS_TOKEN)
```

## Next Steps

- [Repository Settings](repository.md) - Repository configuration options
- [Patterns](patterns.md) - File pattern configuration
- [Global Configuration](../configuration/overview.md#global-configuration) - Additional global settings
