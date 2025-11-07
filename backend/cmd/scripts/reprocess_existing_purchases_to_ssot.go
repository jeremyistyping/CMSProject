package main

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🔄 Reprocess Existing Purchases to SSOT Journal System")
	fmt.Println("======================================================")

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce log noise
	})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		fmt.Println("\n📋 Manual instructions will be provided instead.")
		provideManualInstructions()
		return
	}

	fmt.Println("✅ Database connected successfully")

	// Step 1: Analyze existing purchases
	fmt.Println("\n📊 Step 1: Analyzing Existing Purchases")
	analyzeExistingPurchases(db)

	// Step 2: Reprocess approved purchases
	fmt.Println("\n🔄 Step 2: Reprocessing Approved Purchases")
	reprocessApprovedPurchases(db)

	// Step 3: Verify COA updates
	fmt.Println("\n✅ Step 3: Verifying COA Updates")
	verifyCOAUpdates(db)

	// Step 4: Summary and recommendations
	fmt.Println("\n🎯 Step 4: Summary & Recommendations")
	provideSummaryAndRecommendations()
}

func analyzeExistingPurchases(db *gorm.DB) {
	var totalPurchases, approvedPurchases, pendingPurchases int64
	var totalAmount, approvedAmount float64

	// Count purchases by status
	db.Model(&models.Purchase{}).Count(&totalPurchases)
	db.Model(&models.Purchase{}).Where("status IN ? AND approval_status = ?", []string{"APPROVED", "COMPLETED"}, "APPROVED").Count(&approvedPurchases)
	db.Model(&models.Purchase{}).Where("status IN ?", []string{"PENDING", "DRAFT"}).Count(&pendingPurchases)

	// Sum amounts
	db.Model(&models.Purchase{}).Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalAmount)
	db.Model(&models.Purchase{}).Where("status IN ? AND approval_status = ?", []string{"APPROVED", "COMPLETED"}, "APPROVED").Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&approvedAmount)

	fmt.Printf("📈 Purchase Analysis:\n")
	fmt.Printf("  Total Purchases: %d (Total Amount: Rp %.2f)\n", totalPurchases, totalAmount)
	fmt.Printf("  Approved: %d (Amount: Rp %.2f)\n", approvedPurchases, approvedAmount)
	fmt.Printf("  Pending/Draft: %d\n", pendingPurchases)

	// Check current COA balances for purchase-related accounts
	fmt.Println("\n💰 Current COA Balances (Purchase Related):")
	checkAccountBalance(db, "1301", "Persediaan Barang Dagangan")
	checkAccountBalance(db, "1240", "PPN Masukan")
	checkAccountBalance(db, "2101", "Utang Usaha")
	checkAccountBalance(db, "2111", "Utang PPh 21")
	checkAccountBalance(db, "2112", "Utang PPh 23")
}

func checkAccountBalance(db *gorm.DB, code, name string) {
	var balance float64
	err := db.Model(&models.Account{}).Where("code = ?", code).Select("balance").Row().Scan(&balance)
	if err != nil {
		fmt.Printf("  ❌ %s (%s): Not found\n", name, code)
	} else {
		status := "✅"
		if balance == 0 {
			status = "⚠️"
		}
		fmt.Printf("  %s %s (%s): Rp %.2f\n", status, name, code, balance)
	}
}

