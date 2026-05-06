package nucleus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/marstid/nuc/pkg/domain"
)

// ListTeams returns all teams in a project, excluding the default no-access team.
func (c *Client) ListTeams(ctx context.Context, projectID string) ([]domain.Team, error) {
	body, err := c.get(ctx, fmt.Sprintf("/projects/%s/teams", projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("listing teams for project %s: %w", projectID, err)
	}

	var teams []domain.Team
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, fmt.Errorf("decoding teams response: %w", err)
	}

	filtered := make([]domain.Team, 0, len(teams))
	for _, t := range teams {
		if t.TeamName == "(not) Default no access" {
			continue
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}
