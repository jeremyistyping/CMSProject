package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	// Connect to database using existing ConnectDB function
	db := database.ConnectDB()

	// Clean existing account data
	log.Println("Cleaning existing account data...")
	if err := cleanAccountData(db); err != nil {
		log.Fatalf("Failed to clean account data: %v", err)
	}

	// Reseed accounts
	log.Println("Reseeding account data...")
	if err := database.SeedAccounts(db); err != nil {
		log.Fatalf("Failed to seed account data: %v", err)
	}

	log.Println("Account data reset completed successfully!")
}

func cleanAccountData(db *gorm.DB) error {
	// Delete data in the right order to respect foreign key constraints
	
	// Delete journal entries
	if err := db.Exec("DELETE FROM journal_entries").Error; err != nil {
		log.Printf("Warning: Failed to delete journal_entries: %v", err)
	}
	
	// Delete transactions
	if err := db.Exec("DELETE FROM transactions").Error; err != nil {
		log.Printf("Warning: Failed to delete transactions: %v", err)
	}
	
	// Delete sale items
	if err := db.Exec("DELETE FROM sale_items").Error; err != nil {
		log.Printf("Warning: Failed to delete sale_items: %v", err)
	}
	
	// Delete purchase items
	if err := db.Exec("DELETE FROM purchase_items").Error; err != nil {
		log.Printf("Warning: Failed to delete purchase_items: %v", err)
	}
	
	// Delete assets
	if err := db.Exec("DELETE FROM assets").Error; err != nil {
		log.Printf("Warning: Failed to delete assets: %v", err)
	}
	
	// Finally delete accounts
	if err := db.Exec("DELETE FROM accounts").Error; err != nil {
		return fmt.Errorf("failed to delete accounts: %v", err)
	}

	// Reset auto increment sequences (PostgreSQL)
	db.Exec("ALTER SEQUENCE IF EXISTS accounts_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE IF EXISTS transactions_id_seq RESTART WITH 1")
	
	log.Println("Account data cleaned successfully")
	return nil
}
