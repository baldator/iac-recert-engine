# Infrastructure as Code Recertification Tool – Specification (Go Implementation)
## 1.1 Executive Summary

### Project Name: 

IaC Recertification Engine (ICE)

### Purpose: 

Automate recertification of Infrastructure as Code (IaC) by scanning repositories, determining which files are due for review based on git last modification date and configured thresholds, and opening pull/merge requests on common git platforms (GitHub, Azure DevOps, GitLab) to trigger review workflows.​

### Key Value Proposition:

- Enforce periodic recertification of IaC without manual tracking.​

- Integrate with existing Git PR/MR workflows and CI/CD (e.g. GitHub Actions).​

- Provide flexible grouping of changes (per file, per pattern, per committer, all-files, plugin-driven).

- Provide flexible assignment (static lists, last committer, CMDB-based via plugins).

- Offer auditability for compliance and governance around IaC.​

- Implemented as a Go service/CLI for portability and easy integration into pipelines.​

## 1.2 Problem Statement

### Challenges:

- Manual or ad‑hoc tracking of when IaC definitions were last reviewed.​

- Difficulty proving regular recertification for audits and compliance frameworks.​

- Scaling review processes across large multi-repo environments.

- Fragmented audit trails across ticketing, CI systems, and git.

### Target Users:

- Platform / DevOps / SRE teams owning IaC.

- Security, risk, and compliance teams requiring evidence of periodic code review.​

- Enterprise IT teams running GitHub, Azure DevOps, or GitLab.

## 1.3 Solution Overview

### ICE performs:

- Load configuration (YAML) describing patterns, thresholds, strategies, and plugins.

- Discover files matching configured patterns.

- For each file, query the git provider for last modification date and commit metadata.

- Compute whether recertification is due based on recertification_days.

- Group files into PRs according to configured strategy:

1. Per file

2. Per pattern

3. Per last committer

4. Single PR for all files

5. Plugin-defined grouping

- Resolve assignees and reviewers (static, last committer, plugin, composite).

- Create PRs/MRs on:

1. GitHub

2. Azure DevOps

3. GitLab

- Log an audit trail for all steps.

### Architecture Philosophies:

- Go 1.21+ binary, stateless core, config-driven.​

- CLI and optional long-running HTTP service.

- Plugin-based extensibility.

- Provider-agnostic GitProvider abstraction.

### 1.4 Core Features

### File Pattern Matching
**Example configuration**:
```yaml
patterns:
  - name: "terraform-aws"
    paths:
      - "terraform/aws/**/*.tf"
      - "!terraform/aws/modules/**"
    exclude:
      - "**/test*.tf"
      - "**/.terraform/**"
    recertification_days: 180

  - name: "kubernetes-manifests"
    paths:
      - "k8s/**/*.yaml"
    recertification_days: 90
```
**Behaviour**:
*   Uses glob semantics with `**` via a Go glob library such as `doublestar`.
*   Includes and excludes applied in order.
*   Per-pattern recertification thresholds.

### Git History Analysis
For each matched file:
*   Retrieve last commit affecting the file.
*   Get timestamp (UTC), author name/email, commit hash, commit message.
*   Use provider APIs efficiently (batch queries where supported).

**Edge cases**:
*   **Newly added files**: creation commit used.
*   **Deleted/moved files**: normally ignored unless configured otherwise.
*   **Binary files**: timestamps matter; content not analyzed.

### Recertification Logic
For each file:
*   Compute days since last modification.
*   Compare with `recertification_days`.
*   Mark as needing recertification when threshold reached.
*   Compute priority bucket:
    *   **Critical**: > 150% of threshold.
    *   **High**: 100–150%.
    *   **Medium**: 80–100%.
    *   **Low**: < 80%.
*   Outputs `RecertCheckResult` containing file metadata, days since, threshold, priority, `NeedsRecert`, and next due date.

## 1.5 PR Grouping Strategies
**Supported strategies**:
*   **Per file**: One PR per file; branch `recert/{pattern}/{file-hash}`.
*   **Per pattern**: One PR per pattern; branch `recert/{pattern}`.
*   **Per last committer**: Group by last commit author; branch `recert/{pattern}/{author}`.
*   **Single PR**: All files in one PR; branch `recert/all-files`.
*   **Plugin-based**: Custom Go logic for grouping (e.g., by team, environment, CMDB ownership).

**Constraints**:
*   Configurable `max_files_per_pr`.
*   Optional merging of small groups.
*   Validation to avoid files appearing in multiple groups.

## 1.6 PR Content and Templates
**Content**:
*   Title template, e.g., `🔄 Recertification: {pattern_name} ({file_count} files)`.
*   Description in Markdown:
    *   Summary and reason.
    *   Table of files (path, size, last modified, author).
    *   Certification checklist.
    *   Pattern and configuration reference.
    *   Run ID and timestamps.

**Config snippet**:
```text
pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true
  custom_instructions: |
    Please review and certify these infrastructure configurations.
    Ensure all security and governance controls are applied.
```
Simple string templating (replace placeholders like `{pattern_name}`, `{file_count}`) is sufficient.

## 1.7 Assignment Strategies
**Assignment strategies**:
*   **Static**: Direct assignees and reviewers.
*   **Last committer**: Use `FileInfo.CommitAuthor` to drive accountability.
*   **Plugin-based**: Consult CMDB or ownership service via plugin.
*   **Composite**: Path-based rules combining several strategies.

