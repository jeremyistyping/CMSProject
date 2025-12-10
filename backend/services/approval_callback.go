package services

import (
	"fmt"
)

// ApprovalCallbackHandler implements PostApprovalCallback interface
// It connects approval service to business logic services (PR, Purchase, etc.)
type ApprovalCallbackHandler struct {
	prService PurchaseRequestService
}

// NewApprovalCallbackHandler creates a new callback handler
func NewApprovalCallbackHandler(prService PurchaseRequestService) *ApprovalCallbackHandler {
	return &ApprovalCallbackHandler{
		prService: prService,
	}
}

// OnPurchaseApproved handles post-approval logic for purchases
func (h *ApprovalCallbackHandler) OnPurchaseApproved(purchaseID uint) error {
	// TODO: Implement purchase approval logic when needed
	fmt.Printf("Purchase %d approved - no additional processing configured\n", purchaseID)
	return nil
}

// OnPurchaseRequestApproved handles post-approval logic for purchase requests
// This creates expense transactions automatically when a PR is approved
func (h *ApprovalCallbackHandler) OnPurchaseRequestApproved(prID uint) error {
	fmt.Printf("🔄 Processing approved PR %d - creating expense transactions...\n", prID)
	
	if err := h.prService.CreateExpenseFromApprovedPR(prID); err != nil {
		return fmt.Errorf("failed to create expenses from PR %d: %w", prID, err)
	}
	
	fmt.Printf("✅ Successfully created expense transactions for PR %d\n", prID)
	return nil
}
