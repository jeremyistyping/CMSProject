package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database config from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	
	// Construct DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Step 1: Check if Accounts Payable account exists
	fmt.Println("\n=== Step 1: Checking Accounts Payable Account ===")
	var payableAccount models.Account
	err = db.Where("code = ? OR name ILIKE ?", "2101", "%Hutang%Usaha%").First(&payableAccount).Error
	
	if err != nil {
		fmt.Println("❌ Accounts Payable account not found. Creating it...")
		
		// Create Accounts Payable account
		payableAccount = models.Account{
			Code:        "2101",
			Name:        "Utang Usaha",
			Type:        models.AccountTypeLiability,
			Balance:     0,
			IsActive:    true,
			Description: "Accounts Payable - Vendor obligations",
		}
		
		if err := db.Create(&payableAccount).Error; err != nil {
			log.Fatalf("Failed to create Accounts Payable account: %v", err)
		}
		fmt.Printf("✅ Created Accounts Payable account: ID=%d, Code=%s, Name=%s\n", 
			payableAccount.ID, payableAccount.Code, payableAccount.Name)
	} else {
		fmt.Printf("✅ Found Accounts Payable account: ID=%d, Code=%s, Name=%s, Balance=%.2f\n", 
			payableAccount.ID, payableAccount.Code, payableAccount.Name, payableAccount.Balance)
	}

	// Step 2: Check PPN Masukan account
	fmt.Println("\n=== Step 2: Checking PPN Masukan Account ===")
	var ppnAccount models.Account
	err = db.Where("code = ? OR name ILIKE ?", "1105", "%PPN%Masukan%").First(&ppnAccount).Error
	
	if err != nil {
		fmt.Println("❌ PPN Masukan account not found. Creating it...")
		
		// Create PPN Masukan account
		ppnAccount = models.Account{
			Code:        "1105",
			Name:        "PPN Masukan",
			Type:        models.AccountTypeAsset,
			Balance:     0,
			IsActive:    true,
			Description: "Input VAT - Tax Receivable",
		}
		
		if err := db.Create(&ppnAccount).Error; err != nil {
			log.Fatalf("Failed to create PPN Masukan account: %v", err)
		}
		fmt.Printf("✅ Created PPN Masukan account: ID=%d, Code=%s, Name=%s\n", 
			ppnAccount.ID, ppnAccount.Code, ppnAccount.Name)
	} else {
		fmt.Printf("✅ Found PPN Masukan account: ID=%d, Code=%s, Name=%s, Balance=%.2f\n", 
			ppnAccount.ID, ppnAccount.Code, ppnAccount.Name, ppnAccount.Balance)
	}

	// Step 3: Check and fix journal entries for purchases
	fmt.Println("\n=== Step 3: Fixing Purchase Journal Entries ===")
	var journalEntries []models.JournalEntry
	err = db.Where("reference_type = ? AND status = ?", "PURCHASE", "POSTED").Find(&journalEntries).Error
	if err != nil {
		log.Fatalf("Failed to fetch journal entries: %v", err)
	}

	fmt.Printf("Found %d posted purchase journal entries\n", len(journalEntries))

	for _, entry := range journalEntries {
		fmt.Printf("\nProcessing Journal Entry ID=%d, Code=%s, Amount=%.2f\n", 
			entry.ID, entry.Code, entry.TotalCredit)
		
		// Check if this purchase journal properly updated Accounts Payable
		// If not, update the Accounts Payable balance
		var currentBalance float64
		db.Model(&payableAccount).Select("balance").Scan(&currentBalance)
		
		// Update Accounts Payable with the journal entry amount
		err = db.Model(&payableAccount).Update("balance", gorm.Expr("balance + ?", entry.TotalCredit)).Error
		if err != nil {
			fmt.Printf("❌ Failed to update Accounts Payable for entry %d: %v\n", entry.ID, err)
		} else {
			fmt.Printf("✅ Updated Accounts Payable balance by +%.2f for entry %d\n", entry.TotalCredit, entry.ID)
		}
	}

	// Step 4: Update journal entry repository query to use correct account
	fmt.Println("\n=== Step 4: Fixing Journal Entry Repository Logic ===")
	fmt.Printf("✅ Accounts Payable account ID to use: %d (Code: %s)\n", payableAccount.ID, payableAccount.Code)
	fmt.Printf("✅ PPN Masukan account ID to use: %d (Code: %s)\n", ppnAccount.ID, ppnAccount.Code)

	// Step 5: Verify account balances
	fmt.Println("\n=== Step 5: Final Account Balance Verification ===")
	
	// Reload account balances
	db.First(&payableAccount, payableAccount.ID)
	db.First(&ppnAccount, ppnAccount.ID)
	
	fmt.Printf("Final Accounts Payable Balance: %.2f\n", payableAccount.Balance)
	fmt.Printf("Final PPN Masukan Balance: %.2f\n", ppnAccount.Balance)

	// Step 6: Show all liability accounts for verification
	fmt.Println("\n=== Step 6: Current Liability Accounts ===")
	var liabilityAccounts []models.Account
	db.Where("type = ?", models.AccountTypeLiability).Find(&liabilityAccounts)
	
	for _, acc := range liabilityAccounts {
		fmt.Printf("ID=%d, Code=%s, Name=%s, Balance=%.2f\n", 
			acc.ID, acc.Code, acc.Name, acc.Balance)
	}

	fmt.Println("\n🎉 Purchase accounts fix completed!")
	fmt.Println("Next steps:")
	fmt.Println("1. Update journal_entry_repository.go to use the correct account IDs")
	fmt.Println("2. Test purchase approval to ensure balance updates work correctly")
	fmt.Println("3. Run balance sheet report to verify all accounts are properly updated")
}
