package services

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// PurchasePaymentCOASyncService handles COA balance synchronization after purchase payments
type PurchasePaymentCOASyncService struct {
	db          *gorm.DB
	accountRepo repositories.AccountRepository
}

// NewPurchasePaymentCOASyncService creates a new instance
func NewPurchasePaymentCOASyncService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
) *PurchasePaymentCOASyncService {
	return &PurchasePaymentCOASyncService{
		db:          db,
		accountRepo: accountRepo,
	}
}

// SyncCOABalanceAfterPayment ensures COA balance is updated after purchase payment
// This fixes the issue where UnifiedJournalService creates SSOT entries but doesn't update COA balance
func (s *PurchasePaymentCOASyncService) SyncCOABalanceAfterPayment(
	purchaseID uint,
	paymentAmount float64,
	bankAccountID uint,
	userID uint,
	reference string,
	notes string,
) error {
	log.Printf("🔧 Starting COA balance sync after payment for purchase %d, amount %.2f", purchaseID, paymentAmount)

	// Get purchase details
	var purchase models.Purchase
	if err := s.db.Preload("Vendor").First(&purchase, purchaseID).Error; err != nil {
		return fmt.Errorf("failed to get purchase: %v", err)
	}

	// Create fallback SimpleSSOTJournal entry to ensure COA balance is updated
	return s.createPaymentSimpleSSOTJournal(purchase, paymentAmount, bankAccountID, userID, reference, notes)
}

// createPaymentSimpleSSOTJournal creates a Simple SSOT journal entry for payment to update COA balance
func (s *PurchasePaymentCOASyncService) createPaymentSimpleSSOTJournal(
	purchase models.Purchase,
	paymentAmount float64,
	bankAccountID uint,
	userID uint,
	reference string,
	notes string,
) error {
	log.Printf("📝 Creating Simple SSOT journal entry for payment - purchase %s, amount %.2f", purchase.Code, paymentAmount)

	// Helper to get account ID by code
	getAccID := func(code string) (uint, error) {
		acc, err := s.accountRepo.GetAccountByCode(code)
		if err != nil {
			return 0, fmt.Errorf("account %s not found: %v", code, err)
		}
		return acc.ID, nil
	}

	// Get required accounts
	apID, err := getAccID("2101") // Hutang Usaha
	if err != nil {
		return err
	}

	// Resolve bank account ID
	var bankAccountGLID uint
	if bankAccountID > 0 {
		var cashBank models.CashBank
		if err := s.db.Select("account_id").First(&cashBank, bankAccountID).Error; err == nil && cashBank.AccountID > 0 {
			bankAccountGLID = cashBank.AccountID
		}
	}
	if bankAccountGLID == 0 {
		// Default to Kas (1101)
		bankAccountGLID, err = getAccID("1101")
		if err != nil {
			return err
		}
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create Simple SSOT Journal Entry
	ssotEntry := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("PURCHASE_PAYMENT-%d-%d", purchase.ID, time.Now().Unix()),
		TransactionType:   "PURCHASE_PAYMENT",
		TransactionID:     purchase.ID,
		TransactionNumber: reference,
		Date:              time.Now(),
		Description:       fmt.Sprintf("Payment for Purchase #%s - %s (%.2f)", purchase.Code, purchase.Vendor.Name, paymentAmount),
		TotalAmount:       paymentAmount,
		Status:            "POSTED",
		CreatedBy:         userID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := tx.Create(ssotEntry).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create Simple SSOT journal entry: %v", err)
	}

	// Create journal items
	journalItems := []models.SimpleSSOTJournalItem{
		// DEBIT: Reduce Hutang Usaha (2101)
		{
			JournalID:   ssotEntry.ID,
			AccountID:   apID,
			AccountCode: "2101",
			AccountName: "Hutang Usaha",
			Debit:       paymentAmount,
			Credit:      0,
			Description: fmt.Sprintf("Reduce payable for Purchase #%s", purchase.Code),
		},
		// CREDIT: Bank/Cash account
		{
			JournalID:   ssotEntry.ID,
			AccountID:   bankAccountGLID,
			AccountCode: s.getAccountCode(bankAccountGLID),
			AccountName: s.getAccountName(bankAccountGLID),
			Debit:       0,
			Credit:      paymentAmount,
			Description: fmt.Sprintf("Payment for Purchase #%s", purchase.Code),
		},
	}

	// Create journal items
	for _, item := range journalItems {
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create journal item: %v", err)
		}
	}

	// Update COA balances
	for _, item := range journalItems {
		if err := s.updateCOABalance(tx, item.AccountID, item.Debit, item.Credit); err != nil {
			log.Printf("⚠️ Warning: Failed to update COA balance for account %d: %v", item.AccountID, err)
			// Continue processing - don't fail the entire transaction for balance update issues
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit Simple SSOT journal transaction: %v", err)
	}

	log.Printf("✅ Simple SSOT journal entry created and COA balances updated for payment")
	return nil
}

