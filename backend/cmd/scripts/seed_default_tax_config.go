package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database
	db := database.InitDB()
	
	// Check if tax config already exists
	var existingConfig models.TaxConfig
	err := db.Where("is_active = ? AND is_default = ?", true, true).First(&existingConfig).Error
	
	if err == nil {
		log.Printf("✅ Tax Config already exists: %s", existingConfig.ConfigName)
		return
	}
	
	// Create default tax config
	log.Println("📝 Creating default Tax Configuration...")
	
	defaultConfig := models.GetDefaultTaxConfig()
	defaultConfig.UpdatedBy = 1 // Admin user ID
	
	if err := db.Create(defaultConfig).Error; err != nil {
		log.Fatalf("❌ Failed to create tax config: %v", err)
	}
	
	log.Printf("✅ Successfully created Tax Config: %s", defaultConfig.ConfigName)
	log.Printf("   - Sales PPN Rate: %.2f%%", defaultConfig.SalesPPNRate)
	log.Printf("   - Purchase PPN Rate: %.2f%%", defaultConfig.PurchasePPNRate)
	log.Printf("   - Purchase PPh23 Rate: %.2f%%", defaultConfig.PurchasePPh23Rate)
	log.Println("🎉 Tax Configuration is now active!")
}
