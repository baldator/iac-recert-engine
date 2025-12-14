package servicenow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/baldator/iac-recert-engine/pkg/api"
	"go.uber.org/zap"
)

// ServiceNowPlugin implements assignment based on ServiceNow business applications
type ServiceNowPlugin struct {
	apiURL      string
	username    string
	password    string
	logger      *zap.Logger
	httpClient  *http.Client
	appRegex    *regexp.Regexp
}

// ServiceNowResponse represents the response from ServiceNow API
type ServiceNowResponse struct {
	Result []ServiceNowBusinessApp `json:"result"`
}

// ServiceNowBusinessApp represents a business application in ServiceNow
type ServiceNowBusinessApp struct {
	Name       string `json:"name"`
	SupportedBy struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"supported_by"`
}

// ServiceNowUser represents a user in ServiceNow
type ServiceNowUser struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

// NewServiceNowPlugin creates a new ServiceNow assignment plugin
func NewServiceNowPlugin(logger *zap.Logger) api.AssignmentPlugin {
	return &ServiceNowPlugin{
		logger:     logger,
		httpClient: &http.Client{},
		appRegex:   regexp.MustCompile(`application\s*=\s*["']([^"']+)["']`),
	}
}

func (p *ServiceNowPlugin) Init(config map[string]string) error {
	p.apiURL = config["api_url"]
	p.username = config["username"]
	p.password = config["password"]
	return nil
}

func (p *ServiceNowPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
	// Extract application name from files
	appName := p.extractApplicationName(files)
	if appName == "" {
		p.logger.Warn("no application name found in files")
		return api.AssignmentResult{}, nil
	}

	// Query ServiceNow for the supported by user
	user, err := p.getSupportedByUser(appName)
	if err != nil {
		p.logger.Error("failed to get supported by user", zap.Error(err))
		return api.AssignmentResult{}, err
	}

	return api.AssignmentResult{
		Assignees: []string{user},
	}, nil
}

func (p *ServiceNowPlugin) extractApplicationName(files []api.FileInfo) string {
	for _, file := range files {
		content, err := p.readFile(file.Path)
		if err != nil {
			p.logger.Warn("failed to read file", zap.String("path", file.Path), zap.Error(err))
			continue
		}

		matches := p.appRegex.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func (p *ServiceNowPlugin) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}
	return content.String(), scanner.Err()
}

func (p *ServiceNowPlugin) getSupportedByUser(appName string) (string, error) {
	// Query business application
	url := fmt.Sprintf("%s/api/now/table/cmdb_ci_business_app?sysparm_query=name=%s&sysparm_fields=name,supported_by", p.apiURL, appName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(p.username, p.password)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ServiceNow API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var snResp ServiceNowResponse
	if err := json.Unmarshal(body, &snResp); err != nil {
		return "", err
	}

	if len(snResp.Result) == 0 {
		return "", fmt.Errorf("no business application found with name %s", appName)
	}

	app := snResp.Result[0]
	if app.SupportedBy.Value == "" {
		return "", fmt.Errorf("no supported_by user found for application %s", appName)
	}

	// Get user details
	userURL := fmt.Sprintf("%s/api/now/table/sys_user/%s?sysparm_fields=user_name,email", p.apiURL, app.SupportedBy.Value)
	userReq, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return "", err
	}
	userReq.SetBasicAuth(p.username, p.password)

	userResp, err := p.httpClient.Do(userReq)
	if err != nil {
		return "", err
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ServiceNow user API returned status %d", userResp.StatusCode)
	}

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return "", err
	}

	var userRespStruct struct {
		Result ServiceNowUser `json:"result"`
	}
	if err := json.Unmarshal(userBody, &userRespStruct); err != nil {
		return "", err
	}

	user := userRespStruct.Result
	if user.Email != "" {
		return user.Email, nil
	}
	return user.UserName, nil
}
