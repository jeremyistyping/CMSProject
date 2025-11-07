package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔧 FIXING SPECIFIC PRODUCT COST PRICES")
	log.Println("========================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()
	log.Println("✅ Database connected\n")

	// Update products with realistic cost prices (60% of selling price)
	updates := []struct {
		ProductName string
		SalePrice   float64
		NewCostPrice float64
	}{
		{"Kertas A4 80gsm", 5000000, 3000000},          // 60% of 5M = 3M
		{"Mouse Wireless", 999999999999.99, 600000000000}, // 60% of 1T = 600B
	}

	log.Println("📊 Updating product cost prices...")
	log.Println("   Product Name         | Old Cost    | New Cost        | Sale Price")
	log.Println("   ---------------------|-------------|-----------------|------------------")

	for _, update := range updates {
		var product models.Product
		if err := db.Where("name LIKE ?", update.ProductName+"%").First(&product).Error; err != nil {
			log.Printf("   ⚠️  Product '%s' not found, skipping", update.ProductName)
			continue
		}

		oldCost := product.CostPrice
		product.CostPrice = update.NewCostPrice

		if err := db.Save(&product).Error; err != nil {
			log.Printf("   ❌ Failed to update '%s': %v", update.ProductName, err)
			continue
		}

		log.Printf("   %-20s | Rp %10.0f | Rp %13.0f | Rp %15.0f",
			truncate(product.Name, 20), oldCost, product.CostPrice, product.SalePrice)
	}

	log.Println("\n========================================")
	log.Println("✅ Product cost prices updated!")
	log.Println("\n💡 Next steps:")
	log.Println("   1. Delete old COGS journal entries")
	log.Println("   2. Run backfill script to create new COGS with correct values")
	log.Println("========================================")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
