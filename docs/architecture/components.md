# Core Components

This document provides detailed information about ICE's core components, their responsibilities, interactions, and implementation details.

## Component Overview

ICE is organized into several layered components that work together to provide comprehensive Infrastructure as Code recertification capabilities:

```
┌─────────────────┐
│   CLI Layer     │  Command-line interface and argument parsing
├─────────────────┤
│ Configuration   │  Configuration loading, validation, and management
├─────────────────┤
│    Engine       │  Main orchestration and workflow management
├─────────────────┤
│   Scanning      │  Repository scanning and file analysis
├─────────────────┤
│   Strategy      │  File grouping and PR organization
├─────────────────┤
│  Assignment     │  Reviewer and assignee determination
├─────────────────┤
│   Provider      │  Git provider abstraction and API calls
├─────────────────┤
│    Plugin       │  Extensibility and custom integrations
├─────────────────┤
│    Audit        │  Logging, monitoring, and compliance tracking
└─────────────────┘
```

## CLI Layer

### Purpose
Provides the command-line interface for user interaction, argument parsing, and execution initiation.

### Components

#### `cmd/ice/main.go`
- Application entry point
- Basic setup and initialization

#### `cmd/ice/root.go`
- Root command definition
- Global flags and configuration
- Help system integration

#### `cmd/ice/run.go`
- Run command implementation
- Configuration loading and validation
- Engine initialization and execution

### Key Features
- **Argument Parsing**: Uses Cobra for robust CLI handling
- **Configuration Resolution**: Automatic config file discovery
- **Environment Integration**: Environment variable support
- **Help System**: Comprehensive built-in help and documentation

### Usage Examples
```bash
# Basic execution
ice run --config config.yaml

# Dry run with verbose output
ice run --dry-run --verbose --config config.yaml

# Override repository
ice run --repo-url https://github.com/org/repo --config config.yaml
```

## Configuration Layer

### Purpose
Manages all configuration loading, validation, and access throughout the application.

### Components

#### `internal/config/types.go`
Defines all configuration structures:

```go
type Config struct {
    Version    string           `yaml:"version"`
    Repository RepositoryConfig `yaml:"repository"`
    Auth       AuthConfig       `yaml:"auth"`
    Global     GlobalConfig     `yaml:"global"`
    Patterns   []Pattern        `yaml:"patterns"`
    PRStrategy PRStrategyConfig `yaml:"pr_strategy"`
    Assignment AssignmentConfig `yaml:"assignment"`
    Plugins    PluginConfigs    `yaml:"plugins"`
    Audit      AuditConfig      `yaml:"audit"`
}
```

#### Configuration Validation
- **Schema Validation**: Ensures required fields are present
- **Type Validation**: Validates data types and constraints
- **Cross-Reference Validation**: Checks relationships between settings
- **Custom Validators**: Business logic validation rules

### Key Features
- **YAML Parsing**: Full YAML 1.2 support
- **Environment Variables**: `${VAR_NAME}` substitution
- **Validation**: Comprehensive error reporting
- **Defaults**: Sensible default values for optional settings

### Configuration Resolution Order
1. Explicit `--config` flag
2. `ICE_CONFIG_PATH` environment variable
3. `./ice.yaml`, `./ice.yml`
4. `~/.ice.yaml`, `~/.ice.yml`
5. `~/.config/ice/config.yaml`

## Engine Layer

### Purpose
Orchestrates the entire recertification workflow from start to finish.

### Components

#### `internal/engine/engine.go`
Main engine implementation with the following phases:

1. **Initialization**: Load configuration and setup components
2. **Scanning**: Discover files requiring recertification
3. **Analysis**: Analyze git history and determine requirements
4. **Grouping**: Organize files into logical PR groups
5. **Assignment**: Determine reviewers and assignees
6. **Execution**: Create branches, commits, and PRs
7. **Cleanup**: Resource cleanup and final reporting

