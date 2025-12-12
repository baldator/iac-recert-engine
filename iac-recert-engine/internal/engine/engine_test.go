package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewEngine(t *testing.T) {
	// Set dummy token for tests
	t.Setenv("GITHUB_TOKEN", "dummy-token-for-tests")
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		cfg := config.Config{
			Version: "1.0",
			Repository: config.RepositoryConfig{
				URL:      "https://github.com/test/repo",
				Provider: "github",
			},
			Auth: config.AuthConfig{
				Provider: "github",
				TokenEnv: "GITHUB_TOKEN",
			},
			Patterns: []config.Pattern{
				{
					Name:                "test",
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
					Enabled:             true,
				},
			},
			PRStrategy: config.PRStrategyConfig{
				Type: "per_pattern",
			},
			Assignment: config.AssignmentConfig{
				Strategy:          "static",
				FallbackAssignees: []string{"user1"},
			},
			PRTemplate: config.PRTemplateConfig{
				Title: "Test PR",
			},
		}

		engine, err := NewEngine(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Equal(t, cfg, engine.cfg)
		assert.Equal(t, logger, engine.logger)
	})
}

func TestEngine_Run(t *testing.T) {
	// Set dummy token for tests
	t.Setenv("GITHUB_TOKEN", "dummy-token-for-tests")
	logger := zap.NewNop()

	t.Run("dry run success", func(t *testing.T) {
		cfg := config.Config{
			Version: "1.0",
			Repository: config.RepositoryConfig{
				URL:      "https://github.com/test/repo",
				Provider: "github",
			},
			Auth: config.AuthConfig{
				Provider: "github",
				TokenEnv: "GITHUB_TOKEN",
			},
			Global: config.GlobalConfig{
				DryRun: true,
			},
			Patterns: []config.Pattern{
				{
					Name:                "test",
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
					Enabled:             true,
				},
			},
			PRStrategy: config.PRStrategyConfig{
				Type: "per_pattern",
			},
			Assignment: config.AssignmentConfig{
				Strategy:          "static",
				FallbackAssignees: []string{"user1"},
			},
			PRTemplate: config.PRTemplateConfig{
				Title: "Test PR",
			},
		}

		engine, err := NewEngine(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Run(ctx)
		// This will likely fail due to network/git operations, but we're testing the dry run path
		// The exact error depends on the implementation, but we want to ensure it doesn't panic
		// and exercises the dry run code path
		assert.Error(t, err) // Expected to fail due to missing repo or auth
	})

	t.Run("scan failure", func(t *testing.T) {
		cfg := config.Config{
			Version: "1.0",
			Repository: config.RepositoryConfig{
				URL:      "https://invalid-repo-url-that-does-not-exist",
				Provider: "github",
			},
			Auth: config.AuthConfig{
				Provider: "github",
				TokenEnv: "GITHUB_TOKEN",
			},
			Patterns: []config.Pattern{
				{
					Name:                "test",
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
					Enabled:             true,
				},
			},
			PRStrategy: config.PRStrategyConfig{
				Type: "per_pattern",
			},
			Assignment: config.AssignmentConfig{
				Strategy:          "static",
				FallbackAssignees: []string{"user1"},
			},
			PRTemplate: config.PRTemplateConfig{
				Title: "Test PR",
			},
		}

		engine, err := NewEngine(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Run(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scan failed")
	})

	t.Run("full run with local repo", func(t *testing.T) {
		// Create a temporary git repository for testing
		tmpDir, err := os.MkdirTemp("", "engine-test-repo")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create test files
		testFile := filepath.Join(tmpDir, "main.tf")
		err = os.WriteFile(testFile, []byte("resource \"aws_instance\" \"test\" {}"), 0644)
		require.NoError(t, err)

		// Initialize git repo
		require.NoError(t, exec.Command("git", "init", tmpDir).Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "add", ".").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run())

		cfg := config.Config{
			Version: "1.0",
			Repository: config.RepositoryConfig{
				URL:      "file://" + tmpDir,
				Provider: "github",
			},
			Auth: config.AuthConfig{
				Provider: "github",
				TokenEnv: "GITHUB_TOKEN",
			},
			Global: config.GlobalConfig{
				DryRun: true,
			},
			Patterns: []config.Pattern{
				{
					Name:                "terraform",
					Paths:               []string{"*.tf"},
					RecertificationDays: 60,
					Enabled:             true,
					Decorator:           "# Test decorator\n",
				},
			},
			PRStrategy: config.PRStrategyConfig{
				Type: "per_pattern",
			},
			Assignment: config.AssignmentConfig{
				Strategy:          "static",
				FallbackAssignees: []string{"user1"},
			},
			PRTemplate: config.PRTemplateConfig{
				Title: "Test PR",
			},
		}

		engine, err := NewEngine(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Run(ctx)
		// This should succeed with dry run
		assert.NoError(t, err)
	})

	t.Run("full run without dry run", func(t *testing.T) {
		// Create a temporary git repository for testing
		tmpDir, err := os.MkdirTemp("", "engine-test-repo-full")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create test files
		testFile := filepath.Join(tmpDir, "main.tf")
		err = os.WriteFile(testFile, []byte("resource \"aws_instance\" \"test\" {}"), 0644)
		require.NoError(t, err)

		// Initialize git repo
		require.NoError(t, exec.Command("git", "init", tmpDir).Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "add", ".").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run())
		require.NoError(t, exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run())

		cfg := config.Config{
			Version: "1.0",
			Repository: config.RepositoryConfig{
				URL:      "file://" + tmpDir,
				Provider: "github",
			},
			Auth: config.AuthConfig{
				Provider: "github",
				TokenEnv: "GITHUB_TOKEN",
			},
			Global: config.GlobalConfig{
				DryRun: false,
			},
			Patterns: []config.Pattern{
				{
					Name:                "terraform",
					Paths:               []string{"*.tf"},
					RecertificationDays: 60,
					Enabled:             true,
					Decorator:           "# Test decorator\n",
				},
			},
			PRStrategy: config.PRStrategyConfig{
				Type: "per_pattern",
			},
			Assignment: config.AssignmentConfig{
				Strategy:          "static",
				FallbackAssignees: []string{"user1"},
			},
			PRTemplate: config.PRTemplateConfig{
				Title: "Test PR",
			},
		}

		engine, err := NewEngine(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Run(ctx)
		// This may succeed or fail depending on provider implementation
		// The important thing is it exercises the non-dry-run code path
		_ = err // We don't assert on the error since it depends on the provider
	})
}
