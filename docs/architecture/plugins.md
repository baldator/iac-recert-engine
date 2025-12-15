# Plugin System

ICE features a comprehensive plugin architecture that enables extensibility and integration with external systems. This document details the plugin system design, implementation, and usage patterns.

## Plugin Architecture Overview

### Core Principles

The ICE plugin system is built on several key principles:

- **Type Safety**: Strongly typed interfaces prevent runtime errors
- **Isolation**: Plugin failures don't affect core ICE functionality
- **Lifecycle Management**: Proper initialization, execution, and cleanup
- **Configuration**: Plugin-specific configuration support
- **Observability**: Logging, metrics, and error tracking

### Plugin Types

ICE supports multiple plugin types for different extension points:

```go
// Assignment plugins determine PR reviewers
type AssignmentPlugin interface {
    Plugin
    Resolve(files []FileInfo) (AssignmentResult, error)
}

// Filter plugins provide additional file filtering
type FilterPlugin interface {
    Plugin
    Filter(files []FileInfo) ([]FileInfo, error)
}

// Strategy plugins implement custom PR grouping
type StrategyPlugin interface {
    Plugin
    Group(results []RecertCheckResult) ([]FileGroup, error)
}
```

### Plugin Lifecycle

```
1. Discovery    → Plugin module is loaded
2. Validation   → Plugin interface compliance checked
3. Initialization → Plugin.Init() called with config
4. Execution    → Plugin methods called during processing
5. Cleanup      → Plugin resources released
```

## Plugin Manager

### Architecture

The plugin manager coordinates plugin loading, execution, and lifecycle:

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

### Plugin Loading Process

1. **Module Resolution**: Locate plugin module by name
2. **Interface Validation**: Ensure plugin implements required interfaces
3. **Configuration Parsing**: Convert config map to plugin format
4. **Initialization**: Call plugin Init method
5. **Registration**: Store plugin in manager registry

### Error Isolation

Plugin errors are contained to prevent system-wide failures:

```go
func (m *Manager) executePlugin(name string, operation func(Plugin) error) error {
    plugin, exists := m.plugins[name]
    if !exists {
        return fmt.Errorf("plugin %s not found", name)
    }

    // Execute in isolated context
    defer func() {
        if r := recover(); r != nil {
            m.logger.Error("plugin panic recovered",
                zap.String("plugin", name),
                zap.Any("panic", r))
        }
    }()

    return operation(plugin)
}
```

## Plugin Implementation

### Base Plugin Interface

All plugins must implement the base Plugin interface:

```go
type Plugin interface {
    Init(config map[string]string) error
}
```

**Init Method Requirements:**
- **Idempotent**: Multiple calls should be safe
- **Validation**: Validate configuration parameters
- **Resource Allocation**: Initialize connections, caches, etc.
- **Error Handling**: Return descriptive errors

### Assignment Plugins

#### Interface
```go
type AssignmentPlugin interface {
    Plugin
    Resolve(files []FileInfo) (AssignmentResult, error)
}
```

#### Implementation Example
```go
type CMDBAssignmentPlugin struct {
    apiClient *http.Client
    baseURL   string
    apiKey    string
}

func (p *CMDBAssignmentPlugin) Init(config map[string]string) error {
    p.baseURL = config["api_url"]
    if p.baseURL == "" {
        return errors.New("api_url is required")
    }

    p.apiKey = config["api_key"]
    if p.apiKey == "" {
        return errors.New("api_key is required")
    }

    p.apiClient = &http.Client{
        Timeout: 30 * time.Second,
    }

    return nil
}

func (p *CMDBAssignmentPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    if len(files) == 0 {
        return api.AssignmentResult{}, errors.New("no files provided")
    }

    // Extract application identifiers from file paths
    apps := p.extractApplications(files)

    // Query CMDB for ownership information
    owners, err := p.queryCMDBOwnership(apps)
    if err != nil {
        p.logger.Warn("CMDB query failed, using fallback", zap.Error(err))
        return api.AssignmentResult{
            Assignees: []string{"platform-team"}, // fallback
        }, nil
    }

    return api.AssignmentResult{
        Assignees: owners,
        Reviewers: []string{"security-team"},
        Team:      "platform-team",
    }, nil
}
```

