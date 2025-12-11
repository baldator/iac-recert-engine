// Last Recertification: 2025-12-11T22:49:52+01:00
package plugin

import (
	"fmt"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
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

		// In a real implementation, we would load plugins dynamically or from a registry.
		// Here we can support built-in "plugins" or mock them.
		// For now, we just log that we found a plugin config.
		logger.Info("loading plugin", zap.String("name", name), zap.String("type", cfg.Type))
		
		// TODO: Implement actual plugin loading logic.
		// For now, we can't load external Go plugins easily without compiling them.
		// We might support internal implementations registered by name.
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
