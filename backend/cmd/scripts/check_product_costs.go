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
	SellPrice float64 `gorm:"column:sell_price"`
}

type SaleItem struct {
	ID        uint    `gorm:"column:id"`
	SaleID    uint    `gorm:"column:sale_id"`
	ProductID uint    `gorm:"column:product_id"`
	Quantity  float64 `gorm:"column:quantity"`
}

type Sale struct {
	ID            uint   `gorm:"column:id"`
	InvoiceNumber string `gorm:"column:invoice_number"`
	Status        string `gorm:"column:status"`
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
	fmt.Println("PRODUCT COST PRICE CHECK")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Check products
	var products []Product
	db.Table("products").Order("id").Find(&products)

	fmt.Printf("Found %d products:\n\n", len(products))
	fmt.Printf("%-5s %-30s %-15s %15s %15s\n", "ID", "Name", "SKU", "Cost Price", "Sell Price")
	fmt.Println(string(make([]byte, 80)))

	for _, p := range products {
		fmt.Printf("%-5d %-30s %-15s %15.2f %15.2f\n", p.ID, p.Name, p.SKU, p.CostPrice, p.SellPrice)
	}

	fmt.Println()
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("SALES & SALE ITEMS DETAIL")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Check sales
	var sales []Sale
	db.Table("sales").Where("status IN ('INVOICED', 'PAID')").Order("id").Find(&sales)

	for _, sale := range sales {
		fmt.Printf("Sale #%d - %s (Status: %s)\n", sale.ID, sale.InvoiceNumber, sale.Status)
		
		var items []SaleItem
		db.Table("sale_items").Where("sale_id = ?", sale.ID).Find(&items)
		
		totalEstCOGS := 0.0
		for _, item := range items {
			var product Product
			db.Table("products").Where("id = ?", item.ProductID).First(&product)
			
			estCOGS := item.Quantity * product.CostPrice
			totalEstCOGS += estCOGS
			
			fmt.Printf("  - Item #%d: Product #%d (%s) x %.2f = COGS: Rp %.2f\n", 
				item.ID, product.ID, product.Name, item.Quantity, estCOGS)
		}
		
		fmt.Printf("  Total Est. COGS: Rp %.2f\n\n", totalEstCOGS)
	}
}

