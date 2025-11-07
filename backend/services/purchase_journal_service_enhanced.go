package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// PurchaseJournalServiceEnhanced handles purchase journal entries using configurable tax account settings
// Now also supports TaxConfig for tax rates and calculation rules
type PurchaseJournalServiceEnhanced struct {
	db                *gorm.DB
	taxAccountService *TaxAccountService
}

// NewPurchaseJournalServiceEnhanced creates a new instance of PurchaseJournalServiceEnhanced
func NewPurchaseJournalServiceEnhanced(
	db *gorm.DB,
	taxAccountService *TaxAccountService,
) *PurchaseJournalServiceEnhanced {
	return &PurchaseJournalServiceEnhanced{
		db:                db,
		taxAccountService: taxAccountService,
	}
}

// ShouldPostToJournal checks if a status should create journal entries
func (s *PurchaseJournalServiceEnhanced) ShouldPostToJournal(status string) bool {
	// KRITERIA POSTING JURNAL - APPROVED, COMPLETED, atau PAID
	allowedStatuses := []string{"APPROVED", "COMPLETED", "PAID"}
	for _, allowed := range allowedStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

// CreatePurchaseJournal creates journal entries for a purchase using configurable account settings
func (s *PurchaseJournalServiceEnhanced) CreatePurchaseJournal(purchase *models.Purchase, tx *gorm.DB) error {
	// VALIDASI STATUS
	if !s.ShouldPostToJournal(purchase.Status) {
		log.Printf("⚠️ Skipping journal creation for Purchase #%d with status: %s", purchase.ID, purchase.Status)
		return nil
	}

	log.Printf("📝 Creating enhanced journal entries for Purchase #%d (Status: %s, Payment Method: %s)", 
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

	// Prepare journal items using tax account settings
	var journalItems []models.SimpleSSOTJournalItem

	// DEBIT SIDE - Expense/Inventory Accounts
	for _, item := range purchase.PurchaseItems {
		// Use the expense account from the item, or default from settings
		expenseAccountID := item.ExpenseAccountID
		if expenseAccountID == 0 {
			// Use default purchase expense account from settings
			var err error
			expenseAccountID, err = s.taxAccountService.GetAccountID("purchase_expense")
			if err != nil {
				log.Printf("⚠️ Failed to get default purchase expense account: %v", err)
				// Fallback to hardcoded
				expenseAccountID, err = s.taxAccountService.GetAccountByCode("6001")
				if err != nil {
					return fmt.Errorf("failed to find expense account for item %d: %v", item.ID, err)
				}
			}
		}

		var account models.Account
		if err := dbToUse.First(&account, expenseAccountID).Error; err != nil {
			return fmt.Errorf("failed to load expense account %d: %v", expenseAccountID, err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       item.TotalPrice,
			Credit:      0,
			Description: fmt.Sprintf("Purchase - %s", item.Product.Name),
		})
		log.Printf("🛒 Expense: Debit to %s (%s) = %.2f", account.Name, account.Code, item.TotalPrice)
	}

	// DEBIT: PPN Masukan (if applicable) using tax account settings
	// Note: PPN can now be calculated using TaxConfig:
	//   ppnRate, _ := s.taxAccountService.GetPurchaseTaxRate("ppn")
	//   calculatedPPN := s.taxAccountService.CalculateTax(purchase.Subtotal, ppnRate, discountBeforeTax, discount)
	if purchase.PPNAmount > 0 {
		accountID, err := s.taxAccountService.GetAccountID("purchase_input_vat")
		if err != nil {
			log.Printf("⚠️ Failed to get purchase input VAT account from settings: %v", err)
			// Fallback to hardcoded (1240 is the standard PPN Masukan account)
			accountID, err = s.taxAccountService.GetAccountByCode("1240")
			if err != nil {
				return fmt.Errorf("failed to find input VAT account: %v", err)
			}
		}

		var account models.Account
		if err := dbToUse.First(&account, accountID).Error; err != nil {
			return fmt.Errorf("failed to load input VAT account: %v", err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       purchase.PPNAmount,
			Credit:      0,
			Description: "PPN Masukan",
		})
		log.Printf("📊 Input VAT: Debit to %s (%s) = %.2f", account.Name, account.Code, purchase.PPNAmount)
	}

	// CREDIT SIDE - Based on payment method using tax account settings
	switch strings.ToUpper(strings.TrimSpace(purchase.PaymentMethod)) {
	case "CASH":
		// CASH → Credit Kas (from settings)
		accountID, err := s.taxAccountService.GetAccountID("purchase_cash")
		if err != nil {
			log.Printf("⚠️ Failed to get purchase cash account from settings: %v", err)
			// Fallback to hardcoded
			accountID, err = s.taxAccountService.GetAccountByCode("1101")
			if err != nil {
				return fmt.Errorf("failed to find cash account: %v", err)
			}
		}

		var account models.Account
		if err := dbToUse.First(&account, accountID).Error; err != nil {
			return fmt.Errorf("failed to load cash account: %v", err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       0,
			Credit:      purchase.TotalAmount,
			Description: fmt.Sprintf("Cash paid for Purchase #%s", purchase.Code),
		})
		log.Printf("💵 CASH Payment: Credit to %s (%s) = %.2f", account.Name, account.Code, purchase.TotalAmount)

	case "BANK_TRANSFER", "BANK":
		// BANK → Credit Bank (specific account if available, otherwise from settings)
		var accountID uint
		if purchase.BankAccountID != nil && *purchase.BankAccountID > 0 {
			// Use specific bank account
			var cashBank models.CashBank
			if err := dbToUse.First(&cashBank, *purchase.BankAccountID).Error; err == nil {
				accountID = cashBank.AccountID
			}
		}

		if accountID == 0 {
			// Use default bank account from settings
			var err error
			accountID, err = s.taxAccountService.GetAccountID("purchase_bank")
			if err != nil {
				log.Printf("⚠️ Failed to get purchase bank account from settings: %v", err)
				// Fallback to hardcoded
				accountID, err = s.taxAccountService.GetAccountByCode("1102")
				if err != nil {
					return fmt.Errorf("failed to find bank account: %v", err)
				}
			}
		}

		var account models.Account
		if err := dbToUse.First(&account, accountID).Error; err != nil {
			return fmt.Errorf("failed to load bank account: %v", err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       0,
			Credit:      purchase.TotalAmount,
			Description: fmt.Sprintf("Bank transfer for Purchase #%s", purchase.Code),
		})
		log.Printf("🏦 BANK Payment: Credit to %s (%s) = %.2f", account.Name, account.Code, purchase.TotalAmount)

	default:
		// CREDIT or unknown → Credit Hutang Usaha (from settings)
		accountID, err := s.taxAccountService.GetAccountID("purchase_payable")
		if err != nil {
			log.Printf("⚠️ Failed to get purchase payable account from settings: %v", err)
			// Fallback to hardcoded
			accountID, err = s.taxAccountService.GetAccountByCode("2001")
			if err != nil {
				return fmt.Errorf("failed to find payable account: %v", err)
			}
		}

		var account models.Account
		if err := dbToUse.First(&account, accountID).Error; err != nil {
			return fmt.Errorf("failed to load payable account: %v", err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       0,
			Credit:      purchase.TotalAmount,
			Description: fmt.Sprintf("Account payable for Purchase #%s", purchase.Code),
		})
		log.Printf("📋 CREDIT Purchase: Credit to %s (%s) = %.2f", account.Name, account.Code, purchase.TotalAmount)
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

	log.Printf("✅ Successfully created enhanced journal entries for Purchase #%d", purchase.ID)
	return nil
}

// UpdatePurchaseJournal updates journal entries when purchase is updated
func (s *PurchaseJournalServiceEnhanced) UpdatePurchaseJournal(purchase *models.Purchase, oldStatus string, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Check if we need to create or delete journal based on status change
	oldShouldPost := s.ShouldPostToJournal(oldStatus)
	newShouldPost := s.ShouldPostToJournal(purchase.Status)

	if !oldShouldPost && newShouldPost {
		// Status changed to postable status - Create journal
		log.Printf("📈 Status changed from %s to %s - Creating enhanced journal entries", oldStatus, purchase.Status)
		return s.CreatePurchaseJournal(purchase, dbToUse)
	} else if oldShouldPost && !newShouldPost {
		// Status changed from postable status - Delete journal
		log.Printf("📉 Status changed from %s to %s - Removing journal entries", oldStatus, purchase.Status)
		return s.DeletePurchaseJournal(purchase.ID, dbToUse)
	} else if oldShouldPost && newShouldPost {
		// Both statuses require journal - Update existing
		log.Printf("🔄 Updating existing enhanced journal entries for Purchase #%d", purchase.ID)
		
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
func (s *PurchaseJournalServiceEnhanced) DeletePurchaseJournal(purchaseID uint, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	log.Printf("🗑️ Deleting enhanced journal entries for Purchase #%d", purchaseID)

	// Find the SSOT journal entry
	var ssotJournal models.SimpleSSOTJournal
	err := dbToUse.Where("transaction_type = ? AND transaction_id = ?", "PURCHASE", purchaseID).First(&ssotJournal).Error
	if err != nil {
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

	log.Printf("🗑️ Deleted enhanced journal entries for Purchase #%d", purchaseID)
	return nil
}

// CreatePurchasePaymentJournal creates journal entries for purchase payment using configurable accounts
func (s *PurchaseJournalServiceEnhanced) CreatePurchasePaymentJournal(payment *models.PurchasePayment, purchase *models.Purchase, tx *gorm.DB) error {
	// Only create journal if purchase status allows posting
	if !s.ShouldPostToJournal(purchase.Status) {
		log.Printf("⚠️ Skipping payment journal for Purchase #%d with status: %s", purchase.ID, purchase.Status)
		return nil
	}

	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
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

	// DEBIT: Reduce Hutang Usaha using tax account settings
	payableAccountID, err := s.taxAccountService.GetAccountID("purchase_payable")
	if err != nil {
		log.Printf("⚠️ Failed to get purchase payable account from settings: %v", err)
		// Fallback to hardcoded
		payableAccountID, err = s.taxAccountService.GetAccountByCode("2001")
		if err != nil {
			return fmt.Errorf("failed to find payable account: %v", err)
		}
	}

	var payableAccount models.Account
	if err := dbToUse.First(&payableAccount, payableAccountID).Error; err != nil {
		return fmt.Errorf("failed to load payable account: %v", err)
	}

	journalItems = append(journalItems, models.SimpleSSOTJournalItem{
		JournalID:   ssotEntry.ID,
		AccountID:   payableAccount.ID,
		AccountCode: payableAccount.Code,
		AccountName: payableAccount.Name,
		Debit:       payment.Amount,
		Credit:      0,
		Description: fmt.Sprintf("Reduce payable for Purchase #%s", purchase.Code),
	})

	// CREDIT: Cash/Bank based on payment method using tax account settings
	method := strings.ToUpper(strings.TrimSpace(payment.Method))
	var creditAccountID uint
	
	if payment.CashBankID != nil && *payment.CashBankID > 0 {
		// Use specific cash/bank account when provided
		var cashBank models.CashBank
		if err := dbToUse.First(&cashBank, *payment.CashBankID).Error; err == nil {
			creditAccountID = cashBank.AccountID
		}
	}
	
	if creditAccountID == 0 {
		// Use tax account settings
		var err error
		if method == "BANK" || strings.HasPrefix(method, "BANK") || strings.Contains(method, "TRANSFER") {
			// Use bank account from settings
			creditAccountID, err = s.taxAccountService.GetAccountID("purchase_bank")
			if err != nil {
				// Fallback to hardcoded
				creditAccountID, err = s.taxAccountService.GetAccountByCode("1102")
			}
		} else {
			// Use cash account from settings
			creditAccountID, err = s.taxAccountService.GetAccountID("purchase_cash")
			if err != nil {
				// Fallback to hardcoded
				creditAccountID, err = s.taxAccountService.GetAccountByCode("1101")
			}
		}
		
		if err != nil {
			return fmt.Errorf("failed to find payment account: %v", err)
		}
	}

	// Load credit account
	var creditAccount models.Account
	if err := dbToUse.First(&creditAccount, creditAccountID).Error; err != nil {
		return fmt.Errorf("failed to load payment account: %v", err)
	}

	journalItems = append(journalItems, models.SimpleSSOTJournalItem{
		JournalID:   ssotEntry.ID,
		AccountID:   creditAccount.ID,
		AccountCode: creditAccount.Code,
		AccountName: creditAccount.Name,
		Debit:       0,
		Credit:      payment.Amount,
		Description: fmt.Sprintf("Payment for Purchase #%s", purchase.Code),
	})

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

	log.Printf("✅ Created enhanced purchase payment journal for Payment #%d (Amount: %.2f)", payment.ID, payment.Amount)
	return nil
}

// Helper function to update COA balance
func (s *PurchaseJournalServiceEnhanced) updateCOABalance(db *gorm.DB, accountID uint, debit, credit float64) error {
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