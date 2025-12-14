package api

// Plugin represents a generic plugin interface
type Plugin interface {
	Init(config map[string]string) error
}

// AssignmentPlugin represents a plugin that can resolve assignees for PRs
type AssignmentPlugin interface {
	Plugin
	Resolve(files []FileInfo) (AssignmentResult, error)
}

// FileInfo contains information about a file
type FileInfo struct {
	Path         string
	Size         int64
	LastModified string // ISO 8601 format
	CommitHash   string
	CommitAuthor string
	CommitEmail  string
	CommitMsg    string
}

// AssignmentResult contains the assignment information
type AssignmentResult struct {
	Assignees []string
	Reviewers []string
	Team      string
	Priority  string
}
