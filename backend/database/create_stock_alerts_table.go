package database

import (
	"log"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// CreateStockAlertsTable creates the stock_alerts table if it doesn't exist
func CreateStockAlertsTable(db *gorm.DB) error {
	log.Println("Creating stock_alerts table...")
	
	// Create the stock_alerts table
	if err := db.AutoMigrate(&models.StockAlert{}); err != nil {
		log.Printf("Failed to create stock_alerts table: %v", err)
		return err
	}
	
	log.Println("stock_alerts table created successfully")
	return nil
}

// RunStockAlertsMigration runs the stock alerts migration
func RunStockAlertsMigration(db *gorm.DB) {
	log.Println("Running stock alerts migration...")
	
	// Check if table exists first
	if !db.Migrator().HasTable(&models.StockAlert{}) {
		log.Println("stock_alerts table doesn't exist, creating...")
		if err := CreateStockAlertsTable(db); err != nil {
			log.Fatalf("Failed to create stock_alerts table: %v", err)
		}
	} else {
		log.Println("stock_alerts table already exists, skipping creation")
		
		// Still run AutoMigrate to ensure all columns are up to date
		if err := db.AutoMigrate(&models.StockAlert{}); err != nil {
			log.Printf("Failed to update stock_alerts table: %v", err)
		} else {
			log.Println("stock_alerts table structure updated successfully")
		}
	}
	
	log.Println("Stock alerts migration completed")
}
