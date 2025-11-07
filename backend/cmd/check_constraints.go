package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database: %s\n", databaseURL)
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 Checking migration_logs constraints...")

	// Check constraints on migration_logs table
	var constraints []struct {
		ConstraintName   string `gorm:"column:constraint_name"`
		ConstraintType   string `gorm:"column:constraint_type"`
		CheckClause      string `gorm:"column:check_clause"`
	}

	result := db.Raw(`
		SELECT 
			tc.constraint_name,
			tc.constraint_type,
			COALESCE(cc.check_clause, '') as check_clause
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.check_constraints cc ON tc.constraint_name = cc.constraint_name
		WHERE tc.table_name = 'migration_logs'
		ORDER BY tc.constraint_type, tc.constraint_name
	`).Scan(&constraints)

	if result.Error != nil {
		fmt.Printf("❌ Failed to get constraints: %v\n", result.Error)
		return
	}

	fmt.Println("\n📊 Constraints on migration_logs table:")
	fmt.Println("======================================")
	for _, constraint := range constraints {
		fmt.Printf("  - %s (%s)\n", constraint.ConstraintName, constraint.ConstraintType)
		if constraint.CheckClause != "" {
			fmt.Printf("    Check: %s\n", constraint.CheckClause)
		}
	}

	// Check allowed status values
	fmt.Println("\n📋 Checking existing status values:")
	fmt.Println("==================================")
	var statusValues []struct {
		Status string `gorm:"column:status"`
		Count  int    `gorm:"column:count"`
	}

	result = db.Raw("SELECT status, COUNT(*) as count FROM migration_logs GROUP BY status ORDER BY count DESC").Scan(&statusValues)
	if result.Error != nil {
		fmt.Printf("❌ Failed to get status values: %v\n", result.Error)
		return
	}

	for _, status := range statusValues {
		fmt.Printf("  - %s: %d records\n", status.Status, status.Count)
	}

	fmt.Println("\n✅ Constraint check completed!")
}