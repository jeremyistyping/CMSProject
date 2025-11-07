package main

import (
	"encoding/json"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== TESTING ZERO VALUES UPDATE FIX ===")

	// Get first product
	var product models.Product
	result := db.First(&product)
	if result.Error != nil {
		fmt.Printf("❌ Error getting test product: %v\n", result.Error)
		return
	}

	fmt.Printf("Testing zero values update for product: %s (ID: %d)\n", product.Name, product.ID)
	fmt.Printf("Original values:\n")
	fmt.Printf("- Stock: %d\n", product.Stock)
	fmt.Printf("- Purchase Price: %.2f\n", product.PurchasePrice)
	fmt.Printf("- Sale Price: %.2f\n", product.SalePrice)
	fmt.Printf("- Is Active: %t\n", product.IsActive)

	// Backup original values
	originalStock := product.Stock
	originalPurchasePrice := product.PurchasePrice
	originalSalePrice := product.SalePrice
	originalIsActive := product.IsActive

	fmt.Println("\n=== TESTING OLD METHOD (Updates without Select) ===")
	
	// Test 1: Update with zero values using old method (should fail to update zeros)
	updateDataOld := models.Product{
		Name:          product.Name,
		Code:          product.Code,
		Stock:         0,                // Zero value - should NOT be updated with old method
		PurchasePrice: 0,                // Zero value - should NOT be updated with old method
		SalePrice:     1000.0,           // Non-zero - should be updated
		IsActive:      false,            // Zero value (false) - should NOT be updated with old method
	}

	// Simulate old method
	err := db.Model(&product).Updates(updateDataOld).Error
	if err != nil {
		fmt.Printf("❌ Error with old method: %v\n", err)
	} else {
		fmt.Println("✅ Old method executed without error")
	}

	// Check what was actually updated
	var productAfterOld models.Product
	db.First(&productAfterOld, product.ID)
	
	fmt.Printf("After old method Updates():\n")
	fmt.Printf("- Stock: %d (should still be %d - zero not updated)\n", productAfterOld.Stock, originalStock)
	fmt.Printf("- Purchase Price: %.2f (should still be %.2f - zero not updated)\n", productAfterOld.PurchasePrice, originalPurchasePrice)
	fmt.Printf("- Sale Price: %.2f (should be 1000.00 - non-zero updated)\n", productAfterOld.SalePrice)
	fmt.Printf("- Is Active: %t (should still be %t - false not updated)\n", productAfterOld.IsActive, originalIsActive)

	// Verify the issue
	if productAfterOld.Stock == originalStock && productAfterOld.PurchasePrice == originalPurchasePrice {
		fmt.Printf("❌ CONFIRMED: Old method ignores zero values (Stock: %d, Price: %.2f not updated)\n", 
			productAfterOld.Stock, productAfterOld.PurchasePrice)
	}

	fmt.Println("\n=== TESTING NEW METHOD (Select + Updates) ===")

	// Test 2: Update with zero values using new method (should work)
	updateDataNew := models.Product{
		Name:          product.Name,
		Code:          product.Code,
		Stock:         0,                // Zero value - SHOULD be updated with new method
		PurchasePrice: 0,                // Zero value - SHOULD be updated with new method  
		SalePrice:     2000.0,           // Non-zero - should be updated
		IsActive:      false,            // Zero value (false) - SHOULD be updated with new method
	}

	// Use new method with Select("*")
	err = db.Model(&product).Select("*").Updates(updateDataNew).Error
	if err != nil {
		fmt.Printf("❌ Error with new method: %v\n", err)
	} else {
		fmt.Println("✅ New method executed without error")
	}

	// Check what was actually updated
	var productAfterNew models.Product
	db.First(&productAfterNew, product.ID)
	
	fmt.Printf("After new method Select(*).Updates():\n")
	fmt.Printf("- Stock: %d (should be 0 - zero value updated)\n", productAfterNew.Stock)
	fmt.Printf("- Purchase Price: %.2f (should be 0.00 - zero value updated)\n", productAfterNew.PurchasePrice)
	fmt.Printf("- Sale Price: %.2f (should be 2000.00 - non-zero updated)\n", productAfterNew.SalePrice)
	fmt.Printf("- Is Active: %t (should be false - zero value updated)\n", productAfterNew.IsActive)

	// Verify the fix
	if productAfterNew.Stock == 0 && productAfterNew.PurchasePrice == 0.0 && !productAfterNew.IsActive {
		fmt.Println("✅ SUCCESS: New method correctly updates zero values!")
	} else {
		fmt.Printf("❌ FAILED: New method did not update zero values correctly\n")
	}

	fmt.Println("\n=== TESTING EDGE CASE: Selective Updates ===")
	
	// Test 3: Test selective updates (what frontend might actually send)
	frontendUpdate := map[string]interface{}{
		"stock": 5,
		"purchase_price": 0.0, // User wants to set price to 0
		"sale_price": 1500.0,
	}

	jsonData, _ := json.Marshal(frontendUpdate)
	fmt.Printf("Frontend JSON: %s\n", string(jsonData))

	var partialUpdate models.Product
	json.Unmarshal(jsonData, &partialUpdate)
	
	// Apply selective update with new method
	err = db.Model(&product).Select("stock", "purchase_price", "sale_price").Updates(partialUpdate).Error
	if err != nil {
		fmt.Printf("❌ Error with selective update: %v\n", err)
	}

	var productFinal models.Product
	db.First(&productFinal, product.ID)
	
	fmt.Printf("After selective update:\n")
	fmt.Printf("- Stock: %d (should be 5)\n", productFinal.Stock)
	fmt.Printf("- Purchase Price: %.2f (should be 0.00)\n", productFinal.PurchasePrice)
	fmt.Printf("- Sale Price: %.2f (should be 1500.00)\n", productFinal.SalePrice)

	if productFinal.Stock == 5 && productFinal.PurchasePrice == 0.0 && productFinal.SalePrice == 1500.0 {
		fmt.Println("✅ SUCCESS: Selective update with zero values works!")
	}

	// Restore original values
	fmt.Println("\n=== RESTORING ORIGINAL VALUES ===")
	restoreData := models.Product{
		Stock:         originalStock,
		PurchasePrice: originalPurchasePrice,
		SalePrice:     originalSalePrice,
		IsActive:      originalIsActive,
	}
	
	db.Model(&product).Select("stock", "purchase_price", "sale_price", "is_active").Updates(restoreData)
	fmt.Println("✅ Original values restored")
}