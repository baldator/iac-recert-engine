// Last Recertification: 2025-12-11T20:58:01+01:00
package scan

import (
	"os"
	"path/filepath"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"
)

type Scanner struct {
	logger *zap.Logger
}

func NewScanner(logger *zap.Logger) *Scanner {
	return &Scanner{logger: logger}
}

func (s *Scanner) Scan(root string, patterns []config.Pattern) ([]types.FileInfo, error) {
	var files []types.FileInfo
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if !pattern.Enabled {
			continue
		}

		for _, pathPattern := range pattern.Paths {
			// Handle exclusions in the pattern config (if any, though doublestar handles ! in pattern string too,
			// but config has separate Exclude list)

			fullPattern := filepath.Join(root, pathPattern)
			matches, err := doublestar.FilepathGlob(fullPattern)
			if err != nil {
				s.logger.Error("failed to glob pattern", zap.String("pattern", fullPattern), zap.Error(err))
				continue
			}

			for _, match := range matches {
				// Check exclusions
				excluded := false
				for _, excludePattern := range pattern.Exclude {
					fullExclude := filepath.Join(root, excludePattern)
					matched, err := doublestar.PathMatch(fullExclude, match)
					if err == nil && matched {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}

				// Check if already seen to avoid duplicates across patterns (or maybe we want them?
				// Spec says "Validation to avoid files appearing in multiple groups", but here we just list them.
				// Let's deduplicate by path for now, or maybe we return a map of file -> pattern?
				// The prompt says "Return []FileInfo".
				if seen[match] {
					continue
				}

				info, err := os.Stat(match)
				if err != nil {
					s.logger.Warn("failed to stat file", zap.String("file", match), zap.Error(err))
					continue
				}

				if info.IsDir() {
					continue
				}

				files = append(files, types.FileInfo{
					Path: match,
					Size: info.Size(),
				})
				seen[match] = true
			}
		}
	}

	return files, nil
}
