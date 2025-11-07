package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	_ = cfg // avoid unused variable warning
	
	// Connect to database
	db := database.ConnectDB()
	
	// Check if user with ID 6 already exists
	var existingUser models.User
	result := db.Where("id = ?", 6).First(&existingUser)
	if result.Error == nil {
		fmt.Printf("User with ID 6 already exists: %s (%s)\n", existingUser.Username, existingUser.Email)
		return
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("salespassword"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	
	// Create new sales person user
	newUser := models.User{
		ID:         6, // Explicitly set ID to 6
		Username:   "salesperson",
		Email:      "salesperson@company.com",
		Password:   string(hashedPassword),
		Role:       "employee", // or "sales" if you have that role
		FirstName:  "Sales",
		LastName:   "Person",
		Phone:      "0812-3456-7890",
		Department: "Sales",
		Position:   "Sales Executive",
		HireDate:   timePointer(time.Now()),
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Insert the user (we need to use raw SQL to set specific ID)
	err = db.Exec(`
		INSERT INTO users (id, username, email, password, role, first_name, last_name, phone, department, position, hire_date, is_active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 
		newUser.ID,
		newUser.Username,
		newUser.Email,
		newUser.Password,
		newUser.Role,
		newUser.FirstName,
		newUser.LastName,
		newUser.Phone,
		newUser.Department,
		newUser.Position,
		newUser.HireDate,
		newUser.IsActive,
		newUser.CreatedAt,
		newUser.UpdatedAt,
	).Error
	
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
	}
	
	fmt.Printf("Successfully created sales person user with ID 6:\n")
	fmt.Printf("Username: %s\n", newUser.Username)
	fmt.Printf("Email: %s\n", newUser.Email)
	fmt.Printf("Password: salespassword (change this in production!)\n")
	fmt.Printf("Role: %s\n", newUser.Role)
	fmt.Printf("Name: %s %s\n", newUser.FirstName, newUser.LastName)
	
	// Update the sequence to avoid conflicts with future auto-incremented users
	err = db.Exec("SELECT setval('users_id_seq', (SELECT MAX(id) FROM users))").Error
	if err != nil {
		log.Printf("Warning: Could not update user sequence: %v", err)
	} else {
		fmt.Println("Updated user ID sequence successfully")
	}
}

func timePointer(t time.Time) *time.Time {
	return &t
}
