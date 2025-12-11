// Last Recertification: 2025-12-11T22:33:10+01:00
package config

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Version    string           `yaml:"version" mapstructure:"version" validate:"required"`
	Repository RepositoryConfig `yaml:"repository" mapstructure:"repository" validate:"required"`
	Auth       AuthConfig       `yaml:"auth" mapstructure:"auth" validate:"required"`
	Global     GlobalConfig     `yaml:"global" mapstructure:"global"`
	Patterns   []Pattern        `yaml:"patterns" mapstructure:"patterns" validate:"required,dive"`
	PRStrategy PRStrategyConfig `yaml:"pr_strategy" mapstructure:"pr_strategy"`
	Assignment AssignmentConfig `yaml:"assignment" mapstructure:"assignment"`
	Plugins    PluginConfigs    `yaml:"plugins" mapstructure:"plugins"`
	Schedule   ScheduleConfig   `yaml:"schedule" mapstructure:"schedule"`
	PRTemplate PRTemplateConfig `yaml:"pr_template" mapstructure:"pr_template"`
	Audit      AuditConfig      `yaml:"audit" mapstructure:"audit"`
}

type RepositoryConfig struct {
	URL      string `yaml:"url" mapstructure:"url" validate:"required,url"`
	Provider string `yaml:"provider" mapstructure:"provider" validate:"required,oneof=github azure gitlab"`
}

type AuthConfig struct {
	Provider string `yaml:"provider" mapstructure:"provider" validate:"required,oneof=github azure gitlab"`
	TokenEnv string `yaml:"token_env" mapstructure:"token_env" validate:"required"`
}

type GlobalConfig struct {
	DryRun            bool   `yaml:"dry_run" mapstructure:"dry_run"`
	VerboseLogging    bool   `yaml:"verbose_logging" mapstructure:"verbose_logging"`
	MaxConcurrentPRs  int    `yaml:"max_concurrent_prs" mapstructure:"max_concurrent_prs" validate:"min=1"`
	DefaultBaseBranch string `yaml:"default_base_branch" mapstructure:"default_base_branch"`
}

type Pattern struct {
	Name                string   `yaml:"name" mapstructure:"name" validate:"required"`
	Description         string   `yaml:"description" mapstructure:"description"`
	Paths               []string `yaml:"paths" mapstructure:"paths" validate:"required"`
	Exclude             []string `yaml:"exclude" mapstructure:"exclude"`
	RecertificationDays int      `yaml:"recertification_days" mapstructure:"recertification_days" validate:"required,min=1"`
	Enabled             bool     `yaml:"enabled" mapstructure:"enabled"`
}

type PRStrategyConfig struct {
	Type          string `yaml:"type" mapstructure:"type" validate:"required,oneof=per_file per_pattern per_committer single_pr plugin"`
	MaxFilesPerPR int    `yaml:"max_files_per_pr" mapstructure:"max_files_per_pr"`
	PluginName    string `yaml:"plugin_name" mapstructure:"plugin_name"` // For plugin strategy
}

type AssignmentConfig struct {
	Strategy          string           `yaml:"strategy" mapstructure:"strategy" validate:"required,oneof=static last_committer plugin composite"`
	Rules             []AssignmentRule `yaml:"rules" mapstructure:"rules"`
	FallbackAssignees []string         `yaml:"fallback_assignees" mapstructure:"fallback_assignees"`
}

type AssignmentRule struct {
	Pattern           string   `yaml:"pattern" mapstructure:"pattern" validate:"required"`
	Strategy          string   `yaml:"strategy" mapstructure:"strategy" validate:"required,oneof=static last_committer plugin"`
	Plugin            string   `yaml:"plugin" mapstructure:"plugin"`
	FallbackAssignees []string `yaml:"fallback_assignees" mapstructure:"fallback_assignees"`
}

type PluginConfigs map[string]PluginConfig

type PluginConfig struct {
	Enabled bool              `yaml:"enabled" mapstructure:"enabled"`
	Type    string            `yaml:"type" mapstructure:"type" validate:"required"`
	Module  string            `yaml:"module" mapstructure:"module" validate:"required"`
	Config  map[string]string `yaml:"config" mapstructure:"config"`
}

type ScheduleConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Cron    string `yaml:"cron" mapstructure:"cron"`
}

type PRTemplateConfig struct {
	Title              string `yaml:"title" mapstructure:"title" validate:"required"`
	IncludeFileList    bool   `yaml:"include_file_list" mapstructure:"include_file_list"`
	IncludeChecklist   bool   `yaml:"include_checklist" mapstructure:"include_checklist"`
	CustomInstructions string `yaml:"custom_instructions" mapstructure:"custom_instructions"`
}

type AuditConfig struct {
	Enabled bool              `yaml:"enabled" mapstructure:"enabled"`
	Storage string            `yaml:"storage" mapstructure:"storage" validate:"required,oneof=file s3"`
	Config  map[string]string `yaml:"config" mapstructure:"config"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
