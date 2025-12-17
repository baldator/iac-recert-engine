package scan

import (
	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

// MatchFilePattern checks if a relative file path matches any of the pattern's include paths
// and does not match any of the pattern's exclude paths.
// The path should be relative to the repository root and use forward slashes.
func MatchFilePattern(pattern config.Pattern, relPath string) (bool, error) {
	if !pattern.Enabled {
		return false, nil
	}

	// Check inclusions
	matched := false
	for _, p := range pattern.Paths {
		m, err := doublestar.PathMatch(p, relPath)
		if err != nil {
			return false, err
		}
		if m {
			matched = true
			break
		}
	}

	if !matched {
		return false, nil
	}

	// Check exclusions
	for _, ex := range pattern.Exclude {
		m, err := doublestar.PathMatch(ex, relPath)
		if err != nil {
			return false, err
		}
		if m {
			return false, nil
		}
	}

	return true, nil
}

// FindMatchingPattern returns the first pattern that matches the file path,
// or nil if no pattern matches.
func FindMatchingPattern(patterns []config.Pattern, relPath string) (*config.Pattern, error) {
	for i := range patterns {
		pattern := &patterns[i]
		matched, err := MatchFilePattern(*pattern, relPath)
		if err != nil {
			return nil, err
		}
		if matched {
			return pattern, nil
		}
	}
	return nil, nil
}
