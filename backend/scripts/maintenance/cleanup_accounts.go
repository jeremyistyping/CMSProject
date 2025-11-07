package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get the SQL DB instance
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get SQL database:", err)
	}
	defer sqlDB.Close()

	fmt.Println("🔍 Checking for account database issues...")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Check for duplicates
	fmt.Println("1. Checking for duplicate accounts...")
	rows, err := sqlDB.Query(`
		SELECT code, COUNT(*) as count 
		FROM accounts 
		GROUP BY code 
		HAVING COUNT(*) > 1
	`)
	if err != nil {
		log.Fatal("Failed to check duplicates:", err)
	}

	duplicateFound := false
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			log.Fatal("Failed to scan duplicate row:", err)
		}
		fmt.Printf("   ⚠️  Code %s has %d duplicates\n", code, count)
		duplicateFound = true
	}
	rows.Close()

	if !duplicateFound {
		fmt.Println("   ✅ No duplicate codes found")
	}

	// 2. Check specific problematic account
	fmt.Println("\n2. Checking for problematic account '1009'...")
	rows, err = sqlDB.Query(`
		SELECT code, name, deleted_at, id
		FROM accounts 
		WHERE code = '1009'
		ORDER BY id
	`)
	if err != nil {
		log.Fatal("Failed to check account 1009:", err)
	}

	account1009Found := false
	for rows.Next() {
		var code, name string
		var deletedAt interface{}
		var id int
		if err := rows.Scan(&code, &name, &deletedAt, &id); err != nil {
			log.Fatal("Failed to scan 1009 row:", err)
		}
		fmt.Printf("   Found: ID=%d, Code=%s, Name=%s, Deleted=%v\n", id, code, name, deletedAt)
		account1009Found = true
	}
	rows.Close()

	if !account1009Found {
		fmt.Println("   ✅ No account with code '1009' found")
	}

	// 3. Clean up problematic account
	if account1009Found {
		fmt.Println("\n3. Cleaning up problematic account '1009'...")
		result, err := sqlDB.Exec("DELETE FROM accounts WHERE code = '1009' AND name = 'bank bri2'")
		if err != nil {
			log.Printf("   ⚠️  Failed to delete problematic account: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("   ✅ Deleted %d problematic account(s)\n", rowsAffected)
			} else {
				fmt.Println("   ℹ️  No problematic accounts to delete")
			}
		}
	}

	// 4. Check for any other duplicates
	fmt.Println("\n4. Checking for other duplicate issues...")
	rows, err = sqlDB.Query(`
		SELECT a1.id, a1.code, a1.name, a1.deleted_at
		FROM accounts a1
		INNER JOIN accounts a2 ON a1.code = a2.code AND a1.id != a2.id
		ORDER BY a1.code, a1.id
	`)
	if err != nil {
		log.Fatal("Failed to check other duplicates:", err)
	}

	otherDuplicatesFound := false
	for rows.Next() {
		var id int
		var code, name string
		var deletedAt interface{}
		if err := rows.Scan(&id, &code, &name, &deletedAt); err != nil {
			log.Fatal("Failed to scan duplicate row:", err)
		}
		fmt.Printf("   ⚠️  Duplicate: ID=%d, Code=%s, Name=%s, Deleted=%v\n", id, code, name, deletedAt)
		otherDuplicatesFound = true
	}
	rows.Close()

	if !otherDuplicatesFound {
		fmt.Println("   ✅ No other duplicates found")
	}

	// 5. Show current accounts status
	fmt.Println("\n5. Current accounts status:")
	fmt.Println("   Code   | Name                     | ID | Active | Deleted")
	fmt.Println("   -------|--------------------------|----|---------|---------")
	
	rows, err = sqlDB.Query(`
		SELECT code, name, id, is_active, CASE WHEN deleted_at IS NULL THEN 'No' ELSE 'Yes' END as deleted
		FROM accounts 
		ORDER BY code
	`)
	if err != nil {
		log.Fatal("Failed to get accounts status:", err)
	}

	for rows.Next() {
		var code, name, deleted string
		var id int
		var isActive bool
		if err := rows.Scan(&code, &name, &id, &isActive, &deleted); err != nil {
			log.Fatal("Failed to scan status row:", err)
		}
		
		activeStr := "No"
		if isActive {
			activeStr = "Yes"
		}
		
		fmt.Printf("   %-6s | %-24s | %-2d | %-7s | %-7s\n", 
			code, 
			name[:min(24, len(name))], 
			id, 
			activeStr, 
			deleted)
	}
	rows.Close()

	fmt.Println("\n✅ Database cleanup completed!")
	fmt.Println("You can now try to create new accounts again.")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
