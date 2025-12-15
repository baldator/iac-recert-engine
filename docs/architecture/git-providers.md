# Git Providers

ICE supports multiple Git hosting platforms through a unified provider abstraction layer. This document details the supported providers, their capabilities, and implementation specifics.

## Provider Architecture

### Unified Interface

All Git providers implement the same interface, ensuring consistent behavior across platforms:

```go
type GitProvider interface {
    // Repository metadata
    GetRepository(ctx context.Context, url string) (*Repository, error)

    // File history analysis
    GetLastModificationDate(ctx context.Context, filePath string) (time.Time, Commit, error)

    // Branch management
    CreateBranch(ctx context.Context, name, baseRef string) error
    BranchExists(ctx context.Context, name string) (bool, error)

    // Commit operations
    CreateCommit(ctx context.Context, branch, message string, changes []Change) (string, error)

    // Pull request operations
    CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error)
    PullRequestExists(ctx context.Context, headBranch, baseBranch string) (bool, error)
    UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error
    ClosePullRequest(ctx context.Context, id string, reason string) error

    // Assignment operations
    AssignPullRequest(ctx context.Context, id string, assignees []string) error
    RequestReviewers(ctx context.Context, id string, reviewers []string) error

    // Metadata operations
    AddLabels(ctx context.Context, id string, labels []string) error
    AddComment(ctx context.Context, id string, comment string) error
}
```

### Provider Factory

Providers are instantiated through a factory pattern:

```go
func NewProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
    switch repoCfg.Provider {
    case "github":
        return NewGitHubProvider(ctx, repoCfg, authCfg, logger)
    case "azure":
        return NewAzureDevOpsProvider(ctx, repoCfg, authCfg, logger)
    case "gitlab":
        return NewGitLabProvider(ctx, repoCfg, authCfg, logger)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", repoCfg.Provider)
    }
}
```

## GitHub Provider

### Overview
The GitHub provider integrates with GitHub's REST and GraphQL APIs for comprehensive repository management.

### Authentication
- **Token Types**: Personal Access Tokens (PAT), GitHub App tokens
- **Scopes Required**:
  - `repo`: Full repository access
  - `workflow`: GitHub Actions integration
  - `read:org`: Organization access

### Key Features

#### REST API Integration
- **Rate Limiting**: 5,000 requests/hour for PATs
- **Pagination**: Automatic handling of paginated responses
- **Retry Logic**: Exponential backoff for rate limit handling

#### GraphQL Support
- **Efficient Queries**: Reduced API calls for complex data
- **Batch Operations**: Multiple operations in single request
- **Type Safety**: Strongly typed GraphQL schema

#### Branch Protection
- **Detection**: Identifies protected branches
- **Requirements**: Ensures PR creation meets branch protection rules
- **Status Checks**: Integration with required status checks

### Implementation Details

#### Repository Operations
```go
func (p *GitHubProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
    // Parse repository URL
    owner, repo := parseGitHubURL(url)

    // API call to get repository metadata
    repoData, err := p.client.Repositories.Get(ctx, owner, repo)
    if err != nil {
        return nil, fmt.Errorf("failed to get repository: %w", err)
    }

    return &Repository{
        Name:          repoData.GetName(),
        Owner:         repoData.GetOwner().GetLogin(),
        URL:           repoData.GetHTMLURL(),
        Provider:      "github",
        DefaultBranch: repoData.GetDefaultBranch(),
    }, nil
}
```

#### Pull Request Creation
```go
func (p *GitHubProvider) CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error) {
    prReq := &github.NewPullRequest{
        Title:               &cfg.Title,
        Head:                &cfg.HeadBranch,
        Base:                &cfg.BaseBranch,
        Body:                &cfg.Description,
        Draft:               &cfg.Draft,
    }

    pr, err := p.client.PullRequests.Create(ctx, p.owner, p.repo, prReq)
    if err != nil {
        return nil, fmt.Errorf("failed to create PR: %w", err)
    }

    return &PullRequest{
        ID:    pr.GetNumber(),
        URL:   pr.GetHTMLURL(),
        Title: pr.GetTitle(),
    }, nil
}
```

### Rate Limiting Strategy

GitHub provider implements sophisticated rate limiting:

- **Primary Limit**: 5,000 requests/hour
- **Secondary Limit**: 100 requests/hour for GraphQL
- **Reset Handling**: Automatic waiting for rate limit reset
- **Backoff Strategy**: Exponential backoff with jitter
- **Request Batching**: Combine multiple operations when possible

### Error Handling

