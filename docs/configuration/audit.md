# Audit Configuration

ICE provides comprehensive audit logging to track all recertification activities for compliance and governance purposes. Audit logs capture key events throughout the recertification process.

## Overview

The audit system logs structured events for:
- Run start and completion
- File scanning operations
- History analysis
- Recertification checks
- Grouping operations
- Pull request creation
- Errors and failures

## Configuration

```yaml
audit:
  enabled: true
  storage: "file"  # file or s3
  config:
    directory: "./audit"  # For file storage
    # bucket: "audit-logs-bucket"  # For S3 storage
    # prefix: "iac-recert/"  # For S3 storage
```

## Storage Options

### File Storage
Stores audit events in local JSON files, one file per day.

```yaml
audit:
  enabled: true
  storage: "file"
  config:
    directory: "./audit"  # Directory to store audit files
```

Files are created with the pattern `audit-YYYY-MM-DD.log` and contain one JSON event per line.

### S3 Storage
Stores audit events in Amazon S3 for centralized logging and long-term retention.

```yaml
audit:
  enabled: true
  storage: "s3"
  config:
    bucket: "my-audit-logs"
    prefix: "iac-recert/"  # Optional prefix for S3 keys
```

Events are stored as individual objects with keys like `iac-recert/audit-2025-12-12-runid.log`.

## Audit Events

### Event Types

| Event Type | Description |
|------------|-------------|
| `run_start` | Recertification run initiated |
| `run_end` | Recertification run completed |
| `scan_complete` | File scanning phase finished |
| `enrich_complete` | History analysis phase finished |
| `check_complete` | Recertification check phase finished |
| `group_complete` | File grouping phase finished |
| `pr_created` | Pull request successfully created |
| `pr_error` | Pull request creation failed |
| `error` | General error occurred |

### Event Structure

Each audit event is a JSON object:

```json
{
  "timestamp": "2025-12-12T10:30:45Z",
  "run_id": "a1b2c3d4e5f6",
  "event_type": "run_start",
  "message": "Starting recertification run",
  "details": {
    "repository": "https://github.com/org/repo",
    "dry_run": false
  },
  "error": "",
  "repository": "https://github.com/org/repo",
  "user": ""
}
```

## Audit Log Analysis

### File Storage Analysis
```bash
# View recent audit events
tail -f audit/audit-$(date +%Y-%m-%d).log

# Search for specific events
grep "pr_created" audit/audit-*.log

# Count events by type
cat audit/audit-*.log | jq -r '.event_type' | sort | uniq -c
```

### S3 Storage Analysis
```bash
# List audit objects
aws s3 ls s3://my-audit-logs/iac-recert/

# Download and analyze
aws s3 cp s3://my-audit-logs/iac-recert/audit-2025-12-12-runid.log - | jq .
```

## Security Considerations

- Audit logs may contain sensitive information about repositories and users
- Store audit logs securely with appropriate access controls
- Consider log retention policies for compliance requirements
- Use encryption for S3-stored audit logs

## Performance Impact

- Audit logging has minimal performance impact when using file storage
- S3 storage may add network latency for each event
- Consider batching events for high-volume scenarios (future enhancement)

## Troubleshooting

### Common Issues

**Audit files not created**
- Check that the audit directory is writable
- Verify audit.enabled is set to true
- Check file permissions

**S3 upload failures**
- Verify AWS credentials are configured
- Check S3 bucket permissions
- Ensure bucket exists and is accessible

**Large audit files**
- Files grow with each run
- Consider log rotation or compression
- Use S3 storage for automatic scaling

## Integration with Monitoring

Audit events are also logged to the structured logger and can be integrated with monitoring systems:

```bash
# Extract metrics from audit logs
cat audit/audit-*.log | jq -r 'select(.event_type == "pr_created") | .timestamp' | wc -l
```

## Best Practices

1. **Enable audit logging** for production deployments
2. **Use S3 storage** for multi-environment setups
3. **Monitor audit log size** and implement rotation
4. **Secure audit storage** with appropriate permissions
5. **Regularly review audit logs** for compliance
6. **Set up alerts** for audit failures
7. **Retain logs** according to your compliance requirements
