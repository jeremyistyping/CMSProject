package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔍 Database Structure Analysis")
	fmt.Println("=============================")
	fmt.Println()

	db := database.ConnectDB()
	
	// Check what journal-related tables exist
	fmt.Println("📊 Checking Journal-Related Tables...")
	checkTableExists(db, "journal_entries")
	checkTableExists(db, "journal_details") 
	checkTableExists(db, "journal_entry_items")
	checkTableExists(db, "journal_entry_accounts")
	checkTableExists(db, "journal_items")
	fmt.Println()

	// Show actual journal entries structure
	fmt.Println("🏗️  Analyzing journal_entries table structure...")
	showTableStructure(db, "journal_entries")
	fmt.Println()

	// Check accounts table
	fmt.Println("🏗️  Analyzing accounts table structure...")
	showTableStructure(db, "accounts")
	fmt.Println()

	// Check if there are any related detail tables
	fmt.Println("🔍 Looking for transaction detail tables...")
	checkDetailTables(db)
	fmt.Println()

	// Create proper account_balances view based on actual structure
	fmt.Println("🔧 Creating Account Balances View...")
	createProperAccountBalancesView(db)
}

func checkTableExists(db *gorm.DB, tableName string) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Count(&count)
	
	if err != nil || count == 0 {
		fmt.Printf("❌ Table '%s': Does not exist\n", tableName)
	} else {
		fmt.Printf("✅ Table '%s': Exists\n", tableName)
		
		// Count records
		var recordCount int64
		db.Table(tableName).Count(&recordCount)
		fmt.Printf("   Records: %d\n", recordCount)
	}
}

func showTableStructure(db *gorm.DB, tableName string) {
	type ColumnInfo struct {
		ColumnName    string `json:"column_name"`
		DataType      string `json:"data_type"`
		IsNullable    string `json:"is_nullable"`
		ColumnDefault string `json:"column_default"`
		Extra         string `json:"extra"`
	}

	var columns []ColumnInfo
	err := db.Raw(`
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			COALESCE(column_default, 'NULL') as column_default,
			COALESCE(extra, '') as extra
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = ? 
		ORDER BY ordinal_position
	`, tableName).Find(&columns)

	if err != nil {
		fmt.Printf("❌ Error analyzing table '%s': %v\n", tableName, err)
		return
	}

	if len(columns) == 0 {
		fmt.Printf("❌ No columns found for table '%s'\n", tableName)
		return
	}

	fmt.Printf("📋 Table '%s' columns:\n", tableName)
	for _, col := range columns {
		nullable := "NOT NULL"
		if col.IsNullable == "YES" {
			nullable = "NULL"
		}
		fmt.Printf("   - %s: %s %s %s %s\n", 
			col.ColumnName, col.DataType, nullable, col.ColumnDefault, col.Extra)
	}
}

func checkDetailTables(db *gorm.DB) {
	// Look for tables that might contain journal detail records
	possibleTables := []string{
		"sale_items",
		"purchase_items", 
		"payment_items",
		"transaction_details",
		"ledger_entries",
		"account_transactions",
	}

	for _, table := range possibleTables {
		checkTableExists(db, table)
	}
}

