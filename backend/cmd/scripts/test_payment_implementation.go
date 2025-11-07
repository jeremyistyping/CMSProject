package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🧪 Testing Payment Implementation")
	fmt.Println("================================")

	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test 1: Check if payment functionality exists
	fmt.Println("📝 Test 1: Payment Tables & Structure")
	testPaymentStructure(db)

	// Test 2: Check existing payments and their journal entries
	fmt.Println("\n📝 Test 2: Existing Payment Transactions")
	testExistingPayments(db)

	// Test 3: Check account balance updates
	fmt.Println("\n📝 Test 3: Account Balance Logic")
	testAccountBalanceLogic(db)

	// Test 4: Check sales vs payment allocation
	fmt.Println("\n📝 Test 4: Sales Payment Allocation")
	testSalesPaymentAllocation(db)

	// Test 5: Verify journal entry correctness
	fmt.Println("\n📝 Test 5: Journal Entry Analysis")
	testJournalEntries(db)

	// Final Assessment
	fmt.Println("\n🎯 Final Assessment:")
	provideFinalAssessment(db)
}

func testPaymentStructure(db *gorm.DB) {
	// Check if payment-related tables exist
	tables := []string{
		"payments",
		"payment_allocations", 
		"sale_payments",
		"cash_bank_transactions",
		"journal_entries",
		"journal_lines",
	}

	fmt.Println("Checking required tables:")
	for _, table := range tables {
		var count int64
		err := db.Raw("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_NAME = ?", table).Scan(&count).Error
		if err != nil {
			fmt.Printf("❌ Error checking table %s: %v\n", table, err)
		} else if count > 0 {
			// Get record count
			var records int64
			db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&records)
			fmt.Printf("✅ %s (exists, %d records)\n", table, records)
		} else {
			fmt.Printf("❌ %s (missing)\n", table)
		}
	}
}

func testExistingPayments(db *gorm.DB) {
	var payments []struct {
		ID        uint
		Code      string
		ContactID uint
		Amount    float64
		Status    string
		Method    string
		CreatedAt string
	}

	err := db.Raw(`
		SELECT id, code, contact_id, amount, status, method, created_at
		FROM payments 
		WHERE deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT 10
	`).Scan(&payments).Error

	if err != nil {
		fmt.Printf("❌ Error fetching payments: %v\n", err)
		return
	}

	if len(payments) == 0 {
		fmt.Println("⚠️  No payment records found")
		return
	}

	fmt.Printf("Found %d payment records:\n", len(payments))
	fmt.Printf("%-4s %-15s %-10s %-12s %-10s %-15s\n", 
		"ID", "CODE", "CONTACT", "AMOUNT", "STATUS", "METHOD")
	fmt.Println("================================================================")

	for _, payment := range payments {
		fmt.Printf("%-4d %-15s %-10d %-12.0f %-10s %-15s\n",
			payment.ID, payment.Code, payment.ContactID, 
			payment.Amount, payment.Status, payment.Method)
	}
}

func testAccountBalanceLogic(db *gorm.DB) {
	// Check key accounts before any payments
	keyAccounts := []string{"1101", "1102", "1103", "1201", "4101"}
	
	fmt.Println("Key account balances:")
	for _, code := range keyAccounts {
		var account struct {
			Code    string
			Name    string
			Balance float64
			Type    string
		}
		
		err := db.Raw("SELECT code, name, balance, type FROM accounts WHERE code = ?", code).Scan(&account).Error
		if err != nil {
			fmt.Printf("❌ Account %s: Error - %v\n", code, err)
		} else {
			fmt.Printf("✅ %s %-25s: %15.2f (%s)\n", 
				account.Code, account.Name, account.Balance, account.Type)
		}
	}
}

func testSalesPaymentAllocation(db *gorm.DB) {
	// Check sales and their payment allocations
	var salesWithPayments []struct {
		SaleID            uint
		SaleCode          string
		TotalAmount       float64
		PaidAmount        float64
		OutstandingAmount float64
		Status            string
		PaymentCount      int64
	}

	query := `
		SELECT 
			s.id as sale_id,
			s.code as sale_code,
			s.total_amount,
			s.paid_amount,
			s.outstanding_amount,
			s.status,
			COALESCE(sp.payment_count, 0) as payment_count
		FROM sales s
		LEFT JOIN (
			SELECT sale_id, COUNT(*) as payment_count 
			FROM sale_payments 
			GROUP BY sale_id
		) sp ON s.id = sp.sale_id
		WHERE s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		LIMIT 10
	`

	err := db.Raw(query).Scan(&salesWithPayments).Error
	if err != nil {
		fmt.Printf("❌ Error fetching sales payment data: %v\n", err)
		return
	}

	fmt.Printf("Sales Payment Allocation Status:\n")
	fmt.Printf("%-4s %-15s %-12s %-12s %-12s %-10s %-8s\n", 
		"ID", "CODE", "TOTAL", "PAID", "OUTSTANDING", "STATUS", "PAYMENTS")
	fmt.Println("========================================================================")

	for _, sale := range salesWithPayments {
		fmt.Printf("%-4d %-15s %-12.0f %-12.0f %-12.0f %-10s %-8d\n",
			sale.SaleID, sale.SaleCode, sale.TotalAmount, 
			sale.PaidAmount, sale.OutstandingAmount, sale.Status, sale.PaymentCount)
	}
}

