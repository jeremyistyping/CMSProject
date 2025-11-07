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
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Debug payment journal creation
	if err := debugPaymentJournal(db); err != nil {
		log.Fatal("Debug failed:", err)
	}

	fmt.Println("✅ Payment journal debug completed!")
}

func debugPaymentJournal(db *gorm.DB) error {
	fmt.Println("🔍 Debugging Payment Journal Creation")

	// Find recent payments
	fmt.Println("\n📋 Step 1: Find recent payments")
	var payments []models.Payment
	err := db.Order("created_at DESC").Limit(3).Find(&payments).Error
	if err != nil {
		return fmt.Errorf("failed to get recent payments: %v", err)
	}

	fmt.Printf("  ✓ Found %d recent payments\n", len(payments))
	for i, payment := range payments {
		fmt.Printf("    %d. Payment %s (ID: %d) - Amount: %.2f - Status: %s\n", 
			i+1, payment.Code, payment.ID, payment.Amount, payment.Status)
	}

	if len(payments) == 0 {
		return fmt.Errorf("no payments found to debug")
	}

	// Focus on the latest payment
	latestPayment := payments[0]
	fmt.Printf("\n🎯 Focusing on payment: %s (ID: %d)\n", latestPayment.Code, latestPayment.ID)

	// Check if legacy journal entries exist for this payment
	fmt.Println("\n📗 Step 2: Check legacy journal entries")
	var journalEntries []models.JournalEntry
	err = db.Where("reference_type = ? AND reference_id = ?", models.JournalRefPayment, latestPayment.ID).Find(&journalEntries).Error
	if err != nil {
		fmt.Printf("  ⚠️ Error checking journal entries: %v\n", err)
	} else {
		fmt.Printf("  📝 Found %d legacy journal entries for payment %s\n", len(journalEntries), latestPayment.Code)
		
		for _, entry := range journalEntries {
			fmt.Printf("    - Entry: %s | Status: %s | Debit: %.2f | Credit: %.2f\n", 
				entry.Code, entry.Status, entry.TotalDebit, entry.TotalCredit)
			
			// Get journal lines
			var lines []models.JournalLine
			db.Where("journal_entry_id = ?", entry.ID).Find(&lines)
			for _, line := range lines {
				var account models.Account
				db.First(&account, line.AccountID)
				fmt.Printf("      L%d: %s (%s) - Debit: %.2f | Credit: %.2f\n", 
					line.LineNumber, account.Code, account.Name, line.DebitAmount, line.CreditAmount)
			}
		}
	}

	// Check SSOT journal entries
	fmt.Println("\n🌟 Step 3: Check SSOT journal entries")
	
	// Check simple_ssot_journals first
	type SimpleSSOTJournal struct {
		ID          uint      `json:"id"`
		Status      string    `json:"status"`
		EntryDate   string    `json:"entry_date"`
		Description string    `json:"description"`
	}
	
	var ssotJournals []SimpleSSOTJournal
	err = db.Table("simple_ssot_journals").Where("description LIKE ?", 
		fmt.Sprintf("%%Payment%%for%%Purchase%%")).Find(&ssotJournals).Error
	if err != nil {
		fmt.Printf("  ⚠️ Error checking simple_ssot_journals: %v\n", err)
	} else {
		fmt.Printf("  📝 Found %d SSOT journal entries related to payments\n", len(ssotJournals))
		for _, journal := range ssotJournals {
			fmt.Printf("    - SSOT Journal ID: %d | Status: %s | Description: %s\n", 
				journal.ID, journal.Status, journal.Description)
		}
	}

	// Check if there are any SSOT journal entries with payment source type
	var ssotPaymentJournals []models.SSOTJournalEntry
	err = db.Where("source_type = ?", models.SSOTSourceTypePayment).Find(&ssotPaymentJournals).Error
	if err != nil {
		fmt.Printf("  ⚠️ Error checking SSOTJournalEntry: %v\n", err)
	} else {
		fmt.Printf("  📝 Found %d SSOT journal entries with PAYMENT source type\n", len(ssotPaymentJournals))
		for _, journal := range ssotPaymentJournals {
			fmt.Printf("    - SSOT Entry: %s | Source ID: %d | Status: %s\n", 
				journal.EntryNumber, journal.SourceID, journal.Status)
		}
	}

	// Check if any account balances have changed recently
	fmt.Println("\n💰 Step 4: Check recent account balance changes")
	
	// Get Utang Usaha account info
	var utangAccount models.Account
	err = db.Where("code = ?", "2101").First(&utangAccount).Error
	if err != nil {
		fmt.Printf("  ⚠️ Utang Usaha account (2101) not found: %v\n", err)
	} else {
		fmt.Printf("  💳 Utang Usaha (2101) current balance: %.2f\n", utangAccount.Balance)
		fmt.Printf("      - Name: %s\n", utangAccount.Name)
		fmt.Printf("      - Type: %s\n", utangAccount.Type)
		fmt.Printf("      - Updated: %v\n", utangAccount.UpdatedAt)
	}

	// Get cash accounts info
	var kasAccount models.Account  
	err = db.Where("code = ?", "1101").First(&kasAccount).Error
	if err != nil {
		fmt.Printf("  ⚠️ Kas account (1101) not found: %v\n", err)
	} else {
		fmt.Printf("  💰 Kas (1101) current balance: %.2f\n", kasAccount.Balance)
		fmt.Printf("      - Updated: %v\n", kasAccount.UpdatedAt)
	}

	// Check if payment service is properly configured
	fmt.Println("\n⚙️ Step 5: Check payment service configuration")
	
	// Look for any configuration or flags that might be disabling journal creation
	fmt.Printf("  💡 The payment service logs show: 'Skipping legacy journal creation - SSOT journal will be created separately'\n")
	fmt.Printf("  💡 This means payment service expects SSOT to handle journal creation\n")
	fmt.Printf("  💡 But SSOT journal creation might not be properly integrated\n")

	// Check if manual balance updates are being made anywhere
	fmt.Println("\n🔧 Step 6: Recommendations")
	fmt.Printf("  1. The payment service is NOT creating journal entries that update COA balances\n")
	fmt.Printf("  2. The comment says 'SSOT journal will be created separately' but this isn't happening\n")
	fmt.Printf("  3. Only cash/bank balance is updated directly in cash_banks table\n")
	fmt.Printf("  4. Utang Usaha balance in accounts table is NOT being updated\n")
	
	fmt.Printf("\n💡 Solution needed:\n")
	fmt.Printf("  - Either enable legacy journal creation in payment service\n")
	fmt.Printf("  - OR properly integrate SSOT journal creation for payments\n")
	fmt.Printf("  - OR manually update COA account balances in payment service\n")

	return nil
}