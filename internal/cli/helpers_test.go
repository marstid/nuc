package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		// Trailing star matches anything after prefix.
		{"/teams/team-eu*", "/teams/team-euc", true},
		{"/teams/team-eu*", "/teams/team-euc/container", true},
		{"/teams/team-eu*", "/teams/team-euc/services/foo", true},
		// Exact match without glob.
		{"/teams/team-euc", "/teams/team-euc", true},
		{"/teams/team-euc", "/teams/team-euc/child", false},
		// Middle star.
		{"/teams/*/services", "/teams/team-euc/services", true},
		{"/teams/*/services", "/teams/team-euc/services/foo", false},
		// Question mark.
		{"/teams/team-eu?", "/teams/team-euc", true},
		{"/teams/team-eu?", "/teams/team-eu", false},
		{"/teams/team-eu?", "/teams/team-eucc", false},
		// Non-matching.
		{"/teams/team-mdd*", "/teams/team-euc", false},
		{"/teams/team-eu*", "/docker/repo/foo", false},
		// Star matches slash (unlike filepath.Match).
		{"*euc*", "/teams/team-euc/container", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_vs_"+tt.name, func(t *testing.T) {
			got := globMatch(tt.pattern, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDedupChildren(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		want   []string
	}{
		{
			name:  "single group",
			input: []string{"/teams/team-euc"},
			want:  []string{"/teams/team-euc"},
		},
		{
			name: "parent with children",
			input: []string{
				"/teams/team-euc",
				"/teams/team-euc/container",
				"/teams/team-euc/container/newcorpdev/foo",
				"/teams/team-euc/services",
				"/teams/team-euc/services/bar",
			},
			want: []string{"/teams/team-euc"},
		},
		{
			name: "multiple top-level matches",
			input: []string{
				"/teams/team-euc",
				"/teams/team-euc/container",
				"/teams/team-euw",
				"/teams/team-euw/services",
			},
			want: []string{"/teams/team-euc", "/teams/team-euw"},
		},
		{
			name: "no parent-child relationship",
			input: []string{
				"/teams/team-alpha/services/foo",
				"/teams/team-beta/services/bar",
				"/docker/repo/baz",
			},
			want: []string{"/docker/repo/baz", "/teams/team-alpha/services/foo", "/teams/team-beta/services/bar"},
		},
		{
			name: "does not match partial prefix",
			input: []string{
				"/teams/team-eu",
				"/teams/team-euc",
			},
			want: []string{"/teams/team-eu", "/teams/team-euc"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupChildren(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
