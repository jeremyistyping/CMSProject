package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== Checking Sales Foreign Key Constraint Issue ===")
	
	// 1. Check contacts table structure and data
	fmt.Println("\n1. Checking contacts table...")
	var contacts []models.Contact
	result := db.Where("type = ?", "EMPLOYEE").Find(&contacts)
	if result.Error != nil {
		fmt.Printf("Error querying contacts: %v\n", result.Error)
	} else {
		fmt.Printf("Found %d employees:\n", len(contacts))
		for _, contact := range contacts {
			fmt.Printf("  ID: %d, Name: %s, Type: %s, Active: %t\n", 
				contact.ID, contact.Name, contact.Type, contact.IsActive)
		}
	}

	// 2. Check if sales_person_id 8 exists
	fmt.Println("\n2. Checking specific sales person ID 8...")
	var contact models.Contact
	result = db.First(&contact, 8)
	if result.Error != nil {
		fmt.Printf("Contact with ID 8 not found: %v\n", result.Error)
		
		// Create a default employee contact
		fmt.Println("Creating default employee contact...")
		defaultEmployee := models.Contact{
			Code:        "EMP001",
			Name:        "Default Sales Person",
			Type:        "EMPLOYEE",
			IsActive:    true,
			Email:       "sales@company.com",
			Phone:       "0123456789",
		}
		
		createResult := db.Create(&defaultEmployee)
		if createResult.Error != nil {
			fmt.Printf("Failed to create default employee: %v\n", createResult.Error)
		} else {
			fmt.Printf("Created default employee with ID: %d\n", defaultEmployee.ID)
		}
	} else {
		fmt.Printf("Found contact ID 8: Name=%s, Type=%s, Active=%t\n", 
			contact.Name, contact.Type, contact.IsActive)
		
		// If found but not an employee, update it
		if contact.Type != "EMPLOYEE" {
			fmt.Println("Updating contact to be an EMPLOYEE...")
			contact.Type = "EMPLOYEE"
			contact.IsActive = true
			updateResult := db.Save(&contact)
			if updateResult.Error != nil {
				fmt.Printf("Failed to update contact: %v\n", updateResult.Error)
			} else {
				fmt.Println("Successfully updated contact to EMPLOYEE type")
			}
		}
	}

	// 3. Check sales table constraints
	fmt.Println("\n3. Checking sales table foreign key constraints...")
	var constraintInfo []struct {
		ConstraintName string `gorm:"column:constraint_name"`
		TableName      string `gorm:"column:table_name"`
		ColumnName     string `gorm:"column:column_name"`
		ForeignTable   string `gorm:"column:foreign_table_name"`
		ForeignColumn  string `gorm:"column:foreign_column_name"`
	}
	
	query := `
		SELECT 
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
			AND tc.table_name = 'sales'
			AND kcu.column_name = 'sales_person_id'
	`
	
	result = db.Raw(query).Scan(&constraintInfo)
	if result.Error != nil {
		fmt.Printf("Error checking constraints: %v\n", result.Error)
	} else {
		fmt.Printf("Found %d constraints:\n", len(constraintInfo))
		for _, constraint := range constraintInfo {
			fmt.Printf("  Constraint: %s, Column: %s -> %s.%s\n", 
				constraint.ConstraintName, constraint.ColumnName, 
				constraint.ForeignTable, constraint.ForeignColumn)
		}
	}

	// 4. Create some sample employee contacts if needed
	fmt.Println("\n4. Ensuring we have sample employees...")
	var employeeCount int64
	db.Model(&models.Contact{}).Where("type = ? AND is_active = ?", "EMPLOYEE", true).Count(&employeeCount)
	
	if employeeCount < 3 {
		fmt.Printf("Only %d active employees found. Creating samples...\n", employeeCount)
		
		sampleEmployees := []models.Contact{
			{
				Code:        "EMP002",
				Name:        "John Sales",
				Type:        "EMPLOYEE",
				IsActive:    true,
				Email:       "john.sales@company.com",
				Phone:       "0123456790",
			},
			{
				Code:        "EMP003", 
				Name:        "Jane Marketing",
				Type:        "EMPLOYEE",
				IsActive:    true,
				Email:       "jane.marketing@company.com",
				Phone:       "0123456791",
			},
		}
		
		for _, emp := range sampleEmployees {
			// Check if already exists
			var existing models.Contact
			if db.Where("code = ?", emp.Code).First(&existing).Error != nil {
				// Doesn't exist, create it
				if createErr := db.Create(&emp).Error; createErr != nil {
					fmt.Printf("Failed to create employee %s: %v\n", emp.Code, createErr)
				} else {
					fmt.Printf("Created employee: %s (ID: %d)\n", emp.Name, emp.ID)
				}
			}
		}
	}

	// 5. List all active employees for reference
	fmt.Println("\n5. Final list of active employees:")
	var activeEmployees []models.Contact
	db.Where("type = ? AND is_active = ?", "EMPLOYEE", true).Find(&activeEmployees)
	
	for _, emp := range activeEmployees {
		fmt.Printf("  ID: %d, Code: %s, Name: %s\n", emp.ID, emp.Code, emp.Name)
	}

	fmt.Println("\n=== Fix Complete ===")
	fmt.Println("You can now use any of the employee IDs listed above as sales_person_id")
}
