package scan

import (
	"context"
	"path/filepath"

	"github.com/baldator/iac-recert-engine/internal/provider"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type HistoryAnalyzer struct {
	provider provider.GitProvider
	logger   *zap.Logger
}

func NewHistoryAnalyzer(provider provider.GitProvider, logger *zap.Logger) *HistoryAnalyzer {
	return &HistoryAnalyzer{
		provider: provider,
		logger:   logger,
	}
}

func (h *HistoryAnalyzer) Enrich(ctx context.Context, files []types.FileInfo, repoRoot string) ([]types.FileInfo, error) {
	var enriched []types.FileInfo

	h.logger.Debug("starting history analysis", zap.Int("files", len(files)), zap.String("repo_root", repoRoot))

	for i, file := range files {
		// Calculate relative path from repo root if needed.
		// We assume repoRoot is the local path where we scanned.
		relPath, err := filepath.Rel(repoRoot, file.Path)
		if err != nil {
			h.logger.Warn("failed to get relative path", zap.String("file", file.Path), zap.Error(err))
			relPath = file.Path
		}
		// Normalize path separators to forward slashes for git APIs
		relPath = filepath.ToSlash(relPath)

		h.logger.Debug("analyzing file history", zap.Int("index", i+1), zap.String("file", relPath))

		ts, commit, err := h.provider.GetLastModificationDate(ctx, relPath)
		if err != nil {
			h.logger.Warn("failed to get last modification date", zap.String("file", relPath), zap.Error(err))
			// We still include the file but with zero time, which will likely trigger recertification or error handling downstream
		} else {
			file.LastModified = ts
			file.CommitHash = commit.Hash
			file.CommitAuthor = commit.Author
			file.CommitEmail = commit.Email
			file.CommitMsg = commit.Message

			h.logger.Debug("file history retrieved",
				zap.String("file", relPath),
				zap.Time("last_modified", ts),
				zap.String("commit_hash", commit.Hash),
				zap.String("author", commit.Author))
		}

		enriched = append(enriched, file)
	}

	h.logger.Debug("history analysis completed", zap.Int("enriched_files", len(enriched)))
	return enriched, nil
}
