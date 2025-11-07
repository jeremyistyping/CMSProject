package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== COLUMN STRUCTURE CHECK ===\n")

	// Check accounts table columns
	fmt.Println("1. ACCOUNTS TABLE COLUMNS:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'accounts' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Error fetching accounts columns: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			err := rows.Scan(&columnName, &dataType, &isNullable)
			if err != nil {
				fmt.Printf("Error scanning column: %v\n", err)
				continue
			}
			fmt.Printf("- %s (%s) - Nullable: %s\n", columnName, dataType, isNullable)
		}
	}

	// Check purchases table columns
	fmt.Println("\n2. PURCHASES TABLE COLUMNS:")
	rows, err = db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'purchases' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Error fetching purchases columns: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			err := rows.Scan(&columnName, &dataType, &isNullable)
			if err != nil {
				fmt.Printf("Error scanning column: %v\n", err)
				continue
			}
			fmt.Printf("- %s (%s) - Nullable: %s\n", columnName, dataType, isNullable)
		}
	}

	// Check journal_entries columns
	fmt.Println("\n3. JOURNAL_ENTRIES TABLE COLUMNS:")
	rows, err = db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'journal_entries' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Error fetching journal_entries columns: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			err := rows.Scan(&columnName, &dataType, &isNullable)
			if err != nil {
				fmt.Printf("Error scanning column: %v\n", err)
				continue
			}
			fmt.Printf("- %s (%s) - Nullable: %s\n", columnName, dataType, isNullable)
		}
	}

	// Sample data from accounts with correct column names
	fmt.Println("\n4. SAMPLE ACCOUNTS DATA:")
	rows, err = db.Query(`
		SELECT code, name, type, balance 
		FROM accounts 
		ORDER BY code 
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("Error reading accounts: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var code, name, accountType string
			var balance float64
			err := rows.Scan(&code, &name, &accountType, &balance)
			if err != nil {
				fmt.Printf("Error scanning account: %v\n", err)
				continue
			}
			fmt.Printf("%s - %s (%s): %.0f\n", code, name, accountType, balance)
		}
	}

	// Sample data from purchases
	fmt.Println("\n5. SAMPLE PURCHASES DATA:")
	rows, err = db.Query(`
		SELECT id, code, total_amount, status
		FROM purchases 
		ORDER BY code
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("Error reading purchases: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var code, status string
			var totalAmount float64
			err := rows.Scan(&id, &code, &totalAmount, &status)
			if err != nil {
				fmt.Printf("Error scanning purchase: %v\n", err)
				continue
			}
			fmt.Printf("ID:%d - %s: %.0f (%s)\n", id, code, totalAmount, status)
		}
	}

	fmt.Println("\n=== COLUMN CHECK SELESAI ===")
}