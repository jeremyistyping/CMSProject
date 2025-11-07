package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  FIXING PRODUCT STOCK SETTINGS")
	fmt.Println("========================================")

	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Get all active products
	var products []models.Product
	err := db.Where("is_active = ?", true).Find(&products).Error
	if err != nil {
		log.Fatalf("Error fetching products: %v", err)
	}

	fmt.Printf("Found %d active products\n\n", len(products))

	// Update products with proper stock settings
	updates := []struct {
		ProductID   uint
		MinStock    int
		ReorderLevel int
	}{
		{1, 5, 10},   // Laptop Dell XPS 13
		{2, 10, 20},  // Mouse Wireless Logitech  
		{3, 20, 30},  // Kertas A4 80gsm
	}

	for _, update := range updates {
		var product models.Product
		err := db.First(&product, update.ProductID).Error
		if err != nil {
			fmt.Printf("❌ Product ID %d not found: %v\n", update.ProductID, err)
			continue
		}

		// Update min_stock and reorder_level
		product.MinStock = update.MinStock
		product.ReorderLevel = update.ReorderLevel

		err = db.Save(&product).Error
		if err != nil {
			fmt.Printf("❌ Error updating Product ID %d: %v\n", update.ProductID, err)
			continue
		}

		fmt.Printf("✅ Updated Product %s (ID: %d):\n", product.Name, product.ID)
		fmt.Printf("   - Min Stock: %d\n", product.MinStock)
		fmt.Printf("   - Reorder Level: %d\n", product.ReorderLevel)
		fmt.Printf("   - Current Stock: %d\n", product.Stock)
		
		// Check if notification should be triggered
		if product.Stock <= product.MinStock {
			fmt.Printf("   - 🚨 Should trigger MIN_STOCK notification\n")
		}
		if product.Stock <= product.ReorderLevel {
			fmt.Printf("   - 📋 Should trigger REORDER_ALERT notification\n")
		}
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Println("  STOCK SETTINGS UPDATE COMPLETED")
	fmt.Println("========================================")
	
	// Verify the updates
	fmt.Println("\n🔍 Verification - Current product stock settings:")
	var updatedProducts []models.Product
	err = db.Where("is_active = ?", true).Find(&updatedProducts).Error
	if err != nil {
		log.Fatalf("Error fetching updated products: %v", err)
	}

	for _, product := range updatedProducts {
		fmt.Printf("Product %s (ID: %d):\n", product.Name, product.ID)
		fmt.Printf("  - Current Stock: %d\n", product.Stock)
		fmt.Printf("  - Min Stock: %d\n", product.MinStock)
		fmt.Printf("  - Reorder Level: %d\n", product.ReorderLevel)
		
		if product.MinStock > 0 && product.Stock <= product.MinStock {
			fmt.Printf("  - 🔴 NOTIFICATION SHOULD BE TRIGGERED!\n")
		} else if product.MinStock > 0 {
			fmt.Printf("  - ✅ Stock is above minimum\n")
		} else {
			fmt.Printf("  - ⚪ No min stock set\n")
		}
		fmt.Println()
	}
}