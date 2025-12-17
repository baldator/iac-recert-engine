package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.uber.org/zap"
)

type GitLabProvider struct {
	client  *gitlab.Client
	project int64 // Project ID
	logger  *zap.Logger
}

func NewGitLabProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
	token := os.Getenv(authCfg.TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("token environment variable %s not set", authCfg.TokenEnv)
	}

	// Parse URL to get project path
	// https://gitlab.com/group/project.git
	u, err := url.Parse(repoCfg.URL)
	if err != nil {
		return nil, err
	}

	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	// For GitLab, we need the project ID (numeric) or the URL-encoded path
	// First try to get by path, then extract the ID
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(u.Scheme+"://"+u.Host+"/api/v4"))
	if err != nil {
		return nil, err
	}

	project, _, err := client.Projects.GetProject(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &GitLabProvider{
		client:  client,
		project: project.ID,
		logger:  logger,
	}, nil
}


func (p *GitLabProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	project, _, err := p.client.Projects.GetProject(p.project, nil)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Name:          project.Name,
		Owner:         project.PathWithNamespace,
		URL:           project.WebURL,
		Provider:      "gitlab",
		DefaultBranch: project.DefaultBranch,
	}, nil
}

func (p *GitLabProvider) GetLastModificationDate(ctx context.Context, filePath string) (time.Time, types.Commit, error) {
	// Use the repository commits API
	opts := &gitlab.ListCommitsOptions{
		Path: &filePath,
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
		},
	}

	commits, _, err := p.client.Commits.ListCommits(p.project, opts)
	if err != nil {
		return time.Time{}, types.Commit{}, err
	}

	if len(commits) == 0 {
		return time.Time{}, types.Commit{}, fmt.Errorf("no commits found for file: %s", filePath)
	}

	c := commits[0]
	commit := types.Commit{
		Hash:      c.ID,
		Author:    c.AuthorName,
		Email:     c.AuthorEmail,
		Message:   c.Message,
		Timestamp: *c.CommittedDate,
	}

	return *c.CommittedDate, commit, nil
}

func (p *GitLabProvider) BranchExists(ctx context.Context, name string) (bool, error) {
	_, _, err := p.client.Branches.GetBranch(p.project, name)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *GitLabProvider) CreateBranch(ctx context.Context, name, baseRef string) error {
	opts := &gitlab.CreateBranchOptions{
		Branch: &name,
		Ref:    &baseRef,
	}
	_, _, err := p.client.Branches.CreateBranch(p.project, opts)
	return err
}

func (p *GitLabProvider) CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error) {
	// Build actions for each change
	actions := make([]*gitlab.CommitAction, len(changes))
	for i, change := range changes {
		actions[i] = &gitlab.CommitAction{
			Action:   gitlab.FileCreate,
			FilePath: change.Path,
			Content:  change.Content,
			Encoding: "text",
		}
	}

	opts := &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: &message,
		Actions:       actions,
	}

	commit, _, err := p.client.Commits.CreateCommit(p.project, opts)
	if err != nil {
		return "", err
	}

	return commit.ID, nil
}

func (p *GitLabProvider) PullRequestExists(ctx context.Context, headBranch, baseBranch string) (bool, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: &headBranch,
		TargetBranch: &baseBranch,
		State:        &[]string{"opened"}[0],
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
		},
	}

	mrs, _, err := p.client.MergeRequests.ListProjectMergeRequests(p.project, opts)
	if err != nil {
		return false, err
	}

	return len(mrs) > 0, nil
}

func (p *GitLabProvider) CreatePullRequest(ctx context.Context, cfg types.PRConfig) (*types.PullRequest, error) {
	opts := &gitlab.CreateMergeRequestOptions{
		SourceBranch: &cfg.Branch,
		TargetBranch: &cfg.BaseBranch,
		Title:        &cfg.Title,
		Description:  &cfg.Description,
	}

	mr, _, err := p.client.MergeRequests.CreateMergeRequest(p.project, opts)
	if err != nil {
		return nil, err
	}

	return &types.PullRequest{
		ID:        strconv.FormatInt(mr.IID, 10),
		URL:       mr.WebURL,
		Number:    int(mr.IID),
		State:     mr.State,
		CreatedAt: *mr.CreatedAt,
	}, nil
}

func (p *GitLabProvider) UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error {
	mrIID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	opts := &gitlab.UpdateMergeRequestOptions{}
	if updates.Title != nil {
		opts.Title = updates.Title
	}
	if updates.Description != nil {
		opts.Description = updates.Description
	}
	if updates.State != nil {
		opts.StateEvent = updates.State
	}

	_, _, err = p.client.MergeRequests.UpdateMergeRequest(p.project, mrIID, opts)
	return err
}

func (p *GitLabProvider) ClosePullRequest(ctx context.Context, id string, reason string) error {
	state := "close"
	return p.UpdatePullRequest(ctx, id, PRUpdate{State: &state})
}

func (p *GitLabProvider) AssignPullRequest(ctx context.Context, id string, assignees []string) error {
	mrIID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	// Resolve usernames to user IDs
	var assigneeIDs []int
	for _, username := range assignees {
		user, _, err := p.client.Users.GetUser(username, gitlab.GetUsersOptions{})
		if err != nil {
			p.logger.Warn("Failed to resolve username to user ID", zap.String("username", username), zap.Error(err))
			continue
		}
		assigneeIDs = append(assigneeIDs, user.ID)
	}

	if len(assigneeIDs) == 0 {
		return fmt.Errorf("no valid assignees found")
	}

	opts := &gitlab.UpdateMergeRequestOptions{
		AssigneeIDs: &assigneeIDs,
	}

	_, _, err = p.client.MergeRequests.UpdateMergeRequest(p.project, mrIID, opts)
	return err
}

func (p *GitLabProvider) RequestReviewers(ctx context.Context, id string, reviewers []string) error {
	// GitLab doesn't have the same reviewer system as GitHub
	// Reviewers are typically assigned as assignees or mentioned in comments
	return fmt.Errorf("RequestReviewers not supported for GitLab")
}

func (p *GitLabProvider) AddLabels(ctx context.Context, id string, labels []string) error {
	mrIID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	// Convert labels to LabelOptions
	labelOpts := gitlab.LabelOptions(labels)
	opts := &gitlab.UpdateMergeRequestOptions{
		Labels: &labelOpts,
	}

	_, _, err = p.client.MergeRequests.UpdateMergeRequest(p.project, mrIID, opts)
	return err
}

func (p *GitLabProvider) AddComment(ctx context.Context, id string, comment string) error {
	mrIID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	opts := &gitlab.CreateMergeRequestNoteOptions{
		Body: &comment,
	}

	_, _, err = p.client.Notes.CreateMergeRequestNote(p.project, mrIID, opts)
	return err
}
