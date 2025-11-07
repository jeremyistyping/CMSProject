package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Checking products and stock levels...")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get all products
	var products []models.Product
	err = db.Find(&products).Error
	if err != nil {
		log.Fatalf("❌ Failed to get products: %v", err)
	}

	log.Printf("📊 Found %d products:", len(products))
	for _, product := range products {
		log.Printf("   Product ID: %d, Name: %s, Stock: %d, SalePrice: %.2f", 
			product.ID, product.Name, product.Stock, product.SalePrice)
	}

	// Update products with stock if they have zero stock
	for _, product := range products {
		if product.Stock == 0 {
			product.Stock = 100
			err = db.Save(&product).Error
			if err != nil {
				log.Printf("❌ Failed to update stock for product %d: %v", product.ID, err)
			} else {
				log.Printf("✅ Updated stock for product %d (%s) to 100", product.ID, product.Name)
			}
		}
	}
}