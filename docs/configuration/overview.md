# Configuration Overview

ICE is configured through a YAML file that defines repositories, patterns, strategies, and other settings. This guide provides a comprehensive overview of the configuration schema and options.

## Configuration File Structure

The configuration file follows this high-level structure:

```yaml
version: "1.0"

repository:
  # Repository settings

auth:
  # Authentication settings

global:
  # Global options

patterns:
  # File patterns and rules

pr_strategy:
  # Pull request grouping strategy

assignment:
  # Reviewer assignment strategy

plugins:
  # Plugin configurations

pr_template:
  # PR template settings

audit:
  # Audit logging settings
```

## Configuration Loading

ICE supports multiple ways to load configuration:

### File-based Configuration
```bash
ice run --config config.yaml
ice run --config /path/to/config.yaml
```

### Environment Variable Override
```bash
export ICE_CONFIG_PATH=config.yaml
ice run
```

### Inline Configuration (Limited)
```bash
ice run --repo-url https://github.com/org/repo --token-env GITHUB_TOKEN
```

## Core Configuration Sections

### Version
Specifies the configuration schema version. Currently supports `"1.0"`.

```yaml
version: "1.0"
```

### Repository Configuration
Defines the target repository and provider.

```yaml
repository:
  url: "https://github.com/your-org/your-repo"
  provider: "github"  # github, azure, or gitlab
```

### Authentication Configuration
Configures how ICE authenticates with the Git provider.

```yaml
auth:
  provider: "github"  # Must match repository.provider
  token_env: "GITHUB_TOKEN"  # Environment variable containing the token
```

### Global Configuration
Global settings that affect all operations.

```yaml
global:
  dry_run: false  # Set to true to preview changes without creating PRs
  verbose_logging: true  # Enable detailed logging
  max_concurrent_prs: 5  # Maximum PRs to create in one run
  default_base_branch: "main"  # Default branch for PRs
```

### Patterns Configuration
File patterns define which files to scan and their recertification rules.

```yaml
patterns:
  - name: "terraform-prod"
    description: "Production Terraform configurations"
    paths:
      - "terraform/prod/**/*.tf"
    exclude:
      - "**/*.test.tf"
    recertification_days: 180
    enabled: true
    decorator: "# Last Recertification: {timestamp}\n"
```

### PR Strategy Configuration
Controls how files are grouped into pull requests.

```yaml
pr_strategy:
  type: "per_pattern"  # per_file, per_pattern, per_committer, single_pr, plugin
  max_files_per_pr: 50  # Optional limit
  plugin_name: "custom-grouping"  # For plugin strategy
```

### Assignment Configuration
Defines how reviewers and assignees are determined.

```yaml
assignment:
  strategy: "composite"  # static, last_committer, plugin, composite
  rules:
    - pattern: "terraform/prod/**"
      strategy: "static"
      fallback_assignees: ["infra-team"]
  fallback_assignees:
    - "devops-team"
```

### Plugin Configuration
Configures custom plugins for extended functionality.

```yaml
plugins:
  cmdb_assignment:
    enabled: true
    type: "assignment"
    module: "cmdb"
    config:
      api_url: "https://cmdb.example.com"
      api_key: "${CMDB_API_KEY}"
```

### PR Template Configuration
Customizes the content and format of created pull requests.

```yaml
pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true
  custom_instructions: |
    Please review these files for security and compliance.
    Ensure all governance controls are applied.
```

### Audit Configuration
Configures audit logging and storage.

```yaml
audit:
  enabled: true
  storage: "file"  # file or s3
  config:
    path: "audit.log"  # For file storage
    # bucket: "audit-logs"  # For S3 storage
    # prefix: "iac-recert/"  # For S3 storage
```

## Configuration Validation

ICE validates configuration on startup and reports errors clearly:

```bash
# Validate configuration without running
ice config validate --config config.yaml

# Run with validation (default behavior)
ice run --config config.yaml
```

Common validation errors:
- Missing required fields
- Invalid provider names
- Malformed URLs
- Invalid pattern syntax
- Plugin configuration errors

## Environment Variables

Configuration supports environment variable substitution:

```yaml
repository:
  url: "${REPO_URL}"

auth:
  token_env: "${TOKEN_ENV_VAR}"

plugins:
  cmdb:
    config:
      api_key: "${CMDB_API_KEY}"
```

## Configuration Examples

### Minimal Configuration
```yaml
version: "1.0"
repository:
  url: "https://github.com/org/repo"
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
  fallback_assignees: ["team"]
```

### Advanced Configuration
See `example.config.yaml` in the repository root for a comprehensive example with all options.

## Configuration Best Practices

### Organization
- Use descriptive pattern names
- Group related patterns logically
- Document complex configurations with comments

### Security
- Store tokens in environment variables or secret managers
- Use minimal required token scopes
- Rotate tokens regularly
- Avoid committing sensitive configuration

### Performance
- Limit concurrent PR creation to avoid rate limits
- Use appropriate file patterns to avoid scanning irrelevant files
- Configure reasonable recertification periods

### Maintainability
- Use version control for configuration files
- Document custom plugins and their requirements
- Test configuration changes in dry-run mode first
- Keep audit logs for compliance tracking

## Next Steps

- [Repository Settings](repository.md) - Detailed repository configuration
- [Authentication](authentication.md) - Authentication options and setup
- [Patterns](patterns.md) - File pattern configuration
- [PR Strategies](pr-strategies.md) - Pull request grouping options
- [Assignment Strategies](assignment-strategies.md) - Reviewer assignment configuration