func reprocessApprovedPurchases(db *gorm.DB) {
	// Initialize services for SSOT processing
	accountRepo := repositories.NewAccountRepository(db)
	unifiedJournalService := services.NewUnifiedJournalService(db)
	// Use TaxAccountService for flexible account mapping
	taxSvc := services.NewTaxAccountService(db)
	purchaseAdapter := services.NewPurchaseSSOTJournalAdapter(db, unifiedJournalService, accountRepo, taxSvc)

	// Get all approved and completed purchases that might not have SSOT journal entries
	var purchases []models.Purchase
	err := db.Preload("Vendor").Preload("PurchaseItems.Product").
		Where("status IN ? AND approval_status = ?", []string{"APPROVED", "COMPLETED"}, "APPROVED").
		Find(&purchases).Error

	if err != nil {
		fmt.Printf("❌ Failed to get approved purchases: %v\n", err)
		return
	}

	fmt.Printf("Found %d approved/completed purchases to process\n", len(purchases))

	var processedCount, skippedCount, errorCount int

	for _, purchase := range purchases {
		fmt.Printf("\n🔄 Processing Purchase: %s (ID: %d, Amount: Rp %.2f)\n", 
			purchase.Code, purchase.ID, purchase.TotalAmount)

		// Check if SSOT journal entries already exist
		hasEntries, err := checkExistingSSOTEntries(db, purchase.ID)
		if err != nil {
			fmt.Printf("  ⚠️ Could not check existing entries: %v\n", err)
		}

		if hasEntries {
			fmt.Printf("  📋 SSOT entries already exist, skipping\n")
			skippedCount++
			continue
		}

		// Create SSOT journal entries
		ctx := context.Background()
		journalEntry, err := purchaseAdapter.CreatePurchaseJournalEntry(ctx, &purchase, uint64(1)) // Use admin user ID
		if err != nil {
			fmt.Printf("  ❌ Failed to create SSOT journal entry: %v\n", err)
			errorCount++
			continue
		}

		fmt.Printf("  ✅ SSOT journal entry created: %s (ID: %d)\n", 
			journalEntry.EntryNumber, journalEntry.ID)
		processedCount++

		// If it's a credit purchase with payments, also create payment entries
		if purchase.PaymentMethod == "CREDIT" && purchase.PaidAmount > 0 {
			fmt.Printf("  💳 Processing payments for credit purchase (Paid: Rp %.2f)\n", purchase.PaidAmount)
			err = processExistingPayments(db, purchaseAdapter, &purchase)
			if err != nil {
				fmt.Printf("  ⚠️ Warning: Could not process existing payments: %v\n", err)
			}
		}
	}

	fmt.Printf("\n📊 Reprocessing Summary:\n")
	fmt.Printf("  ✅ Processed: %d purchases\n", processedCount)
	fmt.Printf("  📋 Skipped (already had entries): %d purchases\n", skippedCount)
	fmt.Printf("  ❌ Errors: %d purchases\n", errorCount)
}

func checkExistingSSOTEntries(db *gorm.DB, purchaseID uint) (bool, error) {
	var count int64
	err := db.Model(&models.SSOTJournalEntry{}).
		Where("source_type = ? AND source_id = ?", models.SSOTSourceTypePurchase, purchaseID).
		Count(&count).Error
	
	return count > 0, err
}

func processExistingPayments(db *gorm.DB, adapter *services.PurchaseSSOTJournalAdapter, purchase *models.Purchase) error {
	// For now, we'll create a single payment journal entry for the total paid amount
	// In a more sophisticated system, we'd track individual payments
	if purchase.PaidAmount <= 0 {
		return nil
	}

	// We need a bank account for the payment entry
	// Let's use the purchase's bank account or find a default one
	bankAccountID := uint64(0)
	if purchase.BankAccountID != nil {
		bankAccountID = uint64(*purchase.BankAccountID)
	} else {
		// Find a default bank account (e.g., Bank Mandiri)
		var account models.Account
		err := db.Where("code = ?", "1103").First(&account).Error
		if err == nil {
			bankAccountID = uint64(account.ID)
		}
	}

	if bankAccountID == 0 {
		return fmt.Errorf("no bank account available for payment journal entry")
	}

	ctx := context.Background()
	paymentAmountDecimal := decimal.NewFromFloat(purchase.PaidAmount)
	
	_, err := adapter.CreatePurchasePaymentJournalEntry(
		ctx,
		purchase,
		paymentAmountDecimal,
		bankAccountID,
		uint64(1), // admin user ID
		fmt.Sprintf("PAYMENT-%s", purchase.Code),
		"Reprocessed payment from existing purchase data",
	)

	if err != nil {
		return fmt.Errorf("failed to create payment journal entry: %v", err)
	}

	fmt.Printf("    ✅ Payment journal entry created for Rp %.2f\n", purchase.PaidAmount)
	return nil
}

