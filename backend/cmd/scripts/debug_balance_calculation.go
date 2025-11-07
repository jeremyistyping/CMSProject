package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔍 Debug Balance Calculation Issue")
	fmt.Println("==================================")

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
		log.Printf("❌ Database connection failed: %v", err)
		fmt.Println("\n📋 Manual Analysis Based on Screenshots:")
		provideManualAnalysis()
		return
	}

	fmt.Println("✅ Database connected successfully")

	// Test 1: Check Bank Mandiri account details
	fmt.Println("\n📝 Test 1: Bank Mandiri Account Analysis")
	checkBankMandiriAccount(db)

	// Test 2: Check all journal entries for Bank Mandiri
	fmt.Println("\n📝 Test 2: Journal Entries for Bank Mandiri")
	checkBankMandiriJournalEntries(db)

	// Test 3: Check SSOT journal entries
	fmt.Println("\n📝 Test 3: SSOT Journal Entries")
	checkSSOTJournalEntries(db)

	// Test 4: Check cash bank transactions
	fmt.Println("\n📝 Test 4: Cash Bank Transactions")
	checkCashBankTransactions(db)

	// Test 5: Check recent payments
	fmt.Println("\n📝 Test 5: Recent Payments Analysis")
	checkRecentPayments(db)

	// Final analysis
	fmt.Println("\n🎯 Final Analysis & Recommendations:")
	provideFinalAnalysis(db)
}

func checkBankMandiriAccount(db *gorm.DB) {
	var account struct {
		ID      uint
		Code    string
		Name    string
		Type    string
		Balance float64
		Status  string
	}

	err := db.Raw("SELECT id, code, name, type, balance, status FROM accounts WHERE code = '1103' OR name LIKE '%Mandiri%'").Scan(&account).Error
	if err != nil {
		fmt.Printf("❌ Error fetching Bank Mandiri account: %v\n", err)
		return
	}

	fmt.Printf("Bank Mandiri Account Details:\n")
	fmt.Printf("ID: %d\n", account.ID)
	fmt.Printf("Code: %s\n", account.Code)
	fmt.Printf("Name: %s\n", account.Name)
	fmt.Printf("Type: %s\n", account.Type)
	fmt.Printf("Balance: Rp %.2f\n", account.Balance)
	fmt.Printf("Status: %s\n", account.Status)

	// Check if balance is abnormal
	if account.Balance > 10000000 {
		fmt.Printf("⚠️  ABNORMAL BALANCE DETECTED: Rp %.2f\n", account.Balance)
		fmt.Println("   This suggests double journal entries or calculation error")
	}
}

