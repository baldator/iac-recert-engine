# PR Strategies

PR strategies control how files requiring recertification are grouped into pull requests. Different strategies optimize for different workflows, team sizes, and review processes.

## Configuration Schema

```yaml
pr_strategy:
  type: string              # Required: per_file, per_pattern, per_committer, single_pr, plugin
  max_files_per_pr: int     # Optional: Maximum files per PR
  plugin_name: string       # Optional: Plugin name for plugin strategy
```

## Available Strategies

### Per File Strategy (`per_file`)

Creates one pull request per file that needs recertification.

**Use Cases**:
- Small teams with fast review cycles
- Files that can be reviewed independently
- When changes are isolated and low-risk

**Configuration**:
```yaml
pr_strategy:
  type: "per_file"
  max_files_per_pr: 1  # Always 1 file per PR
```

**Example Output**:
- PR 1: `Recertify: terraform/prod/main.tf`
- PR 2: `Recertify: terraform/prod/variables.tf`
- PR 3: `Recertify: k8s/prod/deployment.yaml`

**Pros**:
- Isolated changes, easy to review
- Minimal merge conflicts
- Clear audit trail per file

**Cons**:
- Many PRs for bulk recertification
- Higher administrative overhead
- May overwhelm reviewers

### Per Pattern Strategy (`per_pattern`)

Groups all files from the same pattern into a single pull request.

**Use Cases**:
- Related files should be reviewed together
- Pattern-based ownership (e.g., all Terraform files)
- Medium-sized teams with pattern expertise

**Configuration**:
```yaml
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 50  # Optional limit
```

