package strategy

import (
	"fmt"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type Strategy interface {
	Group(results []types.RecertCheckResult) ([]types.FileGroup, error)
}

func NewStrategy(cfg config.PRStrategyConfig, logger *zap.Logger) (Strategy, error) {
	switch cfg.Type {
	case "per_file":
		return &PerFileStrategy{logger: logger}, nil
	case "per_pattern":
		return &PerPatternStrategy{logger: logger}, nil
	case "per_committer":
		return &PerCommitterStrategy{logger: logger}, nil
	case "single_pr":
		return &SinglePRStrategy{logger: logger}, nil
	case "plugin":
		return nil, fmt.Errorf("plugin strategy not implemented yet")
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", cfg.Type)
	}
}

type PerFileStrategy struct {
	logger *zap.Logger
}

func (s *PerFileStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	s.logger.Debug("grouping files individually", zap.Int("total_results", len(results)))

	var groups []types.FileGroup
	for _, res := range results {
		if !res.NeedsRecert {
			s.logger.Debug("skipping file that doesn't need recertification", zap.String("file", res.File.Path))
			continue
		}
		group := types.FileGroup{
			ID:       fmt.Sprintf("file-%s", res.File.Path),
			Strategy: "per_file",
			Files:    []types.RecertCheckResult{res},
		}
		groups = append(groups, group)
		s.logger.Debug("created file group", zap.String("file", res.File.Path), zap.String("group_id", group.ID))
	}

	s.logger.Debug("per-file grouping completed", zap.Int("groups_created", len(groups)))
	return groups, nil
}

type PerPatternStrategy struct {
	logger *zap.Logger
}

func (s *PerPatternStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	s.logger.Debug("grouping files by pattern", zap.Int("total_results", len(results)))

	groups := make(map[string][]types.RecertCheckResult)
	for _, res := range results {
		if !res.NeedsRecert {
			s.logger.Debug("skipping file that doesn't need recertification", zap.String("file", res.File.Path))
			continue
		}
		groups[res.PatternName] = append(groups[res.PatternName], res)
	}

	var fileGroups []types.FileGroup
	for pattern, files := range groups {
		group := types.FileGroup{
			ID:       fmt.Sprintf("pattern-%s", pattern),
			Strategy: "per_pattern",
			Files:    files,
		}
		fileGroups = append(fileGroups, group)
		s.logger.Debug("created pattern group", zap.String("pattern", pattern), zap.Int("files", len(files)), zap.String("group_id", group.ID))
	}

	s.logger.Debug("pattern grouping completed", zap.Int("groups_created", len(fileGroups)))
	return fileGroups, nil
}

type PerCommitterStrategy struct {
	logger *zap.Logger
}

func (s *PerCommitterStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	s.logger.Debug("grouping files by committer", zap.Int("total_results", len(results)))

	groups := make(map[string][]types.RecertCheckResult)
	for _, res := range results {
		if !res.NeedsRecert {
			s.logger.Debug("skipping file that doesn't need recertification", zap.String("file", res.File.Path))
			continue
		}
		author := res.File.CommitAuthor
		if author == "" {
			author = "unknown"
		}
		groups[author] = append(groups[author], res)
	}

	var fileGroups []types.FileGroup
	for author, files := range groups {
		group := types.FileGroup{
			ID:       fmt.Sprintf("author-%s", author),
			Strategy: "per_committer",
			Files:    files,
		}
		fileGroups = append(fileGroups, group)
		s.logger.Debug("created committer group", zap.String("author", author), zap.Int("files", len(files)), zap.String("group_id", group.ID))
	}

	s.logger.Debug("committer grouping completed", zap.Int("groups_created", len(fileGroups)))
	return fileGroups, nil
}

type SinglePRStrategy struct {
	logger *zap.Logger
}

func (s *SinglePRStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	s.logger.Debug("grouping all files into single PR", zap.Int("total_results", len(results)))

	var files []types.RecertCheckResult
	for _, res := range results {
		if res.NeedsRecert {
			files = append(files, res)
		} else {
			s.logger.Debug("skipping file that doesn't need recertification", zap.String("file", res.File.Path))
		}
	}

	if len(files) == 0 {
		s.logger.Debug("no files need recertification, no groups created")
		return nil, nil
	}

	group := types.FileGroup{
		ID:       "all-files",
		Strategy: "single_pr",
		Files:    files,
	}

	s.logger.Debug("created single PR group", zap.Int("files", len(files)), zap.String("group_id", group.ID))
	return []types.FileGroup{group}, nil
}
