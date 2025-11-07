package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	ID        uint    `gorm:"column:id"`
	Name      string  `gorm:"column:name"`
	SKU       string  `gorm:"column:sku"`
	CostPrice float64 `gorm:"column:cost_price"`
}

type PurchaseItem struct {
	ProductID  uint    `gorm:"column:product_id"`
	UnitPrice  float64 `gorm:"column:unit_price"`
	PurchaseID uint    `gorm:"column:purchase_id"`
}

type Purchase struct {
	ID     uint   `gorm:"column:id"`
	Status string `gorm:"column:status"`
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("UPDATE PRODUCT COST PRICE FROM PURCHASE TRANSACTIONS")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Get all products
	var products []Product
	db.Table("products").Where("cost_price = 0 OR cost_price IS NULL").Order("id").Find(&products)

	if len(products) == 0 {
		fmt.Println("✅ All products already have cost price set!")
		return
	}

	fmt.Printf("Found %d products without cost price:\n\n", len(products))

	updatedCount := 0
	for _, product := range products {
		// Get latest purchase item for this product from RECEIVED/PAID purchases
		var purchaseItem PurchaseItem
		err := db.Table("purchase_items").
			Select("purchase_items.*").
			Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
			Where("purchase_items.product_id = ?", product.ID).
			Where("purchases.status IN ('RECEIVED', 'PAID')").
			Order("purchases.id DESC").
			First(&purchaseItem).Error

		if err == nil && purchaseItem.UnitPrice > 0 {
			// Update product cost price
			db.Table("products").
				Where("id = ?", product.ID).
				Update("cost_price", purchaseItem.UnitPrice)

			fmt.Printf("✅ Product #%d (%s): Cost Price updated to Rp %.2f\n", 
				product.ID, product.Name, purchaseItem.UnitPrice)
			updatedCount++
		} else {
			fmt.Printf("⚠️  Product #%d (%s): No purchase found, using default Rp 1000.00\n", 
				product.ID, product.Name)
			
			// Set default cost price (can be adjusted later)
			db.Table("products").
				Where("id = ?", product.ID).
				Update("cost_price", 1000.0)
			updatedCount++
		}
	}

	fmt.Println()
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Printf("✅ Updated %d products\n", updatedCount)
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Verify product cost prices in the UI")
	fmt.Println("  2. Run: go run cmd/scripts/backfill_cogs_entries.go -start=2025-01-01 -end=2025-12-31 -dry-run")
	fmt.Println("  3. If preview looks good, run without -dry-run flag")
	fmt.Println()
}

