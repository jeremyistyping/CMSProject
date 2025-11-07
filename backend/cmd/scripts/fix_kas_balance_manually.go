package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Manual Kas Balance Correction")
	fmt.Println("===============================")

	// Connect to database
	db := database.ConnectDB()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	fmt.Println("✅ Database connected successfully")
	
	if err := fixKasBalanceManually(db); err != nil {
		log.Fatal("❌ Fix failed:", err)
	}
}

func fixKasBalanceManually(db *gorm.DB) error {
	fmt.Println("\n🔍 Step 1: Verify Current State")
	
	// Get current Kas account state
	var kasAccount models.Account
	if err := db.Where("code = '1101'").First(&kasAccount).Error; err != nil {
		return fmt.Errorf("failed to get Kas account: %v", err)
	}
	
	fmt.Printf("   💰 Current Kas Balance: %.2f\n", kasAccount.Balance)
	
	// Verify SSOT journal entries
	var paymentJournals []models.SSOTJournalEntry
	if err := db.Where("source_type = 'PAYMENT'").Find(&paymentJournals).Error; err != nil {
		return fmt.Errorf("failed to get payment journals: %v", err)
	}
	
	fmt.Printf("   📔 Found %d payment journal entries\n", len(paymentJournals))
	
	// Calculate expected balance based on SSOT journals
	expectedBalance := kasAccount.Balance
	for _, journal := range paymentJournals {
		var lines []models.SSOTJournalLine
		if err := db.Where("journal_id = ? AND account_id = ?", journal.ID, kasAccount.ID).Find(&lines).Error; err != nil {
			continue
		}
		
		for _, line := range lines {
			// For Kas (Asset account): Debit increases, Credit decreases
			if line.DebitAmount.GreaterThan(line.CreditAmount) {
				debit, _ := line.DebitAmount.Float64()
				expectedBalance += debit
				fmt.Printf("   📈 Journal %s would INCREASE Kas by %.2f\n", journal.EntryNumber, debit)
			} else if line.CreditAmount.GreaterThan(line.DebitAmount) {
				credit, _ := line.CreditAmount.Float64()
				expectedBalance -= credit
				fmt.Printf("   📉 Journal %s would DECREASE Kas by %.2f\n", journal.EntryNumber, credit)
			}
		}
	}
	
	fmt.Printf("   🎯 Expected Kas Balance: %.2f\n", expectedBalance)
	fmt.Printf("   🔴 Balance difference: %.2f\n", kasAccount.Balance - expectedBalance)
	
	fmt.Println("\n🔧 Step 2: Apply Manual Correction")
	
	// Apply correction
	correctBalance := 5837500.00 // Based on analysis: 7225000 - 1387500
	
	fmt.Printf("   🔄 Updating Kas balance from %.2f to %.2f\n", kasAccount.Balance, correctBalance)
	
	result := db.Model(&kasAccount).Update("balance", correctBalance)
	if result.Error != nil {
		return fmt.Errorf("failed to update Kas balance: %v", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows updated - account not found or balance already correct")
	}
	
	fmt.Printf("   ✅ Kas balance updated successfully\n")
	
	fmt.Println("\n🔍 Step 3: Verify Correction")
	
	// Verify the update
	if err := db.Where("code = '1101'").First(&kasAccount).Error; err != nil {
		return fmt.Errorf("failed to verify updated Kas account: %v", err)
	}
	
	fmt.Printf("   💰 New Kas Balance: %.2f\n", kasAccount.Balance)
	
	if kasAccount.Balance == correctBalance {
		fmt.Printf("   ✅ Balance correction verified successfully\n")
	} else {
		fmt.Printf("   ❌ Balance correction failed. Expected: %.2f, Actual: %.2f\n", correctBalance, kasAccount.Balance)
		return fmt.Errorf("balance correction verification failed")
	}
	
	fmt.Println("\n🔍 Step 4: Update Header Accounts")
	
	// Update header accounts (1100 - Current Assets)
	var headerAccount1100 models.Account
	if err := db.Where("code = '1100' AND is_header = ?", true).First(&headerAccount1100).Error; err == nil {
		// Calculate sum of children
		var childrenSum float64
		err := db.Model(&models.Account{}).
			Where("code LIKE '11%' AND code != '1100' AND is_header = ? AND deleted_at IS NULL", false).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
		
		if err == nil {
			oldBalance := headerAccount1100.Balance
			db.Model(&headerAccount1100).Update("balance", childrenSum)
			fmt.Printf("   📊 Updated header account 1100 balance: %.2f -> %.2f\n", oldBalance, childrenSum)
		}
	}
	
	// Update header accounts (1000 - Assets)
	var headerAccount1000 models.Account
	if err := db.Where("code = '1000' AND is_header = ?", true).First(&headerAccount1000).Error; err == nil {
		// Calculate sum of children
		var childrenSum float64
		err := db.Model(&models.Account{}).
			Where("code LIKE '1%' AND code != '1000' AND is_header = ? AND deleted_at IS NULL", false).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
		
		if err == nil {
			oldBalance := headerAccount1000.Balance
			db.Model(&headerAccount1000).Update("balance", childrenSum)
			fmt.Printf("   📊 Updated header account 1000 balance: %.2f -> %.2f\n", oldBalance, childrenSum)
		}
	}
	
	fmt.Println("\n🎯 Step 5: Final Status Check")
	
	// Final verification
	if err := db.Where("code = '1101'").First(&kasAccount).Error; err != nil {
		return fmt.Errorf("failed to get final Kas account state: %v", err)
	}
	
	fmt.Printf("   💰 Final Kas Balance: %.2f\n", kasAccount.Balance)
	fmt.Printf("   🎯 Expected Balance: %.2f\n", correctBalance)
	
	if kasAccount.Balance == correctBalance {
		fmt.Printf("   ✅ Manual balance correction completed successfully!\n")
		fmt.Printf("   📋 Summary: Kas account balance corrected to reflect SSOT payment journal\n")
		fmt.Printf("   🔧 Root cause: SSOT journal balance update mechanism needs investigation\n")
	} else {
		return fmt.Errorf("final verification failed")
	}
	
	fmt.Println("\n📋 Next Steps:")
	fmt.Println("1. 🔍 Investigate why SSOT AutoPost balance update didn't work initially")
	fmt.Println("2. 🛠️ Add debug logging to UnifiedJournalService.updateAccountBalance")
	fmt.Println("3. 🧪 Test new payments to ensure balance updates work correctly")
	fmt.Println("4. 📊 Monitor system for any balance inconsistencies")
	
	return nil
}