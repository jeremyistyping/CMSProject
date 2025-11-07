package main

import (
	"fmt"
	"log"

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

	fmt.Println("=== CHECKING CASH BANKS AND ACCOUNT RELATIONSHIPS ===")

	// Check cash_banks and their account relationships
	rows, err := sqlDB.Query(`
		SELECT 
			cb.id,
			cb.code,
			cb.name,
			cb.type,
			cb.account_id,
			COALESCE(a.code, 'NO ACCOUNT') as account_code,
			COALESCE(a.name, 'MISSING ACCOUNT') as account_name,
			cb.is_active
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id 
		ORDER BY cb.id
	`)
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}
	defer rows.Close()

	fmt.Printf("%-4s %-15s %-20s %-6s %-10s %-15s %-25s %-6s\n", 
		"ID", "Code", "Name", "Type", "AcctID", "AcctCode", "AcctName", "Active")
	fmt.Println(fmt.Sprintf("%s", "======================================================================================"))

	for rows.Next() {
		var id int
		var accountID *int // Use pointer to handle NULL
		var code, name, cbType, accountCode, accountName string
		var isActive bool
		
		if err := rows.Scan(&id, &code, &name, &cbType, &accountID, &accountCode, &accountName, &isActive); err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		
		activeMark := "Yes"
		if !isActive {
			activeMark = "No"
		}
		
		actIdStr := "NULL"
		if accountID != nil {
			actIdStr = fmt.Sprintf("%d", *accountID)
		}
		
		fmt.Printf("%-4d %-15s %-20s %-6s %-10s %-15s %-25s %-6s\n", 
			id, code, name, cbType, actIdStr, accountCode, accountName, activeMark)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}

	fmt.Println("\n=== CHECKING ACCOUNTS WITHOUT CASH_BANKS ===")
	
	// Check accounts that could be used for cash_banks
	rows2, err := sqlDB.Query(`
		SELECT 
			a.id,
			a.code,
			a.name,
			a.type,
			a.category
		FROM accounts a 
		WHERE a.type = 'ASSET' 
		AND a.category = 'CURRENT_ASSET'
		AND a.is_active = true
		AND a.code LIKE '11%'
		AND NOT EXISTS (SELECT 1 FROM cash_banks cb WHERE cb.account_id = a.id)
		ORDER BY a.code
	`)
	if err != nil {
		log.Fatal("Failed to execute accounts query:", err)
	}
	defer rows2.Close()

	fmt.Printf("%-4s %-10s %-30s %-10s %-15s\n", 
		"ID", "Code", "Name", "Type", "Category")
	fmt.Println("=================================================================")

	for rows2.Next() {
		var id int
		var code, name, accType, category string
		
		if err := rows2.Scan(&id, &code, &name, &accType, &category); err != nil {
			log.Fatal("Failed to scan accounts row:", err)
		}
		
		fmt.Printf("%-4d %-10s %-30s %-10s %-15s\n", 
			id, code, name, accType, category)
	}

	fmt.Println("\n✅ Check completed!")
}
