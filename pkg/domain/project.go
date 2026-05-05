package domain

import "time"

// Project represents a Nucleus Security project.
// Note: the API returns project_id as a string (e.g. "3").
type Project struct {
	ID          string    `json:"project_id"`
	Name        string    `json:"project_name"`
	Description string    `json:"project_description"`
	Org         string    `json:"project_org"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RiskScore represents the risk score for a project.
type RiskScore struct {
	ProjectID string `json:"project_id"`
	Score     int    `json:"score"`
}
