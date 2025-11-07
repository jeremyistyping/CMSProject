package services

import (
	"fmt"
	"os"
)

// StatusValidationHelper provides centralized status validation logic
// consistent with SalesJournalServiceV2.ShouldPostToJournal
type StatusValidationHelper struct{}

// NewStatusValidationHelper creates a new instance
func NewStatusValidationHelper() *StatusValidationHelper {
	return &StatusValidationHelper{}
}

// ShouldAllowPayment checks if payment is allowed for given sale status
// KONSISTENSI DENGAN SalesJournalServiceV2.ShouldPostToJournal
// Hanya INVOICED, OVERDUE, dan PAID yang diizinkan
func (h *StatusValidationHelper) ShouldAllowPayment(saleStatus string) bool {
	// Menggunakan logika yang sama dengan SalesJournalServiceV2
	allowedStatuses := []string{"INVOICED", "PAID", "OVERDUE"}
	for _, allowed := range allowedStatuses {
		if saleStatus == allowed {
			return true
		}
	}
	return false
}

// ValidatePaymentAllocation validates if payment allocation is allowed for sale status
func (h *StatusValidationHelper) ValidatePaymentAllocation(saleStatus string, saleID uint) error {
	if !h.ShouldAllowPayment(saleStatus) {
		return fmt.Errorf("payment allocation not allowed for sale #%d with status '%s' - only INVOICED/OVERDUE/PAID are allowed (consistent with SalesJournalServiceV2)", saleID, saleStatus)
	}
	return nil
}

// IsDeprecatedMethodDisabled checks if deprecated payment methods should be blocked
func (h *StatusValidationHelper) IsDeprecatedMethodDisabled() bool {
	return os.Getenv("DISABLE_DEPRECATED_PAYMENT_METHODS") == "true"
}

// ValidateDeprecatedMethodUsage returns error if deprecated methods are disabled
func (h *StatusValidationHelper) ValidateDeprecatedMethodUsage(methodName string) error {
	if h.IsDeprecatedMethodDisabled() {
		return fmt.Errorf("deprecated payment method '%s' is disabled - use SalesJournalServiceV2 and SingleSourcePostingService instead", methodName)
	}
	return nil
}

// GetValidStatusesForPayment returns list of valid statuses for payment allocation
func (h *StatusValidationHelper) GetValidStatusesForPayment() []string {
	return []string{"INVOICED", "PAID", "OVERDUE"}
}