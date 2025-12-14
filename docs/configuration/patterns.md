# Patterns

Patterns define which files ICE will scan for recertification and how they should be processed. Each pattern specifies file matching rules, recertification schedules, and optional decorators.

## Configuration Schema

```yaml
patterns:
  - name: string              # Required: Unique pattern identifier
    description: string       # Optional: Human-readable description
    paths:                    # Required: Glob patterns to include
      - string
    exclude:                  # Optional: Glob patterns to exclude
      - string
    recertification_days: int # Required: Days before recertification needed
    enabled: bool             # Optional: Enable/disable pattern (default: true)
    decorator: string         # Optional: Text to add when recertifying
```

## Basic Configuration

### Simple Pattern
```yaml
patterns:
  - name: "terraform"
    description: "All Terraform files"
    paths:
      - "**/*.tf"
    recertification_days: 90
```

### Multiple Paths
```yaml
patterns:
  - name: "infrastructure"
    description: "Infrastructure as Code"
    paths:
      - "**/*.tf"
      - "**/*.yaml"
      - "**/*.yml"
    recertification_days: 180
```

## Path Matching

### Glob Patterns
ICE uses glob patterns for file matching:

```yaml
patterns:
  - name: "terraform-prod"
    paths:
      - "terraform/prod/**/*.tf"    # All .tf files in terraform/prod/
      - "infrastructure/**/*.tf"     # All .tf files in infrastructure/
      - "!modules/**"                # Exclude modules directory
```

### Supported Glob Syntax
- `*`: Matches any sequence of characters (excluding path separator)
- `**`: Matches any sequence of characters (including path separators)
- `?`: Matches any single character
- `[abc]`: Matches any character in the set
- `{a,b,c}`: Matches any of the comma-separated alternatives

### Directory-Specific Patterns
```yaml
patterns:
  - name: "kubernetes-prod"
    paths:
      - "k8s/prod/**/*.yaml"
      - "k8s/prod/**/*.yml"
    recertification_days: 90

  - name: "kubernetes-staging"
    paths:
      - "k8s/staging/**/*.yaml"
      - "k8s/staging/**/*.yml"
    recertification_days: 30
```

## Exclusion Patterns

### Basic Exclusions
```yaml
patterns:
  - name: "terraform"
    paths:
      - "**/*.tf"
    exclude:
      - "**/*.test.tf"        # Exclude test files
      - "**/modules/**"       # Exclude modules
      - "**/examples/**"      # Exclude examples
    recertification_days: 90
```

### Complex Exclusions
```yaml
patterns:
  - name: "infrastructure"
    paths:
      - "**/*"
    exclude:
      - "**/.git/**"          # Exclude git directory
      - "**/node_modules/**"  # Exclude dependencies
      - "**/*.log"            # Exclude log files
      - "**/tmp/**"           # Exclude temporary files
    recertification_days: 180
```

## Recertification Schedule

### Time-Based Recertification
```yaml
patterns:
  - name: "critical-infrastructure"
    paths: ["terraform/prod/**/*.tf"]
    recertification_days: 90    # Recertify every 90 days

  - name: "development"
    paths: ["terraform/dev/**/*.tf"]
    recertification_days: 30    # Recertify every 30 days

  - name: "experimental"
    paths: ["terraform/lab/**/*.tf"]
    recertification_days: 7     # Recertify weekly
```

### Pattern-Specific Schedules
Different patterns can have different recertification requirements:

```yaml
patterns:
  - name: "security-policies"
    description: "Security policies and compliance rules"
    paths: ["policies/**/*.rego"]
    recertification_days: 30

  - name: "network-config"
    description: "Network infrastructure configuration"
    paths: ["network/**/*.cfg"]
    recertification_days: 180

  - name: "application-code"
    description: "Application source code"
    paths: ["src/**/*.go"]
    recertification_days: 365
```

## File Decorators

### Timestamp Decorators
Add recertification timestamps to files:

```yaml
patterns:
  - name: "terraform"
    paths: ["**/*.tf"]
    recertification_days: 90
    decorator: "# Last Recertification: {timestamp}\n"
```

### Language-Specific Decorators
```yaml
patterns:
  - name: "golang"
    paths: ["**/*.go"]
    recertification_days: 180
    decorator: "// Last Recertification: {timestamp}\n"

  - name: "python"
    paths: ["**/*.py"]
    recertification_days: 180
    decorator: "# Last Recertification: {timestamp}\n"

  - name: "yaml"
    paths: ["**/*.yaml", "**/*.yml"]
    recertification_days: 90
    decorator: "# Last Recertification: {timestamp}\n"
```

