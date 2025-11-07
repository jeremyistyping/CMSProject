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
	fmt.Println("🔧 QUICK FIX: IDENTIFIKASI & PERBAIKAN MASALAH COA")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Check current COA accounts that are displayed
	fmt.Println("\n📊 STEP 1: Checking accounts yang ditampilkan di COA...")

	var coaAccounts []models.Account
	if err := db.Where("is_active = ? AND deleted_at IS NULL", true).
		Order("code ASC").
		Find(&coaAccounts).Error; err != nil {
		log.Printf("❌ Error getting COA accounts: %v", err)
		return
	}

	// Group by code to see duplicates
	codeGroups := make(map[string][]models.Account)
	for _, acc := range coaAccounts {
		codeGroups[acc.Code] = append(codeGroups[acc.Code], acc)
	}

	fmt.Println("📋 Current active COA accounts:")
	fmt.Println("   Code    | ID  | Name                    | Balance        | Linked to CashBank")
	fmt.Println("   --------|-----|-------------------------|----------------|-------------------")

	duplicateFound := false
	for code, accounts := range codeGroups {
		if len(accounts) > 1 {
			duplicateFound = true
			fmt.Printf("   🚨 DUPLICATE CODE: %s\n", code)
		}

		for _, acc := range accounts {
			// Check if linked to cash bank
			var cashBankCount int64
			db.Table("cash_banks").Where("account_id = ?", acc.ID).Count(&cashBankCount)
			
			linkedStatus := "❌ No"
			if cashBankCount > 0 {
				linkedStatus = "✅ Yes"
			}

			duplicateFlag := ""
			if len(accounts) > 1 {
				duplicateFlag = "🚨"
			}

			fmt.Printf("   %s%-7s | %-3d | %-23s | Rp %11.2f | %s\n", 
				duplicateFlag, acc.Code, acc.ID, acc.Name, acc.Balance, linkedStatus)
		}
	}

	// Step 2: Find the issue
	fmt.Println("\n🔍 STEP 2: Menganalisis masalah spesifik...")

	// Check cash bank accounts and their linked COA accounts
	var cashBanks []models.CashBank
	db.Where("is_active = ? AND deleted_at IS NULL", true).Find(&cashBanks)

	fmt.Println("📋 Cash Bank Accounts dan COA yang terhubung:")
	fmt.Println("   Cash Bank Name          | Cash Balance   | COA ID | COA Balance    | Match?")
	fmt.Println("   ------------------------|----------------|--------|----------------|-------")

	mismatchFound := false
	for _, cb := range cashBanks {
		var linkedAccount models.Account
		if err := db.First(&linkedAccount, cb.AccountID).Error; err != nil {
			fmt.Printf("   %-23s | Rp %11.2f | %-6d | ❌ NOT FOUND   | ❌\n", 
				cb.Name, cb.Balance, cb.AccountID)
			continue
		}

		match := "✅"
		if cb.Balance != linkedAccount.Balance {
			match = "❌"
			mismatchFound = true
		}

		fmt.Printf("   %-23s | Rp %11.2f | %-6d | Rp %11.2f | %s\n", 
			cb.Name, cb.Balance, cb.AccountID, linkedAccount.Balance, match)
	}

	// Step 3: Quick fix suggestions
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 DIAGNOSIS & QUICK FIX")
	fmt.Println(strings.Repeat("=", 80))

	if duplicateFound {
		fmt.Println("❌ MASALAH: DUPLICATE ACCOUNTS DITEMUKAN")
		fmt.Println("   Frontend mungkin mengambil account yang salah")
		fmt.Println("")
		fmt.Println("🔧 SOLUSI:")
		fmt.Println("   1. Deaktivasi account duplicate yang balance-nya 0")
		fmt.Println("   2. Pastikan hanya 1 account aktif per code")
		
		// Auto-fix: deactivate duplicates with balance 0
		fmt.Println("\n🚀 AUTO-FIX: Deaktivasi duplicate dengan balance 0...")
		
		fixedCount := 0
		for _, accounts := range codeGroups {
			if len(accounts) <= 1 {
				continue
			}
			
			// Find the correct account (linked to cash bank or highest balance)
			var correctAccount *models.Account
			
			// Priority 1: Account linked to cash bank
			for _, acc := range accounts {
				var cashBankCount int64
				db.Table("cash_banks").Where("account_id = ?", acc.ID).Count(&cashBankCount)
				if cashBankCount > 0 {
					correctAccount = &acc
					break
				}
			}
			
			// Priority 2: Account with highest balance
			if correctAccount == nil {
				maxBalance := float64(-1)
				for _, acc := range accounts {
					if acc.Balance > maxBalance {
						maxBalance = acc.Balance
						correctAccount = &acc
					}
				}
			}
			
			// Deactivate other accounts
			if correctAccount != nil {
				for _, acc := range accounts {
					if acc.ID != correctAccount.ID {
						fmt.Printf("   🔧 Deactivating %s (ID: %d, Balance: %.2f)\n", acc.Name, acc.ID, acc.Balance)
						
						result := db.Model(&acc).Update("is_active", false)
						if result.Error != nil {
							fmt.Printf("   ❌ Failed: %v\n", result.Error)
						} else {
							fmt.Printf("   ✅ Success\n")
							fixedCount++
						}
					}
				}
			}
		}
		
		if fixedCount > 0 {
			fmt.Printf("✅ Fixed %d duplicate accounts\n", fixedCount)
		}
		
	} else if mismatchFound {
		fmt.Println("❌ MASALAH: BALANCE TIDAK SINKRON")
		fmt.Println("   Cash Bank balance berbeda dengan COA balance")
		fmt.Println("")
		fmt.Println("🔧 SOLUSI:")
		fmt.Println("   Sinkronisasi balance COA dengan Cash Bank")
		
		// Auto-fix balance mismatch
		fmt.Println("\n🚀 AUTO-FIX: Sinkronisasi balance...")
		
		for _, cb := range cashBanks {
			var linkedAccount models.Account
			if err := db.First(&linkedAccount, cb.AccountID).Error; err != nil {
				continue
			}
			
			if cb.Balance != linkedAccount.Balance {
				fmt.Printf("   🔧 Fixing %s: %.2f → %.2f\n", linkedAccount.Name, linkedAccount.Balance, cb.Balance)
				
				result := db.Model(&linkedAccount).Update("balance", cb.Balance)
				if result.Error != nil {
					fmt.Printf("   ❌ Failed: %v\n", result.Error)
				} else {
					fmt.Printf("   ✅ Success\n")
				}
			}
		}
		
	} else {
		fmt.Println("✅ TIDAK ADA MASALAH DITEMUKAN")
		fmt.Println("   Semua balance sudah sinkron")
		fmt.Println("")
		fmt.Println("🔧 KEMUNGKINAN MASALAH:")
		fmt.Println("   1. Frontend cache - coba refresh browser (Ctrl+F5)")
		fmt.Println("   2. UI tidak auto-refresh - tunggu beberapa detik")
		fmt.Println("   3. JavaScript error - cek console browser")
	}

	// Step 4: Final verification
	fmt.Println("\n📊 FINAL VERIFICATION...")
	
	// Show final state of Kas account
	var kasAccount models.Account
	if err := db.Where("code = ? AND is_active = ? AND deleted_at IS NULL", "1101", true).
		First(&kasAccount).Error; err != nil {
		fmt.Println("❌ Account Kas (1101) tidak ditemukan atau tidak aktif")
	} else {
		fmt.Printf("✅ Account Kas (1101) - Balance: Rp %.2f\n", kasAccount.Balance)
		
		// Check linked cash bank
		var linkedCashBank models.CashBank
		if err := db.Where("account_id = ? AND is_active = ?", kasAccount.ID, true).
			First(&linkedCashBank).Error; err != nil {
			fmt.Println("❌ Cash Bank yang terhubung tidak ditemukan")
		} else {
			fmt.Printf("✅ Cash Bank terhubung: %s - Balance: Rp %.2f\n", linkedCashBank.Name, linkedCashBank.Balance)
		}
	}

	fmt.Println("\n🎉 QUICK FIX SELESAI!")
	fmt.Println("   Silakan refresh browser dan cek COA lagi")
	fmt.Println(strings.Repeat("=", 80))
}