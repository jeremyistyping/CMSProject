package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	
	// Get all sales
	var sales []models.Sale
	result := db.Select("id, code, type, date, status, created_at").Find(&sales)
	if result.Error != nil {
		log.Fatalf("Error fetching sales: %v", result.Error)
	}
	
	fmt.Printf("Found %d sales records:\n", len(sales))
	fmt.Println("ID\tCode\t\tType\tStatus\tDate")
	fmt.Println("--\t----\t\t----\t------\t----")
	
	// Print all sales
	for _, sale := range sales {
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", 
			sale.ID, 
			sale.Code,
			sale.Type,
			sale.Status, 
			sale.Date.Format("2006-01-02"))
	}
	
	// Check for duplicate codes
	fmt.Println("\nChecking for duplicate codes...")
	codeCount := make(map[string][]uint)
	
	for _, sale := range sales {
		codeCount[sale.Code] = append(codeCount[sale.Code], sale.ID)
	}
	
	duplicatesFound := false
	for code, ids := range codeCount {
		if len(ids) > 1 {
			fmt.Printf("DUPLICATE CODE: %s found in sales IDs: %v\n", code, ids)
			duplicatesFound = true
		}
	}
	
	if !duplicatesFound {
		fmt.Println("No duplicate codes found!")
	}
	
	// Show count by type and year
	fmt.Println("\nSales count by type and year:")
	typeYearCount := make(map[string]map[int]int)
	
	for _, sale := range sales {
		year := sale.Date.Year()
		if typeYearCount[sale.Type] == nil {
			typeYearCount[sale.Type] = make(map[int]int)
		}
		typeYearCount[sale.Type][year]++
	}
	
	for saleType, yearCount := range typeYearCount {
		fmt.Printf("%s:\n", saleType)
		for year, count := range yearCount {
			fmt.Printf("  %d: %d sales\n", year, count)
		}
	}
}
