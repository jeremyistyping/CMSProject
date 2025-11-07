package migrations

import (
	"fmt"

	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

// MigrateWarehouseLocations adds warehouse location support to the database
func MigrateWarehouseLocations(db *gorm.DB) error {

	fmt.Println("Starting warehouse locations migration...")

	// Create warehouse locations table
	err := db.AutoMigrate(&models.WarehouseLocation{})
	if err != nil {
		return fmt.Errorf("failed to create warehouse_locations table: %v", err)
	}
	fmt.Println("✓ Created warehouse_locations table")

	// Add warehouse_location_id column to products table if it doesn't exist
	if !db.Migrator().HasColumn(&models.Product{}, "warehouse_location_id") {
		err = db.Migrator().AddColumn(&models.Product{}, "warehouse_location_id")
		if err != nil {
			return fmt.Errorf("failed to add warehouse_location_id column to products table: %v", err)
		}
		fmt.Println("✓ Added warehouse_location_id column to products table")
	} else {
		fmt.Println("✓ warehouse_location_id column already exists in products table")
	}

	// Create default warehouse locations
	if err := createDefaultWarehouseLocations(db); err != nil {
		return fmt.Errorf("failed to create default warehouse locations: %v", err)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}

func createDefaultWarehouseLocations(db *gorm.DB) error {
	defaultLocations := []models.WarehouseLocation{
		{
			Code:        "WH-001",
			Name:        "Main Warehouse",
			Description: "Primary storage facility",
			Address:     "Jl. Gudang Utama No. 1",
			IsActive:    true,
		},
		{
			Code:        "WH-002",
			Name:        "Storage Room A",
			Description: "Small items storage",
			Address:     "Jl. Gudang Utama No. 2",
			IsActive:    true,
		},
		{
			Code:        "WH-003",
			Name:        "Cold Storage",
			Description: "Temperature controlled storage",
			Address:     "Jl. Gudang Utama No. 3",
			IsActive:    true,
		},
	}

	for _, location := range defaultLocations {
		// Check if location already exists
		var existing models.WarehouseLocation
		if err := db.Where("code = ?", location.Code).First(&existing).Error; err == gorm.ErrRecordNotFound {
			// Create new location
			if err := db.Create(&location).Error; err != nil {
				return fmt.Errorf("failed to create warehouse location %s: %v", location.Code, err)
			} else {
				fmt.Printf("✓ Created warehouse location: %s - %s\n", location.Code, location.Name)
			}
		} else {
			fmt.Printf("✓ Warehouse location already exists: %s - %s\n", location.Code, location.Name)
		}
	}
	return nil
}
