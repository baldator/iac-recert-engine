# Assignment Strategies

Assignment strategies determine how reviewers and assignees are assigned to recertification pull requests. Different strategies support various team structures, ownership models, and approval workflows.

## Configuration Schema

```yaml
assignment:
  strategy: string              # Required: static, last_committer, plugin, composite
  plugin_name: string           # Optional: Plugin name for plugin strategy
  rules:                        # Optional: Rules for composite strategy
    - pattern: string           # File pattern to match
      strategy: string          # Strategy for this pattern
      plugin: string            # Plugin name (if strategy is plugin)
      fallback_assignees:       # Fallback assignees for this rule
        - string
  fallback_assignees:           # Required: Default assignees
    - string
```

## Available Strategies

### Static Strategy (`static`)

Assigns the same reviewers to all pull requests.

**Use Cases**:
- Centralized approval teams
- Small organizations with shared responsibility
- Compliance-focused environments

**Configuration**:
```yaml
assignment:
  strategy: "static"
  fallback_assignees:
    - "security-team"
    - "compliance-officer"
```

**Behavior**:
- All PRs get the same assignees
- Simple and predictable
- No analysis of file content or history

**Pros**:
- Consistent approval process
- Easy to configure
- Clear ownership

**Cons**:
- No specialization by technology/domain
- May bottleneck on busy teams
- Doesn't leverage subject matter expertise

### Last Committer Strategy (`last_committer`)

Assigns the last person who modified each file as the reviewer.

**Use Cases**:
- Individual accountability
- Subject matter experts review their code
- Code ownership by last modifier

**Configuration**:
```yaml
assignment:
  strategy: "last_committer"
  fallback_assignees:
    - "devops-team"  # Used when committer not found
```

**Behavior**:
- Analyzes git history to find last committer
- Assigns most recent committer across all files in PR
- Falls back to default assignees if no committer found

**Example**:
- PR contains files last modified by `john.doe@example.com`
- Assigns `john.doe` as reviewer
- If email not found in git, uses fallback assignees

**Pros**:
- Individual accountability
- Subject matter experts
- Natural ownership model

**Cons**:
- Requires accurate git history
- May assign inactive users
- Doesn't consider team structure

### Plugin Strategy (`plugin`)

Uses external systems or custom logic for assignment.

**Use Cases**:
- Integration with CMDB or ticketing systems
- Complex organizational rules
- Dynamic assignment based on external data

**Configuration**:
```yaml
assignment:
  strategy: "plugin"
  plugin_name: "servicenow_assignment"
  fallback_assignees:
    - "devops-team"
```

**Behavior**:
- Calls configured plugin with file information
- Plugin returns assignees based on custom logic
- Falls back to default assignees on plugin failure

**Pros**:
- Highly flexible
- Can integrate with external systems
- Adaptable to complex requirements

**Cons**:
- Requires plugin development
- Additional complexity
- Plugin maintenance overhead

### Composite Strategy (`composite`)

Applies different assignment strategies based on file patterns.

**Use Cases**:
- Different technologies have different owners
- Mixed environments with specialized teams
- Complex organizational structures

**Configuration**:
```yaml
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/prod/**"
      strategy: "static"
      fallback_assignees: ["infra-team"]
    - pattern: "k8s/**"
      strategy: "last_committer"
    - pattern: "policies/**"
      strategy: "static"
      fallback_assignees: ["security-team", "compliance"]
  fallback_assignees:
    - "devops-team"
```

**Behavior**:
- Matches files against patterns in order
- Applies first matching rule's strategy
- Uses fallback assignees if no rule matches

**Example**:
- `terraform/prod/main.tf` → assigned to `infra-team`
- `k8s/deployment.yaml` → assigned to last committer
- `policies/security.rego` → assigned to `security-team` and `compliance`
- `docs/README.md` → assigned to `devops-team` (fallback)

**Pros**:
- Fine-grained control
- Supports complex organizations
- Combines multiple strategies

**Cons**:
- Complex configuration
- Pattern matching overhead
- Requires careful rule ordering

## Advanced Configuration

### Pattern Matching in Composite Strategy

Rules are evaluated in order, with the first match winning:

```yaml
assignment:
  strategy: "composite"
  rules:
    # Most specific patterns first
    - pattern: "terraform/prod/critical/**"
      strategy: "static"
      fallback_assignees: ["security-team", "infra-lead"]
    - pattern: "terraform/prod/**"
      strategy: "static"
      fallback_assignees: ["infra-team"]
    # General patterns last
    - pattern: "terraform/**"
      strategy: "last_committer"
```

### Plugin Integration

Plugins can return multiple types of assignments:

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

### Fallback Handling

Multiple fallback levels ensure assignments always work:

