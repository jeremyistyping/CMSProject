package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	// Helper function to update permission
	updatePermission := func(roleName string, module string, updates map[string]interface{}) {
		var user models.User
		if err := db.Where("role = ?", roleName).First(&user).Error; err != nil {
			// Try finding by username if role lookup fails
			if err := db.Where("username = ?", roleName).First(&user).Error; err != nil {
				log.Printf("User with role/username '%s' not found, skipping...", roleName)
				return
			}
		}

		var perm models.ModulePermissionRecord
		result := db.Where("user_id = ? AND module = ?", user.ID, module).First(&perm)

		if result.Error != nil {
			// Create if not exists
			log.Printf("Creating permission record for %s on %s", roleName, module)
			perm = models.ModulePermissionRecord{
				UserID: user.ID,
				Module: module,
			}
			// Apply updates to new record
			for k, v := range updates {
				switch k {
				case "CanView":
					perm.CanView = v.(bool)
				case "CanCreate":
					perm.CanCreate = v.(bool)
				case "CanEdit":
					perm.CanEdit = v.(bool)
				case "CanDelete":
					perm.CanDelete = v.(bool)
				case "CanApprove":
					perm.CanApprove = v.(bool)
				case "CanMenu":
					perm.CanMenu = v.(bool)
				}
			}
			if err := db.Create(&perm).Error; err != nil {
				log.Printf("Error creating permission: %v", err)
			}
		} else {
			// Update existing
			log.Printf("Updating permission record for %s on %s", roleName, module)
			if err := db.Model(&perm).Updates(updates).Error; err != nil {
				log.Printf("Error updating permission: %v", err)
			}
		}
	}

	fmt.Println("Updating RBAC permissions...")

	// 1. Employee: Cannot create purchases, Can create daily updates
	updatePermission("employee", "purchases", map[string]interface{}{
		"CanCreate": false,
		"CanEdit":   false,
		"CanDelete": false,
	})
	updatePermission("employee", "daily_updates", map[string]interface{}{
		"CanView":   true,
		"CanCreate": true,
		"CanEdit":   true, // Can edit their own usually, but RBAC is coarse
		"CanMenu":   true,
	})

	// 2. Purchasing: Can create purchases, Can approve daily updates
	updatePermission("purchasing", "purchases", map[string]interface{}{
		"CanView":   true,
		"CanCreate": true,
		"CanEdit":   true,
		"CanDelete": true,
		"CanMenu":   true,
	})
	updatePermission("purchasing", "daily_updates", map[string]interface{}{
		"CanView":    true,
		"CanApprove": true,
		"CanMenu":    true,
	})

	// 3. Inventory Manager: Can create purchases
	updatePermission("inventory_manager", "purchases", map[string]interface{}{
		"CanView":   true,
		"CanCreate": true,
		"CanEdit":   true,
		"CanMenu":   true,
	})

	fmt.Println("RBAC permissions updated successfully.")
	os.Exit(0)
}