#### Workflow Management
```go
func (e *Engine) Run(ctx context.Context) error {
    // Phase 1: Initialize components
    if err := e.initialize(ctx); err != nil {
        return fmt.Errorf("initialization failed: %w", err)
    }

    // Phase 2: Scan repository
    results, err := e.scan(ctx)
    if err != nil {
        return fmt.Errorf("scanning failed: %w", err)
    }

    // Phase 3: Group files
    groups, err := e.group(ctx, results)
    if err != nil {
        return fmt.Errorf("grouping failed: %w", err)
    }

    // Phase 4: Create PRs
    if err := e.execute(ctx, groups); err != nil {
        return fmt.Errorf("execution failed: %w", err)
    }

    return nil
}
```

### Key Features
- **Phase-Based Execution**: Clear separation of concerns
- **Error Recovery**: Graceful handling of partial failures
- **Concurrency Control**: Configurable parallel execution
- **Observability**: Comprehensive logging and metrics
- **Dry Run Support**: Safe testing without side effects

## Scanning Layer

### Purpose
Discovers and analyzes Infrastructure as Code files in Git repositories.

### Components

#### `internal/scan/file_scanner.go`
- **Pattern Matching**: Uses glob patterns to find files
- **Directory Traversal**: Efficient repository scanning
- **File Filtering**: Include/exclude pattern support
- **Metadata Collection**: Basic file information gathering

#### `internal/scan/git_history.go`
- **Git Integration**: Uses git commands for history analysis
- **Commit Analysis**: Extracts last modification dates
- **Author Tracking**: Identifies last committers
- **Performance Optimization**: Caches git operations

#### `internal/scan/recert_checker.go`
- **Date Calculations**: Determines recertification requirements
- **Priority Assessment**: Calculates urgency levels
- **Decorator Detection**: Identifies existing recertification markers
- **Business Logic**: Implements recertification rules

### Scanning Process

```go
func (s *Scanner) Scan(ctx context.Context) ([]RecertCheckResult, error) {
    // 1. Find matching files
    files, err := s.findFiles(ctx)
    if err != nil {
        return nil, err
    }

    // 2. Enrich with git history
    enriched, err := s.enrichWithHistory(ctx, files)
    if err != nil {
        return nil, err
    }

    // 3. Check recertification requirements
    results, err := s.checkRecertification(ctx, enriched)
    if err != nil {
        return nil, err
    }

    return results, nil
}
```

### Key Features
- **Efficient Pattern Matching**: Optimized glob operations
- **Git History Caching**: Reduces redundant git queries
- **Parallel Processing**: Concurrent file analysis
- **Memory Efficient**: Streaming processing for large repositories

## Strategy Layer

### Purpose
Groups files requiring recertification into logical pull request units.

### Components

#### `internal/strategy/strategy.go`
Strategy interface and factory:

```go
type Strategy interface {
    Group(results []RecertCheckResult) ([]FileGroup, error)
}

func NewStrategy(cfg PRStrategyConfig, logger *zap.Logger) (Strategy, error) {
    switch cfg.Type {
    case "per_file":
        return &PerFileStrategy{logger: logger}, nil
    case "per_pattern":
        return &PerPatternStrategy{logger: logger}, nil
    case "per_committer":
        return &PerCommitterStrategy{logger: logger}, nil
    case "single_pr":
        return &SinglePRStrategy{logger: logger}, nil
    case "plugin":
        return nil, fmt.Errorf("plugin strategy not implemented")
    }
}
```

#### Strategy Implementations

##### Per File Strategy
- Creates one PR per file
- Simplest grouping approach
- Maximum isolation between changes

##### Per Pattern Strategy
- Groups files by pattern configuration
- Logical organization by technology/domain
- Balances isolation with efficiency

##### Per Committer Strategy
- Groups files by last committer
- Individual accountability model
- Natural ownership boundaries

