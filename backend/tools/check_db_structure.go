package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== DATABASE STRUCTURE CHECK ===\n")

	// Check all tables
	fmt.Println("1. TABLES IN DATABASE:")
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatal("Error fetching tables:", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatal("Error scanning table name:", err)
		}
		tables = append(tables, tableName)
		fmt.Printf("- %s\n", tableName)
	}

	// Check accounts table specifically
	fmt.Println("\n2. ACCOUNTS TABLE STRUCTURE:")
	if contains(tables, "accounts") {
		rows, err := db.Query(`
			SELECT code, name, account_type, balance 
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
	}

	// Check purchases table
	fmt.Println("\n3. PURCHASES TABLE DATA:")
	if contains(tables, "purchases") {
		rows, err := db.Query(`
			SELECT id, purchase_code, total_amount, status, approval_status
			FROM purchases 
			ORDER BY purchase_code
			LIMIT 10
		`)
		if err != nil {
			fmt.Printf("Error reading purchases: %v\n", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var id int
				var purchaseCode, status, approvalStatus string
				var totalAmount float64
				err := rows.Scan(&id, &purchaseCode, &totalAmount, &status, &approvalStatus)
				if err != nil {
					fmt.Printf("Error scanning purchase: %v\n", err)
					continue
				}
				fmt.Printf("ID:%d - %s: %.0f (%s/%s)\n", id, purchaseCode, totalAmount, status, approvalStatus)
			}
		}
	}

	// Check journal tables
	fmt.Println("\n4. JOURNAL TABLES:")
	if contains(tables, "journals") {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM journals").Scan(&count)
		if err != nil {
			fmt.Printf("Error counting journals: %v\n", err)
		} else {
			fmt.Printf("Total journals: %d\n", count)
		}
	}

	if contains(tables, "journal_entries") {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM journal_entries").Scan(&count)
		if err != nil {
			fmt.Printf("Error counting journal entries: %v\n", err)
		} else {
			fmt.Printf("Total journal entries: %d\n", count)
		}
	}

	fmt.Println("\n=== STRUCTURE CHECK SELESAI ===")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}