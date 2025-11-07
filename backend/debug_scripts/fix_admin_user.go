package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	_ = cfg

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== CHECKING ADMIN USER STATUS ===")
	
	// Check if admin user exists
	var adminUser models.User
	err := db.Where("email = ?", "admin@company.com").First(&adminUser).Error
	
	if err != nil {
		fmt.Printf("❌ Admin user not found: %v\n", err)
		fmt.Println("Creating new admin user...")
		
		// Create new admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash password:", err)
		}
		
		newAdmin := models.User{
			Username:  "admin",
			Email:     "admin@company.com",
			Password:  string(hashedPassword),
			Role:      "admin",
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
		}
		
		if err := db.Create(&newAdmin).Error; err != nil {
			log.Fatal("Failed to create admin user:", err)
		}
		
		fmt.Println("✅ Admin user created successfully!")
		
	} else {
		fmt.Printf("✅ Admin user found: ID=%d, Username=%s, Email=%s, Active=%t\n", 
			adminUser.ID, adminUser.Username, adminUser.Email, adminUser.IsActive)
			
		// Check if user is active
		if !adminUser.IsActive {
			fmt.Println("❌ Admin user is inactive! Activating...")
			adminUser.IsActive = true
			db.Save(&adminUser)
			fmt.Println("✅ Admin user activated!")
		}
		
		// Test password verification
		err := bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte("password123"))
		if err != nil {
			fmt.Printf("❌ Password verification failed: %v\n", err)
			fmt.Println("Updating password to 'password123'...")
			
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			if err != nil {
				log.Fatal("Failed to hash new password:", err)
			}
			
			adminUser.Password = string(hashedPassword)
			if err := db.Save(&adminUser).Error; err != nil {
				log.Fatal("Failed to update password:", err)
			}
			
			fmt.Println("✅ Admin password updated successfully!")
		} else {
			fmt.Println("✅ Password verification successful!")
		}
	}
	
	fmt.Println("\n=== CHECKING ALL USERS ===")
	var users []models.User
	db.Find(&users)
	
	for _, user := range users {
		status := "INACTIVE"
		if user.IsActive {
			status = "ACTIVE"
		}
		fmt.Printf("User: %s (%s) - Role: %s - Status: %s\n", 
			user.Email, user.Username, user.Role, status)
	}
	
	fmt.Println("\n=== FIX COMPLETED ===")
	fmt.Println("Please try logging in with admin@company.com / password123")
}