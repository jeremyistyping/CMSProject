package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	log.Println("Database connected successfully")
	fmt.Println("=== DEBUG ASSETS CODES ===\n")

	// 1. Check all assets with their codes
	var assets []models.Asset
	result := db.Unscoped().Find(&assets) // Include soft deleted records
	
	if result.Error != nil {
		log.Printf("Error querying assets: %v", result.Error)
		return
	}

	fmt.Printf("1. ALL ASSETS (including soft deleted):\n")
	fmt.Printf("ID\tCode\t\tName\t\t\tCategory\tDeleted At\n")
	fmt.Printf("--\t----\t\t----\t\t\t--------\t----------\n")
	
	for _, asset := range assets {
		deletedAt := "NULL"
		if asset.DeletedAt.Valid {
			deletedAt = asset.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%d\t%s\t\t%s\t\t%s\t%s\n", asset.ID, asset.Code, asset.Name, asset.Category, deletedAt)
	}

	// 2. Check for specific code VH-2025-001
	fmt.Printf("\n2. CHECKING FOR VH-2025-001 SPECIFICALLY:\n")
	var specificAsset models.Asset
	
	// Check with soft deleted
	err := db.Unscoped().Where("code = ?", "VH-2025-001").First(&specificAsset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("   Code VH-2025-001 NOT FOUND in database (including soft deleted)")
		} else {
			fmt.Printf("   Error checking for VH-2025-001: %v\n", err)
		}
	} else {
		deletedAt := "NULL (Active)"
		if specificAsset.DeletedAt.Valid {
			deletedAt = "SOFT DELETED: " + specificAsset.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("   FOUND VH-2025-001: ID=%d, Name=%s, Status=%s\n", specificAsset.ID, specificAsset.Name, deletedAt)
	}

	// 3. Check Vehicle category assets
	fmt.Printf("\n3. ALL VEHICLE CATEGORY ASSETS:\n")
	var vehicleAssets []models.Asset
	db.Unscoped().Where("category = ?", "Vehicle").Find(&vehicleAssets)
	
	if len(vehicleAssets) == 0 {
		fmt.Println("   No Vehicle category assets found")
	} else {
		fmt.Printf("   Found %d Vehicle assets:\n", len(vehicleAssets))
		for _, asset := range vehicleAssets {
			deletedAt := "Active"
			if asset.DeletedAt.Valid {
				deletedAt = "Deleted"
			}
			fmt.Printf("   - %s: %s (%s)\n", asset.Code, asset.Name, deletedAt)
		}
	}

	// 4. Check database constraints
	fmt.Printf("\n4. DATABASE CONSTRAINTS ON ASSETS TABLE:\n")
	var constraints []struct {
		ConstraintName string `gorm:"column:conname"`
		ConstraintType string `gorm:"column:contype"`
	}

	query := `
		SELECT conname, contype 
		FROM pg_constraint 
		WHERE conrelid = 'assets'::regclass 
		AND contype IN ('u', 'p')
	`
	
	db.Raw(query).Scan(&constraints)
	
	if len(constraints) > 0 {
		fmt.Println("   Constraints found:")
		for _, c := range constraints {
			constraintType := "Unknown"
			switch c.ConstraintType {
			case "p":
				constraintType = "Primary Key"
			case "u":
				constraintType = "Unique"
			}
			fmt.Printf("   - %s (%s)\n", c.ConstraintName, constraintType)
		}
	} else {
		fmt.Println("   No constraints info available or query failed")
	}

	// 5. Try to check index information
	fmt.Printf("\n5. UNIQUE INDEXES ON ASSETS:\n")
	var indexes []struct {
		IndexName string `gorm:"column:indexname"`
		IndexDef  string `gorm:"column:indexdef"`
	}

	indexQuery := `
		SELECT indexname, indexdef 
		FROM pg_indexes 
		WHERE tablename = 'assets' 
		AND indexdef LIKE '%UNIQUE%'
	`
	
	db.Raw(indexQuery).Scan(&indexes)
	
	if len(indexes) > 0 {
		fmt.Println("   Unique indexes found:")
		for _, idx := range indexes {
			fmt.Printf("   - %s: %s\n", idx.IndexName, idx.IndexDef)
		}
	} else {
		fmt.Println("   No unique indexes found or query failed")
	}

	fmt.Println("\n=== DEBUG COMPLETE ===")
}
