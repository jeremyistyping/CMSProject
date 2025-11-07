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
	fmt.Println("🔧 PERBAIKAN DUPLICATE ACCOUNTS")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Identify all duplicate accounts
	fmt.Println("\n🔍 STEP 1: Mengidentifikasi semua duplicate accounts...")

	var duplicates []struct {
		Code  string
		Count int
	}

	duplicateQuery := `
		SELECT code, COUNT(*) as count
		FROM accounts 
		WHERE deleted_at IS NULL
		GROUP BY code 
		HAVING COUNT(*) > 1
		ORDER BY code
	`

	if err := db.Raw(duplicateQuery).Scan(&duplicates).Error; err != nil {
		log.Printf("❌ Error finding duplicates: %v", err)
		return
	}

	fmt.Printf("📋 Found %d account codes with duplicates:\n", len(duplicates))
	for _, dup := range duplicates {
		fmt.Printf("   Code: %s - %d duplicates\n", dup.Code, dup.Count)
	}

	// Step 2: Detailed analysis for each duplicate
	fmt.Println("\n📊 STEP 2: Detail analisis untuk setiap duplicate...")

	for _, dup := range duplicates {
		fmt.Printf("\n🔍 Analyzing code: %s\n", dup.Code)
		
		var accounts []models.Account
		db.Where("code = ? AND deleted_at IS NULL", dup.Code).Find(&accounts)
		
		fmt.Println("   ID  | Name                    | Balance        | Active | Created At")
		fmt.Println("   ----|-------------------------|----------------|--------|------------------")
		
		var correctAccount *models.Account
		var maxBalance float64 = -1
		var linkedToCashBank *models.Account
		
		for _, acc := range accounts {
			status := "❌"
			if acc.IsActive {
				status = "✅"
			}
			
			fmt.Printf("   %-3d | %-23s | Rp %11.2f | %s | %s\n", 
				acc.ID, acc.Name, acc.Balance, status, acc.CreatedAt.Format("2006-01-02"))
			
			// Check if linked to cash bank
			var cashBankCount int64
			db.Table("cash_banks").Where("account_id = ?", acc.ID).Count(&cashBankCount)
			if cashBankCount > 0 {
				linkedToCashBank = &acc
				fmt.Printf("       └─ 💰 LINKED TO CASH BANK\n")
			}
			
			// Track account with highest balance
			if acc.Balance > maxBalance {
				maxBalance = acc.Balance
				correctAccount = &acc
			}
		}
		
		// Determine the correct account
		if linkedToCashBank != nil {
			correctAccount = linkedToCashBank
			fmt.Printf("   ✅ CORRECT ACCOUNT: ID %d (linked to cash bank)\n", correctAccount.ID)
		} else if correctAccount != nil {
			fmt.Printf("   ✅ CORRECT ACCOUNT: ID %d (highest balance: %.2f)\n", correctAccount.ID, correctAccount.Balance)
		}
		
		// Show which accounts should be deactivated/deleted
		fmt.Printf("   ❌ ACCOUNTS TO DEACTIVATE:\n")
		for _, acc := range accounts {
			if correctAccount != nil && acc.ID != correctAccount.ID {
				fmt.Printf("       - ID %d (Balance: %.2f)\n", acc.ID, acc.Balance)
			}
		}
	}

	// Step 3: Create backup before fixing
	fmt.Println("\n💾 STEP 3: Creating backup...")
	
	backupQuery := `
		DROP TABLE IF EXISTS accounts_backup_duplicate_fix;
		CREATE TABLE accounts_backup_duplicate_fix AS 
		SELECT *, NOW() as backup_timestamp 
		FROM accounts 
		WHERE deleted_at IS NULL
	`

	if err := db.Exec(backupQuery).Error; err != nil {
		log.Printf("❌ Failed to create backup: %v", err)
		return
	}
	fmt.Println("✅ Backup created: accounts_backup_duplicate_fix")

	// Step 4: Fix duplicates (deactivate wrong ones)
	fmt.Println("\n🔧 STEP 4: Memperbaiki duplicate accounts...")
	
	fixedCount := 0
	for _, dup := range duplicates {
		var accounts []models.Account
		db.Where("code = ? AND deleted_at IS NULL", dup.Code).Find(&accounts)
		
		var correctAccount *models.Account
		var maxBalance float64 = -1
		
		// Find account linked to cash bank first
		for _, acc := range accounts {
			var cashBankCount int64
			db.Table("cash_banks").Where("account_id = ?", acc.ID).Count(&cashBankCount)
			if cashBankCount > 0 {
				correctAccount = &acc
				break
			}
		}
		
		// If no cash bank link, use highest balance
		if correctAccount == nil {
			for _, acc := range accounts {
				if acc.Balance > maxBalance {
					maxBalance = acc.Balance
					correctAccount = &acc
				}
			}
		}
		
		// Deactivate wrong accounts
		if correctAccount != nil {
			for _, acc := range accounts {
				if acc.ID != correctAccount.ID {
					fmt.Printf("🔧 Deactivating duplicate account: %s (ID: %d, Balance: %.2f)\n", 
						acc.Name, acc.ID, acc.Balance)
					
					// Set is_active = false instead of deleting
					result := db.Model(&acc).Update("is_active", false)
					if result.Error != nil {
						fmt.Printf("❌ Failed to deactivate account ID %d: %v\n", acc.ID, result.Error)
					} else {
						fmt.Printf("✅ Successfully deactivated account ID %d\n", acc.ID)
						fixedCount++
					}
				}
			}
		}
	}

	// Step 5: Verification
	fmt.Println("\n📊 STEP 5: Verifikasi hasil perbaikan...")
	
	var remainingDuplicates []struct {
		Code  string
		Count int
	}

	verifyQuery := `
		SELECT code, COUNT(*) as count
		FROM accounts 
		WHERE deleted_at IS NULL AND is_active = true
		GROUP BY code 
		HAVING COUNT(*) > 1
		ORDER BY code
	`

	db.Raw(verifyQuery).Scan(&remainingDuplicates)
	
	// Final summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 HASIL PERBAIKAN DUPLICATE ACCOUNTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📋 Duplicate codes yang diperbaiki: %d\n", len(duplicates))
	fmt.Printf("📋 Account yang di-deaktivasi: %d\n", fixedCount)
	fmt.Printf("📋 Duplicate yang tersisa: %d\n", len(remainingDuplicates))

	if len(remainingDuplicates) == 0 {
		fmt.Println("🎉 SEMUA DUPLICATE BERHASIL DIPERBAIKI!")
		
		fmt.Println("\n✅ LANGKAH SELANJUTNYA:")
		fmt.Println("   1. REFRESH browser dengan Ctrl+F5")
		fmt.Println("   2. Clear cache browser") 
		fmt.Println("   3. COA seharusnya menampilkan balance yang benar sekarang")
		
	} else {
		fmt.Println("⚠️  Masih ada duplicate yang tersisa:")
		for _, dup := range remainingDuplicates {
			fmt.Printf("   - Code: %s (%d duplicates)\n", dup.Code, dup.Count)
		}
	}

	fmt.Printf("💾 Backup tersedia di: accounts_backup_duplicate_fix\n")
	fmt.Println(strings.Repeat("=", 80))
}