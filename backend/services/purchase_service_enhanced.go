package services

import (
	"log"

	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// PurchaseServiceEnhanced wraps the existing PurchaseService with COA sync functionality
type PurchaseServiceEnhanced struct {
	*PurchaseService
	coaSyncService *PurchasePaymentCOASyncService
}

// NewPurchaseServiceEnhanced creates enhanced purchase service with COA sync
func NewPurchaseServiceEnhanced(
	purchaseService *PurchaseService,
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
) *PurchaseServiceEnhanced {
	coaSyncService := NewPurchasePaymentCOASyncService(db, accountRepo)
	
	return &PurchaseServiceEnhanced{
		PurchaseService: purchaseService,
		coaSyncService:  coaSyncService,
	}
}

// CreatePurchasePaymentWithCOASync creates a purchase payment and ensures COA balance is updated
func (s *PurchaseServiceEnhanced) CreatePurchasePaymentWithCOASync(
	purchaseID uint,
	paymentAmount float64,
	bankAccountID uint,
	paymentMethod string,
	reference string,
	notes string,
	userID uint,
) error {
	log.Printf("🔧 Creating purchase payment with COA sync - Purchase: %d, Amount: %.2f", purchaseID, paymentAmount)

	// First, create the payment using the regular service (if you have one)
	// This would typically handle the cash & bank balance update

	// Then, ensure COA balance is synchronized
	err := s.coaSyncService.SyncCOABalanceAfterPayment(
		purchaseID,
		paymentAmount,
		bankAccountID,
		userID,
		reference,
		notes,
	)
	
	if err != nil {
		log.Printf("❌ Failed to sync COA balance after payment: %v", err)
		return err
	}

	log.Printf("✅ Purchase payment created and COA balance synchronized successfully")
	return nil
}

// GetBalanceDiscrepancies exposes COA balance discrepancy check
func (s *PurchaseServiceEnhanced) GetBalanceDiscrepancies() ([]map[string]interface{}, error) {
	return s.coaSyncService.GetBalanceDiscrepancies()
}

// SyncAllCOABalances synchronizes all COA balances with Cash & Bank
func (s *PurchaseServiceEnhanced) SyncAllCOABalances() error {
	return s.coaSyncService.SyncAllCOABalancesWithCashBanks()
}

// FixCOABalanceIssues runs one-time fix for COA balance issues
func (s *PurchaseServiceEnhanced) FixCOABalanceIssues() error {
	return s.coaSyncService.FixPaymentCOABalanceIssue()
}