package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"not null"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"not null;default:'employee'"`
	FirstName *string
	LastName  *string
}

func main() {
	fmt.Println("🔧 Creating Test User for SSOT API Testing")
	fmt.Println("==========================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Check if test user already exists
	var existingUser User
	err := db.Where("email = ?", "admin@company.com").First(&existingUser).Error
	if err == nil {
		fmt.Printf("✅ Test user already exists: %s (ID: %d, Role: %s)\n", 
			existingUser.Email, existingUser.ID, existingUser.Role)
		
		// Update password to admin123 for testing
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		existingUser.Password = string(hashedPassword)
		db.Save(&existingUser)
		fmt.Println("✅ Password updated to 'admin123' for test user")
		return
	}

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	firstName := "Test"
	lastName := "Admin"
	testUser := User{
		Username:  "testadmin",
		Email:     "admin@example.com",
		Password:  string(hashedPassword),
		Role:      "admin",
		FirstName: &firstName,
		LastName:  &lastName,
	}

	if err := db.Create(&testUser).Error; err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}

	fmt.Printf("✅ Test user created successfully:\n")
	fmt.Printf("   Email: %s\n", testUser.Email)
	fmt.Printf("   Password: admin123\n")
	fmt.Printf("   Role: %s\n", testUser.Role)
	fmt.Printf("   ID: %d\n", testUser.ID)

	fmt.Println("\n💡 You can now use these credentials for API testing:")
	fmt.Println("   Email: admin@example.com")
	fmt.Println("   Password: admin123")
}