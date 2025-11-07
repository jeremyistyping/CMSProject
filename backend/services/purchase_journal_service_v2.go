package services

import (
	"fmt"
	"log"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// PurchaseJournalServiceV2 handles all purchase-related journal entries with proper status validation
type PurchaseJournalServiceV2 struct {
	db          *gorm.DB
	journalRepo repositories.JournalEntryRepository
	coaService  *COAService
}

// NewPurchaseJournalServiceV2 creates a new instance of PurchaseJournalServiceV2
func NewPurchaseJournalServiceV2(
	db *gorm.DB,
	journalRepo repositories.JournalEntryRepository,
	coaService *COAService,
) *PurchaseJournalServiceV2 {
	return &PurchaseJournalServiceV2{
		db:              db,
		journalRepo:     journalRepo,
		coaService:      coaService,
	}
}

// ShouldPostToJournal checks if a status should create journal entries
// KRITERIA POSTING JURNAL - HANYA STATUS APPROVED, COMPLETED ATAU PAID
func (s *PurchaseJournalServiceV2) ShouldPostToJournal(status string) bool {
	// DRAFT dan PENDING TIDAK BOLEH POSTING
	allowedStatuses := []string{"APPROVED", "COMPLETED", "PAID"}
	for _, allowed := range allowedStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

// CreatePurchaseJournal creates journal entries for a purchase based on status
func (s *PurchaseJournalServiceV2) CreatePurchaseJournal(purchase *models.Purchase, tx *gorm.DB) error {
	// VALIDASI STATUS - HANYA APPROVED/COMPLETED/PAID YANG BOLEH POSTING
	if !s.ShouldPostToJournal(purchase.Status) {
		log.Printf("⚠️ Skipping journal creation for Purchase #%d with status: %s (only APPROVED/COMPLETED/PAID allowed)", purchase.ID, purchase.Status)
		return nil // Return nil karena ini bukan error, tapi by design
	}

	log.Printf("📝 Creating journal entries for Purchase #%d (Status: %s, Payment Method: %s)", 
		purchase.ID, purchase.Status, purchase.PaymentMethod)

	// Tentukan database yang akan digunakan
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Create Simple SSOT Journal Entry
	ssotEntry := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("PURCHASE-%d", purchase.ID),
		TransactionType:   "PURCHASE",
		TransactionID:     purchase.ID,
		TransactionNumber: purchase.Code,
		Date:              purchase.Date,
		Description:       fmt.Sprintf("Purchase Order #%s - %s", purchase.Code, purchase.Vendor.Name),
		TotalAmount:       purchase.TotalAmount,
		Status:            "POSTED",
		CreatedBy:         purchase.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Create SSOT Journal Entry
	if err := dbToUse.Create(ssotEntry).Error; err != nil {
		return fmt.Errorf("failed to create SSOT journal: %v", err)
	}

	// Helper to resolve account by code
	resolveByCode := func(code string) (*models.Account, error) {
		var acc models.Account
		if err := dbToUse.Where("code = ?", code).First(&acc).Error; err != nil {
			return nil, fmt.Errorf("account code %s not found: %v", code, err)
		}
		return &acc, nil
	}

	// Prepare journal items
	var journalItems []models.SimpleSSOTJournalItem

	// DEBIT SIDE - Inventory/Expense accounts and PPN Masukan
	
	// 1. Debit Inventory/Expense accounts for each item
	for _, item := range purchase.PurchaseItems {
		// Determine account based on item type
		if item.ExpenseAccountID != 0 {
			// Use specific expense account if provided
			var acc models.Account
			if err := dbToUse.First(&acc, item.ExpenseAccountID).Error; err == nil {
				journalItems = append(journalItems, models.SimpleSSOTJournalItem{
					JournalID:   ssotEntry.ID,
					AccountID:   acc.ID,
					AccountCode: acc.Code,
					AccountName: acc.Name,
					Debit:       item.TotalPrice,
					Credit:      0,
					Description: fmt.Sprintf("Purchase expense - %s", item.Product.Name),
				})
				log.Printf("📋 Expense: Debit to %s (%s) = %.2f", acc.Name, acc.Code, item.TotalPrice)
			} else {
				log.Printf("⚠️ Failed to load expense account %d: %v", item.ExpenseAccountID, err)
			}
		} else {
			// Use default inventory account (1301 - Persediaan Barang Dagangan)
			if acc, err := resolveByCode("1301"); err == nil {
				journalItems = append(journalItems, models.SimpleSSOTJournalItem{
					JournalID:   ssotEntry.ID,
					AccountID:   acc.ID,
					AccountCode: acc.Code,
					AccountName: acc.Name,
					Debit:       item.TotalPrice,
					Credit:      0,
					Description: fmt.Sprintf("Inventory - %s", item.Product.Name),
				})
				log.Printf("📦 Inventory: Debit to %s (%s) = %.2f", acc.Name, acc.Code, item.TotalPrice)
			} else {
				log.Printf("⚠️ Missing inventory account 1301: %v", err)
			}
		}
	}

	// 2. Debit PPN Masukan (1240) - Input VAT jika ada
	if purchase.PPNAmount > 0 {
		if acc, err := resolveByCode("1240"); err == nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       purchase.PPNAmount,
				Credit:      0,
				Description: fmt.Sprintf("Input VAT for Purchase #%s", purchase.Code),
			})
			log.Printf("📊 Input VAT: Debit to %s (%s) = %.2f", acc.Name, acc.Code, purchase.PPNAmount)
		} else {
			log.Printf("⚠️ Missing PPN Masukan account 1240: %v", err)
		}
	}

	// CREDIT SIDE - Based on payment method
	// Calculate net amount after withholdings (for cash/bank/AP credit)
	netAmount := purchase.TotalAmount - purchase.PPh21Amount - purchase.PPh23Amount - purchase.OtherTaxDeductions
	
	// Determine CREDIT side based on payment method
	switch strings.ToUpper(strings.TrimSpace(purchase.PaymentMethod)) {
	case "CASH":
		// CASH → Credit Kas (1101) with net amount
		if acc, err := resolveByCode("1101"); err == nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       0,
				Credit:      netAmount,
				Description: fmt.Sprintf("Cash payment for Purchase #%s", purchase.Code),
			})
			log.Printf("💵 CASH Payment: Credit to %s (%s) = %.2f", acc.Name, acc.Code, netAmount)
		} else {
			log.Printf("⚠️ Missing cash account 1101: %v", err)
		}
	case "BANK_TRANSFER", "BANK":
		// BANK → Credit ke Bank (specific cash_bank account if available, otherwise 1102) with net amount
		var acc *models.Account
		if purchase.BankAccountID != nil && *purchase.BankAccountID > 0 {
			var cashBank models.CashBank
			if err := dbToUse.First(&cashBank, *purchase.BankAccountID).Error; err == nil {
				// Load account for code/name
				var tmp models.Account
				if e2 := dbToUse.First(&tmp, cashBank.AccountID).Error; e2 == nil {
					acc = &tmp
				}
			}
		}
		if acc == nil {
			if resolved, err := resolveByCode("1102"); err == nil {
				acc = resolved
			} else {
				log.Printf("⚠️ Default bank account 1102 not found: %v", err)
			}
		}
		if acc != nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       0,
				Credit:      netAmount,
				Description: fmt.Sprintf("Bank transfer for Purchase #%s", purchase.Code),
			})
			log.Printf("🏦 BANK Payment: Credit to %s (%s) = %.2f", acc.Name, acc.Code, netAmount)
		}
	case "KREDIT", "CREDIT", "HUTANG":
		// CREDIT Purchase → Credit Hutang Usaha (2101) with net amount (after withholdings)
		if acc, err := resolveByCode("2101"); err == nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       0,
				Credit:      netAmount,
				Description: fmt.Sprintf("Account payable for Purchase #%s", purchase.Code),
			})
			log.Printf("📋 CREDIT Purchase: Credit to %s (%s) = %.2f (net after withholdings)", acc.Name, acc.Code, netAmount)
		} else {
			log.Printf("⚠️ Missing payable account 2101: %v", err)
		}
	default:
		// ❌ FIX: NO DEFAULT! Force explicit payment method
		log.Printf("❌ CRITICAL ERROR: Invalid payment method '%s' for Purchase #%s. Journal NOT created!", 
			purchase.PaymentMethod, purchase.Code)
		return fmt.Errorf("invalid payment method '%s' for Purchase #%s. Valid: TUNAI/CASH, TRANSFER/BANK, KREDIT/CREDIT", 
			purchase.PaymentMethod, purchase.Code)
	}

	// Handle tax withholdings if any
	
	// PPh 21 Payable (2111) - Credit jika ada
	if purchase.PPh21Amount > 0 {
		if acc, err := resolveByCode("2111"); err == nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       0,
				Credit:      purchase.PPh21Amount,
				Description: fmt.Sprintf("PPh 21 withholding for Purchase #%s", purchase.Code),
			})
			log.Printf("📊 PPh 21: Credit to %s (%s) = %.2f", acc.Name, acc.Code, purchase.PPh21Amount)
		} else {
			log.Printf("⚠️ Missing PPh 21 Payable account 2111: %v", err)
		}
	}

	// PPh 23 Payable (2112) - Credit jika ada
	if purchase.PPh23Amount > 0 {
		if acc, err := resolveByCode("2112"); err == nil {
			journalItems = append(journalItems, models.SimpleSSOTJournalItem{
				JournalID:   ssotEntry.ID,
				AccountID:   acc.ID,
				AccountCode: acc.Code,
				AccountName: acc.Name,
				Debit:       0,
				Credit:      purchase.PPh23Amount,
				Description: fmt.Sprintf("PPh 23 withholding for Purchase #%s", purchase.Code),
			})
			log.Printf("📊 PPh 23: Credit to %s (%s) = %.2f", acc.Name, acc.Code, purchase.PPh23Amount)
		} else {
			log.Printf("⚠️ Missing PPh 23 Payable account 2112: %v", err)
		}
	}

	// Create all journal items
	for _, item := range journalItems {
		if err := dbToUse.Create(&item).Error; err != nil {
			return fmt.Errorf("failed to create journal item: %v", err)
		}
	}

	// Update COA balances
	for _, item := range journalItems {
		if err := s.updateCOABalance(dbToUse, item.AccountID, item.Debit, item.Credit); err != nil {
			log.Printf("⚠️ Warning: Failed to update COA balance for account %d: %v", item.AccountID, err)
			// Continue processing, don't fail the entire transaction
		}
	}

	log.Printf("✅ Successfully created journal entries for Purchase #%d", purchase.ID)
	return nil
}

