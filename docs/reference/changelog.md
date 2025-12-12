# Changelog

All notable changes to IaC Recertification Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial public release of IaC Recertification Engine
- Support for GitHub, Azure DevOps, and GitLab repositories
- YAML-based configuration system
- Multiple PR grouping strategies (per-file, per-pattern, per-committer, single PR)
- Flexible assignment strategies (static, last-committer, plugin-based, composite)
- Plugin architecture for extensibility
- Docker container support
- Comprehensive audit logging
- Dry-run mode for testing configurations

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

## [1.0.0] - 2025-01-15

### Added
- Core recertification engine with git history analysis
- File pattern matching with glob support and exclusions
- Priority-based recertification scheduling (Critical, High, Medium, Low)
- GitHub provider implementation with full PR support
- Azure DevOps provider implementation
- GitLab provider implementation
- Configuration validation with detailed error messages
- Structured logging with configurable verbosity
- Environment variable substitution in configuration
- File decorator support for marking recertified files
- Concurrent PR creation with configurable limits
- Comprehensive CLI with multiple commands

### Technical Features
- Go 1.24+ compatibility
- Stateless architecture for reliable execution
- Provider abstraction layer for Git platform agnosticism
- Plugin system with multiple extension points
- Comprehensive test coverage
- Docker multi-stage build support
- Kubernetes CronJob manifests
- GitHub Actions CI/CD integration examples

---

## Version History

### Pre-1.0.0 Development

The project evolved through several internal versions before the 1.0.0 release:

- **0.1.0-alpha**: Basic file scanning and git history analysis
- **0.2.0-alpha**: Initial PR creation support for GitHub
- **0.3.0-alpha**: Configuration system and validation
- **0.4.0-alpha**: Multiple provider support (Azure DevOps, GitLab)
- **0.5.0-alpha**: Plugin architecture and assignment strategies
- **0.6.0-alpha**: Docker containerization and CI/CD integration
- **0.7.0-alpha**: Audit logging and observability features
- **0.8.0-alpha**: Performance optimizations and concurrent execution
- **0.9.0-alpha**: Production hardening and comprehensive testing

## Migration Guide

### From 0.x to 1.0.0

If you're migrating from pre-1.0.0 versions:

1. **Configuration Schema Changes**
   - The `version` field is now required and must be set to `"1.0"`
   - Provider configuration moved from global to repository-specific
   - Plugin configuration schema standardized

2. **Breaking Changes**
   - Environment variable names for tokens changed (e.g., `GITHUB_TOKEN` instead of `GIT_TOKEN`)
   - Default branch configuration moved to `global.default_base_branch`
   - Audit configuration requires explicit storage type

3. **New Features to Configure**
   - Set up audit logging for compliance requirements
   - Configure dry-run mode for testing
   - Review and set appropriate concurrency limits

### Example Migration

**Before (0.9.0-alpha):**
```yaml
git_provider: "github"
token_env: "GIT_TOKEN"
patterns:
  - name: "terraform"
    paths: ["**/*.tf"]
    days: 90
```

**After (1.0.0):**
```yaml
version: "1.0"
repository:
  url: "https://github.com/your-org/repo"
  provider: "github"
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"
patterns:
  - name: "terraform"
    paths: ["**/*.tf"]
    recertification_days: 90
audit:
  enabled: true
  storage: "file"
  config:
    path: "audit.log"
```

## Future Releases

### Planned for 1.1.0
- Enhanced plugin ecosystem
- Webhook support for real-time triggers
- Advanced reporting and dashboards
- Integration with external CMDB systems

### Planned for 1.2.0
- Multi-repository batch operations
- Advanced scheduling with calendar integration
- Machine learning-based priority scoring
- Integration with security scanning tools

### Long-term Vision (2.0.0)
- GraphQL API for programmatic access
- Event-driven architecture
- Advanced analytics and insights
- Enterprise federation support

## Contributing to Changelog

When contributing changes that affect users:

1. Add entries to the "Unreleased" section above
2. Categorize changes as Added, Changed, Deprecated, Removed, Fixed, or Security
3. Use present tense for changes ("Add feature" not "Added feature")
4. Reference issue/PR numbers where applicable
5. Group similar changes together

### Example Entry
```
### Added
- Add support for custom PR templates (#123)
- Implement Azure DevOps provider integration (#124)

### Fixed
- Resolve memory leak in file scanning (#125)
```

## Release Process

1. Update version in relevant files
2. Move unreleased changes to new version section
3. Update release date
4. Create git tag
5. Build and publish artifacts
6. Update documentation
7. Announce release

## Support

For questions about specific versions or migration help, please:

- Check the [Troubleshooting](troubleshooting/common-issues.md) guide
- Review the [Configuration](configuration/overview.md) documentation
- Open an issue on GitHub
- Join community discussions
