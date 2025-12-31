package main

import (
	"fmt"
	"log"
	"time"

	"github.com/baldator/iac-recert-csvlookup-plugin/csvlookup"
	"github.com/baldator/iac-recert-engine/pkg/api"
	"go.uber.org/zap"
)

func main() {
	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	// Create plugin
	plugin := csvlookup.NewCSVLookupPlugin(logger)

	// Initialize plugin
	config := map[string]string{
		"csv_file":    "test_reviewers.csv",
		"key_regex":   `team\s*=\s*["']([^"']+)["']`,
		"key_column":  "0",
		"value_column": "1",
	}

	err = plugin.Init(config)
	if err != nil {
		log.Fatal("Failed to init plugin:", err)
	}

	// Create test file info
	files := []api.FileInfo{
		{
			Path:         "test_terraform.tf",
			Size:         1024,
			LastModified: time.Now().Format("2006-01-02T15:04:05Z07:00"),
			CommitHash:   "abc123",
			CommitAuthor: "test@example.com",
			CommitEmail:  "test@example.com",
			CommitMsg:    "Test commit",
		},
	}

	// Resolve assignment
	result, err := plugin.Resolve(files)
	if err != nil {
		log.Fatal("Failed to resolve:", err)
	}

	fmt.Printf("Assignees: %v\n", result.Assignees)
	fmt.Printf("Reviewers: %v\n", result.Reviewers)
	fmt.Printf("Team: %s\n", result.Team)
	fmt.Printf("Priority: %s\n", result.Priority)
}
