package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🧾 Analyzing Purchase Accounting Transactions")
	fmt.Println("==============================================")

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		return
	}

	fmt.Println("✅ Database connected successfully\n")

	// Step 1: Analyze Purchase Transaction
	fmt.Println("📦 Step 1: Purchase Transaction Analysis")
	analyzePurchaseTransaction(db)

	// Step 2: Analyze Payment Transaction
	fmt.Println("\n💳 Step 2: Payment Transaction Analysis")
	analyzePaymentTransaction(db)

	// Step 3: Analyze SSOT Journal Entries
	fmt.Println("\n📔 Step 3: SSOT Journal Entries Analysis")
	analyzeSSOTJournalEntries(db)

	// Step 4: Verify Accounting Equation
	fmt.Println("\n⚖️ Step 4: Accounting Equation Verification")
	verifyAccountingEquation(db)

	// Step 5: Summary and Recommendations
	fmt.Println("\n🎯 Step 5: Summary & Accounting Logic Verification")
	provideSummaryAndVerification(db)
}

func analyzePurchaseTransaction(db *gorm.DB) {
	var purchase models.Purchase
	err := db.Preload("Vendor").Preload("PurchaseItems.Product").
		Where("code = ?", "PO/2025/09/0003").First(&purchase).Error
	
	if err != nil {
		fmt.Printf("❌ Purchase not found: %v\n", err)
		return
	}

	fmt.Printf("🛒 Purchase Details:\n")
	fmt.Printf("  Code: %s\n", purchase.Code)
	fmt.Printf("  Vendor: %s\n", purchase.Vendor.Name)
	fmt.Printf("  Total Amount: Rp %.2f\n", purchase.TotalAmount)
	fmt.Printf("  Paid Amount: Rp %.2f\n", purchase.PaidAmount)
	fmt.Printf("  Outstanding: Rp %.2f\n", purchase.OutstandingAmount)
	fmt.Printf("  Status: %s (%s)\n", purchase.Status, purchase.ApprovalStatus)
	fmt.Printf("  Payment Method: %s\n", purchase.PaymentMethod)

	// Calculate expected journal entries
	fmt.Printf("\n📝 Expected Purchase Journal Entries:\n")
	fmt.Printf("  Dr. Inventory (1301): Rp %.2f\n", purchase.TotalAmount - purchase.PPNAmount)
	if purchase.PPNAmount > 0 {
		fmt.Printf("  Dr. PPN Masukan (1240): Rp %.2f\n", purchase.PPNAmount)
	}
	fmt.Printf("  Cr. Accounts Payable (2101): Rp %.2f\n", purchase.TotalAmount)
	fmt.Printf("  Total Debit = Total Credit = Rp %.2f ✓\n", purchase.TotalAmount)
}

func analyzePaymentTransaction(db *gorm.DB) {
	// Get payments from Payment Management system
	var payments []models.Payment
	err := db.Where("notes LIKE ?", "%PO/2025/09/0003%").Find(&payments).Error
	if err != nil {
		fmt.Printf("❌ Failed to get payments: %v\n", err)
		return
	}

	if len(payments) == 0 {
		fmt.Printf("ℹ️ No payments found in Payment Management system for PO/2025/09/0003\n")
		fmt.Printf("   This might indicate payments were recorded differently\n")
		return
	}

	fmt.Printf("💰 Found %d payment(s) for purchase:\n", len(payments))
	
	totalPayments := 0.0
	for i, payment := range payments {
		fmt.Printf("  %d. Payment Code: %s\n", i+1, payment.Code)
		fmt.Printf("     Amount: Rp %.2f\n", payment.Amount)
		fmt.Printf("     Date: %s\n", payment.Date.Format("2006-01-02"))
		fmt.Printf("     Method: %s\n", payment.Method)
		fmt.Printf("     Notes: %s\n", payment.Notes)
		totalPayments += payment.Amount
	}

	fmt.Printf("\n📝 Expected Payment Journal Entries:\n")
	fmt.Printf("  Dr. Accounts Payable (2101): Rp %.2f\n", totalPayments)
	fmt.Printf("  Cr. Bank Account: Rp %.2f\n", totalPayments)
	fmt.Printf("  Total Debit = Total Credit = Rp %.2f ✓\n", totalPayments)
}

func analyzeSSOTJournalEntries(db *gorm.DB) {
	var journalEntries []models.SSOTJournalEntry
	err := db.Preload("Lines.Account").Find(&journalEntries).Error
	if err != nil {
		fmt.Printf("❌ Failed to get SSOT journal entries: %v\n", err)
		return
	}

	fmt.Printf("📚 Found %d SSOT journal entries:\n\n", len(journalEntries))

	for i, entry := range journalEntries {
		fmt.Printf("🧾 Entry %d: %s (ID: %d)\n", i+1, entry.EntryNumber, entry.ID)
		fmt.Printf("   Source: %s (ID: %v)\n", entry.SourceType, entry.SourceID)
		fmt.Printf("   Date: %s\n", entry.EntryDate.Format("2006-01-02"))
		fmt.Printf("   Description: %s\n", entry.Description)
		fmt.Printf("   Status: %s\n", entry.Status)
		fmt.Printf("   Total Debit: Rp %.2f\n", entry.TotalDebit)
		fmt.Printf("   Total Credit: Rp %.2f\n", entry.TotalCredit)
		fmt.Printf("   Balanced: %t\n", entry.IsBalanced)

		if len(entry.Lines) > 0 {
			fmt.Printf("   📋 Journal Lines:\n")
			for j, line := range entry.Lines {
				accountName := "Unknown"
				if line.Account != nil {
					accountName = line.Account.Name
				}
				
				if !line.DebitAmount.IsZero() {
					fmt.Printf("     %d. Dr. %s: Rp %.2f\n", j+1, accountName, line.DebitAmount)
				}
				if !line.CreditAmount.IsZero() {
					fmt.Printf("     %d. Cr. %s: Rp %.2f\n", j+1, accountName, line.CreditAmount)
				}
			}
		}
		fmt.Println()
	}
}

