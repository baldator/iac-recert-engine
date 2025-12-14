# ServiceNow Assignment Plugin

This plugin assigns pull requests to the "supported by" user for business applications extracted from IaC files.

## Features

- Extracts business application names from IaC files using regex pattern `application = "value"`
- Queries ServiceNow CMDB API to find the supported_by user for the application
- Returns the user's email or username as the assignee

## Configuration

Add to your ICE configuration:

```yaml
plugins:
  servicenow_assignment:
    enabled: true
    type: "assignment"
    module: "servicenow"
    config:
      api_url: "https://your-instance.service-now.com"
      username: "${SERVICENOW_USERNAME}"
      password: "${SERVICENOW_PASSWORD}"

assignment:
  strategy: "plugin"
  plugin_name: "servicenow_assignment"
```

## Environment Variables

- `SERVICENOW_USERNAME`: ServiceNow username
- `SERVICENOW_PASSWORD`: ServiceNow password

## API Requirements

The plugin requires access to these ServiceNow tables:
- `cmdb_ci_business_app` (read)
- `sys_user` (read)

## Installation

This plugin is designed to be used as a separate Go module. In production:

1. Clone this repository
2. Import it in your main ICE codebase
3. Build with the plugin included

For local development, use Go module replace directives as shown in the main repository's go.mod.
