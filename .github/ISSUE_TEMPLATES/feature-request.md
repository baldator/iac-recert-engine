---
name: Feature Request
description: Suggest a new feature or enhancement
title: "[FEATURE] "
labels: ["enhancement", "triage"]
assignees: []
body:
  - type: markdown
    attributes:
      value: |
        ## Feature Request

        Thank you for suggesting a new feature! Your ideas help make ICE better for everyone.

  - type: textarea
    id: summary
    attributes:
      label: Feature Summary
      description: Provide a brief summary of the feature you'd like to see.
      placeholder: "A short description of what you want to add or change."
    validations:
      required: true

  - type: textarea
    id: problem
    attributes:
      label: Problem/Use Case
      description: Describe the problem this feature would solve or the use case it would enable.
      placeholder: "What's the problem you're trying to solve? What's your use case?"
    validations:
      required: true

  - type: textarea
    id: solution
    attributes:
      label: Proposed Solution
      description: Describe your proposed solution or implementation approach.
      placeholder: "How would you like this feature to work? What would the user experience be like?"
    validations:
      required: true

  - type: textarea
    id: alternatives
    attributes:
      label: Alternative Solutions
      description: Have you considered any alternative approaches?
      placeholder: "What other solutions have you considered? Why might they be less ideal?"
    validations:
      required: false

  - type: dropdown
    id: priority
    attributes:
      label: Priority
      description: How important is this feature to you?
      options:
        - Nice to have
        - Would be helpful
        - Important for my use case
        - Critical/blocking my adoption of ICE
    validations:
      required: true

  - type: dropdown
    id: complexity
    attributes:
      label: Estimated Complexity
      description: How complex do you think this feature would be to implement?
      options:
        - Simple (small changes, low risk)
        - Medium (moderate changes, some risk)
        - Complex (major changes, high risk)
        - Unknown
    validations:
      required: false

  - type: checkboxes
    id: areas
    attributes:
      label: Affected Areas
      description: Which parts of ICE would this feature affect?
      options:
        - label: Core scanning functionality
        - label: Configuration management
        - label: Git provider integrations (GitHub, GitLab, Azure DevOps)
        - label: Plugin system
        - label: CLI interface
        - label: Docker/container support
        - label: Documentation
        - label: CI/CD integration
        - label: Audit/logging
        - label: Assignment strategies
        - label: PR creation and management

  - type: textarea
    id: additional-context
    attributes:
      label: Additional Context
      description: Add any other context, screenshots, or examples that would be helpful.
      placeholder: "Links, screenshots, code examples, or anything else that might help explain your request."
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
        - label: This feature would benefit other users, not just my specific use case
          required: false
        - label: I am willing to contribute to the implementation if needed
          required: false