// UpdatePurchaseJournal updates journal entries when purchase is updated
func (s *PurchaseJournalServiceV2) UpdatePurchaseJournal(purchase *models.Purchase, oldStatus string, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Check if we need to create or delete journal based on status change
	oldShouldPost := s.ShouldPostToJournal(oldStatus)
	newShouldPost := s.ShouldPostToJournal(purchase.Status)

	if !oldShouldPost && newShouldPost {
		// Status changed from DRAFT/PENDING to APPROVED/COMPLETED/PAID - Create journal
		log.Printf("📈 Status changed from %s to %s - Creating journal entries", oldStatus, purchase.Status)
		return s.CreatePurchaseJournal(purchase, dbToUse)
	} else if oldShouldPost && !newShouldPost {
		// Status changed from APPROVED/COMPLETED/PAID to DRAFT/PENDING - Delete journal
		log.Printf("📉 Status changed from %s to %s - Removing journal entries", oldStatus, purchase.Status)
		return s.DeletePurchaseJournal(purchase.ID, dbToUse)
	} else if oldShouldPost && newShouldPost {
		// Both statuses require journal - Update existing
		log.Printf("🔄 Updating existing journal entries for Purchase #%d", purchase.ID)
		
		// Delete old entries
		if err := s.DeletePurchaseJournal(purchase.ID, dbToUse); err != nil {
			return err
		}
		
		// Create new entries
		return s.CreatePurchaseJournal(purchase, dbToUse)
	}

	log.Printf("ℹ️ No journal update needed for Purchase #%d (Status: %s)", purchase.ID, purchase.Status)
	return nil
}

