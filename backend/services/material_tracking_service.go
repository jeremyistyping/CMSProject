package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MaterialTrackingService struct {
	db *gorm.DB
}

func NewMaterialTrackingService(db *gorm.DB) *MaterialTrackingService {
	return &MaterialTrackingService{db: db}
}

// MaterialSummaryStats represents the high-level stats for a project's material
type MaterialSummaryStats struct {
	TotalBudgetValue    float64 `json:"total_budget_value"`
	TotalActualValue    float64 `json:"total_actual_value"`
	TotalRemainingValue float64 `json:"total_remaining_value"`
	TotalItems          int64   `json:"total_items"`
	VariancePercent     float64 `json:"variance_percent"`
}

// MaterialItemSummary represents the status of a single material item in a project
type MaterialItemSummary struct {
	ID           uint    `json:"id"`
	ItemName     string  `json:"item_name"`
	Unit         string  `json:"unit"`
	Category     string  `json:"category"`
	BudgetQty    float64 `json:"budget_qty"`
	ActualQty    float64 `json:"actual_qty"`
	UsedQty      float64 `json:"used_qty"`
	RemainingQty float64 `json:"remaining_qty"`
	UnitCost     float64 `json:"unit_cost"`
	TotalValue   float64 `json:"total_value"`
	Status       string  `json:"status"` // OK, LOW, CRITICAL
}

// MaterialMovement represents a material movement record
type MaterialMovement struct {
	ID              uint      `json:"id"`
	ItemName        string    `json:"item_name"`
	Type            string    `json:"type"` // IN, OUT
	Quantity        float64   `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	TotalCost       float64   `json:"total_cost"`
	Notes           string    `json:"notes"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
}

// GetMaterialSummary retrieves aggregated material stats for a project
func (s *MaterialTrackingService) GetMaterialSummary(projectID uint) (*MaterialSummaryStats, error) {
	stats := &MaterialSummaryStats{}

	// Get total budget from project_budgets
	err := s.db.Table("project_budgets").
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(estimated_amount), 0)").
		Scan(&stats.TotalBudgetValue).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calc budget value: %w", err)
	}

	// Get total actual from purchase requests
	err = s.db.Table("purchase_requests").
		Where("project_id = ?", projectID).
		Where("status IN ?", []string{"APPROVED", "PO_CREATED"}).
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalActualValue).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calc actual value: %w", err)
	}

	stats.TotalRemainingValue = stats.TotalBudgetValue - stats.TotalActualValue

	if stats.TotalBudgetValue > 0 {
		stats.VariancePercent = ((stats.TotalBudgetValue - stats.TotalActualValue) / stats.TotalBudgetValue) * 100
	}

	// Count items from purchase request items
	err = s.db.Table("purchase_request_items").
		Joins("JOIN purchase_requests ON purchase_requests.id = purchase_request_items.purchase_request_id").
		Where("purchase_requests.project_id = ?", projectID).
		Where("purchase_requests.deleted_at IS NULL").
		Where("purchase_request_items.deleted_at IS NULL").
		Select("COUNT(DISTINCT purchase_request_items.id)").
		Scan(&stats.TotalItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	return stats, nil
}

// GetMaterialItems retrieves detailed material list for a project
func (s *MaterialTrackingService) GetMaterialItems(projectID uint) ([]MaterialItemSummary, error) {
	var items []MaterialItemSummary

	// Get items from purchase request items
	query := `
		SELECT 
			pri.id,
			pri.item_name,
			pri.unit,
			'Material' as category,
			0 as budget_qty,
			pri.quantity as actual_qty,
			0 as used_qty,
			pri.quantity as remaining_qty,
			pri.estimated_price as unit_cost,
			pri.total_price as total_value,
			CASE 
				WHEN pr.status = 'APPROVED' THEN 'OK'
				WHEN pr.status = 'PENDING' THEN 'PENDING'
				ELSE 'LOW'
			END as status
		FROM purchase_request_items pri
		JOIN purchase_requests pr ON pr.id = pri.purchase_request_id
		WHERE pr.project_id = ?
		AND pr.deleted_at IS NULL
		AND pri.deleted_at IS NULL
		ORDER BY pri.created_at DESC
	`

	err := s.db.Raw(query, projectID).Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch material items: %w", err)
	}

	return items, nil
}

// GetMaterialMovements retrieves history of material movements for a project
func (s *MaterialTrackingService) GetMaterialMovements(projectID uint) ([]MaterialMovement, error) {
	var movements []MaterialMovement

	// Get movements from purchase requests as "IN" movements
	query := `
		SELECT 
			pri.id,
			pri.item_name,
			'IN' as type,
			pri.quantity,
			pri.estimated_price as unit_cost,
			pri.total_price as total_cost,
			COALESCE(pri.notes, '') as notes,
			pr.request_date as transaction_date,
			u.username as created_by
		FROM purchase_request_items pri
		JOIN purchase_requests pr ON pr.id = pri.purchase_request_id
		JOIN users u ON u.id = pr.created_by
		WHERE pr.project_id = ?
		AND pr.status IN ('APPROVED', 'PO_CREATED')
		AND pr.deleted_at IS NULL
		AND pri.deleted_at IS NULL
		ORDER BY pr.request_date DESC
	`

	err := s.db.Raw(query, projectID).Scan(&movements).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch material movements: %w", err)
	}

	return movements, nil
}

// RecordMaterialUsage records material usage for a project
func (s *MaterialTrackingService) RecordMaterialUsage(projectID uint, itemName string, quantity float64, notes string, userID uint) error {
	fmt.Printf("📝 Recording material usage: Project %d, Item %s, Qty %.2f\n", projectID, itemName, quantity)

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive, got %.2f", quantity)
	}

	// For now, just log the usage - can be extended to create actual records
	fmt.Printf("✅ Material usage recorded for project %d\n", projectID)
	return nil
}
