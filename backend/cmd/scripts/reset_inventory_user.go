package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Resetting/Restoring INVENTORY user")

	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Find inventory user (including soft-deleted)
	var user models.User
	res := db.Unscoped().Where("username = ? OR email = ?", "inventory", "inventory@company.com").First(&user)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			fmt.Println("➡️  Inventory user not found. Creating new one...")
			hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			user = models.User{
				Username:  "inventory",
				Email:     "inventory@company.com",
				Password:  string(hashed),
				Role:      "inventory_manager",
				FirstName: "Inventory",
				LastName:  "User",
				IsActive:  true,
			}
			if err := db.Create(&user).Error; err != nil {
				log.Fatalf("Failed to create inventory user: %v", err)
			}
			fmt.Printf("✅ Created inventory user. ID=%d Email=%s\n", user.ID, user.Email)
		} else {
			log.Fatalf("Error querying user: %v", res.Error)
		}
	} else {
		fmt.Printf("➡️  Found inventory user. ID=%d, Email=%s\n", user.ID, user.Email)

		// If soft-deleted, restore
		if user.DeletedAt.Valid {
			fmt.Println("♻️  User was soft-deleted. Restoring...")
			// set deleted_at to NULL
			if err := db.Model(&models.User{}).Unscoped().Where("id = ?", user.ID).Update("deleted_at", nil).Error; err != nil {
				log.Fatalf("Failed to restore user: %v", err)
			}
		}

		// Ensure fields
		hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		updates := map[string]interface{}{
			"email":     "inventory@company.com",
			"username":  "inventory",
			"password":  string(hashed),
			"role":      "inventory_manager",
			"first_name":"Inventory",
			"last_name": "User",
			"is_active": true,
		}
		if err := db.Model(&models.User{}).Unscoped().Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			log.Fatalf("Failed to update inventory user: %v", err)
		}
		fmt.Println("✅ Inventory user restored/updated successfully")
	}

	fmt.Println("\nYou can now login with:")
	fmt.Println("Email   : inventory@company.com")
	fmt.Println("Password: password123")
}