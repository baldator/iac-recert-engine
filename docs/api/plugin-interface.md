# Plugin Interface

ICE provides a plugin architecture that allows extending functionality through external modules. This document details the plugin interfaces, data structures, and development guidelines.

## Overview

Plugins enable organizations to integrate ICE with external systems and implement custom logic for:

- **Assignment Plugins**: Custom reviewer assignment logic
- **Filter Plugins**: Additional file filtering beyond patterns
- **Strategy Plugins**: Custom PR grouping strategies

## Core Interfaces

### Plugin Base Interface

All plugins must implement the base `Plugin` interface:

```go
type Plugin interface {
    Init(config map[string]string) error
}
```

**Methods:**
- `Init(config map[string]string) error`: Initialize the plugin with configuration

### Assignment Plugin Interface

Assignment plugins determine how reviewers are assigned to pull requests:

```go
type AssignmentPlugin interface {
    Plugin
    Resolve(files []FileInfo) (AssignmentResult, error)
}
```

**Methods:**
- `Resolve(files []FileInfo) (AssignmentResult, error)`: Return assignment for given files

### Filter Plugin Interface

Filter plugins provide additional file filtering logic:

```go
type FilterPlugin interface {
    Plugin
    Filter(files []FileInfo) ([]FileInfo, error)
}
```

**Methods:**
- `Filter(files []FileInfo) ([]FileInfo, error)`: Filter and return subset of files

### Strategy Plugin Interface

Strategy plugins implement custom PR grouping logic:

```go
type StrategyPlugin interface {
    Plugin
    Group(results []RecertCheckResult) ([]FileGroup, error)
}
```

**Methods:**
- `Group(results []RecertCheckResult) ([]FileGroup, error)`: Group files into PRs

## Data Structures

### FileInfo

Contains information about a file for plugin processing:

```go
type FileInfo struct {
    Path         string    // File path relative to repository root
    Size         int64     // File size in bytes
    LastModified time.Time // Last modification timestamp (ISO 8601 string in JSON)
    CommitHash   string    // Latest commit hash
    CommitAuthor string    // Last committer name
    CommitEmail  string    // Last committer email
    CommitMsg    string    // Latest commit message
}
```

**JSON Representation:**
```json
{
  "Path": "terraform/prod/main.tf",
  "Size": 1024,
  "LastModified": "2025-12-14T10:30:00Z",
  "CommitHash": "a1b2c3d4e5f6",
  "CommitAuthor": "John Doe",
  "CommitEmail": "john.doe@example.com",
  "CommitMsg": "Update infrastructure configuration"
}
```

### AssignmentResult

Contains assignment information returned by assignment plugins:

```go
type AssignmentResult struct {
    Assignees []string // Primary assignees (required reviewers)
    Reviewers []string // Additional reviewers (optional)
    Team      string   // Team assignment
    Priority  string   // Priority level
}
```

**Fields:**
- `Assignees`: Users who must review the PR
- `Reviewers`: Additional reviewers (may be optional)
- `Team`: Team assignment (provider-specific)
- `Priority`: Priority level (provider-specific)

**Example:**
```go
AssignmentResult{
    Assignees: []string{"security-team", "infra-lead"},
    Reviewers: []string{"devops-team"},
    Team: "platform-team",
    Priority: "high",
}
```

### RecertCheckResult

Contains recertification check results for strategy plugins:

```go
type RecertCheckResult struct {
    File       FileInfo // File information
    PatternName string  // Pattern that matched this file
    NeedsRecert bool    // Whether file needs recertification
}
```

### FileGroup

Represents a group of files for a single PR:

```go
type FileGroup struct {
    ID       string                // Unique group identifier
    Strategy string                // Grouping strategy used
    Files    []RecertCheckResult   // Files in this group
}
```

## Plugin Development

### Project Structure

Create plugins in the `plugins/` directory:

```
plugins/
└── myplugin/
    ├── go.mod
    ├── myplugin/
    │   └── plugin.go
    └── README.md
```

### Assignment Plugin Example