#### Data Structures
```go
type FileInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    CommitHash   string
    CommitAuthor string
    CommitEmail  string
    CommitMsg    string
}

type AssignmentResult struct {
    Assignees []string // Primary reviewers
    Reviewers []string // Additional reviewers
    Team      string   // Team assignment
    Priority  string   // Priority level
}
```

### Filter Plugins

#### Interface
```go
type FilterPlugin interface {
    Plugin
    Filter(files []FileInfo) ([]FileInfo, error)
}
```

#### Use Cases
- Security scanning integration
- File size limits
- Content-based filtering
- Compliance rule enforcement

#### Implementation Example
```go
type SecurityFilterPlugin struct {
    scannerURL string
    apiKey     string
    riskThreshold string
}

func (p *SecurityFilterPlugin) Init(config map[string]string) error {
    p.scannerURL = config["scanner_url"]
    p.apiKey = config["api_key"]
    p.riskThreshold = config["risk_threshold"]

    if p.riskThreshold == "" {
        p.riskThreshold = "medium"
    }

    return nil
}

func (p *SecurityFilterPlugin) Filter(files []api.FileInfo) ([]api.FileInfo, error) {
    var filtered []api.FileInfo

    for _, file := range files {
        // Perform security scan
        risk, err := p.scanFile(file)
        if err != nil {
            return nil, fmt.Errorf("security scan failed for %s: %w", file.Path, err)
        }

        // Filter based on risk threshold
        if p.shouldInclude(risk) {
            filtered = append(filtered, file)
        } else {
            p.logger.Info("file filtered due to security risk",
                zap.String("file", file.Path),
                zap.String("risk", risk))
        }
    }

    return filtered, nil
}
```

### Strategy Plugins

#### Interface
```go
type StrategyPlugin interface {
    Plugin
    Group(results []RecertCheckResult) ([]FileGroup, error)
}
```

#### Implementation Example
```go
type BusinessUnitStrategyPlugin struct {
    groupByField string
    maxFilesPerPR int
}

func (p *BusinessUnitStrategyPlugin) Init(config map[string]string) error {
    p.groupByField = config["group_by_field"]
    if p.groupByField == "" {
        p.groupByField = "business_unit"
    }

    if maxFiles := config["max_files_per_pr"]; maxFiles != "" {
        if _, err := fmt.Sscanf(maxFiles, "%d", &p.maxFilesPerPR); err != nil {
            return fmt.Errorf("invalid max_files_per_pr: %v", err)
        }
    }

    return nil
}

func (p *BusinessUnitStrategyPlugin) Group(results []api.RecertCheckResult) ([]api.FileGroup, error) {
    groups := make(map[string][]api.RecertCheckResult)

    for _, result := range results {
        if !result.NeedsRecert {
            continue
        }

        // Extract business unit from file path or metadata
        bu := p.extractBusinessUnit(result.File.Path)

        groups[bu] = append(groups[bu], result)
    }

    var fileGroups []api.FileGroup
    for bu, files := range groups {
        // Apply max files per PR limit
        if p.maxFilesPerPR > 0 && len(files) > p.maxFilesPerPR {
            fileGroups = append(fileGroups, p.splitGroup(bu, files)...)
        } else {
            fileGroups = append(fileGroups, api.FileGroup{
                ID:       fmt.Sprintf("bu-%s", bu),
                Strategy: "business_unit",
                Files:    files,
            })
        }
    }

    return fileGroups, nil
}
```

## Plugin Configuration

### Configuration Schema

Plugins are configured in the main ICE configuration:

```yaml
plugins:
  cmdb_assignment:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      api_url: "https://company.service-now.com"
      username: "${SNOW_USERNAME}"
      password: "${SNOW_PASSWORD}"
      assignment_group_field: "assignment_group"

  security_filter:
    enabled: true
    type: "filter"
    module: "security"
    config:
      scanner_url: "https://security.company.com/scan"
      risk_threshold: "high"
```

