# Configuration Schema

This document provides the complete YAML schema reference for ICE configuration files. All configuration options are documented with their types, defaults, and usage examples.

## Complete Schema

```yaml
# ICE Configuration Schema v1.0
version: "1.0"                    # string, required: Configuration schema version

repository:                        # object, required: Repository settings
  url: string                     # required: Full repository URL
  provider: string                # required: github, azure, or gitlab

auth:                             # object, required: Authentication settings
  provider: string                # required: github, azure, or gitlab
  token_env: string               # required: Environment variable containing token

global:                           # object, optional: Global settings
  dry_run: boolean                # optional, default: false
  verbose_logging: boolean        # optional, default: false
  max_concurrent_prs: integer     # optional, default: 5, min: 1
  default_base_branch: string     # optional, default: "main"

patterns:                         # array, required: File patterns
  - name: string                  # required: Unique pattern identifier
    description: string           # optional: Human-readable description
    paths:                        # array, required: Glob patterns to include
      - string
    exclude:                      # array, optional: Glob patterns to exclude
      - string
    recertification_days: integer # required: Days before recertification needed, min: 1
    enabled: boolean              # optional, default: true
    decorator: string             # optional: Text to add when recertifying

pr_strategy:                      # object, optional: PR grouping strategy
  type: string                    # required: per_file, per_pattern, per_committer, single_pr, plugin
  max_files_per_pr: integer       # optional: Maximum files per PR
  plugin_name: string             # optional: Plugin name for plugin strategy

assignment:                       # object, optional: Reviewer assignment strategy
  strategy: string                # required: static, last_committer, plugin, composite
  plugin_name: string             # optional: Plugin name for plugin strategy
  rules:                          # array, optional: Rules for composite strategy
    - pattern: string             # required: File pattern to match
      strategy: string            # required: static, last_committer, plugin
      plugin: string              # optional: Plugin name if strategy is plugin
      fallback_assignees:         # array, optional: Fallback assignees for this rule
        - string
  fallback_assignees:             # array, required: Default assignees
    - string

plugins:                          # object, optional: Plugin configurations
  plugin_name:                    # object: Plugin-specific settings
    enabled: boolean              # required: Enable/disable plugin
    type: string                  # required: assignment, filter, strategy
    module: string                # required: Plugin module name
    config:                       # object, optional: Plugin-specific configuration
      key: value

pr_template:                      # object, optional: PR template settings
  title: string                   # required: PR title template
  include_file_list: boolean      # optional, default: true
  include_checklist: boolean      # optional, default: true
  custom_instructions: string     # optional: Additional instructions

audit:                            # object, optional: Audit logging settings
  enabled: boolean                # optional, default: false
  storage: string                 # required: file, s3
  config:                         # object, optional: Storage-specific configuration
    directory: string             # for file storage
    bucket: string                # for S3 storage
    prefix: string                # for S3 storage

schedule:                         # object, optional: Cron scheduling (for reference)
  enabled: boolean                # optional, default: false
  cron: string                    # required: Cron expression
```

## Field Reference

### Version

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Configuration schema version. Currently supports `"1.0"`. |

### Repository

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repository.url` | string | Yes | Full repository URL (HTTPS format) |
| `repository.provider` | string | Yes | Git provider: `github`, `azure`, `gitlab` |

**Examples:**
```yaml
repository:
  url: "https://github.com/organization/repository"
  provider: "github"
```

### Authentication

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `auth.provider` | string | Yes | Must match `repository.provider` |
| `auth.token_env` | string | Yes | Environment variable containing access token |

**Examples:**
```yaml
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"
```

### Global Settings

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `global.dry_run` | boolean | No | `false` | Run without creating PRs |
| `global.verbose_logging` | boolean | No | `false` | Enable detailed logging |
| `global.max_concurrent_prs` | integer | No | `5` | Maximum concurrent PRs (min: 1) |
| `global.default_base_branch` | string | No | `"main"` | Default branch for PRs |

### Patterns

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `patterns[].name` | string | Yes | Unique pattern identifier |
| `patterns[].description` | string | No | Human-readable description |
| `patterns[].paths[]` | string | Yes | Glob patterns to include |
| `patterns[].exclude[]` | string | No | Glob patterns to exclude |
| `patterns[].recertification_days` | integer | Yes | Days before recertification (min: 1) |
| `patterns[].enabled` | boolean | No | Enable/disable pattern (default: true) |
| `patterns[].decorator` | string | No | Text to add when recertifying |

### PR Strategy

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pr_strategy.type` | string | Yes | Grouping strategy |
| `pr_strategy.max_files_per_pr` | integer | No | Maximum files per PR |
| `pr_strategy.plugin_name` | string | No | Plugin name for plugin strategy |

**Valid Values for `pr_strategy.type`:**
- `per_file`: One PR per file
- `per_pattern`: Group by pattern
- `per_committer`: Group by last committer
- `single_pr`: All files in one PR
- `plugin`: Custom plugin logic

