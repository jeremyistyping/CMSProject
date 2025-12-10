package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"time"
)

type PurchaseOrderService struct {
	poRepo *repositories.PurchaseOrderRepository
	prRepo repositories.PurchaseRequestRepository
}

func NewPurchaseOrderService(poRepo *repositories.PurchaseOrderRepository, prRepo repositories.PurchaseRequestRepository) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo: poRepo,
		prRepo: prRepo,
	}
}

// CreateFromPR creates a Purchase Order from an approved Purchase Request
func (s *PurchaseOrderService) CreateFromPR(req *models.CreatePORequest, userID uint) (*models.PurchaseOrder, error) {
	// Get the PR
	pr, err := s.prRepo.GetByID(req.PurchaseRequestID)
	if err != nil {
		return nil, errors.New("purchase request not found")
	}

	// Validate PR status - must be APPROVED
	if pr.Status != models.PRStatusApproved {
		return nil, errors.New("purchase request must be approved before creating PO")
	}

	// Check if PO already exists for this PR
	existingPOs, err := s.poRepo.GetByPRID(pr.ID)
	if err == nil && len(existingPOs) > 0 {
		// Check if any existing PO is not cancelled
		for _, po := range existingPOs {
			if po.Status != models.POStatusCancelled {
				return nil, errors.New("a purchase order already exists for this purchase request")
			}
		}
	}

	// Generate PO code
	code, err := s.poRepo.GeneratePOCode()
	if err != nil {
		return nil, err
	}

	// Use vendor from PR if not specified
	vendorID := req.VendorID
	if vendorID == nil && pr.VendorID != nil {
		vendorID = pr.VendorID
	}

	// Create PO
	po := &models.PurchaseOrder{
		Code:              code,
		PurchaseRequestID: pr.ID,
		ProjectID:         pr.ProjectID,
		VendorID:          vendorID,
		OrderDate:         time.Now(),
		DeliveryAddress:   req.DeliveryAddress,
		PaymentTerms:      req.PaymentTerms,
		Notes:             req.Notes,
		Status:            models.POStatusDraft,
		CreatedBy:         userID,
	}

	// Set expected delivery date if valid
	if req.ExpectedDeliveryDate.Valid {
		po.ExpectedDeliveryDate = &req.ExpectedDeliveryDate.Time
	}

	// Calculate totals from items
	var subtotal float64
	var items []models.PurchaseOrderItem

	// If items provided in request, use them
	if len(req.Items) > 0 {
		for _, itemReq := range req.Items {
			totalPrice := itemReq.Quantity * itemReq.UnitPrice
			subtotal += totalPrice

			item := models.PurchaseOrderItem{
				PRItemID:   &itemReq.PRItemID,
				MaterialID: itemReq.MaterialID,
				ItemName:   itemReq.ItemName,
				Quantity:   itemReq.Quantity,
				Unit:       itemReq.Unit,
				UnitPrice:  itemReq.UnitPrice,
				TotalPrice: totalPrice,
			}
			items = append(items, item)
		}
	} else {
		// Copy items from PR
		for _, prItem := range pr.Items {
			totalPrice := prItem.Quantity * prItem.EstimatedPrice
			subtotal += totalPrice

			item := models.PurchaseOrderItem{
				PRItemID:   &prItem.ID,
				MaterialID: prItem.MaterialID,
				ItemName:   prItem.ItemName,
				Quantity:   prItem.Quantity,
				Unit:       prItem.Unit,
				UnitPrice:  prItem.EstimatedPrice,
				TotalPrice: totalPrice,
			}
			items = append(items, item)
		}
	}

	po.Subtotal = subtotal
	po.TotalAmount = subtotal // Can add tax/discount logic later
	po.Items = items
	
	// Auto-send PO when created from approved PR (set status to SENT)
	po.Status = models.POStatusSent

	// Save PO
	if err := s.poRepo.Create(po); err != nil {
		return nil, err
	}

	// Update PR status to PO_CREATED
	if err := s.prRepo.UpdateStatus(pr.ID, models.PRStatusPOCreated); err != nil {
		// Log but don't fail
	}

	return s.poRepo.GetByID(po.ID)
}

// GetByID retrieves a purchase order by ID
func (s *PurchaseOrderService) GetByID(id uint) (*models.PurchaseOrder, error) {
	return s.poRepo.GetByID(id)
}

