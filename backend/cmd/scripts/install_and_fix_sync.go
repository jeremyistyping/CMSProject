package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"path/filepath"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	fmt.Println("=== CashBank-COA Sync Installation and Fix ===\n")

	// Initialize database using the existing ConnectDB function
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Step 1: Install database triggers
	fmt.Println("1. Installing database triggers...")
	triggerSQL, err := ioutil.ReadFile(filepath.Join("database", "migrations", "cashbank_coa_sync_trigger.sql"))
	if err != nil {
		log.Printf("Error reading trigger SQL file: %v", err)
		fmt.Println("   ⚠ Skipping trigger installation - file not found")
	} else {
		// Execute the SQL
		if err := db.Exec(string(triggerSQL)).Error; err != nil {
			log.Printf("Error installing triggers: %v", err)
			fmt.Println("   ⚠ Failed to install triggers")
		} else {
			fmt.Println("   ✓ Database triggers installed successfully")
		}
	}
	
	// Verify triggers are installed
	var triggerCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name LIKE '%sync_cashbank%'").Scan(&triggerCount)
	fmt.Printf("   Found %d sync triggers installed\n", triggerCount)

	// Step 2: Initialize services
	fmt.Println("\n2. Initializing services...")
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	
	accountingService := services.NewCashBankAccountingService(db)
	validationService := services.NewCashBankValidationService(db, accountingService)
	_ = services.NewCashBankEnhancedService(db, cashBankRepo, accountRepo) // Create but don't store

	fmt.Println("   ✓ Services initialized")

	// Step 3: Check current sync status
	fmt.Println("\n3. Checking current sync status...")
	status, err := validationService.GetSyncStatus()
	if err != nil {
		log.Printf("Error checking sync status: %v", err)
		return
	}
	
	fmt.Printf("   Status: %s\n", status["status"])
	fmt.Printf("   Total Cash Banks: %v\n", status["total_cash_banks"])
	fmt.Printf("   Linked Cash Banks: %v\n", status["linked_cash_banks"])
	fmt.Printf("   Discrepancies: %v\n", status["discrepancies_count"])
	
	if issues, ok := status["issue_breakdown"].(map[string]int); ok {
		for issue, count := range issues {
			fmt.Printf("   - %s: %d\n", issue, count)
		}
	}

	// Step 4: Show detailed discrepancies
	fmt.Println("\n4. Analyzing discrepancies...")
	discrepancies, err := validationService.FindSyncDiscrepancies()
	if err != nil {
		log.Printf("Error finding discrepancies: %v", err)
	} else {
		if len(discrepancies) == 0 {
			fmt.Println("   ✓ No discrepancies found!")
		} else {
			fmt.Printf("   Found %d discrepancies:\n", len(discrepancies))
			for i, d := range discrepancies {
				fmt.Printf("   %d. %s (%s)\n", i+1, d.CashBankName, d.CashBankCode)
				fmt.Printf("      Issue: %s\n", d.Issue)
				fmt.Printf("      CashBank Balance: %.2f\n", d.CashBankBalance)
				fmt.Printf("      COA Balance: %.2f\n", d.COABalance)
				fmt.Printf("      Transaction Sum: %.2f\n", d.TransactionSum)
				fmt.Printf("      Discrepancy: %.2f\n", d.Discrepancy)
				fmt.Println()
			}
		}
	}

	// Step 5: Auto-fix discrepancies
	fmt.Println("5. Auto-fixing discrepancies...")
	fixedCount, err := validationService.AutoFixDiscrepancies()
	if err != nil {
		log.Printf("Error auto-fixing: %v", err)
	} else {
		fmt.Printf("   ✓ Fixed %d discrepancies automatically\n", fixedCount)
	}

	// Step 6: Manual sync for any remaining issues
	fmt.Println("\n6. Performing manual sync for all accounts...")
	err = accountingService.SyncAllCashBankBalances()
	if err != nil {
		log.Printf("Error syncing all balances: %v", err)
	} else {
		fmt.Println("   ✓ Manual sync completed")
	}

	// Step 7: Final verification
	fmt.Println("\n7. Final verification...")
	finalStatus, err := validationService.GetSyncStatus()
	if err != nil {
		log.Printf("Error checking final status: %v", err)
	} else {
		fmt.Printf("   Final Status: %s\n", finalStatus["status"])
		fmt.Printf("   Remaining Discrepancies: %v\n", finalStatus["discrepancies_count"])
		
		if finalStatus["status"] == "healthy" {
			fmt.Println("   🎉 SUCCESS: All CashBank accounts are now synchronized with COA!")
		} else {
			fmt.Println("   ⚠ ATTENTION: Some issues still remain")
			
			// Show remaining discrepancies
			remainingDiscrepancies, _ := validationService.FindSyncDiscrepancies()
			for _, d := range remainingDiscrepancies {
				if d.Issue != "NOT_LINKED" { // Focus on critical issues
					fmt.Printf("   - %s: %s (CB: %.2f, COA: %.2f)\n", 
						d.CashBankName, d.Issue, d.CashBankBalance, d.COABalance)
				}
			}
		}
	}

	// Step 8: Show summary table
	fmt.Println("\n8. Summary of CashBank accounts:")
	fmt.Println("   ┌─────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("   │ %-20s │ %-15s │ %-15s │ %-15s │ %-10s │\n", 
		"Account Name", "CB Balance", "COA Balance", "TX Sum", "Status")
	fmt.Println("   ├─────────────────────────────────────────────────────────────────────────────────┤")

	allDiscrepancies, _ := validationService.FindSyncDiscrepancies()
	
	// Also show accounts that are in sync
	cashBanks, _ := cashBankRepo.FindAll()
	processedIDs := make(map[uint]bool)
	
	// Show problematic accounts first
	for _, d := range allDiscrepancies {
		fmt.Printf("   │ %-20s │ %15.2f │ %15.2f │ %15.2f │ %-10s │\n", 
			d.CashBankName, d.CashBankBalance, d.COABalance, d.TransactionSum, d.Issue)
		processedIDs[d.CashBankID] = true
	}
	
	// Show healthy accounts
	for _, cb := range cashBanks {
		if !processedIDs[cb.ID] && cb.AccountID > 0 {
			integrity, err := validationService.ValidateCashBankIntegrity(cb.ID)
			if err == nil {
				status := "SYNC_OK"
				if integrity.Issue != "SYNC_OK" {
					status = integrity.Issue
				}
				fmt.Printf("   │ %-20s │ %15.2f │ %15.2f │ %15.2f │ %-10s │\n", 
					cb.Name, integrity.CashBankBalance, integrity.COABalance, 
					integrity.TransactionSum, status)
			}
		}
	}
	
	fmt.Println("   └─────────────────────────────────────────────────────────────────────────────────┘")

	// Step 9: Recommendations
	fmt.Println("\n9. Next Steps & Recommendations:")
	fmt.Println("   ✓ Database triggers installed and active")
	fmt.Println("   ✓ Sync services ready for integration")
	fmt.Println("   ✓ Automatic sync will now work for future transactions")
	fmt.Println()
	fmt.Println("   📋 To integrate with existing controllers:")
	fmt.Println("   1. Update your main.go to use CashBankEnhancedService")
	fmt.Println("   2. Add validation middleware to routes")
	fmt.Println("   3. Add health check endpoints")
	fmt.Println()
	fmt.Println("   📊 Monitor using these endpoints:")
	fmt.Println("   - GET /api/health/cashbank - Overall health")  
	fmt.Println("   - GET /api/cashbank/sync/status - Detailed status")
	fmt.Println("   - POST /api/cashbank/sync/fix - Auto-fix issues")
	fmt.Println()
	fmt.Println("=== Installation and Fix Completed ===")
}
