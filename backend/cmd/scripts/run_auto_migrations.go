package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔄 Running auto SQL migrations...")
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	if err := database.RunAutoMigrations(db); err != nil {
		log.Fatalf("❌ Auto-migrations failed: %v", err)
	}
	fmt.Println("✅ Auto-migrations completed successfully")
}
