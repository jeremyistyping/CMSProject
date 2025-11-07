package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  MANUALLY TRIGGER STOCK NOTIFICATIONS")
	fmt.Println("========================================")

	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Initialize services
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(db, notificationRepo)
	stockMonitoringService := services.NewStockMonitoringService(db, notificationService)

	fmt.Println("🔍 Step 1: Check current stock status...")

	var products []models.Product
	err := db.Where("is_active = ?", true).
		Preload("Category").
		Order("id ASC").
		Find(&products).Error
	if err != nil {
		log.Fatalf("Error fetching products: %v", err)
	}

	for _, product := range products {
		fmt.Printf("Product %s (ID: %d):\n", product.Name, product.ID)
		fmt.Printf("  - Current Stock: %d\n", product.Stock)
		fmt.Printf("  - Min Stock: %d\n", product.MinStock)
		fmt.Printf("  - Reorder Level: %d\n", product.ReorderLevel)
		
		shouldTriggerMin := product.MinStock > 0 && product.Stock <= product.MinStock
		shouldTriggerReorder := product.ReorderLevel > 0 && product.Stock <= product.ReorderLevel
		
		if shouldTriggerMin {
			fmt.Printf("  - 🚨 SHOULD trigger MIN_STOCK notification\n")
		}
		if shouldTriggerReorder {
			fmt.Printf("  - 📋 SHOULD trigger REORDER_ALERT notification\n")
		}
		if !shouldTriggerMin && !shouldTriggerReorder {
			fmt.Printf("  - ✅ No notifications needed\n")
		}
		fmt.Println()
	}

	fmt.Println("🔥 Step 2: Force trigger notifications for all products...")

	for _, product := range products {
		fmt.Printf("Checking Product ID %d (%s)...\n", product.ID, product.Code)
		
		err := stockMonitoringService.CheckSingleProductStock(product.ID)
		if err != nil {
			fmt.Printf("❌ Error checking product %d: %v\n", product.ID, err)
		} else {
			fmt.Printf("✅ Check completed for product %d\n", product.ID)
		}
	}

	fmt.Println("\n🔍 Step 3: Verify notifications were created...")

	// Check notifications created in last 5 minutes
	var recentNotifications []models.Notification
	err = db.Where("type IN ? AND created_at > NOW() - INTERVAL '5 minutes'", 
		[]string{models.NotificationTypeLowStock, models.NotificationTypeReorderAlert}).
		Order("created_at DESC").
		Find(&recentNotifications).Error
	if err != nil {
		log.Fatalf("Error fetching recent notifications: %v", err)
	}

	fmt.Printf("📬 Found %d notifications created in last 5 minutes:\n", len(recentNotifications))
	for _, notif := range recentNotifications {
		fmt.Printf("  - ID: %d, Type: %s, User ID: %d\n", notif.ID, notif.Type, notif.UserID)
		fmt.Printf("  - Title: %s\n", notif.Title)
		fmt.Printf("  - Created: %s\n", notif.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	fmt.Println("🔍 Step 4: Check stock alerts...")

	var stockAlerts []models.StockAlert
	err = db.Where("status = ?", models.StockAlertStatusActive).
		Preload("Product").
		Order("created_at DESC").
		Find(&stockAlerts).Error
	if err != nil {
		log.Fatalf("Error fetching stock alerts: %v", err)
	}

	fmt.Printf("🚨 Found %d active stock alerts:\n", len(stockAlerts))
	for _, alert := range stockAlerts {
		fmt.Printf("  - Product: %s (ID: %d)\n", alert.Product.Name, alert.ProductID)
		fmt.Printf("  - Alert Type: %s\n", alert.AlertType)
		fmt.Printf("  - Current Stock: %d, Threshold: %d\n", alert.CurrentStock, alert.ThresholdStock)
		fmt.Printf("  - Status: %s\n", alert.Status)
		fmt.Printf("  - Created: %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Println("  NOTIFICATION TRIGGER COMPLETED")
	fmt.Println("========================================")

	fmt.Println("\n💡 SUMMARY:")
	fmt.Printf("  - Products checked: %d\n", len(products))
	fmt.Printf("  - Recent notifications: %d\n", len(recentNotifications))
	fmt.Printf("  - Active stock alerts: %d\n", len(stockAlerts))
	
	fmt.Println("\n🎯 NEXT STEPS:")
	fmt.Println("1. Check your frontend notification API endpoints")
	fmt.Println("2. Verify that notifications are being fetched properly")
	fmt.Println("3. Check notification polling/refresh mechanism")
	fmt.Println("4. Ensure notification UI components are working")
}