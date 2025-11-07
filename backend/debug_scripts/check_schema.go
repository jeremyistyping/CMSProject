package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🔍 Checking Database Schema")
	fmt.Println("===========================")

	// Check journal_entries table structure
	fmt.Println("\n📊 journal_entries table:")
	rows, err := db.Raw("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'journal_entries' ORDER BY ordinal_position").Rows()
	if err != nil {
		log.Printf("Error querying journal_entries: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			rows.Scan(&columnName, &dataType, &isNullable)
			fmt.Printf("   - %s: %s (Nullable: %s)\n", columnName, dataType, isNullable)
		}
	}

	// Check journal_lines table structure
	fmt.Println("\n📝 journal_lines table:")
	rows2, err := db.Raw("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'journal_lines' ORDER BY ordinal_position").Rows()
	if err != nil {
		log.Printf("Error querying journal_lines: %v", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var columnName, dataType, isNullable string
			rows2.Scan(&columnName, &dataType, &isNullable)
			fmt.Printf("   - %s: %s (Nullable: %s)\n", columnName, dataType, isNullable)
		}
	}

	// Check foreign key constraints
	fmt.Println("\n🔗 Foreign key constraints:")
	constraintRows, err := db.Raw(`
		SELECT 
			tc.constraint_name, 
			tc.table_name, 
			kcu.column_name, 
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name 
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND (tc.table_name = 'journal_lines' OR tc.table_name = 'journal_entries')
		ORDER BY tc.table_name, tc.constraint_name
	`).Rows()
	
	if err != nil {
		log.Printf("Error querying constraints: %v", err)
	} else {
		defer constraintRows.Close()
		for constraintRows.Next() {
			var constraintName, tableName, columnName, foreignTableName, foreignColumnName string
			constraintRows.Scan(&constraintName, &tableName, &columnName, &foreignTableName, &foreignColumnName)
			fmt.Printf("   - %s: %s.%s -> %s.%s\n", constraintName, tableName, columnName, foreignTableName, foreignColumnName)
		}
	}

	// Check actual journal entries
	fmt.Println("\n📋 Existing journal entries:")
	entryRows, err := db.Raw("SELECT id, code, description, status FROM journal_entries ORDER BY id").Rows()
	if err != nil {
		log.Printf("Error querying entries: %v", err)
	} else {
		defer entryRows.Close()
		for entryRows.Next() {
			var id int
			var code, description, status string
			entryRows.Scan(&id, &code, &description, &status)
			fmt.Printf("   - ID: %d, Code: %s, Status: %s, Description: %s\n", id, code, status, description)
		}
	}

	// Check existing journal lines
	fmt.Println("\n📝 Existing journal lines:")
	lineRows, err := db.Raw("SELECT id, journal_entry_id, account_id, debit_amount, credit_amount FROM journal_lines ORDER BY id LIMIT 10").Rows()
	if err != nil {
		log.Printf("Error querying lines: %v", err)
	} else {
		defer lineRows.Close()
		for lineRows.Next() {
			var id, journalEntryID, accountID int
			var debitAmount, creditAmount float64
			lineRows.Scan(&id, &journalEntryID, &accountID, &debitAmount, &creditAmount)
			fmt.Printf("   - ID: %d, EntryID: %d, AccountID: %d, Debit: %.2f, Credit: %.2f\n", id, journalEntryID, accountID, debitAmount, creditAmount)
		}
	}

	fmt.Println("\n✅ Schema check completed!")
}