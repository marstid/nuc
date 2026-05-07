package tools

import (
	"testing"

	"github.com/marstid/nuc/pkg/domain"
)

func TestTeamsWithGroupPrefix(t *testing.T) {
	tests := []struct {
		name      string
		groups    []domain.AssetGroup
		inGroup   string
		wantTeams map[string]bool
	}{
		{
			name: "filters teams by /service prefix",
			groups: []domain.AssetGroup{
				{Name: "/teams/team-euc/service/api-gw", AssetCount: 5},
				{Name: "/teams/team-euc/container", AssetCount: 3},
				{Name: "/teams/team-mdd/service/auth", AssetCount: 2},
				{Name: "/teams/kandji/container", AssetCount: 1},
				{Name: "/service/standalone", AssetCount: 1},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{"team-euc": true, "team-mdd": true},
		},
		{
			name: "no matches returns empty map",
			groups: []domain.AssetGroup{
				{Name: "/teams/team-euc/container", AssetCount: 3},
				{Name: "/teams/kandji", AssetCount: 1},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{},
		},
		{
			name: "team group without sub-path is skipped",
			groups: []domain.AssetGroup{
				{Name: "/teams/team-euc", AssetCount: 10},
				{Name: "/teams/team-euc/service/api", AssetCount: 2},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{"team-euc": true},
		},
		{
			name: "standalone service groups are ignored",
			groups: []domain.AssetGroup{
				{Name: "/service/standalone", AssetCount: 1},
				{Name: "/teams/team-euc/service/api", AssetCount: 2},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{"team-euc": true},
		},
		{
			name: "non-teams groups are ignored",
			groups: []domain.AssetGroup{
				{Name: "production", AssetCount: 100},
				{Name: "/docker/repo/nginx", AssetCount: 5},
				{Name: "/teams/team-euc/service/svc", AssetCount: 1},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{"team-euc": true},
		},
		{
			name: "multiple sub-groups under same team count once",
			groups: []domain.AssetGroup{
				{Name: "/teams/team-euc/service/api-gw", AssetCount: 5},
				{Name: "/teams/team-euc/service/auth", AssetCount: 2},
				{Name: "/teams/team-euc/service/web", AssetCount: 3},
			},
			inGroup:   "/service",
			wantTeams: map[string]bool{"team-euc": true},
		},
		{
			name: "different in_group prefix",
			groups: []domain.AssetGroup{
				{Name: "/teams/team-euc/container/nginx", AssetCount: 5},
				{Name: "/teams/team-mdd/service/api", AssetCount: 2},
			},
			inGroup:   "/container",
			wantTeams: map[string]bool{"team-euc": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := teamsWithGroupPrefix(tt.groups, tt.inGroup)
			if len(got) != len(tt.wantTeams) {
				t.Fatalf("expected %d teams, got %d: %v", len(tt.wantTeams), len(got), got)
			}
			for k := range tt.wantTeams {
				if !got[k] {
					t.Errorf("expected team %q in results", k)
				}
			}
			for k := range got {
				if !tt.wantTeams[k] {
					t.Errorf("unexpected team %q in results", k)
				}
			}
		})
	}
}
