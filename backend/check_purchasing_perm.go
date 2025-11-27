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

	var user models.User
	if err := db.Where("username = ?", "purchasing").First(&user).Error; err != nil {
		log.Fatalf("Failed to find purchasing user: %v", err)
	}

	fmt.Printf("User Found: %s (ID: %d, Role: %s)\n", user.Username, user.ID, user.Role)

	var permissions []models.ModulePermissionRecord
	if err := db.Where("user_id = ?", user.ID).Find(&permissions).Error; err != nil {
		log.Fatalf("Failed to fetch permissions: %v", err)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("%-15s | %-5s | %-5s | %-5s | %-5s | %-5s\n", "Module", "View", "Create", "Edit", "Delete", "Approve")
	fmt.Println("----------------------------------------------------------------")
	for _, p := range permissions {
		fmt.Printf("%-15s | %-5v | %-5v | %-5v | %-5v | %-5v\n", p.Module, p.CanView, p.CanCreate, p.CanEdit, p.CanDelete, p.CanApprove)
	}
	fmt.Println("----------------------------------------------------------------")

	os.Exit(0)
}