### Assignment Strategy

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `assignment.strategy` | string | Yes | Assignment strategy |
| `assignment.plugin_name` | string | No | Plugin name for plugin strategy |
| `assignment.rules[]` | object | No | Rules for composite strategy |
| `assignment.fallback_assignees[]` | string | Yes | Default assignees |

**Valid Values for `assignment.strategy`:**
- `static`: Same assignees for all PRs
- `last_committer`: Assign to last file modifier
- `plugin`: Custom plugin logic
- `composite`: Pattern-based rules

### Assignment Rules (for composite strategy)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `assignment.rules[].pattern` | string | Yes | File pattern to match |
| `assignment.rules[].strategy` | string | Yes | Strategy for this pattern |
| `assignment.rules[].plugin` | string | No | Plugin name if strategy is plugin |
| `assignment.rules[].fallback_assignees[]` | string | No | Fallback assignees for this rule |

### Plugins

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `plugins.{name}.enabled` | boolean | Yes | Enable/disable plugin |
| `plugins.{name}.type` | string | Yes | Plugin type |
| `plugins.{name}.module` | string | Yes | Plugin module name |
| `plugins.{name}.config` | object | No | Plugin-specific configuration |

**Valid Values for `plugins.{name}.type`:**
- `assignment`: Assignment plugins
- `filter`: Filter plugins
- `strategy`: Strategy plugins

### PR Template

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `pr_template.title` | string | Yes | - | PR title template |
| `pr_template.include_file_list` | boolean | No | `true` | Include file list in PR |
| `pr_template.include_checklist` | boolean | No | `true` | Include checklist in PR |
| `pr_template.custom_instructions` | string | No | - | Additional instructions |

### Audit Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `audit.enabled` | boolean | No | Enable audit logging (default: false) |
| `audit.storage` | string | Yes | Storage type: `file`, `s3` |
| `audit.config` | object | No | Storage-specific configuration |

**File Storage Config:**
```yaml
audit:
  config:
    directory: "./audit"  # Directory for audit files
```

**S3 Storage Config:**
```yaml
audit:
  config:
    bucket: "my-audit-bucket"
    prefix: "iac-recert/"  # Optional S3 prefix
```

### Schedule Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schedule.enabled` | boolean | No | Enable scheduled runs (default: false) |
| `schedule.cron` | string | Yes | Cron expression for scheduling |

## Environment Variables

Configuration supports environment variable substitution using `${VAR_NAME}` syntax:

```yaml
repository:
  url: "${REPO_URL}"

auth:
  token_env: "${TOKEN_ENV_VAR}"

plugins:
  myplugin:
    config:
      api_key: "${API_KEY}"
```

## Validation Rules

### Required Fields
- `version`
- `repository` (with `url` and `provider`)
- `auth` (with `provider` and `token_env`)
- `patterns` (at least one pattern)
- `assignment.fallback_assignees`

### Field Constraints
- `version`: Must be `"1.0"`
- `repository.provider`: Must be `github`, `azure`, or `gitlab`
- `auth.provider`: Must match `repository.provider`
- `patterns[].recertification_days`: Must be ≥ 1
- `global.max_concurrent_prs`: Must be ≥ 1

### Pattern Validation
- Pattern names must be unique
- At least one path must be specified per pattern
- Glob patterns are validated for syntax

## Examples

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
assignment:
  strategy: "static"
  fallback_assignees: ["team"]
```

### Advanced Configuration
```yaml
version: "1.0"
repository:
  url: "https://github.com/org/infrastructure"
  provider: "github"
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"
global:
  dry_run: false
  verbose_logging: true
  max_concurrent_prs: 10
patterns:
  - name: "terraform-prod"
    description: "Production Terraform"
    paths: ["terraform/prod/**/*.tf"]
    exclude: ["**/modules/**"]
    recertification_days: 180
    decorator: "# Last Recertification: {timestamp}\n"
  - name: "kubernetes"
    paths: ["k8s/**/*.yaml"]
    recertification_days: 90
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 25
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/prod/**"
      strategy: "static"
      fallback_assignees: ["infra-team"]
    - pattern: "k8s/**"
      strategy: "last_committer"
  fallback_assignees: ["devops-team"]
plugins:
  servicenow:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      instance_url: "https://company.service-now.com"
      username: "${SNOW_USER}"
      password: "${SNOW_PASS}"
pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true
audit:
  enabled: true
  storage: "file"
  config:
    directory: "./audit"
```

## Schema Extensions

The configuration schema is designed to be extensible. New fields can be added to objects without breaking existing configurations. Unknown fields are ignored during parsing.

## Version Compatibility

- **v1.0**: Initial schema version
- Future versions will maintain backward compatibility
- Breaking changes will require version bumps

## Next Steps

- [CLI Commands](cli-commands.md) - Command-line interface reference
- [Plugin Interface](plugin-interface.md) - Plugin development API
- [Configuration Overview](../configuration/overview.md) - Configuration guide
