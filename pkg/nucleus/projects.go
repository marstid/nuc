package nucleus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/marstid/nuc/pkg/domain"
)

// List returns all projects accessible to the authenticated user.
func (c *Client) List(ctx context.Context) ([]domain.Project, error) {
	body, err := c.get(ctx, "/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	var projects []domain.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("decoding projects response: %w", err)
	}

	return projects, nil
}

// Get returns a specific project by ID.
func (c *Client) Get(ctx context.Context, projectID string) (*domain.Project, error) {
	body, err := c.get(ctx, "/projects/"+projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("getting project %s: %w", projectID, err)
	}

	var project domain.Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("decoding project response: %w", err)
	}

	return &project, nil
}

// GetRiskScore returns the risk score for a project.
func (c *Client) GetRiskScore(ctx context.Context, projectID string) (*domain.RiskScore, error) {
	body, err := c.get(ctx, "/projects/"+projectID+"/riskscore", nil)
	if err != nil {
		return nil, fmt.Errorf("getting risk score for project %s: %w", projectID, err)
	}

	var riskScore domain.RiskScore
	if err := json.Unmarshal(body, &riskScore); err != nil {
		return nil, fmt.Errorf("decoding risk score response: %w", err)
	}

	riskScore.ProjectID = projectID // Normalise — API response may omit this.
	return &riskScore, nil
}
