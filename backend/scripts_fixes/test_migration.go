package main

import (
	"log"
	
	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🧪 Testing Database Migration with Balance Sync...")
	
	// Connect to database and run full migration
	log.Println("📡 Connecting to database...")
	db := database.ConnectDB()
	
	log.Println("🔄 Running AutoMigrate (includes balance sync migration)...")
	database.AutoMigrate(db)
	
	log.Println("✅ Migration test completed successfully!")
}