### Custom Messages
```yaml
patterns:
  - name: "policies"
    paths: ["policies/**/*.rego"]
    recertification_days: 30
    decorator: |
      # Last Recertification: {timestamp}
      # Reviewed by: Security Team
      # Approved for: Production Use
```

## Pattern Management

### Enabling/Disabling Patterns
```yaml
patterns:
  - name: "active-pattern"
    paths: ["**/*.tf"]
    recertification_days: 90
    enabled: true

  - name: "disabled-pattern"
    paths: ["legacy/**/*.tf"]
    recertification_days: 90
    enabled: false
```

### Pattern Organization
```yaml
patterns:
  # Production Infrastructure
  - name: "prod-terraform"
    description: "Production Terraform"
    paths: ["terraform/prod/**/*.tf"]
    recertification_days: 90

  - name: "prod-kubernetes"
    description: "Production Kubernetes manifests"
    paths: ["k8s/prod/**/*.yaml"]
    recertification_days: 90

  # Development Infrastructure
  - name: "dev-terraform"
    description: "Development Terraform"
    paths: ["terraform/dev/**/*.tf"]
    recertification_days: 30

  - name: "dev-kubernetes"
    description: "Development Kubernetes manifests"
    paths: ["k8s/dev/**/*.yaml"]
    recertification_days: 30
```

## Advanced Patterns

### Multi-Environment Patterns
```yaml
patterns:
  - name: "multi-env-terraform"
    description: "Terraform across all environments"
    paths:
      - "terraform/*/**/*.tf"
    exclude:
      - "terraform/modules/**"
    recertification_days: 90

  - name: "environment-specific"
    description: "Environment-specific configurations"
    paths:
      - "config/prod/**/*.yaml"
      - "config/staging/**/*.yaml"
      - "config/dev/**/*.yaml"
    recertification_days: 60
```

### File Type Combinations
```yaml
patterns:
  - name: "infrastructure-code"
    description: "All infrastructure code"
    paths:
      - "**/*.tf"       # Terraform
      - "**/*.yaml"     # Kubernetes, CloudFormation
      - "**/*.yml"      # Kubernetes, CloudFormation
      - "**/*.json"     # CloudFormation, ARM templates
      - "**/*.py"       # Python infrastructure code
    exclude:
      - "**/test/**"
      - "**/tests/**"
    recertification_days: 180
```

## Pattern Validation

ICE validates patterns on startup:

```bash
# Validate configuration
ice config validate --config config.yaml

# Test pattern matching
ice run --config config.yaml --dry-run --verbose
```

Common pattern issues:
- Invalid glob syntax
- Missing required fields
- Conflicting include/exclude patterns
- Overlapping pattern names

## Performance Considerations

### Pattern Optimization
- Use specific paths instead of `**/*`
- Minimize exclusion patterns
- Group related files in same directories
- Avoid overly broad patterns

### Large Repositories
```yaml
# Efficient scanning
patterns:
  - name: "terraform-only"
    paths:
      - "terraform/**/*.tf"    # Specific directory
    recertification_days: 90

# Instead of
patterns:
  - name: "all-files"
    paths:
      - "**/*.tf"              # Scans entire repository
    recertification_days: 90
```

## Troubleshooting

### Files Not Matched
- Check glob pattern syntax
- Verify file paths match patterns
- Use `--verbose` to see matching process
- Test patterns with `ice run --dry-run`

### Unexpected Matches
- Review exclusion patterns
- Check for overlapping patterns
- Use more specific path patterns
- Validate with dry-run mode

### Performance Issues
- Reduce pattern complexity
- Use directory-specific patterns
- Limit broad `**` usage
- Monitor scan times with verbose logging

## Best Practices

1. **Use descriptive names** for patterns
2. **Group related files** logically
3. **Set appropriate recertification periods** based on risk
4. **Use decorators** for audit trails
5. **Test patterns** with dry-run mode
6. **Document complex patterns** with descriptions
7. **Regularly review** pattern effectiveness
8. **Monitor performance** of pattern matching

## Next Steps

- [PR Strategies](pr-strategies.md) - How files are grouped into pull requests
- [Assignment Strategies](assignment-strategies.md) - Reviewer assignment configuration
- [Configuration Overview](overview.md) - Complete configuration reference