func createProperAccountBalancesView(db *gorm.DB) {
	// First, let's see what data is actually in journal_entries
	var sampleEntry struct {
		ID          uint    `json:"id"`
		EntryDate   string  `json:"entry_date"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
	
	err := db.Table("journal_entries").Select("id, DATE(entry_date) as entry_date, reference, description, total_debit, total_credit").First(&sampleEntry)
	if err != nil {
		fmt.Printf("❌ Cannot read journal_entries: %v\n", err)
		return
	}

	fmt.Printf("📝 Sample journal entry: ID=%d, TotalDebit=%.2f, TotalCredit=%.2f\n", 
		sampleEntry.ID, sampleEntry.TotalDebit, sampleEntry.TotalCredit)

	// Check if there's a way to link journal entries to specific accounts
	// Let's look for foreign key relationships
	fmt.Println("🔗 Checking for account relationships...")
	
	// Try to find account-related fields in journal_entries
	var journalWithAccount struct {
		ID        uint `json:"id"`
		AccountID uint `json:"account_id"`
		DebitAccountID uint `json:"debit_account_id"`
		CreditAccountID uint `json:"credit_account_id"`
	}
	
	err = db.Table("journal_entries").Select("id, account_id, debit_account_id, credit_account_id").First(&journalWithAccount)
	if err == nil {
		fmt.Printf("✅ Found account relationships in journal_entries\n")
		fmt.Printf("   AccountID: %d, DebitAccountID: %d, CreditAccountID: %d\n", 
			journalWithAccount.AccountID, journalWithAccount.DebitAccountID, journalWithAccount.CreditAccountID)
		
		// Create account_balances view based on actual structure
		createAccountBalancesFromJournalEntries(db)
	} else {
		fmt.Printf("⚠️  No direct account relationships found in journal_entries\n")
		fmt.Println("💡 Creating basic account balances view from chart of accounts...")
		createBasicAccountBalancesView(db)
	}
}

func createAccountBalancesFromJournalEntries(db *gorm.DB) {
	fmt.Println("🔧 Creating account_balances view from journal_entries...")

	// Drop existing view
	db.Exec("DROP VIEW IF EXISTS account_balances")

	// Create comprehensive view
	createViewQuery := `
	CREATE VIEW account_balances AS
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.account_type,
		
		-- Calculate debits (when this account is debited)
		COALESCE(SUM(CASE WHEN je.debit_account_id = a.id THEN je.total_debit ELSE 0 END), 0) as total_debit,
		
		-- Calculate credits (when this account is credited) 
		COALESCE(SUM(CASE WHEN je.credit_account_id = a.id THEN je.total_credit ELSE 0 END), 0) as total_credit,
		
		-- Calculate balance based on account type
		CASE 
			WHEN a.account_type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(SUM(CASE WHEN je.debit_account_id = a.id THEN je.total_debit ELSE 0 END), 0) - 
				COALESCE(SUM(CASE WHEN je.credit_account_id = a.id THEN je.total_credit ELSE 0 END), 0)
			ELSE 
				COALESCE(SUM(CASE WHEN je.credit_account_id = a.id THEN je.total_credit ELSE 0 END), 0) - 
				COALESCE(SUM(CASE WHEN je.debit_account_id = a.id THEN je.total_debit ELSE 0 END), 0)
		END as balance,
		
		NOW() as last_updated
		
	FROM accounts a
	LEFT JOIN journal_entries je ON (a.id = je.debit_account_id OR a.id = je.credit_account_id)
	WHERE a.is_active = 1
	GROUP BY a.id, a.code, a.name, a.account_type
	`

	err := db.Exec(createViewQuery)
	if err != nil {
		fmt.Printf("❌ Error creating account_balances view: %v\n", err.Error)
	} else {
		fmt.Println("✅ Account balances view created successfully!")
		testAccountBalancesView(db)
	}
}

func createBasicAccountBalancesView(db *gorm.DB) {
	fmt.Println("🔧 Creating basic account_balances view...")

	// Drop existing view
	db.Exec("DROP VIEW IF EXISTS account_balances")

	// Create basic view with zero balances for now
	createViewQuery := `
	CREATE VIEW account_balances AS
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.account_type,
		0.00 as total_debit,
		0.00 as total_credit,
		0.00 as balance,
		NOW() as last_updated
	FROM accounts a
	WHERE a.is_active = 1
	`

	err := db.Exec(createViewQuery)
	if err != nil {
		fmt.Printf("❌ Error creating basic account_balances view: %v\n", err.Error)
	} else {
		fmt.Println("✅ Basic account balances view created successfully!")
		testAccountBalancesView(db)
	}
}

func testAccountBalancesView(db *gorm.DB) {
	fmt.Println("🧪 Testing account_balances view...")

	var count int64
	err := db.Table("account_balances").Count(&count)
	if err != nil {
		fmt.Printf("❌ Error counting account_balances: %v\n", err)
		return
	}

	fmt.Printf("✅ Account balances view contains %d records\n", count)

	// Show sample records
	var balances []struct {
		AccountCode string  `json:"account_code"`
		AccountName string  `json:"account_name"`
		AccountType string  `json:"account_type"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
		Balance     float64 `json:"balance"`
	}
	
	db.Table("account_balances").
		Select("account_code, account_name, account_type, total_debit, total_credit, balance").
		Limit(10).Find(&balances)

	if len(balances) > 0 {
		fmt.Println("💰 Sample Account Balances:")
		for _, bal := range balances {
			fmt.Printf("   %s - %s (%s): Debit=%.2f, Credit=%.2f, Balance=%.2f\n", 
				bal.AccountCode, bal.AccountName, bal.AccountType, 
				bal.TotalDebit, bal.TotalCredit, bal.Balance)
		}
	}
}