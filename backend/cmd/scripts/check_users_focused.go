package main

import (
	"fmt"
	"log"
	"os"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("=== CHECKING USERS IN DATABASE ===")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
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

	// Check for admin user
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	fmt.Printf("\nFound %d admin users\n", adminCount)

	// Check for specific admin user with common credentials
	commonAdminCredentials := []string{
		"admin", "admin@admin.com", "administrator", "root",
	}
	
	fmt.Println("\nChecking for common admin credentials:")
	for _, cred := range commonAdminCredentials {
		var user models.User
		if err := db.Where("username = ? OR email = ?", cred, cred).First(&user).Error; err == nil {
			fmt.Printf("✅ Found admin user with credential '%s': Username=%s, Email=%s\n", 
				cred, user.Username, user.Email)
		} else {
			fmt.Printf("❌ No admin user found with credential '%s'\n", cred)
		}
	}

	// Exit with success
	os.Exit(0)
}