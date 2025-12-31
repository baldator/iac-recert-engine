# CSV Lookup Assignment Plugin

This plugin assigns pull requests to reviewers based on a CSV file lookup. It extracts a key from IaC files using a regex pattern and looks up the corresponding reviewer in a CSV file.

## Features

- Extracts keys from IaC files using configurable regex patterns
- Looks up reviewers in a CSV file based on key-value mapping
- Supports configurable CSV columns for keys and values
- Handles CSV files with or without headers

## Configuration

Add to your ICE configuration:

```yaml
plugins:
  csv_assignment:
    enabled: true
    type: "assignment"
    module: "csvlookup"
    config:
      csv_file: "/path/to/reviewers.csv"
      key_regex: "team\\s*=\\s*[\"']([^\"']+)[\"']"
      key_column: "0"
      value_column: "1"

assignment:
  strategy: "plugin"
  plugin_name: "csv_assignment"
```

## Configuration Parameters

- `csv_file`: Path to the CSV file containing the key-value mappings (required)
- `key_regex`: Regex pattern to extract the key from IaC files. Must contain a capture group (required)
- `key_column`: Zero-based column index containing the keys in the CSV file (required)
- `value_column`: Zero-based column index containing the reviewer values in the CSV file (required)

## CSV File Format

The CSV file should contain key-value pairs where:
- The key column contains the values that will be matched against extracted keys
- The value column contains the reviewer usernames/emails to assign

Example CSV file:
```csv
team,reviewer
platform,alice@example.com
security,bob@example.com
network,charlie@example.com
```

## Example Usage

Given a Terraform file with:
```hcl
resource "aws_instance" "example" {
  team = "platform"
  # ...
}
```

And the regex `team\s*=\s*["']([^"']+)["']`, the plugin will:
1. Extract "platform" as the key
2. Look up "platform" in the CSV file
3. Return "alice@example.com" as the assignee

## Installation

This plugin is designed to be used as a separate Go module. In production:

1. Clone this repository
2. Import it in your main ICE codebase
3. Build with the plugin included

For local development, use Go module replace directives as shown in the main repository's go.mod.
