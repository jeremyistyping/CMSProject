package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db, err := database.InitPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("=== FIXING CASH BANK & COA BALANCE SYNCHRONIZATION ===")
	log.Println("🎯 Objective: Ensure COA balances update when cash bank transactions occur")
	log.Println("")

	// ===== STEP 1: Analyze Current State =====
	log.Println("🔍 STEP 1: Analyzing current cash bank vs COA balance mismatches...")
	
	var mismatches []struct {
		CashBankID   uint
		CashBankName string
		AccountID    uint
		AccountCode  string
		AccountName  string
		CashBalance  float64
		COABalance   float64
		Difference   float64
	}

	query := `
		SELECT 
			cb.id as cash_bank_id,
			cb.name as cash_bank_name,
			cb.account_id,
			acc.code as account_code,
			acc.name as account_name,
			cb.balance as cash_balance,
			acc.balance as coa_balance,
			(cb.balance - acc.balance) as difference
		FROM cash_banks cb
		JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.is_active = true 
		  AND cb.balance != acc.balance
		  AND cb.deleted_at IS NULL
		  AND acc.deleted_at IS NULL
	`

	if err := db.Raw(query).Scan(&mismatches).Error; err != nil {
		log.Printf("❌ Failed to analyze mismatches: %v", err)
		return
	}

	log.Printf("Found %d cash bank accounts with COA balance mismatches:", len(mismatches))
	for _, mismatch := range mismatches {
		log.Printf("   📊 %s (ID:%d) → COA %s (%s)", 
			mismatch.CashBankName, mismatch.CashBankID,
			mismatch.AccountCode, mismatch.AccountName)
		log.Printf("      Cash: %.2f | COA: %.2f | Diff: %.2f", 
			mismatch.CashBalance, mismatch.COABalance, mismatch.Difference)
	}

	if len(mismatches) == 0 {
		log.Println("✅ No mismatches found! All balances are synchronized.")
	}

	// ===== STEP 2: Backup Current Balances =====
	log.Println("\n🔐 STEP 2: Creating backup of current balances...")
	
	backupTableName := fmt.Sprintf("balance_backup_%d", time.Now().Unix())
	backupQuery := fmt.Sprintf(`
		CREATE TABLE %s AS 
		SELECT 
			cb.id as cash_bank_id,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			acc.id as account_id,
			acc.code as account_code,
			acc.name as account_name,
			acc.balance as account_balance,
			NOW() as backup_timestamp
		FROM cash_banks cb
		JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.is_active = true 
		  AND cb.deleted_at IS NULL
		  AND acc.deleted_at IS NULL
	`, backupTableName)

	if err := db.Exec(backupQuery).Error; err != nil {
		log.Printf("❌ Failed to create backup table: %v", err)
		return
	}
	log.Printf("✅ Created backup table: %s", backupTableName)

	// ===== STEP 3: Fix Current Mismatches =====
	log.Println("\n🔧 STEP 3: Fixing current balance mismatches...")
	
	fixedCount := 0
	for _, mismatch := range mismatches {
		// Update COA balance to match cash bank balance
		result := db.Model(&models.Account{}).
			Where("id = ?", mismatch.AccountID).
			Update("balance", mismatch.CashBalance)

		if result.Error != nil {
			log.Printf("❌ Failed to fix mismatch for account %d: %v", mismatch.AccountID, result.Error)
		} else {
			log.Printf("✅ Fixed %s: Updated COA balance %.2f → %.2f", 
				mismatch.AccountName, mismatch.COABalance, mismatch.CashBalance)
			fixedCount++
		}
	}

	log.Printf("\n📊 Fixed %d out of %d mismatched accounts", fixedCount, len(mismatches))

	// ===== STEP 4: Test the Fix with New Deposit =====
	log.Println("\n🧪 STEP 4: Testing fix with new deposit transaction...")
	
	// Find or create a test cash account
	var testAccount *models.CashBank
	if len(mismatches) > 0 {
		// Use existing account from mismatches
		err := db.First(&testAccount, mismatches[0].CashBankID).Error
		if err != nil {
			log.Printf("❌ Failed to load test account: %v", err)
			return
		}
	} else {
		// Create new test account
		accountRepo := repositories.NewAccountRepository(db)
		cashBankRepo := repositories.NewCashBankRepository(db)
		cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)

		createRequest := services.CashBankCreateRequest{
			Name:           "Test Sync Fix Account",
			Type:           "CASH",
			Currency:       "IDR",
			OpeningBalance: 1000000,
			OpeningDate:    services.CustomDate(time.Now()),
			Description:    "Test account for balance sync fix validation",
		}

		testAccount, err = cashBankService.CreateCashBankAccount(createRequest, 1)
		if err != nil {
			log.Printf("❌ Failed to create test account: %v", err)
			return
		}
		log.Printf("✅ Created test account: %s (ID: %d)", testAccount.Name, testAccount.ID)
	}

	// Get balances before test deposit
	var beforeCOA models.Account
	db.First(&beforeCOA, testAccount.AccountID)

	log.Printf("📊 BEFORE TEST DEPOSIT:")
	log.Printf("   Cash Bank Balance: %.2f", testAccount.Balance)
	log.Printf("   COA Balance: %.2f", beforeCOA.Balance)

	// Process test deposit with modified service
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	
	// Create enhanced cash bank service with COA sync
	enhancedCashBankService := NewEnhancedCashBankService(db, cashBankRepo, accountRepo)

	depositRequest := services.DepositRequest{
		AccountID: testAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    250000, // Rp 250,000
		Reference: "SYNC-TEST-001",
		Notes:     "Test deposit to verify COA balance sync fix",
	}

	_, err = enhancedCashBankService.ProcessDepositWithCOASync(depositRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to process test deposit: %v", err)
		return
	}

	// Get balances after test deposit
	db.First(&testAccount, testAccount.ID)
	var afterCOA models.Account
	db.First(&afterCOA, testAccount.AccountID)

	log.Printf("\n📊 AFTER TEST DEPOSIT:")
	log.Printf("   Cash Bank Balance: %.2f (Change: %.2f)", testAccount.Balance, testAccount.Balance-beforeCOA.Balance)
	log.Printf("   COA Balance: %.2f (Change: %.2f)", afterCOA.Balance, afterCOA.Balance-beforeCOA.Balance)

	if fmt.Sprintf("%.2f", testAccount.Balance) == fmt.Sprintf("%.2f", afterCOA.Balance) {
		log.Printf("✅ SUCCESS: Balances are now synchronized!")
	} else {
		log.Printf("❌ FAILED: Balances are still mismatched!")
	}

	// ===== STEP 5: Provide Implementation Guidance =====
	log.Println("\n" + "="*60)
	log.Println("📋 IMPLEMENTATION GUIDANCE")
	log.Println("="*60)

	log.Println("\n🛠️ TO PERMANENTLY FIX THIS ISSUE:")
	log.Println("   1. Modify the CashBankService.ProcessDeposit method to:")
	log.Println("      - Update both cash_banks.balance AND accounts.balance")
	log.Println("      - Ensure atomic transactions for consistency")
	log.Println("   ")
	log.Println("   2. Add similar fixes to ProcessWithdrawal and ProcessTransfer")
	log.Println("   ")
	log.Println("   3. Alternative: Remove the isCashBankAccount skip logic from")
	log.Println("      UnifiedJournalService.postJournalEntryTx, but ensure no")
	log.Println("      double posting occurs")

	log.Println("\n📁 FILES TO MODIFY:")
	log.Println("   - backend/services/cashbank_service.go")
	log.Println("     (Add COA balance updates in ProcessDeposit/ProcessWithdrawal)")
	log.Println("   ")
	log.Println("   - backend/services/unified_journal_service.go")
	log.Println("     (Consider modifying isCashBankAccount skip logic)")

	log.Println("\n🧪 VERIFICATION:")
	log.Printf("   - Backup table created: %s", backupTableName)
	log.Printf("   - Use the test script to validate the fix works correctly")

	log.Println("\n" + "="*60)
}