// DeletePurchaseJournal deletes all journal entries for a purchase
func (s *PurchaseJournalServiceV2) DeletePurchaseJournal(purchaseID uint, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Find SSOT journal entry
	var ssotJournal models.SimpleSSOTJournal
	if err := dbToUse.Where("transaction_type = ? AND transaction_id = ?", "PURCHASE", purchaseID).
		First(&ssotJournal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ℹ️ No journal entry found for Purchase #%d", purchaseID)
			return nil
		}
		return fmt.Errorf("failed to find journal entry: %v", err)
	}

	// Get all journal items for balance reversal
	var journalItems []models.SimpleSSOTJournalItem
	if err := dbToUse.Where("journal_id = ?", ssotJournal.ID).Find(&journalItems).Error; err != nil {
		return fmt.Errorf("failed to find journal items: %v", err)
	}

	// Reverse COA balances
	for _, item := range journalItems {
		// Reverse the effect: subtract debit, add credit back
		if err := s.updateCOABalance(dbToUse, item.AccountID, -item.Debit, -item.Credit); err != nil {
			log.Printf("⚠️ Warning: Failed to reverse COA balance for account %d: %v", item.AccountID, err)
		}
	}

	// Delete journal items
	if err := dbToUse.Where("journal_id = ?", ssotJournal.ID).Delete(&models.SimpleSSOTJournalItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete journal items: %v", err)
	}

	// Delete journal entry
	if err := dbToUse.Delete(&ssotJournal).Error; err != nil {
		return fmt.Errorf("failed to delete journal entry: %v", err)
	}

	log.Printf("🗑️ Deleted journal entries for Purchase #%d", purchaseID)
	return nil
}