```go
package myplugin

import (
    "context"
    "fmt"
    "strings"

    "github.com/baldator/iac-recert-engine/pkg/api"
)

type MyAssignmentPlugin struct {
    apiURL    string
    apiKey    string
    teamMap   map[string]string // pattern -> team mapping
}

func (p *MyAssignmentPlugin) Init(config map[string]string) error {
    p.apiURL = config["api_url"]
    if p.apiURL == "" {
        return fmt.Errorf("api_url is required")
    }

    p.apiKey = config["api_key"]
    if p.apiKey == "" {
        return fmt.Errorf("api_key is required")
    }

    // Initialize team mapping
    p.teamMap = map[string]string{
        "terraform": "infra-team",
        "kubernetes": "k8s-team",
        "policies": "security-team",
    }

    return nil
}

func (p *MyAssignmentPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    if len(files) == 0 {
        return api.AssignmentResult{}, fmt.Errorf("no files provided")
    }

    // Determine primary pattern from first file
    primaryFile := files[0].Path
    var primaryTeam string

    for pattern, team := range p.teamMap {
        if strings.Contains(primaryFile, pattern) {
            primaryTeam = team
            break
        }
    }

    // Default fallback
    if primaryTeam == "" {
        primaryTeam = "devops-team"
    }

    // Check for production files
    hasProd := false
    for _, file := range files {
        if strings.Contains(file.Path, "/prod/") || strings.Contains(file.Path, "/production/") {
            hasProd = true
            break
        }
    }

    assignees := []string{primaryTeam}
    if hasProd {
        assignees = append(assignees, "security-team")
    }

    return api.AssignmentResult{
        Assignees: assignees,
        Reviewers: []string{"optional-reviewer"},
        Team: primaryTeam,
        Priority: "medium",
    }, nil
}

// Export for plugin loading
var Plugin = &MyAssignmentPlugin{}
```

### Filter Plugin Example

```go
package myplugin

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/baldator/iac-recert-engine/pkg/api"
)

type MyFilterPlugin struct {
    excludePatterns []string
    maxFileSize     int64
}

func (p *MyFilterPlugin) Init(config map[string]string) error {
    // Parse exclude patterns
    if patterns := config["exclude_patterns"]; patterns != "" {
        p.excludePatterns = strings.Split(patterns, ",")
        for i, pattern := range p.excludePatterns {
            p.excludePatterns[i] = strings.TrimSpace(pattern)
        }
    }

    // Parse max file size
    if size := config["max_file_size"]; size != "" {
        if _, err := fmt.Sscanf(size, "%d", &p.maxFileSize); err != nil {
            return fmt.Errorf("invalid max_file_size: %v", err)
        }
    }

    return nil
}

func (p *MyFilterPlugin) Filter(files []api.FileInfo) ([]api.FileInfo, error) {
    var filtered []api.FileInfo

    for _, file := range files {
        // Check file size
        if p.maxFileSize > 0 && file.Size > p.maxFileSize {
            continue // Skip large files
        }

        // Check exclude patterns
        excluded := false
        for _, pattern := range p.excludePatterns {
            if matched, err := filepath.Match(pattern, file.Path); err == nil && matched {
                excluded = true
                break
            }
        }

        if !excluded {
            filtered = append(filtered, file)
        }
    }

    return filtered, nil
}

var Plugin = &MyFilterPlugin{}
```

### Strategy Plugin Example

```go
package myplugin

import (
    "fmt"
    "sort"
    "strings"

    "github.com/baldator/iac-recert-engine/pkg/api"
)

type MyStrategyPlugin struct {
    maxFilesPerPR int
    groupBy       string // "directory", "owner", "size"
}

func (p *MyStrategyPlugin) Init(config map[string]string) error {
    p.groupBy = config["group_by"]
    if p.groupBy == "" {
        p.groupBy = "directory"
    }

    if maxFiles := config["max_files_per_pr"]; maxFiles != "" {
        if _, err := fmt.Sscanf(maxFiles, "%d", &p.maxFilesPerPR); err != nil {
            return fmt.Errorf("invalid max_files_per_pr: %v", err)
        }
    }

    if p.maxFilesPerPR <= 0 {
        p.maxFilesPerPR = 25 // default
    }

    return nil
}

func (p *MyStrategyPlugin) Group(results []api.RecertCheckResult) ([]api.FileGroup, error) {
    if len(results) == 0 {
        return nil, nil
    }

    var groups []api.FileGroup

    switch p.groupBy {
    case "directory":
        groups = p.groupByDirectory(results)
    case "owner":
        groups = p.groupByOwner(results)
    case "size":
        groups = p.groupBySize(results)
    default:
        return nil, fmt.Errorf("unsupported group_by: %s", p.groupBy)
    }

    // Apply max files per PR limit
    if p.maxFilesPerPR > 0 {
        groups = p.splitLargeGroups(groups)
    }

    return groups, nil
}

func (p *MyStrategyPlugin) groupByDirectory(results []api.RecertCheckResult) []api.FileGroup {
    groups := make(map[string][]api.RecertCheckResult)

    for _, result := range results {
        if !result.NeedsRecert {
            continue
        }

        // Extract directory from path
        dir := filepath.Dir(result.File.Path)
        groups[dir] = append(groups[dir], result)
    }

    var fileGroups []api.FileGroup
    for dir, files := range groups {
        fileGroups = append(fileGroups, api.FileGroup{
            ID:       fmt.Sprintf("dir-%s", dir),
            Strategy: "directory",
            Files:    files,
        })
    }

    return fileGroups
}

func (p *MyStrategyPlugin) splitLargeGroups(groups []api.FileGroup) []api.FileGroup {
    var result []api.FileGroup

    for _, group := range groups {
        if len(group.Files) <= p.maxFilesPerPR {
            result = append(result, group)
            continue
        }

        // Split into multiple groups
        for i := 0; i < len(group.Files); i += p.maxFilesPerPR {
            end := i + p.maxFilesPerPR
            if end > len(group.Files) {
                end = len(group.Files)
            }

            splitGroup := api.FileGroup{
                ID:       fmt.Sprintf("%s-%d", group.ID, i/p.maxFilesPerPR+1),
                Strategy: group.Strategy,
                Files:    group.Files[i:end],
            }
            result = append(result, splitGroup)
        }
    }

    return result
}

var Plugin = &MyStrategyPlugin{}
```

