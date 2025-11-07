package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 PostgreSQL SSOT Fix & Test")
	fmt.Println("=============================")
	fmt.Println()

	db := database.ConnectDB()
	
	// Step 1: Check what tables actually exist
	fmt.Println("📊 Step 1: Checking Existing Tables...")
	checkPostgreSQLTables(db)
	fmt.Println()

	// Step 2: Analyze journal_entries table structure
	fmt.Println("🏗️  Step 2: Analyzing journal_entries Structure...")
	analyzeJournalEntriesStructure(db)
	fmt.Println()

	// Step 3: Create proper account_balances view
	fmt.Println("🔧 Step 3: Creating Account Balances View...")
	createPostgreSQLAccountBalancesView(db)
	fmt.Println()

	// Step 4: Test the final result
	fmt.Println("🧪 Step 4: Final Testing...")
	testFinalResult(db)
}

func checkPostgreSQLTables(db *gorm.DB) {
	// PostgreSQL compatible query to list tables
	var tables []string
	db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name LIKE '%journal%'
		ORDER BY table_name
	`).Pluck("table_name", &tables)

	fmt.Println("📋 Journal-related tables found:")
	for _, table := range tables {
		var count int64
		db.Table(table).Count(&count)
		fmt.Printf("   ✅ %s: %d records\n", table, count)
	}

	// Also check accounts table
	var accountCount int64
	db.Table("accounts").Count(&accountCount)
	fmt.Printf("   ✅ accounts: %d records\n", accountCount)

	// Check if account_balances exists
	var viewExists bool
	var viewCount int
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.views 
		WHERE table_schema = 'public' 
		AND table_name = 'account_balances'
	`).Row().Scan(&viewCount)
	viewExists = (err == nil && viewCount > 0)
	
	if viewExists {
		var recordCount int64
		db.Table("account_balances").Count(&recordCount)
		fmt.Printf("   ✅ account_balances (view): %d records\n", recordCount)
	} else {
		fmt.Printf("   ❌ account_balances: Does not exist\n")
	}
}

func analyzeJournalEntriesStructure(db *gorm.DB) {
	// Get column information for journal_entries
	type ColumnInfo struct {
		ColumnName string `json:"column_name"`
		DataType   string `json:"data_type"`
		IsNullable string `json:"is_nullable"`
	}

	var columns []ColumnInfo
	db.Raw(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'journal_entries' 
		ORDER BY ordinal_position
	`).Find(&columns)

	fmt.Println("📋 journal_entries table columns:")
	for _, col := range columns {
		nullable := "NOT NULL"
		if col.IsNullable == "YES" {
			nullable = "NULL"
		}
		fmt.Printf("   - %s: %s (%s)\n", col.ColumnName, col.DataType, nullable)
	}

	// Get sample data to understand the structure
	var sampleEntries []map[string]interface{}
	db.Raw("SELECT * FROM journal_entries LIMIT 3").Find(&sampleEntries)
	
	if len(sampleEntries) > 0 {
		fmt.Println("\n📝 Sample journal entries data:")
		for i, entry := range sampleEntries {
			fmt.Printf("   Entry %d:\n", i+1)
			for key, value := range entry {
				fmt.Printf("      %s: %v\n", key, value)
			}
			fmt.Println()
		}
	}
}

func createPostgreSQLAccountBalancesView(db *gorm.DB) {
	fmt.Println("🔧 Creating account_balances view for PostgreSQL...")

	// Drop existing view if it exists
	db.Exec("DROP VIEW IF EXISTS account_balances")

	// First, let's check what account-related fields exist in journal_entries
	var hasDebitAccountID bool
	var hasCreditAccountID bool
	
	err1 := db.Raw("SELECT debit_account_id FROM journal_entries LIMIT 1").Row().Scan(new(interface{}))
	hasDebitAccountID = (err1 == nil)
	
	err2 := db.Raw("SELECT credit_account_id FROM journal_entries LIMIT 1").Row().Scan(new(interface{}))
	hasCreditAccountID = (err2 == nil)

	fmt.Printf("🔍 Journal entries structure check:\n")
	fmt.Printf("   - Has debit_account_id: %t\n", hasDebitAccountID)
	fmt.Printf("   - Has credit_account_id: %t\n", hasCreditAccountID)

	var createViewQuery string

	if hasDebitAccountID && hasCreditAccountID {
		// Create view with proper account linking
		createViewQuery = `
		CREATE VIEW account_balances AS
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.account_type,
			
			-- Calculate total debits for this account
			COALESCE(SUM(CASE WHEN je.debit_account_id = a.id THEN je.total_debit ELSE 0 END), 0) as total_debit,
			
			-- Calculate total credits for this account
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
		WHERE a.is_active = true
		GROUP BY a.id, a.code, a.name, a.account_type
		`
		fmt.Println("✅ Creating comprehensive account_balances view...")
	} else {
		// Create basic view without account linking
		createViewQuery = `
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
		WHERE a.is_active = true
		`
		fmt.Println("⚠️  Creating basic account_balances view (no journal linkage)...")
	}

	err := db.Exec(createViewQuery)
	if err.Error != nil {
		fmt.Printf("❌ Error creating view: %v\n", err.Error)
	} else {
		fmt.Println("✅ Account balances view created successfully!")
	}
}

func testFinalResult(db *gorm.DB) {
	// Test the account_balances view
	var count int64
	err := db.Table("account_balances").Count(&count)
	if err != nil {
		fmt.Printf("❌ Error testing account_balances view: %v\n", err)
		return
	}

	fmt.Printf("✅ Account balances view contains %d records\n", count)

	// Get sample balances
	var balances []struct {
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		AccountType string  `gorm:"column:account_type"`
		TotalDebit  float64 `gorm:"column:total_debit"`
		TotalCredit float64 `gorm:"column:total_credit"`
		Balance     float64 `gorm:"column:balance"`
	}
	
	db.Table("account_balances").
		Select("account_code, account_name, account_type, total_debit, total_credit, balance").
		Order("balance DESC").
		Limit(10).
		Find(&balances)

	if len(balances) > 0 {
		fmt.Println("\n💰 Top Account Balances:")
		for _, bal := range balances {
			fmt.Printf("   %s - %s (%s)\n", bal.AccountCode, bal.AccountName, bal.AccountType)
			fmt.Printf("      Debit: %.2f, Credit: %.2f, Balance: %.2f\n", 
				bal.TotalDebit, bal.TotalCredit, bal.Balance)
		}
	}

	// Also show journal entries summary
	var journalCount int64
	var totalDebit float64
	db.Table("journal_entries").Count(&journalCount)
	db.Table("journal_entries").Select("COALESCE(SUM(total_debit), 0)").Row().Scan(&totalDebit)

	fmt.Printf("\n📊 SSOT Summary:\n")
	fmt.Printf("   - Journal Entries: %d\n", journalCount)
	fmt.Printf("   - Total Transaction Value: %.2f\n", totalDebit)
	fmt.Printf("   - Account Balances View: %d accounts\n", count)

	if journalCount > 0 && count > 0 {
		fmt.Println("\n🎉 SSOT System Status: READY!")
		fmt.Println("✅ Journal entries exist")
		fmt.Println("✅ Account balances view created")
		fmt.Println("✅ Frontend should now be able to retrieve financial data")
	} else {
		fmt.Println("\n⚠️  SSOT System Status: INCOMPLETE")
		if journalCount == 0 {
			fmt.Println("❌ No journal entries found")
		}
		if count == 0 {
			fmt.Println("❌ Account balances view is empty")
		}
	}
}