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
	for _, file := range files {
		// Calculate relative path from repo root if needed.
		// We assume repoRoot is the local path where we scanned.
		relPath, err := filepath.Rel(repoRoot, file.Path)
		if err != nil {
			h.logger.Warn("failed to get relative path", zap.String("file", file.Path), zap.Error(err))
			relPath = file.Path
		}
		// Normalize path separators to forward slashes for git APIs
		relPath = filepath.ToSlash(relPath)

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
		}
		
		enriched = append(enriched, file)
	}
	return enriched, nil
}
