package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🧪 Testing Payment Code Display in API Response")

	db := database.ConnectDB()

	// Get recent payments
	var payments []models.Payment
	if err := db.Preload("Contact").Order("id DESC").Limit(10).Find(&payments).Error; err != nil {
		log.Printf("❌ Failed to get payments: %v", err)
		return
	}

	log.Printf("\n💰 Recent Payments (testing JSON field mapping):")
	log.Printf("%-5s %-20s %-20s %-15s %-10s %s", "ID", "Code", "Contact", "Amount", "Status", "Date")
	log.Printf("%s", "================================================================================")
	
	for _, payment := range payments {
		contactName := "Unknown"
		if payment.Contact.ID > 0 {
			contactName = payment.Contact.Name
		}
		
		log.Printf("%-5d %-20s %-20s %15.2f %-10s %s", 
			payment.ID, 
			payment.Code,  // This should now show the payment code
			contactName, 
			payment.Amount, 
			payment.Status,
			payment.Date.Format("2006-01-02"))
	}

	// Test JSON marshaling
	log.Printf("\n🔍 JSON Field Test:")
	if len(payments) > 0 {
		testPayment := payments[0]
		log.Printf("Payment ID: %d", testPayment.ID)
		log.Printf("Payment Code Field: '%s'", testPayment.Code)
		log.Printf("Contact: %s", testPayment.Contact.Name)
		log.Printf("Amount: %.2f", testPayment.Amount)
		log.Printf("Method: %s", testPayment.Method)
		log.Printf("Status: %s", testPayment.Status)
	}

	// Test API endpoint format simulation
	log.Printf("\n📊 Simulating API Response Structure:")
	for i, payment := range payments {
		if i >= 3 { // Only show first 3 for brevity
			break
		}
		
		// This mimics how the API would return the data
		log.Printf("Payment %d:", i+1)
		log.Printf("  id: %d", payment.ID)
		log.Printf("  code: \"%s\"", payment.Code)  // Should now be accessible as 'code'
		log.Printf("  contact_id: %d", payment.ContactID)
		log.Printf("  amount: %.2f", payment.Amount)
		log.Printf("  method: \"%s\"", payment.Method)
		log.Printf("  status: \"%s\"", payment.Status)
		log.Printf("  date: \"%s\"", payment.Date.Format("2006-01-02T15:04:05Z"))
		log.Printf("")
	}

	log.Printf("✅ Payment Code Display Test Completed!")
	log.Printf("📋 Summary: Payment codes should now be accessible via 'code' field in API responses")
}