package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🎯 Final Fix for Last CashBank-COA Discrepancy...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Initialize services
	accountingService := services.NewCashBankAccountingService(db)
	validationService := services.NewCashBankValidationService(db, accountingService)
	
	fmt.Println("✅ Services initialized")
	
	// Get detailed discrepancies
	fmt.Println("\n🔍 Identifying remaining discrepancy...")
	discrepancies, err := validationService.FindSyncDiscrepancies()
	if err != nil {
		log.Fatalf("❌ Failed to find discrepancies: %v", err)
	}
	
	if len(discrepancies) == 0 {
		fmt.Println("🎉 No discrepancies found! All accounts are synced!")
		return
	}
	
	fmt.Printf("Found %d discrepancy(ies):\n", len(discrepancies))
	for i, disc := range discrepancies {
		fmt.Printf("\n%d. CashBank: %s (ID: %d)\n", i+1, disc.CashBankName, disc.CashBankID)
		fmt.Printf("   Code: %s\n", disc.CashBankCode)
		fmt.Printf("   COA Account: %s (ID: %d)\n", disc.COAAccountCode, disc.COAAccountID)
		fmt.Printf("   CashBank Balance: %.2f\n", disc.CashBankBalance)
		fmt.Printf("   COA Balance: %.2f\n", disc.COABalance)
		fmt.Printf("   Transaction Sum: %.2f\n", disc.TransactionSum)
		fmt.Printf("   Discrepancy: %.2f\n", disc.Discrepancy)
		fmt.Printf("   Issue: %s\n", disc.Issue)
	}
	
	// Try multiple fix approaches
	for i, disc := range discrepancies {
		fmt.Printf("\n🔧 Fixing discrepancy %d (%s)...\n", i+1, disc.CashBankName)
		
		// Approach 1: Use transaction sum as source of truth
		correctBalance := disc.TransactionSum
		fmt.Printf("   Setting correct balance to: %.2f (from transaction sum)\n", correctBalance)
		
		// Manual database update
		err = db.Transaction(func(tx *gorm.DB) error {
			// Update CashBank balance
			if err := tx.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", correctBalance, disc.CashBankID).Error; err != nil {
				return fmt.Errorf("failed to update cash bank balance: %v", err)
			}
			
			// Update COA balance
			if err := tx.Exec("UPDATE accounts SET balance = ? WHERE id = ?", correctBalance, disc.COAAccountID).Error; err != nil {
				return fmt.Errorf("failed to update COA balance: %v", err)
			}
			
		return nil
		})
		
		if err != nil {
			fmt.Printf("   ❌ Manual fix failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Manual fix completed\n")
		}
	}
	
	// Final verification
	fmt.Println("\n🔍 Final verification...")
	remaining, err := validationService.FindSyncDiscrepancies()
	if err != nil {
		log.Printf("⚠️ Error during final verification: %v", err)
	} else if len(remaining) == 0 {
		fmt.Println("🎉 SUCCESS! All discrepancies resolved!")
		fmt.Println("✅ 100% CashBank-COA synchronization achieved!")
	} else {
		fmt.Printf("⚠️ %d discrepancy(ies) still remain:\n", len(remaining))
		for _, disc := range remaining {
			fmt.Printf("   - %s: %s (Discrepancy: %.2f)\n", disc.CashBankName, disc.Issue, disc.Discrepancy)
		}
	}
}
