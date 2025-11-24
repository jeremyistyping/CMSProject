package services

import (
	"app-sistem-akuntansi/models"
	"fmt"

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
	TotalPurchasedValue float64 `json:"total_purchased_value"`
	TotalUsedValue      float64 `json:"total_used_value"`
	TotalRemainingValue float64 `json:"total_remaining_value"`
	TotalItems          int64   `json:"total_items"`
	LowStockItems       int64   `json:"low_stock_items"`
}

// MaterialItemSummary represents the status of a single material item in a project
type MaterialItemSummary struct {
	ProductID    uint    `json:"product_id"`
	ProductCode  string  `json:"product_code"`
	ProductName  string  `json:"product_name"`
	Unit         string  `json:"unit"`
	Category     string  `json:"category"`
	BudgetQty    float64 `json:"budget_qty"`
	PurchasedQty float64 `json:"purchased_qty"`
	UsedQty      float64 `json:"used_qty"`
	RemainingQty float64 `json:"remaining_qty"`
	AvgUnitCost  float64 `json:"avg_unit_cost"`
	TotalValue   float64 `json:"total_value"` // Remaining * AvgUnitCost
	Status       string  `json:"status"`      // OK, LOW, CRITICAL
}

// GetMaterialSummary retrieves aggregated material stats for a project
func (s *MaterialTrackingService) GetMaterialSummary(projectID uint) (*MaterialSummaryStats, error) {
	stats := &MaterialSummaryStats{}

	// 1. Total Purchased Value (from approved purchases linked to project)
	// We look at purchase_items where purchase is approved and project_id matches
	err := s.db.Table("purchase_items").
		Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
		Where("purchases.project_id = ?", projectID).
		Where("purchases.status IN ?", []string{"APPROVED", "COMPLETED", "PAID"}).
		Where("purchase_items.deleted_at IS NULL").
		Select("COALESCE(SUM(purchase_items.total_price), 0)").
		Scan(&stats.TotalPurchasedValue).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calc purchased value: %w", err)
	}

	// 2. Total Used Value (from inventory OUT movements linked to project)
	err = s.db.Table("inventories").
		Where("project_id = ?", projectID).
		Where("type = ?", "OUT").
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&stats.TotalUsedValue).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calc used value: %w", err)
	}

	stats.TotalRemainingValue = stats.TotalPurchasedValue - stats.TotalUsedValue

	// 3. Count items involved
	// Count distinct products purchased for this project
	err = s.db.Table("purchase_items").
		Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
		Where("purchases.project_id = ?", projectID).
		Where("purchases.status IN ?", []string{"APPROVED", "COMPLETED", "PAID"}).
		Where("purchase_items.deleted_at IS NULL").
		Select("COUNT(DISTINCT purchase_items.product_id)").
		Scan(&stats.TotalItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	return stats, nil
}