##### Single PR Strategy
- Combines all files into one PR
- Simplest management overhead
- Best for small repositories

### Key Features
- **Configurable Limits**: Maximum files per PR
- **Automatic Splitting**: Handles large groups gracefully
- **Strategy Selection**: Runtime strategy switching
- **Extensible Design**: Easy addition of new strategies

## Assignment Layer

### Purpose
Determines reviewers and assignees for pull requests.

### Components

#### `internal/assign/resolver.go`
Main assignment resolution logic:

```go
func (r *Resolver) Resolve(ctx context.Context, group FileGroup) (AssignmentResult, error) {
    // Determine strategy for this group
    strategy := r.determineStrategy(group)

    // Apply assignment logic
    switch strategy {
    case "static":
        return r.assignStatic(group)
    case "last_committer":
        return r.assignLastCommitter(group)
    case "plugin":
        return r.assignPlugin(group)
    case "composite":
        return r.assignComposite(group)
    }

    // Fallback to default assignees
    return AssignmentResult{Assignees: r.cfg.FallbackAssignees}, nil
}
```

#### Assignment Strategies

##### Static Assignment
- Assigns same reviewers to all PRs
- Simple and predictable
- Good for centralized teams

##### Last Committer Assignment
- Assigns based on git history
- Subject matter expertise
- Automatic ownership tracking

##### Plugin Assignment
- Custom logic via plugins
- Integration with external systems
- Maximum flexibility

##### Composite Assignment
- Pattern-based strategy selection
- Complex organizational rules
- Fine-grained control

### Key Features
- **Fallback Handling**: Ensures assignments always work
- **Plugin Integration**: Extensible assignment logic
- **Performance Optimized**: Efficient git history queries
- **Error Resilient**: Graceful degradation on failures

## Provider Layer

### Purpose
Abstracts Git provider APIs for repository operations.

### Components

#### `internal/provider/provider.go`
Common interface definition:

```go
type GitProvider interface {
    // Repository operations
    GetRepository(ctx context.Context, url string) (*Repository, error)

    // File operations
    GetLastModificationDate(ctx context.Context, filePath string) (time.Time, Commit, error)

    // Branch operations
    CreateBranch(ctx context.Context, name, baseRef string) error
    BranchExists(ctx context.Context, name string) (bool, error)

    // Commit operations
    CreateCommit(ctx context.Context, branch, message string, changes []Change) (string, error)

    // PR operations
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

#### Provider Implementations

##### GitHub Provider (`internal/provider/github.go`)
- REST API integration
- GraphQL support for complex queries
- Rate limiting and retry logic
- Branch protection handling

##### Azure DevOps Provider (`internal/provider/azure.go`)
- Azure DevOps REST APIs
- Work item integration
- Policy and approval handling
- Organization/project scoping

##### GitLab Provider (`internal/provider/gitlab.go`)
- GitLab REST and GraphQL APIs
- Group and project permissions
- Merge request approvals
- CI/CD pipeline integration

### Key Features
- **Unified Interface**: Consistent API across providers
- **Error Handling**: Provider-specific error translation
- **Rate Limiting**: Automatic retry with backoff
- **Caching**: Response caching for improved performance
- **Observability**: Detailed logging of API interactions

## Plugin Layer

### Purpose
Provides extensibility through custom plugins and integrations.

### Components

#### `internal/plugin/manager.go`
Plugin lifecycle management:

```go
type Manager struct {
    plugins map[string]Plugin
    logger  *zap.Logger
}