// CreatePurchasePaymentJournal creates journal entries for purchase payment
func (s *PurchaseJournalServiceV2) CreatePurchasePaymentJournal(payment *models.PurchasePayment, purchase *models.Purchase, tx *gorm.DB) error {
	// Only create journal if purchase status allows posting
	if !s.ShouldPostToJournal(purchase.Status) {
		log.Printf("⚠️ Skipping payment journal for Purchase #%d with status: %s", purchase.ID, purchase.Status)
		return nil
	}

	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Helper to resolve account by code
	resolveByCode := func(code string) (*models.Account, error) {
		var acc models.Account
		if err := dbToUse.Where("code = ?", code).First(&acc).Error; err != nil {
			return nil, fmt.Errorf("account code %s not found: %v", code, err)
		}
		return &acc, nil
	}

	// Create SSOT Journal Entry for payment
	ssotEntry := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("PURCHASE_PAYMENT-%d", payment.ID),
		TransactionType:   "PURCHASE_PAYMENT",
		TransactionID:     payment.ID,
		TransactionNumber: payment.PaymentNumber,
		Date:              payment.Date,
		Description:       fmt.Sprintf("Payment for Purchase #%s", purchase.Code),
		TotalAmount:       payment.Amount,
		Status:            "POSTED",
		CreatedBy:         payment.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := dbToUse.Create(ssotEntry).Error; err != nil {
		return fmt.Errorf("failed to create payment journal: %v", err)
	}

	// Prepare journal items
	var journalItems []models.SimpleSSOTJournalItem

	// DEBIT: Reduce Hutang Usaha (2101)
	if payable, err := resolveByCode("2101"); err == nil {
		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   payable.ID,
			AccountCode: payable.Code,
			AccountName: payable.Name,
			Debit:       payment.Amount,
			Credit:      0,
			Description: fmt.Sprintf("Reduce payable for Purchase #%s", purchase.Code),
		})
	}

	// CREDIT: Cash/Bank based on payment method
	var creditAcc *models.Account
	method := strings.ToUpper(strings.TrimSpace(payment.Method))
	
	if payment.CashBankID != nil && *payment.CashBankID > 0 {
		// Use specific cash/bank account when provided
		var cashBank models.CashBank
		if err := dbToUse.First(&cashBank, *payment.CashBankID).Error; err == nil {
			var tmp models.Account
			if e2 := dbToUse.First(&tmp, cashBank.AccountID).Error; e2 == nil {
				creditAcc = &tmp
			}
		}
	}
	if creditAcc == nil {
		if method == "BANK" || strings.HasPrefix(method, "BANK") {
			// Default bank account if method indicates bank but cash_bank_id missing
			if acc, err := resolveByCode("1102"); err == nil { creditAcc = acc }
		} else {
			// Default cash account
			if acc, err := resolveByCode("1101"); err == nil { creditAcc = acc }
		}
	}
	if creditAcc != nil {
		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   creditAcc.ID,
			AccountCode: creditAcc.Code,
			AccountName: creditAcc.Name,
			Debit:       0,
			Credit:      payment.Amount,
			Description: fmt.Sprintf("Payment made for Purchase #%s", purchase.Code),
		})
	}

	// Create all journal items
	for _, item := range journalItems {
		if err := dbToUse.Create(&item).Error; err != nil {
			return fmt.Errorf("failed to create payment journal item: %v", err)
		}
	}

	// Update COA balances
	for _, item := range journalItems {
		if err := s.updateCOABalance(dbToUse, item.AccountID, item.Debit, item.Credit); err != nil {
			log.Printf("⚠️ Warning: Failed to update COA balance for account %d: %v", item.AccountID, err)
		}
	}

	log.Printf("✅ Created payment journal for Purchase Payment #%d (Amount: %.2f)", payment.ID, payment.Amount)
	return nil
}

// Helper function to update COA balance
func (s *PurchaseJournalServiceV2) updateCOABalance(db *gorm.DB, accountID uint, debit, credit float64) error {
	var coa models.COA
	if err := db.First(&coa, accountID).Error; err != nil {
		return fmt.Errorf("COA account %d not found: %v", accountID, err)
	}

	// Update balance based on account type
	netChange := debit - credit
	
	switch coa.Type {
	case "ASSET", "EXPENSE":
		// Assets and Expenses increase with debit
		coa.Balance += netChange
	case "LIABILITY", "EQUITY", "REVENUE":
		// Liabilities, Equity, and Revenue increase with credit
		coa.Balance -= netChange
	}

	return db.Save(&coa).Error
}

// GetDisplayBalance returns the balance formatted for frontend display
func (s *PurchaseJournalServiceV2) GetDisplayBalance(accountID uint) (float64, error) {
	var coa models.COA
	if err := s.db.First(&coa, accountID).Error; err != nil {
		return 0, err
	}

	// For LIABILITY accounts (like Hutang Usaha), 
	// return positive value for display
	if coa.Type == "LIABILITY" {
		return -coa.Balance, nil // Convert negative to positive for display
	}

	return coa.Balance, nil
}