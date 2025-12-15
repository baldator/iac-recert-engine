# CI/CD Integration

ICE integrates seamlessly with continuous integration and deployment pipelines. This guide covers automated recertification workflows, scheduling, and integration with popular CI/CD platforms.

## Overview

CI/CD integration enables:

- **Automated scheduling** - Regular recertification runs
- **Multi-environment support** - Different configs per environment
- **Parallel processing** - Handle multiple repositories
- **Reporting and notifications** - Integration with monitoring systems
- **Approval workflows** - Integration with change management

## GitHub Actions

### Basic Workflow

Schedule weekly recertification:

```yaml
# .github/workflows/recertification.yml
name: Infrastructure Recertification

on:
  schedule:
    # Run weekly on Sunday at 2 AM UTC
    - cron: '0 2 * * 0'
  workflow_dispatch:
    inputs:
      dry_run:
        description: 'Run in dry-run mode'
        required: false
        default: 'false'
        type: boolean

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  recertify:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Set up ICE
      run: |
        curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
        chmod +x ice

    - name: Run recertification
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        if [ "${{ github.event.inputs.dry_run }}" = "true" ] || [ "${{ github.event_name }}" = "pull_request" ]; then
          echo "Running in dry-run mode"
          ./ice run --dry-run --config .ice/config.yaml
        else
          echo "Running production recertification"
          ./ice run --config .ice/config.yaml
        fi
```

### Matrix Strategy for Multiple Repositories

Process multiple repositories in parallel:

```yaml
# .github/workflows/multi-repo-recert.yml
name: Multi-Repository Recertification

on:
  schedule:
    - cron: '0 3 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  recertify:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        repo: [
          'org/infrastructure',
          'org/networking',
          'org/security'
        ]
    steps:
    - name: Checkout ICE config
      uses: actions/checkout@v4
      with:
        repository: 'myorg/ice-configs'
        path: 'configs'

    - name: Set up ICE
      run: |
        curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
        chmod +x ice

    - name: Run recertification
      env:
        GITHUB_TOKEN: ${{ secrets.ICE_GITHUB_TOKEN }}
      run: |
        ./ice run \
          --repo-url "https://github.com/${{ matrix.repo }}" \
          --config "configs/${{ matrix.repo }}/config.yaml"
```

### Docker-based Workflow

Use Docker for consistent execution:

```yaml
# .github/workflows/docker-recert.yml
name: Docker-based Recertification

on:
  schedule:
    - cron: '0 2 * * 0'
  workflow_dispatch:

jobs:
  recertify:
    runs-on: ubuntu-latest
    container:
      image: ghcr.io/baldator/iac-recert-engine:latest
    steps:
    - name: Checkout configuration
      uses: actions/checkout@v4

    - name: Run recertification
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        ice run --config .ice/config.yaml
```

### Pull Request Triggers

Run on pull request events for testing:

```yaml
# .github/workflows/pr-test.yml
name: Test Recertification on PR

on:
  pull_request:
    paths:
      - '.ice/**'
      - 'infrastructure/**'

jobs:
  test-recert:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Set up ICE
      run: |
        curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
        chmod +x ice

    - name: Test configuration
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        ./ice run --dry-run --verbose --config .ice/config.yaml
```

## GitLab CI/CD

### Basic Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - recertify

recertification:
  stage: recertify
  image: ghcr.io/baldator/iac-recert-engine:latest
  only:
    - schedules
  script:
    - ice run --config .ice/config.yaml
  dependencies: []
