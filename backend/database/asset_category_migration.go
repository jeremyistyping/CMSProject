package database

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// AssetCategoryMigration creates default asset categories and fixes existing asset data
func AssetCategoryMigration(db *gorm.DB) {
	log.Println("🔧 Starting Asset Category Migration...")

	migrationID := "asset_category_v1.0"
	
	// Check if this migration has already been run
	var existingMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", migrationID).First(&existingMigration).Error; err == nil {
		log.Printf("✅ Asset Category Migration already applied at %v", existingMigration.AppliedAt)
		return
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("❌ Failed to start asset category migration transaction: %v", tx.Error)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Asset category migration rolled back due to panic: %v", r)
		}
	}()

	var fixesApplied []string

	// Step 1: Create default asset categories
	if err := createDefaultAssetCategories(tx); err == nil {
		fixesApplied = append(fixesApplied, "Default asset categories created")
	}

	// Step 2: Migrate existing assets to use category relations
	if err := migrateExistingAssetCategories(tx); err == nil {
		fixesApplied = append(fixesApplied, "Existing assets migrated to category relations")
	}

	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: fmt.Sprintf("Asset category migration applied: %v", fixesApplied),
		Version:     "1.0",
		AppliedAt:   time.Now(),
	}

	if err := tx.Create(&migrationRecord).Error; err != nil {
		log.Printf("❌ Failed to record asset category migration: %v", err)
		tx.Rollback()
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit asset category migration: %v", err)
		return
	}

	log.Printf("✅ Asset Category Migration completed successfully. Applied fixes: %v", fixesApplied)
}

// createDefaultAssetCategories creates the default asset categories
func createDefaultAssetCategories(tx *gorm.DB) error {
	log.Println("  🔧 Creating default asset categories...")

	defaultCategories := []models.AssetCategory{
		{Code: "FIXED", Name: "Fixed Asset", Description: "Long-term tangible assets used in business operations", IsActive: true},
		{Code: "REAL", Name: "Real Estate", Description: "Land, buildings, and property assets", IsActive: true},
		{Code: "COMP", Name: "Computer Equipment", Description: "Computers, servers, and IT hardware", IsActive: true},
		{Code: "VEHICLE", Name: "Vehicle", Description: "Cars, trucks, and transportation assets", IsActive: true},
		{Code: "OFFICE", Name: "Office Equipment", Description: "Office furniture, equipment, and supplies", IsActive: true},
		{Code: "FURNITURE", Name: "Furniture", Description: "Furniture and fixtures", IsActive: true},
		{Code: "IT", Name: "IT Infrastructure", Description: "Network equipment, software licenses", IsActive: true},
		{Code: "MACHINE", Name: "Machinery", Description: "Manufacturing and industrial machinery", IsActive: true},
		{Code: "LAND", Name: "Tanah", Description: "Tanah dan lahan", IsActive: true},
	}

	createdCount := 0
	for _, category := range defaultCategories {
		// Check if category already exists
		var existingCategory models.AssetCategory
		if err := tx.Where("code = ? OR name = ?", category.Code, category.Name).First(&existingCategory).Error; err != nil {
			// Category doesn't exist, create it
			if err := tx.Create(&category).Error; err != nil {
				log.Printf("    ❌ Failed to create category %s: %v", category.Name, err)
			} else {
				log.Printf("    ✅ Created asset category: %s (%s)", category.Name, category.Code)
				createdCount++
			}
		} else {
			log.Printf("    ⚠️  Category already exists: %s", category.Name)
		}
	}

	log.Printf("    ✅ Created %d new asset categories", createdCount)
	return nil
}

// migrateExistingAssetCategories migrates existing assets to use the new category relations
func migrateExistingAssetCategories(tx *gorm.DB) error {
	log.Println("  🔧 Migrating existing assets to category relations...")

	// Get all assets that have category string values but no CategoryID
	var assets []models.Asset
	err := tx.Where("category IS NOT NULL AND category != '' AND category_id IS NULL").Find(&assets).Error
	if err != nil {
		log.Printf("    ❌ Error finding assets to migrate: %v", err)
		return err
	}

	if len(assets) == 0 {
		log.Println("    ✅ No assets need category migration")
		return nil
	}

	log.Printf("    📊 Found %d assets to migrate", len(assets))

	migratedCount := 0
	for _, asset := range assets {
		// Try to find matching category
		var category models.AssetCategory
		
		// First try exact name match
		if err := tx.Where("name = ? AND is_active = ?", asset.Category, true).First(&category).Error; err != nil {
			// Try case-insensitive match
			if err := tx.Where("LOWER(name) = LOWER(?) AND is_active = ?", asset.Category, true).First(&category).Error; err != nil {
				// Try to match with common variations
				categoryName := mapLegacyCategoryName(asset.Category)
				if err := tx.Where("LOWER(name) = LOWER(?) AND is_active = ?", categoryName, true).First(&category).Error; err != nil {
					// Create new category for unmatched ones
					newCategory := models.AssetCategory{
						Code:        generateCategoryCode(asset.Category),
						Name:        asset.Category,
						Description: fmt.Sprintf("Migrated from legacy asset: %s", asset.Category),
						IsActive:    true,
					}
					if err := tx.Create(&newCategory).Error; err != nil {
						log.Printf("    ❌ Failed to create category for asset %d: %v", asset.ID, err)
						continue
					}
					category = newCategory
					log.Printf("    🆕 Created new category: %s for asset %d", asset.Category, asset.ID)
				}
			}
		}

		// Update asset with category ID
		if err := tx.Model(&asset).Update("category_id", category.ID).Error; err != nil {
			log.Printf("    ❌ Failed to update asset %d with category ID: %v", asset.ID, err)
		} else {
			log.Printf("    ✅ Migrated asset %d (%s) to category: %s", asset.ID, asset.Name, category.Name)
			migratedCount++
		}
	}

	log.Printf("    ✅ Successfully migrated %d assets to category relations", migratedCount)
	return nil
}

// mapLegacyCategoryName maps legacy category names to standard names
func mapLegacyCategoryName(legacyName string) string {
	categoryMap := map[string]string{
		"tanah":           "Tanah",
		"land":            "Tanah", 
		"fixed asset":     "Fixed Asset",
		"real estate":     "Real Estate",
		"computer":        "Computer Equipment",
		"vehicle":         "Vehicle",
		"office":          "Office Equipment",
		"furniture":       "Furniture",
		"it":              "IT Infrastructure",
		"machinery":       "Machinery",
	}
	
	if mapped, exists := categoryMap[legacyName]; exists {
		return mapped
	}
	return legacyName
}

// generateCategoryCode generates a unique category code
func generateCategoryCode(name string) string {
	// Take first 3-6 characters and make uppercase
	code := name
	if len(code) > 6 {
		code = code[:6]
	}
	if len(code) < 3 {
		code = code + "CAT"
	}
	
	// Remove spaces and special characters
	result := ""
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result += string(r)
		}
	}
	
	return result
}