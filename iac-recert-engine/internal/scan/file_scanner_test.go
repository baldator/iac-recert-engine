package scan

import (
	"os"
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

	logger := zap.NewNop()
	scanner := NewScanner(logger)

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
				filepath.Join(tmpDir, "main.tf"),
				filepath.Join(tmpDir, "modules/vpc/main.tf"),
				filepath.Join(tmpDir, "modules/vpc/variables.tf"),
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
				filepath.Join(tmpDir, "README.md"),
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
				filepath.Join(tmpDir, "main.tf"),
				filepath.Join(tmpDir, "modules/vpc/main.tf"),
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
			got, err := scanner.Scan(tmpDir, tt.patterns)
			require.NoError(t, err)

			var gotPaths []string
			for _, f := range got {
				gotPaths = append(gotPaths, f.Path)
			}

			assert.ElementsMatch(t, tt.want, gotPaths)
		})
	}
}
