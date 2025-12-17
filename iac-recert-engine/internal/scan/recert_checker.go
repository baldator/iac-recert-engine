package scan

import (
	"path/filepath"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type Checker struct {
	logger *zap.Logger
}

func NewChecker(logger *zap.Logger) *Checker {
	return &Checker{logger: logger}
}

func (c *Checker) Check(files []types.FileInfo, patterns []config.Pattern, repoRoot string) ([]types.RecertCheckResult, error) {
	var results []types.RecertCheckResult

	c.logger.Info("checking files for recertification", zap.Int("count", len(files)))
	c.logger.Debug("starting recertification check", zap.String("repo_root", repoRoot), zap.Int("patterns", len(patterns)))

	for i, file := range files {
		// Calculate relative path for matching
		relPath, err := filepath.Rel(repoRoot, file.Path)
		if err != nil {
			c.logger.Warn("failed to get relative path", zap.String("file", file.Path), zap.Error(err))
			continue
		}
		// Normalize for doublestar (forward slashes)
		relPath = filepath.ToSlash(relPath)

		c.logger.Debug("checking file", zap.Int("index", i+1), zap.String("file", relPath), zap.Time("last_modified", file.LastModified))

		matchedPattern, err := FindMatchingPattern(patterns, relPath)
		if err != nil {
			c.logger.Warn("failed to match pattern", zap.String("file", relPath), zap.Error(err))
			continue
		}

		if matchedPattern == nil {
			// File doesn't match any pattern (shouldn't happen if Scanner uses same patterns,
			// but possible if Scanner logic differs or if patterns changed)
			c.logger.Debug("file did not match any pattern", zap.String("file", relPath))
			continue
		}

		// Compute recertification status
		daysSince := int(time.Since(file.LastModified).Hours() / 24)
		threshold := matchedPattern.RecertificationDays
		needsRecert := daysSince >= threshold

		priority := "Low"
		if needsRecert {
			ratio := float64(daysSince) / float64(threshold)
			if ratio > 1.5 {
				priority = "Critical"
			} else if ratio >= 1.0 {
				priority = "High"
			} else if ratio >= 0.8 {
				priority = "Medium"
			}
		} else {
			// Check if approaching threshold
			ratio := float64(daysSince) / float64(threshold)
			if ratio >= 0.8 {
				priority = "Medium"
			}
		}

		nextDueDate := file.LastModified.AddDate(0, 0, threshold)

		result := types.RecertCheckResult{
			File:        file,
			PatternName: matchedPattern.Name,
			DaysSince:   daysSince,
			Threshold:   threshold,
			Priority:    priority,
			NeedsRecert: needsRecert,
			NextDueDate: nextDueDate,
		}

		results = append(results, result)

		c.logger.Debug("file check result",
			zap.String("file", relPath),
			zap.String("pattern", matchedPattern.Name),
			zap.Int("days_since", daysSince),
			zap.Int("threshold", threshold),
			zap.String("priority", priority),
			zap.Bool("needs_recert", needsRecert),
			zap.Time("next_due", nextDueDate))
	}

	c.logger.Info("recertification check completed", zap.Int("results", len(results)))
	return results, nil
}