```

### Advanced GitLab Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - recertify
  - report

validate_config:
  stage: validate
  image: ghcr.io/baldator/iac-recert-engine:latest
  script:
    - ice run --dry-run --config .ice/config.yaml
  only:
    - merge_requests

recertify_production:
  stage: recertify
  image: ghcr.io/baldator/iac-recert-engine:latest
  environment:
    name: production
  script:
    - ice run --config .ice/config.yaml
  only:
    schedules:
      - weekly_recert
  dependencies: []

recertify_staging:
  stage: recertify
  image: ghcr.io/baldator/iac-recert-engine:latest
  environment:
    name: staging
  script:
    - ice run --repo-url $STAGING_REPO_URL --config .ice/staging.yaml
  only:
    schedules:
      - daily_recert
  dependencies: []

generate_report:
  stage: report
  image: alpine:latest
  script:
    - echo "Recertification completed"
    - cat recertification.log || true
  artifacts:
    paths:
      - recertification.log
    expire_in: 1 week
  only:
    schedules:
```

### GitLab Scheduled Pipelines

Set up scheduled recertification:

1. Go to **CI/CD → Schedules**
2. Create new schedule:
   - **Description**: Weekly Infrastructure Recertification
   - **Interval Pattern**: `0 2 * * 0` (Sunday 2 AM)
   - **Cron Timezone**: UTC
   - **Target Branch**: `main`
3. Set variables:
   - `GITHUB_TOKEN`: Your GitHub token
   - `STAGING_REPO_URL`: Staging repository URL

## Azure DevOps

### Classic Pipeline

```yaml
# azure-pipelines.yml
trigger: none

schedules:
- cron: "0 2 * * 0"
  displayName: Weekly recertification
  branches:
    include:
    - main

pool:
  vmImage: 'ubuntu-latest'

steps:
- script: |
    curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
    chmod +x ice
  displayName: 'Download ICE'

- script: |
    ./ice run --config .ice/config.yaml
  displayName: 'Run Recertification'
  env:
    GITHUB_TOKEN: $(GITHUB_TOKEN)
```

### YAML Pipeline with Docker

```yaml
# azure-pipelines.yml
trigger: none

schedules:
- cron: "0 2 * * 0"
  displayName: Weekly recertification
  branches:
    include:
    - main

pool:
  vmImage: 'ubuntu-latest'

container:
  image: ghcr.io/baldator/iac-recert-engine:latest

steps:
- script: |
    ice run --config .ice/config.yaml
  displayName: 'Run Recertification'
  env:
    GITHUB_TOKEN: $(GITHUB_TOKEN)
```

### Multi-Environment Pipeline

```yaml
# azure-pipelines.yml
parameters:
- name: environment
  displayName: Target Environment
  type: string
  default: production
  values:
  - development
  - staging
  - production

trigger: none

pool:
  vmImage: 'ubuntu-latest'

steps:
- script: |
    curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
    chmod +x ice
  displayName: 'Download ICE'

- script: |
    ./ice run --config .ice/${{ parameters.environment }}.yaml
  displayName: 'Run Recertification'
  env:
    GITHUB_TOKEN: $(GITHUB_TOKEN)
```

## Jenkins

### Freestyle Job

1. Create new **Freestyle project**
2. Configure build triggers:
   - **Build periodically**: `H 2 * * 0` (Weekly Sunday 2 AM)
3. Add build step **Execute shell**:

```bash
#!/bin/bash
set -e

# Download ICE
curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
chmod +x ice

# Run recertification
export GITHUB_TOKEN=$GITHUB_TOKEN
./ice run --config .ice/config.yaml
```

### Pipeline Job

```groovy
// Jenkinsfile
pipeline {
    agent any

    triggers {
        cron('H 2 * * 0')  // Weekly
    }

    environment {
        GITHUB_TOKEN = credentials('github-token')
    }

    stages {
        stage('Recertify') {
            steps {
                sh '''
                    # Download ICE
                    curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
                    chmod +x ice

                    # Run recertification
                    ./ice run --config .ice/config.yaml
                '''
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: '*.log', allowEmptyArchive: true
        }
        failure {
            mail to: 'team@example.com',
                 subject: "Recertification Failed",
                 body: "Recertification job failed. Check logs."
        }
    }
}
```