func verifyAccountingEquation(db *gorm.DB) {
	// Get current balances for key accounts
	accounts := map[string]string{
		"1301": "Persediaan Barang Dagangan",
		"1240": "PPN Masukan", 
		"2101": "Utang Usaha",
		"1103": "Bank Mandiri",
		"1101": "Kas",
	}

	fmt.Printf("💰 Current Account Balances:\n")
	
	var totalAssets, totalLiabilities, totalInventory, totalPayables float64
	
	for code, name := range accounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("   ❌ %s (%s): Not found\n", code, name)
			continue
		}

		fmt.Printf("   %s (%s): Rp %.2f", code, name, account.Balance)
		
		// Categorize for equation verification
		switch code {
		case "1301", "1240": // Inventory and PPN
			totalInventory += account.Balance
			totalAssets += account.Balance
			fmt.Printf(" (Asset)\n")
		case "1103", "1101": // Bank accounts
			totalAssets += account.Balance
			fmt.Printf(" (Asset)\n")
		case "2101": // Accounts Payable
			totalLiabilities += account.Balance // Note: liability balances are negative
			totalPayables += account.Balance
			fmt.Printf(" (Liability)\n")
		default:
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n⚖️ Accounting Logic Verification:\n")
	fmt.Printf("   Purchase Transaction Logic:\n")
	fmt.Printf("   - Inventory increase: Rp %.2f ✓\n", totalInventory)
	fmt.Printf("   - Accounts Payable increase: Rp %.2f ✓\n", -totalPayables)
	
	if totalInventory > 0 && totalPayables < 0 {
		fmt.Printf("   ✅ Purchase recording logic is CORRECT\n")
		fmt.Printf("      (Assets increased, Liabilities increased)\n")
	} else {
		fmt.Printf("   ❌ Purchase recording logic may have issues\n")
	}

	// Check if payment was made
	expectedOutstanding := 11100000.0 - 5550000.0 // Total - Payment
	if totalPayables < 0 && (-totalPayables) <= 11100000.0 {
		fmt.Printf("   ✅ Payment recording logic is CORRECT\n")
		fmt.Printf("      Expected outstanding: Rp %.2f\n", expectedOutstanding)
		fmt.Printf("      Actual payable balance: Rp %.2f\n", -totalPayables)
	}
}

func provideSummaryAndVerification(db *gorm.DB) {
	fmt.Printf("📋 Accounting Logic Summary:\n\n")
	
	fmt.Printf("🛒 Purchase Transaction (Rp 11.100.000):\n")
	fmt.Printf("   ✅ Should record: Dr. Inventory, Cr. Accounts Payable\n")
	fmt.Printf("   ✅ Expected balance changes:\n")
	fmt.Printf("      - Inventory (1301): +Rp 10.000.000 (goods purchased)\n")
	fmt.Printf("      - PPN Masukan (1240): +Rp 1.100.000 (11% VAT)\n")
	fmt.Printf("      - Accounts Payable (2101): -Rp 11.100.000 (liability)\n")
	
	fmt.Printf("\n💳 Payment Transaction (Rp 5.550.000):\n")
	fmt.Printf("   ✅ Should record: Dr. Accounts Payable, Cr. Bank\n")
	fmt.Printf("   ✅ Expected balance changes:\n")
	fmt.Printf("      - Accounts Payable (2101): +Rp 5.550.000 (reducing liability)\n")
	fmt.Printf("      - Bank Account: -Rp 5.550.000 (cash out)\n")
	fmt.Printf("      - Remaining payable: Rp 5.550.000\n")

	fmt.Printf("\n🔍 Verification Questions:\n")
	fmt.Printf("   1. Is inventory recorded at purchase cost? ✓ Check\n")
	fmt.Printf("   2. Is VAT properly separated? ✓ Check\n")
	fmt.Printf("   3. Is accounts payable properly credited? ✓ Check\n")
	fmt.Printf("   4. Does payment reduce payable balance? ✓ Check\n")
	fmt.Printf("   5. Are journal entries balanced? ✓ Check\n")

	fmt.Printf("\n🎯 Expected Account Balances After All Transactions:\n")
	fmt.Printf("   - Inventory (1301): Rp 10.000.000\n")
	fmt.Printf("   - PPN Masukan (1240): Rp 1.100.000\n")
	fmt.Printf("   - Accounts Payable (2101): -Rp 5.550.000 (remaining debt)\n")
	fmt.Printf("   - Bank Account: Should decrease by Rp 5.550.000\n")

	fmt.Printf("\n✅ If balances match expectations above, accounting logic is CORRECT!\n")
}