// updateCOABalance updates account balance based on journal entry
func (s *PurchasePaymentCOASyncService) updateCOABalance(tx *gorm.DB, accountID uint, debit, credit float64) error {
	var account models.Account
	if err := tx.First(&account, accountID).Error; err != nil {
		return fmt.Errorf("Account %d not found: %v", accountID, err)
	}

	// Update balance based on account type
	netChange := debit - credit

	switch account.Type {
	case "ASSET", "EXPENSE":
		// Assets and Expenses increase with debit
		account.Balance += netChange
	case "LIABILITY", "EQUITY", "REVENUE":
		// Liabilities, Equity, and Revenue increase with credit
		account.Balance -= netChange
	}

	account.UpdatedAt = time.Now()
	return tx.Save(&account).Error
}

// getAccountCode gets account code by ID
func (s *PurchasePaymentCOASyncService) getAccountCode(accountID uint) string {
	var account models.Account
	if err := s.db.Select("code").First(&account, accountID).Error; err != nil {
		return ""
	}
	return account.Code
}

// getAccountName gets account name by ID
func (s *PurchasePaymentCOASyncService) getAccountName(accountID uint) string {
	var account models.Account
	if err := s.db.Select("name").First(&account, accountID).Error; err != nil {
		return ""
	}
	return account.Name
}

// SyncAllCOABalancesWithCashBanks synchronizes all COA balances with Cash & Bank balances
func (s *PurchasePaymentCOASyncService) SyncAllCOABalancesWithCashBanks() error {
	log.Printf("🔄 Starting full COA balance synchronization with Cash & Bank balances")

	// Update COA balances based on Cash & Bank balances for bank/cash accounts
	query := `
		UPDATE accounts 
		SET balance = cb.balance, updated_at = NOW()
		FROM cash_banks cb
		WHERE accounts.id = cb.account_id
		  AND cb.is_active = true 
		  AND accounts.is_active = true
		  AND accounts.type = 'ASSET'
		  AND (accounts.code LIKE '110%' OR accounts.code = '1101' OR accounts.code = '1102')
		  AND ABS(cb.balance - COALESCE(accounts.balance, 0)) > 0.01
	`

	if err := s.db.Exec(query).Error; err != nil {
		return fmt.Errorf("failed to sync COA balances with Cash & Bank: %v", err)
	}

	log.Printf("✅ COA balance synchronization completed")
	return nil
}

// GetBalanceDiscrepancies returns discrepancies between Cash & Bank and COA balances
func (s *PurchasePaymentCOASyncService) GetBalanceDiscrepancies() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			cb.id as cash_bank_id,
			cb.name as cash_bank_name,
			cb.account_id,
			cb.balance as cash_bank_balance,
			acc.code as account_code,
			acc.name as account_name,
			COALESCE(acc.balance, 0) as account_balance,
			(cb.balance - COALESCE(acc.balance, 0)) as difference,
			acc.type as account_type
		FROM cash_banks cb
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.is_active = true 
		  AND acc.is_active = true
		  AND ABS(cb.balance - COALESCE(acc.balance, 0)) > 0.01
		ORDER BY ABS(cb.balance - COALESCE(acc.balance, 0)) DESC
	`

	var discrepancies []map[string]interface{}
	if err := s.db.Raw(query).Scan(&discrepancies).Error; err != nil {
		return nil, fmt.Errorf("failed to get balance discrepancies: %v", err)
	}

	return discrepancies, nil
}

// FixPaymentCOABalanceIssue is a one-time fix for existing payment issues
func (s *PurchasePaymentCOASyncService) FixPaymentCOABalanceIssue() error {
	log.Printf("🛠️ Running one-time fix for payment COA balance issues")

	// Run the SQL script to fix existing issues
	queries := []string{
		// Update account balance based on Cash & Bank balance for bank/cash accounts
		`UPDATE accounts 
		 SET balance = cb.balance, updated_at = NOW()
		 FROM cash_banks cb
		 WHERE accounts.id = cb.account_id
		   AND cb.is_active = true
		   AND accounts.is_active = true
		   AND accounts.type = 'ASSET'
		   AND (accounts.code LIKE '110%' OR accounts.code = '1101' OR accounts.code = '1102')
		   AND ABS(cb.balance - COALESCE(accounts.balance, 0)) > 0.01`,

		// Update Hutang Usaha balance based on outstanding purchases
		`UPDATE accounts 
		 SET balance = COALESCE((
			SELECT SUM(outstanding_amount) 
			FROM purchases 
			WHERE payment_method = 'CREDIT' 
			  AND status IN ('APPROVED', 'COMPLETED', 'PAID')
		 ), 0), updated_at = NOW()
		 WHERE code = '2101' AND is_active = true`,
	}

	for i, query := range queries {
		log.Printf("Executing fix query %d...", i+1)
		if err := s.db.Exec(query).Error; err != nil {
			log.Printf("⚠️ Warning: Fix query %d failed: %v", i+1, err)
		} else {
			log.Printf("✅ Fix query %d completed successfully", i+1)
		}
	}

	log.Printf("🎉 One-time COA balance fix completed")
	return nil
}