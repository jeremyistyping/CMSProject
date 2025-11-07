package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TaxPaymentService handles tax-related payment operations (PPN Input/Output payments)
type TaxPaymentService struct {
	db                 *gorm.DB
	taxAccountService  *TaxAccountService
	journalService     *UnifiedJournalService
}

// NewTaxPaymentService creates a new TaxPaymentService instance
func NewTaxPaymentService(db *gorm.DB) *TaxPaymentService {
	return &TaxPaymentService{
		db:                db,
		taxAccountService: NewTaxAccountService(db),
		journalService:    NewUnifiedJournalService(db),
	}
}

// CreatePPNPaymentRequest represents a request to remit PPN (Setor PPN)
type CreatePPNPaymentRequest struct {
	PPNType     string    `json:"ppn_type" binding:"required"` // REMIT (Setor PPN) - deprecated but kept for compatibility
	Amount      float64   `json:"amount" binding:"required"`    // Amount to remit (optional, will be calculated if 0)
	Date        time.Time `json:"date" binding:"required"`
	CashBankID  uint      `json:"cash_bank_id" binding:"required"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
	PeriodMonth int       `json:"period_month"` // Tax period month
	PeriodYear  int       `json:"period_year"`  // Tax period year
}

// CreatePPNPayment creates PPN remittance (Setor PPN) with proper PSAK-compliant double-entry
// Logic per PSAK:
// - Calculate: PPN Terutang = PPN Keluaran - PPN Masukan
// - If positive (Kurang Bayar): Debit PPN Keluaran, Credit PPN Masukan, Credit Cash/Bank
// - If negative (Lebih Bayar): No payment needed, create adjustment entry
func (s *TaxPaymentService) CreatePPNPayment(req CreatePPNPaymentRequest, userID uint) (*models.Payment, error) {
	log.Printf("🏦 Starting PPN Remittance (Setor PPN): Amount=%.2f", req.Amount)

	// Note: ppn_type is deprecated but kept for backward compatibility
	// Modern implementation auto-calculates from balances

	// Amount validation removed - will be calculated from PPN balances if not provided

	// Validate date
	if req.Date.IsZero() {
		return nil, fmt.Errorf("payment date is required")
	}

	// Begin transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ PANIC in CreatePPNPayment: %v", r)
		}
	}()

	// Get Cash/Bank account
	var cashBank models.CashBank
	if err := tx.Preload("Account").First(&cashBank, req.CashBankID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("cash/bank account not found: %v", err)
	}
	log.Printf("📋 Cash/Bank Account: %s (ID: %d)", cashBank.Account.Name, cashBank.AccountID)

	// Get PPN accounts from settings
	settings, err := s.taxAccountService.GetSettings()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get tax settings: %v", err)
	}

	// Validate both PPN accounts are configured
	if settings.PurchaseInputVATAccountID == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("PPN Masukan account not configured in tax settings")
	}
	if settings.SalesOutputVATAccountID == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("PPN Keluaran account not configured in tax settings")
	}

	// Get PPN Masukan account (Asset - 1240)
	var ppnMasukanAccount models.Account
	if err := tx.First(&ppnMasukanAccount, settings.PurchaseInputVATAccountID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("PPN Masukan account not found: %v", err)
	}

	// Get PPN Keluaran account (Liability - 2103)
	var ppnKeluaranAccount models.Account
	if err := tx.First(&ppnKeluaranAccount, settings.SalesOutputVATAccountID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("PPN Keluaran account not found: %v", err)
	}

	log.Printf("📋 PPN Masukan: %s - %s (Balance: %.2f)", ppnMasukanAccount.Code, ppnMasukanAccount.Name, ppnMasukanAccount.Balance)
	log.Printf("📋 PPN Keluaran: %s - %s (Balance: %.2f)", ppnKeluaranAccount.Code, ppnKeluaranAccount.Name, ppnKeluaranAccount.Balance)

	// Calculate PPN Terutang (PPN yang harus dibayar)
	// PPN Terutang = PPN Keluaran (Liability) - PPN Masukan (Asset)
	// Note: PPN Keluaran balance is positive for liability, PPN Masukan balance is positive for asset
	ppnTerutang := ppnKeluaranAccount.Balance - ppnMasukanAccount.Balance

	log.Printf("📊 Calculated PPN Terutang: %.2f (Keluaran: %.2f - Masukan: %.2f)", ppnTerutang, ppnKeluaranAccount.Balance, ppnMasukanAccount.Balance)

	// Validate PPN Terutang
	if ppnTerutang <= 0 {
		tx.Rollback()
		return nil, fmt.Errorf("tidak ada PPN yang harus dibayar. PPN Terutang: %.2f (PPN Masukan lebih besar dari PPN Keluaran)", ppnTerutang)
	}

	// Use calculated amount or provided amount
	paymentAmount := req.Amount
	if paymentAmount <= 0 {
		paymentAmount = ppnTerutang
		log.Printf("✅ Using calculated PPN Terutang: %.2f", paymentAmount)
	} else if paymentAmount > ppnTerutang {
		tx.Rollback()
		return nil, fmt.Errorf("payment amount (%.2f) cannot exceed PPN Terutang (%.2f)", paymentAmount, ppnTerutang)
	}

	// Validate cash/bank balance
	if cashBank.Balance < paymentAmount {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient cash/bank balance. Available: %.2f, Required: %.2f", cashBank.Balance, paymentAmount)
	}

	// Generate payment code for PPN remittance
	code := s.generatePaymentCode("SETOR-PPN", tx)

	// Use default system contact (Tax Office)
	contactID := uint(1)

	// Create payment record
	payment := &models.Payment{
		Code:        code,
		ContactID:   contactID,
		UserID:      userID,
		Date:        req.Date,
		Amount:      paymentAmount,
		Method:      models.PaymentMethodBankTransfer, // Default untuk setor PPN
		Reference:   req.Reference,
		Status:      models.PaymentStatusCompleted, // PPN remittance langsung completed
		Notes:       fmt.Sprintf("Setor PPN - Terutang: %.2f. %s", ppnTerutang, req.Notes),
		PaymentType: models.PaymentTypeTaxPPN,
	}

	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment record: %v", err)
	}
	log.Printf("✅ Payment record created: ID=%d, Code=%s", payment.ID, payment.Code)

	// Create journal entry per PSAK
	// Logika Setor PPN yang benar:
	// Debit:  PPN Keluaran (Liability berkurang) - sesuai payment amount + kompensasi
	// Credit: PPN Masukan (Asset berkurang) - kompensasi dari masukan
	// Credit: Cash/Bank (kas keluar) - payment amount (neto)
	//
	// Example: PPN Keluaran 50jt, PPN Masukan 30jt, Terutang 20jt, Bayar 20jt
	// Debit:  PPN Keluaran 50jt (sesuai balance yang akan dikosongkan)
	// Credit: PPN Masukan 30jt (kompensasi sesuai balance yang akan dikosongkan)
	// Credit: Cash/Bank 20jt (pembayaran neto)
	
	// Calculate kompensasi amount (PPN Masukan yang akan dikompensasi)
	// Kompensasi tidak boleh lebih dari balance PPN Masukan
	kompensasiAmount := ppnMasukanAccount.Balance
	if kompensasiAmount > ppnKeluaranAccount.Balance {
		kompensasiAmount = ppnKeluaranAccount.Balance
	}
	
	// Debit amount untuk PPN Keluaran = payment amount + kompensasi
	// Karena kita bayar net (payment) tapi juga kompensasi (masukan)
	ppnKeluaranDebit := paymentAmount + kompensasiAmount
	
	log.Printf("📋 Journal Entry breakdown: PPN Keluaran Debit=%.2f, PPN Masukan Credit=%.2f, Cash Credit=%.2f",
		ppnKeluaranDebit, kompensasiAmount, paymentAmount)
	
	// Build journal lines (only include non-zero amounts)
	journalLines := []JournalLineRequest{}
	
	// Debit PPN Keluaran (liability berkurang) - always include if > 0
	if ppnKeluaranDebit > 0 {
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    uint64(settings.SalesOutputVATAccountID),
			Description:  fmt.Sprintf("Setor PPN - %s", payment.Code),
			DebitAmount:  decimal.NewFromFloat(ppnKeluaranDebit),
			CreditAmount: decimal.Zero,
		})
	}
	
	// Credit PPN Masukan (asset berkurang - kompensasi) - only if > 0
	if kompensasiAmount > 0 {
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    uint64(settings.PurchaseInputVATAccountID),
			Description:  fmt.Sprintf("Setor PPN - Kompensasi - %s", payment.Code),
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(kompensasiAmount),
		})
	}
	
	// Credit Cash/Bank (pembayaran neto) - always include
	journalLines = append(journalLines, JournalLineRequest{
		AccountID:    uint64(cashBank.AccountID),
		Description:  fmt.Sprintf("Setor PPN - Pembayaran Neto - %s", payment.Code),
		DebitAmount:  decimal.Zero,
		CreditAmount: decimal.NewFromFloat(paymentAmount),
	})

	// Create SSOT journal entry
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    uint64(payment.ID),
		Reference:   payment.Code,
		EntryDate:   payment.Date,
		Description: fmt.Sprintf("Setor PPN ke Negara - %s (Terutang: %.2f)", payment.Code, ppnTerutang),
		Lines:       journalLines,
		AutoPost:    true,
		CreatedBy:   uint64(userID),
	}

	journalService := NewUnifiedJournalService(tx)
	journalResponse, err := journalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create journal entry: %v", err)
	}

	log.Printf("✅ Journal entry created: ID=%d, EntryNumber=%s", journalResponse.ID, journalResponse.EntryNumber)

	// Update payment with journal reference
	journalEntryID := uint(journalResponse.ID)
	payment.JournalEntryID = &journalEntryID
	if err := tx.Save(payment).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to update payment with journal reference: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ PPN Remittance (Setor PPN) successfully created: %s (Amount: %.2f, PPN Terutang: %.2f)", payment.Code, paymentAmount, ppnTerutang)

	// Load relations for response
	if err := s.db.Preload("Contact").Preload("User").First(payment, payment.ID).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to load payment relations: %v", err)
	}

	return payment, nil
}

// generatePaymentCode generates unique payment code with prefix
func (s *TaxPaymentService) generatePaymentCode(prefix string, tx *gorm.DB) string {
	now := time.Now()
	yearMonth := now.Format("0601") // YYMM format

	// Get last sequence for this month
	var lastPayment models.Payment
	pattern := fmt.Sprintf("%s-%s-%%", prefix, yearMonth)
	
	err := tx.Where("code LIKE ?", pattern).Order("code DESC").First(&lastPayment).Error
	
	sequence := 1
	if err == nil && lastPayment.Code != "" {
		// Extract sequence from last code more reliably
		// Expected format: PREFIX-YYMM-NNNN
		// Example: SETOR-PPN-2511-0001
		parts := strings.Split(lastPayment.Code, "-")
		if len(parts) >= 3 {
			// Last part should be the sequence number
			lastPart := parts[len(parts)-1]
			if lastSeq, err := strconv.Atoi(lastPart); err == nil {
				sequence = lastSeq + 1
			}
		}
	}

	// Format: PREFIX-YYMM-NNNN (max 30 chars)
	// Example: SETOR-PPN-2511-0001 (20 chars for SETOR-PPN)
	return fmt.Sprintf("%s-%s-%04d", prefix, yearMonth, sequence)
}

// GetPPNPaymentsByType retrieves PPN payments filtered by type
func (s *TaxPaymentService) GetPPNPaymentsByType(paymentType string, startDate, endDate time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	
	query := s.db.Preload("Contact").Preload("User").Preload("TaxAccount").Preload("CashBank").
		Where("payment_type = ?", paymentType)
	
	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}
	
	if err := query.Order("date DESC, created_at DESC").Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get PPN payments: %v", err)
	}
	
	return payments, nil
}

// GetPPNPaymentSummary returns summary of PPN payments
func (s *TaxPaymentService) GetPPNPaymentSummary(startDate, endDate time.Time) (map[string]interface{}, error) {
	type SummaryResult struct {
		PaymentType string
		TotalAmount float64
		Count       int64
	}
	
	var results []SummaryResult
	
	query := s.db.Model(&models.Payment{}).
		Select("payment_type, SUM(amount) as total_amount, COUNT(*) as count").
		Where("payment_type IN ?", []string{models.PaymentTypeTaxPPNInput, models.PaymentTypeTaxPPNOutput})
	
	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}
	
	if err := query.Group("payment_type").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get PPN payment summary: %v", err)
	}
	
	summary := map[string]interface{}{
		"ppn_masukan": map[string]interface{}{
			"total_amount": 0.0,
			"count":        0,
		},
		"ppn_keluaran": map[string]interface{}{
			"total_amount": 0.0,
			"count":        0,
		},
	}
	
	for _, result := range results {
		if result.PaymentType == models.PaymentTypeTaxPPNInput {
			summary["ppn_masukan"] = map[string]interface{}{
				"total_amount": result.TotalAmount,
				"count":        result.Count,
			}
		} else if result.PaymentType == models.PaymentTypeTaxPPNOutput {
			summary["ppn_keluaran"] = map[string]interface{}{
				"total_amount": result.TotalAmount,
				"count":        result.Count,
			}
		}
	}
	
	return summary, nil
}

// GetPPNBalance retrieves current balance of PPN account (Masukan or Keluaran)
func (s *TaxPaymentService) GetPPNBalance(ppnType string) (float64, error) {
	// Validate PPN type
	if ppnType != "INPUT" && ppnType != "OUTPUT" {
		return 0, fmt.Errorf("ppn_type must be INPUT or OUTPUT")
	}
	
	// Get PPN account from settings
	settings, err := s.taxAccountService.GetSettings()
	if err != nil {
		return 0, fmt.Errorf("failed to get tax settings: %v", err)
	}
	
	var accountID uint
	if ppnType == "INPUT" {
		// PPN Masukan (Purchase VAT)
		if settings.PurchaseInputVATAccountID == 0 {
			return 0, fmt.Errorf("PPN Masukan account not configured in tax settings")
		}
		accountID = settings.PurchaseInputVATAccountID
	} else {
		// PPN Keluaran (Sales VAT)
		if settings.SalesOutputVATAccountID == 0 {
			return 0, fmt.Errorf("PPN Keluaran account not configured in tax settings")
		}
		accountID = settings.SalesOutputVATAccountID
	}
	
	// Get account balance
	var account models.Account
	if err := s.db.First(&account, accountID).Error; err != nil {
		return 0, fmt.Errorf("account not found: %v", err)
	}
	
	log.Printf("📊 PPN %s Balance: %.2f (Account: %s - %s)", ppnType, account.Balance, account.Code, account.Name)
	
	return account.Balance, nil
}