```go
func (p *GitHubProvider) handleError(err error) error {
    if ghErr, ok := err.(*github.ErrorResponse); ok {
        switch ghErr.Response.StatusCode {
        case 401:
            return fmt.Errorf("authentication failed: invalid token")
        case 403:
            return fmt.Errorf("access forbidden: insufficient permissions")
        case 404:
            return fmt.Errorf("repository not found or access denied")
        }
    }
    return err
}
```

## Azure DevOps Provider

### Overview
The Azure DevOps provider integrates with Azure DevOps Services and Server for enterprise Git operations.

### Authentication
- **Token Types**: Personal Access Tokens (PAT)
- **Scopes Required**:
  - `Code (read, write)`: Repository access
  - `Pull Request Threads (read, write)`: PR comment access
  - `Project and Team (read)`: Project metadata

### Key Features

#### Organization Structure
- **Organizations**: Top-level grouping
- **Projects**: Repository containers
- **Repositories**: Git repositories within projects
- **Teams**: User groupings for assignments

#### Work Item Integration
- **PR-Work Item Linking**: Associate PRs with work items
- **Automatic Linking**: Link based on branch naming or commit messages
- **Status Updates**: Update work item status on PR events

#### Policy Support
- **Branch Policies**: Integration with branch protection policies
- **Required Reviewers**: Support for required reviewer policies
- **Build Validation**: Integration with CI/CD pipelines

### Implementation Details

#### Repository URL Parsing
```go
func parseAzureURL(url string) (organization, project, repository string, err error) {
    // Parse URLs like:
    // https://dev.azure.com/org/project/_git/repo
    // https://org.visualstudio.com/project/_git/repo

    parts := strings.Split(strings.Trim(url, "/"), "/")
    if len(parts) < 5 {
        return "", "", "", fmt.Errorf("invalid Azure DevOps URL format")
    }

    organization = parts[3]
    project = parts[4]
    repository = strings.TrimSuffix(parts[6], "_git")

    return organization, project, repository, nil
}
```

#### Pull Request Management
```go
func (p *AzureDevOpsProvider) CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error) {
    prCreate := git.CreatePullRequestArgs{
        SourceRefName: &cfg.HeadBranch,
        TargetRefName: &cfg.BaseBranch,
        Title:         &cfg.Title,
        Description:   &cfg.Description,
    }

    pr, err := p.gitClient.CreatePullRequest(ctx, prCreate, p.project, p.repository)
    if err != nil {
        return nil, fmt.Errorf("failed to create PR: %w", err)
    }

    return &PullRequest{
        ID:    strconv.Itoa(*pr.PullRequestId),
        URL:   *pr.URL,
        Title: *pr.Title,
    }, nil
}
```

### Enterprise Features

#### Azure DevOps Server Support
- **On-Premises**: Support for Azure DevOps Server installations
- **Authentication**: Windows Authentication and PAT support
- **Custom Domains**: Support for custom domain configurations

#### Advanced Security
- **Conditional Access**: Integration with Azure AD conditional access
- **Audit Logs**: Comprehensive audit trail integration
- **Compliance**: SOC 2 and other compliance framework support

## GitLab Provider

### Overview
The GitLab provider supports both GitLab.com and self-hosted GitLab instances for comprehensive Git operations.

### Authentication
- **Token Types**: Personal Access Tokens, Project Access Tokens, Group Access Tokens
- **Scopes Required**:
  - `api`: Full API access
  - `read_repository`: Repository read access
  - `write_repository`: Repository write access

### Key Features

#### Instance Flexibility
- **GitLab.com**: Public SaaS platform
- **Self-Hosted**: Support for GitLab CE and EE
- **Version Compatibility**: Support for multiple GitLab versions

#### Group Hierarchy
- **Groups**: Hierarchical organization structure
- **Subgroups**: Nested group support
- **Permissions**: Inherited and direct permissions

#### Merge Request Features
- **Approvals**: Support for approval rules and workflows
- **Pipelines**: Integration with GitLab CI/CD
- **Environments**: Environment-specific deployments

### Implementation Details

#### Multi-Version Support
```go
func (p *GitLabProvider) detectVersion(ctx context.Context) error {
    version, _, err := p.client.Version.GetVersion()
    if err != nil {
        return fmt.Errorf("failed to get GitLab version: %w", err)
    }

    p.version = version.Version
    p.isCE = strings.Contains(version.Edition, "Community")

    // Adjust API calls based on version capabilities
    if p.compareVersion("13.0.0") >= 0 {
        p.supportsGraphQL = true
    }

    return nil
}
```

