// Last Recertification: 2025-12-11T20:58:01+01:00
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
	var groups []types.FileGroup
	for _, res := range results {
		if !res.NeedsRecert {
			continue
		}
		groups = append(groups, types.FileGroup{
			ID:       fmt.Sprintf("file-%s", res.File.Path),
			Strategy: "per_file",
			Files:    []types.RecertCheckResult{res},
		})
	}
	return groups, nil
}

type PerPatternStrategy struct {
	logger *zap.Logger
}

func (s *PerPatternStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	groups := make(map[string][]types.RecertCheckResult)
	for _, res := range results {
		if !res.NeedsRecert {
			continue
		}
		groups[res.PatternName] = append(groups[res.PatternName], res)
	}

	var fileGroups []types.FileGroup
	for pattern, files := range groups {
		fileGroups = append(fileGroups, types.FileGroup{
			ID:       fmt.Sprintf("pattern-%s", pattern),
			Strategy: "per_pattern",
			Files:    files,
		})
	}
	return fileGroups, nil
}

type PerCommitterStrategy struct {
	logger *zap.Logger
}

func (s *PerCommitterStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	groups := make(map[string][]types.RecertCheckResult)
	for _, res := range results {
		if !res.NeedsRecert {
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
		fileGroups = append(fileGroups, types.FileGroup{
			ID:       fmt.Sprintf("author-%s", author),
			Strategy: "per_committer",
			Files:    files,
		})
	}
	return fileGroups, nil
}

type SinglePRStrategy struct {
	logger *zap.Logger
}

func (s *SinglePRStrategy) Group(results []types.RecertCheckResult) ([]types.FileGroup, error) {
	var files []types.RecertCheckResult
	for _, res := range results {
		if res.NeedsRecert {
			files = append(files, res)
		}
	}
	if len(files) == 0 {
		return nil, nil
	}
	return []types.FileGroup{{
		ID:       "all-files",
		Strategy: "single_pr",
		Files:    files,
	}}, nil
}