### Docker-based Jenkins

```groovy
// Jenkinsfile
pipeline {
    agent {
        docker {
            image 'ghcr.io/baldator/iac-recert-engine:latest'
        }
    }

    triggers {
        cron('H 2 * * 0')
    }

    environment {
        GITHUB_TOKEN = credentials('github-token')
    }

    stages {
        stage('Recertify') {
            steps {
                sh 'ice run --config .ice/config.yaml'
            }
        }
    }
}
```

## CircleCI

### Basic Configuration

```yaml
# .circleci/config.yml
version: 2.1

workflows:
  weekly-recert:
    triggers:
      - schedule:
          cron: "0 2 * * 0"  # Weekly Sunday 2 AM
          filters:
            branches:
              only: main
    jobs:
      - recertify

jobs:
  recertify:
    docker:
      - image: ghcr.io/baldator/iac-recert-engine:latest
    steps:
      - checkout
      - run:
          name: Run recertification
          command: ice run --config .ice/config.yaml
          environment:
            GITHUB_TOKEN: $GITHUB_TOKEN
```

### Advanced CircleCI with Workflows

```yaml
# .circleci/config.yml
version: 2.1

executors:
  ice-executor:
    docker:
      - image: ghcr.io/baldator/iac-recert-engine:latest

workflows:
  recertification:
    triggers:
      - schedule:
          cron: "0 2 * * 0"
          filters:
            branches:
              only: main
    jobs:
      - validate-config
      - recertify:
          requires:
            - validate-config

jobs:
  validate-config:
    executor: ice-executor
    steps:
      - checkout
      - run:
          name: Validate configuration
          command: ice run --dry-run --config .ice/config.yaml

  recertify:
    executor: ice-executor
    steps:
      - checkout
      - run:
          name: Run recertification
          command: ice run --config .ice/config.yaml
          environment:
            GITHUB_TOKEN: $GITHUB_TOKEN
```

## AWS CodeBuild

### Buildspec

```yaml
# buildspec.yml
version: 0.2

phases:
  install:
    commands:
      - curl -L https://github.com/baldator/iac-recert-engine/releases/latest/download/ice-linux-amd64 -o ice
      - chmod +x ice

  build:
    commands:
      - ./ice run --config .ice/config.yaml

environment:
  GITHUB_TOKEN:
    - GITHUB_TOKEN
```

### Scheduled Execution

Use CloudWatch Events to trigger CodeBuild:

```json
{
  "source": ["aws.events"],
  "detail-type": ["Scheduled Event"],
  "detail": {
    "schedule": "cron(0 2 ? * SUN *)"
  }
}
```

## Google Cloud Build

### Cloud Build Configuration

```yaml
# cloudbuild.yaml
steps:
  - name: 'ghcr.io/baldator/iac-recert-engine:latest'
    args:
      - 'run'
      - '--config'
      - '.ice/config.yaml'
    env:
      - 'GITHUB_TOKEN=$GITHUB_TOKEN'

timeout: '3600s'
```

### Scheduled Triggers

Create scheduled triggers in Cloud Build:

1. Go to **Cloud Build → Triggers**
2. Create new trigger
3. Set schedule: `0 2 * * 0` (Sunday 2 AM)
4. Configure source and build config

## Multi-Repository Strategies

### Monorepo Approach

Single repository with multiple configurations:

```yaml
# .ice/config.yaml
repositories:
  - name: "infra-prod"
    url: "https://github.com/org/infra"
    patterns:
      - name: "terraform-prod"
        paths: ["terraform/prod/**/*.tf"]
        recertification_days: 90

  - name: "infra-staging"
    url: "https://github.com/org/infra"
    patterns:
      - name: "terraform-staging"
        paths: ["terraform/staging/**/*.tf"]
        recertification_days: 30
```

### Separate Config Repository

Centralized configuration management:

