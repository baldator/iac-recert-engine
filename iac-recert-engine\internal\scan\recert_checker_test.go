// Last Recertification: 2025-12-11T22:33:11+01:00
package scan

import (
	"fmt"
	"testing"
	"time"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestChecker_Check(t *testing.T) {
	logger := zap.NewNop()
	checker := NewChecker(logger)

	now := time.Now()
	repoRoot := "."

	tests := []struct {
		name     string
		files    []types.FileInfo
		patterns []config.Pattern
		want     []types.RecertCheckResult
	}{
		{
			name: "needs recertification - critical",
			files: []types.FileInfo{
				{
					Path:         "main.tf",
					LastModified: now.AddDate(0, 0, -100), // 100 days ago
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: []types.RecertCheckResult{
				{
					PatternName: "terraform",
					DaysSince:   100,
					Threshold:   60,
					Priority:    "Critical", // 100/60 = 1.66 > 1.5
					NeedsRecert: true,
				},
			},
		},
		{
			name: "needs recertification - high",
			files: []types.FileInfo{
				{
					Path:         "main.tf",
					LastModified: now.AddDate(0, 0, -70), // 70 days ago
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: []types.RecertCheckResult{
				{
					PatternName: "terraform",
					DaysSince:   70,
					Threshold:   60,
					Priority:    "High", // 70/60 = 1.16 > 1.0
					NeedsRecert: true,
				},
			},
		},
		{
			name: "needs recertification - medium",
			files: []types.FileInfo{
				{
					Path:         "main.tf",
					LastModified: now.AddDate(0, 0, -60), // 60 days ago
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: []types.RecertCheckResult{
				{
					PatternName: "terraform",
					DaysSince:   60,
					Threshold:   60,
					Priority:    "High",
					NeedsRecert: true,
				},
			},
		},
		{
			name: "approaching threshold - medium",
			files: []types.FileInfo{
				{
					Path:         "main.tf",
					LastModified: now.AddDate(0, 0, -50), // 50 days ago
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: []types.RecertCheckResult{
				{
					PatternName: "terraform",
					DaysSince:   50,
					Threshold:   60,
					Priority:    "Medium", // 50/60 = 0.83 >= 0.8
					NeedsRecert: false,
				},
			},
		},
		{
			name: "no recertification needed - low",
			files: []types.FileInfo{
				{
					Path:         "main.tf",
					LastModified: now.AddDate(0, 0, -10), // 10 days ago
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: []types.RecertCheckResult{
				{
					PatternName: "terraform",
					DaysSince:   10,
					Threshold:   60,
					Priority:    "Low",
					NeedsRecert: false,
				},
			},
		},
		{
			name: "no match",
			files: []types.FileInfo{
				{
					Path:         "other.txt",
					LastModified: now.AddDate(0, 0, -100),
				},
			},
			patterns: []config.Pattern{
				{
					Name:                "terraform",
					Enabled:             true,
					Paths:               []string{"**/*.tf"},
					RecertificationDays: 60,
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.Check(tt.files, tt.patterns, repoRoot)
			require.NoError(t, err)

			if len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}

			require.Len(t, got, len(tt.want))
			for i, want := range tt.want {
				fmt.Printf("Checking result %d: DaysSince=%d, Threshold=%d, Priority=%s, NeedsRecert=%v\n", i, got[i].DaysSince, got[i].Threshold, got[i].Priority, got[i].NeedsRecert)
				assert.Equal(t, want.PatternName, got[i].PatternName)
				assert.Equal(t, want.DaysSince, got[i].DaysSince)
				assert.Equal(t, want.Threshold, got[i].Threshold)
				assert.Equal(t, want.Priority, got[i].Priority)
				assert.Equal(t, want.NeedsRecert, got[i].NeedsRecert)
			}
		})
	}
}
