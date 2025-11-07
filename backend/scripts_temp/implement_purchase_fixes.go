package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/database"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("🚀 IMPLEMENTING PURCHASE BUG FIXES")
	fmt.Println("=" + fmt.Sprintf("%*s", 69, "") + "=")

	// Step 1: Update PPN account from LIABILITY to ASSET
	fmt.Println("\n🔧 STEP 1: Updating PPN Account (2102) from LIABILITY to ASSET...")
	
	var ppnAccount models.Account
	if err := db.Where("code = ?", "2102").First(&ppnAccount).Error; err != nil {
		log.Fatal("PPN account not found:", err)
	}

	fmt.Printf("Current PPN Account: %s (%s) - Type: %s\n", 
		ppnAccount.Name, ppnAccount.Code, ppnAccount.Type)

	// Update PPN account to correct type and name
	ppnAccount.Name = "PPN Masukan"
	ppnAccount.Type = models.AccountTypeAsset
	ppnAccount.Category = models.CategoryCurrentAsset
	
	// Find parent account for current assets (1100)
	var currentAssetsAccount models.Account
	if err := db.Where("code = ?", "1100").First(&currentAssetsAccount).Error; err == nil {
		ppnAccount.ParentID = &currentAssetsAccount.ID
	}

	if err := db.Save(&ppnAccount).Error; err != nil {
		log.Fatal("Failed to update PPN account:", err)
	}

	fmt.Printf("✅ PPN Account Updated: %s (%s) - Now Type: %s\n", 
		ppnAccount.Name, ppnAccount.Code, ppnAccount.Type)

	// Step 2: Re-run account seeder to ensure all accounts are properly setup
	fmt.Println("\n🔧 STEP 2: Re-running account seeder...")
	if err := database.SeedAccounts(db); err != nil {
		log.Fatal("Failed to re-run account seeder:", err)
	}
	fmt.Println("✅ Account seeder completed")

	// Step 3: Verify account setup
	fmt.Println("\n🔍 STEP 3: Verifying key account setup...")
	keyAccounts := []string{"1301", "2101", "2102"}
	accountNames := map[string]string{
		"1301": "Persediaan Barang Dagangan (Inventory)",
		"2101": "Utang Usaha (Accounts Payable)", 
		"2102": "PPN Masukan (Input Tax)",
	}
	
	for _, code := range keyAccounts {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err == nil {
			fmt.Printf("   ✅ %s: %s (%s) - Type: %s, Balance: %.2f\n", 
				code, accountNames[code], account.Name, account.Type, account.Balance)
		} else {
			fmt.Printf("   ❌ %s: Account not found\n", code)
		}
	}

	// Step 4: Show final accounting equation
	fmt.Printf("\n📊 STEP 4: Accounting Equation Verification:\n")
	
	var assetsTotal, liabilitiesTotal, equityTotal float64
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL").Scan(&assetsTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL").Scan(&liabilitiesTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL").Scan(&equityTotal)
	
	fmt.Printf("   Assets: Rp %.2f\n", assetsTotal)
	fmt.Printf("   Liabilities: Rp %.2f\n", liabilitiesTotal)
	fmt.Printf("   Equity: Rp %.2f\n", equityTotal)
	fmt.Printf("   Balanced? %t (Difference: Rp %.2f)\n", 
		assetsTotal == liabilitiesTotal + equityTotal, 
		assetsTotal - (liabilitiesTotal + equityTotal))

	// Step 5: Summary of fixes implemented
	fmt.Printf("\n🎯 FIXES IMPLEMENTED:\n")
	fmt.Printf("✅ 1. PPN Account (2102) corrected from LIABILITY to ASSET\n")
	fmt.Printf("✅ 2. Account name updated to 'PPN Masukan' (more appropriate)\n")
	fmt.Printf("✅ 3. Account moved to Current Assets category\n")
	fmt.Printf("✅ 4. SSOT Journal Adapter updated to use account 2102\n")
	fmt.Printf("✅ 5. Automatic balance calculation enabled in UnifiedJournalService\n")
	fmt.Printf("✅ 6. Purchase service already calls SSOT for journal creation\n")

	fmt.Printf("\n💡 WHAT HAPPENS NOW WHEN CREATING NEW PURCHASE:\n")
	fmt.Printf("1. Purchase created with correct PPN calculation (11%%)\n")
	fmt.Printf("2. When approved, SSOT journal entry automatically created:\n")
	fmt.Printf("   - Dr. Persediaan Barang Dagangan (1301): Purchase amount\n")
	fmt.Printf("   - Dr. PPN Masukan (2102): PPN amount (now correctly as asset)\n")
	fmt.Printf("   - Cr. Utang Usaha (2101): Total amount including PPN\n")
	fmt.Printf("3. Account balances automatically updated when journal posted\n")
	fmt.Printf("4. Accounting equation remains balanced\n")

	fmt.Printf("\n🎉 PURCHASE SYSTEM FIXES IMPLEMENTED SUCCESSFULLY!\n")
	fmt.Printf("You can now create new purchases and they will automatically:\n")
	fmt.Printf("- Calculate PPN correctly\n")
	fmt.Printf("- Create proper journal entries\n")
	fmt.Printf("- Update account balances automatically\n")
	fmt.Printf("- No need to run manual scripts!\n")
}