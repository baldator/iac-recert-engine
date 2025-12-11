// Last Recertification: 2025-12-11T22:33:11+01:00
package strategy

import (
	"testing"

	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPerFileStrategy_Group(t *testing.T) {
	logger := zap.NewNop()
	s := &PerFileStrategy{logger: logger}

	results := []types.RecertCheckResult{
		{
			File:        types.FileInfo{Path: "file1.tf"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file2.tf"},
			NeedsRecert: false,
		},
		{
			File:        types.FileInfo{Path: "file3.tf"},
			NeedsRecert: true,
		},
	}

	groups, err := s.Group(results)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	assert.Equal(t, "file-file1.tf", groups[0].ID)
	assert.Equal(t, "per_file", groups[0].Strategy)
	assert.Len(t, groups[0].Files, 1)

	assert.Equal(t, "file-file3.tf", groups[1].ID)
	assert.Equal(t, "per_file", groups[1].Strategy)
	assert.Len(t, groups[1].Files, 1)
}

func TestPerPatternStrategy_Group(t *testing.T) {
	logger := zap.NewNop()
	s := &PerPatternStrategy{logger: logger}

	results := []types.RecertCheckResult{
		{
			File:        types.FileInfo{Path: "file1.tf"},
			PatternName: "terraform",
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file2.tf"},
			PatternName: "terraform",
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "README.md"},
			PatternName: "markdown",
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file3.tf"},
			PatternName: "terraform",
			NeedsRecert: false,
		},
	}

	groups, err := s.Group(results)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	// Order is not guaranteed, so we need to find the groups
	var tfGroup, mdGroup *types.FileGroup
	for i := range groups {
		if groups[i].ID == "pattern-terraform" {
			tfGroup = &groups[i]
		} else if groups[i].ID == "pattern-markdown" {
			mdGroup = &groups[i]
		}
	}

	require.NotNil(t, tfGroup)
	assert.Equal(t, "per_pattern", tfGroup.Strategy)
	assert.Len(t, tfGroup.Files, 2)

	require.NotNil(t, mdGroup)
	assert.Equal(t, "per_pattern", mdGroup.Strategy)
	assert.Len(t, mdGroup.Files, 1)
}

func TestPerCommitterStrategy_Group(t *testing.T) {
	logger := zap.NewNop()
	s := &PerCommitterStrategy{logger: logger}

	results := []types.RecertCheckResult{
		{
			File:        types.FileInfo{Path: "file1.tf", CommitAuthor: "alice"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file2.tf", CommitAuthor: "bob"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file3.tf", CommitAuthor: "alice"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file4.tf", CommitAuthor: "alice"},
			NeedsRecert: false,
		},
	}

	groups, err := s.Group(results)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	var aliceGroup, bobGroup *types.FileGroup
	for i := range groups {
		if groups[i].ID == "author-alice" {
			aliceGroup = &groups[i]
		} else if groups[i].ID == "author-bob" {
			bobGroup = &groups[i]
		}
	}

	require.NotNil(t, aliceGroup)
	assert.Equal(t, "per_committer", aliceGroup.Strategy)
	assert.Len(t, aliceGroup.Files, 2)

	require.NotNil(t, bobGroup)
	assert.Equal(t, "per_committer", bobGroup.Strategy)
	assert.Len(t, bobGroup.Files, 1)
}

func TestSinglePRStrategy_Group(t *testing.T) {
	logger := zap.NewNop()
	s := &SinglePRStrategy{logger: logger}

	results := []types.RecertCheckResult{
		{
			File:        types.FileInfo{Path: "file1.tf"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file2.tf"},
			NeedsRecert: true,
		},
		{
			File:        types.FileInfo{Path: "file3.tf"},
			NeedsRecert: false,
		},
	}

	groups, err := s.Group(results)
	require.NoError(t, err)
	require.Len(t, groups, 1)

	assert.Equal(t, "all-files", groups[0].ID)
	assert.Equal(t, "single_pr", groups[0].Strategy)
	assert.Len(t, groups[0].Files, 2)
}