### Environment Variable Substitution

Plugin configurations support environment variables:

```yaml
plugins:
  myplugin:
    config:
      api_key: "${PLUGIN_API_KEY}"
      database_url: "${DATABASE_URL}"
      timeout: "${PLUGIN_TIMEOUT}"
```

### Configuration Validation

Plugins validate their configuration during initialization:

```go
func (p *MyPlugin) Init(config map[string]string) error {
    // Required parameters
    if config["api_url"] == "" {
        return errors.New("api_url is required")
    }

    // Optional parameters with defaults
    p.timeout = 30 * time.Second
    if timeoutStr := config["timeout"]; timeoutStr != "" {
        if timeout, err := time.ParseDuration(timeoutStr); err == nil {
            p.timeout = timeout
        }
    }

    // Type validation
    if maxRetriesStr := config["max_retries"]; maxRetriesStr != "" {
        if maxRetries, err := strconv.Atoi(maxRetriesStr); err == nil {
            p.maxRetries = maxRetries
        } else {
            return fmt.Errorf("invalid max_retries: %s", maxRetriesStr)
        }
    }

    return nil
}
```

## Plugin Execution Model

### Synchronous Execution

Plugins execute synchronously within the main ICE workflow:

```
ICE Workflow:
  1. Scan files
  2. Apply filters (plugin)
  3. Check recertification
  4. Group files (strategy plugin)
  5. Assign reviewers (assignment plugin)
  6. Create PRs
```

### Error Handling

Plugin errors are handled gracefully:

```go
func (m *Manager) executeWithFallback(operation string, pluginName string, fallback func() error) error {
    if plugin, exists := m.plugins[pluginName]; exists {
        if err := m.executePlugin(pluginName, func(p Plugin) error {
            // Execute plugin operation
            return operation(p)
        }); err != nil {
            m.logger.Warn("plugin execution failed, using fallback",
                zap.String("plugin", pluginName),
                zap.String("operation", operation),
                zap.Error(err))

            return fallback()
        }
    }

    return fallback()
}
```

### Timeout Protection

Plugins are protected against hanging:

```go
func (m *Manager) executeWithTimeout(pluginName string, timeout time.Duration, operation func(Plugin) error) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    resultChan := make(chan error, 1)

    go func() {
        resultChan <- m.executePlugin(pluginName, operation)
    }()

    select {
    case err := <-resultChan:
        return err
    case <-ctx.Done():
        return fmt.Errorf("plugin %s timed out after %v", pluginName, timeout)
    }
}
```

## Plugin Packaging and Distribution

### Directory Structure

Plugins are organized in the `plugins/` directory:

```
plugins/
├── servicenow/
│   ├── go.mod
│   ├── servicenow/
│   │   └── plugin.go
│   └── README.md
├── ldap/
│   ├── go.mod
│   ├── ldap/
│   │   └── plugin.go
│   └── README.md
└── custom/
    ├── go.mod
    ├── custom/
    │   └── plugin.go
    └── README.md
```

### Module Naming

Plugins use Go module naming conventions:

```go
// go.mod
module github.com/myorg/ice-plugin-servicenow

go 1.21

require (
    github.com/baldator/iac-recert-engine v1.0.0
    github.com/levigross/grequests v0.0.0-20231203190009-47c5c1d3fab8
)
```

### Plugin Registration

Plugins register themselves in the main ICE binary:

```go
// internal/plugin/manager.go
import (
    // ... other imports
    _ "github.com/myorg/ice-plugin-servicenow/servicenow"
)

func (m *Manager) loadPlugin(name string, cfg PluginConfig) (Plugin, error) {
    switch cfg.Module {
    case "servicenow":
        return servicenow.NewPlugin(), nil
    // ... other plugins
    default:
        return nil, fmt.Errorf("unknown plugin module: %s", cfg.Module)
    }
}
```

