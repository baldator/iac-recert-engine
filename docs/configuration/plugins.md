# Plugins

Plugins extend ICE functionality by integrating with external systems and providing custom logic for assignment, filtering, and other operations. ICE supports a plugin architecture that allows organizations to adapt the tool to their specific workflows and systems.

## Overview

ICE plugins are external modules that implement specific interfaces to provide additional capabilities:

- **Assignment Plugins**: Custom reviewer assignment logic
- **Filter Plugins**: Custom file filtering beyond patterns
- **Strategy Plugins**: Custom PR grouping strategies

Plugins run in the same process as ICE and have access to configuration data and file information.

## Configuration Schema

```yaml
plugins:
  plugin_name:
    enabled: bool        # Enable/disable the plugin
    type: string         # Plugin type: assignment, filter, strategy
    module: string       # Plugin module name
    config:              # Plugin-specific configuration
      key: value
```

## Plugin Types

### Assignment Plugins

Assignment plugins determine how reviewers are assigned to pull requests.

**Interface**:
```go
type AssignmentPlugin interface {
    Init(config map[string]string) error
    Resolve(files []FileInfo) (AssignmentResult, error)
}
```

**Configuration Example**:
```yaml
plugins:
  servicenow_assignment:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      api_url: "https://company.service-now.com"
      username: "${SNOW_USERNAME}"
      password: "${SNOW_PASSWORD}"
```

**Usage in Assignment Strategy**:
```yaml
assignment:
  strategy: "plugin"
  plugin_name: "servicenow_assignment"
  fallback_assignees: ["devops-team"]
```

### Filter Plugins

Filter plugins provide additional file filtering logic beyond glob patterns.

**Interface**:
```go
type FilterPlugin interface {
    Init(config map[string]string) error
    Filter(files []FileInfo) ([]FileInfo, error)
}
```

**Configuration Example**:
```yaml
plugins:
  security_filter:
    enabled: true
    type: "filter"
    module: "security"
    config:
      risk_threshold: "high"
```

### Strategy Plugins

Strategy plugins implement custom PR grouping logic.

**Interface**:
```go
type StrategyPlugin interface {
    Init(config map[string]string) error
    Group(results []RecertCheckResult) ([]FileGroup, error)
}
```

**Configuration Example**:
```yaml
plugins:
  custom_grouping:
    enabled: true
    type: "strategy"
    module: "custom"
    config:
      group_by: "business_unit"
```

**Usage in PR Strategy**:
```yaml
pr_strategy:
  type: "plugin"
  plugin_name: "custom_grouping"
```

## Built-in Plugins

### ServiceNow Assignment Plugin

Integrates with ServiceNow CMDB for automated assignment based on configuration items.

**Features**:
- Queries CMDB for application ownership
- Assigns based on support groups and technical contacts
- Supports change management integration

**Configuration**:
```yaml
plugins:
  servicenow:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      instance_url: "https://company.service-now.com"
      username: "${SNOW_USERNAME}"
      password: "${SNOW_PASSWORD}"
      client_id: "${SNOW_CLIENT_ID}"        # Optional: OAuth client ID
      client_secret: "${SNOW_CLIENT_SECRET}" # Optional: OAuth client secret
      assignment_group_field: "assignment_group"
      technical_contact_field: "u_technical_contact"
      business_owner_field: "owned_by"
```

**Assignment Logic**:
1. Extracts application names from file paths
2. Queries ServiceNow CMDB for matching configuration items
3. Returns assignees based on assignment groups and contacts
4. Falls back to configured defaults if no matches found

**Required Permissions**:
- `cmdb_read` role for CMDB access
- `itil` role for incident/change management
- Access to `cmdb_ci` table

## Developing Custom Plugins

### Plugin Structure

Plugins are Go modules that implement specific interfaces. Create a new directory under `plugins/`:

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
    "encoding/json"
    "net/http"

    "github.com/baldator/iac-recert-engine/pkg/api"
)

type MyPlugin struct {
    apiURL string
    apiKey string
}

func (p *MyPlugin) Init(config map[string]string) error {
    p.apiURL = config["api_url"]
    p.apiKey = config["api_key"]
    return nil
}

func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    // Custom assignment logic
    assignees := []string{"team-lead"}

    for _, file := range files {
        // Analyze file path, content, etc.
        if strings.Contains(file.Path, "production") {
            assignees = append(assignees, "security-team")
        }
    }

    return api.AssignmentResult{
        Assignees: assignees,
        Reviewers: []string{"optional-reviewer"},
        Team:      "platform-team",
        Priority:  "high",
    }, nil
}

// Export for plugin loading
var Plugin = &MyPlugin{}
```

### Plugin Registration

Update `internal/plugin/plugin.go` to register your plugin:

```go
case "myplugin":
    if cfg.Type != "assignment" {
        return nil, fmt.Errorf("myplugin plugin must be of type assignment")
    }
    apiPlugin := myplugin.Plugin
    plugin = &assignmentPluginWrapper{apiPlugin: apiPlugin}
