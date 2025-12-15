---
name: Bug Report
description: Report a bug or unexpected behavior
title: "[BUG] "
labels: ["bug", "triage"]
assignees: []
body:
  - type: markdown
    attributes:
      value: |
        ## Bug Report

        Thank you for reporting a bug! Please provide as much detail as possible to help us reproduce and fix the issue.

  - type: textarea
    id: description
    attributes:
      label: Describe the Bug
      description: A clear and concise description of what the bug is.
      placeholder: "What happened? What did you expect to happen?"
    validations:
      required: true

  - type: textarea
    id: steps
    attributes:
      label: Steps to Reproduce
      description: Provide step-by-step instructions to reproduce the issue.
      placeholder: |
        1. Go to '...'
        2. Click on '....'
        3. Scroll down to '....'
        4. See error
    validations:
      required: true

  - type: textarea
    id: expected-behavior
    attributes:
      label: Expected Behavior
      description: What should have happened?
      placeholder: "Describe what you expected to happen instead."
    validations:
      required: true

  - type: textarea
    id: actual-behavior
    attributes:
      label: Actual Behavior
      description: What actually happened?
      placeholder: "Describe what actually happened."
    validations:
      required: true

  - type: textarea
    id: environment
    attributes:
      label: Environment
      description: Please provide details about your environment.
      value: |
        - OS: [e.g., Windows 11, macOS 12.1, Ubuntu 20.04]
        - ICE Version: [e.g., v1.0.0]
        - Go Version: [e.g., 1.24]
        - Git Provider: [e.g., GitHub, GitLab, Azure DevOps]
        - Installation Method: [e.g., binary, Docker, from source]
    validations:
      required: true

  - type: textarea
    id: config
    attributes:
      label: Configuration
      description: If applicable, provide your configuration (redact sensitive information).
      placeholder: |
        ```yaml
        # Your config.yaml (with sensitive data removed)
        ```
    validations:
      required: false

  - type: textarea
    id: logs
    attributes:
      label: Logs
      description: Please include relevant log output.
      placeholder: |
        ```
        Paste logs here
        ```
      render: shell
    validations:
      required: false

  - type: textarea
    id: additional-context
    attributes:
      label: Additional Context
      description: Add any other context about the problem here.
      placeholder: "Any additional information that might be helpful."
    validations:
      required: false

  - type: checkboxes
    id: checklist
    attributes:
      label: Checklist
      description: Please confirm the following before submitting.
      options:
        - label: I have searched existing issues to ensure this is not a duplicate
          required: true
        - label: I have provided all the information requested above
          required: true
        - label: I am using the latest version of ICE
          required: false
