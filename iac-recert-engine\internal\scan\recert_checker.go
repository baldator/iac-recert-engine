// Last Recertification: 2025-12-11T22:51:29+01:00
package scan

import (
	"path/filepath"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/bmatcuk/doublestar/v4"
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

	for _, file := range files {
		// Calculate relative path for matching
		relPath, err := filepath.Rel(repoRoot, file.Path)
		if err != nil {
			c.logger.Warn("failed to get relative path", zap.String("file", file.Path), zap.Error(err))
			continue
		}
		// Normalize for doublestar (forward slashes)
		relPath = filepath.ToSlash(relPath)

		var matchedPattern *config.Pattern
		for i := range patterns {
			pattern := &patterns[i]
			if !pattern.Enabled {
				continue
			}

			matched := false
			for _, p := range pattern.Paths {
				m, err := doublestar.PathMatch(p, relPath)
				if err == nil && m {
					matched = true
					break
				}
			}

			if matched {
				// Check exclusions
				excluded := false
				for _, ex := range pattern.Exclude {
					m, err := doublestar.PathMatch(ex, relPath)
					if err == nil && m {
						excluded = true
						break
					}
				}
				if !excluded {
					matchedPattern = pattern
					break // Found the first matching pattern
				}
			}
		}

		if matchedPattern == nil {
			// File doesn't match any pattern (shouldn't happen if Scanner uses same patterns,
			// but possible if Scanner logic differs or if patterns changed)
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

		results = append(results, types.RecertCheckResult{
			File:        file,
			PatternName: matchedPattern.Name,
			DaysSince:   daysSince,
			Threshold:   threshold,
			Priority:    priority,
			NeedsRecert: needsRecert,
			NextDueDate: nextDueDate,
		})
	}

	return results, nil
}