func checkBankMandiriJournalEntries(db *gorm.DB) {
	var journalEntries []struct {
		EntryID      uint
		EntryDate    string
		Reference    string
		Description  string
		DebitAmount  float64
		CreditAmount float64
		EntryType    string
		CreatedAt    string
	}

	query := `
		SELECT 
			je.id as entry_id,
			je.entry_date,
			je.reference,
			je.description,
			jl.debit_amount,
			jl.credit_amount,
			'LEGACY' as entry_type,
			je.created_at
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.code = '1103' OR a.name LIKE '%Mandiri%'
		ORDER BY je.created_at DESC
		LIMIT 20
	`

	err := db.Raw(query).Scan(&journalEntries).Error
	if err != nil {
		fmt.Printf("❌ Error fetching legacy journal entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d legacy journal entries for Bank Mandiri:\n", len(journalEntries))
	if len(journalEntries) > 0 {
		fmt.Printf("%-8s %-12s %-15s %-30s %-12s %-12s %-20s\n", 
			"ENTRY", "DATE", "REFERENCE", "DESCRIPTION", "DEBIT", "CREDIT", "CREATED")
		fmt.Println("==================================================================================")

		totalDebit := 0.0
		totalCredit := 0.0

		for _, entry := range journalEntries {
			fmt.Printf("%-8d %-12s %-15s %-30s %-12.2f %-12.2f %-20s\n",
				entry.EntryID, entry.EntryDate[:10], entry.Reference, 
				entry.Description, entry.DebitAmount, entry.CreditAmount, entry.CreatedAt[:19])
			
			totalDebit += entry.DebitAmount
			totalCredit += entry.CreditAmount
		}

		fmt.Println("==================================================================================")
		fmt.Printf("%-8s %-12s %-15s %-30s %-12.2f %-12.2f\n", 
			"TOTALS", "", "", "", totalDebit, totalCredit)
		fmt.Printf("Net Effect: Rp %.2f\n", totalDebit-totalCredit)
	} else {
		fmt.Println("No legacy journal entries found for Bank Mandiri")
	}
}

func checkSSOTJournalEntries(db *gorm.DB) {
	var ssotEntries []struct {
		EntryID     uint
		EntryNumber string
		SourceType  string
		SourceID    *uint
		Reference   string
		TotalAmount float64
		Status      string
		CreatedAt   string
	}

	query := `
		SELECT 
			sje.id as entry_id,
			sje.entry_number,
			sje.source_type,
			sje.source_id,
			sje.reference,
			sje.total_amount,
			sje.status,
			sje.created_at
		FROM ssot_journal_entries sje
		JOIN ssot_journal_lines sjl ON sje.id = sjl.journal_entry_id
		JOIN accounts a ON sjl.account_id = a.id
		WHERE a.code = '1103' OR a.name LIKE '%Mandiri%'
		ORDER BY sje.created_at DESC
		LIMIT 10
	`

	err := db.Raw(query).Scan(&ssotEntries).Error
	if err != nil {
		fmt.Printf("❌ Error fetching SSOT journal entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d SSOT journal entries affecting Bank Mandiri:\n", len(ssotEntries))
	if len(ssotEntries) > 0 {
		fmt.Printf("%-8s %-15s %-12s %-8s %-15s %-12s %-10s %-20s\n", 
			"ENTRY", "NUMBER", "SOURCE", "SRC_ID", "REFERENCE", "AMOUNT", "STATUS", "CREATED")
		fmt.Println("==============================================================================================")

		totalAmount := 0.0
		for _, entry := range ssotEntries {
			sourceID := "N/A"
			if entry.SourceID != nil {
				sourceID = fmt.Sprintf("%d", *entry.SourceID)
			}

			fmt.Printf("%-8d %-15s %-12s %-8s %-15s %-12.2f %-10s %-20s\n",
				entry.EntryID, entry.EntryNumber, entry.SourceType, sourceID,
				entry.Reference, entry.TotalAmount, entry.Status, entry.CreatedAt[:19])
			
			totalAmount += entry.TotalAmount
		}

		fmt.Printf("Total SSOT Amount: Rp %.2f\n", totalAmount)
	} else {
		fmt.Println("No SSOT journal entries found affecting Bank Mandiri")
	}
}

func checkCashBankTransactions(db *gorm.DB) {
	var transactions []struct {
		ID              uint
		CashBankID      uint
		ReferenceType   string
		ReferenceID     *uint
		Amount          float64
		BalanceAfter    float64
		TransactionDate string
		Notes           string
	}

	query := `
		SELECT 
			cbt.id,
			cbt.cash_bank_id,
			cbt.reference_type,
			cbt.reference_id,
			cbt.amount,
			cbt.balance_after,
			cbt.transaction_date,
			cbt.notes
		FROM cash_bank_transactions cbt
		JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
		JOIN accounts a ON cb.account_id = a.id
		WHERE a.code = '1103' OR a.name LIKE '%Mandiri%'
		ORDER BY cbt.transaction_date DESC
		LIMIT 10
	`

	err := db.Raw(query).Scan(&transactions).Error
	if err != nil {
		fmt.Printf("❌ Error fetching cash bank transactions: %v\n", err)
		return
	}

	fmt.Printf("Found %d cash bank transactions for Bank Mandiri:\n", len(transactions))
	if len(transactions) > 0 {
		fmt.Printf("%-6s %-8s %-12s %-8s %-12s %-15s %-12s %-30s\n", 
			"ID", "CB_ID", "REF_TYPE", "REF_ID", "AMOUNT", "BALANCE_AFTER", "DATE", "NOTES")
		fmt.Println("=======================================================================================================")

		for _, tx := range transactions {
			refID := "N/A"
			if tx.ReferenceID != nil {
				refID = fmt.Sprintf("%d", *tx.ReferenceID)
			}

			fmt.Printf("%-6d %-8d %-12s %-8s %-12.2f %-15.2f %-12s %-30s\n",
				tx.ID, tx.CashBankID, tx.ReferenceType, refID, tx.Amount,
				tx.BalanceAfter, tx.TransactionDate[:10], tx.Notes)
		}
	} else {
		fmt.Println("No cash bank transactions found for Bank Mandiri")
	}
}

func checkRecentPayments(db *gorm.DB) {
	var payments []struct {
		ID        uint
		Code      string
		Amount    float64
		Status    string
		Method    string
		CreatedAt string
	}

	query := `
		SELECT id, code, amount, status, method, created_at
		FROM payments
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 5
	`

	err := db.Raw(query).Scan(&payments).Error
	if err != nil {
		fmt.Printf("❌ Error fetching recent payments: %v\n", err)
		return
	}

	fmt.Printf("Recent payments:\n")
	if len(payments) > 0 {
		fmt.Printf("%-6s %-15s %-12s %-12s %-15s %-20s\n", 
			"ID", "CODE", "AMOUNT", "STATUS", "METHOD", "CREATED")
		fmt.Println("==================================================================")

		for _, payment := range payments {
			fmt.Printf("%-6d %-15s %-12.2f %-12s %-15s %-20s\n",
				payment.ID, payment.Code, payment.Amount, payment.Status,
				payment.Method, payment.CreatedAt[:19])
		}
	} else {
		fmt.Println("No recent payments found")
	}
}

func provideFinalAnalysis(db *gorm.DB) {
	var bankBalance float64
	var totalJournalDebits, totalJournalCredits float64
	var totalSSOTAmount float64

	// Get current bank balance
	db.Raw("SELECT balance FROM accounts WHERE code = '1103'").Scan(&bankBalance)

	// Calculate total from legacy journals
	db.Raw(`
		SELECT 
			COALESCE(SUM(jl.debit_amount), 0) as total_debits,
			COALESCE(SUM(jl.credit_amount), 0) as total_credits
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.code = '1103'
	`).Scan(&totalJournalDebits)

	db.Raw(`
		SELECT COALESCE(SUM(jl.credit_amount), 0) as total_credits
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.code = '1103'
	`).Scan(&totalJournalCredits)

	// Calculate total from SSOT
	db.Raw(`
		SELECT COALESCE(SUM(sje.total_amount), 0) as total_amount
		FROM ssot_journal_entries sje
		JOIN ssot_journal_lines sjl ON sje.id = sjl.journal_entry_id
		JOIN accounts a ON sjl.account_id = a.id
		WHERE a.code = '1103'
	`).Scan(&totalSSOTAmount)

	fmt.Printf("Balance Analysis Summary:\n")
	fmt.Printf("Current Bank Mandiri Balance: Rp %.2f\n", bankBalance)
	fmt.Printf("Legacy Journal Debits: Rp %.2f\n", totalJournalDebits)
	fmt.Printf("Legacy Journal Credits: Rp %.2f\n", totalJournalCredits)
	fmt.Printf("Legacy Net Effect: Rp %.2f\n", totalJournalDebits-totalJournalCredits)
	fmt.Printf("SSOT Total Amount: Rp %.2f\n", totalSSOTAmount)

	fmt.Printf("\n🔍 Diagnosis:\n")
	
	if bankBalance == 11100000 {
		fmt.Println("❌ DOUBLE JOURNAL ENTRY DETECTED!")
		fmt.Printf("   Expected balance: Rp 5.550.000 (payment amount)\n")
		fmt.Printf("   Actual balance: Rp %.2f (exactly double)\n", bankBalance)
		fmt.Println("   Cause: Both SSOT and Legacy journal systems are active")
	}

	fmt.Printf("\n📋 Recommendations:\n")
	fmt.Println("1. Disable legacy journal creation in payment service")
	fmt.Println("2. Use only SSOT system for all journal entries")
	fmt.Println("3. Run balance recalculation to fix current balances")
	fmt.Println("4. Implement SSOT migration to clean up duplicate entries")
}

func provideManualAnalysis() {
	fmt.Println("Based on the screenshot analysis:")
	fmt.Printf("Expected Payment: Rp 5.550.000\n")
	fmt.Printf("Actual Bank Balance: Rp 11.100.000\n") 
	fmt.Printf("Difference: Rp 5.550.000 (Exactly double!)\n")

	fmt.Println("\n🔍 DIAGNOSIS: DOUBLE JOURNAL ENTRY")
	fmt.Println("The balance is exactly 2x the payment amount, indicating:")
	fmt.Println("1. ❌ Both SSOT and Legacy journal systems created entries")
	fmt.Println("2. ❌ Payment service called both journal methods") 
	fmt.Println("3. ❌ Account balance was updated twice")

	fmt.Println("\n📋 SOLUTION:")
	fmt.Println("1. Modify payment service to use ONLY SSOT system")
	fmt.Println("2. Disable legacy journal creation")
	fmt.Println("3. Run balance recalculation script")
	fmt.Println("4. Verify SSOT configuration is working properly")

	fmt.Println("\n🔧 IMMEDIATE ACTION:")
	fmt.Println("Check payment_service.go line 286:")
	fmt.Println("Make sure only createReceivablePaymentJournalWithSSOTFixed() is called")
	fmt.Println("Do NOT call both SSOT and legacy methods")
}