```yaml
# Repository structure
# org/ice-configs/
# ├── repos/
# │   ├── infra-prod.yaml
# │   ├── infra-staging.yaml
# │   └── networking.yaml
# └── scripts/
#     └── run-all.sh
```

```bash
#!/bin/bash
# run-all.sh

repos=("infra-prod" "infra-staging" "networking")

for repo in "${repos[@]}"; do
    echo "Processing $repo..."
    ice run --config "repos/$repo.yaml"
    sleep 60  # Rate limiting
done
```

## Notifications and Reporting

### Slack Notifications

```yaml
# .github/workflows/recertification.yml
- name: Notify Slack
  if: always()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: "Infrastructure recertification ${{ job.status }}"
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Email Reports

```bash
#!/bin/bash
# Generate report
ice run --config config.yaml > recert.log 2>&1

# Send email
mail -s "Recertification Report" team@example.com < recert.log
```

### Dashboard Integration

Send metrics to monitoring systems:

```bash
#!/bin/bash
# Run recertification and capture metrics
ice run --config config.yaml 2>&1 | tee recert.log

# Extract metrics
prs_created=$(grep "Pull request created" recert.log | wc -l)
files_processed=$(grep "Found.*files requiring recertification" recert.log | awk '{print $2}')

# Send to monitoring
curl -X POST https://monitoring.example.com/metrics \
  -d "prs_created=$prs_created&files_processed=$files_processed"
```

## Security Considerations

### Secret Management

Use platform-specific secret management:

```yaml
# GitHub Actions
env:
  GITHUB_TOKEN: ${{ secrets.ICE_GITHUB_TOKEN }}

# GitLab CI
variables:
  GITHUB_TOKEN: $ICE_GITHUB_TOKEN

# Azure DevOps
env:
  GITHUB_TOKEN: $(ICE_GITHUB_TOKEN)
```

### Least Privilege Tokens

Create dedicated tokens with minimal permissions:

```bash
# GitHub token with minimal scopes
# - repo (read/write)
# - pull_requests (read/write)
```

### Audit Logging

Enable comprehensive audit logging:

```yaml
audit:
  enabled: true
  storage: "file"
  config:
    directory: "/var/log/ice/audit"
```

## Troubleshooting CI/CD

### Authentication Issues

```bash
# Test token in CI
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check token permissions
# Ensure token has repo and pull_request scopes
```

### Rate Limiting

```bash
# Add delays between runs
sleep 60

# Reduce concurrency
global:
  max_concurrent_prs: 2
```

### Resource Constraints

```yaml
# Increase runner resources
jobs:
  recertify:
    runs-on: ubuntu-latest
    # Increase timeout and resources
    timeout-minutes: 60
```

### Configuration Validation

```yaml
# Validate before running
- name: Validate config
  run: |
    ice run --dry-run --config .ice/config.yaml
```

## Best Practices

### Scheduling

- **Weekly runs**: Balance compliance with operational overhead
- **Off-peak hours**: Run during low-traffic periods
- **Timezone consideration**: Align with team working hours
- **Backup schedules**: Have manual trigger capability

### Environment Management

- **Separate configurations**: Different settings per environment
- **Gradual rollout**: Test in staging before production
- **Rollback plans**: Ability to revert changes quickly
- **Monitoring**: Track success/failure rates

### Error Handling

- **Retry logic**: Automatic retries for transient failures
- **Notifications**: Alert teams on failures
- **Logging**: Comprehensive logs for debugging
- **Timeouts**: Prevent hanging jobs

### Security

- **Token rotation**: Regular token renewal
- **Access control**: Limit who can trigger runs
- **Audit trails**: Log all activities
- **Compliance**: Meet regulatory requirements

## Next Steps

- [Command Line Interface](cli.md) - Direct binary usage
- [Docker Usage](docker.md) - Containerized deployment
- [Configuration Overview](../configuration/overview.md) - Configuration reference
