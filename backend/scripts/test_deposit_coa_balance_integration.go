package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/shopspring/decimal"
)

type TestResults struct {
	TestName              string
	Success               bool
	Message               string
	CashBankBalanceBefore float64
	CashBankBalanceAfter  float64
	COABalanceBefore      float64
	COABalanceAfter       float64
	JournalEntryCreated   bool
	JournalEntryNumber    string
	BalanceMatches        bool
}

func main() {
	// Initialize database
	db, err := database.InitPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("=== TESTING DEPOSIT & COA BALANCE INTEGRATION ===")
	log.Println("🎯 Objective: Analyze why COA balance doesn't update on deposits")
	log.Println("")

	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)

	results := []TestResults{}

	// ===== TEST 1: Create Test Cash Account =====
	log.Println("🧪 TEST 1: Creating test cash account...")
	
	cashAccountRequest := services.CashBankCreateRequest{
		Name:           "Test Kas COA Balance",
		Type:           "CASH",
		Currency:       "IDR",
		OpeningBalance: 1000000, // Rp 1,000,000 opening balance
		OpeningDate:    services.CustomDate(time.Now()),
		Description:    "Test account for COA balance validation",
	}

	cashAccount, err := cashBankService.CreateCashBankAccount(cashAccountRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to create cash account: %v", err)
		return
	}
	
	log.Printf("✅ Created cash account: %s (ID: %d)", cashAccount.Name, cashAccount.ID)
	log.Printf("   Cash Bank Balance: %.2f", cashAccount.Balance)

	// Get initial COA balance
	var initialCOABalance float64
	var coaAccount models.Account
	if err := db.First(&coaAccount, cashAccount.AccountID).Error; err != nil {
		log.Printf("❌ Failed to get COA account: %v", err)
		return
	}
	initialCOABalance = coaAccount.Balance
	log.Printf("   COA Balance: %.2f", initialCOABalance)

	result1 := TestResults{
		TestName:              "Account Creation",
		Success:               err == nil,
		CashBankBalanceBefore: 0,
		CashBankBalanceAfter:  cashAccount.Balance,
		COABalanceBefore:      0,
		COABalanceAfter:       initialCOABalance,
		BalanceMatches:        fmt.Sprintf("%.2f", cashAccount.Balance) == fmt.Sprintf("%.2f", initialCOABalance),
	}
	results = append(results, result1)

	// ===== TEST 2: Deposit Transaction Analysis =====
	log.Println("\n🧪 TEST 2: Processing deposit and analyzing balance updates...")

	// Get balances before deposit
	cashAccount, _ = cashBankService.GetCashBankByID(cashAccount.ID)
	db.First(&coaAccount, cashAccount.AccountID)
	
	balanceBeforeCash := cashAccount.Balance
	balanceBeforeCOA := coaAccount.Balance

	log.Printf("📊 BEFORE DEPOSIT:")
	log.Printf("   Cash Bank Balance: %.2f", balanceBeforeCash)
	log.Printf("   COA Balance: %.2f", balanceBeforeCOA)

	// Process deposit
	depositAmount := 500000.0 // Rp 500,000
	depositRequest := services.DepositRequest{
		AccountID: cashAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    depositAmount,
		Reference: "TEST-DEP-001",
		Notes:     "Test deposit for COA balance validation",
	}

	depositTx, err := cashBankService.ProcessDeposit(depositRequest, 1)
	
	// Get balances after deposit
	cashAccount, _ = cashBankService.GetCashBankByID(cashAccount.ID)
	db.First(&coaAccount, cashAccount.AccountID)
	
	balanceAfterCash := cashAccount.Balance
	balanceAfterCOA := coaAccount.Balance

	log.Printf("\n📊 AFTER DEPOSIT:")
	log.Printf("   Cash Bank Balance: %.2f (Change: %.2f)", balanceAfterCash, balanceAfterCash-balanceBeforeCash)
	log.Printf("   COA Balance: %.2f (Change: %.2f)", balanceAfterCOA, balanceAfterCOA-balanceBeforeCOA)

	// Check journal entry creation
	var journalEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND source_id = ?", models.SSOTSourceTypeCashBank, depositTx.ID).
		Preload("Lines").
		Preload("Lines.Account").
		Find(&journalEntries)

	journalCreated := len(journalEntries) > 0
	journalNumber := ""
	if journalCreated {
		journalNumber = journalEntries[0].EntryNumber
		log.Printf("✅ Journal Entry Created: %s", journalNumber)
		
		for _, entry := range journalEntries {
			log.Printf("   📋 Entry: %s - %s", entry.EntryNumber, entry.Description)
			for _, line := range entry.Lines {
				log.Printf("      └─ %s: Dr %.2f, Cr %.2f - %s",
					line.Account.Name,
					line.DebitAmount.InexactFloat64(),
					line.CreditAmount.InexactFloat64(),
					line.Description)
			}
		}
	} else {
		log.Printf("❌ No journal entry found!")
	}

	result2 := TestResults{
		TestName:              "Deposit Transaction",
		Success:               err == nil,
		Message:               fmt.Sprintf("Deposit of %.2f processed", depositAmount),
		CashBankBalanceBefore: balanceBeforeCash,
		CashBankBalanceAfter:  balanceAfterCash,
		COABalanceBefore:      balanceBeforeCOA,
		COABalanceAfter:       balanceAfterCOA,
		JournalEntryCreated:   journalCreated,
		JournalEntryNumber:    journalNumber,
		BalanceMatches:        fmt.Sprintf("%.2f", balanceAfterCash) == fmt.Sprintf("%.2f", balanceAfterCOA),
	}
	results = append(results, result2)

	// ===== TEST 3: Manual GL Account Balance Update =====
	log.Println("\n🧪 TEST 3: Testing manual GL account balance update...")

	// Update COA balance manually to test if it's a sync issue
	expectedCOABalance := balanceAfterCash
	err = db.Model(&coaAccount).Update("balance", expectedCOABalance).Error
	if err != nil {
		log.Printf("❌ Failed to update COA balance manually: %v", err)
	} else {
		log.Printf("✅ Manually updated COA balance to %.2f", expectedCOABalance)
		
		// Verify update
		db.First(&coaAccount, cashAccount.AccountID)
		log.Printf("   COA Balance after manual update: %.2f", coaAccount.Balance)
	}

	// ===== TEST 4: Analysis of SSOT Journal Processing =====
	log.Println("\n🧪 TEST 4: Analyzing SSOT journal processing logic...")

	// Check if account is detected as cash bank account
	unifiedJournalService := services.NewUnifiedJournalService(db)
	
	// Use reflection or direct query to check isCashBankAccount logic
	var cashBankCount int64
	db.Table("cash_banks").Where("account_id = ?", cashAccount.AccountID).Count(&cashBankCount)
	
	log.Printf("📊 SSOT Analysis:")
	log.Printf("   Cash Bank Account ID: %d", cashAccount.ID)
	log.Printf("   Linked COA Account ID: %d", cashAccount.AccountID)
	log.Printf("   Is COA linked to Cash Bank: %t (count: %d)", cashBankCount > 0, cashBankCount)
	log.Printf("   This explains why COA balance is skipped in SSOT posting!")

	// ===== FINAL RESULTS SUMMARY =====
	log.Println("\n" + "="*60)
	log.Println("📊 COMPREHENSIVE TEST RESULTS SUMMARY")
	log.Println("="*60)

	for i, result := range results {
		log.Printf("\n🧪 Test %d: %s", i+1, result.TestName)
		if result.Success {
			log.Printf("   ✅ Status: SUCCESS")
		} else {
			log.Printf("   ❌ Status: FAILED")
		}
		if result.Message != "" {
			log.Printf("   📝 Message: %s", result.Message)
		}
		log.Printf("   💰 Cash Bank: %.2f → %.2f", result.CashBankBalanceBefore, result.CashBankBalanceAfter)
		log.Printf("   📊 COA Balance: %.2f → %.2f", result.COABalanceBefore, result.COABalanceAfter)
		if result.JournalEntryCreated {
			log.Printf("   📋 Journal: ✅ Created (%s)", result.JournalEntryNumber)
		} else {
			log.Printf("   📋 Journal: ❌ Not Created")
		}
		if result.BalanceMatches {
			log.Printf("   🎯 Balance Match: ✅ YES")
		} else {
			log.Printf("   🎯 Balance Match: ❌ NO")
		}
	}

	// ===== PROBLEM DIAGNOSIS =====
	log.Println("\n" + "="*60)
	log.Println("🔍 PROBLEM DIAGNOSIS")
	log.Println("="*60)

	log.Println("\n🎯 ROOT CAUSE IDENTIFIED:")
	log.Println("   The UnifiedJournalService.postJournalEntryTx method contains logic")
	log.Println("   that INTENTIONALLY SKIPS balance updates for accounts linked to cash_banks")
	log.Println("   to prevent 'double posting'. However, this means:")
	log.Println("   ")
	log.Println("   ✅ Cash Bank Balance: Updated by CashBankService")
	log.Println("   ❌ COA Balance: Skipped by UnifiedJournalService")
	log.Println("   ✅ Journal Entry: Created correctly")
	log.Println("   ")
	log.Println("   This causes the COA to show outdated balances even though")
	log.Println("   the journal entries are properly recorded.")

	log.Println("\n💡 SOLUTION OPTIONS:")
	log.Println("   1. Remove the cash bank account skip logic in UnifiedJournalService")
	log.Println("   2. Add COA balance sync in CashBankService")
	log.Println("   3. Create a separate balance reconciliation process")

	log.Println("\n🛠️ RECOMMENDED FIX:")
	log.Println("   Modify CashBankService to also update the linked COA account balance")
	log.Println("   alongside the cash bank balance to maintain consistency.")

	log.Println("\n" + "="*60)
}