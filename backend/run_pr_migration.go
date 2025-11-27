package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Connection with correct credentials from .env
	dsn := "host=localhost user=postgres password=Moon dbname=CMSNew port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to database successfully")

	// Read SQL file
	sqlBytes, err := os.ReadFile("migrations/056_create_purchase_requests_table.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Execute SQL
	log.Println("Executing migration SQL...")
	err = db.Exec(string(sqlBytes)).Error
	if err != nil {
		log.Fatal("Failed to execute SQL:", err)
	}

	log.Println("✅ Migration completed successfully!")
	log.Println("Tables 'purchase_requests' and 'purchase_request_items' created")
}
