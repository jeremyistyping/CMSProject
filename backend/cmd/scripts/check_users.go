package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("=== CHECKING USERS IN DATABASE ===")
	
	// Connect to database
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	// Get all users
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	fmt.Printf("\n📋 Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("ID: %d, Username: %s, Email: %s, Role: %s, Active: %t\n", 
			user.ID, user.Username, user.Email, user.Role, user.IsActive)
	}

	// Check for specific admin user
	var adminUser models.User
	err = db.Where("username = ? OR email = ?", "admin", "admin@admin.com").First(&adminUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("\n❌ No admin user found with username 'admin' or email 'admin@admin.com'")
		} else {
			fmt.Printf("\n❌ Error checking for admin user: %v\n", err)
		}
	} else {
		fmt.Printf("\n✅ Found admin user: ID=%d, Username=%s, Email=%s, Role=%s, Active=%t\n", 
			adminUser.ID, adminUser.Username, adminUser.Email, adminUser.Role, adminUser.IsActive)
	}
}