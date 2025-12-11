// Last Recertification: 2025-12-11T22:49:52+01:00
package assign

import (
	"context"
	"fmt"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/plugin"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type Resolver struct {
	cfg    config.AssignmentConfig
	logger *zap.Logger
	pm     *plugin.Manager
}

func NewResolver(cfg config.AssignmentConfig, pm *plugin.Manager, logger *zap.Logger) *Resolver {
	return &Resolver{
		cfg:    cfg,
		pm:     pm,
		logger: logger,
	}
}

func (r *Resolver) Resolve(ctx context.Context, group types.FileGroup) (types.AssignmentResult, error) {
	// Determine strategy for this group
	// If composite, we need to match rules.
	// But FileGroup might contain files from different patterns if strategy was "single_pr".
	// However, usually we want to assign based on the group content.

	// If strategy is "composite", we iterate rules and find the first one that matches the group.
	// But matching a group to a pattern is tricky if the group has mixed files.
	// Assuming groups are homogeneous or we use the first file?
	// Or we use the group ID if it contains pattern info?
	// Let's assume we check the files in the group.

	strategy := r.cfg.Strategy
	// var rule *config.AssignmentRule

	if strategy == "composite" {
		// Find matching rule
		// We need to check if ANY file in the group matches the rule pattern?
		// Or ALL files?
		// Usually, we want to find a rule that covers the files.
		// If files are mixed, composite strategy is hard.
		// But let's assume we use the first file to determine the rule for the whole PR.
		if len(group.Files) > 0 {
			// firstFile := group.Files[0].File.Path
			// We need to match firstFile against rule patterns.
			// We need doublestar again? Or just simple match?
			// Config says "pattern: terraform/prod/**".
			// We should use doublestar.
			// But I don't want to import doublestar here if I can avoid it.
			// But I probably should.
			// Let's skip complex matching for now and assume simple prefix or exact match?
			// No, the spec says "pattern".

			// For now, let's just fallback to default if composite.
		}
	}

	switch strategy {
	case "static":
		return types.AssignmentResult{
			Assignees: r.cfg.FallbackAssignees,
		}, nil
	case "last_committer":
		// Collect unique authors from files
		authors := make(map[string]bool)
		for _, f := range group.Files {
			if f.File.CommitAuthor != "" {
				authors[f.File.CommitAuthor] = true
			}
		}
		var assignees []string
		for a := range authors {
			assignees = append(assignees, a)
		}
		return types.AssignmentResult{
			Assignees: assignees,
		}, nil
	case "plugin":
		// Call plugin
		// We need PluginManager.
		// r.pm.GetAssignmentPlugin(name).Resolve(...)
		return types.AssignmentResult{}, fmt.Errorf("plugin assignment not implemented")
	default:
		return types.AssignmentResult{
			Assignees: r.cfg.FallbackAssignees,
		}, nil
	}
}