```

### Plugin API

The plugin API provides access to:

```go
type FileInfo struct {
    Path         string    // File path
    Size         int64     // File size in bytes
    LastModified time.Time // Last modification time
    CommitHash   string    // Latest commit hash
    CommitAuthor string    // Last committer
    CommitEmail  string    // Committer email
    CommitMsg    string    // Commit message
}

type AssignmentResult struct {
    Assignees []string // Primary assignees
    Reviewers []string // Additional reviewers
    Team      string   // Team assignment
    Priority  string   // Priority level
}
```

## Plugin Configuration

### Environment Variables

Plugins support environment variable substitution:

```yaml
plugins:
  myplugin:
    config:
      api_key: "${MYPLUGIN_API_KEY}"
      database_url: "${DATABASE_URL}"
```

### Secure Configuration

Store sensitive configuration in environment variables or secret managers:

```bash
export MYPLUGIN_API_KEY="secret-key"
export DATABASE_URL="postgres://user:pass@host/db"
```

### Configuration Validation

Plugins should validate their configuration in the `Init` method:

```go
func (p *MyPlugin) Init(config map[string]string) error {
    if config["api_url"] == "" {
        return errors.New("api_url is required")
    }
    if config["api_key"] == "" {
        return errors.New("api_key is required")
    }
    // Additional validation...
    return nil
}
```

## Plugin Deployment

### Local Development

1. Create plugin directory under `plugins/`
2. Implement plugin interface
3. Register plugin in `internal/plugin/plugin.go`
4. Test with local configuration

### Production Deployment

1. Build plugin as part of ICE binary
2. Configure plugin settings
3. Set environment variables
4. Test in staging environment

### Plugin Updates

- Plugins are loaded at startup
- Configuration changes require restart
- Update plugin code and rebuild ICE binary

## Plugin Best Practices

### Error Handling

```go
func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    result, err := p.callExternalAPI()
    if err != nil {
        // Log error but return fallback result
        log.Printf("Plugin error: %v", err)
        return api.AssignmentResult{
            Assignees: []string{"fallback-team"},
        }, nil
    }
    return result, nil
}
```

### Performance

- Cache external API calls when possible
- Implement timeouts for network requests
- Avoid blocking operations

### Logging

```go
import "go.uber.org/zap"

type MyPlugin struct {
    logger *zap.Logger
}

func (p *MyPlugin) Init(config map[string]string) error {
    p.logger = zap.L() // Use ICE's logger
    p.logger.Info("Plugin initialized", zap.String("config", "..."))
    return nil
}
```

### Testing

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
        {Path: "test/file.tf"},
    }

    result, err := plugin.Resolve(files)
    assert.NoError(t, err)
    assert.Contains(t, result.Assignees, "expected-assignee")
}
```

## Troubleshooting Plugins

### Plugin Not Loading

- Check plugin is enabled in configuration
- Verify module name matches registration
- Check plugin initialization errors in logs

### Plugin Errors

- Review plugin logs for error messages
- Test plugin configuration independently
- Verify external system connectivity

### Performance Issues

- Monitor plugin execution time
- Check for network timeouts
- Implement caching for repeated calls

### Configuration Issues

- Validate environment variables are set
- Check configuration syntax
- Use `--verbose` flag for detailed logging

## Security Considerations

### Authentication

- Store API keys in environment variables
- Use OAuth when possible
- Rotate credentials regularly

### Data Handling

- Don't log sensitive information
- Encrypt data in transit
- Validate input data

### Network Security

- Use HTTPS for external calls
- Implement request timeouts
- Validate SSL certificates

## Advanced Topics

### Multiple Plugins

Configure multiple plugins of different types:

```yaml
plugins:
  servicenow_assignment:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config: {...}

  security_filter:
    enabled: true
    type: "filter"
    module: "security"
    config: {...}

  custom_strategy:
    enabled: true
    type: "strategy"
    module: "custom"
    config: {...}
```

### Plugin Chains

Plugins can be combined for complex workflows:

```yaml
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/**"
      strategy: "plugin"
      plugin: "servicenow"
    - pattern: "k8s/**"
      strategy: "plugin"
      plugin: "kubernetes_owners"
```

### Custom Plugin Interfaces

Extend plugin interfaces for specialized use cases:

```go
type AdvancedAssignmentPlugin interface {
    AssignmentPlugin
    ValidateConfig(config map[string]string) error
    GetCapabilities() []string
}
```

## Next Steps

- [Assignment Strategies](assignment-strategies.md) - Using assignment plugins
- [PR Strategies](pr-strategies.md) - Custom strategy plugins
- [Development](../development/setup.md) - Setting up development environment
