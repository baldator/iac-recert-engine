package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Version: "1.0",
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "github",
				},
				Auth: AuthConfig{
					Provider: "github",
					TokenEnv: "GITHUB_TOKEN",
				},
				Global: GlobalConfig{
					MaxConcurrentPRs: 1,
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 90,
					},
				},
				PRStrategy: PRStrategyConfig{
					Type: "per_file",
				},
				Assignment: AssignmentConfig{
					Strategy: "static",
				},
				PRTemplate: PRTemplateConfig{
					Title: "Recertification",
				},
				Audit: AuditConfig{
					Storage: "file",
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: Config{
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "github",
				},
				Auth: AuthConfig{
					Provider: "github",
					TokenEnv: "GITHUB_TOKEN",
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 90,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid repository provider",
			config: Config{
				Version: "1.0",
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "invalid",
				},
				Auth: AuthConfig{
					Provider: "github",
					TokenEnv: "GITHUB_TOKEN",
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 90,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid auth provider",
			config: Config{
				Version: "1.0",
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "github",
				},
				Auth: AuthConfig{
					Provider: "invalid",
					TokenEnv: "GITHUB_TOKEN",
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 90,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid pattern recertification days",
			config: Config{
				Version: "1.0",
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "github",
				},
				Auth: AuthConfig{
					Provider: "github",
					TokenEnv: "GITHUB_TOKEN",
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid pr strategy type",
			config: Config{
				Version: "1.0",
				Repository: RepositoryConfig{
					URL:      "https://github.com/org/repo",
					Provider: "github",
				},
				Auth: AuthConfig{
					Provider: "github",
					TokenEnv: "GITHUB_TOKEN",
				},
				Patterns: []Pattern{
					{
						Name:                "terraform",
						Paths:               []string{"**/*.tf"},
						RecertificationDays: 90,
					},
				},
				PRStrategy: PRStrategyConfig{
					Type: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
