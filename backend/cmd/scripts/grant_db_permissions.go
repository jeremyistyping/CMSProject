package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// grantStatements are safe, non-hardcoded privilege grants using CURRENT_USER
var grantStatements = []string{
	// Ensure current user can use and create in public schema
	"GRANT USAGE, CREATE ON SCHEMA public TO CURRENT_USER",

	// Ensure current user can create objects by default in public schema
	"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO CURRENT_USER",
	"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO CURRENT_USER",
	"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO CURRENT_USER",

	// Grant on existing objects
	"GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO CURRENT_USER",
	"GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO CURRENT_USER",
	"GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO CURRENT_USER",
}

func main() {
	fmt.Println("🔐 Setting up database permissions for CURRENT_USER (from DATABASE_URL)...")

	// Load .env so we don't hardcode credentials
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect via GORM to reuse pg driver
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Print current user for context (without hardcoding)
	printCurrentUser(sqlDB)

	// Execute each grant statement, continue on errors to be resilient
	for _, stmt := range grantStatements {
		fmt.Printf("➡️  %s...\n", stmt)
		if _, err := sqlDB.Exec(stmt); err != nil {
			fmt.Printf("   ⚠️  Skipped (not critical or insufficient rights): %v\n", err)
		} else {
			fmt.Println("   ✅ Done")
		}
	}

	fmt.Println("\n✅ Permission setup completed (best-effort). If creation previously failed, retry the setup now.")
}

func printCurrentUser(db *sql.DB) {
	var user, dbname string
	_ = db.QueryRow("SELECT CURRENT_USER, current_database()").Scan(&user, &dbname)
	if user != "" {
		fmt.Printf("Connected as: %s on database: %s\n", user, dbname)
	}
}