## Built-in Plugins

### ServiceNow Assignment Plugin

Integrates with ServiceNow CMDB for automated assignment:

**Features:**
- CMDB CI relationship queries
- Assignment group resolution
- Technical contact lookup
- Change management integration

**Configuration:**
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
      client_id: "${SNOW_CLIENT_ID}"
      client_secret: "${SNOW_CLIENT_SECRET}"
```

### LDAP Assignment Plugin

Integrates with LDAP/Active Directory:

**Features:**
- User and group lookups
- Organizational hierarchy
- Attribute-based assignment
- Kerberos authentication

**Configuration:**
```yaml
plugins:
  ldap:
    enabled: true
    type: "assignment"
    module: "ldap"
    config:
      server: "ldap.company.com"
      port: "389"
      base_dn: "ou=users,dc=company,dc=com"
      bind_user: "${LDAP_BIND_USER}"
      bind_password: "${LDAP_BIND_PASSWORD}"
      user_filter: "(sAMAccountName=%s)"
      group_filter: "(member=%s)"
```

## Plugin Development Best Practices

### Error Handling

Always provide meaningful error messages:

```go
func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    if len(files) == 0 {
        return api.AssignmentResult{}, fmt.Errorf("cannot resolve assignment: no files provided")
    }

    result, err := p.callExternalAPI()
    if err != nil {
        return api.AssignmentResult{}, fmt.Errorf("external API call failed: %w", err)
    }

    return result, nil
}
```

### Logging

Use structured logging for observability:

```go
func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    p.logger.Info("resolving assignment",
        zap.Int("file_count", len(files)),
        zap.String("first_file", files[0].Path))

    // ... resolution logic

    p.logger.Info("assignment resolved",
        zap.Strings("assignees", result.Assignees),
        zap.String("team", result.Team))

    return result, nil
}
```

### Resource Management

Properly manage resources:

```go
type MyPlugin struct {
    httpClient *http.Client
    dbConn     *sql.DB
}

func (p *MyPlugin) Init(config map[string]string) error {
    p.httpClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns: 10,
        },
    }

    // Initialize database connection
    db, err := sql.Open("postgres", config["database_url"])
    if err != nil {
        return err
    }

    if err := db.Ping(); err != nil {
        return err
    }

    p.dbConn = db
    return nil
}

func (p *MyPlugin) Close() error {
    if p.dbConn != nil {
        return p.dbConn.Close()
    }
    return nil
}
```

### Testing

Comprehensive testing for plugins:

```go
func TestCMDBAssignmentPlugin_Resolve(t *testing.T) {
    plugin := &CMDBAssignmentPlugin{}

    config := map[string]string{
        "api_url": "http://test",
        "api_key": "test-key",
    }

    err := plugin.Init(config)
    assert.NoError(t, err)

    files := []api.FileInfo{
        {Path: "apps/myapp/main.tf"},
    }

    result, err := plugin.Resolve(files)
    assert.NoError(t, err)
    assert.Contains(t, result.Assignees, "platform-team")
}

