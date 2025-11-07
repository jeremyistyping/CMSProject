package main

import (
	"fmt"
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username" gorm:"unique;not null"`
	Email     string `json:"email" gorm:"unique;not null"`
	Password  string `json:"-" gorm:"not null"` // "-" means don't include in JSON
	Role      string `json:"role" gorm:"not null;default:'employee'"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
}

func main() {
	fmt.Println("=== CHECKING LOGIN CREDENTIALS ===")
	
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		dsn = fmt.Sprintf("host=localhost user=postgres password=%s dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta", dbPassword)
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("✅ Database connected successfully")

	// Get all users
	var users []User
	if err := db.Find(&users).Error; err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	fmt.Printf("\n📋 Total users found: %d\n", len(users))
	fmt.Println("================================")
	
	if len(users) == 0 {
		fmt.Println("❌ NO USERS FOUND IN DATABASE!")
		fmt.Println("🔧 You need to run database seeding first:")
		fmt.Println("   go run cmd/main.go --seed")
		return
	}
	
	for i, user := range users {
		fmt.Printf("%d. ID: %d\n", i+1, user.ID)
		fmt.Printf("   Username: '%s'\n", user.Username)
		fmt.Printf("   Email: '%s'\n", user.Email) 
		fmt.Printf("   Role: '%s'\n", user.Role)
		fmt.Printf("   Active: %t\n", user.IsActive)
		fmt.Printf("   Password Hash: %s...\n", user.Password[:20]) // Show first 20 chars of hash
		fmt.Println("   -------------------------")
	}

	// Test password for admin user
	fmt.Println("\n🔍 TESTING PASSWORD FOR ADMIN USER:")
	var adminUser User
	err = db.Where("username = ?", "admin").First(&adminUser).Error
	if err != nil {
		fmt.Printf("❌ No user found with username 'admin': %v\n", err)
		
		// Try to find any user with admin role
		err = db.Where("role = ?", "admin").First(&adminUser).Error
		if err != nil {
			fmt.Printf("❌ No user found with admin role: %v\n", err)
			return
		} else {
			fmt.Printf("✅ Found admin role user: %s\n", adminUser.Username)
		}
	} else {
		fmt.Printf("✅ Found admin user: %s\n", adminUser.Username)
	}

	// Test passwords
	testPasswords := []string{"password123", "admin", "admin123", "password", "123456"}
	
	fmt.Printf("\n🔐 Testing passwords for user '%s':\n", adminUser.Username)
	for _, testPass := range testPasswords {
		err := bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(testPass))
		if err == nil {
			fmt.Printf("✅ CORRECT PASSWORD: '%s'\n", testPass)
			fmt.Printf("\n🎯 USE THESE CREDENTIALS IN SWAGGER:\n")
			fmt.Printf("   Username: %s\n", adminUser.Username)
			fmt.Printf("   Password: %s\n", testPass)
			return
		} else {
			fmt.Printf("❌ Wrong password: '%s'\n", testPass)
		}
	}
	
	fmt.Printf("\n⚠️  None of the common passwords work!\n")
	fmt.Printf("💡 Try resetting password or check if seeding was successful\n")
}