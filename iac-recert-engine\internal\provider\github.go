// Last Recertification: 2025-12-11T20:58:01+01:00
package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/google/go-github/v57/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	client *github.Client
	owner  string
	repo   string
	logger *zap.Logger
}

func NewGitHubProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
	token := os.Getenv(authCfg.TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("token environment variable %s not set", authCfg.TokenEnv)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Parse owner and repo from URL
	// Expected format: https://github.com/owner/repo or git@github.com:owner/repo.git
	// Simplified parsing for now
	parts := strings.Split(strings.TrimSuffix(repoCfg.URL, ".git"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository URL: %s", repoCfg.URL)
	}
	repo := parts[len(parts)-1]
	owner := parts[len(parts)-2]

	return &GitHubProvider{
		client: client,
		owner:  owner,
		repo:   repo,
		logger: logger,
	}, nil
}

func (p *GitHubProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	repo, _, err := p.client.Repositories.Get(ctx, p.owner, p.repo)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Name:          repo.GetName(),
		Owner:         repo.GetOwner().GetLogin(),
		URL:           repo.GetHTMLURL(),
		Provider:      "github",
		DefaultBranch: repo.GetDefaultBranch(),
	}, nil
}

func (p *GitHubProvider) GetLastModificationDate(ctx context.Context, filePath string) (time.Time, types.Commit, error) {
	opts := &github.CommitsListOptions{
		Path: filePath,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	}

	commits, _, err := p.client.Repositories.ListCommits(ctx, p.owner, p.repo, opts)
	if err != nil {
		return time.Time{}, types.Commit{}, err
	}

	if len(commits) == 0 {
		return time.Time{}, types.Commit{}, fmt.Errorf("no commits found for file: %s", filePath)
	}

	commit := commits[0]
	ts := commit.GetCommit().GetCommitter().GetDate().Time

	c := types.Commit{
		Hash:      commit.GetSHA(),
		Author:    commit.GetCommit().GetAuthor().GetName(),
		Email:     commit.GetCommit().GetAuthor().GetEmail(),
		Message:   commit.GetCommit().GetMessage(),
		Timestamp: ts,
	}

	return ts, c, nil
}

func (p *GitHubProvider) CreateBranch(ctx context.Context, name, baseRef string) error {
	// Get the reference to the base branch
	ref, _, err := p.client.Git.GetRef(ctx, p.owner, p.repo, "refs/heads/"+baseRef)
	if err != nil {
		return err
	}

	// Create the new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + name),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}

	_, _, err = p.client.Git.CreateRef(ctx, p.owner, p.repo, newRef)
	return err
}

