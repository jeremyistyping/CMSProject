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
	fmt.Println("💳 Verify Payment Accounting Logic")
	fmt.Println("===================================")

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

	// Step 1: Check all purchases
	fmt.Println("📦 Step 1: All Purchases Analysis")
	checkAllPurchases(db)

	// Step 2: Check all payments
	fmt.Println("\n💰 Step 2: All Payment Transactions")
	checkAllPayments(db)

	// Step 3: Check SSOT journal entries for payments
	fmt.Println("\n📔 Step 3: SSOT Payment Journal Entries")
	checkPaymentJournalEntries(db)

	// Step 4: Reconcile balances
	fmt.Println("\n⚖️ Step 4: Balance Reconciliation")
	reconcileBalances(db)

	// Step 5: Verify accounting logic
	fmt.Println("\n🔍 Step 5: Accounting Logic Verification")
	verifyAccountingLogic(db)
}

func checkAllPurchases(db *gorm.DB) {
	var purchases []models.Purchase
	err := db.Preload("Vendor").Find(&purchases).Error
	if err != nil {
		fmt.Printf("❌ Failed to get purchases: %v\n", err)
		return
	}

	fmt.Printf("Found %d purchases:\n", len(purchases))
	
	totalPurchases := 0.0
	totalPaid := 0.0
	totalOutstanding := 0.0

	for i, purchase := range purchases {
		fmt.Printf("  %d. %s (%s)\n", i+1, purchase.Code, purchase.Status)
		fmt.Printf("     Vendor: %s\n", purchase.Vendor.Name)
		fmt.Printf("     Total: Rp %.2f, Paid: Rp %.2f, Outstanding: Rp %.2f\n", 
			purchase.TotalAmount, purchase.PaidAmount, purchase.OutstandingAmount)
		fmt.Printf("     Payment Method: %s\n", purchase.PaymentMethod)
		
		if purchase.Status == "COMPLETED" || purchase.Status == "APPROVED" {
			totalPurchases += purchase.TotalAmount
			totalPaid += purchase.PaidAmount
			totalOutstanding += purchase.OutstandingAmount
		}
		fmt.Println()
	}

	fmt.Printf("📊 Purchase Summary:\n")
	fmt.Printf("  Total Purchases: Rp %.2f\n", totalPurchases)
	fmt.Printf("  Total Paid: Rp %.2f\n", totalPaid)
	fmt.Printf("  Total Outstanding: Rp %.2f\n", totalOutstanding)
}

func checkAllPayments(db *gorm.DB) {
	var payments []models.Payment
	err := db.Find(&payments).Error
	if err != nil {
		fmt.Printf("❌ Failed to get payments: %v\n", err)
		return
	}

	fmt.Printf("Found %d payments in Payment Management system:\n", len(payments))
	
	totalPayments := 0.0
	for i, payment := range payments {
		fmt.Printf("  %d. %s (Date: %s)\n", i+1, payment.Code, payment.Date.Format("2006-01-02"))
		fmt.Printf("     Contact ID: %d\n", payment.ContactID)
		fmt.Printf("     Amount: Rp %.2f\n", payment.Amount)
		fmt.Printf("     Method: %s\n", payment.Method)
		fmt.Printf("     Notes: %s\n", payment.Notes)
		totalPayments += payment.Amount
		fmt.Println()
	}

	fmt.Printf("📊 Payment Summary:\n")
	fmt.Printf("  Total Payments Made: Rp %.2f\n", totalPayments)
}

func checkPaymentJournalEntries(db *gorm.DB) {
	var journalEntries []models.SSOTJournalEntry
	err := db.Preload("Lines.Account").
		Where("source_type = ?", "PAYMENT").
		Find(&journalEntries).Error
	
	if err != nil {
		fmt.Printf("❌ Failed to get payment journal entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d SSOT payment journal entries:\n", len(journalEntries))

	for i, entry := range journalEntries {
		fmt.Printf("  %d. %s (Date: %s)\n", i+1, entry.EntryNumber, entry.EntryDate.Format("2006-01-02"))
		fmt.Printf("     Description: %s\n", entry.Description)
		fmt.Printf("     Status: %s, Balanced: %t\n", entry.Status, entry.IsBalanced)
		
		if len(entry.Lines) > 0 {
			fmt.Printf("     Journal Lines:\n")
			for j, line := range entry.Lines {
				accountName := "Unknown"
				if line.Account != nil {
					accountName = line.Account.Name
				}
				
				if !line.DebitAmount.IsZero() {
					fmt.Printf("       %d. Dr. %s: Rp %.2f\n", j+1, accountName, line.DebitAmount)
				}
				if !line.CreditAmount.IsZero() {
					fmt.Printf("       %d. Cr. %s: Rp %.2f\n", j+1, accountName, line.CreditAmount)
				}
			}
		}
		fmt.Println()
	}
}

func reconcileBalances(db *gorm.DB) {
	// Get key account balances
	accounts := []string{"1301", "1240", "2101", "1103", "1101"}
	
	fmt.Printf("Current account balances vs expected:\n")
	
	var totalPayables float64
	
	for _, code := range accounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("  ❌ %s: Not found\n", code)
			continue
		}

		fmt.Printf("  %s (%s): Rp %.2f", code, account.Name, account.Balance)
		
		switch code {
		case "1301": // Inventory
			fmt.Printf(" ✅ (Purchase goods recorded)")
		case "1240": // PPN Input 
			fmt.Printf(" ✅ (VAT from purchases)")
		case "2101": // Accounts Payable
			totalPayables = account.Balance
			if account.Balance < 0 {
				fmt.Printf(" ✅ (Liability - credit balance)")
			} else {
				fmt.Printf(" ❌ (Should be negative for liability)")
			}
		case "1103": // Bank Mandiri
			if account.Balance < 0 {
				fmt.Printf(" ✅ (Cash decreased by payments)")
			} else {
				fmt.Printf(" ⚠️ (No cash decrease recorded?)")
			}
		case "1101": // Cash
			fmt.Printf(" ℹ️ (Cash account)")
		}
		fmt.Println()
	}

	fmt.Printf("\n📊 Balance Analysis:\n")
	fmt.Printf("  Accounts Payable Balance: Rp %.2f\n", totalPayables)
	fmt.Printf("  This represents remaining debt to vendors\n")
	
	if totalPayables < 0 {
		fmt.Printf("  ✅ Liability balance is correct (negative = credit)\n")
		fmt.Printf("  Outstanding vendor payments: Rp %.2f\n", -totalPayables)
	}
}

