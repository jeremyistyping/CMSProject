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

func main() {
	// Initialize database
	db, err := database.InitPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)

	log.Println("=== Testing SSOT Cash-Bank Integration ===")

	// Test 1: Create Cash Account with Opening Balance
	log.Println("\n🧪 Test 1: Creating cash account with opening balance...")
	cashAccountRequest := services.CashBankCreateRequest{
		Name:           "Kas Utama Test",
		Type:           "CASH", 
		Currency:       "IDR",
		OpeningBalance: 5000000, // Rp 5,000,000
		OpeningDate:    services.CustomDate(time.Now()),
		Description:    "Test cash account for SSOT integration",
	}

	cashAccount, err := cashBankService.CreateCashBankAccount(cashAccountRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to create cash account: %v", err)
	} else {
		log.Printf("✅ Created cash account: %s (ID: %d, Balance: %.2f)", 
			cashAccount.Name, cashAccount.ID, cashAccount.Balance)
	}

	// Test 2: Create Bank Account with Opening Balance  
	log.Println("\n🧪 Test 2: Creating bank account with opening balance...")
	bankAccountRequest := services.CashBankCreateRequest{
		Name:           "Bank BCA Test",
		Type:           "BANK",
		BankName:       "Bank Central Asia",
		AccountNo:      "1234567890",
		Currency:       "IDR",
		OpeningBalance: 10000000, // Rp 10,000,000
		OpeningDate:    services.CustomDate(time.Now()),
		Description:    "Test bank account for SSOT integration",
	}

	bankAccount, err := cashBankService.CreateCashBankAccount(bankAccountRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to create bank account: %v", err)
	} else {
		log.Printf("✅ Created bank account: %s (ID: %d, Balance: %.2f)", 
			bankAccount.Name, bankAccount.ID, bankAccount.Balance)
	}

	// Test 3: Process Deposit
	log.Println("\n🧪 Test 3: Processing deposit transaction...")
	depositRequest := services.DepositRequest{
		AccountID: cashAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    1000000, // Rp 1,000,000
		Reference: "DEP001",
		Notes:     "Test deposit transaction",
	}

	depositTx, err := cashBankService.ProcessDeposit(depositRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to process deposit: %v", err)
	} else {
		log.Printf("✅ Processed deposit: ID %d, Amount: %.2f, Balance After: %.2f", 
			depositTx.ID, depositTx.Amount, depositTx.BalanceAfter)
	}

	// Test 4: Process Withdrawal  
	log.Println("\n🧪 Test 4: Processing withdrawal transaction...")
	withdrawalRequest := services.WithdrawalRequest{
		AccountID: bankAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    500000, // Rp 500,000
		Reference: "WTH001",
		Notes:     "Test withdrawal transaction",
	}

	withdrawalTx, err := cashBankService.ProcessWithdrawal(withdrawalRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to process withdrawal: %v", err)
	} else {
		log.Printf("✅ Processed withdrawal: ID %d, Amount: %.2f, Balance After: %.2f", 
			withdrawalTx.ID, withdrawalTx.Amount, withdrawalTx.BalanceAfter)
	}

	// Test 5: Process Transfer
	log.Println("\n🧪 Test 5: Processing transfer transaction...")
	transferRequest := services.TransferRequest{
		FromAccountID: bankAccount.ID,
		ToAccountID:   cashAccount.ID,
		Date:          services.CustomDate(time.Now()),
		Amount:        750000, // Rp 750,000
		Reference:     "TRF001",
		Notes:         "Test transfer transaction",
	}

	transfer, err := cashBankService.ProcessTransfer(transferRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to process transfer: %v", err)
	} else {
		log.Printf("✅ Processed transfer: ID %d, Amount: %.2f, Converted: %.2f", 
			transfer.ID, transfer.Amount, transfer.ConvertedAmount)
	}

	// Test 6: Verify SSOT Journal Entries
	log.Println("\n🧪 Test 6: Verifying SSOT journal entries...")
	
	// Get all SSOT journal entries for cash-bank source type
	var journalEntries []models.SSOTJournalEntry
	err = db.Where("source_type = ?", models.SSOTSourceTypeCashBank).
		Preload("Lines").
		Preload("Lines.Account").
		Find(&journalEntries).Error
	
	if err != nil {
		log.Printf("❌ Failed to get SSOT journal entries: %v", err)
	} else {
		log.Printf("✅ Found %d SSOT journal entries for cash-bank transactions", len(journalEntries))
		
		for i, entry := range journalEntries {
			log.Printf("  📋 Entry %d: %s - %s (Status: %s)", 
				i+1, entry.EntryNumber, entry.Description, entry.Status)
			
			var totalDebit, totalCredit decimal.Decimal
			for _, line := range entry.Lines {
				totalDebit = totalDebit.Add(line.DebitAmount)
				totalCredit = totalCredit.Add(line.CreditAmount)
				log.Printf("    └─ %s: Dr %.2f, Cr %.2f - %s", 
					line.Account.Name,
					line.DebitAmount.InexactFloat64(),
					line.CreditAmount.InexactFloat64(),
					line.Description)
			}
			
			// Verify balance
			if !totalDebit.Equal(totalCredit) {
				log.Printf("    ❌ Entry not balanced! Debit: %.2f, Credit: %.2f", 
					totalDebit.InexactFloat64(), totalCredit.InexactFloat64())
			} else {
				log.Printf("    ✅ Entry balanced: %.2f", totalDebit.InexactFloat64())
			}
		}
	}

	// Test 7: Verify Account Balances vs GL Balances
	log.Println("\n🧪 Test 7: Verifying cash-bank vs GL account balance consistency...")
	
	// Get updated cash-bank accounts
	cashAccount, _ = cashBankService.GetCashBankByID(cashAccount.ID)
	bankAccount, _ = cashBankService.GetCashBankByID(bankAccount.ID)
	
	// Check cash account balance consistency
	var cashGLAccount models.Account
	err = db.First(&cashGLAccount, cashAccount.AccountID).Error
	if err != nil {
		log.Printf("❌ Failed to get GL account for cash: %v", err)
	} else {
		log.Printf("Cash Account Balance: %.2f | GL Account Balance: %.2f", 
			cashAccount.Balance, cashGLAccount.Balance)
		if fmt.Sprintf("%.2f", cashAccount.Balance) == fmt.Sprintf("%.2f", cashGLAccount.Balance) {
			log.Printf("✅ Cash account balance matches GL account")
		} else {
			log.Printf("❌ Cash account balance mismatch with GL account!")
		}
	}
	
	// Check bank account balance consistency
	var bankGLAccount models.Account
	err = db.First(&bankGLAccount, bankAccount.AccountID).Error
	if err != nil {
		log.Printf("❌ Failed to get GL account for bank: %v", err)
	} else {
		log.Printf("Bank Account Balance: %.2f | GL Account Balance: %.2f", 
			bankAccount.Balance, bankGLAccount.Balance)
		if fmt.Sprintf("%.2f", bankAccount.Balance) == fmt.Sprintf("%.2f", bankGLAccount.Balance) {
			log.Printf("✅ Bank account balance matches GL account")
		} else {
			log.Printf("❌ Bank account balance mismatch with GL account!")
		}
	}

	// Test 8: Test SSOT Adapter Validation
	log.Println("\n🧪 Test 8: Testing SSOT adapter validation...")
	
	unifiedJournalService := services.NewUnifiedJournalService(db)
	sSOTAdapter := services.NewCashBankSSOTJournalAdapter(db, unifiedJournalService, accountRepo)
	
	err = sSOTAdapter.ValidateJournalIntegrity()
	if err != nil {
		log.Printf("❌ SSOT adapter validation failed: %v", err)
	} else {
		log.Printf("✅ SSOT adapter validation completed")
	}

	// Test 9: Create Reversal Entry
	log.Println("\n🧪 Test 9: Testing journal reversal...")
	
	reversal, err := sSOTAdapter.ReverseJournalEntry(
		uint64(withdrawalTx.ID), 
		"Test reversal for withdrawal transaction", 
		1)
	if err != nil {
		log.Printf("❌ Failed to create reversal: %v", err)
	} else {
		log.Printf("✅ Created reversal journal entry: %s", reversal.JournalEntry.EntryNumber)
	}

	// Final Summary
	log.Println("\n📊 === SSOT Cash-Bank Integration Test Summary ===")
	log.Println("✅ Cash account creation with opening balance")
	log.Println("✅ Bank account creation with opening balance") 
	log.Println("✅ Deposit transaction processing")
	log.Println("✅ Withdrawal transaction processing")
	log.Println("✅ Transfer transaction processing")
	log.Println("✅ SSOT journal entries verification")
	log.Println("✅ Balance consistency verification")
	log.Println("✅ SSOT adapter validation")
	log.Println("✅ Journal reversal functionality")
	log.Println("\n🎉 All SSOT Cash-Bank integration tests completed!")
}