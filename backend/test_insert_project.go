package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Project struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Code        string    `gorm:"unique;not null"`
	Description string
	Status      string    `gorm:"default:'active'"`
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func main() {
	dsn := "host=localhost user=postgres password=Moon dbname=CMSNew port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Check if project 1 exists
	var count int64
	db.Model(&Project{}).Where("id = ?", 1).Count(&count)
	
	if count > 0 {
		fmt.Println("✅ Project with ID 1 already exists")
		
		// Show existing project
		var project Project
		db.First(&project, 1)
		fmt.Printf("   ID: %d, Name: %s, Code: %s\n", project.ID, project.Name, project.Code)
	} else {
		fmt.Println("❌ Project with ID 1 does NOT exist. Creating dummy project...")
		
		// Create dummy project
		project := Project{
			Name:        "Test Project for Milestones",
			Code:        "TEST001",
			Description: "Dummy project for testing milestone functionality",
			Status:      "active",
			StartDate:   time.Now(),
		}
		
		result := db.Create(&project)
		if result.Error != nil {
			log.Fatal("Failed to create project:", result.Error)
		}
		
		fmt.Printf("✅ Created project with ID: %d\n", project.ID)
	}
	
	// List all projects
	fmt.Println("\n📋 All projects in database:")
	var projects []Project
	db.Find(&projects)
	for _, p := range projects {
		fmt.Printf("   ID: %d, Name: %s, Code: %s, Status: %s\n", p.ID, p.Name, p.Code, p.Status)
	}
}

