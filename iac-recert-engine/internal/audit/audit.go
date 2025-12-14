package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/baldator/iac-recert-engine/internal/config"
	"go.uber.org/zap"
)

// EventType represents different types of audit events
type EventType string

const (
	EventRunStart       EventType = "run_start"
	EventRunEnd         EventType = "run_end"
	EventScanComplete   EventType = "scan_complete"
	EventEnrichComplete EventType = "enrich_complete"
	EventCheckComplete  EventType = "check_complete"
	EventGroupComplete  EventType = "group_complete"
	EventPRCreated      EventType = "pr_created"
	EventPRError        EventType = "pr_error"
	EventError          EventType = "error"
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	RunID       string     `json:"run_id"`
	EventType   EventType  `json:"event_type"`
	Message     string     `json:"message"`
	Details     any        `json:"details,omitempty"`
	Error       string     `json:"error,omitempty"`
	Repository  string     `json:"repository,omitempty"`
	User        string     `json:"user,omitempty"`
}

// Auditor handles audit logging
type Auditor struct {
	cfg    config.AuditConfig
	logger *zap.Logger
	runID  string
}

// NewAuditor creates a new auditor
func NewAuditor(cfg config.AuditConfig, logger *zap.Logger, runID string) *Auditor {
	return &Auditor{
		cfg:    cfg,
		logger: logger,
		runID:  runID,
	}
}

// LogEvent logs an audit event
func (a *Auditor) LogEvent(ctx context.Context, eventType EventType, message string, details any, err error) {
	if !a.cfg.Enabled {
		return
	}

	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		RunID:     a.runID,
		EventType: eventType,
		Message:   message,
		Details:   details,
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Log to structured logger
	a.logger.Info("audit event",
		zap.String("run_id", event.RunID),
		zap.String("event_type", string(event.EventType)),
		zap.String("message", event.Message),
		zap.Any("details", event.Details),
		zap.String("error", event.Error),
	)

	// Store to configured backend
	if storeErr := a.storeEvent(ctx, event); storeErr != nil {
		a.logger.Error("failed to store audit event", zap.Error(storeErr))
	}
}

// storeEvent stores the event based on the configured storage type
func (a *Auditor) storeEvent(ctx context.Context, event AuditEvent) error {
	switch a.cfg.Storage {
	case "file":
		return a.storeToFile(event)
	case "s3":
		return a.storeToS3(ctx, event)
	default:
		return fmt.Errorf("unsupported audit storage type: %s", a.cfg.Storage)
	}
}

// storeToFile stores the event to a local file
func (a *Auditor) storeToFile(event AuditEvent) error {
	dir, ok := a.cfg.Config["directory"]
	if !ok {
		dir = "./audit"
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Create filename with date
	filename := filepath.Join(dir, fmt.Sprintf("audit-%s.log", event.Timestamp.Format("2006-01-02")))

	// Marshal to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Append to file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return nil
}

// storeToS3 stores the event to S3
func (a *Auditor) storeToS3(ctx context.Context, event AuditEvent) error {
	bucket, ok := a.cfg.Config["bucket"]
	if !ok {
		return fmt.Errorf("bucket not configured for S3 audit storage")
	}

	prefix, ok := a.cfg.Config["prefix"]
	if !ok {
		prefix = "iac-recert/"
	}

	// Create S3 session
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	uploader := s3manager.NewUploader(sess)

	// Create key with date and run ID
	key := fmt.Sprintf("%saudit-%s-%s.log", prefix, event.Timestamp.Format("2006-01-02"), event.RunID)

	// Marshal to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Upload to S3
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return fmt.Errorf("failed to upload audit event to S3: %w", err)
	}

	return nil
}
