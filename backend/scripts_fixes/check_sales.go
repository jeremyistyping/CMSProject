package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Checking existing sales...")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get all sales
	var sales []models.Sale
	err = db.Order("id DESC").Limit(10).Find(&sales).Error
	if err != nil {
		log.Fatalf("❌ Failed to get sales: %v", err)
	}

	log.Printf("📊 Found %d sales:", len(sales))
	for _, sale := range sales {
		log.Printf("   Sale ID: %d, Code: %s, Status: %s, PaymentMethod: %s, Total: %.2f", 
			sale.ID, sale.Code, sale.Status, sale.PaymentMethodType, sale.TotalAmount)
	}

	// Check cash banks
	var cashBanks []models.CashBank
	err = db.Where("is_active = ?", true).Find(&cashBanks).Error
	if err == nil {
		log.Printf("\n💰 Available cash banks:")
		for _, cb := range cashBanks {
			log.Printf("   CashBank ID: %d, Name: %s, AccountID: %v", cb.ID, cb.Name, cb.AccountID)
		}
	}
}