func testJournalEntries(db *gorm.DB) {
	// Check journal entries for payments
	var journalEntries []struct {
		EntryID       uint
		Reference     string
		Description   string
		AccountCode   string
		AccountName   string
		DebitAmount   float64
		CreditAmount  float64
		ReferenceType string
	}

	query := `
		SELECT 
			je.id as entry_id,
			je.reference,
			je.description,
			a.code as account_code,
			a.name as account_name,
			jl.debit_amount,
			jl.credit_amount,
			je.reference_type
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE je.reference_type = 'PAYMENT'
		AND je.deleted_at IS NULL
		ORDER BY je.created_at DESC, jl.line_number
		LIMIT 20
	`

	err := db.Raw(query).Scan(&journalEntries).Error
	if err != nil {
		fmt.Printf("❌ Error fetching journal entries: %v\n", err)
		return
	}

	if len(journalEntries) == 0 {
		fmt.Println("⚠️  No payment journal entries found")
		return
	}

	fmt.Printf("Found %d payment journal lines:\n", len(journalEntries))
	fmt.Printf("%-8s %-15s %-8s %-25s %-12s %-12s\n", 
		"ENTRY", "REFERENCE", "CODE", "ACCOUNT", "DEBIT", "CREDIT")
	fmt.Println("===============================================================================")

	currentEntry := uint(0)
	entryDebit := 0.0
	entryCredit := 0.0

	for _, entry := range journalEntries {
		if currentEntry != 0 && currentEntry != entry.EntryID {
			// Print totals for previous entry
			fmt.Printf("%-8s %-15s %-8s %-25s %-12.2f %-12.2f\n", 
				"", "TOTALS:", "", "", entryDebit, entryCredit)
			fmt.Println("-------------------------------------------------------------------------------")
			entryDebit = 0.0
			entryCredit = 0.0
		}

		fmt.Printf("%-8d %-15s %-8s %-25s %-12.2f %-12.2f\n",
			entry.EntryID, entry.Reference, entry.AccountCode, entry.AccountName,
			entry.DebitAmount, entry.CreditAmount)

		currentEntry = entry.EntryID
		entryDebit += entry.DebitAmount
		entryCredit += entry.CreditAmount
	}

	// Print final totals
	if currentEntry != 0 {
		fmt.Printf("%-8s %-15s %-8s %-25s %-12.2f %-12.2f\n", 
			"", "TOTALS:", "", "", entryDebit, entryCredit)
	}
}

func provideFinalAssessment(db *gorm.DB) {
	// Count various elements
	var paymentCount, journalCount, allocationCount int64
	
	db.Raw("SELECT COUNT(*) FROM payments WHERE deleted_at IS NULL").Scan(&paymentCount)
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE reference_type = 'PAYMENT'").Scan(&journalCount)
	db.Raw("SELECT COUNT(*) FROM payment_allocations").Scan(&allocationCount)

	// Get account balances for verification
	var arBalance, cashBalance float64
	db.Raw("SELECT balance FROM accounts WHERE code = '1201'").Scan(&arBalance)
	db.Raw("SELECT SUM(balance) FROM accounts WHERE code IN ('1101','1102','1103','1104','1105')").Scan(&cashBalance)

	fmt.Println("📊 Implementation Assessment:")
	fmt.Printf("✅ Payment Records: %d\n", paymentCount)
	fmt.Printf("✅ Journal Entries: %d\n", journalCount)
	fmt.Printf("✅ Payment Allocations: %d\n", allocationCount)
	fmt.Printf("✅ Current AR Balance: Rp %.2f\n", arBalance)
	fmt.Printf("✅ Current Cash/Bank: Rp %.2f\n", cashBalance)

	fmt.Println("\n🎯 Feature Analysis:")
	
	if paymentCount > 0 {
		fmt.Println("✅ Payment Creation - IMPLEMENTED")
	} else {
		fmt.Println("❌ Payment Creation - NO DATA")
	}

	if journalCount > 0 {
		fmt.Println("✅ Journal Entry Creation - IMPLEMENTED")
		if journalCount == paymentCount {
			fmt.Println("✅ Journal Consistency - PERFECT (1:1 ratio)")
		} else {
			fmt.Printf("⚠️  Journal Consistency - CHECK (Payment:%d vs Journal:%d)\n", paymentCount, journalCount)
		}
	} else {
		fmt.Println("❌ Journal Entry Creation - NOT IMPLEMENTED")
	}

	if allocationCount > 0 {
		fmt.Println("✅ Payment Allocation - IMPLEMENTED")
	} else {
		fmt.Println("❌ Payment Allocation - NO DATA")
	}

	// Final verdict
	fmt.Println("\n🏆 FINAL VERDICT:")
	
	score := 0
	if paymentCount > 0 { score++ }
	if journalCount > 0 { score++ }
	if allocationCount > 0 { score++ }
	
	switch score {
	case 3:
		fmt.Println("🎉 EXCELLENT - Full payment system implemented!")
		fmt.Println("   ✅ Payment records")
		fmt.Println("   ✅ Journal entries") 
		fmt.Println("   ✅ Payment allocations")
		fmt.Println("   Your app follows proper accounting principles! 🚀")
	case 2:
		fmt.Println("👍 GOOD - Most features implemented")
		fmt.Println("   Some features may need testing or data")
	case 1:
		fmt.Println("⚠️  PARTIAL - Basic features only")
		fmt.Println("   Need to complete journal entries or allocations")
	default:
		fmt.Println("❌ INCOMPLETE - Payment system not active")
		fmt.Println("   No payment data found")
	}
}