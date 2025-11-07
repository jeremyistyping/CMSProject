package main

import (
	"log"
	"os"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	log.Println("🔄 Recreating sales tables...")

	// Drop existing sales related tables
	log.Println("Dropping existing sales tables...")
	db.Exec("DROP TABLE IF EXISTS sale_return_items CASCADE")
	db.Exec("DROP TABLE IF EXISTS sale_returns CASCADE")
	db.Exec("DROP TABLE IF EXISTS sale_payments CASCADE")
	db.Exec("DROP TABLE IF EXISTS sale_items CASCADE")
	db.Exec("DROP TABLE IF EXISTS sales CASCADE")

	// Recreate tables using AutoMigrate
	log.Println("Creating fresh sales tables...")
	err := db.AutoMigrate(
		&models.Sale{},
		&models.SaleItem{},
		&models.SalePayment{},
		&models.SaleReturn{},
		&models.SaleReturnItem{},
	)

	if err != nil {
		log.Fatalf("Failed to recreate sales tables: %v", err)
	}

	log.Println("✅ Sales tables recreated successfully!")
	log.Println("You can now run the main application.")

	os.Exit(0)
}
