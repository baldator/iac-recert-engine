package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestAuditor_FileStorage(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.AuditConfig{
		Enabled: true,
		Storage: "file",
		Config: map[string]string{
			"directory": tmpDir,
		},
	}

	logger := zaptest.NewLogger(t)
	auditor := NewAuditor(cfg, logger, "test-run-123")

	ctx := context.Background()

	// Log an event
	testDetails := map[string]any{
		"test_key": "test_value",
		"count":    42,
	}
	auditor.LogEvent(ctx, EventRunStart, "Test message", testDetails, nil)

	// Check if file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "audit-*.log"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 audit file, got %d", len(files))
	}

	// Read and parse the file
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}

	// Should contain one JSON line
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var event AuditEvent
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatal(err)
	}

	// Verify event
	if event.RunID != "test-run-123" {
		t.Errorf("Expected run ID 'test-run-123', got '%s'", event.RunID)
	}
	if event.EventType != EventRunStart {
		t.Errorf("Expected event type '%s', got '%s'", EventRunStart, event.EventType)
	}
	if event.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", event.Message)
	}
	if event.Details.(map[string]any)["test_key"] != "test_value" {
		t.Errorf("Expected details.test_key = 'test_value', got %v", event.Details)
	}
	if event.Error != "" {
		t.Errorf("Expected no error, got '%s'", event.Error)
	}
}

func TestAuditor_ErrorLogging(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.AuditConfig{
		Enabled: true,
		Storage: "file",
		Config: map[string]string{
			"directory": tmpDir,
		},
	}

	logger := zaptest.NewLogger(t)
	auditor := NewAuditor(cfg, logger, "test-run-456")

	ctx := context.Background()

	// Log an error event
	testErr := os.ErrNotExist
	auditor.LogEvent(ctx, EventError, "Test error message", nil, testErr)

	// Check if file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "audit-*.log"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 audit file, got %d", len(files))
	}

	// Read and parse the file
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}

	var event AuditEvent
	if err := json.Unmarshal(content, &event); err != nil {
		t.Fatal(err)
	}

	// Verify error is logged
	if event.Error == "" {
		t.Error("Expected error to be logged")
	}
	if !strings.Contains(event.Error, "file does not exist") {
		t.Errorf("Expected error message to contain 'file does not exist', got '%s'", event.Error)
	}
}

func TestAuditor_Disabled(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.AuditConfig{
		Enabled: false, // Disabled
		Storage: "file",
		Config: map[string]string{
			"directory": tmpDir,
		},
	}

	logger := zaptest.NewLogger(t)
	auditor := NewAuditor(cfg, logger, "test-run-789")

	ctx := context.Background()

	// Log an event
	auditor.LogEvent(ctx, EventRunStart, "Test message", nil, nil)

	// Check that no file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "audit-*.log"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf("Expected no audit files when disabled, got %d", len(files))
	}
}

func TestAuditor_InvalidStorage(t *testing.T) {
	cfg := config.AuditConfig{
		Enabled: true,
		Storage: "invalid",
		Config:  map[string]string{},
	}

	logger := zaptest.NewLogger(t)
	auditor := NewAuditor(cfg, logger, "test-run-invalid")

	ctx := context.Background()

	// This should not panic, but log an error
	auditor.LogEvent(ctx, EventRunStart, "Test message", nil, nil)

	// Test passes if no panic occurs
}

func TestAuditor_S3Storage_MissingConfig(t *testing.T) {
	cfg := config.AuditConfig{
		Enabled: true,
		Storage: "s3",
		Config:  map[string]string{}, // Missing bucket
	}

	logger := zaptest.NewLogger(t)
	auditor := NewAuditor(cfg, logger, "test-run-s3")

	ctx := context.Background()

	// This should handle the error gracefully
	auditor.LogEvent(ctx, EventRunStart, "Test message", nil, nil)

	// Test passes if no panic occurs
}

func TestAuditEvent_JSON(t *testing.T) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		RunID:     "test-run",
		EventType: EventRunStart,
		Message:   "Test message",
		Details: map[string]any{
			"key": "value",
		},
		Error: "test error",
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	// Test JSON unmarshaling
	var decoded AuditEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.RunID != event.RunID {
		t.Errorf("RunID mismatch")
	}
	if decoded.EventType != event.EventType {
		t.Errorf("EventType mismatch")
	}
	if decoded.Message != event.Message {
		t.Errorf("Message mismatch")
	}
	if decoded.Error != event.Error {
		t.Errorf("Error mismatch")
	}
}
