// Last Recertification: 2025-12-11T22:33:10+01:00
package pr

import (
	"fmt"
	"strings"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/types"
)

type Generator struct {
	cfg config.PRTemplateConfig
}

func NewGenerator(cfg config.PRTemplateConfig) *Generator {
	return &Generator{cfg: cfg}
}

func (g *Generator) Generate(group types.FileGroup) (types.PRConfig, error) {
	// 1. Generate Title
	title := g.cfg.Title
	// Simple replacement
	title = strings.ReplaceAll(title, "{pattern_name}", getPatternName(group))
	title = strings.ReplaceAll(title, "{file_count}", fmt.Sprintf("%d", len(group.Files)))

	// 2. Generate Description
	var sb strings.Builder
	sb.WriteString("## Recertification Required\n\n")
	sb.WriteString("The following infrastructure files are due for recertification.\n\n")

	if g.cfg.IncludeFileList {
		sb.WriteString("### Files\n\n")
		sb.WriteString("| Path | Last Modified | Author | Priority |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, f := range group.Files {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				f.File.Path,
				f.File.LastModified.Format("2006-01-02"),
				f.File.CommitAuthor,
				f.Priority,
			))
		}
		sb.WriteString("\n")
	}

	if g.cfg.IncludeChecklist {
		sb.WriteString("### Checklist\n\n")
		sb.WriteString("- [ ] Reviewed configuration for security compliance\n")
		sb.WriteString("- [ ] Verified ownership and necessity\n")
		sb.WriteString("- [ ] Approved for recertification\n\n")
	}

	if g.cfg.CustomInstructions != "" {
		sb.WriteString("### Instructions\n\n")
		sb.WriteString(g.cfg.CustomInstructions)
		sb.WriteString("\n")
	}

	// 3. Build PRConfig
	// Note: Branch name generation should probably happen here or in Engine.
	// Spec says: branch recert/{pattern}/{file-hash} etc.
	// Let's assume Engine handles branch naming or we do it here.
	// Let's do it here based on group ID.
	branchName := fmt.Sprintf("recert/%s", group.ID)

	return types.PRConfig{
		Title:       title,
		Description: sb.String(),
		Branch:      branchName,
		// BaseBranch, Assignees, Reviewers will be filled by Engine/Resolver
	}, nil
}

func getPatternName(group types.FileGroup) string {
	// Try to extract from ID or use first file's pattern
	if len(group.Files) > 0 {
		return group.Files[0].PatternName
	}
	return "unknown"
}
