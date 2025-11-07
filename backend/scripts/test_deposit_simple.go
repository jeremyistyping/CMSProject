package main

import (
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Println("🧪 Simple Deposit Test")
	log.Println("======================")

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	log.Println("✅ Database connected successfully")

	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)

	// Get first available cash bank account
	log.Println("🔍 Finding available cash-bank account...")
	accounts, err := cashBankService.GetCashBankAccounts()
	if err != nil {
		log.Printf("❌ Failed to get accounts: %v", err)
		return
	}

	if len(accounts) == 0 {
		log.Println("❌ No cash-bank accounts found. Creating one first...")
		
		// Create a test account
		createRequest := services.CashBankCreateRequest{
			Name:        "Test Cash Account",
			Type:        "CASH",
			Currency:    "IDR",
			Description: "Test account for deposit",
		}
		
		newAccount, err := cashBankService.CreateCashBankAccount(createRequest, 1)
		if err != nil {
			log.Printf("❌ Failed to create test account: %v", err)
			return
		}
		log.Printf("✅ Created test account: %s (ID: %d)", newAccount.Name, newAccount.ID)
		accounts = []models.CashBank{*newAccount}
	}

	testAccount := accounts[0]
	log.Printf("📋 Using account: %s (ID: %d, Balance: %.2f)", 
		testAccount.Name, testAccount.ID, testAccount.Balance)

	// Test deposit
	log.Println("\n💰 Testing deposit transaction...")
	depositRequest := services.DepositRequest{
		AccountID: testAccount.ID,
		Date:      services.CustomDate(time.Now()),
		Amount:    100000, // Rp 100,000
		Reference: "TEST-DEP-001",
		Notes:     "Test deposit transaction",
	}

	// Measure time
	start := time.Now()
	transaction, err := cashBankService.ProcessDeposit(depositRequest, 1)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Deposit failed after %v: %v", duration, err)
		return
	}

	log.Printf("🎉 Deposit successful! Duration: %v", duration)
	log.Printf("📊 Transaction details:")
	log.Printf("   - ID: %d", transaction.ID)
	log.Printf("   - Amount: %.2f", transaction.Amount)
	log.Printf("   - Balance After: %.2f", transaction.BalanceAfter)
	log.Printf("   - Reference: %s", transaction.ReferenceType)
	log.Printf("   - Notes: %s", transaction.Notes)

	// Verify account balance
	updatedAccount, err := cashBankService.GetCashBankByID(testAccount.ID)
	if err != nil {
		log.Printf("⚠️ Warning: Could not verify updated balance: %v", err)
	} else {
		log.Printf("✅ Account balance verified: %.2f", updatedAccount.Balance)
	}

	log.Println("\n✅ Deposit test completed successfully!")
}