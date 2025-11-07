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

	fmt.Println("=== MANUAL API TEST SIMULATION ===")
	fmt.Println("This simulates what the frontend should be doing:")

	// Simulate Step 1: Frontend calls GET /api/v1/assets/categories
	fmt.Println("\n1. 📄 GET /api/v1/assets/categories (Before Create)")
	var categoriesBefore []models.AssetCategory
	err = db.Where("is_active = ?", true).Order("name ASC").Find(&categoriesBefore).Error
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("   Response: %d categories found\n", len(categoriesBefore))
	for _, cat := range categoriesBefore {
		fmt.Printf("   - %s (%s)\n", cat.Name, cat.Code)
	}

	// Simulate Step 2: Frontend calls POST /api/v1/assets/categories
	fmt.Println("\n2. 📝 POST /api/v1/assets/categories")
	timestamp := time.Now().Format("150405")
	newCategory := models.AssetCategory{
		Code:        "TEST" + timestamp,
		Name:        "test01",
		Description: "Test category from manual simulation",
		IsActive:    true,
	}

	err = db.Create(&newCategory).Error
	if err != nil {
		fmt.Printf("   ❌ CREATE FAILED: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Response: Category created with ID %d\n", newCategory.ID)

	// Simulate Step 3: Frontend immediately calls GET again (refresh)
	fmt.Println("\n3. 🔄 GET /api/v1/assets/categories (After Create - Immediate)")
	var categoriesAfter []models.AssetCategory
	err = db.Where("is_active = ?", true).Order("name ASC").Find(&categoriesAfter).Error
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("   Response: %d categories found\n", len(categoriesAfter))
	
	foundNewCategory := false
	for _, cat := range categoriesAfter {
		fmt.Printf("   - %s (%s)\n", cat.Name, cat.Code)
		if cat.ID == newCategory.ID {
			foundNewCategory = true
		}
	}

	if foundNewCategory {
		fmt.Println("   ✅ NEW CATEGORY FOUND - API Working Correctly!")
	} else {
		fmt.Println("   ❌ NEW CATEGORY NOT FOUND - API Bug!")
	}

	// Simulate Step 4: User refreshes browser (new request)
	fmt.Println("\n4. 🌐 Browser Refresh (New Connection)")
	time.Sleep(1 * time.Second)
	
	var categoriesRefresh []models.AssetCategory
	err = db.Where("is_active = ?", true).Order("name ASC").Find(&categoriesRefresh).Error
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("   Response: %d categories found\n", len(categoriesRefresh))
	
	foundAfterRefresh := false
	for _, cat := range categoriesRefresh {
		if cat.ID == newCategory.ID {
			foundAfterRefresh = true
		}
	}

	if foundAfterRefresh {
		fmt.Println("   ✅ CATEGORY STILL FOUND AFTER REFRESH - Persistent!")
	} else {
		fmt.Println("   ❌ CATEGORY MISSING AFTER REFRESH - Data Loss!")
	}

	// Cleanup
	fmt.Println("\n5. 🧹 Cleanup")
	db.Delete(&newCategory)
	fmt.Println("   ✅ Test category removed")

	fmt.Println("\n=== CONCLUSION ===")
	if foundNewCategory && foundAfterRefresh {
		fmt.Println("✅ BACKEND API IS WORKING CORRECTLY")
		fmt.Println("📋 The issue is likely in the FRONTEND:")
		fmt.Println("   - Check browser cache")
		fmt.Println("   - Check frontend state management") 
		fmt.Println("   - Check API call error handling")
		fmt.Println("   - Check frontend refresh logic")
	} else {
		fmt.Println("❌ BACKEND API HAS ISSUES")
		fmt.Println("📋 Backend needs investigation")
	}
}