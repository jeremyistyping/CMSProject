package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize database connection  
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=== Starting Duplicate Notification Cleanup ===")

	// Clean up duplicate notifications
	if err := cleanupDuplicateNotifications(db); err != nil {
		log.Fatalf("Failed to cleanup duplicate notifications: %v", err)
	}

	fmt.Println("✅ Duplicate notification cleanup completed successfully!")
}

func cleanupDuplicateNotifications(db *gorm.DB) error {
	// Find and remove duplicate notifications based on:
	// - Same user_id
	// - Same type
	// - Same purchase_id (from JSON data)
	// - Created within same hour
	
	var duplicates []struct {
		UserID     uint
		Type       string
		PurchaseID string
		Count      int64
	}

	// Find duplicates
	err := db.Raw(`
		SELECT 
			user_id,
			type,
			data::json->>'purchase_id' as purchase_id,
			COUNT(*) as count
		FROM notifications 
		WHERE data::json->>'purchase_id' IS NOT NULL
		AND created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY user_id, type, data::json->>'purchase_id'
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error

	if err != nil {
		return fmt.Errorf("failed to find duplicates: %v", err)
	}

	fmt.Printf("Found %d sets of duplicate notifications\n", len(duplicates))

	totalRemoved := 0
	for _, dup := range duplicates {
		if dup.PurchaseID == "" {
			continue
		}

		// Keep the latest notification, remove older ones
		var notifications []models.Notification
		err := db.Where("user_id = ? AND type = ? AND data::json->>'purchase_id' = ?",
			dup.UserID, dup.Type, dup.PurchaseID).
			Order("created_at DESC").
			Find(&notifications).Error

		if err != nil {
			fmt.Printf("Error finding notifications for user %d, type %s, purchase %s: %v\n",
				dup.UserID, dup.Type, dup.PurchaseID, err)
			continue
		}

		// Keep first (latest), remove the rest
		if len(notifications) > 1 {
			var idsToDelete []uint
			for i := 1; i < len(notifications); i++ {
				idsToDelete = append(idsToDelete, notifications[i].ID)
			}

			if len(idsToDelete) > 0 {
				err := db.Where("id IN ?", idsToDelete).Delete(&models.Notification{}).Error
				if err != nil {
					fmt.Printf("Error deleting duplicates: %v\n", err)
					continue
				}
				
				removed := len(idsToDelete)
				totalRemoved += removed
				fmt.Printf("✅ Removed %d duplicate notifications for user %d, type %s, purchase %s\n",
					removed, dup.UserID, dup.Type, dup.PurchaseID)
			}
		}
	}

	fmt.Printf("\n=== Cleanup Summary ===\n")
	fmt.Printf("Total duplicate notifications removed: %d\n", totalRemoved)

	return nil
}
