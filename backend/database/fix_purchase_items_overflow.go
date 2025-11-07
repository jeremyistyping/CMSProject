package database

import (
	"log"
	"gorm.io/gorm"
)

// FixPurchaseItemsFieldOverflow updates purchase_items decimal fields to prevent numeric overflow
func FixPurchaseItemsFieldOverflow(db *gorm.DB) {
	log.Println("🔧 Fixing purchase_items decimal field overflow...")

	// Update discount and tax fields from decimal(8,2) to decimal(15,2)
	queries := []string{
		`ALTER TABLE purchase_items ALTER COLUMN discount TYPE decimal(15,2)`,
		`ALTER TABLE purchase_items ALTER COLUMN tax TYPE decimal(15,2)`,
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to execute migration query: %v", err)
			log.Printf("Query: %s", query)
			// Continue with other queries even if one fails
		}
	}

	log.Println("✅ Purchase items field overflow fix completed")
}
