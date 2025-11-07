package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🔧 PERBAIKAN HISTORICAL BALANCE SYNC")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Find all mismatched balances
	fmt.Println("\n📊 STEP 1: Mencari semua balance yang tidak sinkron...")
	
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
		ORDER BY ABS(cb.balance - acc.balance) DESC
	`

	if err := db.Raw(query).Scan(&mismatches).Error; err != nil {
		log.Printf("❌ Failed to find mismatches: %v", err)
		return
	}

	fmt.Printf("📋 Found %d accounts with balance mismatches:\n", len(mismatches))

	if len(mismatches) == 0 {
		fmt.Println("✅ Semua balance sudah sinkron! Tidak perlu perbaikan.")
		return
	}

	// Step 2: Show mismatches
	for i, mismatch := range mismatches {
		fmt.Printf("\n%d. ❌ %s (ID:%d)\n", i+1, mismatch.CashBankName, mismatch.CashBankID)
		fmt.Printf("   COA Account: %s (%s)\n", mismatch.AccountName, mismatch.AccountCode)
		fmt.Printf("   Cash Balance: %.2f\n", mismatch.CashBalance)
		fmt.Printf("   COA Balance: %.2f\n", mismatch.COABalance)
		fmt.Printf("   Difference: %.2f\n", mismatch.Difference)
	}

	// Step 3: Create backup
	fmt.Println("\n🔐 STEP 2: Membuat backup balance saat ini...")
	
	backupTable := "balance_backup_historical_fix"
	backupQuery := fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
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
	`, backupTable, backupTable)

	if err := db.Exec(backupQuery).Error; err != nil {
		log.Printf("❌ Failed to create backup: %v", err)
		return
	}
	fmt.Printf("✅ Backup created: %s\n", backupTable)

	// Step 4: Fix mismatches
	fmt.Println("\n🔧 STEP 3: Memperbaiki balance mismatch...")
	
	fixedCount := 0
	for _, mismatch := range mismatches {
		fmt.Printf("\n🔧 Fixing %s...\n", mismatch.CashBankName)
		
		// Update COA balance to match cash bank balance
		result := db.Model(&models.Account{}).
			Where("id = ?", mismatch.AccountID).
			Update("balance", mismatch.CashBalance)

		if result.Error != nil {
			fmt.Printf("❌ Failed to fix %s: %v\n", mismatch.CashBankName, result.Error)
		} else {
			fmt.Printf("✅ Fixed %s: COA balance %.2f → %.2f\n", 
				mismatch.CashBankName, mismatch.COABalance, mismatch.CashBalance)
			fixedCount++
		}
	}

	// Step 5: Verification
	fmt.Printf("\n📊 STEP 4: Verifikasi hasil perbaikan...\n")

	var remainingMismatches int64
	db.Raw(`
		SELECT COUNT(*)
		FROM cash_banks cb
		JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.is_active = true 
		  AND cb.balance != acc.balance
		  AND cb.deleted_at IS NULL
		  AND acc.deleted_at IS NULL
	`).Scan(&remainingMismatches)

	// Final results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 HASIL PERBAIKAN HISTORICAL BALANCE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📋 Total accounts yang diperbaiki: %d dari %d\n", fixedCount, len(mismatches))
	fmt.Printf("📋 Balance mismatch yang tersisa: %d\n", remainingMismatches)
	
	if remainingMismatches == 0 {
		fmt.Println("🎉 SEMUA BALANCE SUDAH SINKRON!")
	} else {
		fmt.Println("⚠️  Masih ada balance yang tidak sinkron")
	}

	fmt.Printf("💾 Backup data tersimpan di: %s\n", backupTable)
	fmt.Println("\n✅ Perbaikan historical balance selesai!")
	fmt.Println(strings.Repeat("=", 80))
}