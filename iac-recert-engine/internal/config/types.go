package config

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Version     string           `yaml:"version" validate:"required"`
	Repository  RepositoryConfig `yaml:"repository" validate:"required"`
	Auth        AuthConfig       `yaml:"auth" validate:"required"`
	Global      GlobalConfig     `yaml:"global"`
	Patterns    []Pattern        `yaml:"patterns" validate:"required,dive"`
	PRStrategy  PRStrategyConfig `yaml:"pr_strategy"`
	Assignment  AssignmentConfig `yaml:"assignment"`
	Plugins     PluginConfigs    `yaml:"plugins"`
	Schedule    ScheduleConfig   `yaml:"schedule"`
	PRTemplate  PRTemplateConfig `yaml:"pr_template"`
	Audit       AuditConfig      `yaml:"audit"`
}

type RepositoryConfig struct {
	URL      string `yaml:"url" validate:"required,url"`
	Provider string `yaml:"provider" validate:"required,oneof=github azure gitlab"`
}

type AuthConfig struct {
	Provider string `yaml:"provider" validate:"required,oneof=github azure gitlab"`
	TokenEnv string `yaml:"token_env" validate:"required"`
}

type GlobalConfig struct {
	DryRun            bool   `yaml:"dry_run"`
	VerboseLogging    bool   `yaml:"verbose_logging"`
	MaxConcurrentPRs  int    `yaml:"max_concurrent_prs" validate:"min=1"`
	DefaultBaseBranch string `yaml:"default_base_branch"`
}

type Pattern struct {
	Name                string   `yaml:"name" validate:"required"`
	Description         string   `yaml:"description"`
	Paths               []string `yaml:"paths" validate:"required"`
	Exclude             []string `yaml:"exclude"`
	RecertificationDays int      `yaml:"recertification_days" validate:"required,min=1"`
	Enabled             bool     `yaml:"enabled"`
}

type PRStrategyConfig struct {
	Type           string `yaml:"type" validate:"required,oneof=per_file per_pattern per_committer single_pr plugin"`
	MaxFilesPerPR  int    `yaml:"max_files_per_pr"`
	PluginName     string `yaml:"plugin_name"` // For plugin strategy
}

type AssignmentConfig struct {
	Strategy          string           `yaml:"strategy" validate:"required,oneof=static last_committer plugin composite"`
	Rules             []AssignmentRule `yaml:"rules"`
	FallbackAssignees []string         `yaml:"fallback_assignees"`
}

type AssignmentRule struct {
	Pattern           string   `yaml:"pattern" validate:"required"`
	Strategy          string   `yaml:"strategy" validate:"required,oneof=static last_committer plugin"`
	Plugin            string   `yaml:"plugin"`
	FallbackAssignees []string `yaml:"fallback_assignees"`
}

type PluginConfigs map[string]PluginConfig

type PluginConfig struct {
	Enabled bool              `yaml:"enabled"`
	Type    string            `yaml:"type" validate:"required"`
	Module  string            `yaml:"module" validate:"required"`
	Config  map[string]string `yaml:"config"`
}

type ScheduleConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cron    string `yaml:"cron"`
}

type PRTemplateConfig struct {
	Title              string `yaml:"title" validate:"required"`
	IncludeFileList    bool   `yaml:"include_file_list"`
	IncludeChecklist   bool   `yaml:"include_checklist"`
	CustomInstructions string `yaml:"custom_instructions"`
}

type AuditConfig struct {
	Enabled bool              `yaml:"enabled"`
	Storage string            `yaml:"storage" validate:"required,oneof=file s3"`
	Config  map[string]string `yaml:"config"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
