# System Architecture

This document describes the high-level architecture of IaC Recertification Engine (ICE), including core components, data flow, and design principles.

## Architectural Principles

ICE follows several key architectural principles:

- **Stateless Design**: Each execution is independent and repeatable
- **Configuration-Driven**: All behavior controlled through YAML configuration
- **Plugin Architecture**: Extensible through custom plugins
- **Provider Abstraction**: Git provider agnostic through common interfaces
- **Observable**: Comprehensive logging, metrics, and audit trails

## Core Components

### 1. Configuration Layer

**Purpose**: Load, validate, and provide access to configuration settings.

**Components**:
- `config/types.go`: Configuration structs and validation
- `config/loader.go`: Configuration file parsing and environment variable substitution
- `config/validator.go`: Configuration validation logic

**Responsibilities**:
- Parse YAML configuration files
- Validate configuration schema
- Substitute environment variables
- Provide typed configuration access

### 2. Engine Layer

**Purpose**: Orchestrate the recertification workflow.

**Components**:
- `engine/engine.go`: Main orchestration logic
- `engine/workflow.go`: Step-by-step execution flow
- `engine/executor.go`: Concurrent execution management

**Responsibilities**:
- Coordinate scanning, analysis, and PR creation
- Manage concurrency and resource limits
- Handle errors and recovery
- Provide execution observability

### 3. Scanning Layer

**Purpose**: Discover and analyze IaC files in repositories.

**Components**:
- `scan/file_scanner.go`: File system scanning and pattern matching
- `scan/git_history.go`: Git history analysis and metadata extraction
- `scan/recert_checker.go`: Recertification logic and priority calculation

**Responsibilities**:
- Scan repositories for matching files
- Query git history for modification dates
- Calculate recertification requirements
- Enrich file metadata

### 4. Strategy Layer

**Purpose**: Group files into logical PR units.

**Components**:
- `strategy/strategy.go`: Strategy interface and implementations
- `strategy/per_file.go`: One PR per file strategy
- `strategy/per_pattern.go`: One PR per pattern strategy
- `strategy/per_committer.go`: Group by last committer strategy

**Responsibilities**:
- Implement different grouping algorithms
- Validate grouping constraints
- Optimize PR sizes
- Handle edge cases

### 5. Assignment Layer

**Purpose**: Determine PR assignees and reviewers.

**Components**:
- `assign/resolver.go`: Assignment resolution logic
- `assign/static.go`: Static assignment strategy
- `assign/last_committer.go`: Last committer assignment
- `assign/composite.go`: Composite assignment rules

**Responsibilities**:
- Resolve assignees based on configuration
- Handle fallback scenarios
- Integrate with external systems via plugins
- Validate assignment results

### 6. Provider Layer

**Purpose**: Abstract Git provider APIs for repository operations.

**Components**:
- `provider/provider.go`: GitProvider interface definition
- `provider/github.go`: GitHub API implementation
- `provider/azure.go`: Azure DevOps API implementation
- `provider/gitlab.go`: GitLab API implementation

**Responsibilities**:
- Create and manage branches
- Create commits and pull requests
- Assign reviewers and labels
- Handle provider-specific features

### 7. Plugin Layer

**Purpose**: Extend functionality through custom plugins.

**Components**:
- `plugin/manager.go`: Plugin loading and management
- `plugin/interface.go`: Plugin interfaces
- `plugin/types/`: Plugin type definitions

**Responsibilities**:
- Load plugins from configuration
- Provide plugin execution context
- Handle plugin errors and timeouts
- Support different plugin types

## Data Flow

### Execution Flow

1. **Configuration Loading**
   - Parse YAML configuration
   - Validate schema and values
   - Substitute environment variables

2. **Repository Discovery**
   - Connect to Git provider
   - Scan repository for matching files
   - Retrieve git history metadata

3. **Recertification Analysis**
   - Calculate days since last modification
   - Determine recertification priority
   - Filter files requiring action

4. **Grouping and Assignment**
   - Apply configured grouping strategy
   - Resolve assignees and reviewers
   - Prepare PR content and metadata

5. **PR Creation**
   - Create feature branches
   - Commit recertification markers
   - Create pull requests with proper metadata

6. **Audit and Cleanup**
   - Log all actions and results
   - Clean up temporary resources
   - Report execution summary

### Data Structures

#### FileInfo
```go
type FileInfo struct {
    Path           string
    Pattern        string
    LastCommit     Commit
    DaysSinceMod   int
    NeedsRecert    bool
    Priority       Priority
    NextDueDate    time.Time
}
```

#### RecertCheckResult
```go
type RecertCheckResult struct {
    File           FileInfo
    Threshold      int
    DaysOverdue    int
    Priority       Priority
    Recommendation string
}
```

#### PRGroup
```go
type PRGroup struct {
    ID          string
    Pattern     string
    Files       []FileInfo
    Assignees   []string
    Reviewers   []string
    BranchName  string
    Title       string
    Description string
}
```

## Design Patterns

### Strategy Pattern
Used for PR grouping and assignment strategies:
- `Strategy` interface defines common contract
- Concrete implementations for each strategy type
- Runtime selection based on configuration

### Provider Pattern
Abstracts Git provider differences:
- `GitProvider` interface for common operations
- Provider-specific implementations
- Factory pattern for provider instantiation

### Plugin Pattern
Enables extensibility:
- Plugin interfaces for different extension points
- Runtime loading and execution
- Error isolation and fallback handling

### Observer Pattern
Provides observability:
- Event publishing for key operations
- Multiple subscribers (logging, metrics, audit)
- Decoupled monitoring and alerting

## Security Considerations

### Authentication
- Token-based authentication for all providers
- Environment variable storage for secrets
- Minimal required permissions principle

### Data Protection
- No sensitive data storage in configuration
- Secure token handling and cleanup
- Audit logging without exposing secrets

### Network Security
- HTTPS-only communication
- Certificate validation
- Timeout and retry limits

## Performance Characteristics

### Scalability
- Concurrent PR creation with configurable limits
- Efficient git history queries
- Batch API operations where supported

### Resource Usage
- Minimal memory footprint
- Configurable concurrency limits
- Automatic cleanup of temporary resources

### Monitoring
- Structured logging with configurable levels
- Execution metrics and timing
- Error tracking and alerting

## Deployment Patterns

### CLI Tool
- Direct command-line execution
- Suitable for manual runs and scripts
- Easy integration with existing tooling

### Docker Container
- Containerized execution
- Consistent runtime environment
- Kubernetes and orchestration friendly

### Scheduled Service
- Cron-based execution
- Kubernetes CronJobs
- CI/CD pipeline integration

## Extensibility Points

### Plugin Types
- **Assignment Plugins**: Custom assignee resolution
- **Filter Plugins**: Custom file filtering logic
- **Transform Plugins**: File content transformation
- **Hook Plugins**: Execution lifecycle hooks
- **Validator Plugins**: Custom validation rules

### Provider Extensions
- Support for additional Git providers
- Custom provider implementations
- Provider-specific feature extensions

### Strategy Extensions
- Custom grouping algorithms
- Advanced assignment rules
- Organization-specific logic
