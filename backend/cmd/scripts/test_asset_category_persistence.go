package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"time"

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

	fmt.Println("=== ASSET CATEGORY PERSISTENCE TEST ===")

	// Step 1: Get initial count
	var initialCount int64
	db.Model(&models.AssetCategory{}).Where("is_active = ?", true).Count(&initialCount)
	fmt.Printf("Initial active categories count: %d\n", initialCount)

	// Step 2: Create test category (similar to what user did in frontend)
	// Generate unique code with timestamp
	timestamp := time.Now().Format("060102150405")
	testCategory := models.AssetCategory{
		Code:        "TEST" + timestamp, 
		Name:        "test01", // Exact same name as user input
		Description: "Testing category persistence - simulating frontend input",
		IsActive:    true,
	}

	fmt.Println("\n=== CREATING CATEGORY ===")
	fmt.Printf("Creating category: Code='%s', Name='%s'\n", testCategory.Code, testCategory.Name)

	err = db.Create(&testCategory).Error
	if err != nil {
		log.Printf("❌ Failed to create category: %v", err)
		return
	}

	fmt.Printf("✅ Category created successfully with ID: %d\n", testCategory.ID)

	// Step 3: Verify creation immediately
	var createdCategory models.AssetCategory
	err = db.Where("id = ?", testCategory.ID).First(&createdCategory).Error
	if err != nil {
		fmt.Printf("❌ Failed to find created category: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Category found immediately after creation: ID=%d, Name='%s'\n", 
		createdCategory.ID, createdCategory.Name)

	// Step 4: Simulate refresh by getting all categories (like frontend GET request)
	fmt.Println("\n=== SIMULATING BROWSER REFRESH (GET ALL CATEGORIES) ===")
	var categories []models.AssetCategory
	err = db.Where("is_active = ?", true).Order("name ASC").Find(&categories).Error
	if err != nil {
		fmt.Printf("❌ Failed to fetch categories: %v\n", err)
		return
	}

	fmt.Printf("Retrieved %d active categories:\n", len(categories))
	foundTestCategory := false
	for _, cat := range categories {
		fmt.Printf("- ID: %d, Code: %s, Name: %s\n", cat.ID, cat.Code, cat.Name)
		if cat.Name == "test01" {
			foundTestCategory = true
		}
	}

	if foundTestCategory {
		fmt.Println("✅ Test category 'test01' FOUND after refresh - PERSISTENT!")
	} else {
		fmt.Println("❌ Test category 'test01' NOT FOUND after refresh - BUG CONFIRMED!")
	}

	// Step 5: Check database transaction state
	fmt.Println("\n=== DATABASE TRANSACTION VERIFICATION ===")
	var finalCount int64
	db.Model(&models.AssetCategory{}).Where("is_active = ?", true).Count(&finalCount)
	fmt.Printf("Final active categories count: %d\n", finalCount)
	fmt.Printf("Expected count increase: %d\n", finalCount - initialCount)

	// Step 6: Direct SQL query to verify (bypassing GORM)
	fmt.Println("\n=== RAW SQL VERIFICATION ===")
	var rawCount int
	err = db.Raw("SELECT COUNT(*) FROM asset_categories WHERE name = ? AND is_active = true", "test01").Scan(&rawCount).Error
	if err != nil {
		fmt.Printf("❌ Raw SQL query failed: %v\n", err)
	} else {
		fmt.Printf("Raw SQL count for 'test01': %d\n", rawCount)
	}

	// Step 7: Check for soft deletes
	fmt.Println("\n=== SOFT DELETE CHECK ===")
	var withDeleted []models.AssetCategory
	err = db.Unscoped().Where("name = ?", "test01").Find(&withDeleted).Error
	if err != nil {
		fmt.Printf("❌ Soft delete check failed: %v\n", err)
	} else {
		fmt.Printf("Categories with name 'test01' (including soft-deleted): %d\n", len(withDeleted))
		for _, cat := range withDeleted {
			var deletedStatus string
			if cat.DeletedAt.Valid {
				deletedStatus = "SOFT DELETED at " + cat.DeletedAt.Time.Format("2006-01-02 15:04:05")
			} else {
				deletedStatus = "ACTIVE"
			}
			fmt.Printf("- ID: %d, Status: %s\n", cat.ID, deletedStatus)
		}
	}

	// Step 8: Wait and test again (simulating time delay like user experienced)
	fmt.Println("\n=== WAITING AND TESTING AGAIN ===")
	time.Sleep(2 * time.Second)

	var categoriesAfterWait []models.AssetCategory
	err = db.Where("is_active = ?", true).Order("name ASC").Find(&categoriesAfterWait).Error
	if err != nil {
		fmt.Printf("❌ Failed to fetch categories after wait: %v\n", err)
		return
	}

	foundAfterWait := false
	for _, cat := range categoriesAfterWait {
		if cat.Name == "test01" {
			foundAfterWait = true
			break
		}
	}

	if foundAfterWait {
		fmt.Println("✅ Test category 'test01' STILL FOUND after wait - PERSISTENT!")
	} else {
		fmt.Println("❌ Test category 'test01' DISAPPEARED after wait - BUG CONFIRMED!")
	}

	// Clean up - Delete the test category
	fmt.Println("\n=== CLEANUP ===")
	err = db.Delete(&testCategory).Error
	if err != nil {
		fmt.Printf("❌ Failed to delete test category: %v\n", err)
	} else {
		fmt.Println("✅ Test category deleted successfully")
	}

	fmt.Println("\n=== TEST COMPLETE ===")
}