#### Merge Request Operations
```go
func (p *GitLabProvider) CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error) {
    mrOpts := &gitlab.CreateMergeRequestOptions{
        SourceBranch: &cfg.HeadBranch,
        TargetBranch: &cfg.BaseBranch,
        Title:        &cfg.Title,
        Description:  &cfg.Description,
    }

    mr, _, err := p.client.MergeRequests.CreateMergeRequest(p.projectID, mrOpts)
    if err != nil {
        return nil, fmt.Errorf("failed to create merge request: %w", err)
    }

    return &PullRequest{
        ID:    mr.IID,
        URL:   mr.WebURL,
        Title: mr.Title,
    }, nil
}
```

### Self-Hosted Considerations

#### Custom Certificates
```go
func NewGitLabProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
    client, err := gitlab.NewClient(authCfg.TokenEnv,
        gitlab.WithBaseURL(repoCfg.URL),
        gitlab.WithCustomCA("/path/to/ca.crt"),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create GitLab client: %w", err)
    }

    return &GitLabProvider{
        client: client,
        logger: logger,
    }, nil
}
```

#### Network Configuration
- **Proxy Support**: HTTP/HTTPS proxy configuration
- **Timeouts**: Configurable request timeouts
- **Retries**: Automatic retry with exponential backoff

## Provider Comparison

| Feature | GitHub | Azure DevOps | GitLab |
|---------|--------|--------------|--------|
| **Authentication** | PAT, GitHub Apps | PAT | PAT, OAuth |
| **API Rate Limits** | 5,000/hour | Configurable | Configurable |
| **GraphQL Support** | Yes | Limited | Yes |
| **Branch Protection** | Yes | Yes | Yes |
| **Work Item Integration** | Issues | Work Items | Issues |
| **CI/CD Integration** | Actions | Pipelines | CI/CD |
| **Self-Hosted Support** | GitHub Enterprise | Azure DevOps Server | GitLab CE/EE |
| **Organization Structure** | Flat | Hierarchical | Hierarchical |

## Common Provider Operations

### Branch Management

All providers support consistent branch operations:

```go
// Create feature branch
err := provider.CreateBranch(ctx, "recert/terraform-prod-2025-12-14", "main")

// Check if branch exists
exists, err := provider.BranchExists(ctx, "recert/terraform-prod-2025-12-14")
```

### Commit Creation

Standardized commit operations across providers:

```go
changes := []types.Change{
    {
        Path:    "terraform/prod/main.tf",
        Content: "# Last Recertification: 2025-12-14T10:30:00Z\n" + originalContent,
    },
}

commitSHA, err := provider.CreateCommit(ctx, "recert/terraform-prod-2025-12-14",
    "Recertify terraform/prod/main.tf", changes)
```

### Pull Request Lifecycle

Consistent PR management interface:

```go
// Create PR
pr, err := provider.CreatePullRequest(ctx, types.PRConfig{
    Title:       "Recertify Terraform Production",
    HeadBranch:  "recert/terraform-prod-2025-12-14",
    BaseBranch:  "main",
    Description: "Automated recertification of production Terraform files",
})

// Assign reviewers
err = provider.AssignPullRequest(ctx, pr.ID, []string{"infra-team", "security"})

// Add labels
err = provider.AddLabels(ctx, pr.ID, []string{"recertification", "automated"})
```

## Error Handling and Resilience

### Provider-Specific Errors

Each provider handles errors appropriately:

```go
func (p *GitHubProvider) handleAPIError(err error) error {
    if ghErr, ok := err.(*github.ErrorResponse); ok {
        switch ghErr.Response.StatusCode {
        case 422:
            return fmt.Errorf("validation failed: %s", ghErr.Message)
        case 404:
            return fmt.Errorf("resource not found")
        }
    }
    return fmt.Errorf("GitHub API error: %w", err)
}
```

### Retry and Backoff

Automatic retry logic for transient failures:

```go
func (p *GitHubProvider) withRetry(ctx context.Context, operation func() error) error {
    backoff := time.Second

    for attempt := 1; attempt <= 5; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }

        p.logger.Warn("operation failed, retrying",
            zap.Int("attempt", attempt),
            zap.Duration("backoff", backoff),
            zap.Error(err))

        select {
        case <-time.After(backoff):
        case <-ctx.Done():
            return ctx.Err()
        }

        backoff *= 2 // Exponential backoff
    }

    return fmt.Errorf("operation failed after 5 attempts")
}
```

### Circuit Breaker Pattern

Prevent cascading failures:

