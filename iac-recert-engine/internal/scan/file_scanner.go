package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"
)

type Scanner struct {
	logger  *zap.Logger
	repoURL string
}

func NewScanner(repoURL string, logger *zap.Logger) *Scanner {
	return &Scanner{
		logger:  logger,
		repoURL: repoURL,
	}
}

func (s *Scanner) Scan(root string, patterns []config.Pattern) ([]types.FileInfo, string, error) {
	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "iac-recert-scan-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	s.logger.Debug("cloning repository", zap.String("url", s.repoURL), zap.String("temp_dir", tempDir))

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth", "1", s.repoURL, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup on error
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("failed to clone repository: %w, output: %s", err, string(output))
	}

	s.logger.Debug("repository cloned successfully", zap.String("temp_dir", tempDir))

	// Scan the cloned repository
	var files []types.FileInfo
	seen := make(map[string]bool)

	s.logger.Debug("starting file scan", zap.String("root", root), zap.String("scan_dir", tempDir), zap.Int("patterns", len(patterns)))

	for _, pattern := range patterns {
		if !pattern.Enabled {
			s.logger.Debug("skipping disabled pattern", zap.String("pattern", pattern.Name))
			continue
		}

		s.logger.Debug("processing pattern", zap.String("name", pattern.Name), zap.Strings("paths", pattern.Paths))

		for _, pathPattern := range pattern.Paths {
			// Handle exclusions in the pattern config (if any, though doublestar handles ! in pattern string too,
			// but config has separate Exclude list)

			fullPattern := filepath.Join(tempDir, pathPattern)
			s.logger.Debug("globbing pattern", zap.String("full_pattern", fullPattern))
			matches, err := doublestar.FilepathGlob(fullPattern)
			if err != nil {
				s.logger.Error("failed to glob pattern", zap.String("pattern", fullPattern), zap.Error(err))
				continue
			}

			s.logger.Debug("pattern matched files", zap.String("pattern", pathPattern), zap.Int("matches", len(matches)))

			for _, match := range matches {
				// Check exclusions
				excluded := false
				for _, excludePattern := range pattern.Exclude {
					fullExclude := filepath.Join(tempDir, excludePattern)
					matched, err := doublestar.PathMatch(fullExclude, match)
					if err == nil && matched {
						s.logger.Debug("file excluded", zap.String("file", match), zap.String("exclude_pattern", excludePattern))
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
					s.logger.Debug("skipping duplicate file", zap.String("file", match))
					continue
				}

				info, err := os.Stat(match)
				if err != nil {
					s.logger.Warn("failed to stat file", zap.String("file", match), zap.Error(err))
					continue
				}

				if info.IsDir() {
					s.logger.Debug("skipping directory", zap.String("path", match))
					continue
				}

				files = append(files, types.FileInfo{
					Path: match,
					Size: info.Size(),
				})
				seen[match] = true
				s.logger.Debug("added file to scan results", zap.String("file", match), zap.Int64("size", info.Size()))
			}
		}
	}

	s.logger.Debug("scan completed", zap.Int("total_files", len(files)))
	return files, tempDir, nil
}
