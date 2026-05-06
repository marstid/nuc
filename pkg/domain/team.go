package domain

type Team struct {
	TeamID      string   `json:"team_id"`
	TeamName    string   `json:"team_name"`
	ProjectID   string   `json:"project_id"`
	AssetGroups []string `json:"asset_groups"`
}
