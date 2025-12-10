# iac-recert-engine

[![Build](https://github.com/baldator/iac-recert-engine/actions/workflows/docker-build.yml/badge.svg)](https://github.com/baldator/iac-recert-engine/actions/workflows/docker-build.yml)
[![Docker](https://img.shields.io/docker/pulls/baldator/iac-recert-engine)](https://hub.docker.com/r/baldator/iac-recert-engine)

<img src="./resources/iac-recert-engine-logo.png" alt="Project Logo" />


**Infrastructure-as-Code Recertification Engine**

A lightweight Go tool that scans your IaC (Infrastructure as Code) repositories, identifies resources that need recertification, and helps you enforce compliance policies.

## Features
- Detects outdated or missing certifications in Terraform, CloudFormation, and other IaC files.
- Generates actionable reports and can create PRs to remediate issues.
- Configurable via `config.yaml` or command‑line flags.
- Built with Go 1.24, easy to integrate into CI/CD pipelines.
- Docker image available on GitHub Container Registry (GHCR).

## Getting Started
```yaml
# example.config.yaml
scan_paths:
  - ./terraform
  - ./cloudformation

recertification_rules:
  - name: "Expired Recertification Tag"
    type: "tag_date_check"
    tag_name: "recertification_date"
    # Resources with a recertification_date tag older than this duration will be flagged.
    # Format: 24h, 7d, 30d, 1y
    max_age: "365d" # 1 year

  - name: "Missing Recertification Tag"
    type: "tag_presence_check"
    tag_name: "recertification_date"
    # Resources missing the 'recertification_date' tag will be flagged.
    required: true

output:
  format: "markdown" # or "json", "csv"
  file: "recertification_report.md"

# remediations:
#   create_pr: true
#   pr_template: "templates/pr_template.md"
#   pr_branch_prefix: "recertification-fix-"
```
## Configuration File Fields

The `config.yaml` file allows you to customize the behavior of the IaC Recertification Engine. Below is a detailed description of each field:

| Field| Type| Description| Example|
|---|---|---|---|
| `scan_paths`| `List[str]`| A list of paths to directories or files that should be scanned for IaC resources.| A list of directories to be scanned for IaC resources. Each path can be relative or absolute. | `scan_paths:`<br>`  - ./terraform`<br>`  - /home/user/cloudformation`
| `recertification_rules`| `List[dict]`| A list of recertification rules to apply to the scanned resources.| A list of recertification rules to apply to the scanned resources. Each rule has a name, type, tag name, max age, and required flag. | `recertification_rules:`<br>`  - name: "Expired Recertification Tag"`<br>`    type: "tag_date_check"`<br>`    tag_name: "recertification_date"`<br>`    max_age: "365d"`<br>`    required: true`
| `output`| `dict`| A dictionary of output options.| A dictionary of output options. | `output:`<br>`  format: "markdown"`<br>`  file: "recertification_report.md"`
| `remediations`| `dict`| A dictionary of remediation options.| A dictionary of remediation options. | `remediations:`<br>`  create_pr: true`<br>`  pr_template: "templates/pr_template.md"`<br>`  pr_branch_prefix: "recertification-fix-"`



### Prerequisites
- Go 1.24 or later
- Docker (for container builds)

### Build the binary
```bash
# Clone the repository
git clone https://github.com/baldator/iac-recert-engine.git
cd iac-recert-engine

# Build the executable
go build -o ice ./cmd/ice
```

### Run the tool
```bash
./ice run --config config.yaml
```

## Docker
A Docker image is built and pushed on every push to the `main` branch.
```bash
# Pull the image
docker pull ghcr.io/baldator/iac-recert-engine:latest

# Run the container
docker run --rm -v $(pwd):/app ghcr.io/baldator/iac-recert-engine:latest run --config /app/config.yaml
```

## GitHub Actions
The repository includes a workflow (`.github/workflows/docker-build.yml`) that:
- Builds the Docker image on pushes to `main`.
- Logs in to GHCR using the `GITHUB_TOKEN`.
- Pushes the image with tags `latest` and the commit SHA.

## Configuration
Edit `config.yaml` to define your recertification rules, target directories, and output formats. See `example.config.yaml` for a full example.

## Contributing
Contributions are welcome! Please:
1. Fork the repository.
2. Create a feature branch.
3. Write tests for new functionality.
4. Submit a pull request.

## License
This project is licensed under the MIT License.