```go
type CircuitBreaker struct {
    failures int
    lastFail time.Time
    state    string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Call(operation func() error) error {
    if cb.state == "open" {
        if time.Since(cb.lastFail) > cb.timeout {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := operation()
    if err != nil {
        cb.failures++
        cb.lastFail = time.Now()
        if cb.failures >= cb.threshold {
            cb.state = "open"
        }
        return err
    }

    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = "closed"
    return nil
}
```

## Testing and Validation

### Provider Testing

Unit tests for provider implementations:

```go
func TestGitHubProvider_CreatePullRequest(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock GitHub API responses
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(201)
        json.NewEncoder(w).Encode(github.PullRequest{
            Number: github.Int(123),
            Title:  github.String("Test PR"),
        })
    }))
    defer mockServer.Close()

    provider := &GitHubProvider{
        client: github.NewClient(nil, github.WithBaseURL(mockServer.URL)),
    }

    pr, err := provider.CreatePullRequest(context.Background(), PRConfig{
        Title: "Test PR",
        HeadBranch: "feature/test",
        BaseBranch: "main",
    })

    assert.NoError(t, err)
    assert.Equal(t, "123", pr.ID)
    assert.Contains(t, pr.Title, "Test PR")
}
```

### Integration Testing

End-to-end testing with real providers:

```go
func TestProviderIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    provider := NewGitHubProvider(context.Background(), config, auth, logger)

    // Test repository access
    repo, err := provider.GetRepository(context.Background(), "https://github.com/octocat/Hello-World")
    assert.NoError(t, err)
    assert.Equal(t, "Hello-World", repo.Name)
}
```

## Performance Optimization

### Connection Pooling

Reuse HTTP connections for better performance:

```go
func NewGitHubProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
    client := &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
        Timeout: 30 * time.Second,
    }

    ghClient := github.NewClient(nil).WithHTTPClient(client)
    // ... rest of initialization
}
```

### Request Batching

Combine multiple operations when possible:

```go
func (p *GitHubProvider) batchLabelUpdates(ctx context.Context, prIDs []string, labels []string) error {
    // Use GraphQL mutation to update multiple PRs
    mutation := `
    mutation($input: UpdatePullRequestInput!) {
        updatePullRequest(input: $input) {
            pullRequest {
                id
            }
        }
    }`

    // Batch updates for better performance
    // ...
}
```

### Caching Strategy

Cache frequently accessed data:

```go
type ProviderCache struct {
    repositories sync.Map // map[string]*Repository
    branches     sync.Map // map[string]bool
    commits      sync.Map // map[string]*Commit
}

func (c *ProviderCache) GetRepository(key string) (*Repository, bool) {
    if val, ok := c.repositories.Load(key); ok {
        return val.(*Repository), true
    }
    return nil, false
}

func (c *ProviderCache) SetRepository(key string, repo *Repository) {
    c.repositories.Store(key, repo)
}
```

## Monitoring and Observability

### Metrics Collection

Track provider performance and reliability:

```go
type ProviderMetrics struct {
    requestsTotal    prometheus.Counter
    requestDuration  prometheus.Histogram
    errorsTotal      prometheus.Counter
    rateLimitHits    prometheus.Counter
}

func (p *GitHubProvider) observeRequest(method, endpoint string, start time.Time, err error) {
    duration := time.Since(start).Seconds()

    p.metrics.requestDuration.WithLabelValues(method, endpoint).Observe(duration)

    if err != nil {
        p.metrics.errorsTotal.WithLabelValues(method, endpoint).Inc()
    }

    // Check for rate limiting
    if isRateLimitError(err) {
        p.metrics.rateLimitHits.WithLabelValues(method, endpoint).Inc()
    }
}
```

### Structured Logging

Comprehensive logging for debugging:

```go
func (p *GitHubProvider) CreatePullRequest(ctx context.Context, cfg PRConfig) (*PullRequest, error) {
    start := time.Now()
    p.logger.Info("creating pull request",
        zap.String("repository", fmt.Sprintf("%s/%s", p.owner, p.repo)),
        zap.String("head", cfg.HeadBranch),
        zap.String("base", cfg.BaseBranch),
        zap.String("title", cfg.Title))

    pr, err := p.client.PullRequests.Create(ctx, p.owner, p.repo, prReq)

    p.logger.Info("pull request created",
        zap.Duration("duration", time.Since(start)),
        zap.String("url", pr.GetHTMLURL()),
        zap.Error(err))

    return pr, err
}
```

## Next Steps

- [Core Components](components.md) - Component architecture details
- [Plugin System](plugins.md) - Plugin architecture
- [Security](security.md) - Security considerations