## Plugin Registration

Plugins are registered in `internal/plugin/plugin.go`:

```go
case "myplugin":
    if cfg.Type != "assignment" {
        return nil, fmt.Errorf("myplugin plugin must be of type assignment")
    }
    apiPlugin := myplugin.Plugin
    plugin = &assignmentPluginWrapper{apiPlugin: apiPlugin}
```

## Configuration

Plugins receive configuration through the `Init` method:

```yaml
plugins:
  myassignment:
    enabled: true
    type: "assignment"
    module: "myplugin"
    config:
      api_url: "https://api.example.com"
      api_key: "${API_KEY}"
      team_mapping: "terraform=infra-team,kubernetes=k8s-team"
```

## Error Handling

Plugins should handle errors gracefully:

```go
func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    // Attempt external API call
    result, err := p.callExternalAPI()
    if err != nil {
        // Log error but return fallback
        log.Printf("Plugin API error: %v", err)
        return api.AssignmentResult{
            Assignees: []string{"fallback-team"},
        }, nil
    }

    return result, nil
}
```

## Testing Plugins

```go
func TestMyPlugin_Resolve(t *testing.T) {
    plugin := &MyPlugin{}

    config := map[string]string{
        "api_url": "http://test",
        "api_key": "test-key",
    }

    err := plugin.Init(config)
    assert.NoError(t, err)

    files := []api.FileInfo{
        {
            Path:         "terraform/prod/main.tf",
            Size:         1024,
            CommitAuthor: "john.doe",
        },
    }

    result, err := plugin.Resolve(files)
    assert.NoError(t, err)
    assert.Contains(t, result.Assignees, "infra-team")
}
```

## Best Practices

### Error Handling
- Always provide fallback behavior
- Log errors but don't fail the entire process
- Validate input parameters

### Performance
- Cache external API responses when possible
- Implement timeouts for network calls
- Avoid blocking operations

### Configuration
- Validate configuration in `Init`
- Support environment variable substitution
- Provide sensible defaults

### Logging
- Use structured logging
- Include relevant context in log messages
- Don't log sensitive information

### Thread Safety
- Plugins may be called concurrently
- Avoid shared mutable state
- Use locks if necessary

## Plugin Lifecycle

1. **Registration**: Plugin module is registered in ICE
2. **Initialization**: `Init()` called with configuration
3. **Execution**: Plugin methods called during processing
4. **Cleanup**: Plugin instance garbage collected

## Security Considerations

### Input Validation
- Validate all input data
- Sanitize file paths
- Check for malicious patterns

### Authentication
- Store credentials securely
- Use environment variables for secrets
- Implement proper token handling

### Network Security
- Use HTTPS for external calls
- Validate SSL certificates
- Implement request timeouts

## Advanced Topics

### Plugin Dependencies

Plugins can depend on external packages:

```go
// go.mod
module github.com/myorg/ice-plugin

go 1.21

require (
    github.com/baldator/iac-recert-engine v1.0.0
    github.com/external/library v1.2.3
)
```

### Custom Interfaces

Extend plugin interfaces for specialized functionality:

```go
type AdvancedAssignmentPlugin interface {
    AssignmentPlugin
    ValidateConfig(config map[string]string) error
    GetCapabilities() []string
    HealthCheck() error
}
```

### Plugin Metadata

Provide metadata about plugin capabilities:

```go
type PluginMetadata struct {
    Name        string
    Version     string
    Description string
    Author      string
    Capabilities []string
}

func (p *MyPlugin) GetMetadata() PluginMetadata {
    return PluginMetadata{
        Name: "My Custom Plugin",
        Version: "1.0.0",
        Description: "Custom assignment logic",
        Author: "My Organization",
        Capabilities: []string{"assignment", "filter"},
    }
}
```

## Next Steps

- [CLI Commands](cli-commands.md) - Command-line interface
- [Configuration Schema](configuration-schema.md) - Configuration reference
- [Plugins](../configuration/plugins.md) - Plugin configuration guide
