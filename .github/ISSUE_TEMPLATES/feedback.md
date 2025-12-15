---
name: Feedback & General Discussion
description: Provide feedback, ask questions, or start general discussions
title: "[FEEDBACK] "
labels: ["feedback", "question"]
assignees: []
body:
  - type: markdown
    attributes:
      value: |
        ## Feedback & General Discussion

        We welcome your feedback, questions, and general discussions about ICE! This template is for anything that doesn't fit the bug report or feature request categories.

  - type: dropdown
    id: feedback-type
    attributes:
      label: Type of Feedback
      description: What type of feedback are you providing?
      options:
        - General feedback about ICE
        - Question about usage or configuration
        - Documentation feedback or suggestion
        - Performance feedback
        - Security concern or question
        - Integration question
        - Other
    validations:
      required: true

  - type: textarea
    id: summary
    attributes:
      label: Summary
      description: Provide a brief summary of your feedback or question.
      placeholder: "A short description of what you want to discuss."
    validations:
      required: true

  - type: textarea
    id: details
    attributes:
      label: Details
      description: Provide more details about your feedback, question, or discussion topic.
      placeholder: "Please provide as much detail as possible. Include examples, use cases, or specific scenarios if relevant."
    validations:
      required: true

  - type: dropdown
    id: experience-level
    attributes:
      label: Your Experience Level
      description: How familiar are you with ICE and IaC recertification?
      options:
        - New to ICE (first time user)
        - Have used ICE a few times
        - Regular ICE user
        - ICE power user/expert
        - Just learning about IaC recertification
    validations:
      required: false

  - type: dropdown
    id: satisfaction
    attributes:
      label: Overall Satisfaction (if applicable)
      description: If you've used ICE, how satisfied are you with it?
      options:
        - Very satisfied
        - Satisfied
        - Neutral
        - Dissatisfied
        - Very dissatisfied
        - Haven't used ICE yet
    validations:
      required: false

  - type: checkboxes
    id: aspects
    attributes:
      label: Aspects of ICE (select all that apply)
      description: Which aspects of ICE does your feedback relate to?
      options:
        - label: Ease of installation/setup
        - label: Documentation quality
        - label: User interface (CLI)
        - label: Configuration options
        - label: Performance/speed
        - label: Reliability/stability
        - label: Feature completeness
        - label: Integration with Git providers
        - label: Plugin system
        - label: Docker support
        - label: CI/CD integration
        - label: Security features
        - label: Audit/logging capabilities

  - type: textarea
    id: suggestions
    attributes:
      label: Suggestions for Improvement
      description: If you have specific suggestions for how we can improve ICE, please share them here.
      placeholder: "What would you like to see changed or improved?"
    validations:
      required: false

  - type: textarea
    id: environment
    attributes:
      label: Environment (if relevant)
      description: If your feedback is related to a specific environment or setup, please provide details.
      value: |
        - OS: [e.g., Windows 11, macOS 12.1, Ubuntu 20.04]
        - ICE Version: [e.g., v1.0.0]
        - Go Version: [e.g., 1.24]
        - Git Provider: [e.g., GitHub, GitLab, Azure DevOps]
        - Installation Method: [e.g., binary, Docker, from source]
    validations:
      required: false

  - type: textarea
    id: additional-context
    attributes:
      label: Additional Context
      description: Add any other context, links, or information that would be helpful.
      placeholder: "Links, screenshots, code examples, or anything else relevant."
    validations:
      required: false

  - type: checkboxes
    id: contact
    attributes:
      label: Contact Preferences
      description: How would you like us to follow up?
      options:
        - label: I would like to be contacted for clarification if needed
          required: false
        - label: I would like to help test potential solutions
          required: false
        - label: I would like to contribute to improvements
          required: false
