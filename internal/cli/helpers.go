package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/marstid/nuc/pkg/nucleus"
)

// expandGroupGlobs resolves glob patterns in group names.
// If no value contains glob characters (* ? [), returns the input unchanged (no API call).
// Otherwise fetches all groups and returns matching names.
// Unlike filepath.Match, the '*' wildcard matches any character including '/'.
func expandGroupGlobs(client *nucleus.Client, projectID string, patterns []string) ([]string, error) {
	hasGlob := false
	for _, p := range patterns {
		if strings.ContainsAny(p, "*?") {
			hasGlob = true
			break
		}
	}
	if !hasGlob {
		return patterns, nil
	}

	ctx := context.Background()
	groups, err := client.ListAssetGroups(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing groups for pattern expansion: %w", err)
	}

	seen := make(map[string]bool)
	var matched []string
	for _, g := range groups {
		for _, pattern := range patterns {
			if strings.ContainsAny(pattern, "*?") {
				if globMatch(pattern, g.Name) && !seen[g.Name] {
					seen[g.Name] = true
					matched = append(matched, g.Name)
				}
			} else {
				// Literal value — keep as-is without requiring it to exist.
				if !seen[pattern] {
					seen[pattern] = true
					matched = append(matched, pattern)
				}
			}
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no asset groups matched pattern(s): %s", strings.Join(patterns, ", "))
	}

	return dedupChildren(matched), nil
}

// dedupChildren removes groups that are sub-paths of other matched groups.
// For example, if both "/teams/team-euc" and "/teams/team-euc/container" match,
// only "/teams/team-euc" is kept. This prevents sending too many groups to the
// Nucleus API which silently returns empty results when the array is too large.
func dedupChildren(groups []string) []string {
	sort.Strings(groups) // sort so parents come before children
	var result []string
	for _, g := range groups {
		isChild := false
		for _, parent := range result {
			if strings.HasPrefix(g, parent+"/") {
				isChild = true
				break
			}
		}
		if !isChild {
			result = append(result, g)
		}
	}
	return result
}

// globMatch performs simple glob matching where '*' matches any sequence of characters
// (including '/') and '?' matches exactly one character (including '/').
// This differs from filepath.Match which treats '/' as a boundary.
func globMatch(pattern, name string) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '*':
			// Consume consecutive stars.
			for len(pattern) > 0 && pattern[0] == '*' {
				pattern = pattern[1:]
			}
			// Trailing star matches everything.
			if len(pattern) == 0 {
				return true
			}
			// Try matching the rest of the pattern at every position.
			for i := 0; i <= len(name); i++ {
				if globMatch(pattern, name[i:]) {
					return true
				}
			}
			return false
		case '?':
			if len(name) == 0 {
				return false
			}
			name = name[1:]
			pattern = pattern[1:]
		default:
			if len(name) == 0 || name[0] != pattern[0] {
				return false
			}
			name = name[1:]
			pattern = pattern[1:]
		}
	}
	return len(name) == 0
}
