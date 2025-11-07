package main

import (
	"fmt"
	"log"
	"os"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Product model for checking prices
type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Code        string  `gorm:"unique;not null;size:20"`
	Name        string  `gorm:"not null;size:100"`
	CostPrice   float64 `gorm:"type:decimal(15,2);default:0"`
	SalePrice   float64 `gorm:"type:decimal(15,2);default:0"`
	Stock       int     `gorm:"default:0"`
}

func main() {
	// Get database connection from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "accounting_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("========================================")
	fmt.Println("🔍 Checking for products with unreasonable cost prices...")
	fmt.Println("========================================\n")

	// Find products with cost price > 10 Billion or negative
	var problematicProducts []Product
	const maxReasonablePrice = 10000000000.0

	if err := db.Where("cost_price > ? OR cost_price < 0", maxReasonablePrice).Find(&problematicProducts).Error; err != nil {
		log.Fatalf("❌ Failed to query products: %v", err)
	}

	if len(problematicProducts) == 0 {
		fmt.Println("✅ No products found with unreasonable cost prices!")
		fmt.Println("   All products have valid cost prices (0 to 10 Billion)")
		return
	}

	fmt.Printf("⚠️  Found %d products with unreasonable cost prices:\n\n", len(problematicProducts))

	for i, product := range problematicProducts {
		fmt.Printf("%d. Product: %s\n", i+1, product.Name)
		fmt.Printf("   Code: %s\n", product.Code)
		fmt.Printf("   ❌ Cost Price: Rp %.2f\n", product.CostPrice)
		fmt.Printf("   Sale Price: Rp %.2f\n", product.SalePrice)
		fmt.Printf("   Stock: %d\n", product.Stock)
		
		// Suggest reasonable price
		suggestedPrice := product.SalePrice * 0.7 // 70% of sale price as rough estimate
		if suggestedPrice == 0 {
			suggestedPrice = 100000 // Default to 100k if sale price is 0
		}
		fmt.Printf("   💡 Suggested Cost Price: Rp %.2f (70%% of sale price)\n", suggestedPrice)
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Println("🔧 To fix these products:")
	fmt.Println("========================================")
	fmt.Println("Option 1: Via UI")
	fmt.Println("  - Go to Products page")
	fmt.Println("  - Edit each product above")
	fmt.Println("  - Set correct cost_price")
	fmt.Println()
	fmt.Println("Option 2: Via SQL (example for Mouse Logitech)")
	fmt.Println("  UPDATE products SET cost_price = 500000 WHERE name LIKE '%Mouse%Logitech%';")
	fmt.Println()
	fmt.Println("Option 3: Auto-fix with suggested prices (Type 'yes' to proceed)")
	
	// Ask for confirmation
	var response string
	fmt.Print("\nAuto-fix all products with suggested prices? (yes/no): ")
	fmt.Scanln(&response)
	
	if response == "yes" || response == "y" {
		fmt.Println("\n🔧 Auto-fixing products...")
		
		for _, product := range problematicProducts {
			// Calculate suggested price
			suggestedPrice := product.SalePrice * 0.7
			if suggestedPrice == 0 || suggestedPrice > maxReasonablePrice {
				suggestedPrice = 100000 // Default to 100k
			}
			
			// Update cost price
			if err := db.Model(&Product{}).Where("id = ?", product.ID).Update("cost_price", suggestedPrice).Error; err != nil {
				fmt.Printf("❌ Failed to update product %s: %v\n", product.Name, err)
			} else {
				fmt.Printf("✅ Updated %s: %.2f → %.2f\n", product.Name, product.CostPrice, suggestedPrice)
			}
		}
		
		fmt.Println("\n✅ All products updated!")
	} else {
		fmt.Println("\n⚠️  No changes made. Please fix products manually.")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