func TestCMDBAssignmentPlugin_InitValidation(t *testing.T) {
    plugin := &CMDBAssignmentPlugin{}

    // Test missing required config
    err := plugin.Init(map[string]string{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "api_url is required")
}
```

### Documentation

Provide comprehensive documentation:

```go
// Package servicenow provides ServiceNow CMDB integration for ICE
//
// This plugin integrates with ServiceNow Configuration Management Database (CMDB)
// to automatically assign pull request reviewers based on application ownership
// and support group assignments.
//
// Configuration:
//   - instance_url: ServiceNow instance URL
//   - username: ServiceNow username
//   - password: ServiceNow password
//   - assignment_group_field: Field name for assignment groups
//
// Example:
//   plugins:
//     servicenow:
//       enabled: true
//       type: "assignment"
//       module: "servicenow"
//       config:
//         instance_url: "https://company.service-now.com"
//         username: "${SNOW_USER}"
//         password: "${SNOW_PASS}"
package servicenow
```

## Security Considerations

### Authentication

Secure credential handling:

```go
func (p *MyPlugin) Init(config map[string]string) error {
    // Validate HTTPS URLs
    if !strings.HasPrefix(config["api_url"], "https://") {
        return errors.New("api_url must use HTTPS")
    }

    // Use secure credential storage
    p.apiKey = config["api_key"]
    if p.apiKey == "" {
        return errors.New("api_key is required")
    }

    return nil
}
```

### Input Validation

Validate all inputs to prevent injection:

```go
func (p *MyPlugin) sanitizeInput(input string) string {
    // Remove potentially dangerous characters
    return strings.ReplaceAll(input, "'", "''")
}

func (p *MyPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
    // Validate file paths
    for _, file := range files {
        if strings.Contains(file.Path, "..") {
            return api.AssignmentResult{}, errors.New("invalid file path")
        }
    }

    // Sanitize inputs before external calls
    sanitizedPaths := make([]string, len(files))
    for i, file := range files {
        sanitizedPaths[i] = p.sanitizeInput(file.Path)
    }

    return p.queryExternalSystem(sanitizedPaths)
}
```

### Network Security

Implement secure network communications:

```go
func (p *MyPlugin) createHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
                CipherSuites: []uint16{
                    tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
                },
            },
            MaxIdleConns: 10,
            IdleConnTimeout: 90 * time.Second,
        },
    }
}
```

## Performance Optimization

### Caching

Implement caching for expensive operations:

```go
type PluginCache struct {
    mu     sync.RWMutex
    data   map[string]interface{}
    expiry map[string]time.Time
}

func (c *PluginCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if expiry, exists := c.expiry[key]; exists && time.Now().After(expiry) {
        return nil, false
    }

    value, exists := c.data[key]
    return value, exists
}

func (c *PluginCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.data[key] = value
    c.expiry[key] = time.Now().Add(ttl)
}
```

### Connection Pooling

Reuse connections for better performance:

```go
func (p *MyPlugin) Init(config map[string]string) error {
    p.httpClient = &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
        Timeout: 30 * time.Second,
    }

    return nil
}
```

### Batch Operations

Process multiple items efficiently:

```go
func (p *MyPlugin) ResolveBatch(files []api.FileInfo, batchSize int) (api.AssignmentResult, error) {
    var allAssignees []string

    for i := 0; i < len(files); i += batchSize {
        end := i + batchSize
        if end > len(files) {
            end = len(files)
        }

        batch := files[i:end]
        result, err := p.resolveBatch(batch)
        if err != nil {
            return api.AssignmentResult{}, err
        }

        allAssignees = append(allAssignees, result.Assignees...)
    }

    // Deduplicate assignees
    uniqueAssignees := removeDuplicates(allAssignees)

    return api.AssignmentResult{
        Assignees: uniqueAssignees,
    }, nil
}
```

## Monitoring and Observability

### Metrics

Expose plugin performance metrics:

```go
type PluginMetrics struct {
    operationsTotal    prometheus.Counter
    operationDuration  prometheus.Histogram
    errorsTotal        prometheus.Counter
    cacheHits          prometheus.Counter
    cacheMisses        prometheus.Counter
}

func (p *MyPlugin) observeOperation(operation string, start time.Time, err error) {
    duration := time.Since(start).Seconds()

    p.metrics.operationDuration.WithLabelValues(operation).Observe(duration)

    if err != nil {
        p.metrics.errorsTotal.WithLabelValues(operation).Inc()
    }
}
```

### Health Checks

Implement plugin health monitoring:

```go
func (p *MyPlugin) HealthCheck() error {
    // Test external system connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", p.healthCheckURL, nil)
    if err != nil {
        return err
    }

    resp, err := p.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
    }

    return nil
}
```

## Next Steps

- [Core Components](components.md) - Component architecture details
- [Git Providers](git-providers.md) - Provider-specific details
- [Security](security.md) - Security considerations
