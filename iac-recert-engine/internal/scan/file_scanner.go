package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
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

	// Walk the cloned repository and check each file against patterns
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			s.logger.Warn("failed to get relative path", zap.String("file", path), zap.Error(err))
			return nil
		}
		// Normalize to forward slashes for consistent matching
		relPath = filepath.ToSlash(relPath)

		s.logger.Debug("checking file during scan", zap.String("file", relPath))

		// Check if file matches any pattern
		matchedPattern, err := FindMatchingPattern(patterns, relPath)
		if err != nil {
			s.logger.Warn("failed to match pattern", zap.String("file", relPath), zap.Error(err))
			return nil
		}

		if matchedPattern != nil {
			if seen[path] {
				s.logger.Debug("skipping duplicate file", zap.String("file", path))
				return nil
			}

			files = append(files, types.FileInfo{
				Path: path,
				Size: info.Size(),
			})
			seen[path] = true
			s.logger.Debug("added file to scan results", zap.String("file", relPath), zap.String("pattern", matchedPattern.Name), zap.Int64("size", info.Size()))
		}

		return nil
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to walk directory: %w", err)
	}

	s.logger.Debug("scan completed", zap.Int("total_files", len(files)))
	return files, tempDir, nil
}
