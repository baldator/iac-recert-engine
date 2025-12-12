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

## Local Development

This documentation uses [Docsify](https://docsify.js.org/) for a lightweight, fast experience. To work on the documentation locally:

### Quick Start (Recommended)

```bash
# Navigate to docs directory
cd docs

# Install Docsify CLI globally (optional)
npm install -g docsify-cli

# Serve documentation locally
docsify serve .

# Or use npm scripts
npm run serve
```

### Available Commands

- `npm run serve` - Serve documentation locally (default port 3000)
- `npm run dev` - Serve on port 3000 explicitly
- `npm run preview` - Serve and open in browser automatically

### Manual Setup

```bash
# Install Docsify CLI globally
npm install -g docsify-cli

# Navigate to docs directory
cd docs

# Serve the documentation locally
docsify serve .

# The documentation will be available at http://localhost:3000
```

## GitHub Pages Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch. The deployment workflow:

1. **Triggers**: On push to `main` branch affecting `docs/` directory
2. **Process**: Serves Markdown files directly using Docsify
3. **Deployment**: Publishes to `https://baldator.github.io/iac-recert-engine/`

### Setup Requirements

Before the automatic deployment works, you need to enable GitHub Pages in your repository:

1. Go to your repository **Settings** tab
2. Scroll down to **Pages** section
3. Under **Source**, select **"GitHub Actions"**
4. The workflow will automatically deploy to the `gh-pages` branch

### Manual Deployment

You can also trigger the documentation deployment manually:

1. Go to the **Actions** tab in your GitHub repository
2. Select **"Deploy Documentation"** workflow
3. Click **"Run workflow"**

### Local Testing

Before pushing changes, test the documentation locally:

```bash
cd docs
docsify serve .
# Visit http://localhost:3000 to preview
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
