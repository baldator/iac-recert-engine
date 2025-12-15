# iac-recert-engine

[![Build](https://github.com/baldator/iac-recert-engine/actions/workflows/docker-build.yml/badge.svg)](https://github.com/baldator/iac-recert-engine/actions/workflows/docker-build.yml)

<img src="./resources/iac-recert-engine-logo.png" alt="Project Logo" />


**Infrastructure-as-Code Recertification Engine**

A lightweight Go tool that scans your IaC (Infrastructure as Code) repositories, identifies resources that need recertification, and helps you enforce compliance policies.

## The Problem

Infrastructure as Code has revolutionized how organizations manage their cloud infrastructure. However, as IaC configurations age, they can become outdated, insecure, or misaligned with current business requirements. Traditional approaches rely on manual reviews or ad-hoc processes that are:

- **Inefficient**: Manual tracking of when configurations were last reviewed
- **Inconsistent**: Variable review quality and timing
- **Hard to Audit**: Difficult to prove regular compliance and governance
- **Not Scalable**: Challenging to manage across large, multi-repository environments

ICE addresses these challenges by providing automated, policy-driven recertification workflows.

## Features
- Detects outdated or missing certifications in Terraform, CloudFormation, and other IaC files.
- Generates actionable reports and can create PRs to remediate issues.
- Configurable via `config.yaml` or command‑line flags.
- Built with Go 1.24, easy to integrate into CI/CD pipelines.
- Docker image available on GitHub Container Registry (GHCR).

## Demo

Here's a quick demonstration of how ICE works:

### 1. Create a basic configuration
```yaml
version: "1.0"
repository:
  url: "https://github.com/your-org/your-repo"
  provider: "github"
auth:
  provider: "github"
  token_env: "GITHUB_TOKEN"
patterns:
  - name: "terraform"
    paths: ["**/*.tf"]
    recertification_days: 90
assignment:
  strategy: "static"
  fallback_assignees: ["devops-team"]
```

### 2. Set your authentication token
```bash
export GITHUB_TOKEN="your_personal_access_token_here"
```

### 3. Run a dry-run to preview changes
```bash
ice run --dry-run --config config.yaml
```

### 4. Review the output
```
2025-12-14T10:30:00Z INFO Starting recertification run (DRY RUN)
2025-12-14T10:30:01Z INFO Scanning repository for files...
2025-12-14T10:30:05Z INFO Found 25 files requiring recertification
2025-12-14T10:30:06Z INFO Would create 3 file groups
2025-12-14T10:30:07Z INFO Would create PR for terraform-prod (15 files)
2025-12-14T10:30:08Z INFO Would create PR for k8s-manifests (8 files)
2025-12-14T10:30:09Z INFO Would create PR for policies (2 files)
2025-12-14T10:30:10Z INFO Recertification run completed successfully (DRY RUN)
```

### 5. Execute the actual recertification
```bash
ice run --config config.yaml
```

ICE will automatically create pull requests with clear titles, file lists, and certification checklists for team review.

## Use Cases

### Enterprise Governance
Large organizations can use ICE to enforce organization-wide IaC governance policies, ensuring all infrastructure code receives regular review regardless of team or project.

### Security Compliance
Security teams can implement automated recertification schedules that align with compliance frameworks like SOC 2, PCI DSS, or ISO 27001.

### DevSecOps Integration
Integrate ICE into existing CI/CD pipelines to automatically trigger recertification reviews as part of the development workflow.

### Multi-Cloud Management
Manage recertification across heterogeneous infrastructure environments, ensuring consistent governance across AWS, Azure, GCP, and on-premises resources.

### Practical Examples

**Financial Services Company**
- **Challenge**: Required quarterly review of all infrastructure configurations for SOX compliance
- **Solution**: ICE configured to scan 50+ repositories weekly, creating PRs for files older than 90 days
- **Result**: 80% reduction in manual compliance work, provable audit trails

**Cloud-Native Startup**
- **Challenge**: Rapid infrastructure changes needed regular security reviews
- **Solution**: Integrated ICE into GitHub Actions, running daily scans with automated PR creation
- **Result**: Consistent security posture across all environments, early detection of misconfigurations

**Enterprise Platform Team**
- **Challenge**: Managing IaC governance across 200+ development teams
- **Solution**: Organization-wide ICE deployment with custom plugins for team-specific assignment logic
- **Result**: Standardized review processes, improved collaboration between platform and development teams

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

### 🎯 Feedback is Our Highest Priority
At ICE, **your feedback drives everything we do**. We believe the best software is built by deeply understanding user needs. Every bug report, feature suggestion, usage question, or general comment is treated as a valuable opportunity to improve ICE.

**All feedback receives a response.** We prioritize user issues before feature development and value input from all experience levels.

### Ways to Contribute
- **💬 Feedback & Community Input** (Highest Priority): Bug reports, feature suggestions, usage questions, performance feedback
- **🐛 Bug Reports**: Help us identify and fix issues
- **✨ Feature Requests**: Suggest improvements and new capabilities
- **📖 Documentation**: Improve clarity based on user questions
- **🔧 Code Contributions**: Address community-identified needs

### Getting Started
Ready to contribute? Here's how:

1. **📋 Provide Feedback**: Use our [issue templates](.github/ISSUE_TEMPLATES/) for structured feedback
2. **📖 Check Documentation**: Visit our [comprehensive docs](docs/) for detailed contribution guidelines
3. **🛠️ Code Contributions**: See our [Contributing Guide](docs/development/contributing.md) for technical details

### Quick Links
- 📋 [Bug Report Template](.github/ISSUE_TEMPLATES/bug-report.md)
- ✨ [Feature Request Template](.github/ISSUE_TEMPLATES/feature-request.md)
- 💬 [Feedback Template](.github/ISSUE_TEMPLATES/feedback.md)
- 📖 [Full Contributing Guide](docs/development/contributing.md)

## License
This project is licensed under the MIT License.
