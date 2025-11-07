package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	fmt.Println("🔧 Fixing CashBank-COA Data Integrity Issues...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Initialize services
	accountingService := services.NewCashBankAccountingService(db)
	validationService := services.NewCashBankValidationService(db, accountingService)
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	enhancedService := services.NewCashBankEnhancedService(db, cashBankRepo, accountRepo)
	
	fmt.Println("✅ Services initialized")
	
	// 1. Check current sync status
	fmt.Println("\n🔍 STEP 1: Checking current sync status...")
	syncStatus, err := validationService.GetSyncStatus()
	if err != nil {
		log.Fatalf("❌ Failed to get sync status: %v", err)
	}
	
	fmt.Printf("📊 Current Status: %s\n", syncStatus["status"])
	fmt.Printf("💰 Total CashBanks: %v\n", syncStatus["total_cash_banks"])
	fmt.Printf("🔗 Linked CashBanks: %v\n", syncStatus["linked_cash_banks"])
	fmt.Printf("❗ Discrepancies: %v\n", syncStatus["discrepancies_count"])
	
	if discrepancies, ok := syncStatus["discrepancies"].([]interface{}); ok && len(discrepancies) > 0 {
		fmt.Printf("\n🚨 Found %d discrepancies to fix:\n", len(discrepancies))
		for i, disc := range discrepancies {
			if discMap, ok := disc.(map[string]interface{}); ok {
				fmt.Printf("   %d. %s (%s): CB=%.2f, COA=%.2f, Issue=%s\n", 
					i+1, 
					discMap["cash_bank_name"],
					discMap["cash_bank_code"], 
					discMap["cash_bank_balance"],
					discMap["coa_balance"],
					discMap["issue"])
			}
		}
	}
	
	// 2. Auto-fix discrepancies
	fmt.Println("\n🔧 STEP 2: Auto-fixing discrepancies...")
	fixedCount, err := validationService.AutoFixDiscrepancies()
	if err != nil {
		log.Printf("⚠️ Auto-fix encountered errors: %v", err)
	} else {
		fmt.Printf("✅ Auto-fixed %d discrepancies\n", fixedCount)
	}
	
	// 3. Manual sync all balances using enhanced service
	fmt.Println("\n🔄 STEP 3: Running manual sync for all balances...")
	err = enhancedService.SyncAllBalances()
	if err != nil {
		log.Printf("⚠️ Manual sync encountered errors: %v", err)
	} else {
		fmt.Println("✅ Manual sync completed")
	}
	
	// 4. Validate all CashBank integrity individually
	fmt.Println("\n🔍 STEP 4: Validating individual CashBank integrity...")
	
	var cashBankIDs []uint
	db.Model(&struct{}{}).Table("cash_banks").
		Where("deleted_at IS NULL AND is_active = true").
		Pluck("id", &cashBankIDs)
	
	for _, cbID := range cashBankIDs {
		discrepancy, err := validationService.ValidateCashBankIntegrity(cbID)
		if err != nil {
			fmt.Printf("   ⚠️ Error validating CashBank %d: %v\n", cbID, err)
			continue
		}
		
		if discrepancy.Issue != "SYNC_OK" {
			fmt.Printf("   🔧 Fixing CashBank %s (ID: %d)...\n", discrepancy.CashBankName, cbID)
			
			// Try to recalculate balance
			err = enhancedService.RecalculateBalance(cbID)
			if err != nil {
				fmt.Printf("      ❌ Failed to recalculate: %v\n", err)
			} else {
				fmt.Printf("      ✅ Balance recalculated\n")
			}
		} else {
			fmt.Printf("   ✅ %s is properly synced\n", discrepancy.CashBankName)
		}
	}
	
	// 5. Final validation
	fmt.Println("\n🏁 STEP 5: Final validation...")
	finalStatus, err := validationService.GetSyncStatus()
	if err != nil {
		log.Fatalf("❌ Failed to get final sync status: %v", err)
	}
	
	fmt.Printf("\n📊 FINAL RESULTS:\n")
	fmt.Printf("   Status: %s\n", finalStatus["status"])
	fmt.Printf("   Total CashBanks: %v\n", finalStatus["total_cash_banks"])
	fmt.Printf("   Linked CashBanks: %v\n", finalStatus["linked_cash_banks"])
	fmt.Printf("   Remaining Discrepancies: %v\n", finalStatus["discrepancies_count"])
	
	if finalStatus["status"] == "healthy" {
		fmt.Println("\n🎉 ALL DATA INTEGRITY ISSUES FIXED!")
		fmt.Println("✅ 100% CashBank-COA synchronization achieved!")
	} else {
		fmt.Printf("\n⚠️ Some issues remain. Discrepancy count: %v\n", finalStatus["discrepancies_count"])
		if remaining, ok := finalStatus["discrepancies"].([]interface{}); ok && len(remaining) > 0 {
			fmt.Println("\nRemaining issues:")
			for i, disc := range remaining {
				if discMap, ok := disc.(map[string]interface{}); ok {
					fmt.Printf("   %d. %s: %s\n", i+1, discMap["cash_bank_name"], discMap["issue"])
				}
			}
		}
	}
	
	// 6. Test database triggers by creating a test transaction
	fmt.Println("\n🧪 STEP 6: Testing database triggers...")
	
	// Get first active cash bank
	var firstCashBankID uint
	db.Model(&struct{}{}).Table("cash_banks").
		Where("deleted_at IS NULL AND is_active = true AND account_id > 0").
		Order("id").
		Pluck("id", &firstCashBankID)
	
	if firstCashBankID > 0 {
		fmt.Printf("Testing trigger with CashBank ID: %d\n", firstCashBankID)
		
		// Get balance before
		var balanceBefore float64
		db.Model(&struct{}{}).Table("cash_banks").
			Where("id = ?", firstCashBankID).
			Pluck("balance", &balanceBefore)
		
		// Create a small test transaction
		testAmount := 1.00
		err := accountingService.CreateTransactionWithJournal(
			firstCashBankID,
			testAmount,
			"TEST",
			1,
			"Testing database trigger functionality",
			1, // Assuming account ID 1 exists
		)
		
		if err != nil {
			fmt.Printf("   ❌ Test transaction failed: %v\n", err)
		} else {
			// Check if balance updated automatically
			var balanceAfter float64
			db.Model(&struct{}{}).Table("cash_banks").
				Where("id = ?", firstCashBankID).
				Pluck("balance", &balanceAfter)
			
			expectedBalance := balanceBefore + testAmount
			if balanceAfter == expectedBalance {
				fmt.Printf("   ✅ Database trigger working! Balance: %.2f -> %.2f\n", balanceBefore, balanceAfter)
				
				// Clean up test transaction
				db.Exec("DELETE FROM cash_bank_transactions WHERE notes = 'Testing database trigger functionality' AND cash_bank_id = ?", firstCashBankID)
				fmt.Printf("   🧹 Test transaction cleaned up\n")
			} else {
				fmt.Printf("   ⚠️ Trigger might not be working. Expected: %.2f, Got: %.2f\n", expectedBalance, balanceAfter)
			}
		}
	} else {
		fmt.Println("   ⚠️ No active linked CashBank found for trigger testing")
	}
	
	fmt.Println("\n✅ Data integrity fix process completed!")
}
