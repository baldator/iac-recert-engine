# 🤝 Contributing

We welcome contributions to the IaC Recertification Engine! This document outlines the process for contributing to the project.

## Ways to Contribute

### Code Contributions
- **Bug Fixes**: Identify and fix issues in the codebase
- **New Features**: Implement new functionality following our design principles
- **Performance Improvements**: Optimize existing code for better performance
- **Documentation**: Improve or add documentation

### Non-Code Contributions
- **Bug Reports**: Report issues with detailed reproduction steps
- **Feature Requests**: Suggest new features or improvements
- **Documentation**: Help improve or translate documentation
- **Testing**: Add test cases or help with testing

## Development Workflow

### 1. Fork and Clone
```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/your-username/iac-recert-engine.git
cd iac-recert-engine
```

### 2. Set Up Development Environment
Follow the [Development Setup](setup.md) guide to configure your local environment.

### 3. Create a Branch
```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Or create a bug fix branch
git checkout -b fix/issue-number-description
```

### 4. Make Changes
- Write clear, concise commit messages
- Follow the existing code style and conventions
- Add tests for new functionality
- Update documentation as needed

### 5. Test Your Changes
```bash
# Run tests
make test

# Run linting
make lint

# Build the project
make build
```

### 6. Submit a Pull Request
```bash
# Push your branch
git push origin feature/your-feature-name

# Create a Pull Request on GitHub
```

## Pull Request Guidelines

### Before Submitting
- [ ] Tests pass locally
- [ ] Code is properly formatted and linted
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] Branch is up to date with main

### PR Description
Include:
- **What**: What changes are being made
- **Why**: Why these changes are needed
- **How**: How the changes work
- **Testing**: How to test the changes

### Example PR Description
```
## What
Adds support for Azure DevOps repositories

## Why
Customers using Azure DevOps need ICE to work with their repositories

## How
- Added Azure DevOps provider implementation
- Updated configuration schema
- Added tests for Azure DevOps integration

## Testing
Run `go test ./internal/provider/azure_test.go`
```

## Code Standards

### Go Code
- Follow standard Go formatting (`gofmt`)
- Use `golangci-lint` for code quality checks
- Write comprehensive tests (unit and integration)
- Use meaningful variable and function names
- Add comments for complex logic

### Commit Messages
Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

Examples:
```
feat: add Azure DevOps provider support
fix: resolve panic in git history scanning
docs: update installation instructions
```

## Testing Guidelines

### Unit Tests
- Test individual functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for high code coverage

### Integration Tests
- Test component interactions
- Use realistic test data
- Test error conditions
- Validate end-to-end workflows

### Running Tests
```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/scan/...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

## Documentation

### Code Documentation
- Add package comments for exported functions
- Document complex algorithms
- Include examples in doc comments

### User Documentation
- Update relevant docs for new features
- Add configuration examples
- Update troubleshooting guides

## Issue Reporting

### Bug Reports
When reporting bugs, include:
- **Version**: ICE version and Go version
- **Environment**: OS, architecture
- **Steps to Reproduce**: Detailed reproduction steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Logs**: Relevant log output
- **Configuration**: Sanitized config file

### Feature Requests
When requesting features, include:
- **Use Case**: Why this feature is needed
- **Requirements**: Detailed requirements
- **Alternatives**: Considered alternatives
- **Impact**: Expected impact on users

## Community Guidelines

### Communication
- Be respectful and constructive
- Use clear, professional language
- Provide context for questions
- Help others when possible

### Review Process
- All PRs require review before merging
- Reviews focus on code quality, correctness, and maintainability
- Be open to feedback and willing to make changes
- Multiple reviewers may be involved for complex changes

### Recognition
Contributors are recognized through:
- GitHub contributor statistics
- Mention in release notes
- Attribution in documentation

## Getting Help

- **Documentation**: Check the [docs](../) first
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Join community channels for support

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (Apache 2.0).

Thank you for contributing to IaC Recertification Engine! 🚀