func (p *GitHubProvider) CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error) {
	// 1. Get the latest commit of the branch
	ref, _, err := p.client.Git.GetRef(ctx, p.owner, p.repo, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	parentSHA := ref.Object.GetSHA()

	// 2. Create blobs for each change and build a tree
	var entries []*github.TreeEntry
	for _, change := range changes {
		blob, _, err := p.client.Git.CreateBlob(ctx, p.owner, p.repo, &github.Blob{
			Content:  github.String(change.Content),
			Encoding: github.String("utf-8"),
		})
		if err != nil {
			return "", err
		}

		entries = append(entries, &github.TreeEntry{
			Path: github.String(change.Path),
			Type: github.String("blob"),
			Mode: github.String("100644"),
			SHA:  blob.SHA,
		})
	}

	tree, _, err := p.client.Git.CreateTree(ctx, p.owner, p.repo, *ref.Object.SHA, entries)
	if err != nil {
		return "", err
	}

	// 3. Create the commit
	commit := &github.Commit{
		Message: github.String(message),
		Tree:    tree,
		Parents: []*github.Commit{
			{SHA: github.String(parentSHA)},
		},
	}
	newCommit, _, err := p.client.Git.CreateCommit(ctx, p.owner, p.repo, commit, nil)
	if err != nil {
		return "", err
	}

	// 4. Update the reference
	ref.Object.SHA = newCommit.SHA
	_, _, err = p.client.Git.UpdateRef(ctx, p.owner, p.repo, ref, false)
	if err != nil {
		return "", err
	}

	return newCommit.GetSHA(), nil
}

func (p *GitHubProvider) CreatePullRequest(ctx context.Context, cfg types.PRConfig) (*types.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(cfg.Title),
		Head:                github.String(cfg.Branch),
		Base:                github.String(cfg.BaseBranch),
		Body:                github.String(cfg.Description),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := p.client.PullRequests.Create(ctx, p.owner, p.repo, newPR)
	if err != nil {
		return nil, err
	}

	// Add assignees and reviewers if provided
	if len(cfg.Assignees) > 0 {
		_, _, err = p.client.Issues.AddAssignees(ctx, p.owner, p.repo, pr.GetNumber(), cfg.Assignees)
		if err != nil {
			p.logger.Error("failed to add assignees", zap.Error(err))
		}
	}

	if len(cfg.Reviewers) > 0 {
		reviewers := github.ReviewersRequest{
			Reviewers: cfg.Reviewers,
		}
		_, _, err = p.client.PullRequests.RequestReviewers(ctx, p.owner, p.repo, pr.GetNumber(), reviewers)
		if err != nil {
			p.logger.Error("failed to request reviewers", zap.Error(err))
		}
	}

	if len(cfg.Labels) > 0 {
		_, _, err = p.client.Issues.AddLabelsToIssue(ctx, p.owner, p.repo, pr.GetNumber(), cfg.Labels)
		if err != nil {
			p.logger.Error("failed to add labels", zap.Error(err))
		}
	}

	return &types.PullRequest{
		ID:        fmt.Sprintf("%d", pr.GetNumber()),
		URL:       pr.GetHTMLURL(),
		Number:    pr.GetNumber(),
		State:     pr.GetState(),
		CreatedAt: pr.GetCreatedAt().Time,
	}, nil
}

func (p *GitHubProvider) UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error {
	// Parse ID as int
	var prNumber int
	fmt.Sscanf(id, "%d", &prNumber)

	pr := &github.PullRequest{}
	if updates.Title != nil {
		pr.Title = updates.Title
	}
	if updates.Description != nil {
		pr.Body = updates.Description
	}
	if updates.State != nil {
		pr.State = updates.State
	}

	_, _, err := p.client.PullRequests.Edit(ctx, p.owner, p.repo, prNumber, pr)
	return err
}

func (p *GitHubProvider) ClosePullRequest(ctx context.Context, id string, reason string) error {
	state := "closed"
	return p.UpdatePullRequest(ctx, id, PRUpdate{State: &state})
}

func (p *GitHubProvider) AssignPullRequest(ctx context.Context, id string, assignees []string) error {
	var prNumber int
	fmt.Sscanf(id, "%d", &prNumber)
	_, _, err := p.client.Issues.AddAssignees(ctx, p.owner, p.repo, prNumber, assignees)
	return err
}

func (p *GitHubProvider) RequestReviewers(ctx context.Context, id string, reviewers []string) error {
	var prNumber int
	fmt.Sscanf(id, "%d", &prNumber)
	req := github.ReviewersRequest{Reviewers: reviewers}
	_, _, err := p.client.PullRequests.RequestReviewers(ctx, p.owner, p.repo, prNumber, req)
	return err
}

func (p *GitHubProvider) AddLabels(ctx context.Context, id string, labels []string) error {
	var prNumber int
	fmt.Sscanf(id, "%d", &prNumber)
	_, _, err := p.client.Issues.AddLabelsToIssue(ctx, p.owner, p.repo, prNumber, labels)
	return err
}

func (p *GitHubProvider) AddComment(ctx context.Context, id string, comment string) error {
	var prNumber int
	fmt.Sscanf(id, "%d", &prNumber)
	c := &github.IssueComment{Body: &comment}
	_, _, err := p.client.Issues.CreateComment(ctx, p.owner, p.repo, prNumber, c)
	return err
}
