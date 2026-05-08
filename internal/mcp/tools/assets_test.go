package tools

import (
	"testing"

	"github.com/marstid/nuc/pkg/domain"
)

func TestBuildAssetGroupsFilter(t *testing.T) {
	tests := []struct {
		name        string
		assetGroups string
		team        string
		service     string
		want        string
	}{
		{
			name: "all empty returns empty",
			want: "",
		},
		{
			name:        "only asset_groups",
			assetGroups: "/custom/group",
			want:        "/custom/group",
		},
		{
			name: "only team",
			team: "team-euc",
			want: "/teams/team-euc",
		},
		{
			name:    "only service",
			service: "payments",
			want:    "/service/payments",
		},
		{
			name:        "asset_groups and team",
			assetGroups: "/custom/group",
			team:        "team-euc",
			want:        "/custom/group,/teams/team-euc",
		},
		{
			name:        "asset_groups and service",
			assetGroups: "/custom/group",
			service:     "payments",
			want:        "/custom/group,/service/payments",
		},
		{
			name:    "team and service",
			team:    "team-euc",
			service: "payments",
			want:    "/teams/team-euc,/service/payments",
		},
		{
			name:        "all three",
			assetGroups: "/custom/group",
			team:        "team-euc",
			service:     "payments",
			want:        "/custom/group,/teams/team-euc,/service/payments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAssetGroupsFilter(tt.assetGroups, tt.team, tt.service)
			if got != tt.want {
				t.Errorf("buildAssetGroupsFilter(%q, %q, %q) = %q, want %q",
					tt.assetGroups, tt.team, tt.service, got, tt.want)
			}
		})
	}
}

func TestBuildAssetGroupsSlice(t *testing.T) {
	tests := []struct {
		name        string
		assetGroups []string
		team        string
		service     string
		want        []string
	}{
		{
			name: "all empty returns empty slice",
			want: []string{},
		},
		{
			name:        "nil input returns empty slice",
			assetGroups: nil,
			want:        []string{},
		},
		{
			name:        "only asset_groups",
			assetGroups: []string{"/custom/group"},
			want:        []string{"/custom/group"},
		},
		{
			name: "only team",
			team: "team-euc",
			want: []string{"/teams/team-euc"},
		},
		{
			name:    "only service",
			service: "payments",
			want:    []string{"/service/payments"},
		},
		{
			name:        "asset_groups and team",
			assetGroups: []string{"/custom/group"},
			team:        "team-euc",
			want:        []string{"/custom/group", "/teams/team-euc"},
		},
		{
			name:        "asset_groups and service",
			assetGroups: []string{"/custom/group"},
			service:     "payments",
			want:        []string{"/custom/group", "/service/payments"},
		},
		{
			name:    "team and service",
			team:    "team-euc",
			service: "payments",
			want:    []string{"/teams/team-euc", "/service/payments"},
		},
		{
			name:        "all three",
			assetGroups: []string{"/custom/group", "/another"},
			team:        "team-euc",
			service:     "payments",
			want:        []string{"/custom/group", "/another", "/teams/team-euc", "/service/payments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAssetGroupsSlice(tt.assetGroups, tt.team, tt.service)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("buildAssetGroupsSlice() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("buildAssetGroupsSlice()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildAssetGroupsSlice_DoesNotMutateInput(t *testing.T) {
	input := []string{"/original"}
	inputCopy := make([]string, len(input))
	copy(inputCopy, input)

	_ = buildAssetGroupsSlice(input, "team-euc", "payments")

	if len(input) != len(inputCopy) {
		t.Fatal("input slice was mutated")
	}
	for i := range input {
		if input[i] != inputCopy[i] {
			t.Errorf("input[%d] = %q, want %q", i, input[i], inputCopy[i])
		}
	}
}

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
