// Last Recertification: 2025-12-11T20:58:01+01:00
package engine

import (
	"context"
	"fmt"

	"github.com/baldator/iac-recert-engine/internal/assign"
	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/plugin"
	"github.com/baldator/iac-recert-engine/internal/pr"
	"github.com/baldator/iac-recert-engine/internal/provider"
	"github.com/baldator/iac-recert-engine/internal/scan"
	"github.com/baldator/iac-recert-engine/internal/strategy"
	"github.com/baldator/iac-recert-engine/internal/types"
	"go.uber.org/zap"
)

type Engine struct {
	cfg      config.Config
	logger   *zap.Logger
	provider provider.GitProvider
	scanner  *scan.Scanner
	analyzer *scan.HistoryAnalyzer
	checker  *scan.Checker
	strategy strategy.Strategy
	resolver *assign.Resolver
	prGen    *pr.Generator
}

func NewEngine(cfg config.Config, logger *zap.Logger) (*Engine, error) {
	ctx := context.Background()

	// 1. Init Provider
	prov, err := provider.NewProvider(ctx, cfg.Repository, cfg.Auth, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init provider: %w", err)
	}

	// 2. Init Components
	scanner := scan.NewScanner(logger)
	analyzer := scan.NewHistoryAnalyzer(prov, logger)
	checker := scan.NewChecker(logger)

	strat, err := strategy.NewStrategy(cfg.PRStrategy, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init strategy: %w", err)
	}

	pm, err := plugin.NewManager(cfg.Plugins, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init plugin manager: %w", err)
	}

	resolver := assign.NewResolver(cfg.Assignment, pm, logger)
	prGen := pr.NewGenerator(cfg.PRTemplate)

	return &Engine{
		cfg:      cfg,
		logger:   logger,
		provider: prov,
		scanner:  scanner,
		analyzer: analyzer,
		checker:  checker,
		strategy: strat,
		resolver: resolver,
		prGen:    prGen,
	}, nil
}

func (e *Engine) Run(ctx context.Context) error {
	e.logger.Info("starting recertification run")

	// 1. Scan
	repoRoot := "."

	files, err := e.scanner.Scan(repoRoot, e.cfg.Patterns)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	e.logger.Info("scanned files", zap.Int("count", len(files)))

	// 2. Enrich
	enrichedFiles, err := e.analyzer.Enrich(ctx, files, repoRoot)
	if err != nil {
		return fmt.Errorf("history analysis failed: %w", err)
	}

	// 3. Check
	results, err := e.checker.Check(enrichedFiles, e.cfg.Patterns, repoRoot)
	if err != nil {
		return fmt.Errorf("recertification check failed: %w", err)
	}

	// 4. Group
	groups, err := e.strategy.Group(results)
	if err != nil {
		return fmt.Errorf("grouping failed: %w", err)
	}
	e.logger.Info("created groups", zap.Int("count", len(groups)))

	// 5. Process Groups
	for _, group := range groups {
		if err := e.processGroup(ctx, group); err != nil {
			e.logger.Error("failed to process group", zap.String("group_id", group.ID), zap.Error(err))
		}
	}

	e.logger.Info("run completed")
	return nil
}

func (e *Engine) processGroup(ctx context.Context, group types.FileGroup) error {
	// Resolve Assignment
	assignment, err := e.resolver.Resolve(ctx, group)
	if err != nil {
		return fmt.Errorf("assignment resolution failed: %w", err)
	}

	// Generate PR Config
	prCfg, err := e.prGen.Generate(group)
	if err != nil {
		return fmt.Errorf("pr generation failed: %w", err)
	}

	// Apply assignment
	prCfg.Assignees = assignment.Assignees
	prCfg.Reviewers = assignment.Reviewers
	prCfg.BaseBranch = e.cfg.Global.DefaultBaseBranch
	if prCfg.BaseBranch == "" {
		prCfg.BaseBranch = "main" // Default
	}

	if e.cfg.Global.DryRun {
		e.logger.Info("dry run: would create PR",
			zap.String("title", prCfg.Title),
			zap.String("branch", prCfg.Branch),
			zap.Strings("assignees", prCfg.Assignees),
		)
		return nil
	}

	// Create Branch
	err = e.provider.CreateBranch(ctx, prCfg.Branch, prCfg.BaseBranch)
	if err != nil {
		e.logger.Warn("failed to create branch (might exist)", zap.Error(err))
	}

	// Create Commit
	_, err = e.provider.CreateCommit(ctx, prCfg.Branch, "Trigger recertification", nil)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Create PR
	pr, err := e.provider.CreatePullRequest(ctx, prCfg)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	e.logger.Info("created PR", zap.String("url", pr.URL))
	return nil
}
