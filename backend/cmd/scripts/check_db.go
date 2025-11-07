package main

import (
	"encoding/json"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== DATABASE CONNECTION CHECK ===")
	fmt.Println("✅ Database connected successfully")
	
	// Check if products table exists
	fmt.Println("\n=== PRODUCTS TABLE CHECK ===")
	if !db.Migrator().HasTable(&models.Product{}) {
		fmt.Println("❌ Products table does not exist")
		return
	}
	fmt.Println("✅ Products table exists")

	// Check products table structure
	fmt.Println("\n=== PRODUCTS TABLE COLUMNS ===")
	columns, err := db.Migrator().ColumnTypes(&models.Product{})
	if err != nil {
		fmt.Printf("❌ Error getting column types: %v\n", err)
		return
	}

	for _, column := range columns {
		fmt.Printf("- %s: %s\n", column.Name(), column.DatabaseTypeName())
	}

	// Check if stock column exists
	if !db.Migrator().HasColumn(&models.Product{}, "stock") {
		fmt.Println("❌ Stock column does not exist in products table")
		return
	}
	fmt.Println("✅ Stock column exists")

	// Get sample products
	fmt.Println("\n=== SAMPLE PRODUCTS DATA ===")
	var products []models.Product
	result := db.Limit(5).Find(&products)
	if result.Error != nil {
		fmt.Printf("❌ Error fetching products: %v\n", result.Error)
		return
	}

	fmt.Printf("Found %d products in database\n", len(products))
	for _, product := range products {
		fmt.Printf("ID: %d, Code: %s, Name: %s, Stock: %d\n", 
			product.ID, product.Code, product.Name, product.Stock)
	}

	// Test product update
	if len(products) > 0 {
		fmt.Println("\n=== TESTING PRODUCT UPDATE ===")
		testProduct := products[0]
		originalStock := testProduct.Stock
		newStock := originalStock + 10

		fmt.Printf("Testing update for product ID: %d\n", testProduct.ID)
		fmt.Printf("Original stock: %d\n", originalStock)
		fmt.Printf("New stock: %d\n", newStock)

		// Update stock
		updateResult := db.Model(&testProduct).Update("stock", newStock)
		if updateResult.Error != nil {
			fmt.Printf("❌ Error updating product stock: %v\n", updateResult.Error)
			return
		}

		// Verify update
		var updatedProduct models.Product
		db.First(&updatedProduct, testProduct.ID)
		
		if updatedProduct.Stock == newStock {
			fmt.Println("✅ Product stock updated successfully")
			
			// Restore original stock
			db.Model(&testProduct).Update("stock", originalStock)
			fmt.Println("✅ Original stock restored")
		} else {
			fmt.Printf("❌ Stock update failed. Expected: %d, Got: %d\n", newStock, updatedProduct.Stock)
		}
	}

	// Test JSON binding simulation
	fmt.Println("\n=== TESTING JSON BINDING SIMULATION ===")
	testJSONUpdate(db)
}

func testJSONUpdate(db *gorm.DB) {
	// Get first product
	var product models.Product
	result := db.First(&product)
	if result.Error != nil {
		fmt.Printf("❌ Error getting test product: %v\n", result.Error)
		return
	}

	originalData := product
	fmt.Printf("Testing JSON update for product: %s (ID: %d)\n", product.Name, product.ID)
	fmt.Printf("Current stock: %d\n", product.Stock)

	// Simulate JSON payload that might come from frontend
	updateJSON := map[string]interface{}{
		"name":          product.Name,
		"code":          product.Code,
		"stock":         product.Stock + 5,
		"purchase_price": product.PurchasePrice,
		"sale_price":    product.SalePrice,
		"category_id":   product.CategoryID,
		"warehouse_location_id": product.WarehouseLocationID,
	}

	jsonBytes, _ := json.Marshal(updateJSON)
	fmt.Printf("Simulated JSON payload: %s\n", string(jsonBytes))

	// Try to unmarshal into Product struct (simulate ShouldBindJSON)
	var updateData models.Product
	err := json.Unmarshal(jsonBytes, &updateData)
	if err != nil {
		fmt.Printf("❌ JSON unmarshaling error: %v\n", err)
		return
	}

	fmt.Printf("Unmarshaled stock value: %d\n", updateData.Stock)

	// Try update using GORM Updates method (like in UpdateProduct function)
	updateResult := db.Model(&product).Updates(updateData)
	if updateResult.Error != nil {
		fmt.Printf("❌ GORM Updates error: %v\n", updateResult.Error)
		return
	}

	fmt.Printf("Rows affected: %d\n", updateResult.RowsAffected)

	// Check if update was successful
	var updatedProduct models.Product
	db.First(&updatedProduct, product.ID)
	
	if updatedProduct.Stock == updateData.Stock {
		fmt.Println("✅ JSON-based update successful")
	} else {
		fmt.Printf("❌ JSON-based update failed. Expected: %d, Got: %d\n", 
			updateData.Stock, updatedProduct.Stock)
	}

	// Restore original data
	db.Model(&product).Updates(originalData)
	fmt.Println("✅ Original data restored")

	// Check for potential issues
	fmt.Println("\n=== POTENTIAL ISSUES ANALYSIS ===")
	
	// Check if there are any triggers or constraints
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.triggers WHERE event_object_table = 'products'").Scan(&count)
	if count > 0 {
		fmt.Printf("⚠️  Found %d triggers on products table\n", count)
		
		// Get trigger details
		var triggers []struct {
			TriggerName string `json:"trigger_name"`
			EventManipulation string `json:"event_manipulation"`
		}
		db.Raw("SELECT trigger_name, event_manipulation FROM information_schema.triggers WHERE event_object_table = 'products'").Scan(&triggers)
		
		for _, trigger := range triggers {
			fmt.Printf("   - Trigger: %s, Event: %s\n", trigger.TriggerName, trigger.EventManipulation)
		}
	} else {
		fmt.Println("✅ No triggers found on products table")
	}

	// Check foreign key constraints
	var fkCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.key_column_usage 
		WHERE table_name = 'products' 
		AND referenced_table_name IS NOT NULL
	`).Scan(&fkCount)
	
	if fkCount > 0 {
		fmt.Printf("ℹ️  Found %d foreign key constraints on products table\n", fkCount)
	}
}