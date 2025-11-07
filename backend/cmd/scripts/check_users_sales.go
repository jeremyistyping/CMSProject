package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	_ = cfg // avoid unused variable warning
	
	// Connect to database
	db := database.ConnectDB()
	
	// Check users table
	var users []models.User
	result := db.Select("id, username, first_name, last_name, email, role, is_active, created_at").Find(&users)
	if result.Error != nil {
		log.Fatalf("Error fetching users: %v", result.Error)
	}
	
	fmt.Printf("Found %d users:\n", len(users))
	fmt.Println("ID\tUsername\t\tName\t\t\tEmail\t\t\tRole\tActive")
	fmt.Println("--\t--------\t\t----\t\t\t-----\t\t\t----\t------")
	
	for _, user := range users {
		fullName := user.FirstName + " " + user.LastName
		if user.FirstName == "" && user.LastName == "" {
			fullName = user.Username
		}
		fmt.Printf("%d\t%s\t\t%s\t\t%s\t\t%s\t%t\n", 
			user.ID, 
			user.Username,
			fullName,
			user.Email,
			user.Role,
			user.IsActive)
	}
	
	// Check for user with ID 6 specifically
	var user6 models.User
	result = db.Where("id = ?", 6).First(&user6)
	if result.Error != nil {
		fmt.Printf("\nUser with ID 6 NOT FOUND: %v\n", result.Error)
	} else {
		fullName := user6.FirstName + " " + user6.LastName
		if user6.FirstName == "" && user6.LastName == "" {
			fullName = user6.Username
		}
		fmt.Printf("\nUser ID 6 found: %s (%s) - Role: %s, Active: %t\n", 
			fullName, user6.Email, user6.Role, user6.IsActive)
	}
	
	// Check existing sales with sales_person_id
	var sales []models.Sale
	result = db.Select("id, code, sales_person_id").Where("sales_person_id IS NOT NULL").Find(&sales)
	if result.Error != nil {
		log.Printf("Error fetching sales: %v", result.Error)
	} else {
		fmt.Printf("\nFound %d sales with sales_person_id:\n", len(sales))
		for _, sale := range sales {
			fmt.Printf("Sale ID: %d, Code: %s, Sales Person ID: %d\n", 
				sale.ID, sale.Code, *sale.SalesPersonID)
		}
	}
}
