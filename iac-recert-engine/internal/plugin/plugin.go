package plugin

import (
	"fmt"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/baldator/iac-recert-engine/pkg/api"
	"github.com/baldator/iac-recert-servicenow-plugin/servicenow"
	"go.uber.org/zap"
)

type PluginType string

const (
	PluginTypeAssignment PluginType = "assignment"
	PluginTypeFilter     PluginType = "filter"
)

type Plugin interface {
	Init(config map[string]string) error
}

type AssignmentPlugin interface {
	Plugin
	Resolve(files []types.FileInfo) (types.AssignmentResult, error)
}

// assignmentPluginWrapper wraps an api.AssignmentPlugin to implement the internal AssignmentPlugin interface
type assignmentPluginWrapper struct {
	apiPlugin api.AssignmentPlugin
}

func (w *assignmentPluginWrapper) Init(config map[string]string) error {
	return w.apiPlugin.Init(config)
}

func (w *assignmentPluginWrapper) Resolve(files []types.FileInfo) (types.AssignmentResult, error) {
	// Convert internal FileInfo to api.FileInfo
	var apiFiles []api.FileInfo
	for _, f := range files {
		apiFiles = append(apiFiles, api.FileInfo{
			Path:         f.Path,
			Size:         f.Size,
			LastModified: f.LastModified.Format("2006-01-02T15:04:05Z07:00"),
			CommitHash:   f.CommitHash,
			CommitAuthor: f.CommitAuthor,
			CommitEmail:  f.CommitEmail,
			CommitMsg:    f.CommitMsg,
		})
	}

	// Call the API plugin
	result, err := w.apiPlugin.Resolve(apiFiles)
	if err != nil {
		return types.AssignmentResult{}, err
	}

	// Convert back to internal AssignmentResult
	return types.AssignmentResult{
		Assignees: result.Assignees,
		Reviewers: result.Reviewers,
		Team:      result.Team,
		Priority:  result.Priority,
	}, nil
}

type Manager struct {
	plugins map[string]Plugin
	logger  *zap.Logger
}

func NewManager(configs config.PluginConfigs, logger *zap.Logger) (*Manager, error) {
	m := &Manager{
		plugins: make(map[string]Plugin),
		logger:  logger,
	}

	for name, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		logger.Info("loading plugin", zap.String("name", name), zap.String("type", cfg.Type), zap.String("module", cfg.Module))

		var plugin Plugin
		switch cfg.Module {
		case "servicenow":
			if cfg.Type != "assignment" {
				return nil, fmt.Errorf("servicenow plugin must be of type assignment")
			}
			apiPlugin := servicenow.NewServiceNowPlugin(logger)
			plugin = &assignmentPluginWrapper{apiPlugin: apiPlugin}
		default:
			return nil, fmt.Errorf("unknown plugin module: %s", cfg.Module)
		}

		if err := plugin.Init(cfg.Config); err != nil {
			return nil, fmt.Errorf("failed to init plugin %s: %w", name, err)
		}

		m.plugins[name] = plugin
	}

	return m, nil
}

func (m *Manager) GetAssignmentPlugin(name string) (AssignmentPlugin, error) {
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	ap, ok := p.(AssignmentPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s is not an assignment plugin", name)
	}
	return ap, nil
}
