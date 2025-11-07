package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	db := database.ConnectDB()
	var products []models.Product
	db.Limit(10).Find(&products)
	
	fmt.Println("Existing products in database (first 10):")
	fmt.Println("=========================================")
	
	if len(products) == 0 {
		fmt.Println("No products found in database!")
		return
	}
	
	for _, product := range products {
		fmt.Printf("ID: %d, Name: %s, Code: %s, Active: %t\n", 
			product.ID, product.Name, product.Code, product.IsActive)
	}
}