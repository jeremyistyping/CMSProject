package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	log.Println("Database connected successfully")
	fmt.Println("=== TEST ASSET CREATION WITH RETRY LOGIC ===\n")

	// Create asset service
	assetRepo := repositories.NewAssetRepository(db)
	assetService := services.NewAssetService(assetRepo, db)

	// Test creating multiple assets with the same category in sequence
	fmt.Println("1. Testing sequential asset creation (Vehicle category):")
	
	for i := 1; i <= 3; i++ {
		asset := &models.Asset{
			Name:               fmt.Sprintf("Test Vehicle %d", i),
			Category:           "Vehicle",
			Status:             models.AssetStatusActive,
			PurchaseDate:       time.Now(),
			PurchasePrice:      50000.00,
			SalvageValue:       5000.00,
			UsefulLife:         5,
			DepreciationMethod: models.DepreciationMethodStraightLine,
			IsActive:           true,
			Notes:              fmt.Sprintf("Test asset %d for duplicate checking", i),
			Condition:          "Good",
		}

		err := assetService.CreateAsset(asset)
		if err != nil {
			fmt.Printf("   ❌ Asset %d creation failed: %v\n", i, err)
		} else {
			fmt.Printf("   ✅ Asset %d created successfully with code: %s\n", i, asset.Code)
		}
	}

	// Test concurrent-like creation (different categories)
	fmt.Println("\n2. Testing creation with different categories:")
	
	categories := []string{"Computer Equipment", "Office Equipment", "Machinery"}
	for i, category := range categories {
		asset := &models.Asset{
			Name:               fmt.Sprintf("Test %s Asset", category),
			Category:           category,
			Status:             models.AssetStatusActive,
			PurchaseDate:       time.Now(),
			PurchasePrice:      25000.00,
			SalvageValue:       2500.00,
			UsefulLife:         3,
			DepreciationMethod: models.DepreciationMethodStraightLine,
			IsActive:           true,
			Notes:              fmt.Sprintf("Test asset for category %s", category),
			Condition:          "Good",
		}

		err := assetService.CreateAsset(asset)
		if err != nil {
			fmt.Printf("   ❌ %s asset creation failed: %v\n", category, err)
		} else {
			fmt.Printf("   ✅ %s asset created successfully with code: %s\n", category, asset.Code)
		}
	}

	// Test creating an asset with explicit code (should fail if duplicate)
	fmt.Println("\n3. Testing explicit code assignment (should detect existing):")
	
	asset := &models.Asset{
		Code:               "VH-2025-001", // This should already exist from the first test
		Name:               "Duplicate Code Test Vehicle",
		Category:           "Vehicle", 
		Status:             models.AssetStatusActive,
		PurchaseDate:       time.Now(),
		PurchasePrice:      60000.00,
		SalvageValue:       6000.00,
		UsefulLife:         5,
		DepreciationMethod: models.DepreciationMethodStraightLine,
		IsActive:           true,
		Notes:              "This should fail due to duplicate code",
		Condition:          "Good",
	}

	err := assetService.CreateAsset(asset)
	if err != nil {
		fmt.Printf("   ✅ Expected failure detected: %v\n", err)
	} else {
		fmt.Printf("   ❌ Unexpected success - duplicate should have been detected!\n")
	}

	fmt.Println("\n=== TEST COMPLETE ===")
}