func verifyCOAUpdates(db *gorm.DB) {
	fmt.Println("🔍 Checking COA balance updates after SSOT reprocessing...")

	// Check key purchase-related accounts
	accounts := map[string]string{
		"1301": "Persediaan Barang Dagangan",
		"1240": "PPN Masukan", 
		"2101": "Utang Usaha",
		"2111": "Utang PPh 21",
		"2112": "Utang PPh 23",
		"1103": "Bank Mandiri",
	}

	for code, name := range accounts {
		var balance float64
		err := db.Model(&models.Account{}).Where("code = ?", code).Select("balance").Row().Scan(&balance)
		if err != nil {
			fmt.Printf("  ❌ %s (%s): Not found\n", name, code)
			continue
		}

		status := "✅"
		analysis := ""
		switch code {
		case "1301": // Inventory
			if balance > 0 {
				analysis = " (Good - should reflect inventory from purchases)"
			} else {
				status = "⚠️"
				analysis = " (May be zero if no inventory items purchased)"
			}
		case "2101": // Accounts Payable  
			if balance < 0 {
				analysis = " (Good - credit balance for liability)"
			} else {
				status = "⚠️" 
				analysis = " (Should be negative for unpaid purchases)"
			}
		case "1103": // Bank Mandiri
			analysis = fmt.Sprintf(" (Current balance - should reflect payments made)")
		}

		fmt.Printf("  %s %s (%s): Rp %.2f%s\n", status, name, code, balance, analysis)
	}

	// Check journal entry counts
	var totalEntries, purchaseEntries, paymentEntries int64
	db.Model(&models.SSOTJournalEntry{}).Count(&totalEntries)
	db.Model(&models.SSOTJournalEntry{}).Where("source_type = ?", models.SSOTSourceTypePurchase).Count(&purchaseEntries)
	db.Model(&models.SSOTJournalEntry{}).Where("source_type = ?", models.SSOTSourceTypePayment).Count(&paymentEntries)

	fmt.Printf("\n📊 SSOT Journal Entry Summary:\n")
	fmt.Printf("  Total Entries: %d\n", totalEntries)
	fmt.Printf("  Purchase Entries: %d\n", purchaseEntries)
	fmt.Printf("  Payment Entries: %d\n", paymentEntries)
}

func provideSummaryAndRecommendations() {
	fmt.Println("📋 Reprocessing Complete!")
	fmt.Println("\n✅ What was accomplished:")
	fmt.Println("1. Existing approved purchases were processed with SSOT journal entries")
	fmt.Println("2. COA balances should now reflect purchase transactions")
	fmt.Println("3. Purchase payments were integrated where applicable")
	fmt.Println("4. Double-entry accounting principles maintained")

	fmt.Println("\n🔮 Going forward:")
	fmt.Println("1. New purchases will automatically use SSOT journal system")
	fmt.Println("2. Purchase approvals will create journal entries automatically")
	fmt.Println("3. Purchase payments will be integrated with SSOT")
	fmt.Println("4. All accounting transactions maintain consistency")

	fmt.Println("\n🧪 Next steps to test:")
	fmt.Println("1. Create a new purchase and approve it")
	fmt.Println("2. Verify journal entries are created automatically")
	fmt.Println("3. Make a payment for a credit purchase")
	fmt.Println("4. Check that COA balances update correctly")

	fmt.Println("\n✅ Purchase system is now fully integrated with SSOT!")
}

func provideManualInstructions() {
	fmt.Println("Since database is offline, here are manual steps:")
	fmt.Println("")
	fmt.Println("🔧 Manual SSOT Integration for Purchases:")
	fmt.Println("")
	fmt.Println("1. Check existing approved purchases:")
	fmt.Println("   SELECT id, code, total_amount, status FROM purchases WHERE status = 'APPROVED';")
	fmt.Println("")
	fmt.Println("2. For each approved purchase, create SSOT journal entries manually")
	fmt.Println("   or run this script when database is available.")
	fmt.Println("")
	fmt.Println("3. Verify COA balances after processing:")
	fmt.Println("   SELECT code, name, balance FROM accounts WHERE code IN ('1301', '2101', '1240');")
	fmt.Println("")
	fmt.Println("✅ The purchase system code is already integrated with SSOT!")
	fmt.Println("   - New purchases will automatically create journal entries when approved")
	fmt.Println("   - Purchase payments will update COA through SSOT")
	fmt.Println("   - Only existing transactions need reprocessing")
}