// GetAll retrieves all purchase orders with filters
func (s *PurchaseOrderService) GetAll(filter map[string]interface{}) ([]models.PurchaseOrder, error) {
	return s.poRepo.GetAll(filter)
}

// GetByPRID retrieves purchase orders by PR ID
func (s *PurchaseOrderService) GetByPRID(prID uint) ([]models.PurchaseOrder, error) {
	return s.poRepo.GetByPRID(prID)
}

// UpdateStatus updates PO status
func (s *PurchaseOrderService) UpdateStatus(id uint, status string) error {
	return s.poRepo.UpdateStatus(id, status)
}

// SendPO marks PO as sent to vendor
func (s *PurchaseOrderService) SendPO(id uint) error {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		return err
	}

	if po.Status != models.POStatusDraft {
		return errors.New("only draft PO can be sent")
	}

	return s.poRepo.UpdateStatus(id, models.POStatusSent)
}

// CreateGoodsReceipt creates a goods receipt for a PO
func (s *PurchaseOrderService) CreateGoodsReceipt(req *models.CreateGRRequest, userID uint) (*models.GoodsReceipt, error) {
	// Get PO
	po, err := s.poRepo.GetByID(req.PurchaseOrderID)
	if err != nil {
		return nil, errors.New("purchase order not found")
	}

	// Validate PO status
	if po.Status != models.POStatusSent && po.Status != models.POStatusPartialReceived {
		return nil, errors.New("purchase order must be sent before receiving goods")
	}

	// Generate GR code
	code, err := s.poRepo.GenerateGRCode()
	if err != nil {
		return nil, err
	}

	// Parse receipt date
	var receiptDate time.Time
	if req.ReceiptDate != "" {
		// Try parsing date-only format first (YYYY-MM-DD)
		parsedDate, err := time.Parse("2006-01-02", req.ReceiptDate)
		if err != nil {
			// Try RFC3339 format as fallback
			parsedDate, err = time.Parse(time.RFC3339, req.ReceiptDate)
			if err != nil {
				return nil, errors.New("invalid receipt date format, expected YYYY-MM-DD")
			}
		}
		receiptDate = parsedDate
	} else {
		receiptDate = time.Now()
	}

	// Create GR
	gr := &models.GoodsReceipt{
		Code:            code,
		PurchaseOrderID: po.ID,
		ProjectID:       po.ProjectID,
		ReceiptDate:     receiptDate,
		ReceivedBy:      userID,
		Notes:           req.Notes,
		Status:          models.GRStatusPending,
	}

	if err := s.poRepo.CreateGoodsReceipt(gr); err != nil {
		return nil, err
	}

	// Create GR items and update PO item received quantities
	for _, itemReq := range req.Items {
		grItem := &models.GoodsReceiptItem{
			GoodsReceiptID:   gr.ID,
			POItemID:         itemReq.POItemID,
			ReceivedQuantity: itemReq.ReceivedQuantity,
			AcceptedQuantity: itemReq.AcceptedQuantity,
			RejectedQuantity: itemReq.RejectedQuantity,
			RejectionReason:  itemReq.RejectionReason,
		}

		if err := s.poRepo.CreateGoodsReceiptItem(grItem); err != nil {
			return nil, err
		}

		// Update PO item received quantity
		if err := s.poRepo.UpdateItemReceivedQty(itemReq.POItemID, itemReq.AcceptedQuantity); err != nil {
			return nil, err
		}
	}

	// Check if all items received - update PO status
	allReceived, err := s.poRepo.CheckAllItemsReceived(po.ID)
	if err == nil {
		if allReceived {
			s.poRepo.UpdateStatus(po.ID, models.POStatusCompleted)
			// Also update PR status to COMPLETED
			s.prRepo.UpdateStatus(po.PurchaseRequestID, "COMPLETED")
		} else {
			s.poRepo.UpdateStatus(po.ID, models.POStatusPartialReceived)
		}
	}

	return s.poRepo.GetGoodsReceiptByID(gr.ID)
}

// GetGoodsReceiptsByPOID retrieves all goods receipts for a PO
func (s *PurchaseOrderService) GetGoodsReceiptsByPOID(poID uint) ([]models.GoodsReceipt, error) {
	return s.poRepo.GetGoodsReceiptsByPOID(poID)
}

// Delete deletes a PO (only if draft)
func (s *PurchaseOrderService) Delete(id uint) error {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		return err
	}

	if po.Status != models.POStatusDraft {
		return errors.New("only draft PO can be deleted")
	}

	return s.poRepo.Delete(id)
}
