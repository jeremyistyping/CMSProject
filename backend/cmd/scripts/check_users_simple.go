package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("=== CHECKING USERS IN DATABASE ===")
	
	// Initialize database connection
	db, err := database.Database()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connected successfully")

	// Get all users
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	fmt.Printf("\n📋 Total users found: %d\n", len(users))
	fmt.Println("================================")
	
	for i, user := range users {
		fmt.Printf("%d. Username: %s\n", i+1, user.Username)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Role: %s\n", user.Role)
		fmt.Printf("   Active: %t\n", user.IsActive)
		fmt.Printf("   Created: %v\n", user.CreatedAt)
		fmt.Println("   -------------------------")
	}

	// Specific check for admin credentials
	fmt.Println("\n🔍 CHECKING SPECIFIC CREDENTIALS:")
	
	// Check admin@company.com (from seed file)
	var adminUser models.User
	err = db.Where("email = ?", "admin@company.com").First(&adminUser).Error
	if err == nil {
		fmt.Printf("✅ Found user with email 'admin@company.com':\n")
		fmt.Printf("   Username: %s\n", adminUser.Username)
		fmt.Printf("   Role: %s\n", adminUser.Role)
		fmt.Printf("   Active: %t\n", adminUser.IsActive)
	} else {
		fmt.Printf("❌ No user found with email 'admin@company.com'\n")
	}

	// Check username 'admin'
	err = db.Where("username = ?", "admin").First(&adminUser).Error
	if err == nil {
		fmt.Printf("✅ Found user with username 'admin':\n")
		fmt.Printf("   Email: %s\n", adminUser.Email)
		fmt.Printf("   Role: %s\n", adminUser.Role)
		fmt.Printf("   Active: %t\n", adminUser.IsActive)
	} else {
		fmt.Printf("❌ No user found with username 'admin'\n")
	}
}