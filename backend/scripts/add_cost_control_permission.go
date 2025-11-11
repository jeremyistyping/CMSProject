package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	log.Println("🔄 Adding cost_control permissions to all users...")

	// Get all users
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	successCount := 0
	for _, user := range users {
		// Check if user already has cost_control permission
		var existingPerm models.ModulePermissionRecord
		if err := db.Where("user_id = ? AND module = ?", user.ID, "cost_control").First(&existingPerm).Error; err == nil {
			log.Printf("⏭️  User %s already has cost_control permission, skipping", user.Username)
			continue
		}

		// Get default permissions for user's role
		defaultPerms := models.GetDefaultPermissions(user.Role)
		perm := defaultPerms["cost_control"]

		if perm != nil {
			permission := models.ModulePermissionRecord{
				UserID:     user.ID,
				Module:     "cost_control",
				CanView:    perm.CanView,
				CanCreate:  perm.CanCreate,
				CanEdit:    perm.CanEdit,
				CanDelete:  perm.CanDelete,
				CanApprove: perm.CanApprove,
				CanExport:  perm.CanExport,
				CanMenu:    perm.CanMenu,
			}

			if err := db.Create(&permission).Error; err != nil {
				log.Printf("⚠️  Error creating cost_control permission for user %s: %v", user.Username, err)
			} else {
				log.Printf("✅ Cost control permission added for user %s (role: %s) - CanView: %v, CanMenu: %v", 
					user.Username, user.Role, perm.CanView, perm.CanMenu)
				successCount++
			}
		}
	}

	log.Printf("📊 Cost control permission migration completed: %d/%d users updated", successCount, len(users))
}
