# Repository Settings

The repository configuration defines which Git repository ICE will scan and which provider to use for API interactions.

## Configuration Schema

```yaml
repository:
  url: string    # Required: Full repository URL
  provider: string # Required: git, azure, or gitlab
```

## URL Formats

### GitHub
```
https://github.com/organization/repository
https://github.com/user/repository
```

**Examples:**
```yaml
repository:
  url: "https://github.com/acme/infrastructure"
  provider: "github"
```

### Azure DevOps
```
https://dev.azure.com/organization/project/_git/repository
```

**Examples:**
```yaml
repository:
  url: "https://dev.azure.com/acme/Platform/_git/infrastructure"
  provider: "azure"
```

### GitLab
```
https://gitlab.com/group/project
https://gitlab.example.com/group/project  # Self-hosted
```

**Examples:**
```yaml
repository:
  url: "https://gitlab.com/acme/infrastructure"
  provider: "gitlab"
```

## Provider Detection

ICE can automatically detect the provider from the URL, but explicit configuration is recommended for clarity:

```yaml
# Explicit provider (recommended)
repository:
  url: "https://github.com/acme/infra"
  provider: "github"

# Auto-detection (works but less clear)
repository:
  url: "https://github.com/acme/infra"
  # provider will be auto-detected as "github"
```

## Supported Providers

### GitHub
- **URL Pattern**: `github.com`
- **Authentication**: Personal Access Tokens (PAT)
- **Features**: Full PR support, labels, reviewers, comments
- **Rate Limits**: 5,000 requests per hour for PATs

### Azure DevOps
- **URL Pattern**: `dev.azure.com`
- **Authentication**: Personal Access Tokens (PAT)
- **Features**: PR support, work item linking, policies
- **Rate Limits**: Based on organization settings

### GitLab
- **URL Pattern**: `gitlab.com` or custom domains
- **Authentication**: Personal Access Tokens or Project Tokens
- **Features**: Merge request support, approvals, comments
- **Rate Limits**: Configurable per instance

## Multiple Repositories

For organization-wide scanning, create separate configuration files:

```yaml
# config-repo1.yaml
repository:
  url: "https://github.com/acme/infra-prod"
  provider: "github"

# config-repo2.yaml
repository:
  url: "https://github.com/acme/infra-staging"
  provider: "github"
```

Run separately:
```bash
ice run --config config-repo1.yaml
ice run --config config-repo2.yaml
```

## Repository Permissions

Ensure your authentication token has appropriate permissions:

### GitHub
- `repo`: Full repository access
- `workflow`: Access to GitHub Actions (if using CI/CD integration)
- `read:org`: Organization access (for organization-wide features)

### Azure DevOps
- `Code (read, write)`: Repository access
- `Pull Request Threads (read, write)`: PR comment access
- `Project and Team (read)`: Project metadata access

### GitLab
- `api`: Full API access
- `read_repository`: Repository read access
- `write_repository`: Repository write access

## Branch Considerations

ICE operates on the default branch unless specified otherwise:

```yaml
global:
  default_base_branch: "main"  # or "master", "develop", etc.
```

All PRs will be created against this base branch.

## Private Repositories

Private repositories require proper authentication. Ensure:

1. Token has access to the private repository
2. Repository URL is correct
3. Token is not expired
4. Token has necessary scopes

## Troubleshooting

### "Repository not found"
- Verify the URL is correct
- Ensure the repository exists and is accessible
- Check token permissions for private repositories

### "Invalid provider"
- Ensure provider matches URL domain
- Use supported providers: `github`, `azure`, `gitlab`

### "Authentication failed"
- Verify token is valid and not expired
- Check token has correct scopes
- Ensure token environment variable is set

## Best Practices

- Use HTTPS URLs for security
- Store repository URLs in configuration files
- Document repository ownership and access requirements
- Test configurations with dry-run mode first
- Use descriptive repository names in configurations
