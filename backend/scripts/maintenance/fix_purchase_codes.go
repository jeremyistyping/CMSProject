package main

import (
	"fmt"
	"log"
	"os"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get database connection string from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	
	// Initialize database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	fmt.Println("=== Checking Purchase Codes ===")
	
	// Check all purchases with code PO/2025/09/0001
	var purchases []models.Purchase
	err = db.Unscoped().Where("code LIKE ?", "PO/2025/09/%").Order("code, id").Find(&purchases).Error
	if err != nil {
		log.Fatal("Error querying purchases:", err)
	}
	
	fmt.Printf("\nFound %d purchases with code pattern PO/2025/09/:\n", len(purchases))
	for _, p := range purchases {
		deletedStatus := ""
		if p.DeletedAt.Valid {
			deletedStatus = fmt.Sprintf(" (DELETED at %s)", p.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("- ID: %d, Code: %s, Created: %s%s\n", 
			p.ID, p.Code, p.CreatedAt.Format("2006-01-02 15:04:05"), deletedStatus)
	}
	
	// Check for duplicates
	var duplicates []struct {
		Code  string
		Count int64
	}
	err = db.Model(&models.Purchase{}).
		Select("code, COUNT(*) as count").
		Where("code LIKE ?", "PO/2025/09/%").
		Group("code").
		Having("COUNT(*) > 1").
		Find(&duplicates).Error
		
	if err != nil {
		log.Fatal("Error checking duplicates:", err)
	}
	
	if len(duplicates) > 0 {
		fmt.Println("\n=== Found Duplicate Codes ===")
		for _, d := range duplicates {
			fmt.Printf("Code %s appears %d times\n", d.Code, d.Count)
			
			// List all instances of this code
			var dupPurchases []models.Purchase
			db.Unscoped().Where("code = ?", d.Code).Order("id").Find(&dupPurchases)
			for i, p := range dupPurchases {
				deletedStatus := "ACTIVE"
				if p.DeletedAt.Valid {
					deletedStatus = "DELETED"
				}
				fmt.Printf("  - ID: %d, Status: %s, Created: %s [%s]\n", 
					p.ID, p.Status, p.CreatedAt.Format("2006-01-02 15:04:05"), deletedStatus)
				
				// Keep the first one, delete others if they're not already deleted
				if i > 0 && !p.DeletedAt.Valid {
					fmt.Printf("    -> Soft deleting duplicate ID %d\n", p.ID)
					db.Delete(&p)
				}
			}
		}
	} else {
		fmt.Println("\nNo duplicate codes found.")
	}
	
	// Now check if PO/2025/09/0001 exists in active records
	var count int64
	err = db.Model(&models.Purchase{}).Where("code = ?", "PO/2025/09/0001").Count(&count).Error
	if err != nil {
		log.Fatal("Error counting:", err)
	}
	
	fmt.Printf("\n=== Final Check ===\n")
	fmt.Printf("Active purchases with code PO/2025/09/0001: %d\n", count)
	
	// If the code exists but we can't find it with our query, there might be an issue
	if count > 0 {
		fmt.Println("\nChecking why GetLastPurchaseNumberByMonth might fail...")
		
		// Try the exact query that GetLastPurchaseNumberByMonth uses
		var purchase models.Purchase
		err = db.Model(&models.Purchase{}).
			Where("code LIKE ?", "PO/2025/09/%").
			Order("code DESC").
			First(&purchase).Error
			
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Println("Query returns 'record not found' - this is the issue!")
				
				// Let's try without soft delete scope
				err = db.Unscoped().Model(&models.Purchase{}).
					Where("code LIKE ? AND deleted_at IS NULL", "PO/2025/09/%").
					Order("code DESC").
					First(&purchase).Error
					
				if err == nil {
					fmt.Printf("Found with Unscoped: %s (ID: %d)\n", purchase.Code, purchase.ID)
				} else {
					fmt.Printf("Still not found with Unscoped: %v\n", err)
				}
			} else {
				fmt.Printf("Query error: %v\n", err)
			}
		} else {
			fmt.Printf("Query successful: Found %s (ID: %d)\n", purchase.Code, purchase.ID)
		}
	}
	
	fmt.Println("\n=== Complete ===")
}
