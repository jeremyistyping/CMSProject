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

	fmt.Println("🔧 Fixing Foreign Key Constraint")
	fmt.Println("================================")

	// First, drop the incorrect foreign key constraint
	fmt.Println("\n1. Dropping incorrect foreign key constraint...")
	result := db.Exec("ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS fk_journal_entries_journal_lines")
	if result.Error != nil {
		fmt.Printf("   Error dropping constraint: %v\n", result.Error)
	} else {
		fmt.Println("   ✓ Dropped old constraint successfully")
	}

	// Create the correct foreign key constraint
	fmt.Println("\n2. Creating correct foreign key constraint...")
	result = db.Exec("ALTER TABLE journal_lines ADD CONSTRAINT fk_journal_entries_journal_lines FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id) ON DELETE CASCADE")
	if result.Error != nil {
		fmt.Printf("   Error creating constraint: %v\n", result.Error)
	} else {
		fmt.Println("   ✓ Created correct constraint successfully")
	}

	// Verify the fix
	fmt.Println("\n3. Verifying foreign key constraint...")
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
			AND tc.table_name = 'journal_lines'
			AND kcu.column_name = 'journal_entry_id'
	`).Rows()
	
	if err != nil {
		fmt.Printf("   Error checking constraints: %v\n", err)
	} else {
		defer constraintRows.Close()
		found := false
		for constraintRows.Next() {
			found = true
			var constraintName, tableName, columnName, foreignTableName, foreignColumnName string
			constraintRows.Scan(&constraintName, &tableName, &columnName, &foreignTableName, &foreignColumnName)
			fmt.Printf("   ✓ Constraint: %s: %s.%s -> %s.%s\n", constraintName, tableName, columnName, foreignTableName, foreignColumnName)
		}
		if !found {
			fmt.Println("   ❌ No foreign key constraint found")
		}
	}

	fmt.Println("\n✅ Foreign key constraint fix completed!")
	fmt.Println("   Now try running the journal lines creation script again.")
}