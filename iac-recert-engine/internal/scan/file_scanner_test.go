package scan

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScanner_Scan(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "scanner-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []string{
		"main.tf",
		"modules/vpc/main.tf",
		"modules/vpc/variables.tf",
		"README.md",
		"ignored.txt",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		require.NoError(t, err)
		err = os.WriteFile(path, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Initialize git repository
	require.NoError(t, exec.Command("git", "init", tmpDir).Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "add", ".").Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run())

	logger := zap.NewNop()
	scanner := NewScanner(tmpDir, logger)

	tests := []struct {
		name     string
		patterns []config.Pattern
		want     []string
	}{
		{
			name: "match terraform files",
			patterns: []config.Pattern{
				{
					Name:    "terraform",
					Enabled: true,
					Paths:   []string{"**/*.tf"},
				},
			},
			want: []string{
				"main.tf",
				"modules/vpc/main.tf",
				"modules/vpc/variables.tf",
			},
		},
		{
			name: "match markdown files",
			patterns: []config.Pattern{
				{
					Name:    "markdown",
					Enabled: true,
					Paths:   []string{"**/*.md"},
				},
			},
			want: []string{
				"README.md",
			},
		},
		{
			name: "exclude variables.tf",
			patterns: []config.Pattern{
				{
					Name:    "terraform",
					Enabled: true,
					Paths:   []string{"**/*.tf"},
					Exclude: []string{"**/variables.tf"},
				},
			},
			want: []string{
				"main.tf",
				"modules/vpc/main.tf",
			},
		},
		{
			name: "disabled pattern",
			patterns: []config.Pattern{
				{
					Name:    "terraform",
					Enabled: false,
					Paths:   []string{"**/*.tf"},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, scanDir, err := scanner.Scan(tmpDir, tt.patterns)
			require.NoError(t, err)
			defer os.RemoveAll(scanDir)

			var gotPaths []string
			for _, f := range got {
				relPath, err := filepath.Rel(scanDir, f.Path)
				require.NoError(t, err)
				gotPaths = append(gotPaths, filepath.ToSlash(relPath))
			}

			assert.ElementsMatch(t, tt.want, gotPaths)
		})
	}
}
