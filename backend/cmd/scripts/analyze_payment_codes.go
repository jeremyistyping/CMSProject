package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
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

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🔍 Analyzing Payment Code Sequences")
	fmt.Println("=====================================")

	// Check if payment_code_sequences table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = 'payment_code_sequences'", dbName).Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		return
	}

	if tableExists == 0 {
		fmt.Println("❌ payment_code_sequences table does not exist!")
		fmt.Println("This explains why payment codes are not sequential - the system is falling back to timestamp-based codes.")
		return
	}

	// Check payment_code_sequences table
	fmt.Println("✅ payment_code_sequences table exists")
	rows, err := db.Query("SELECT prefix, year, month, sequence_number, created_at, updated_at FROM payment_code_sequences ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		log.Printf("Error querying sequences: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-10s %-6s %-6s %-8s %-20s %-20s\n", "PREFIX", "YEAR", "MONTH", "SEQ_NUM", "CREATED", "UPDATED")
	fmt.Println("=================================================================================")

	hasSequences := false
	for rows.Next() {
		hasSequences = true
		var prefix string
		var year, month, seqNum int
		var createdAt, updatedAt string

		err := rows.Scan(&prefix, &year, &month, &seqNum, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}

		fmt.Printf("%-10s %-6d %-6d %-8d %-20s %-20s\n", prefix, year, month, seqNum, createdAt[:19], updatedAt[:19])
	}

	if !hasSequences {
		fmt.Println("❌ No sequence records found!")
	}

	// Check recent payments and their codes
	fmt.Println("\n🔍 Recent payment codes analysis:")
	paymentRows, err := db.Query("SELECT id, code, created_at FROM payments ORDER BY id DESC LIMIT 20")
	if err != nil {
		log.Printf("Error querying payments: %v", err)
		return
	}
	defer paymentRows.Close()

	fmt.Printf("%-5s %-20s %-20s %-15s\n", "ID", "CODE", "CREATED", "CODE_TYPE")
	fmt.Println("==================================================================")

	for paymentRows.Next() {
		var id int
		var code, createdAt string

		err := paymentRows.Scan(&id, &code, &createdAt)
		if err != nil {
			log.Printf("Error scanning payment: %v", err)
			continue
		}

		// Determine code type
		codeType := "SEQUENTIAL"
		if len(code) > 10 && code[len(code)-8:len(code)-2] == "TS" {
			codeType = "TIMESTAMP"
		}

		fmt.Printf("%-5d %-20s %-20s %-15s\n", id, code, createdAt[:19], codeType)
	}

	// Count different code types
	var sequentialCount, timestampCount int
	db.QueryRow("SELECT COUNT(*) FROM payments WHERE code NOT LIKE '%TS%'").Scan(&sequentialCount)
	db.QueryRow("SELECT COUNT(*) FROM payments WHERE code LIKE '%TS%'").Scan(&timestampCount)

	fmt.Println("\n📊 Payment Code Statistics:")
	fmt.Printf("Sequential codes: %d\n", sequentialCount)
	fmt.Printf("Timestamp codes: %d\n", timestampCount)

	if timestampCount > 0 {
		fmt.Println("\n❌ Issue found: Some payment codes are using timestamp fallback")
		fmt.Println("This indicates that the sequence generation is failing occasionally")
	} else {
		fmt.Println("\n✅ All payment codes are using proper sequence")
	}
}