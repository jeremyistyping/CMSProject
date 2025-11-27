package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"os"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	roles := []string{"employee", "purchasing", "inventory_manager"}
	modules := []string{"purchases", "cost_control", "daily_updates"}

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("%-20s | %-15s | %-10s | %-10s\n", "Role", "Module", "CanView", "CanCreate")
	fmt.Println("----------------------------------------------------------------")

	for _, roleName := range roles {
		var user models.User
		if err := db.Where("role = ?", roleName).First(&user).Error; err != nil {
			// Try finding by username if role lookup fails (assuming username=role for testing)
			if err := db.Where("username = ?", roleName).First(&user).Error; err != nil {
				fmt.Printf("User with role/username '%s' not found\n", roleName)
				continue
			}
		}

		for _, module := range modules {
			var perm models.ModulePermissionRecord
			db.Where("user_id = ? AND module = ?", user.ID, module).First(&perm)

			fmt.Printf("%-20s | %-15s | %-10v | %-10v\n", roleName, module, perm.CanView, perm.CanCreate)
		}
		fmt.Println("----------------------------------------------------------------")
	}

	os.Exit(0)
}
