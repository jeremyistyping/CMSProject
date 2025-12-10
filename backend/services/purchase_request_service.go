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
	VerifyPR(prID uint, mappings []models.PRCBSMapping, verifierID uint, notes string) error
	GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error)
	ValidateCBSAllocation(prID uint, mappings []models.PRCBSMapping) error
	CreateExpenseFromApprovedPR(prID uint) error
}

type purchaseRequestService struct {
	repo            repositories.PurchaseRequestRepository
	cbsRepo         repositories.CBSRepository
	db              *gorm.DB
	approvalService *ApprovalService
	expenseRepo     repositories.ExpenseTransactionRepository
	materialRepo    repositories.MaterialRepository
}

func NewPurchaseRequestService(
	repo repositories.PurchaseRequestRepository,
	cbsRepo repositories.CBSRepository,
	db *gorm.DB,
	approvalService *ApprovalService,
	expenseRepo repositories.ExpenseTransactionRepository,
	materialRepo repositories.MaterialRepository,
) PurchaseRequestService {
	return &purchaseRequestService{
		repo:            repo,
		cbsRepo:         cbsRepo,
		db:              db,
		approvalService: approvalService,
		expenseRepo:     expenseRepo,
		materialRepo:    materialRepo,
	}
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

	// Create the PR
	if err := s.repo.Create(pr); err != nil {
		return err
	}

	// Create approval request for the PR
	if s.approvalService != nil {
		dto := models.CreateApprovalRequestDTO{
			EntityType:     models.EntityTypePurchaseRequest,
			EntityID:       pr.ID,
			Amount:         pr.TotalAmount,
			RequestTitle:   fmt.Sprintf("Purchase Request %s - Project %d", pr.Code, pr.ProjectID),
			RequestMessage: fmt.Sprintf("PR requesting materials for project %d with total amount %.2f", pr.ProjectID, pr.TotalAmount),
			Priority:       models.ApprovalPriorityNormal,
		}

		approvalReq, err := s.approvalService.CreateApprovalRequest(dto, pr.CreatedBy)
		if err != nil {
			// Log error but don't fail PR creation
			fmt.Printf("⚠️  Failed to create approval request for PR %d: %v\n", pr.ID, err)
		} else {
			fmt.Printf("✅ Created approval request %s for PR %s\n", approvalReq.RequestCode, pr.Code)
		}
	}

	return nil
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
		// Get current stock for this item in this project (if tracking exists)
		var currentStock float64
		if item.ProductID != nil {
			err := s.db.Table("inventories").
				Select("COALESCE(SUM(CASE WHEN type = 'IN' THEN quantity ELSE -quantity END), 0)").
				Where("project_id = ?", pr.ProjectID).
				Where("product_id = ?", *item.ProductID).
				Where("deleted_at IS NULL").
				Scan(&currentStock).Error

			if err != nil {
				// Log error but continue with 0 stock
				fmt.Printf("Error getting stock for product %d: %v\n", *item.ProductID, err)
			}
		}

		impact := models.MaterialImpact{
			ProductID:      item.ProductID,
			ProductName:    item.ItemName,
			ProductCode:    "", // No product code since Product model is removed
			Unit:           item.Unit,
			RequestedQty:   item.Quantity,
			CurrentStock:   currentStock,
			ProjectedStock: currentStock + item.Quantity,
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

// VerifyPR verifies a PR and maps it to CBS nodes
func (s *purchaseRequestService) VerifyPR(prID uint, mappings []models.PRCBSMapping, verifierID uint, notes string) error {
	// Get PR
	pr, err := s.repo.FindByID(prID)
	if err != nil {
		return err
	}

	// Only allow verification of PENDING PRs
	if pr.Status != models.PRStatusPending {
		return fmt.Errorf("can only verify PRs with PENDING status")
	}

	// Validate CBS allocation matches PR total
	if err := s.ValidateCBSAllocation(prID, mappings); err != nil {
		return err
	}

	// Delete existing mappings if any
	s.cbsRepo.DeletePRCBSMappings(prID)

	// Create new mappings
	for _, mapping := range mappings {
		mapping.PurchaseRequestID = prID
		mapping.CreatedBy = &verifierID
		if err := s.cbsRepo.CreatePRCBSMapping(&mapping); err != nil {
			return err
		}
	}

	// Update PR verification status
	now := time.Now()
	pr.VerifiedBy = &verifierID
	pr.VerifiedAt = &now
	pr.VerificationNotes = &notes
	pr.Status = "VERIFIED"

	return s.db.Save(pr).Error
}

// GetPRCBSMappings retrieves CBS mappings for a PR
func (s *purchaseRequestService) GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error) {
	return s.cbsRepo.GetPRCBSMappings(prID)
}

// ValidateCBSAllocation ensures mappings total matches PR total
func (s *purchaseRequestService) ValidateCBSAllocation(prID uint, mappings []models.PRCBSMapping) error {
	pr, err := s.repo.FindByID(prID)
	if err != nil {
		return err
	}

	// Calculate total allocated
	var totalAllocated int64
	for _, mapping := range mappings {
		totalAllocated += mapping.AllocatedAmount
	}

	// Convert PR total to cents (assuming TotalAmount is in currency units)
	prTotalCents := int64(pr.TotalAmount * 100)

	if totalAllocated != prTotalCents {
		return fmt.Errorf("CBS allocation (%d) does not match PR total (%d)", totalAllocated, prTotalCents)
	}

	return nil
}

// CreateExpenseFromApprovedPR creates expense transactions when a PR is approved
func (s *purchaseRequestService) CreateExpenseFromApprovedPR(prID uint) error {
	fmt.Printf("📝 [CreateExpenseFromApprovedPR] Starting for PR ID: %d\n", prID)
	
	// Get PR with items
	pr, err := s.repo.FindByID(prID)
	if err != nil {
		fmt.Printf("❌ [CreateExpenseFromApprovedPR] Failed to get PR %d: %v\n", prID, err)
		return fmt.Errorf("failed to get PR: %w", err)
	}
	fmt.Printf("📋 [CreateExpenseFromApprovedPR] Found PR: %s, Status: %s, Items: %d\n", pr.Code, pr.Status, len(pr.Items))

	// Only create expenses for approved PRs
	if pr.Status != models.PRStatusApproved {
		fmt.Printf("⚠️  [CreateExpenseFromApprovedPR] PR is not approved, status: %s\n", pr.Status)
		return fmt.Errorf("PR is not approved, status: %s", pr.Status)
	}

	createdCount := 0
	skippedCount := 0
	errorCount := 0

	// Create expense transaction for each PR item
	for i, item := range pr.Items {
		fmt.Printf("🔍 [CreateExpenseFromApprovedPR] Processing item %d/%d: %s (ID: %d)\n", i+1, len(pr.Items), item.ItemName, item.ID)
		
		// Determine COA account ID
		var coaAccountID uint
		var description string

		// Try to get COA from material
		if item.MaterialID != nil {
			fmt.Printf("   📦 Material ID: %d\n", *item.MaterialID)
			material, err := s.materialRepo.GetByID(*item.MaterialID)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to get material %d: %v\n", *item.MaterialID, err)
			} else if material == nil {
				fmt.Printf("   ⚠️  Material %d not found\n", *item.MaterialID)
			} else {
				fmt.Printf("   📦 Material found: %s (Code: %s)\n", material.Name, material.Code)
				if material.COAAccountID != nil {
					coaAccountID = *material.COAAccountID
					description = fmt.Sprintf("PR-%s: %s (Material: %s)", pr.Code, item.ItemName, material.Name)
					fmt.Printf("   ✅ COA Account ID: %d\n", coaAccountID)
				} else {
					fmt.Printf("   ⚠️  Material has no COA mapping\n")
				}
			}
		} else {
			fmt.Printf("   ⚠️  Item has no material_id\n")
		}

		// If no COA from material, skip this item (or use default COA)
		if coaAccountID == 0 {
			// Log warning but continue
			fmt.Printf("   ⏭️  Skipping: No COA found for PR item %d (%s)\n", item.ID, item.ItemName)
			skippedCount++
			continue
		}

		// Create expense transaction
		expense := &models.ExpenseTransaction{
			ProjectID:       pr.ProjectID,
			TransactionDate: time.Now(),
			COAAccountID:    coaAccountID,
			Description:     description,
			Amount:          item.TotalPrice,
			Unit:            item.Unit,
			Quantity:        item.Quantity,
			TransactionType: models.ExpenseTypeMaterial,
			ReferenceType:   models.ExpenseRefTypePR,
			ReferenceID:     &pr.ID,
			ReferenceNo:     pr.Code,
			Notes:           fmt.Sprintf("Auto-created from approved PR: %s", pr.Code),
			CreatedBy:       pr.CreatedBy,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		fmt.Printf("   💾 Saving expense: Project=%d, COA=%d, Amount=%.2f\n", expense.ProjectID, expense.COAAccountID, expense.Amount)
		
		// Save expense transaction
		if err := s.expenseRepo.Create(expense); err != nil {
			// Log error but continue with other items
			fmt.Printf("   ❌ Error creating expense for PR item %d: %v\n", item.ID, err)
			errorCount++
			continue
		}

		fmt.Printf("   ✅ Created expense transaction ID: %d for PR-%s item: %s (Amount: %.2f)\n", 
			expense.ID, pr.Code, item.ItemName, item.TotalPrice)
		createdCount++
	}

	fmt.Printf("📊 [CreateExpenseFromApprovedPR] Summary - Created: %d, Skipped: %d, Errors: %d\n", createdCount, skippedCount, errorCount)
	
	if createdCount == 0 && len(pr.Items) > 0 {
		return fmt.Errorf("no expenses were created (skipped: %d, errors: %d)", skippedCount, errorCount)
	}

	return nil
}
