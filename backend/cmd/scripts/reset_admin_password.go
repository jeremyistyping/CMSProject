package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db := database.ConnectDB()
	
	// Find admin user
	var admin models.User
	if err := db.Where("email = ? AND role = ?", "admin@company.com", "admin").First(&admin).Error; err != nil {
		log.Fatal("Admin user not found:", err)
	}
	
	// Hash new password
	newPassword := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}
	
	// Update password
	admin.Password = string(hashedPassword)
	if err := db.Save(&admin).Error; err != nil {
		log.Fatal("Failed to update password:", err)
	}
	
	fmt.Printf("✅ Admin password reset successfully!\n")
	fmt.Printf("Email: %s\n", admin.Email)
	fmt.Printf("New Password: %s\n", newPassword)
}