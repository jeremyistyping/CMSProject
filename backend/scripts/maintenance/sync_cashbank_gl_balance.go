package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

type BalanceSyncFix struct {
	CashBankID      uint    `json:"cash_bank_id"`
	CashBankCode    string  `json:"cash_bank_code"`
	CashBankName    string  `json:"cash_bank_name"`
	CashBankBalance float64 `json:"cash_bank_balance"`
	GLAccountID     uint    `json:"gl_account_id"`
	GLCode          string  `json:"gl_code"`
	GLName          string  `json:"gl_name"`
	OldGLBalance    float64 `json:"old_gl_balance"`
	Difference      float64 `json:"difference"`
}

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🔧 Synchronizing Cash/Bank account balances with GL accounts...")
	fmt.Println("================================================================================")

	var results []BalanceSyncFix

	// Get all unsynchronized accounts with details
	err := db.Raw(`
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			cb.account_id as gl_account_id,
			acc.code as gl_code,
			acc.name as gl_name,
			acc.balance as old_gl_balance,
			cb.balance - acc.balance as difference
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND cb.balance != acc.balance
		ORDER BY cb.type, cb.code
	`).Scan(&results).Error

	if err != nil {
		log.Fatal("Failed to query unsynchronized accounts:", err)
	}

	if len(results) == 0 {
		fmt.Println("✅ All cash/bank accounts are already synchronized with GL accounts!")
		return
	}

	fmt.Printf("Found %d unsynchronized accounts to fix:\n\n", len(results))

	// Display what will be fixed
	fmt.Printf("%-15s %-20s %-15s %-15s %-15s %-15s\n", 
		"CB_CODE", "CB_NAME", "CB_BALANCE", "GL_CODE", "OLD_GL_BAL", "DIFFERENCE")
	fmt.Println("================================================================================")

	for _, result := range results {
		fmt.Printf("%-15s %-20s %15.2f %15s %15.2f %15.2f\n",
			result.CashBankCode,
			truncateString(result.CashBankName, 20),
			result.CashBankBalance,
			result.GLCode,
			result.OldGLBalance,
			result.Difference,
		)
	}

	fmt.Println("\n🔄 Starting synchronization process...")

	// Begin transaction for safety
	tx := db.Begin()

	successCount := 0
	errorCount := 0

	for _, result := range results {
		fmt.Printf("Syncing %s (%s) -> %s (%s)... ", 
			result.CashBankCode, result.CashBankName,
			result.GLCode, result.GLName)

		// Update GL account balance to match cash bank balance using raw SQL
		err := tx.Exec(`UPDATE accounts SET balance = ? WHERE id = ?`, 
			result.CashBankBalance, result.GLAccountID).Error

		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			errorCount++
		} else {
			fmt.Printf("✅ SUCCESS (%.2f -> %.2f)\n", 
				result.OldGLBalance, result.CashBankBalance)
			successCount++
		}
	}

	if errorCount > 0 {
		fmt.Printf("\n⚠️  Found %d errors during sync. Rolling back all changes for safety.\n", errorCount)
		tx.Rollback()
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("\n❌ Failed to commit transaction: %v\n", err)
		return
	}

	fmt.Println("\n================================================================================")
	fmt.Printf("🎉 SYNCHRONIZATION COMPLETED SUCCESSFULLY!\n")
	fmt.Printf("✅ Updated %d GL account balances\n", successCount)
	fmt.Printf("❌ Failed: %d\n", errorCount)

	// Verify the sync worked
	fmt.Println("\n🔍 Verifying synchronization...")
	
	var verifyCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND cb.balance != acc.balance
	`).Scan(&verifyCount)

	if verifyCount == 0 {
		fmt.Println("✅ All accounts are now synchronized!")
	} else {
		fmt.Printf("⚠️  Warning: %d accounts are still not synchronized\n", verifyCount)
	}

	fmt.Println("\n💡 IMPORTANT NOTES:")
	fmt.Println("1. GL account balances have been updated to match cash/bank balances")
	fmt.Println("2. This is a one-time sync to fix existing data inconsistency")
	fmt.Println("3. Future transactions should automatically keep both balances in sync")
	fmt.Println("4. Consider implementing automatic journal entries for cash/bank transactions")
	fmt.Println("5. Review your transaction processing to ensure both balances are updated together")
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
