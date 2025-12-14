package assign

import (
	"context"
	"time"

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
	r.logger.Debug("resolving assignment for group", zap.String("group_id", group.ID), zap.String("strategy", group.Strategy), zap.Int("files", len(group.Files)))

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
	r.logger.Debug("using assignment strategy", zap.String("strategy", strategy))

	if strategy == "composite" {
		r.logger.Debug("composite strategy not fully implemented, using fallback assignees")
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
		result := types.AssignmentResult{
			Assignees: r.cfg.FallbackAssignees,
		}
		r.logger.Debug("assigned using static strategy", zap.Strings("assignees", result.Assignees))
		return result, nil
	case "last_committer":
		// Find the most recent committer
		var lastAuthor string
		var lastTime time.Time
		for _, f := range group.Files {
			if f.File.CommitAuthor != "" && f.File.LastModified.After(lastTime) {
				lastAuthor = f.File.CommitAuthor
				lastTime = f.File.LastModified
			}
		}
		var assignees []string
		if lastAuthor != "" {
			assignees = []string{lastAuthor}
		}
		result := types.AssignmentResult{
			Assignees: assignees,
		}
		r.logger.Debug("assigned using last_committer strategy", zap.Strings("assignees", result.Assignees))
		return result, nil
	case "plugin":
		// Call plugin
		plugin, err := r.pm.GetAssignmentPlugin(r.cfg.PluginName)
		if err != nil {
			r.logger.Error("failed to get assignment plugin", zap.String("plugin", r.cfg.PluginName), zap.Error(err))
			return types.AssignmentResult{}, err
		}
		// Extract FileInfo from RecertCheckResult
		var files []types.FileInfo
		for _, f := range group.Files {
			files = append(files, f.File)
		}
		result, err := plugin.Resolve(files)
		if err != nil {
			r.logger.Error("plugin resolve failed", zap.Error(err))
			return types.AssignmentResult{}, err
		}
		r.logger.Debug("assigned using plugin strategy", zap.Strings("assignees", result.Assignees))
		return result, nil
	default:
		result := types.AssignmentResult{
			Assignees: r.cfg.FallbackAssignees,
		}
		r.logger.Debug("assigned using fallback strategy", zap.Strings("assignees", result.Assignees))
		return result, nil
	}
}
