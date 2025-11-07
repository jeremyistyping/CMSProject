package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	// Read SQL file
	sqlFile := "../../fix_chart_of_accounts.sql"
	if len(os.Args) > 1 {
		sqlFile = os.Args[1]
	}

	sqlBytes, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	sqlContent := string(sqlBytes)
	fmt.Println("Executing SQL script:", sqlFile)
	fmt.Println("Content:")
	fmt.Println(sqlContent)
	fmt.Println("\n" + strings.Repeat("=", 50))

	// Execute SQL script
	result, err := sqlDB.Exec(sqlContent)
	if err != nil {
		log.Fatal("Failed to execute SQL script:", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("SQL script executed successfully! Rows affected: %d\n", rowsAffected)

	// Now run the verification query
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("VERIFICATION - Current Account Hierarchy:")
	fmt.Println(strings.Repeat("=", 50))

	rows, err := sqlDB.Query(`
		SELECT 
			code || ' - ' || name as account_hierarchy,
			type,
			is_header,
			level
		FROM accounts 
		ORDER BY code
	`)
	if err != nil {
		log.Fatal("Failed to execute verification query:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hierarchy, accountType string
		var isHeader bool
		var level int
		
		if err := rows.Scan(&hierarchy, &accountType, &isHeader, &level); err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		
		headerMark := ""
		if isHeader {
			headerMark = " [HEADER]"
		}
		
		fmt.Printf("%-40s %-10s Level:%d%s\n", hierarchy, accountType, level, headerMark)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}

	fmt.Println("\n✅ Script executed successfully!")
	fmt.Println("You can now refresh the frontend at http://localhost:3000/accounts")
}