// GetMaterialItems retrieves detailed material list for a project
func (s *MaterialTrackingService) GetMaterialItems(projectID uint) ([]MaterialItemSummary, error) {
	var items []MaterialItemSummary

	// We need to aggregate data from 3 sources:
	// 1. Project Budget (for BudgetQty)
	// 2. Purchases (for PurchasedQty & Cost)
	// 3. Inventory (for UsedQty)

	// Step 1: Get all products relevant to this project (either budgeted or purchased)
	// This query gets the base list of products
	query := `
		SELECT DISTINCT p.id, COALESCE(p.code, ''), COALESCE(p.name, ''), COALESCE(p.unit, ''), COALESCE(c.name, '') as category
		FROM products p
		LEFT JOIN product_categories c ON c.id = p.category_id
		WHERE p.id IN (
			SELECT product_id FROM purchase_items pi
			JOIN purchases pu ON pu.id = pi.purchase_id
			WHERE pu.project_id = ? AND pu.status IN ('APPROVED', 'COMPLETED', 'PAID')
		)
		OR p.id IN (
			SELECT product_id FROM inventories WHERE project_id = ?
		)
	`

	rows, err := s.db.Raw(query, projectID, projectID).Rows()
	if err != nil {
		fmt.Printf("❌ GetMaterialItems: Failed to fetch products for project %d: %v\n", projectID, err)
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item MaterialItemSummary
		if err := rows.Scan(&item.ProductID, &item.ProductCode, &item.ProductName, &item.Unit, &item.Category); err != nil {
			fmt.Printf("⚠️ Error scanning row: %v\n", err)
			continue
		}

		// Get Purchased Qty & Avg Cost
		var purchaseStats struct {
			TotalQty  float64
			TotalCost float64
		}
		s.db.Table("purchase_items").
			Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
			Where("purchases.project_id = ?", projectID).
			Where("purchase_items.product_id = ?", item.ProductID).
			Where("purchases.status IN ?", []string{"APPROVED", "COMPLETED", "PAID"}).
			Select("COALESCE(SUM(quantity), 0) as total_qty, COALESCE(SUM(total_price), 0) as total_cost").
			Scan(&purchaseStats)

		item.PurchasedQty = purchaseStats.TotalQty
		if item.PurchasedQty > 0 {
			item.AvgUnitCost = purchaseStats.TotalCost / item.PurchasedQty
		}

		// Get Used Qty
		var usedQty float64
		s.db.Table("inventories").
			Where("project_id = ?", projectID).
			Where("product_id = ?", item.ProductID).
			Where("type = ?", "OUT").
			Select("COALESCE(SUM(quantity), 0)").
			Scan(&usedQty)
		item.UsedQty = usedQty

		// Calculate Remaining
		// Note: This is "Project Stock".
		// Logic: Project Stock = Purchased for Project - Used in Project
		// This assumes all purchases for a project are "received" into a virtual project warehouse.
		item.RemainingQty = item.PurchasedQty - item.UsedQty
		item.TotalValue = item.RemainingQty * item.AvgUnitCost

		// Get Budget Qty (if mapped via product -> account -> budget)
		// This is tricky because budgets are by Account (COA), not Product.
		// For now, we might leave BudgetQty as 0 unless we have a direct link.
		// Or we can try to link via Product.ExpenseAccountID -> ProjectBudget.AccountID
		// Let's try that best-effort link.

		// This is an approximation. Multiple products might map to same account.
		// We can't easily split the budget qty per product.
		// So we will skip BudgetQty per product for now, or fetch it if we had a specific "Material Budget" table.
		item.BudgetQty = 0

		// Determine Status
		if item.RemainingQty <= 0 {
			item.Status = "CRITICAL"
		} else if item.RemainingQty < item.PurchasedQty*0.2 {
			item.Status = "LOW"
		} else {
			item.Status = "OK"
		}

		items = append(items, item)
	}

	return items, nil
}

// GetMaterialMovements retrieves history of material in/out for a project
func (s *MaterialTrackingService) GetMaterialMovements(projectID uint) ([]models.Inventory, error) {
	var movements []models.Inventory

	// We want to show:
	// 1. IN: When purchase is received (Inventory IN linked to project?)
	//    Actually, usually Inventory IN is to Warehouse.
	//    But for Project Material Tracking, "IN" is when we buy it FOR the project.
	//    However, the inventory table structure is:
	//    - Type: IN/OUT
	//    - ProjectID: (Added now)

	// If we want to show "Purchases" as "IN" movements in this list, we might need to union query
	// or ensure that when we buy for a project, we insert an Inventory record with ProjectID.
	// CURRENTLY: Purchases don't automatically create Inventory records with ProjectID (unless we change that flow).
	// BUT, we can query `inventory` table where `project_id` is set.

	err := s.db.Preload("Product").
		Where("project_id = ?", projectID).
		Order("transaction_date DESC").
		Find(&movements).Error

	return movements, err
}
