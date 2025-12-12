package assign

import (
	"context"
	"testing"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name            string
		cfg             config.AssignmentConfig
		group           types.FileGroup
		expectedResult  types.AssignmentResult
		expectedError   error
	}{
		{
			name: "static strategy",
			cfg: config.AssignmentConfig{
				Strategy: "static",
				FallbackAssignees: []string{"user1", "user2"},
			},
			group: types.FileGroup{
				Files: []types.RecertCheckResult{
					{
						File: types.FileInfo{
							Path: "file1",
						},
					},
					{
						File: types.FileInfo{
							Path: "file2",
						},
					},
				},
			},
			expectedResult: types.AssignmentResult{
				Assignees: []string{"user1", "user2"},
			},
		},
		{
			name: "last committer strategy",
			cfg: config.AssignmentConfig{
				Strategy: "last_committer",
			},
			group: types.FileGroup{
				Files: []types.RecertCheckResult{
					{
						File: types.FileInfo{
							Path:         "file1",
							LastModified: time.Now().Add(-1 * time.Hour),
							CommitAuthor: "user1",
						},
					},
					{
						File: types.FileInfo{
							Path:         "file2",
							LastModified: time.Now(),
							CommitAuthor: "user2",
						},
					},
				},
			},
			expectedResult: types.AssignmentResult{
				Assignees: []string{"user2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver(tt.cfg, nil, zap.NewNop())
			result, err := r.Resolve(context.Background(), tt.group)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
