// Last Recertification: 2025-12-11T22:51:29+01:00
package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type AzureDevOpsProvider struct {
	client  *http.Client
	org     string
	project string
	repo    string
	token   string
	logger  *zap.Logger
	baseURL string
}

func NewAzureDevOpsProvider(ctx context.Context, repoCfg config.RepositoryConfig, authCfg config.AuthConfig, logger *zap.Logger) (GitProvider, error) {
	token := os.Getenv(authCfg.TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("token environment variable %s not set", authCfg.TokenEnv)
	}

	// Parse URL: https://dev.azure.com/{org}/{project}/_git/{repo}
	// or https://{org}.visualstudio.com/{project}/_git/{repo}
	// Simplified parsing
	parts := strings.Split(repoCfg.URL, "/")
	var org, project, repo string
	if strings.Contains(repoCfg.URL, "dev.azure.com") {
		// https://dev.azure.com/org/project/_git/repo
		if len(parts) < 6 {
			return nil, fmt.Errorf("invalid Azure DevOps URL: %s", repoCfg.URL)
		}
		org = parts[3]
		project = parts[4]
		repo = parts[6]
	} else if strings.Contains(repoCfg.URL, "visualstudio.com") {
		// https://org.visualstudio.com/project/_git/repo
		hostParts := strings.Split(parts[2], ".")
		org = hostParts[0]
		project = parts[3]
		repo = parts[5]
	} else {
		return nil, fmt.Errorf("unsupported Azure DevOps URL format: %s", repoCfg.URL)
	}

	return &AzureDevOpsProvider{
		client:  &http.Client{},
		org:     org,
		project: project,
		repo:    repo,
		token:   token,
		logger:  logger,
		baseURL: fmt.Sprintf("https://dev.azure.com/%s/%s/_apis", org, project),
	}, nil
}

func (p *AzureDevOpsProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
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

	auth := base64.StdEncoding.EncodeToString([]byte(":" + p.token))
	req.Header.Set("Authorization", "Basic "+auth)
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

func (p *AzureDevOpsProvider) GetRepository(ctx context.Context, url string) (*Repository, error) {
	// In a real implementation, we would fetch repo details
	return &Repository{
		Name:     p.repo,
		Owner:    p.org,
		URL:      url,
		Provider: "azure",
	}, nil
}

func (p *AzureDevOpsProvider) GetLastModificationDate(ctx context.Context, filePath string) (time.Time, types.Commit, error) {
	// GET /git/repositories/{repositoryId}/commits?searchCriteria.itemPath={filePath}&$top=1
	path := fmt.Sprintf("git/repositories/%s/commits?searchCriteria.itemPath=%s&$top=1&api-version=7.0", p.repo, filePath)

	var result struct {
		Count int `json:"count"`
		Value []struct {
			CommitId string `json:"commitId"`
			Author   struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Comment string `json:"comment"`
		} `json:"value"`
	}

	if err := p.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return time.Time{}, types.Commit{}, err
	}

	if result.Count == 0 {
		return time.Time{}, types.Commit{}, fmt.Errorf("no commits found for file: %s", filePath)
	}

	c := result.Value[0]
	commit := types.Commit{
		Hash:      c.CommitId,
		Author:    c.Author.Name,
		Email:     c.Author.Email,
		Message:   c.Comment,
		Timestamp: c.Author.Date,
	}

	return c.Author.Date, commit, nil
}

func (p *AzureDevOpsProvider) CreateBranch(ctx context.Context, name, baseRef string) error {
	// 1. Get base ref
	// 2. Create ref
	// Simplified: Not implemented fully
	return fmt.Errorf("CreateBranch not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error) {
	// Pushes endpoint
	return "", fmt.Errorf("CreateCommit not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) CreatePullRequest(ctx context.Context, cfg types.PRConfig) (*types.PullRequest, error) {
	// POST /git/repositories/{repositoryId}/pullrequests
	path := fmt.Sprintf("git/repositories/%s/pullrequests?api-version=7.0", p.repo)

	body := map[string]interface{}{
		"sourceRefName": "refs/heads/" + cfg.Branch,
		"targetRefName": "refs/heads/" + cfg.BaseBranch,
		"title":         cfg.Title,
		"description":   cfg.Description,
	}

	var result struct {
		PullRequestId int       `json:"pullRequestId"`
		Url           string    `json:"url"`
		Status        string    `json:"status"`
		CreationDate  time.Time `json:"creationDate"`
	}

	if err := p.doRequest(ctx, "POST", path, body, &result); err != nil {
		return nil, err
	}

	return &types.PullRequest{
		ID:        fmt.Sprintf("%d", result.PullRequestId),
		URL:       result.Url,
		Number:    result.PullRequestId,
		State:     result.Status,
		CreatedAt: result.CreationDate,
	}, nil
}

func (p *AzureDevOpsProvider) UpdatePullRequest(ctx context.Context, id string, updates PRUpdate) error {
	return fmt.Errorf("UpdatePullRequest not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) ClosePullRequest(ctx context.Context, id string, reason string) error {
	return fmt.Errorf("ClosePullRequest not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) AssignPullRequest(ctx context.Context, id string, assignees []string) error {
	return fmt.Errorf("AssignPullRequest not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) RequestReviewers(ctx context.Context, id string, reviewers []string) error {
	return fmt.Errorf("RequestReviewers not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) AddLabels(ctx context.Context, id string, labels []string) error {
	return fmt.Errorf("AddLabels not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) AddComment(ctx context.Context, id string, comment string) error {
	return fmt.Errorf("AddComment not implemented for Azure DevOps")
}