func (m *Manager) LoadPlugins(config PluginConfigs) error {
    for name, cfg := range config {
        if !cfg.Enabled {
            continue
        }

        plugin, err := m.loadPlugin(name, cfg)
        if err != nil {
            return fmt.Errorf("failed to load plugin %s: %w", name, err)
        }

        if err := plugin.Init(cfg.Config); err != nil {
            return fmt.Errorf("failed to init plugin %s: %w", name, err)
        }

        m.plugins[name] = plugin
    }

    return nil
}
```

#### Plugin Types

##### Assignment Plugins
- Custom reviewer assignment logic
- Integration with CMDB, LDAP, etc.
- Business rule implementation

##### Filter Plugins
- Additional file filtering beyond patterns
- Custom inclusion/exclusion logic
- Security and compliance filtering

##### Strategy Plugins
- Custom PR grouping algorithms
- Organization-specific logic
- Advanced grouping rules

### Key Features
- **Runtime Loading**: Dynamic plugin discovery and loading
- **Isolation**: Plugin failures don't affect core functionality
- **Configuration**: Plugin-specific configuration support
- **Lifecycle Management**: Proper initialization and cleanup

## Audit Layer

### Purpose
Provides comprehensive logging, monitoring, and compliance tracking.

### Components

#### `internal/audit/audit.go`
Audit event management:

```go
type Auditor struct {
    storage AuditStorage
    logger  *zap.Logger
}

func (a *Auditor) LogEvent(event AuditEvent) error {
    // Structure audit event
    auditEvent := AuditEvent{
        Timestamp:   time.Now(),
        RunID:       event.RunID,
        EventType:   event.EventType,
        Message:     event.Message,
        Details:     event.Details,
        Error:       event.Error,
        Repository:  event.Repository,
        User:        event.User,
    }

    // Store audit event
    return a.storage.Store(auditEvent)
}
```

#### Audit Storage Options

##### File Storage
- JSON lines format
- Daily rotation
- Local filesystem storage

##### S3 Storage
- Cloud storage integration
- Long-term retention
- Cross-region replication

### Key Features
- **Structured Events**: Consistent audit event format
- **Multiple Storage**: File and cloud storage options
- **Performance**: Minimal impact on execution
- **Compliance**: Detailed activity tracking
- **Searchable**: Query and filter audit events

## Component Interactions

### Data Flow Example

```
CLI Layer → Configuration Layer
    ↓
Configuration Layer → Engine Layer
    ↓
Engine Layer → Scanning Layer → File Discovery
    ↓
Engine Layer → Strategy Layer → File Grouping
    ↓
Engine Layer → Assignment Layer → Reviewer Assignment
    ↓
Engine Layer → Provider Layer → Git Operations
    ↓
Engine Layer → Audit Layer → Event Logging
```

### Error Propagation

```
Component Failure → Error Logging → Audit Recording → Graceful Degradation → User Notification
```

### Configuration Sharing

```
Configuration Layer → All Components (Read-Only Access)
```

## Performance Considerations

### Concurrency Control
- **Engine Level**: Configurable max concurrent PRs
- **Scanning Level**: Parallel file analysis
- **Provider Level**: Rate limiting and batching

### Resource Management
- **Memory**: Streaming processing for large datasets
- **Network**: Connection pooling and reuse
- **Storage**: Efficient audit log rotation

### Monitoring Points
- **Execution Time**: Per-phase timing metrics
- **Resource Usage**: Memory and CPU monitoring
- **Error Rates**: Component failure tracking
- **Throughput**: Files processed per minute

## Testing and Validation

### Unit Testing
- **Component Isolation**: Mock dependencies for testing
- **Interface Compliance**: Verify interface implementations
- **Error Scenarios**: Test failure modes and recovery

### Integration Testing
- **End-to-End Workflows**: Complete execution path testing
- **Provider Compatibility**: Test with real Git providers
- **Plugin Integration**: Validate plugin loading and execution

### Performance Testing
- **Load Testing**: High-volume repository processing
- **Concurrency Testing**: Multi-threaded execution validation
- **Resource Testing**: Memory and CPU usage analysis

## Next Steps

- [System Overview](overview.md) - High-level architecture
- [Git Providers](git-providers.md) - Provider-specific details
- [Plugin System](plugins.md) - Plugin architecture
- [Security](security.md) - Security considerations
