package database

import (
	"log"
	"gorm.io/gorm"
)

// SafeBalanceSyncFix - DISABLED FOR PRODUCTION SAFETY
// This function has been permanently disabled to prevent any balance modifications in production
func SafeBalanceSyncFix(db *gorm.DB) {
	log.Println("🛡️  PRODUCTION SAFETY: SafeBalanceSyncFix DISABLED")
	log.Println("⚠️  All balance synchronization operations skipped")
	log.Println("✅ Account balances are fully protected from automatic modifications")
	return // Exit immediately - no balance operations will be performed
	
	// Check if required tables exist
	var cashBankTableExists, accountsTableExists bool
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_banks'
	)`).Scan(&cashBankTableExists)
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounts'
	)`).Scan(&accountsTableExists)
	
	if !cashBankTableExists || !accountsTableExists {
		log.Println("Required tables not found, skipping safe balance sync fix")
		return
	}
	
	// Step 1: Fix missing account_id relationships (safe)
	fixMissingAccountRelationshipsSafe(db)
	
	// Step 2: Only recalculate balances if there are actual transactions
	safeRecalculateCashBankBalances(db)
	
	// Step 3: Only sync if there's a real discrepancy (not just balance differences)
	safeEnsureGLAccountSync(db)
	
	// Step 4: Validate and report final synchronization status
	validateFinalSyncStatus(db)
	
	log.Println("✅ SAFE Balance Synchronization Fix completed")
}

