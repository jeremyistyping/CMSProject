package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	// Check if asset_categories table exists
	var tableExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'asset_categories')").Scan(&tableExists).Error
	if err != nil {
		log.Fatal("Error checking table existence:", err)
	}

	fmt.Println("\n=== ASSET CATEGORIES TABLE CHECK ===")
	if tableExists {
		fmt.Println("✅ asset_categories table exists")

		// Get table structure
		fmt.Println("\n=== TABLE COLUMNS ===")
		var columns []struct {
			ColumnName string
			DataType   string
		}
		err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'asset_categories' ORDER BY ordinal_position").Scan(&columns).Error
		if err == nil {
			for _, col := range columns {
				fmt.Printf("- %s: %s\n", col.ColumnName, col.DataType)
			}
		}

		// Check data
		var categories []models.AssetCategory
		err = db.Find(&categories).Error
		if err != nil {
			log.Printf("Error fetching asset categories: %v", err)
		} else {
			fmt.Printf("\n=== DATA COUNT ===\n")
			fmt.Printf("Found %d asset categories\n", len(categories))
			
			if len(categories) > 0 {
				fmt.Println("\n=== SAMPLE DATA ===")
				for _, cat := range categories {
					fmt.Printf("ID: %d, Code: %s, Name: %s, Active: %t\n", cat.ID, cat.Code, cat.Name, cat.IsActive)
				}
			}
		}

		// Test create functionality
		fmt.Println("\n=== TEST CREATE CATEGORY ===")
		testCategory := models.AssetCategory{
			Code:        "TEST01",
			Name:        "Test Category",
			Description: "Test category for debugging",
			IsActive:    true,
		}

		// Try to create
		if err := db.Create(&testCategory).Error; err != nil {
			fmt.Printf("❌ Failed to create test category: %v\n", err)
		} else {
			fmt.Printf("✅ Test category created with ID: %d\n", testCategory.ID)

			// Clean up - delete the test category
			db.Delete(&testCategory)
			fmt.Println("🧹 Test category cleaned up")
		}

	} else {
		fmt.Println("❌ asset_categories table does not exist")
		
		// Try to create the table
		fmt.Println("\n=== ATTEMPTING TO CREATE TABLE ===")
		err := db.AutoMigrate(&models.AssetCategory{})
		if err != nil {
			fmt.Printf("❌ Failed to create asset_categories table: %v\n", err)
		} else {
			fmt.Println("✅ asset_categories table created successfully")
		}
	}

	fmt.Println("\n=== CHECK COMPLETE ===")
}