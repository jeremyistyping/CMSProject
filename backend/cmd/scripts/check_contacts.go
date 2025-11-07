package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	db := database.ConnectDB()
	var contacts []models.Contact
	db.Where("type = ? OR type = ?", "CUSTOMER", "BOTH").Find(&contacts)
	
	fmt.Println("Existing customers in database:")
	fmt.Println("===============================")
	
	if len(contacts) == 0 {
		fmt.Println("No customers found in database!")
		return
	}
	
	for _, contact := range contacts {
		fmt.Printf("ID: %d, Name: %s, Type: %s, Active: %t\n", 
			contact.ID, contact.Name, contact.Type, contact.IsActive)
	}
}