**Example Output**:
- PR 1: `Recertify: terraform-prod` (contains all terraform/prod/*.tf files)
- PR 2: `Recertify: k8s-manifests` (contains all k8s/*.yaml files)

**Pros**:
- Logical grouping of related files
- Fewer PRs than per-file strategy
- Pattern owners can review efficiently

**Cons**:
- Large PRs may be overwhelming
- Mixed file types if patterns are broad
- Potential for unrelated changes in same PR

### Per Committer Strategy (`per_committer`)

Groups files by their last committer, creating one PR per person.

**Use Cases**:
- Teams want individuals to review their own changes
- Code ownership by last modifier
- Accountability for recertification

**Configuration**:
```yaml
pr_strategy:
  type: "per_committer"
  max_files_per_pr: 25
```

**Example Output**:
- PR 1: `Recertify: john.doe` (files last modified by John)
- PR 2: `Recertify: jane.smith` (files last modified by Jane)

**Pros**:
- Individual accountability
- Subject matter experts review their code
- Natural ownership boundaries

**Cons**:
- Uneven PR sizes
- May not align with team structure
- Requires accurate git history

### Single PR Strategy (`single_pr`)

Combines all files requiring recertification into one large pull request.

**Use Cases**:
- Small repositories with infrequent changes
- Bulk recertification campaigns
- When all changes can be reviewed together

**Configuration**:
```yaml
pr_strategy:
  type: "single_pr"
  max_files_per_pr: 100  # Optional, but ignored for single_pr
```

**Example Output**:
- PR 1: `Recertify: All Files` (contains all files needing recertification)

**Pros**:
- Single review process
- Easy to track overall progress
- Minimal PR management overhead

**Cons**:
- Very large PRs difficult to review
- High risk of merge conflicts
- All-or-nothing approval process

### Plugin Strategy (`plugin`)

Uses a custom plugin to determine PR grouping logic.

**Use Cases**:
- Complex business rules for grouping
- Integration with external systems
- Custom organizational requirements

**Configuration**:
```yaml
pr_strategy:
  type: "plugin"
  plugin_name: "custom-grouping"
  max_files_per_pr: 30
```

**Pros**:
- Flexible grouping logic
- Can integrate with CMDB, ticketing systems
- Adaptable to organizational needs

**Cons**:
- Requires plugin development
- Additional complexity
- Plugin maintenance overhead

## File Limits

### Max Files Per PR
Control PR size to maintain reviewability:

```yaml
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 25  # Split large groups into multiple PRs
```

**Behavior**:
- When a group exceeds `max_files_per_pr`, it's split into multiple PRs
- Split PRs are numbered (e.g., `Recertify: terraform-prod (1/3)`)
- Maintains grouping strategy while controlling size

### Automatic Splitting
```yaml
# Large pattern split into multiple PRs
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 20

# Result: If pattern has 65 files
# PR 1: Recertify: terraform-prod (1/4) - 20 files
# PR 2: Recertify: terraform-prod (2/4) - 20 files
# PR 3: Recertify: terraform-prod (3/4) - 20 files
# PR 4: Recertify: terraform-prod (4/4) - 5 files
```

## Strategy Selection Guide

### Choose Based on Team Size

**Small Teams (1-5 people)**:
```yaml
pr_strategy:
  type: "single_pr"  # One PR for everything
```

**Medium Teams (6-20 people)**:
```yaml
pr_strategy:
  type: "per_pattern"  # Group by technology/domain
```

**Large Teams (20+ people)**:
```yaml
pr_strategy:
  type: "per_committer"  # Individual accountability
```

### Choose Based on Repository Size

**Small Repositories (< 100 files)**:
```yaml
pr_strategy:
  type: "single_pr"
```

**Medium Repositories (100-1000 files)**:
```yaml
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 50
```

**Large Repositories (1000+ files)**:
```yaml
pr_strategy:
  type: "per_file"
```

### Choose Based on Change Frequency

**Frequent Small Changes**:
```yaml
pr_strategy:
  type: "per_file"
```

**Periodic Bulk Updates**:
```yaml
pr_strategy:
  type: "per_pattern"
  max_files_per_pr: 25
```

**Annual Recertification**:
```yaml
pr_strategy:
  type: "single_pr"
```

## Advanced Configuration

### Multiple Strategies by Pattern
While ICE doesn't directly support different strategies per pattern, you can achieve similar results with plugins or by using multiple configuration files.

### Conditional Strategies
For complex scenarios, consider plugin-based strategies that can make decisions based on:
- File count
- File types
- Directory structure
- External system state

## PR Templates

Strategies work with PR templates to customize the content:

```yaml
pr_template:
  title: "🔄 Recertification: {pattern_name}"
  include_file_list: true
  include_checklist: true
```

## Performance Considerations

### Strategy Performance
- `per_file`: Fastest, minimal grouping overhead
- `per_pattern`: Moderate, requires pattern matching
- `per_committer`: Slower, requires git history analysis
- `single_pr`: Fastest grouping
- `plugin`: Depends on plugin implementation

### Memory Usage
- Large `single_pr` groups consume more memory
- `per_file` with many files creates many small objects
- Consider `max_files_per_pr` to control memory usage

## Troubleshooting

### Unexpected Grouping
- Verify git history is accessible for `per_committer`
- Check pattern definitions for `per_pattern`
- Use `--verbose` to see grouping decisions

### Too Many PRs
- Increase `max_files_per_pr` limit
- Switch to less granular strategy
- Use `single_pr` for bulk operations

### Too Few PRs
- Decrease `max_files_per_pr` limit
- Switch to more granular strategy
- Use `per_file` for maximum isolation

### Plugin Issues
- Verify plugin is loaded and enabled
- Check plugin logs for errors
- Test plugin independently

## Best Practices

1. **Match strategy to team workflow** - Choose what works for your review process
2. **Start with per_pattern** - Good default for most organizations
3. **Set reasonable file limits** - Keep PRs reviewable (10-50 files)
4. **Test strategies** - Use dry-run to see PR grouping before production
5. **Monitor PR sizes** - Adjust limits based on actual review times
6. **Document your choice** - Explain why you chose a particular strategy
7. **Review periodically** - Team size and processes change over time

## Strategy Comparison

| Strategy | PR Count | Review Complexity | Ownership | Setup |
|----------|----------|-------------------|-----------|-------|
| per_file | High | Low | Individual | Simple |
| per_pattern | Medium | Medium | Team | Simple |
| per_committer | Medium | Medium | Individual | Requires git |
| single_pr | Low | High | Team | Simple |
| plugin | Variable | Variable | Custom | Complex |

## Next Steps

- [Assignment Strategies](assignment-strategies.md) - How reviewers are assigned to PRs
- [PR Templates](../configuration/overview.md#pr-template-configuration) - Customizing PR content
- [Plugins](plugins.md) - Custom strategy implementation
