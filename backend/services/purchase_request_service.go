package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PurchaseRequestService interface {
	CreatePR(pr *models.PurchaseRequest) error
	UpdatePR(id uint, pr *models.PurchaseRequest) error
	DeletePR(id uint) error
	GetPRByID(id uint) (*models.PurchaseRequest, error)
	GetAllPRs(filter map[string]interface{}) ([]models.PurchaseRequest, error)
	UpdateStatus(id uint, status string, approverID uint, reason string) error
	GetEstimatedMaterialImpact(prID uint) ([]models.MaterialImpact, error)
}

type purchaseRequestService struct {
	repo repositories.PurchaseRequestRepository
	db   *gorm.DB
}

func NewPurchaseRequestService(repo repositories.PurchaseRequestRepository, db *gorm.DB) PurchaseRequestService {
	return &purchaseRequestService{repo, db}
}

func (s *purchaseRequestService) CreatePR(pr *models.PurchaseRequest) error {
	// Generate Code
	pr.Code = fmt.Sprintf("PR-%d-%s", pr.ProjectID, time.Now().Format("20060102150405"))
	pr.Status = models.PRStatusPending

	// Calculate Total Amount
	var total float64
	for _, item := range pr.Items {
		total += item.TotalPrice
	}
	pr.TotalAmount = total

	return s.repo.Create(pr)
}

func (s *purchaseRequestService) UpdatePR(id uint, pr *models.PurchaseRequest) error {
	existingPR, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existingPR.Status != models.PRStatusPending && existingPR.Status != models.PRStatusRevision {
		return fmt.Errorf("cannot update PR with status %s", existingPR.Status)
	}

	// Update fields
	existingPR.ProjectID = pr.ProjectID
	existingPR.RequestDate = pr.RequestDate
	existingPR.RequiredDate = pr.RequiredDate
	existingPR.VendorID = pr.VendorID
	existingPR.Notes = pr.Notes
	existingPR.Items = pr.Items

	// Recalculate Total
	var total float64
	for _, item := range pr.Items {
		total += item.TotalPrice
	}
	existingPR.TotalAmount = total

	return s.repo.Update(existingPR)
}

func (s *purchaseRequestService) DeletePR(id uint) error {
	existingPR, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existingPR.Status != models.PRStatusPending {
		return fmt.Errorf("cannot delete PR with status %s", existingPR.Status)
	}

	return s.repo.Delete(id)
}

func (s *purchaseRequestService) GetPRByID(id uint) (*models.PurchaseRequest, error) {
	return s.repo.FindByID(id)
}

func (s *purchaseRequestService) GetAllPRs(filter map[string]interface{}) ([]models.PurchaseRequest, error) {
	return s.repo.FindAll(filter)
}

func (s *purchaseRequestService) UpdateStatus(id uint, status string, approverID uint, reason string) error {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	pr.Status = status
	if status == models.PRStatusApproved {
		pr.ApprovedBy = &approverID
		now := time.Now()
		pr.ApprovedAt = &now
	} else if status == models.PRStatusRejected || status == models.PRStatusRevision {
		pr.RejectionReason = reason
	}

	return s.repo.Update(pr)
}

// GetEstimatedMaterialImpact calculates how the PR items will affect the project's material stock
func (s *purchaseRequestService) GetEstimatedMaterialImpact(prID uint) ([]models.MaterialImpact, error) {
	// Get PR with items
	pr, err := s.repo.FindByID(prID)
	if err != nil {
		return nil, err
	}

	var impacts []models.MaterialImpact

	for _, item := range pr.Items {
		// Get current stock for this product in this project
		var currentStock float64
		err := s.db.Table("inventories").
			Select("COALESCE(SUM(CASE WHEN type = 'IN' THEN quantity ELSE -quantity END), 0)").
			Where("project_id = ?", pr.ProjectID).
			Where("product_id = ?", item.ProductID).
			Where("deleted_at IS NULL").
			Scan(&currentStock).Error

		if err != nil {
			// Log error but continue with 0 stock
			fmt.Printf("Error getting stock for product %d: %v\n", item.ProductID, err)
		}

		// Get product details
		var product models.Product
		if err := s.db.First(&product, item.ProductID).Error; err != nil {
			continue
		}

		impact := models.MaterialImpact{
			ProductID:      item.ProductID,
			ProductName:    product.Name,
			ProductCode:    product.Code,
			Unit:           product.Unit,
			RequestedQty:   float64(item.Quantity),
			CurrentStock:   currentStock,
			ProjectedStock: currentStock + float64(item.Quantity),
			Status:         "OK",
		}

		// Determine status (simple logic for now)
		if impact.CurrentStock <= 0 {
			impact.Status = "CRITICAL" // Currently out of stock, so this purchase is critical
		} else if impact.CurrentStock < 10 { // Arbitrary low stock threshold
			impact.Status = "LOW"
		}

		impacts = append(impacts, impact)
	}

	return impacts, nil
}
