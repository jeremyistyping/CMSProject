package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	log.Println("🔄 Starting menu permissions update...")

	// Get all users
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	modules := []string{"accounts", "products", "contacts", "assets", "sales", "purchases", "payments", "cash_bank", "reports", "settings"}
	
	updatedCount := 0
	for _, user := range users {
		log.Printf("Processing user: %s (role: %s)", user.Username, user.Role)
		
		// Get default permissions for this user's role
		defaultPerms := models.GetDefaultPermissions(user.Role)
		
		// Update permissions for each module
		for _, module := range modules {
			perm := defaultPerms[module]
			if perm == nil {
				continue
			}
			
			// Update the existing permission record
			result := db.Model(&models.ModulePermissionRecord{}).
				Where("user_id = ? AND module = ?", user.ID, module).
				Update("can_menu", perm.CanMenu)
			
			if result.Error != nil {
				log.Printf("⚠️  Error updating %s permission for user %s: %v", module, user.Username, result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("✅ Updated %s permission for user %s - CanMenu: %v", module, user.Username, perm.CanMenu)
				updatedCount++
			}
		}
	}

	log.Printf("✅ Menu permissions update completed! Updated %d records", updatedCount)
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total users processed: %d\n", len(users))
	fmt.Printf("Total permissions updated: %d\n", updatedCount)
}
