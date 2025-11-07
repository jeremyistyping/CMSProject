package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// FixProductCostPrices fixes products with zero or null cost prices
func main() {
	log.Println("========================================")
	log.Println("🔧 FIX PRODUCT COST PRICES")
	log.Println("========================================")

	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database
	db := database.ConnectDB()
	log.Println("✅ Database connected")

	// Step 1: Find products with zero or null cost prices
	log.Println("\n📊 Step 1: Finding products with missing cost prices...")
	var productsNoCost []models.Product
	if err := db.Where("cost_price = 0 OR cost_price IS NULL").
		Find(&productsNoCost).Error; err != nil {
		log.Fatalf("❌ Failed to load products: %v", err)
	}

	log.Printf("   Found %d products with zero/null cost price", len(productsNoCost))

	if len(productsNoCost) == 0 {
		log.Println("\n✅ All products have valid cost prices. Nothing to do!")
		return
	}

	// Display products
	log.Println("\n📋 Products without cost price:")
	log.Println("   ID   | Name                          | Price       | Cost Price")
	log.Println("   -----|-------------------------------|-------------|------------")
	for _, p := range productsNoCost {
		log.Printf("   %-4d | %-29s | Rp %10.2f | Rp %10.2f",
			p.ID, truncateString(p.Name, 29), p.SalePrice, p.CostPrice)
	}

	// Step 2: Choose fixing method
	log.Println("\n🔧 Choose cost price fixing method:")
	log.Println("   1. Auto-calculate from selling price (recommended)")
	log.Println("   2. Set from last purchase price")
	log.Println("   3. Manual entry per product")
	log.Println("   4. Cancel")
	
	fmt.Print("\nEnter choice (1-4): ")
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		fixByPercentage(db, productsNoCost)
	case 2:
		fixByPurchasePrice(db, productsNoCost)
	case 3:
		fixManually(db, productsNoCost)
	case 4:
		log.Println("❌ Operation cancelled")
		return
	default:
		log.Println("❌ Invalid choice")
		return
	}
}

// fixByPercentage calculates cost price as percentage of selling price
func fixByPercentage(db *gorm.DB, products []models.Product) {
	log.Println("\n💡 Auto-calculate cost price from selling price")
	log.Println("   Formula: Cost Price = Selling Price × Percentage")
	
	fmt.Print("\nEnter cost percentage (e.g., 60 for 60%): ")
	var percentage float64
	fmt.Scanln(&percentage)

	if percentage <= 0 || percentage >= 100 {
		log.Println("❌ Invalid percentage. Must be between 0 and 100")
		return
	}

	multiplier := percentage / 100

	log.Printf("\n📊 Preview (using %.0f%% of selling price):", percentage)
	log.Println("   ID   | Name                          | Price       | New Cost Price")
	log.Println("   -----|-------------------------------|-------------|----------------")
	for _, p := range products {
		newCost := p.SalePrice * multiplier
		log.Printf("   %-4d | %-29s | Rp %10.2f | Rp %10.2f",
			p.ID, truncateString(p.Name, 29), p.SalePrice, newCost)
	}

	fmt.Print("\n❓ Apply these changes? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" && confirm != "y" {
		log.Println("❌ Operation cancelled")
		return
	}

	// Apply changes
	successCount := 0
	for _, p := range products {
		p.CostPrice = p.SalePrice * multiplier
		if err := db.Save(&p).Error; err != nil {
			log.Printf("   ❌ Failed to update product #%d: %v", p.ID, err)
		} else {
			successCount++
		}
	}

	log.Printf("\n✅ Successfully updated %d/%d products", successCount, len(products))
}

// fixByPurchasePrice sets cost price from last purchase price
func fixByPurchasePrice(db *gorm.DB, products []models.Product) {
	log.Println("\n💡 Set cost price from last purchase price")
	
	successCount := 0
	notFoundCount := 0

	for _, product := range products {
		// Find last purchase for this product
		var lastPurchaseItem models.PurchaseItem
		err := db.Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
			Where("purchase_items.product_id = ?", product.ID).
			Where("purchases.status IN ?", []string{"APPROVED", "COMPLETED", "PAID"}).
			Order("purchases.date DESC").
			First(&lastPurchaseItem).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("   ⚠️  Product #%d '%s': No purchase history found", 
					product.ID, product.Name)
				notFoundCount++
			} else {
				log.Printf("   ❌ Product #%d: Database error: %v", product.ID, err)
			}
			continue
		}

		// Calculate unit cost from purchase
		unitCost := lastPurchaseItem.UnitPrice
		product.CostPrice = unitCost

		if err := db.Save(&product).Error; err != nil {
			log.Printf("   ❌ Failed to update product #%d: %v", product.ID, err)
		} else {
			log.Printf("   ✅ Product #%d '%s': Cost = Rp %.2f (from last purchase)",
				product.ID, truncateString(product.Name, 25), unitCost)
			successCount++
		}
	}

	log.Printf("\n✅ Successfully updated: %d", successCount)
	log.Printf("⚠️  No purchase history: %d", notFoundCount)

	if notFoundCount > 0 {
		log.Println("\n💡 Tip: For products without purchase history, use method 1 or 3")
	}
}

// fixManually allows manual entry for each product
func fixManually(db *gorm.DB, products []models.Product) {
	log.Println("\n💡 Manual cost price entry")
	log.Println("   Enter '0' to skip a product, 'q' to quit")
	
	successCount := 0

	for i, product := range products {
		log.Printf("\n[%d/%d] Product: %s", i+1, len(products), product.Name)
		log.Printf("   Current Price: Rp %.2f", product.SalePrice)
		log.Printf("   Current Cost:  Rp %.2f", product.CostPrice)
		
		fmt.Print("   New Cost Price (or 0 to skip, q to quit): Rp ")
		var input string
		fmt.Scanln(&input)

		if input == "q" || input == "Q" {
			log.Println("   ⏹️  Stopping manual entry")
			break
		}

		var newCost float64
		if _, err := fmt.Sscanf(input, "%f", &newCost); err != nil {
			log.Printf("   ⚠️  Invalid input, skipping")
			continue
		}

		if newCost <= 0 {
			log.Printf("   ⏭️  Skipped")
			continue
		}

		product.CostPrice = newCost
		if err := db.Save(&product).Error; err != nil {
			log.Printf("   ❌ Failed: %v", err)
		} else {
			log.Printf("   ✅ Updated: Rp %.2f", newCost)
			successCount++
		}
	}

	log.Printf("\n✅ Successfully updated %d products", successCount)
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
