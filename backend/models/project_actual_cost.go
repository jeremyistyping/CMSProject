package models

import "time"

// ProjectActualCost represents a derived actual_costs row for a project.
// Backed by the project_actual_costs view (see migrations/20251115_create_project_progress_and_actual_costs.sql).
type ProjectActualCost struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	ProjectID       uint      `json:"project_id"`
	ProjectBudgetID *uint     `json:"project_budget_id"`
	SourceType      string    `json:"source_type"`
	SourceID        uint      `json:"source_id"`
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	Status          string    `json:"status"`
}

// TableName specifies the backing view name for ProjectActualCost
func (ProjectActualCost) TableName() string {
	return "project_actual_costs"
}
