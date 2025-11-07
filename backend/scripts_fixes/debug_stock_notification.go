package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  STOCK NOTIFICATION DEBUG TOOL")
	fmt.Println("========================================")

	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Initialize services
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(db, notificationRepo)
	stockMonitoringService := services.NewStockMonitoringService(db, notificationService)

	fmt.Println("🔍 Step 1: Checking products with zero/low stock...")
	
	var products []models.Product
	err := db.Where("is_active = ?", true).
		Preload("Category").
		Order("stock ASC").
		Find(&products).Error
	if err != nil {
		log.Fatalf("Error fetching products: %v", err)
	}

	fmt.Printf("✅ Found %d active products\n\n", len(products))

	// Show products with potential stock issues
	fmt.Println("📊 Product Stock Analysis:")
	fmt.Println("=========================")
	
	lowStockCount := 0
	zeroStockCount := 0
	noMinStockCount := 0
	
	for i, product := range products {
		if i >= 10 { // Show only first 10 products
			break
		}
		
		fmt.Printf("Product ID %d (%s):\n", product.ID, product.Code)
		fmt.Printf("  - Name: %s\n", product.Name)
		fmt.Printf("  - Current Stock: %d\n", product.Stock)
		fmt.Printf("  - Min Stock: %d\n", product.MinStock)
		fmt.Printf("  - Reorder Level: %d\n", product.ReorderLevel)
		
		if product.Stock == 0 {
			fmt.Printf("  - 🔴 STATUS: OUT OF STOCK\n")
			zeroStockCount++
		} else if product.MinStock > 0 && product.Stock <= product.MinStock {
			fmt.Printf("  - 🟡 STATUS: BELOW MINIMUM STOCK\n")
			lowStockCount++
		} else if product.MinStock == 0 {
			fmt.Printf("  - ⚪ STATUS: NO MIN STOCK SET\n")
			noMinStockCount++
		} else {
			fmt.Printf("  - ✅ STATUS: STOCK OK\n")
		}
		fmt.Println()
	}

	fmt.Printf("📈 Summary:\n")
	fmt.Printf("  - Zero stock products: %d\n", zeroStockCount)
	fmt.Printf("  - Low stock products: %d\n", lowStockCount)
	fmt.Printf("  - No min stock set: %d\n", noMinStockCount)
	fmt.Println()

	fmt.Println("🔍 Step 2: Checking active users with notification permissions...")
	
	var users []models.User
	err = db.Where("role IN ? AND is_active = ?", []string{"inventory_manager", "admin"}, true).
		Find(&users).Error
	if err != nil {
		log.Fatalf("Error fetching users: %v", err)
	}

	fmt.Printf("✅ Found %d users with notification permissions:\n", len(users))
	for _, user := range users {
		fmt.Printf("  - ID: %d, Username: %s, Role: %s\n", user.ID, user.Username, user.Role)
	}
	fmt.Println()

	fmt.Println("🔍 Step 3: Checking existing stock alerts...")
	
	var stockAlerts []models.StockAlert
	err = db.Where("status = ?", models.StockAlertStatusActive).
		Preload("Product").
		Order("created_at DESC").
		Find(&stockAlerts).Error
	if err != nil {
		log.Fatalf("Error fetching stock alerts: %v", err)
	}

	fmt.Printf("✅ Found %d active stock alerts:\n", len(stockAlerts))
	for _, alert := range stockAlerts {
		fmt.Printf("  - Product: %s (ID: %d)\n", alert.Product.Name, alert.ProductID)
		fmt.Printf("  - Alert Type: %s\n", alert.AlertType)
		fmt.Printf("  - Current Stock: %d, Threshold: %d\n", alert.CurrentStock, alert.ThresholdStock)
		fmt.Printf("  - Last Alert: %s\n", alert.LastAlertAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	fmt.Println("🔍 Step 4: Checking existing notifications...")
	
	var notifications []models.Notification
	err = db.Where("type IN ?", []string{models.NotificationTypeLowStock, models.NotificationTypeReorderAlert, models.NotificationTypeMinStock}).
		Order("created_at DESC").
		Limit(10).
		Find(&notifications).Error
	if err != nil {
		log.Fatalf("Error fetching notifications: %v", err)
	}

	fmt.Printf("✅ Found %d stock-related notifications (last 10):\n", len(notifications))
	for _, notif := range notifications {
		fmt.Printf("  - ID: %d, User ID: %d, Type: %s\n", notif.ID, notif.UserID, notif.Type)
		fmt.Printf("  - Title: %s\n", notif.Title)
		fmt.Printf("  - Read: %t, Priority: %s\n", notif.IsRead, notif.Priority)
		fmt.Printf("  - Created: %s\n", notif.CreatedAt.Format("2006-01-02 15:04:05"))
		
		// Parse notification data
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(notif.Data), &data); err == nil {
			if productID, ok := data["product_id"]; ok {
				fmt.Printf("  - Product ID: %.0f\n", productID)
			}
		}
		fmt.Println()
	}

	fmt.Println("🔍 Step 5: Testing stock monitoring service...")
	
	if len(products) > 0 {
		// Test with first product that has min_stock > 0
		var testProduct *models.Product
		for _, p := range products {
			if p.MinStock > 0 {
				testProduct = &p
				break
			}
		}
		
		if testProduct != nil {
			fmt.Printf("🧪 Testing with Product ID %d (%s)\n", testProduct.ID, testProduct.Code)
			fmt.Printf("  - Current Stock: %d, Min Stock: %d\n", testProduct.Stock, testProduct.MinStock)
			
			// Force check this specific product
			err = stockMonitoringService.CheckSingleProductStock(testProduct.ID)
			if err != nil {
				fmt.Printf("❌ Error testing stock monitoring: %v\n", err)
			} else {
				fmt.Printf("✅ Stock monitoring check completed\n")
			}
			
			// Check if new notifications were created
			var newNotifications []models.Notification
			err = db.Where("type IN ? AND created_at > ?", 
				[]string{models.NotificationTypeLowStock, models.NotificationTypeReorderAlert}, 
				time.Now().Add(-5*time.Minute)).
				Find(&newNotifications).Error
			if err != nil {
				fmt.Printf("❌ Error checking new notifications: %v\n", err)
			} else {
				fmt.Printf("📬 New notifications created in last 5 minutes: %d\n", len(newNotifications))
			}
		} else {
			fmt.Println("⚠️  No products with min_stock > 0 found for testing")
		}
	}

	fmt.Println()
	fmt.Println("🔍 Step 6: Running full stock monitoring check...")
	
	err = stockMonitoringService.ScheduledStockCheck()
	if err != nil {
		fmt.Printf("❌ Error running scheduled stock check: %v\n", err)
	} else {
		fmt.Printf("✅ Scheduled stock check completed\n")
	}

	// Final check for notifications
	var finalNotifications []models.Notification
	err = db.Where("type IN ? AND created_at > ?", 
		[]string{models.NotificationTypeLowStock, models.NotificationTypeReorderAlert}, 
		time.Now().Add(-2*time.Minute)).
		Find(&finalNotifications).Error
	if err != nil {
		fmt.Printf("❌ Error checking final notifications: %v\n", err)
	} else {
		fmt.Printf("📬 New notifications created in last 2 minutes: %d\n", len(finalNotifications))
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  DIAGNOSIS COMPLETED")
	fmt.Println("========================================")
	fmt.Println()
	
	fmt.Println("💡 POTENTIAL ISSUES TO CHECK:")
	fmt.Println("1. Products may not have min_stock or reorder_level set")
	fmt.Println("2. No users with 'admin' or 'inventory_manager' roles")
	fmt.Println("3. Stock monitoring service may not be called automatically")
	fmt.Println("4. Database constraints or migration issues")
	fmt.Println("5. Notification type constants mismatch (LOW_STOCK vs MIN_STOCK)")
	
	fmt.Println()
	fmt.Println("🔧 RECOMMENDED FIXES:")
	fmt.Println("1. Set min_stock values for products that should have notifications")
	fmt.Println("2. Ensure users have correct roles (admin/inventory_manager)")
	fmt.Println("3. Call stock monitoring after product stock changes")
	fmt.Println("4. Add automated scheduling for stock monitoring")
	fmt.Println("5. Verify notification API endpoints work correctly")
}