// Last Recertification: 2025-12-11T22:51:29+01:00
package types

import (
	"time"
)

type FileInfo struct {
	Path         string
	Size         int64
	LastModified time.Time
	CommitHash   string
	CommitAuthor string
	CommitEmail  string
	CommitMsg    string
}

type RecertCheckResult struct {
	File        FileInfo
	PatternName string
	DaysSince   int
	Threshold   int
	Priority    string // Critical, High, Medium, Low
	NeedsRecert bool
	NextDueDate time.Time
}

type FileGroup struct {
	ID        string
	Strategy  string
	Files     []RecertCheckResult
	Assignees []string
	Reviewers []string
}

type ExecutionResult struct {
	RunID      string
	StartTime  time.Time
	EndTime    time.Time
	Success    bool
	Errors     []string
	PRsCreated []string
}

type AssignmentResult struct {
	Assignees []string
	Reviewers []string
	Team      string
	Priority  string
}

type PRConfig struct {
	Title       string
	Description string
	Branch      string
	BaseBranch  string
	Files       []string // List of files included
	Assignees   []string
	Reviewers   []string
	Labels      []string
}

type PullRequest struct {
	ID        string
	URL       string
	Number    int
	State     string
	CreatedAt time.Time
}

type Commit struct {
	Hash      string
	Author    string
	Email     string
	Message   string
	Timestamp time.Time
}

type Change struct {
	Path    string
	Content string // For creating commits (if needed)
}
