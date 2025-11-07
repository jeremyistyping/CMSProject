package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== Fixing Sales Person Foreign Key Constraint ===")
	
	// 1. Check current constraint
	fmt.Println("\n1. Checking current constraint...")
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
	
	result := db.Raw(query).Scan(&constraintInfo)
	if result.Error != nil {
		fmt.Printf("Error checking constraints: %v\n", result.Error)
		return
	}
	
	fmt.Printf("Found %d constraints:\n", len(constraintInfo))
	for _, constraint := range constraintInfo {
		fmt.Printf("  Constraint: %s, Column: %s -> %s.%s\n", 
			constraint.ConstraintName, constraint.ColumnName, 
			constraint.ForeignTable, constraint.ForeignColumn)
	}
	
	// 2. Check if any sales records exist with sales_person_id
	fmt.Println("\n2. Checking existing sales with sales_person_id...")
	var salesWithSalesPerson int64
	db.Raw("SELECT COUNT(*) FROM sales WHERE sales_person_id IS NOT NULL").Scan(&salesWithSalesPerson)
	fmt.Printf("Found %d sales records with sales_person_id\n", salesWithSalesPerson)
	
	if salesWithSalesPerson > 0 {
		fmt.Println("Warning: There are existing sales records with sales_person_id.")
		fmt.Println("We need to check if these IDs correspond to valid contacts or users.")
		
		// Check if sales_person_ids exist in users table
		var invalidUserIds []uint
		db.Raw(`
			SELECT DISTINCT s.sales_person_id 
			FROM sales s 
			LEFT JOIN users u ON s.sales_person_id = u.id 
			WHERE s.sales_person_id IS NOT NULL AND u.id IS NULL
		`).Scan(&invalidUserIds)
		
		if len(invalidUserIds) > 0 {
			fmt.Printf("Found %d sales records with invalid user IDs: %v\n", len(invalidUserIds), invalidUserIds)
		}
		
		// Check if sales_person_ids exist in contacts table  
		var validContactIds []uint
		db.Raw(`
			SELECT DISTINCT s.sales_person_id 
			FROM sales s 
			JOIN contacts c ON s.sales_person_id = c.id 
			WHERE s.sales_person_id IS NOT NULL AND c.type = 'EMPLOYEE'
		`).Scan(&validContactIds)
		
		if len(validContactIds) > 0 {
			fmt.Printf("Found %d sales records that would have valid contact IDs: %v\n", len(validContactIds), validContactIds)
		}
	}
	
	// 3. Drop existing constraint
	if len(constraintInfo) > 0 {
		constraintName := constraintInfo[0].ConstraintName
		fmt.Printf("\n3. Dropping existing constraint '%s'...\n", constraintName)
		
		err := db.Exec(fmt.Sprintf("ALTER TABLE sales DROP CONSTRAINT IF EXISTS %s", constraintName)).Error
		if err != nil {
			fmt.Printf("Error dropping constraint: %v\n", err)
			return
		} else {
			fmt.Println("Successfully dropped old constraint")
		}
	}
	
	// 4. Add new constraint pointing to contacts table
	fmt.Println("\n4. Adding new foreign key constraint to contacts table...")
	err := db.Exec(`
		ALTER TABLE sales 
		ADD CONSTRAINT fk_sales_sales_person_contact 
		FOREIGN KEY (sales_person_id) REFERENCES contacts(id)
	`).Error
	
	if err != nil {
		fmt.Printf("Error adding new constraint: %v\n", err)
		
		// Try to identify the issue
		fmt.Println("\n   Checking for constraint violation issues...")
		
		// Check for sales records with sales_person_id not in contacts
		var violatingRecords []struct {
			SalesID       uint `gorm:"column:sales_id"`
			SalesPersonID uint `gorm:"column:sales_person_id"`
		}
		
		db.Raw(`
			SELECT s.id as sales_id, s.sales_person_id 
			FROM sales s 
			LEFT JOIN contacts c ON s.sales_person_id = c.id 
			WHERE s.sales_person_id IS NOT NULL AND c.id IS NULL
		`).Scan(&violatingRecords)
		
		if len(violatingRecords) > 0 {
			fmt.Printf("   Found %d sales records with invalid sales_person_id:\n", len(violatingRecords))
			for _, record := range violatingRecords {
				fmt.Printf("     Sales ID: %d, Invalid Sales Person ID: %d\n", record.SalesID, record.SalesPersonID)
			}
			
			// Option 1: Set them to NULL
			fmt.Println("\n   Option 1: Setting invalid sales_person_id to NULL...")
			updateResult := db.Exec(`
				UPDATE sales 
				SET sales_person_id = NULL 
				WHERE id IN (
					SELECT s.id 
					FROM sales s 
					LEFT JOIN contacts c ON s.sales_person_id = c.id 
					WHERE s.sales_person_id IS NOT NULL AND c.id IS NULL
				)
			`)
			
			if updateResult.Error != nil {
				fmt.Printf("   Error updating sales records: %v\n", updateResult.Error)
			} else {
				fmt.Printf("   Updated %d sales records, set sales_person_id to NULL\n", updateResult.RowsAffected)
				
				// Try adding constraint again
				fmt.Println("\n   Retrying to add constraint...")
				err = db.Exec(`
					ALTER TABLE sales 
					ADD CONSTRAINT fk_sales_sales_person_contact 
					FOREIGN KEY (sales_person_id) REFERENCES contacts(id)
				`).Error
				
				if err != nil {
					fmt.Printf("   Still failed to add constraint: %v\n", err)
				} else {
					fmt.Println("   ✅ Successfully added new constraint!")
				}
			}
		}
	} else {
		fmt.Println("✅ Successfully added new foreign key constraint to contacts table!")
	}
	
	// 5. Verify the new constraint
	fmt.Println("\n5. Verifying new constraint...")
	var newConstraintInfo []struct {
		ConstraintName string `gorm:"column:constraint_name"`
		TableName      string `gorm:"column:table_name"`
		ColumnName     string `gorm:"column:column_name"`
		ForeignTable   string `gorm:"column:foreign_table_name"`
		ForeignColumn  string `gorm:"column:foreign_column_name"`
	}
	
	result = db.Raw(query).Scan(&newConstraintInfo)
	if result.Error != nil {
		fmt.Printf("Error checking new constraints: %v\n", result.Error)
	} else {
		fmt.Printf("Current constraints:\n")
		for _, constraint := range newConstraintInfo {
			fmt.Printf("  ✅ Constraint: %s, Column: %s -> %s.%s\n", 
				constraint.ConstraintName, constraint.ColumnName, 
				constraint.ForeignTable, constraint.ForeignColumn)
		}
	}
	
	fmt.Println("\n=== Foreign Key Fix Complete ===")
	fmt.Println("The sales_person_id foreign key now correctly points to contacts.id")
	fmt.Println("You can now create sales with sales_person_id referencing contact IDs (employees)")
}
