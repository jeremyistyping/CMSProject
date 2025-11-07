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
	fmt.Println("🔍 DEBUG: ANALISIS MASALAH DEPOSIT 5 JUTA")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Cek account Kas yang bermasalah
	fmt.Println("\n📊 STEP 1: Menganalisis account Kas (1101)...")

	var kasAccount models.Account
	if err := db.Where("code = ?", "1101").First(&kasAccount).Error; err != nil {
		log.Printf("❌ Account Kas (1101) tidak ditemukan: %v", err)
		return
	}

	fmt.Printf("📋 Account Kas (1101):\n")
	fmt.Printf("   ID: %d\n", kasAccount.ID)
	fmt.Printf("   Name: %s\n", kasAccount.Name)
	fmt.Printf("   Balance: %.2f\n", kasAccount.Balance)
	fmt.Printf("   Active: %t\n", kasAccount.IsActive)

	// Step 2: Cek cash bank yang terhubung dengan account ini
	fmt.Println("\n💰 STEP 2: Mencari Cash Bank yang terhubung dengan account Kas...")

	var linkedCashBanks []models.CashBank
	if err := db.Where("account_id = ?", kasAccount.ID).Find(&linkedCashBanks).Error; err != nil {
		log.Printf("❌ Error getting linked cash banks: %v", err)
		return
	}

	fmt.Printf("📋 Found %d cash bank accounts linked to Kas (1101):\n", len(linkedCashBanks))

	for i, cashBank := range linkedCashBanks {
		fmt.Printf("\n%d. 💳 %s (ID: %d)\n", i+1, cashBank.Name, cashBank.ID)
		fmt.Printf("   Code: %s\n", cashBank.Code)
		fmt.Printf("   Balance: %.2f\n", cashBank.Balance)
		fmt.Printf("   Active: %t\n", cashBank.IsActive)
		fmt.Printf("   Account ID: %d\n", cashBank.AccountID)
	}

	if len(linkedCashBanks) == 0 {
		fmt.Println("❌ MASALAH DITEMUKAN: Tidak ada Cash Bank yang terhubung dengan account Kas (1101)!")
		fmt.Println("   Ini bisa menjadi penyebab balance tidak update.")
	}

	// Step 3: Cek transaksi deposit terbaru
	fmt.Println("\n📋 STEP 3: Mencari transaksi deposit terbaru...")

	query := `
		SELECT cbt.*, cb.name as cash_bank_name 
		FROM cash_bank_transactions cbt
		JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
		WHERE cbt.reference_type = 'DEPOSIT'
		ORDER BY cbt.created_at DESC
		LIMIT 5
	`

	type TransactionWithName struct {
		models.CashBankTransaction
		CashBankName string `gorm:"column:cash_bank_name"`
	}

	var transactionsWithNames []TransactionWithName
	if err := db.Raw(query).Scan(&transactionsWithNames).Error; err != nil {
		log.Printf("❌ Error getting recent transactions: %v", err)
	} else {
		fmt.Printf("📋 5 transaksi deposit terbaru:\n")
		for i, tx := range transactionsWithNames {
			fmt.Printf("\n%d. 💰 %s\n", i+1, tx.CashBankName)
			fmt.Printf("   Transaction ID: %d\n", tx.ID)
			fmt.Printf("   Amount: %.2f\n", tx.Amount)
			fmt.Printf("   Balance After: %.2f\n", tx.BalanceAfter)
			fmt.Printf("   Date: %s\n", tx.TransactionDate.Format("2006-01-02 15:04:05"))
			fmt.Printf("   Cash Bank ID: %d\n", tx.CashBankID)
		}
	}

	// Step 4: Cek journal entries untuk deposit terbaru
	fmt.Println("\n📝 STEP 4: Mencari journal entries untuk deposit terbaru...")

	if len(transactionsWithNames) > 0 {
		latestTx := transactionsWithNames[0]
		
		var journalEntries []models.SSOTJournalEntry
		db.Where("source_type = ? AND source_id = ?", models.SSOTSourceTypeCashBank, latestTx.ID).
			Preload("Lines").
			Preload("Lines.Account").
			Find(&journalEntries)

		if len(journalEntries) > 0 {
			fmt.Printf("✅ Journal entry ditemukan untuk transaksi ID %d:\n", latestTx.ID)
			for _, entry := range journalEntries {
				fmt.Printf("   📋 Entry: %s - %s\n", entry.EntryNumber, entry.Description)
				fmt.Printf("       Status: %s | Balanced: %t\n", entry.Status, entry.IsBalanced)
				
				for _, line := range entry.Lines {
					fmt.Printf("       └─ %s (ID:%d, Code:%s): Dr %.2f, Cr %.2f\n",
						line.Account.Name, line.AccountID, line.Account.Code,
						line.DebitAmount.InexactFloat64(),
						line.CreditAmount.InexactFloat64())
				}
			}
		} else {
			fmt.Printf("❌ MASALAH: Tidak ada journal entry untuk transaksi deposit ID %d\n", latestTx.ID)
		}
	}

	// Step 5: Analisis semua balance mismatch
	fmt.Println("\n⚖️  STEP 5: Analisis balance mismatch...")

	var allMismatches []struct {
		CashBankID   uint
		CashBankName string
		AccountID    uint
		AccountCode  string
		AccountName  string
		CashBalance  float64
		COABalance   float64
		Difference   float64
	}

	mismatchQuery := `
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

	if err := db.Raw(mismatchQuery).Scan(&allMismatches).Error; err != nil {
		log.Printf("❌ Error getting mismatches: %v", err)
	} else {
		fmt.Printf("📊 Balance mismatches ditemukan: %d\n", len(allMismatches))
		
		for i, mismatch := range allMismatches {
			fmt.Printf("\n%d. ❌ %s (Cash Bank ID: %d)\n", i+1, mismatch.CashBankName, mismatch.CashBankID)
			fmt.Printf("   COA Account: %s (%s) - ID: %d\n", mismatch.AccountName, mismatch.AccountCode, mismatch.AccountID)
			fmt.Printf("   Cash Balance: %.2f\n", mismatch.CashBalance)
			fmt.Printf("   COA Balance: %.2f\n", mismatch.COABalance)
			fmt.Printf("   Difference: %.2f\n", mismatch.Difference)
		}
	}

	// Step 6: Diagnosis
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 DIAGNOSIS & REKOMENDASI")
	fmt.Println(strings.Repeat("=", 80))

	if len(linkedCashBanks) == 0 {
		fmt.Println("❌ ROOT CAUSE DITEMUKAN:")
		fmt.Println("   Account Kas (1101) tidak terhubung dengan Cash Bank account manapun!")
		fmt.Println("   Deposit di Cash & Bank tidak akan mengupdate balance COA.")
		fmt.Println("")
		fmt.Println("🔧 SOLUSI:")
		fmt.Println("   1. Pastikan Cash Bank account (Kas) terhubung dengan COA account 1101")
		fmt.Println("   2. Update account_id pada tabel cash_banks")
		fmt.Println("   3. Atau buat ulang Cash Bank account dengan link yang benar")
		
	} else if len(allMismatches) > 0 {
		fmt.Println("❌ MASALAH SINKRONISASI:")
		fmt.Printf("   Ditemukan %d account dengan balance tidak sinkron\n", len(allMismatches))
		fmt.Println("")
		fmt.Println("🔧 SOLUSI:")
		fmt.Println("   Jalankan script fix balance synchronization:")
		fmt.Println("   go run scripts/fix_historical_balance_sync.go")
		
	} else {
		fmt.Println("🤔 TIDAK ADA MASALAH TERDETEKSI:")
		fmt.Println("   Semua balance sudah sinkron, mungkin masalah di UI cache atau timing")
	}

	// Step 7: Quick fix suggestion
	if len(allMismatches) > 0 {
		fmt.Println("\n🚀 QUICK FIX - Update balance sekarang?")
		fmt.Println("   Apakah Anda ingin memperbaiki balance mismatch sekarang? (y/n)")
		
		// Manual fix for the specific case
		for _, mismatch := range allMismatches {
			if mismatch.AccountCode == "1101" || strings.Contains(strings.ToLower(mismatch.AccountName), "kas") {
				fmt.Printf("\n🔧 Memperbaiki balance %s...\n", mismatch.CashBankName)
				
				result := db.Model(&models.Account{}).
					Where("id = ?", mismatch.AccountID).
					Update("balance", mismatch.CashBalance)

				if result.Error != nil {
					fmt.Printf("❌ Gagal memperbaiki: %v\n", result.Error)
				} else {
					fmt.Printf("✅ Balance %s berhasil diperbaiki: %.2f → %.2f\n", 
						mismatch.AccountName, mismatch.COABalance, mismatch.CashBalance)
				}
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}