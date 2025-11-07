package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := connectDB()

	fmt.Println("=== Starting Smart Notification System Migration ===")
	
	// Run migration
	if err := database.MigrateNotificationConfig(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Check if tables were created
	fmt.Println("\n=== Checking Tables ===")
	checkTable(db, "notification_rules")
	checkTable(db, "notification_preferences")
	checkTable(db, "notification_batches")
	checkTable(db, "notification_queues")

	// Check notification rules
	var rules []models.NotificationRule
	db.Find(&rules)
	fmt.Printf("\n=== Notification Rules Created ===\n")
	fmt.Printf("Total rules: %d\n", len(rules))
	for _, rule := range rules {
		fmt.Printf("- %s (Role: %s, Priority: %s)\n", rule.Name, rule.Role, rule.Priority)
	}

	// Create test users if not exists
	createTestUsers(db)

	fmt.Println("\n✅ Smart Notification System migration completed successfully!")
}

func connectDB() *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func checkTable(db *gorm.DB, tableName string) {
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists).Error
	if err != nil {
		fmt.Printf("❌ Error checking table %s: %v\n", tableName, err)
		return
	}
	if exists {
		fmt.Printf("✅ Table %s exists\n", tableName)
	} else {
		fmt.Printf("❌ Table %s does not exist\n", tableName)
	}
}

func createTestUsers(db *gorm.DB) {
	fmt.Println("\n=== Checking Test Users ===")
	
	testUsers := []models.User{
		{
			Username:  "john_employee",
			Email:     "john@company.com",
			Password:  "$2a$10$YourHashedPasswordHere", // Use proper password hashing
			Role:      "employee",
			FirstName: "John",
			LastName:  "Doe",
			Department: "Sales",
			IsActive:  true,
		},
		{
			Username:  "jack_finance",
			Email:     "jack@company.com",
			Password:  "$2a$10$YourHashedPasswordHere",
			Role:      "finance",
			FirstName: "Jack",
			LastName:  "Smith",
			Department: "Finance",
			IsActive:  true,
		},
		{
			Username:  "josh_director",
			Email:     "josh@company.com",
			Password:  "$2a$10$YourHashedPasswordHere",
			Role:      "director",
			FirstName: "Josh",
			LastName:  "Wilson",
			Department: "Management",
			IsActive:  true,
		},
	}

	for _, user := range testUsers {
		var existing models.User
		err := db.Where("username = ?", user.Username).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&user).Error; err != nil {
				fmt.Printf("❌ Failed to create user %s: %v\n", user.Username, err)
			} else {
				fmt.Printf("✅ Created test user: %s (Role: %s)\n", user.Username, user.Role)
			}
		} else {
			fmt.Printf("ℹ️  User %s already exists (Role: %s)\n", existing.Username, existing.Role)
		}
	}
}
