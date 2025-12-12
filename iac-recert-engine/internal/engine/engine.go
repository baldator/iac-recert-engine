package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
	scanner := scan.NewScanner(cfg.Repository.URL, logger)
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
	e.logger.Debug("run configuration", zap.Bool("dry_run", e.cfg.Global.DryRun), zap.Bool("verbose_logging", e.cfg.Global.VerboseLogging))

	// 1. Scan
	repoRoot := "."
	e.logger.Debug("starting file scan phase", zap.String("repo_root", repoRoot))

	files, scanDir, err := e.scanner.Scan(repoRoot, e.cfg.Patterns)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	defer func() {
		if scanDir != "" {
			if err := os.RemoveAll(scanDir); err != nil {
				e.logger.Warn("failed to cleanup scan dir", zap.String("dir", scanDir), zap.Error(err))
			}
		}
	}()
	e.logger.Info("scanned files", zap.Int("count", len(files)), zap.String("scan_dir", scanDir))

	// 2. Enrich
	e.logger.Debug("starting history analysis phase")
	enrichedFiles, err := e.analyzer.Enrich(ctx, files, scanDir)
	if err != nil {
		return fmt.Errorf("history analysis failed: %w", err)
	}
	e.logger.Debug("history analysis completed", zap.Int("enriched_files", len(enrichedFiles)))

	// 3. Check
	e.logger.Debug("starting recertification check phase")
	results, err := e.checker.Check(enrichedFiles, e.cfg.Patterns, scanDir)
	if err != nil {
		return fmt.Errorf("recertification check failed: %w", err)
	}
	e.logger.Debug("recertification check completed", zap.Int("check_results", len(results)))

	// 4. Group
	e.logger.Debug("starting grouping phase")
	groups, err := e.strategy.Group(results)
	if err != nil {
		return fmt.Errorf("grouping failed: %w", err)
	}
	e.logger.Info("created groups", zap.Int("count", len(groups)))

	// 5. Process Groups
	e.logger.Debug("starting group processing phase", zap.Int("groups", len(groups)))
	processed := 0
	failed := 0
	for _, group := range groups {
		if err := e.processGroup(ctx, group, scanDir, e.cfg.Patterns); err != nil {
			e.logger.Error("failed to process group", zap.String("group_id", group.ID), zap.Error(err))
			failed++
		} else {
			processed++
		}
	}

	e.logger.Info("run completed", zap.Int("groups_processed", processed), zap.Int("groups_failed", failed))
	return nil
}

func (e *Engine) processGroup(ctx context.Context, group types.FileGroup, scanDir string, patterns []config.Pattern) error {
	e.logger.Debug("processing group", zap.String("group_id", group.ID), zap.String("strategy", group.Strategy), zap.Int("files", len(group.Files)))

	// Resolve Assignment
	e.logger.Debug("resolving assignment for group", zap.String("group_id", group.ID))
	assignment, err := e.resolver.Resolve(ctx, group)
	if err != nil {
		return fmt.Errorf("assignment resolution failed: %w", err)
	}

	// Generate PR Config
	e.logger.Debug("generating PR configuration", zap.String("group_id", group.ID))
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

	e.logger.Debug("PR configuration prepared",
		zap.String("group_id", group.ID),
		zap.String("title", prCfg.Title),
		zap.String("branch", prCfg.Branch),
		zap.String("base_branch", prCfg.BaseBranch),
		zap.Strings("assignees", prCfg.Assignees),
		zap.Strings("reviewers", prCfg.Reviewers))

	if e.cfg.Global.DryRun {
		e.logger.Info("dry run: would create PR",
			zap.String("title", prCfg.Title),
			zap.String("branch", prCfg.Branch),
			zap.Strings("assignees", prCfg.Assignees),
		)
		return nil
	}

	// Generate changes with decorators
	var changes []types.Change
	for _, res := range group.Files {
		// Find the pattern
		var pattern *config.Pattern
		for i := range patterns {
			if patterns[i].Name == res.PatternName {
				pattern = &patterns[i]
				break
			}
		}
		if pattern == nil || pattern.Decorator == "" {
			continue
		}

		// Read file content
		relPath, err := filepath.Rel(scanDir, res.File.Path)
		if err != nil {
			e.logger.Warn("failed to get relative path for change", zap.String("file", res.File.Path), zap.Error(err))
			continue
		}
		contentBytes, err := os.ReadFile(res.File.Path)
		if err != nil {
			e.logger.Warn("failed to read file for change", zap.String("file", res.File.Path), zap.Error(err))
			continue
		}
		content := string(contentBytes)

		// Remove existing decorator
		existingDecorator := regexp.MustCompile(`^` + regexp.QuoteMeta(pattern.Decorator) + `.*\n?`)
		content = existingDecorator.ReplaceAllString(content, "")

		// Apply new decorator
		decorator := strings.ReplaceAll(pattern.Decorator, "{timestamp}", time.Now().Format(time.RFC3339))
		content = decorator + content

		changes = append(changes, types.Change{
			Path:    relPath,
			Content: content,
		})
		e.logger.Debug("prepared change for file", zap.String("file", relPath), zap.String("decorator", pattern.Decorator))
	}

	// Check if branch already exists
	e.logger.Debug("checking if branch already exists", zap.String("branch", prCfg.Branch))
	branchExists, err := e.provider.BranchExists(ctx, prCfg.Branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}
	if branchExists {
		e.logger.Info("branch already exists, skipping creation", zap.String("branch", prCfg.Branch), zap.String("group_id", group.ID))
	} else {
		// Create Branch
		e.logger.Debug("creating branch", zap.String("branch", prCfg.Branch), zap.String("base", prCfg.BaseBranch))
		err = e.provider.CreateBranch(ctx, prCfg.Branch, prCfg.BaseBranch)
		if err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
		e.logger.Debug("branch created successfully", zap.String("branch", prCfg.Branch))
	}

	// Create Commit
	e.logger.Debug("creating commit", zap.String("branch", prCfg.Branch), zap.Int("changes", len(changes)))
	_, err = e.provider.CreateCommit(ctx, prCfg.Branch, "Trigger recertification", changes)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	e.logger.Debug("commit created successfully", zap.String("branch", prCfg.Branch))

	// Check if PR already exists
	e.logger.Debug("checking if pull request already exists", zap.String("branch", prCfg.Branch), zap.String("base_branch", prCfg.BaseBranch))
	prExists, err := e.provider.PullRequestExists(ctx, prCfg.Branch, prCfg.BaseBranch)
	if err != nil {
		return fmt.Errorf("failed to check if PR exists: %w", err)
	}
	if prExists {
		e.logger.Info("PR already exists for branch, skipping creation", zap.String("branch", prCfg.Branch), zap.String("group_id", group.ID))
		return nil
	}

	// Create PR
	e.logger.Debug("creating pull request", zap.String("title", prCfg.Title), zap.String("branch", prCfg.Branch))
	pr, err := e.provider.CreatePullRequest(ctx, prCfg)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	e.logger.Info("created PR", zap.String("url", pr.URL), zap.String("group_id", group.ID))
	return nil
}
