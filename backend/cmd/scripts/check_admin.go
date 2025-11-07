package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Connected to database successfully")

	// Query for admin user
	var adminUser models.User
	if err := db.Where("role = ?", "admin").First(&adminUser).Error; err != nil {
		log.Fatal("Failed to find admin user:", err)
	}

	fmt.Printf("Admin User Details:\n")
	fmt.Printf("ID: %d\n", adminUser.ID)
	fmt.Printf("Username: %s\n", adminUser.Username)
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("Role: %s\n", adminUser.Role)
	fmt.Printf("Active: %t\n", adminUser.IsActive)
	fmt.Printf("First Name: %s\n", adminUser.FirstName)
	fmt.Printf("Last Name: %s\n", adminUser.LastName)
}
