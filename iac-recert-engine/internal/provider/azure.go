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

func (p *AzureDevOpsProvider) BranchExists(ctx context.Context, name string) (bool, error) {
	path := fmt.Sprintf("git/repositories/%s/refs?filter=heads/%s&api-version=7.0", p.repo, name)

	var result struct {
		Count int `json:"count"`
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}

	if err := p.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

func (p *AzureDevOpsProvider) CreateBranch(ctx context.Context, name, baseRef string) error {
	// 1. Get base ref SHA
	basePath := fmt.Sprintf("git/repositories/%s/refs?filter=heads/%s&api-version=7.0", p.repo, baseRef)
	var baseResult struct {
		Value []struct {
			ObjectId string `json:"objectId"`
		} `json:"value"`
	}

	if err := p.doRequest(ctx, "GET", basePath, nil, &baseResult); err != nil {
		return err
	}

	if len(baseResult.Value) == 0 {
		return fmt.Errorf("base ref %s not found", baseRef)
	}

	// 2. Create new ref
	path := fmt.Sprintf("git/repositories/%s/refs?api-version=7.0", p.repo)
	body := []map[string]interface{}{
		{
			"name":        "refs/heads/" + name,
			"oldObjectId": "0000000000000000000000000000000000000000",
			"newObjectId": baseResult.Value[0].ObjectId,
		},
	}

	return p.doRequest(ctx, "POST", path, body, nil)
}

func (p *AzureDevOpsProvider) CreateCommit(ctx context.Context, branch, message string, changes []types.Change) (string, error) {
	// Use pushes API to create commit with changes
	path := fmt.Sprintf("git/repositories/%s/pushes?api-version=7.0", p.repo)

	// Get current branch ref
	refPath := fmt.Sprintf("git/repositories/%s/refs?filter=heads/%s&api-version=7.0", p.repo, branch)
	var refResult struct {
		Value []struct {
			ObjectId string `json:"objectId"`
		} `json:"value"`
	}

	if err := p.doRequest(ctx, "GET", refPath, nil, &refResult); err != nil {
		return "", err
	}

	if len(refResult.Value) == 0 {
		return "", fmt.Errorf("branch %s not found", branch)
	}

	// Build commits array
	commits := []map[string]interface{}{
		{
			"comment": message,
			"changes": func() []map[string]interface{} {
				var result []map[string]interface{}
				for _, change := range changes {
					result = append(result, map[string]interface{}{
						"changeType": "add",
						"item": map[string]interface{}{
							"path": change.Path,
						},
						"newContent": map[string]interface{}{
							"content":     change.Content,
							"contentType": "rawtext",
						},
					})
				}
				return result
			}(),
		},
	}

	body := map[string]interface{}{
		"refUpdates": []map[string]interface{}{
			{
				"name":        "refs/heads/" + branch,
				"oldObjectId": refResult.Value[0].ObjectId,
			},
		},
		"commits": commits,
	}

	var result struct {
		PushId int `json:"pushId"`
		Commits []struct {
			CommitId string `json:"commitId"`
		} `json:"commits"`
	}

	if err := p.doRequest(ctx, "POST", path, body, &result); err != nil {
		return "", err
	}

	if len(result.Commits) == 0 {
		return "", fmt.Errorf("no commits created")
	}

	return result.Commits[0].CommitId, nil
}

func (p *AzureDevOpsProvider) PullRequestExists(ctx context.Context, headBranch, baseBranch string) (bool, error) {
	path := fmt.Sprintf("git/repositories/%s/pullrequests?sourceRefName=refs/heads/%s&targetRefName=refs/heads/%s&status=active&api-version=7.0", p.repo, headBranch, baseBranch)

	var result struct {
		Count int `json:"count"`
		Value []struct {
			PullRequestId int `json:"pullRequestId"`
		} `json:"value"`
	}

	if err := p.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return false, err
	}

	return result.Count > 0, nil
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
	path := fmt.Sprintf("git/repositories/%s/pullrequests/%s?api-version=7.0", p.repo, id)

	body := make(map[string]interface{})
	if updates.Title != nil {
		body["title"] = *updates.Title
	}
	if updates.Description != nil {
		body["description"] = *updates.Description
	}
	if updates.State != nil {
		body["status"] = *updates.State
	}

	return p.doRequest(ctx, "PATCH", path, body, nil)
}

func (p *AzureDevOpsProvider) ClosePullRequest(ctx context.Context, id string, reason string) error {
	status := "abandoned" // Azure DevOps uses "abandoned" for closed PRs
	return p.UpdatePullRequest(ctx, id, PRUpdate{State: &status})
}

func (p *AzureDevOpsProvider) AssignPullRequest(ctx context.Context, id string, assignees []string) error {
	// In Azure DevOps, assignment is handled through reviewers
	return p.RequestReviewers(ctx, id, assignees)
}

func (p *AzureDevOpsProvider) RequestReviewers(ctx context.Context, id string, reviewers []string) error {
	path := fmt.Sprintf("git/repositories/%s/pullrequests/%s/reviewers?api-version=7.0", p.repo, id)

	body := make([]map[string]interface{}, len(reviewers))
	for i, reviewer := range reviewers {
		body[i] = map[string]interface{}{
			"uniqueName": reviewer,
			"vote":       0, // 0 = no vote, 10 = approved, -10 = rejected
		}
	}

	return p.doRequest(ctx, "PUT", path, body, nil)
}

func (p *AzureDevOpsProvider) AddLabels(ctx context.Context, id string, labels []string) error {
	return fmt.Errorf("AddLabels not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) AddComment(ctx context.Context, id string, comment string) error {
	path := fmt.Sprintf("git/repositories/%s/pullrequests/%s/threads?api-version=7.0", p.repo, id)

	body := map[string]interface{}{
		"comments": []map[string]interface{}{
			{
				"content": comment,
			},
		},
	}

	return p.doRequest(ctx, "POST", path, body, nil)
}