```yaml
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/**"
      strategy: "plugin"
      plugin: "servicenow"
      fallback_assignees: ["terraform-team"]  # Rule-specific fallback
  fallback_assignees:  # Global fallback
    - "devops-team"
    - "platform-team"
```

## Assignment Results

Strategies return assignment results with multiple fields:

```go
type AssignmentResult struct {
    Assignees []string  // Primary assignees (required reviewers)
    Reviewers []string  // Additional reviewers (optional)
    Team      string    // Team assignment
    Priority  string    // Priority level
}
```

### Using Assignment Results

Different Git providers handle assignments differently:

**GitHub**:
- Assignees become PR assignees
- Reviewers become requested reviewers
- Team assignments may create team reviews

**Azure DevOps**:
- Assignees become work item assignees
- Reviewers become additional reviewers
- Priority may set work item priority

**GitLab**:
- Assignees become MR assignees
- Reviewers become requested reviewers
- Team assignments may use group mentions

## Strategy Selection Guide

### Choose Based on Organization Size

**Small Organizations (< 50 people)**:
```yaml
assignment:
  strategy: "static"
  fallback_assignees: ["platform-team"]
```

**Medium Organizations (50-500 people)**:
```yaml
assignment:
  strategy: "composite"
  rules:
    - pattern: "terraform/**"
      strategy: "static"
      fallback_assignees: ["infra-team"]
    - pattern: "k8s/**"
      strategy: "last_committer"
  fallback_assignees: ["devops-team"]
```

**Large Organizations (500+ people)**:
```yaml
assignment:
  strategy: "plugin"
  plugin_name: "cmdb_assignment"
  fallback_assignees: ["platform-team"]
```

### Choose Based on Team Structure

**Technology-Specialized Teams**:
```yaml
assignment:
  strategy: "composite"
  rules:
    - pattern: "**/*.tf"
      strategy: "static"
      fallback_assignees: ["terraform-team"]
    - pattern: "**/*.yaml"
      strategy: "static"
      fallback_assignees: ["kubernetes-team"]
```

**Feature Teams**:
```yaml
assignment:
  strategy: "last_committer"
  fallback_assignees: ["devops-team"]
```

**Centralized Approval**:
```yaml
assignment:
  strategy: "static"
  fallback_assignees: ["security", "compliance", "architecture"]
```

## Performance Considerations

### Strategy Performance
- `static`: Fastest, no analysis required
- `last_committer`: Moderate, requires git history analysis
- `composite`: Moderate, requires pattern matching
- `plugin`: Variable, depends on plugin implementation

### Git History Analysis
For `last_committer` strategy:
- Scans git log for each file
- Caches results to improve performance
- May be slow for large repositories

### Plugin Performance
- Network calls to external systems
- Database queries in CMDB systems
- Consider caching for frequently accessed data

## Troubleshooting

### No Assignments Applied
- Check fallback assignees are configured
- Verify strategy is valid
- Use `--verbose` to see assignment decisions

### Wrong Assignments
- For `last_committer`: Check git history is accessible
- For `composite`: Verify pattern matching order
- For `plugin`: Check plugin logs and configuration

### Plugin Failures
- Verify plugin is loaded and enabled
- Check plugin configuration and credentials
- Test plugin independently of ICE

### Git History Issues
- Ensure repository has full git history
- Check file permissions for git operations
- Verify git is available in PATH

## Best Practices

1. **Start simple** - Begin with `static` or `last_committer`
2. **Use composite for complexity** - Combine strategies for different technologies
3. **Always configure fallbacks** - Ensure assignments never fail
4. **Test assignments** - Use dry-run to verify correct assignments
5. **Document ownership** - Explain why certain assignments are made
6. **Monitor effectiveness** - Track review times and adjust strategies
7. **Review periodically** - Team structures and ownership change

## Integration Examples

### ServiceNow CMDB Integration
```yaml
assignment:
  strategy: "plugin"
  plugin_name: "servicenow"
  fallback_assignees: ["platform-team"]

plugins:
  servicenow:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      instance_url: "https://company.service-now.com"
      username: "${SNOW_USER}"
      password: "${SNOW_PASS}"
      assignment_group: "Infrastructure Team"
```

### LDAP/Active Directory Integration
```yaml
assignment:
  strategy: "plugin"
  plugin_name: "ldap_assignment"
  fallback_assignees: ["devops"]

plugins:
  ldap_assignment:
    enabled: true
    type: "assignment"
    module: "ldap"
    config:
      server: "ldap.company.com"
      base_dn: "ou=users,dc=company,dc=com"
      bind_user: "${LDAP_USER}"
      bind_pass: "${LDAP_PASS}"
```

## Next Steps

- [Plugins](plugins.md) - Custom assignment plugin development
- [PR Strategies](pr-strategies.md) - How files are grouped into PRs
- [Configuration Overview](../configuration/overview.md) - Complete configuration reference
