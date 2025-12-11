// Last Recertification: 2025-12-11T22:49:52+01:00
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type GitLabProvider struct {
	client  *http.Client
	baseURL string
	token   string
	project string // Project ID or URL-encoded path
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

	// URL encode the project path for API usage
	projectID := url.PathEscape(path)

	baseURL := "https://gitlab.com/api/v4"
	if u.Host != "gitlab.com" {
		baseURL = fmt.Sprintf("https://%s/api/v4", u.Host)
	}

	return &GitLabProvider{
		client:  &http.Client{},
		baseURL: baseURL,
		token:   token,
		project: projectID,
		logger:  logger,
	}, nil
}

func (p *GitLabProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/%s", p.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed: %s %s: %s", method, url, string(bodyBytes))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (p *GitLabProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	var repo struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
		DefaultBranch     string `json:"default_branch"`
	}

	if err := p.doRequest(ctx, "GET", fmt.Sprintf("projects/%s", p.project), nil, &repo); err != nil {
		return nil, err
	}

	return &Repository{
		Name:          repo.Name,
		Owner:         repo.PathWithNamespace, // Approximate mapping
		URL:           repo.WebURL,
		Provider:      "gitlab",
		DefaultBranch: repo.DefaultBranch,
	}, nil
}

func (p *GitLabProvider) GetLastModificationDate(ctx context.Context, filePath string) (time.Time, types.Commit, error) {
	// GET /projects/:id/repository/commits?path=...&per_page=1
	path := fmt.Sprintf("projects/%s/repository/commits?path=%s&per_page=1", p.project, url.QueryEscape(filePath))

	var commits []struct {
		ID            string    `json:"id"`
		AuthorName    string    `json:"author_name"`
		AuthorEmail   string    `json:"author_email"`
		Message       string    `json:"message"`
		CommittedDate time.Time `json:"committed_date"`
	}

	if err := p.doRequest(ctx, "GET", path, nil, &commits); err != nil {
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
		Timestamp: c.CommittedDate,
	}

	return c.CommittedDate, commit, nil
}

func (p *GitLabProvider) CreateBranch(ctx context.Context, name, baseRef string) error {
	return fmt.Errorf("CreateBranch not implemented for GitLab")
}

func (p *GitLabProvider) CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error) {
	return "", fmt.Errorf("CreateCommit not implemented for GitLab")
}

func (p *GitLabProvider) CreatePullRequest(ctx context.Context, cfg types.PRConfig) (*types.PullRequest, error) {
	// POST /projects/:id/merge_requests
	path := fmt.Sprintf("projects/%s/merge_requests", p.project)

	body := map[string]interface{}{
		"source_branch": cfg.Branch,
		"target_branch": cfg.BaseBranch,
		"title":         cfg.Title,
		"description":   cfg.Description,
	}

	var mr struct {
		IID       int       `json:"iid"`
		WebURL    string    `json:"web_url"`
		State     string    `json:"state"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := p.doRequest(ctx, "POST", path, body, &mr); err != nil {
		return nil, err
	}

	return &types.PullRequest{
		ID:        strconv.Itoa(mr.IID),
		URL:       mr.WebURL,
		Number:    mr.IID,
		State:     mr.State,
		CreatedAt: mr.CreatedAt,
	}, nil
}

func (p *GitLabProvider) UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error {
	return fmt.Errorf("UpdatePullRequest not implemented for GitLab")
}

func (p *GitLabProvider) ClosePullRequest(ctx context.Context, id string, reason string) error {
	return fmt.Errorf("ClosePullRequest not implemented for GitLab")
}

func (p *GitLabProvider) AssignPullRequest(ctx context.Context, id string, assignees []string) error {
	return fmt.Errorf("AssignPullRequest not implemented for GitLab")
}

func (p *GitLabProvider) RequestReviewers(ctx context.Context, id string, reviewers []string) error {
	return fmt.Errorf("RequestReviewers not implemented for GitLab")
}

func (p *GitLabProvider) AddLabels(ctx context.Context, id string, labels []string) error {
	return fmt.Errorf("AddLabels not implemented for GitLab")
}

func (p *GitLabProvider) AddComment(ctx context.Context, id string, comment string) error {
	return fmt.Errorf("AddComment not implemented for GitLab")
}
