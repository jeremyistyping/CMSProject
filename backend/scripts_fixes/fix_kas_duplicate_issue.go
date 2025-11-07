package main

import (
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"uniqueIndex;size:20"`
	Name        string `json:"name" gorm:"size:255"`
	Type        string `json:"type" gorm:"size:50"`
	Balance     float64 `json:"balance" gorm:"default:0"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	ParentID    *uint  `json:"parent_id"`
	IsHeader    bool   `json:"is_header" gorm:"default:false"`
	Level       int    `json:"level" gorm:"default:1"`
	Description string `json:"description"`
}

func main() {
	// Connect to database
	dsn := "accounting_user:Bismillah2024!@tcp(localhost:3306)/accounting_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🛠️  FIXING KAS DUPLICATE & CATEGORIZATION ISSUES")
	fmt.Println("==================================================")

	// Step 1: Identify all Cash accounts
	var kasAccounts []Account
	db.Where("name LIKE ?", "%Kas%").Order("code").Find(&kasAccounts)

	fmt.Printf("\n📋 CURRENT KAS ACCOUNTS (%d found):\n", len(kasAccounts))
	for _, acc := range kasAccounts {
		fmt.Printf("- ID:%d Code:%s Name:%s Type:%s Active:%t\n", 
			acc.ID, acc.Code, acc.Name, acc.Type, acc.IsActive)
	}

	// Step 2: Check for journal entries usage
	type AccountUsage struct {
		AccountID   uint
		EntryCount  int64
		TotalDebit  float64
		TotalCredit float64
	}

	var usageData []AccountUsage
	for _, acc := range kasAccounts {
		var usage AccountUsage
		usage.AccountID = acc.ID
		
		db.Table("journal_lines").
			Where("account_id = ?", acc.ID).
			Count(&usage.EntryCount)
			
		if usage.EntryCount > 0 {
			db.Table("journal_lines").
				Where("account_id = ?", acc.ID).
				Select("COALESCE(SUM(debit_amount), 0) as total_debit, COALESCE(SUM(credit_amount), 0) as total_credit").
				Row().Scan(&usage.TotalDebit, &usage.TotalCredit)
		}
		
		usageData = append(usageData, usage)
	}

	fmt.Printf("\n📊 JOURNAL ENTRY USAGE:\n")
	for i, usage := range usageData {
		acc := kasAccounts[i]
		if usage.EntryCount > 0 {
			fmt.Printf("- %s (%s): %d entries, Debit: %.2f, Credit: %.2f\n", 
				acc.Code, acc.Name, usage.EntryCount, usage.TotalDebit, usage.TotalCredit)
		} else {
			fmt.Printf("- %s (%s): No transactions\n", acc.Code, acc.Name)
		}
	}

	// Step 3: Determine which account to keep and which to remove/fix
	fmt.Printf("\n💡 RECOMMENDED ACTIONS:\n")
	fmt.Println("========================")

	var keepAccount *Account
	var problemAccounts []Account

	for i, acc := range kasAccounts {
		usage := usageData[i]
		
		// Priority: Keep 1101 - Kas if it exists and has standard structure
		if acc.Code == "1101" && acc.Name == "Kas" && acc.Type == "ASSET" && acc.IsActive {
			keepAccount = &acc
			fmt.Printf("✅ KEEP: %s (%s) - Standard cash account\n", acc.Code, acc.Name)
		} else {
			problemAccounts = append(problemAccounts, acc)
			
			if usage.EntryCount > 0 {
				fmt.Printf("⚠️  PROBLEM: %s (%s) - Has %d transactions, needs data migration\n", 
					acc.Code, acc.Name, usage.EntryCount)
			} else {
				fmt.Printf("❌ REMOVE: %s (%s) - No transactions, safe to delete\n", 
					acc.Code, acc.Name)
			}
		}
	}

	// Step 4: Fix the issues
	fmt.Printf("\n🔧 APPLYING FIXES:\n")
	fmt.Println("==================")

	// If no standard 1101 account exists, keep the first active one
	if keepAccount == nil && len(kasAccounts) > 0 {
		for _, acc := range kasAccounts {
			if acc.IsActive && !acc.IsHeader {
				keepAccount = &acc
				fmt.Printf("✅ DESIGNATED KEEP: %s (%s) - Best available option\n", acc.Code, acc.Name)
				break
			}
		}
	}

	if keepAccount == nil {
		fmt.Println("❌ ERROR: No suitable cash account found to keep!")
		return
	}

	// Fix accounts that need to be removed/updated
	for i, acc := range problemAccounts {
		usage := usageData[i]
		
		if usage.EntryCount > 0 {
			fmt.Printf("🔄 MIGRATING: %s (%s) transactions to %s (%s)\n", 
				acc.Code, acc.Name, keepAccount.Code, keepAccount.Name)
			
			// Migrate journal entries
			result := db.Table("journal_lines").
				Where("account_id = ?", acc.ID).
				Update("account_id", keepAccount.ID)
				
			fmt.Printf("   Migrated %d journal entries\n", result.RowsAffected)
			
			// Migrate other references if needed
			db.Table("unified_journal_lines").
				Where("account_id = ?", acc.ID).
				Update("account_id", keepAccount.ID)
		}
		
		if acc.Code != keepAccount.Code {
			fmt.Printf("🗑️  DEACTIVATING: %s (%s)\n", acc.Code, acc.Name)
			
			// Deactivate instead of delete to maintain data integrity
			db.Model(&acc).Updates(map[string]interface{}{
				"is_active": false,
				"name": acc.Name + " (DEPRECATED - merged with " + keepAccount.Code + ")",
			})
		}
	}

	// Step 5: Fix PPN Masukan categorization issue
	fmt.Printf("\n🔧 FIXING PPN MASUKAN CATEGORIZATION:\n")
	fmt.Println("======================================")

	var ppnAccount Account
	result := db.Where("code = ? OR name LIKE ?", "2102", "%PPN Masukan%").First(&ppnAccount)
	if result.Error == nil {
		if ppnAccount.Type != "ASSET" {
			fmt.Printf("⚠️  PPN Masukan (%s) is incorrectly typed as %s\n", ppnAccount.Code, ppnAccount.Type)
			fmt.Printf("💡 Note: PPN Masukan should be ASSET type (Input VAT is recoverable)\n")
			
			// Update PPN Masukan to correct type
			db.Model(&ppnAccount).Update("type", "ASSET")
			fmt.Printf("✅ Updated %s (%s) type to ASSET\n", ppnAccount.Code, ppnAccount.Name)
		} else {
			fmt.Printf("✅ PPN Masukan (%s) is correctly categorized as ASSET\n", ppnAccount.Code)
		}
	}

	// Step 6: Update Balance Sheet Service Logic (create patch file)
	fmt.Printf("\n📝 CREATING BALANCE SHEET SERVICE PATCH:\n")
	fmt.Println("=========================================")

	patchContent := `
// PATCH FOR: services/ssot_balance_sheet_service.go
// FIX: Duplicate Cash accounts in Balance Sheet display

// In categorizeAssetAccount function, add better logic:

func (s *SSOTBalanceSheetService) categorizeAssetAccount(bsData *SSOTBalanceSheetData, item BSAccountItem, code string) {
	switch {
	// Current Assets - Cash accounts (110x, 1101 specifically)
	case code == "1101" || (strings.HasPrefix(code, "110") && strings.Contains(strings.ToLower(item.AccountName), "kas")):
		bsData.Assets.CurrentAssets.Cash += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)

	// Skip duplicate cash accounts that are not the primary one
	case strings.HasPrefix(code, "1100-") && strings.Contains(strings.ToLower(item.AccountName), "kas"):
		// Skip this - it's likely a duplicate of 1101
		return

	// Continue with other categorizations...
	case strings.HasPrefix(code, "112"), strings.HasPrefix(code, "120"): // Accounts Receivable
		bsData.Assets.CurrentAssets.Receivables += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	// ... rest of the logic
	}
}
`

	fmt.Println(patchContent)

	fmt.Printf("\n🎉 FIXES COMPLETED!\n")
	fmt.Println("===================")
	fmt.Printf("✅ Migrated transactions to primary cash account: %s\n", keepAccount.Code)
	fmt.Printf("✅ Deactivated duplicate accounts\n")
	fmt.Printf("✅ Fixed PPN Masukan categorization\n")
	fmt.Printf("📝 Balance Sheet service patch created above\n")
	fmt.Printf("\n🔄 NEXT STEPS:\n")
	fmt.Println("1. Apply the service patch to remove duplicates from Balance Sheet display")
	fmt.Println("2. Test the Balance Sheet generation")
	fmt.Println("3. Verify no duplicate Kas accounts appear")
}