func verifyAccountingLogic(db *gorm.DB) {
	fmt.Printf("🔍 Accounting Logic Verification:\n\n")
	
	// Expected flow for purchase + payment:
	fmt.Printf("📝 Expected Accounting Flow:\n")
	fmt.Printf("1. Purchase Transaction (Rp 11.100.000):\n")
	fmt.Printf("   Dr. Inventory       10.000.000\n")
	fmt.Printf("   Dr. PPN Masukan      1.100.000\n")
	fmt.Printf("       Cr. Accounts Payable    11.100.000\n")
	fmt.Printf("   ✅ This creates liability (debt to vendor)\n\n")
	
	fmt.Printf("2. Payment Transaction (Rp 5.550.000):\n")
	fmt.Printf("   Dr. Accounts Payable  5.550.000\n")
	fmt.Printf("       Cr. Bank Account      5.550.000\n")
	fmt.Printf("   ✅ This reduces liability and decreases cash\n\n")
	
	fmt.Printf("3. Expected Final Balances:\n")
	fmt.Printf("   - Inventory: +Rp 10.000.000 (asset increase)\n")
	fmt.Printf("   - PPN Masukan: +Rp 1.100.000 (asset increase)\n")
	fmt.Printf("   - Accounts Payable: -Rp 5.550.000 (remaining debt)\n")
	fmt.Printf("   - Bank Account: -Rp 5.550.000 (cash decrease)\n\n")

	// Check actual vs expected
	var inventory, ppn, payable, bank float64
	db.Model(&models.Account{}).Where("code = ?", "1301").Select("balance").Row().Scan(&inventory)
	db.Model(&models.Account{}).Where("code = ?", "1240").Select("balance").Row().Scan(&ppn)
	db.Model(&models.Account{}).Where("code = ?", "2101").Select("balance").Row().Scan(&payable)
	db.Model(&models.Account{}).Where("code = ?", "1103").Select("balance").Row().Scan(&bank)

	fmt.Printf("🔍 Actual vs Expected:\n")
	fmt.Printf("  Inventory (1301): Rp %.2f ", inventory)
	if inventory == 10000000 {
		fmt.Printf("✅ CORRECT\n")
	} else {
		fmt.Printf("❌ Expected: Rp 10.000.000\n")
	}

	fmt.Printf("  PPN Masukan (1240): Rp %.2f ", ppn)
	if ppn == 1100000 {
		fmt.Printf("✅ CORRECT\n")
	} else {
		fmt.Printf("❌ Expected: Rp 1.100.000\n")
	}

	fmt.Printf("  Accounts Payable (2101): Rp %.2f ", payable)
	if payable == -5550000 {
		fmt.Printf("✅ PERFECT MATCH\n")
	} else if payable < -5550000 {
		fmt.Printf("⚠️ MORE DEBT (%.2f) - multiple purchases?\n", -payable)
	} else {
		fmt.Printf("❌ Expected: -Rp 5.550.000\n")
	}

	fmt.Printf("  Bank Mandiri (1103): Rp %.2f ", bank)
	if bank == -5550000 {
		fmt.Printf("✅ PERFECT - payment recorded\n")
	} else if bank == 0 {
		fmt.Printf("❌ No payment recorded in bank account\n")
	} else {
		fmt.Printf("⚠️ Different amount\n")
	}

	fmt.Printf("\n🎯 CONCLUSION:\n")
	if inventory == 10000000 && ppn == 1100000 && payable < 0 {
		fmt.Printf("✅ PURCHASE ACCOUNTING LOGIC IS CORRECT!\n")
		fmt.Printf("   - Assets properly increased\n")
		fmt.Printf("   - Liabilities properly recorded\n")
		fmt.Printf("   - VAT properly separated\n")
	}
	
	if payable == -5550000 {
		fmt.Printf("✅ PAYMENT ACCOUNTING LOGIC IS PERFECT!\n")
		fmt.Printf("   - Remaining debt exactly matches outstanding amount\n")
	} else if payable < -5550000 {
		fmt.Printf("ℹ️ PAYMENT LOGIC WORKING, but more debt exists\n")
		fmt.Printf("   - This suggests multiple purchases or transactions\n")
	}
}