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

	// Find purchasing user
	var user models.User
	if err := db.Where("username = ?", "purchasing").First(&user).Error; err != nil {
		log.Fatalf("Failed to find purchasing user: %v", err)
	}

	fmt.Printf("Found user: %s (ID: %d)\n", user.Username, user.ID)

	// Update or create cost_control permission with can_menu = true
	var perm models.ModulePermissionRecord
	result := db.Where("user_id = ? AND module = ?", user.ID, "cost_control").First(&perm)

	if result.Error != nil {
		// Create new permission
		perm = models.ModulePermissionRecord{
			UserID:     user.ID,
			Module:     "cost_control",
			CanView:    true,
			CanMenu:    true,
			CanCreate:  false,
			CanEdit:    false,
			CanDelete:  false,
			CanApprove: false,
			CanExport:  false,
		}
		if err := db.Create(&perm).Error; err != nil {
			log.Fatalf("Failed to create permission: %v", err)
		}
		fmt.Println("✅ Created cost_control permission with can_menu=true")
	} else {
		// Update existing
		perm.CanMenu = true
		if err := db.Save(&perm).Error; err != nil {
			log.Fatalf("Failed to update permission: %v", err)
		}
		fmt.Println("✅ Updated cost_control permission with can_menu=true")
	}

	fmt.Println("\nVerifying all permissions for purchasing user:")
	var allPerms []models.ModulePermissionRecord
	db.Where("user_id = ?", user.ID).Find(&allPerms)

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("%-15s | %-5s | %-5s\n", "Module", "View", "Menu")
	fmt.Println("----------------------------------------------------------------")
	for _, p := range allPerms {
		fmt.Printf("%-15s | %-5v | %-5v\n", p.Module, p.CanView, p.CanMenu)
	}
	fmt.Println("----------------------------------------------------------------")

	os.Exit(0)
}
