package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🔧 Fixing Director Permissions...")
	
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Find all users with director role
	var directors []models.User
	if err := db.Where("role = ?", "director").Find(&directors).Error; err != nil {
		log.Fatal("Failed to find directors:", err)
	}

	fmt.Printf("Found %d director(s)\n", len(directors))

	for _, director := range directors {
		fmt.Printf("Updating permissions for director: %s (ID: %d)\n", director.Username, director.ID)
		
		// Start transaction
		tx := db.Begin()
		
		// Delete existing permissions for this director
		if err := tx.Where("user_id = ?", director.ID).Delete(&models.ModulePermissionRecord{}).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to delete existing permissions for user %d: %v", director.ID, err)
			continue
		}
		
		// Get default permissions for director role
		defaultPerms := models.GetDefaultPermissions("director")
		
		// Create new permissions
		for module, perm := range defaultPerms {
			permission := models.ModulePermissionRecord{
				UserID:     director.ID,
				Module:     module,
				CanView:    perm.CanView,
				CanCreate:  perm.CanCreate,
				CanEdit:    perm.CanEdit,
				CanDelete:  perm.CanDelete,
				CanApprove: perm.CanApprove,
				CanExport:  perm.CanExport,
			}
			
			if err := tx.Create(&permission).Error; err != nil {
				tx.Rollback()
				log.Printf("Failed to create permission for user %d, module %s: %v", director.ID, module, err)
				continue
			}
		}
		
		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			log.Printf("Failed to commit permissions for user %d: %v", director.ID, err)
			continue
		}
		
		fmt.Printf("✅ Successfully updated permissions for director: %s\n", director.Username)
	}
	
	fmt.Println("🎉 Director permissions fix completed!")
}