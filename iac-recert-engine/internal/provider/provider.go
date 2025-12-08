package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type GitProvider interface {
	GetRepository(ctx context.Context, url string) (*Repository, error)
	GetLastModificationDate(ctx context.Context, filePath string) (time.Time, types.Commit, error)

	CreateBranch(ctx context.Context, name, baseRef string) error
	CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error)

	CreatePullRequest(ctx context.Context, cfg types.PRConfig) (*types.PullRequest, error)
	UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error
	ClosePullRequest(ctx context.Context, id string, reason string) error

	AssignPullRequest(ctx context.Context, id string, assignees []string) error
	RequestReviewers(ctx context.Context, id string, reviewers []string) error

	AddLabels(ctx context.Context, id string, labels []string) error
	AddComment(ctx context.Context, id string, comment string) error
}

type Repository struct {
	Name          string
	Owner         string
	URL           string
	Provider      string
	DefaultBranch string
}

type PRUpdate struct {
	Title       *string
	Description *string
	State       *string
}

func NewProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
	switch repoCfg.Provider {
	case "github":
		return NewGitHubProvider(ctx, repoCfg, authCfg, logger)
	case "azure":
		return NewAzureDevOpsProvider(ctx, repoCfg, authCfg, logger)
	case "gitlab":
		return NewGitLabProvider(ctx, repoCfg, authCfg, logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", repoCfg.Provider)
	}
}
