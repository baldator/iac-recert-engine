# IaC Recertification Engine Documentation

This documentation provides comprehensive guidance for using and contributing to the IaC Recertification Engine (ICE).

## About ICE

The IaC Recertification Engine is a lightweight Go tool that automates the recertification of Infrastructure as Code (IaC) repositories. It scans your IaC files, identifies resources that need periodic review based on configurable policies, and creates pull requests to trigger review workflows on GitHub, Azure DevOps, and GitLab.

## Key Features

- **Automated Recertification**: Enforce periodic review of IaC configurations
- **Multi-Platform Support**: Works with GitHub, Azure DevOps, and GitLab
- **Flexible Configuration**: YAML-based configuration with extensive customization options
- **Plugin Architecture**: Extensible through plugins for custom logic
- **CI/CD Integration**: Easy integration into existing pipelines
- **Audit Trail**: Comprehensive logging and audit capabilities

## Documentation Structure

This documentation is organized into the following sections:

- **Getting Started**: Introduction, quick start guide, and installation instructions
- **Configuration**: Detailed configuration options and examples
- **Usage**: How to use ICE in different scenarios
- **Architecture**: System design and component descriptions
- **API Reference**: Configuration schema and plugin interfaces
- **Development**: Contributing guidelines and development setup
- **Troubleshooting**: Common issues and debugging tips
- **Reference**: Changelog, license, and FAQ

## Building the Documentation

This documentation is designed to be published as a GitBook. To build and serve locally:

### Quick Start (Recommended)

```bash
# Navigate to docs directory
cd docs

# Install dependencies and serve locally
npm run install-plugins
npm run serve
```

### Manual Setup

```bash
# Install GitBook CLI globally (if not already installed)
npm install -g gitbook-cli

# Navigate to docs directory
cd docs

# Install GitBook plugins
gitbook install

# Serve the documentation locally (with live reload)
gitbook serve

# Build static files for production
gitbook build
```

### Available Commands

- `npm run serve` - Serve documentation locally with live reload
- `npm run build` - Build static HTML files
- `npm run install-plugins` - Install GitBook plugins
- `npm run pdf` - Generate PDF version
- `npm run epub` - Generate EPUB version
- `npm run mobi` - Generate MOBI version

## GitHub Pages Deployment

The documentation is automatically built and deployed to GitHub Pages when changes are pushed to the `main` branch. The deployment workflow:

1. **Triggers**: On push to `main` branch affecting `docs/` directory
2. **Build Process**: Uses GitBook CLI to generate static HTML
3. **Deployment**: Publishes to `https://baldator.github.io/iac-recert-engine/`

### Setup Requirements

Before the automatic deployment works, you need to enable GitHub Pages in your repository:

1. Go to your repository **Settings** tab
2. Scroll down to **Pages** section
3. Under **Source**, select **"GitHub Actions"**
4. The workflow will automatically deploy to the `gh-pages` branch

### Manual Deployment

You can also trigger the documentation build manually:

1. Go to the **Actions** tab in your GitHub repository
2. Select **"Build and Deploy Documentation"** workflow
3. Click **"Run workflow"**

### Local Testing

Before pushing changes, test the build locally:

```bash
cd docs
gitbook build
# Check the _book/ directory for generated files
```

## Contributing to Documentation

We welcome contributions to improve this documentation. Please follow these guidelines:

1. Use clear, concise language
2. Include code examples where helpful
3. Keep examples up-to-date with the latest version
4. Test any commands or configurations you document
5. Follow the existing structure and formatting

## Support

If you need help or have questions:

- Check the [Troubleshooting](troubleshooting/common-issues.md) section
- Review the [FAQ](reference/faq.md)
- Open an issue on GitHub
- Join our community discussions

## License

This documentation is licensed under the MIT License, in line with the main project.