**Example**:
```text
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/prod/**"
      strategy: "plugin"
      plugin: "cmdb_assignment"
    - pattern: "k8s/**"
      strategy: "last_committer"
      fallback_assignees:
        - "k8s-team@company.com"
  fallback_assignees:
    - "devops@company.com"
```
**Plugins**:
*   **Input**: file list and metadata.
*   **Output**: `AssignmentResult` with assignees, reviewers, team, priority.

## 1.8 Git Provider Integrations

### Common GitProvider Interface (Go)
```go
type GitProvider interface {
    GetRepository(ctx context.Context, url string) (*Repository, error)
    GetLastModificationDate(ctx context.Context, filePath string) (time.Time, Commit, error)

    CreateBranch(ctx context.Context, name, baseRef string) error
    CreateCommit(ctx context.Context, branch, message string, changes []Change) (string, error)

    CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error)
    UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error
    ClosePullRequest(ctx context.Context, id string, reason string) error

    AssignPullRequest(ctx context.Context, id string, assignees []string) error
    RequestReviewers(ctx context.Context, id string, reviewers []string) error

    AddLabels(ctx context.Context, id string, labels []string) error
    AddComment(ctx context.Context, id string, comment string) error
}
```

### GitHub
*   Auth via PAT or GitHub App.
*   Use `go-github` or REST with `net/http`.
*   Supports labels, assignees, reviewers, comments, draft PRs.
*   Easy integration with GitHub Actions.

### Azure DevOps
*   Auth via PAT.
*   Use Azure DevOps REST API for git and PR endpoints.
*   Handle organization, project, repo from URLs.
*   Respect branch policies.

### GitLab
*   Auth via PAT or project/group tokens.
*   Support both gitlab.com and self-hosted instances.
*   Use GitLab REST API v4 to create merge requests, assign reviewers and labels.

### Provider Factory
*   Detects provider from `repository.provider` or URL.
*   Returns appropriate `GitProvider` implementation.

## 1.9 Plugin System
**Plugin types**:
*   AssignmentPlugin
*   FilterPlugin
*   TransformPlugin
*   HookPlugin
*   ValidatorPlugin

**Architecture**:
*   Plugins compiled into the binary as Go packages.
*   `PluginManager` loads by name from config.
*   Each plugin has `Init` and main methods (`Resolve`, `Filter` etc.).
*   Errors, panics, and timeouts result in fallbacks and audit entries.
*   **Example**: CMDB assignment plugin querying a CMDB HTTP API to map file path → owning team.

## 1.10 Configuration Schema (YAML)
**Root configuration (`.ice/config.yaml`)**:
```text
version: "1.0"

repository:
  url: "https://github.com/company/infrastructure"
  provider: "github"

auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"

global:
  dry_run: false
  verbose_logging: true
  max_concurrent_prs: 5
  default_base_branch: "main"

patterns:
  - name: "terraform-prod"
    description: "Production Terraform configurations"
    paths:
      - "terraform/prod/**/*.tf"
      - "!terraform/prod/modules/**"
    exclude:
      - "**/*.test.tf"
    recertification_days: 180
    enabled: true

pr_strategy:
  type: "per_pattern"

assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/prod/**"
      strategy: "plugin"
      plugin: "cmdb_assignment"
  fallback_assignees:
    - "devops@company.com"

plugins:
  cmdb_assignment:
    enabled: true
    type: "assignment"
    module: "cmdb"
    config:
      api_url: "${CMDB_API_URL}"
      api_key: "${CMDB_API_KEY}"

schedule:
  enabled: true
  cron: "0 2 * * 0"

pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true

audit:
  enabled: true
  storage: "s3"
  config:
    bucket: "company-audit-logs"
    prefix: "iac-recert/"
```

Go structs with yaml tags and validation enforce this schema.

## 1.11 Execution Flow
### CLI / Service:

CLI binary ice with commands run, dry-run, config validate, status.

Optional HTTP service providing /api/v1/recertify, /api/v1/status, /api/v1/audit.

### Run steps:
1. Load and validate config.
2. Setup logging, metrics, and audit.
3. Instantiate provider via factory.
4. Load plugins.
5. For each pattern:
   a. Scan files.
   b. Enrich via git history.
   c. Compute recertification status.
   d. Collect all files needing recertification.
   e. Group into PR groups.
   f. For each group:
      i. Resolve assignees.
      ii. Build PR template.
      iii. If not dry_run, create branch/commit/PR.
      iv. Log audit and send notifications where configured.​

## 1.12 Observability & Security
### Observability:

Structured logging using a Go logging library such as zap.​

Metrics: counts and latencies for scanning, PR creation, plugin exec, provider calls.

Audit logs stored in local files, S3, or other backends.​

### Security:

Tokens in env or secret managers, not in config files.​

Least-privilege tokens for git providers.​

TLS validation for CMDB and external APIs.​

Optional signed commits and respect branch protections.

## 1.13 Testing and Deployment
### Testing:

Unit tests for config, strategies, plugins.

Integration tests using fake providers and test repos.

End-to-end tests against test projects on each provider.

### Deployment:

Single Go binary packed into:

Container image.

Kubernetes CronJob.

GitHub Actions step.​