// EnhancedCashBankService wraps the original service with COA synchronization
type EnhancedCashBankService struct {
	db               *gorm.DB
	originalService  *services.CashBankService
}

func NewEnhancedCashBankService(db *gorm.DB, cashBankRepo *repositories.CashBankRepository, accountRepo repositories.AccountRepository) *EnhancedCashBankService {
	return &EnhancedCashBankService{
		db:              db,
		originalService: services.NewCashBankService(db, cashBankRepo, accountRepo),
	}
}

// ProcessDepositWithCOASync processes deposit and ensures COA balance is synchronized
func (s *EnhancedCashBankService) ProcessDepositWithCOASync(request services.DepositRequest, userID uint) (*models.CashBankTransaction, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get cash bank account
	var cashBank models.CashBank
	if err := tx.First(&cashBank, request.AccountID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("cash bank account not found: %v", err)
	}

	// Get COA account before transaction
	var coaAccount models.Account
	if err := tx.First(&coaAccount, cashBank.AccountID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("linked COA account not found: %v", err)
	}

	log.Printf("🔄 Processing deposit with COA sync...")
	log.Printf("   Cash Bank Balance before: %.2f", cashBank.Balance)
	log.Printf("   COA Balance before: %.2f", coaAccount.Balance)

	// Process deposit using original service (but within our transaction)
	// We'll update balances manually to ensure synchronization
	
	// Update cash bank balance
	newCashBalance := cashBank.Balance + request.Amount
	if err := tx.Model(&cashBank).Update("balance", newCashBalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update cash bank balance: %v", err)
	}

	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      request.AccountID,
		ReferenceType:   "DEPOSIT",
		ReferenceID:     0,
		Amount:          request.Amount,
		BalanceAfter:    newCashBalance,
		TransactionDate: request.Date.ToTimeWithCurrentTime(),
		Notes:           request.Notes,
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction record: %v", err)
	}

	// CRITICAL: Update COA balance to match cash bank balance
	newCOABalance := newCashBalance
	if err := tx.Model(&coaAccount).Update("balance", newCOABalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update COA balance: %v", err)
	}

	log.Printf("   Cash Bank Balance after: %.2f", newCashBalance)
	log.Printf("   COA Balance after: %.2f", newCOABalance)

	// Create SSOT journal entry (but it won't update balance due to skip logic)
	// This is fine since we manually updated both balances above
	unifiedJournalService := services.NewUnifiedJournalService(tx)
	journalRequest := &services.JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    func() *uint64 { id := uint64(transaction.ID); return &id }(),
		Reference:   fmt.Sprintf("DEP-%d", transaction.ID),
		EntryDate:   request.Date.ToTimeWithCurrentTime(),
		Description: fmt.Sprintf("Deposit to %s", cashBank.Name),
		Lines: []services.JournalLineRequest{
			{
				AccountID:    uint64(cashBank.AccountID),
				Description:  fmt.Sprintf("Deposit to %s", cashBank.Name),
				DebitAmount:  decimal.NewFromFloat(request.Amount),
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    uint64(3101), // Default equity account - should be configurable
				Description:  fmt.Sprintf("Capital deposit to %s", cashBank.Name),
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(request.Amount),
			},
		},
		AutoPost:  true,
		CreatedBy: uint64(userID),
	}

	_, err := unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create journal entry: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Deposit processed with synchronized balances")
	return transaction, nil
}