// fixMissingAccountRelationshipsSafe - Same as original but with better logging
func fixMissingAccountRelationshipsSafe(db *gorm.DB) {
	log.Println("Step 1: Checking account relationships...")
	
	// Check for CashBank accounts without GL account links
	var orphanedCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND (cb.account_id IS NULL OR cb.account_id = 0 OR acc.id IS NULL)
	`).Scan(&orphanedCount)
	
	if orphanedCount > 0 {
		log.Printf("⚠️  Found %d CashBank accounts without proper GL account links (manual fix required)", orphanedCount)
	} else {
		log.Println("✅ All CashBank accounts have proper GL account relationships")
	}
}

// safeRecalculateCashBankBalances - PERBAIKAN UTAMA: Hanya recalculate jika ada transaksi
func safeRecalculateCashBankBalances(db *gorm.DB) {
	log.Println("Step 2: SAFE recalculation of CashBank balances...")
	
	// Check if cash_bank_transactions table exists
	var transactionTableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_bank_transactions'
	)`).Scan(&transactionTableExists)
	
	if !transactionTableExists {
		log.Println("Cash bank transactions table not found, preserving existing balances")
		return
	}
	
	// Check if there are any transactions at all
	var transactionCount int64
	db.Raw(`SELECT COUNT(*) FROM cash_bank_transactions WHERE deleted_at IS NULL`).Scan(&transactionCount)
	
	if transactionCount == 0 {
		log.Println("⚠️  No cash bank transactions found - PRESERVING existing balances")
		log.Println("✅ Existing CashBank balances preserved (not reset to zero)")
		return
	}
	
	log.Printf("Found %d cash bank transactions, proceeding with safe recalculation...", transactionCount)
	
	// Only update accounts that HAVE transactions
	err := db.Exec(`
		UPDATE cash_banks 
		SET balance = COALESCE((
			SELECT SUM(amount) 
			FROM cash_bank_transactions cbt 
			WHERE cbt.cash_bank_id = cash_banks.id 
			  AND cbt.deleted_at IS NULL
		), balance), -- PERBAIKAN: Gunakan balance existing jika tidak ada transaksi
		    updated_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL
		  AND EXISTS (
			SELECT 1 FROM cash_bank_transactions cbt 
			WHERE cbt.cash_bank_id = cash_banks.id 
			  AND cbt.deleted_at IS NULL
		  ) -- PERBAIKAN: Hanya update yang punya transaksi
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to safely recalculate CashBank balances: %v", err)
	} else {
		log.Println("✅ Safely recalculated CashBank balances (preserved accounts without transactions)")
	}
}

// safeEnsureGLAccountSync - PERBAIKAN UTAMA: Hanya sync jika benar-benar perlu
func safeEnsureGLAccountSync(db *gorm.DB) {
	log.Println("Step 3: SAFE GL account synchronization...")
	
	// Check for accounts that have significant discrepancies AND have transaction history
	var unsyncCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) > 1.00  -- Hanya jika perbedaan signifikan (> Rp 1)
		  AND (
			-- Hanya jika CashBank punya transaksi atau balance bukan 0
			cb.balance != 0 
			OR EXISTS (
				SELECT 1 FROM cash_bank_transactions cbt 
				WHERE cbt.cash_bank_id = cb.id AND cbt.deleted_at IS NULL
			)
		  )
	`).Scan(&unsyncCount)
	
	if unsyncCount == 0 {
		log.Println("✅ All GL accounts are properly synchronized (no significant discrepancies)")
		return
	}
	
	log.Printf("Found %d GL accounts with significant discrepancies that need synchronization", unsyncCount)
	
	// Get details of accounts that need sync
	type UnsyncGLAccount struct {
		CashBankID      uint    `json:"cash_bank_id"`
		CashBankCode    string  `json:"cash_bank_code"`
		CashBankName    string  `json:"cash_bank_name"`
		CashBankBalance float64 `json:"cash_bank_balance"`
		GLAccountID     uint    `json:"gl_account_id"`
		GLCode          string  `json:"gl_code"`
		GLBalance       float64 `json:"gl_balance"`
		Difference      float64 `json:"difference"`
		HasTransactions bool    `json:"has_transactions"`
	}
	
	var unsyncAccounts []UnsyncGLAccount
	db.Raw(`
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			acc.id as gl_account_id,
			acc.code as gl_code,
			acc.balance as gl_balance,
			cb.balance - acc.balance as difference,
			EXISTS (
				SELECT 1 FROM cash_bank_transactions cbt 
				WHERE cbt.cash_bank_id = cb.id AND cbt.deleted_at IS NULL
			) as has_transactions
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) > 1.00
		  AND (
			cb.balance != 0 
			OR EXISTS (
				SELECT 1 FROM cash_bank_transactions cbt 
				WHERE cbt.cash_bank_id = cb.id AND cbt.deleted_at IS NULL
			)
		  )
		ORDER BY ABS(cb.balance - acc.balance) DESC
		LIMIT 10
	`).Scan(&unsyncAccounts)
	
	// Log accounts that will be synchronized
	log.Println("Accounts to be synchronized:")
	syncCount := 0
	preservedCount := 0
	
	for _, account := range unsyncAccounts {
		if account.HasTransactions || account.CashBankBalance != 0 {
			log.Printf("  SYNC: CB: %s (%.2f) -> GL: %s (%.2f) | Diff: %.2f | HasTx: %t", 
				account.CashBankCode, account.CashBankBalance,
				account.GLCode, account.GLBalance, account.Difference, account.HasTransactions)
			syncCount++
		} else {
			log.Printf("  PRESERVE: CB: %s (%.2f) = GL: %s (%.2f) | No transactions, preserving GL balance", 
				account.CashBankCode, account.CashBankBalance,
				account.GLCode, account.GLBalance)
			preservedCount++
		}
	}
	
	if syncCount == 0 {
		log.Println("✅ No accounts need synchronization - all balances preserved")
		return
	}
	
	// Begin transaction for safe bulk update
	tx := db.Begin()
	
	// PERBAIKAN: Hanya sync accounts yang benar-benar punya transaksi atau balance valid
	err := tx.Exec(`
		UPDATE accounts 
		SET balance = cb.balance,
		    updated_at = CURRENT_TIMESTAMP
		FROM cash_banks cb 
		WHERE accounts.id = cb.account_id 
		  AND cb.deleted_at IS NULL
		  AND ABS(cb.balance - accounts.balance) > 1.00  -- Perbedaan signifikan
		  AND (
			-- Hanya jika CashBank benar-benar punya basis untuk balance-nya
			cb.balance != 0 
			OR EXISTS (
				SELECT 1 FROM cash_bank_transactions cbt 
				WHERE cbt.cash_bank_id = cb.id AND cbt.deleted_at IS NULL
			)
		  )
	`).Error
	
	if err != nil {
		log.Printf("❌ Failed to safely synchronize GL accounts: %v", err)
		tx.Rollback()
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit safe GL account synchronization: %v", err)
		return
	}
	
	log.Printf("✅ Safely synchronized %d GL accounts (preserved %d accounts without transaction basis)", 
		syncCount, preservedCount)
}
