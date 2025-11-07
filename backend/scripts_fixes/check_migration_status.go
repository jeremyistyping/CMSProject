package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("DB_PASS") 
	if dbPass == "" {
		dbPass = "postgres"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "accounting_system"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔍 Checking purchase_balance migration status...")
	fmt.Println("=" + string(make([]byte, 80)) + "=")
	fmt.Printf("%-35s %-10s %-20s %s\n", "FILENAME", "STATUS", "CREATED_AT", "ERROR_MESSAGE")
	fmt.Println("-" + string(make([]byte, 80)) + "-")

	// Check migration logs for purchase_balance related migrations
	query := `
		SELECT filename, status, created_at, COALESCE(error_message, '') as error_message 
		FROM migration_logs 
		WHERE filename LIKE '%purchase_balance%' 
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying migration logs: %v", err)
		return
	}
	defer rows.Close()

	migrationCount := 0
	for rows.Next() {
		var filename, status, createdAt, errorMessage string
		if err := rows.Scan(&filename, &status, &createdAt, &errorMessage); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		
		// Truncate error message if too long
		if len(errorMessage) > 50 {
			errorMessage = errorMessage[:47] + "..."
		}
		
		// Color coding for status
		statusDisplay := status
		switch status {
		case "SUCCESS":
			statusDisplay = "✅ SUCCESS"
		case "SKIPPED":
			statusDisplay = "⏭️  SKIPPED"
		case "FAILED":
			statusDisplay = "❌ FAILED"
		}

		fmt.Printf("%-35s %-10s %-20s %s\n", filename, statusDisplay, createdAt[:19], errorMessage)
		migrationCount++
	}

	if migrationCount == 0 {
		fmt.Println("No purchase_balance migrations found in migration_logs table.")
	} else {
		fmt.Printf("\nFound %d purchase_balance migration records.\n", migrationCount)
	}

	// Check if purchase balance functions exist
	fmt.Println("\n🔍 Checking if purchase balance functions exist...")
	
	functionChecks := []string{
		"get_purchase_balance_summary",
		"calculate_purchase_outstanding_amount",
		"sync_purchase_balance",
	}

	for _, funcName := range functionChecks {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_proc p 
				JOIN pg_namespace n ON p.pronamespace = n.oid 
				WHERE n.nspname = 'public' AND p.proname = $1
			)
		`, funcName).Scan(&exists)
		
		if err != nil {
			fmt.Printf("❓ %s: Error checking - %v\n", funcName, err)
		} else if exists {
			fmt.Printf("✅ %s: EXISTS\n", funcName)
		} else {
			fmt.Printf("❌ %s: NOT FOUND\n", funcName)
		}
	}
}