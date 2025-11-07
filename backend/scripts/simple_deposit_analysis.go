package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🔍 ANALISIS MASALAH DEPOSIT & COA BALANCE")
	fmt.Println(strings.Repeat("=", 80))

	// Initialize services
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)

	// Step 1: Analisis kondisi saat ini
	fmt.Println("\n📊 STEP 1: Menganalisis kondisi Cash Bank vs COA Balance saat ini...")
	
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
		  AND cb.deleted_at IS NULL
		  AND acc.deleted_at IS NULL
		ORDER BY ABS(cb.balance - acc.balance) DESC
		LIMIT 5
	`

	if err := db.Raw(query).Scan(&mismatches).Error; err != nil {
		log.Printf("❌ Failed to analyze current state: %v", err)
		return
	}

	fmt.Printf("📋 Found %d cash bank accounts (top 5 by difference):\n", len(mismatches))
	
	totalMismatch := 0
	for _, mismatch := range mismatches {
		if mismatch.CashBalance != mismatch.COABalance {
			totalMismatch++
		}
		
		status := "✅ SYNCED"
		if mismatch.CashBalance != mismatch.COABalance {
			status = "❌ MISMATCH"
		}

		fmt.Printf("   %s %s (ID:%d) → %s (%s)\n", 
			status, mismatch.CashBankName, mismatch.CashBankID,
			mismatch.AccountCode, mismatch.AccountName)
		fmt.Printf("      Cash: %.2f | COA: %.2f | Diff: %.2f\n", 
			mismatch.CashBalance, mismatch.COABalance, mismatch.Difference)
	}

	if totalMismatch == 0 {
		fmt.Println("✅ All checked accounts have synchronized balances")
	} else {
		fmt.Printf("⚠️  %d accounts have balance mismatches\n", totalMismatch)
	}

	// Step 2: Create test account untuk testing
	fmt.Println("\n🧪 STEP 2: Membuat test account untuk analisis...")
	
	createRequest := services.CashBankCreateRequest{
		Name:           fmt.Sprintf("Test Deposit Analysis %d", time.Now().Unix()),
		Type:           "CASH",
		Currency:       "IDR",
		OpeningBalance: 1000000, // Rp 1,000,000
		OpeningDate:    services.CustomDate(time.Now()),
		Description:    "Test account for deposit balance analysis",
	}

	testAccount, err := cashBankService.CreateCashBankAccount(createRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to create test account: %v", err)
		return
	}

	fmt.Printf("✅ Created test account: %s (ID: %d)\n", testAccount.Name, testAccount.ID)
	
	// Get initial COA balance
	var coaAccount models.Account
	db.First(&coaAccount, testAccount.AccountID)
	
	fmt.Printf("📊 Initial State:\n")
	fmt.Printf("   Cash Bank Balance: %.2f\n", testAccount.Balance)
	fmt.Printf("   COA Balance: %.2f\n", coaAccount.Balance)
	fmt.Printf("   Balance Match: %t\n", testAccount.Balance == coaAccount.Balance)

	// Step 3: Test deposit process
	fmt.Println("\n🔄 STEP 3: Testing deposit process...")

	// Save before state
	beforeCashBalance := testAccount.Balance
	beforeCOABalance := coaAccount.Balance

	// Process deposit
	depositAmount := 500000.0 // Rp 500,000
	depositRequest := services.DepositRequest{
		AccountID: testAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    depositAmount,
		Reference: "TEST-DEP-001",
		Notes:     "Test deposit for analysis",
	}

	fmt.Printf("🔄 Processing deposit of Rp %.2f...\n", depositAmount)

	depositTx, err := cashBankService.ProcessDeposit(depositRequest, 1)
	if err != nil {
		log.Printf("❌ Deposit failed: %v", err)
		return
	}

	fmt.Printf("✅ Deposit transaction created: ID %d\n", depositTx.ID)

	// Get after state
	db.First(&testAccount, testAccount.ID)
	db.First(&coaAccount, testAccount.AccountID)

	fmt.Printf("\n📊 AFTER DEPOSIT:\n")
	fmt.Printf("   Cash Bank Balance: %.2f (Change: %.2f)\n", testAccount.Balance, testAccount.Balance-beforeCashBalance)
	fmt.Printf("   COA Balance: %.2f (Change: %.2f)\n", coaAccount.Balance, coaAccount.Balance-beforeCOABalance)

	// Step 4: Check journal entries
	fmt.Println("\n📋 STEP 4: Checking journal entries...")
	
	var journalEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND source_id = ?", models.SSOTSourceTypeCashBank, depositTx.ID).
		Preload("Lines").
		Preload("Lines.Account").
		Find(&journalEntries)

	if len(journalEntries) > 0 {
		fmt.Printf("✅ Journal Entry Created: %s\n", journalEntries[0].EntryNumber)
		for _, entry := range journalEntries {
			fmt.Printf("   📋 Entry: %s - %s\n", entry.EntryNumber, entry.Description)
			fmt.Printf("       Status: %s | Balanced: %t\n", entry.Status, entry.IsBalanced)
			
			for _, line := range entry.Lines {
				fmt.Printf("       └─ %s (%s): Dr %.2f, Cr %.2f\n",
					line.Account.Name, line.Account.Code,
					line.DebitAmount.InexactFloat64(),
					line.CreditAmount.InexactFloat64())
			}
		}
	} else {
		fmt.Printf("❌ No journal entry found!\n")
	}

	// Step 5: Root cause analysis
	fmt.Println("\n🎯 STEP 5: ROOT CAUSE ANALYSIS")
	
	// Check if account is identified as cash bank account
	var cashBankCount int64
	db.Table("cash_banks").Where("account_id = ?", testAccount.AccountID).Count(&cashBankCount)

	fmt.Printf("🔍 Analysis Results:\n")
	fmt.Printf("   Cash Bank Account ID: %d\n", testAccount.ID)
	fmt.Printf("   Linked COA Account ID: %d\n", testAccount.AccountID)
	fmt.Printf("   COA Account linked to Cash Bank: %t (count: %d)\n", cashBankCount > 0, cashBankCount)

	// Final diagnosis
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 DIAGNOSIS HASIL TESTING")
	fmt.Println(strings.Repeat("=", 80))

	if testAccount.Balance != coaAccount.Balance {
		fmt.Println("❌ MASALAH TERDETEKSI: Balance tidak sinkron!")
		fmt.Printf("   Cash Bank Balance: %.2f\n", testAccount.Balance)
		fmt.Printf("   COA Balance: %.2f\n", coaAccount.Balance)
		fmt.Printf("   Selisih: %.2f\n", testAccount.Balance-coaAccount.Balance)
		
		fmt.Println("\n🔍 ROOT CAUSE:")
		fmt.Println("   UnifiedJournalService melewati update balance untuk account")
		fmt.Println("   yang terhubung dengan cash_banks (skip logic di line 250-253)")
		fmt.Println("   ")
		fmt.Println("   ✅ Journal Entry: Dibuat dengan benar")
		fmt.Println("   ✅ Cash Bank Balance: Diupdate oleh CashBankService")  
		fmt.Println("   ❌ COA Balance: Dilewati oleh UnifiedJournalService")
		
	} else {
		fmt.Println("✅ BALANCE SUDAH SINKRON")
		fmt.Println("   Tidak ada masalah yang terdeteksi")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}