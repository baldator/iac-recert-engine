# 🤝 Contributing

## 🎯 Our Philosophy: Feedback is Priority #1

At IaC Recertification Engine (ICE), **your feedback is our highest priority**. We believe that the best software is built by deeply understanding and responding to user needs. Every piece of feedback—whether it's a bug report, feature suggestion, usage question, or general comment—is treated as a valuable opportunity to improve ICE.

### Why Feedback Matters Most
- **User-Centric Development**: ICE exists to solve real problems for real users
- **Continuous Learning**: Your experiences teach us how to make ICE better
- **Community-Driven**: The project evolves based on community input and needs
- **Quality over Speed**: We'd rather get it right than get it fast

### How We Handle Feedback
- **Immediate Acknowledgment**: All feedback receives a response
- **Priority Triage**: Critical user issues are addressed before feature development
- **Transparent Communication**: You'll know the status and timeline for addressing your input
- **Inclusive Process**: Feedback from all experience levels is equally valued

## Ways to Contribute

### 💬 Feedback & Community Input (Highest Priority)
Your voice shapes the future of ICE. We prioritize feedback above all other contributions because it ensures we're building what users actually need.

#### Types of Feedback We Love
- **Bug Reports**: Help us identify and fix issues
- **Feature Suggestions**: Tell us what would make ICE more useful
- **Usage Questions**: Help us understand how people use ICE
- **Performance Feedback**: Share your experiences with speed and reliability
- **Documentation Issues**: Point out confusing or missing information
- **Configuration Challenges**: Help us improve setup and configuration
- **Integration Questions**: Share your experiences with CI/CD, plugins, etc.

#### How to Provide Feedback
1. **Use Issue Templates**: We've created specific templates for different types of feedback
2. **Be Specific**: Include context, examples, and your environment details
3. **Share Your Use Case**: Help us understand why this matters to you
4. **Follow Up**: Let us know if our responses address your concerns

### Code Contributions
While feedback is our top priority, we also welcome code contributions that address community-identified needs.

- **Bug Fixes**: Address issues reported by the community
- **Community-Requested Features**: Implement features suggested by users
- **Performance Improvements**: Optimize based on user feedback
- **Documentation**: Improve clarity based on user questions

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

## Feedback Guidelines

### Making Your Feedback Count
To help us address your feedback effectively, please:

#### For Bug Reports
- **Use the Bug Report Template**: It guides you to provide all necessary information
- **Include Reproduction Steps**: Step-by-step instructions help us reproduce and fix issues
- **Share Your Environment**: ICE version, OS, Go version, and Git provider details
- **Provide Logs**: Error messages and relevant output help with debugging
- **Attach Configuration**: Sanitized config files show us your setup

#### For Feature Requests
- **Explain Your Use Case**: Help us understand why this feature matters to you
- **Describe the Problem**: What problem would this feature solve?
- **Share Alternatives**: What workarounds are you currently using?
- **Consider Impact**: How would this benefit other users?

#### For General Feedback
- **Be Specific**: Point to specific areas of improvement
- **Include Context**: Share your experience level and use case
- **Suggest Solutions**: If you have ideas for improvement, share them
- **Ask Questions**: We're here to help you succeed with ICE

## Community Guidelines

### Communication
- **Feedback is Always Welcome**: Share your thoughts openly - we value all input
- Be respectful and constructive in all interactions
- Use clear, professional language
- Provide context for questions and feedback
- Help others when possible, especially newcomers

### Review Process
- **Feedback-First Triage**: Community feedback is reviewed before code contributions
- All PRs require review before merging, but user needs come first
- Reviews focus on code quality, correctness, and user impact
- Be open to feedback and willing to iterate on solutions
- Multiple reviewers may be involved for complex changes

### Recognition
Contributors are recognized through:
- GitHub contributor statistics
- Mention in release notes for significant contributions
- Attribution in documentation
- **Special recognition for valuable feedback** that shapes the project direction

## Getting Help & Providing Feedback

### Primary Channels for Feedback
1. **GitHub Issues**: Use our issue templates for structured feedback
2. **Documentation**: Check the [docs](../) first, then let us know if something's unclear
3. **Discussions**: Use GitHub Discussions for questions and general discussion
4. **Community**: Join community channels for support and feedback

### When to Reach Out
- **Have a question?** Ask it - we want to understand how you use ICE
- **Found a bug?** Report it - detailed bug reports help us improve
- **Have an idea?** Share it - feature requests shape our roadmap
- **Need help?** We're here - configuration and usage questions are valuable feedback
- **Confused by docs?** Tell us - documentation feedback is crucial

### Our Commitment to You
- **No question is too small** - if you're confused, others might be too
- **All feedback is actionable** - we review and respond to every piece of input
- **Your success matters** - we're invested in helping you succeed with ICE

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (Apache 2.0).

Thank you for contributing to IaC Recertification Engine! 🚀
