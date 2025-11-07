package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	config.LoadConfig()
	db := database.ConnectDB()
	
	fmt.Println("=== CHECKING MIGRATION 042 STATUS ===\n")
	
	// 1. Check if migration 042 was executed
	var migrationRecord struct {
		ID          uint
		Name        string
		ExecutedAt  string
		Status      string
	}
	
	err := db.Raw(`
		SELECT id, name, executed_at, status 
		FROM migration_records 
		WHERE name LIKE '%042%' OR name LIKE '%account_2102%'
		ORDER BY executed_at DESC
	`).Scan(&migrationRecord).Error
	
	if err == nil && migrationRecord.ID > 0 {
		fmt.Printf("✅ Migration 042 FOUND in migration_records:\n")
		fmt.Printf("   ID: %d\n", migrationRecord.ID)
		fmt.Printf("   Name: %s\n", migrationRecord.Name)
		fmt.Printf("   Status: %s\n", migrationRecord.Status)
		fmt.Printf("   Executed At: %s\n\n", migrationRecord.ExecutedAt)
	} else {
		fmt.Println("❌ Migration 042 NOT FOUND in migration_records\n")
		fmt.Println("   This means the SQL migration hasn't run yet!\n")
	}
	
	// 2. Check account 2102 current status
	var account2102 struct {
		ID          uint
		Code        string
		Name        string
		Type        string
		Balance     float64
		IsActive    bool
		DeletedAt   *string
	}
	
	err = db.Raw(`
		SELECT id, code, name, type, balance, is_active, deleted_at
		FROM accounts 
		WHERE code = '2102'
	`).Scan(&account2102).Error
	
	fmt.Println("=== ACCOUNT 2102 CURRENT STATUS ===\n")
	
	if err == nil && account2102.ID > 0 {
		fmt.Printf("Account 2102 EXISTS:\n")
		fmt.Printf("   ID: %d\n", account2102.ID)
		fmt.Printf("   Code: %s\n", account2102.Code)
		fmt.Printf("   Name: %s\n", account2102.Name)
		fmt.Printf("   Type: %s\n", account2102.Type)
		fmt.Printf("   Balance: Rp %.0f\n", account2102.Balance)
		fmt.Printf("   IsActive: %v\n", account2102.IsActive)
		if account2102.DeletedAt != nil {
			fmt.Printf("   DeletedAt: %s (SOFT DELETED)\n", *account2102.DeletedAt)
		} else {
			fmt.Printf("   DeletedAt: NULL (NOT DELETED)\n")
		}
		
		if account2102.Balance != 0 && account2102.DeletedAt == nil {
			fmt.Println("\n❌ PROBLEM: Account 2102 still has balance and NOT deleted!")
			fmt.Println("   Migration 042 needs to run to fix this.")
		}
	} else {
		fmt.Println("Account 2102 NOT FOUND - This is OK if migration ran successfully")
	}
	
	// 3. Check if there are any journal entries for account 2102
	var journalCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM unified_journal_lines 
		WHERE account_code = '2102'
	`).Scan(&journalCount)
	
	fmt.Printf("\n=== JOURNAL ENTRIES FOR ACCOUNT 2102 ===\n")
	fmt.Printf("Count: %d entries\n", journalCount)
	
	if journalCount > 0 {
		var journals []struct {
			ID          uint
			AccountCode string
			AccountName string
			Debit       float64
			Credit      float64
			EntryDate   string
		}
		
		db.Raw(`
			SELECT id, account_code, account_name, debit, credit, entry_date
			FROM unified_journal_lines 
			WHERE account_code = '2102'
			ORDER BY entry_date DESC
			LIMIT 5
		`).Scan(&journals)
		
		fmt.Println("\nRecent journal entries:")
		for _, j := range journals {
			fmt.Printf("   ID: %d, Date: %s, Debit: %.0f, Credit: %.0f\n", 
				j.ID, j.EntryDate, j.Debit, j.Credit)
		}
	}
	
	// 4. Check account 1240 (PPN MASUKAN - target account)
	var account1240 struct {
		ID      uint
		Code    string
		Name    string
		Type    string
		Balance float64
	}
	
	db.Raw(`
		SELECT id, code, name, type, balance
		FROM accounts 
		WHERE code = '1240'
	`).Scan(&account1240)
	
	fmt.Println("\n=== ACCOUNT 1240 (PPN MASUKAN - TARGET) ===\n")
	if account1240.ID > 0 {
		fmt.Printf("Account 1240 EXISTS:\n")
		fmt.Printf("   ID: %d\n", account1240.ID)
		fmt.Printf("   Code: %s\n", account1240.Code)
		fmt.Printf("   Name: %s\n", account1240.Name)
		fmt.Printf("   Type: %s\n", account1240.Type)
		fmt.Printf("   Balance: Rp %.0f\n", account1240.Balance)
	} else {
		fmt.Println("❌ Account 1240 NOT FOUND - Migration needs to create it!")
	}
	
	// 5. Check RunAutoMigrations function to see which migrations are tracked
	fmt.Println("\n=== ALL MIGRATION RECORDS ===\n")
	
	var allMigrations []struct {
		Name       string
		ExecutedAt string
		Status     string
	}
	
	db.Raw(`
		SELECT name, executed_at, status 
		FROM migration_records 
		ORDER BY executed_at DESC 
		LIMIT 10
	`).Scan(&allMigrations)
	
	fmt.Println("Recent migrations:")
	for i, m := range allMigrations {
		fmt.Printf("%d. %s (%s) - %s\n", i+1, m.Name, m.Status, m.ExecutedAt)
	}
	
	// 6. Final recommendation
	fmt.Println("\n=== RECOMMENDATION ===\n")
	
	if migrationRecord.ID == 0 {
		fmt.Println("❌ Migration 042 has NOT been executed yet.")
		fmt.Println("\n📋 NEXT STEPS:")
		fmt.Println("1. Check if file 'migrations/042_fix_account_2102_and_balance_sheet.sql' exists")
		fmt.Println("2. Verify RunAutoMigrations() in database/database.go is configured correctly")
		fmt.Println("3. Check database/auto_migrations.go to see if 042 is registered")
		fmt.Println("4. Run: go run migrations/042_fix_account_2102_and_balance_sheet.sql manually")
		fmt.Println("\nOR restart backend to trigger auto-migration")
	} else if account2102.Balance != 0 && account2102.DeletedAt == nil {
		fmt.Println("⚠️  Migration 042 was recorded but account 2102 still has issues!")
		fmt.Println("   The migration may have failed partially.")
		fmt.Println("\n📋 NEXT STEPS:")
		fmt.Println("1. Manually run the SQL migration again")
		fmt.Println("2. Or delete migration record and restart backend")
	} else {
		fmt.Println("✅ Migration appears to have run successfully!")
		fmt.Println("   But balance sheet still shows diff Rp 165.000")
		fmt.Println("\n📋 INVESTIGATION NEEDED:")
		fmt.Println("1. Check if there are other accounts with similar issues")
		fmt.Println("2. Verify balance sheet query is using updated code")
		fmt.Println("3. Check for phantom balances in other accounts")
	}
}
