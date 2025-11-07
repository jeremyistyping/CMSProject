package database

import (
	"log"
	"strings"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// SeedContacts seeds initial contact data
func SeedContacts(db *gorm.DB) {
	log.Println("Seeding contacts...")

	contacts := []models.Contact{
		{
			Code:         "CUST-0001",
			Name:         "PT Maju Jaya",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryRetail,
			Email:        "info@majujaya.com",
			Phone:        "+62-21-5551234",
			Mobile:       "+62-812-3456789",
			Website:      "www.majujaya.com",
			TaxNumber:    "01.234.567.8-901.000",
			CreditLimit:  50000000,
			PaymentTerms: 30,
			IsActive:     true,
			PICName:      "Budi Santoso",
			ExternalID:   "CUST-001",
			Address:      "Jl. Sudirman No. 123, Menteng, Jakarta Pusat 10220, DKI Jakarta, Indonesia",
			Notes:        "Customer utama untuk produk elektronik",
		},
		{
			Code:         "VEND-0001",
			Name:         "CV Sumber Rejeki",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryWholesale,
			Email:        "sales@sumberrejeki.co.id",
			Phone:        "+62-21-5555678",
			Mobile:       "+62-813-9876543",
			Website:      "www.sumberrejeki.co.id",
			TaxNumber:    "02.345.678.9-012.000",
			CreditLimit:  0,
			PaymentTerms: 14,
			IsActive:     true,
			PICName:      "Sari Wulandari",
			ExternalID:   "VEND-001",
			Address:      "Jl. Gatot Subroto No. 456, Setiabudi, Jakarta Selatan 12950, DKI Jakarta, Indonesia",
			Notes:        "Supplier utama untuk bahan baku",
		},
		{
			Code:         "EMP-0001",
			Name:         "Ahmad Subandi",
			Type:         models.ContactTypeEmployee,
			Email:        "ahmad.subandi@company.com",
			Phone:        "+62-21-5557890",
			Mobile:       "+62-812-3456789",
			CreditLimit:  0,
			PaymentTerms: 0,
			IsActive:     true,
			ExternalID:   "EMP-001",
			Address:      "Jl. Kebon Jeruk No. 789, Kebon Jeruk, Jakarta Barat 11530, DKI Jakarta, Indonesia",
			Notes:        "Manager Keuangan",
		},
		{
			Code:         "CUST-0002",
			Name:         "PT Global Tech",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryWholesale,
			Email:        "contact@globaltech.id",
			Phone:        "+62-21-7771234",
			Mobile:       "+62-815-1234567",
			Website:      "www.globaltech.id",
			TaxNumber:    "03.456.789.0-123.000",
			CreditLimit:  75000000,
			PaymentTerms: 45,
			IsActive:     true,
			PICName:      "Andi Wijaya",
			ExternalID:   "CUST-002",
			Address:      "Jl. HR Rasuna Said No. 321, Setiabudi, Jakarta Selatan 12940, DKI Jakarta, Indonesia",
			Notes:        "Customer wholesale teknologi",
		},
		{
			Code:         "VEND-0002",
			Name:         "Toko Elektronik Sejati",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryRetail,
			Email:        "admin@elektroniksejati.com",
			Phone:        "+62-21-6661111",
			Mobile:       "+62-814-5678901",
			TaxNumber:    "04.567.890.1-234.000",
			CreditLimit:  0,
			PaymentTerms: 7,
			IsActive:     true,
			PICName:      "Rudi Hermawan",
			ExternalID:   "VEND-002",
			Address:      "Jl. Mangga Besar No. 88, Tamansari, Jakarta Barat 11150, DKI Jakarta, Indonesia",
			Notes:        "Supplier komponen elektronik",
		},
		{
			Code:         "EMP-0002",
			Name:         "Siti Nurhaliza",
			Type:         models.ContactTypeEmployee,
			Email:        "siti.nurhaliza@company.com",
			Phone:        "+62-21-5559999",
			Mobile:       "+62-813-9876543",
			CreditLimit:  0,
			PaymentTerms: 0,
			IsActive:     true,
			ExternalID:   "EMP-002",
			Address:      "Jl. Cempaka Putih No. 55, Cempaka Putih, Jakarta Pusat 10570, DKI Jakarta, Indonesia",
			Notes:        "Supervisor Inventory",
		},
		{
			Code:         "CUST-0003",
			Name:         "PT Sejahtera Mandiri",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryDistributor,
			Email:        "contact@sejahteramandiri.com",
			Phone:        "+62-21-8881234",
			Mobile:       "+62-816-2345678",
			Website:      "www.sejahteramandiri.com",
			TaxNumber:    "05.678.901.2-345.000",
			CreditLimit:  100000000,
			PaymentTerms: 60,
			IsActive:     true,
			PICName:      "Dewi Sartika",
			ExternalID:   "CUST-003",
			Address:      "Jl. MH Thamrin No. 88, Tanah Abang, Jakarta Pusat 10230, DKI Jakarta, Indonesia",
			Notes:        "Distributor utama wilayah Jakarta",
		},
		{
			Code:         "VEND-0003",
			Name:         "UD Berkah Jaya",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryManufacturer,
			Email:        "info@berkahjaya.com",
			Phone:        "+62-21-2345678",
			Mobile:       "+62-817-3456789",
			TaxNumber:    "06.789.012.3-456.000",
			CreditLimit:  0,
			PaymentTerms: 21,
			IsActive:     true,
			PICName:      "Bambang Sutrisno",
			ExternalID:   "VEND-003",
			Address:      "Jl. Jend Sudirman No. 200, Karet Semanggi, Jakarta Selatan 12930, DKI Jakarta, Indonesia",
			Notes:        "Manufaktur spare parts",
		},
	}

	// Seed addresses for contacts
	addresses := []models.ContactAddress{
		{
			ContactID:  1, // PT Maju Jaya
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Sudirman No. 123",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10220",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  1, // PT Maju Jaya
			Type:       models.AddressTypeShipping,
			Address1:   "Jl. Sudirman No. 123",
			Address2:   "Gudang Belakang",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10220",
			Country:    "Indonesia",
			IsDefault:  false,
		},
		{
			ContactID:  2, // CV Sumber Rejeki
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Gatot Subroto No. 456",
			City:       "Jakarta Selatan",
			State:      "DKI Jakarta",
			PostalCode: "12950",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  3, // Ahmad Subandi
			Type:       models.AddressTypeMailing,
			Address1:   "Jl. Kebon Jeruk No. 789",
			City:       "Jakarta Barat",
			State:      "DKI Jakarta",
			PostalCode: "11530",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  4, // PT Global Tech
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. HR Rasuna Said No. 321",
			City:       "Jakarta Selatan",
			State:      "DKI Jakarta",
			PostalCode: "12940",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  5, // Toko Elektronik Sejati
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Mangga Besar No. 88",
			City:       "Jakarta Barat",
			State:      "DKI Jakarta",
			PostalCode: "11150",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  6, // Siti Nurhaliza
			Type:       models.AddressTypeMailing,
			Address1:   "Jl. Cempaka Putih No. 55",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10570",
			Country:    "Indonesia",
			IsDefault:  true,
		},
	}

	// Check if specific seed contacts already exist
	var seedContactExists int64
	db.Model(&models.Contact{}).Where("code IN ?", []string{"CUST-0001", "VEND-0001", "EMP-0001"}).Count(&seedContactExists)
	
	if seedContactExists == 0 {
	// Use native PostgreSQL UPSERT to avoid all ERROR logs
	successCount := 0
	for _, contact := range contacts {
		// Use ON CONFLICT DO NOTHING to skip existing contacts silently
		query := `
			INSERT INTO contacts (
				code, name, type, category, email, phone, mobile, fax, website, 
				tax_number, credit_limit, payment_terms, is_active, pic_name, 
				external_id, address, notes, created_at, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON CONFLICT (code) DO NOTHING
		`
		
		result := db.Exec(query,
			contact.Code, contact.Name, contact.Type, contact.Category,
			contact.Email, contact.Phone, contact.Mobile, contact.Fax,
			contact.Website, contact.TaxNumber, contact.CreditLimit,
			contact.PaymentTerms, contact.IsActive, contact.PICName,
			contact.ExternalID, contact.Address, contact.Notes,
		)
		
		if result.Error != nil {
			log.Printf("Error seeding contact %s: %v", contact.Code, result.Error)
		} else if result.RowsAffected > 0 {
			successCount++
		}
	}
		log.Printf("Successfully seeded %d out of %d contacts", successCount, len(contacts))

		// Create addresses one by one to handle duplicates gracefully
		addressSuccessCount := 0
		for _, address := range addresses {
			if err := db.Create(&address).Error; err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					log.Printf("Address for contact %d already exists, skipping", address.ContactID)
				} else {
					log.Printf("Error seeding address for contact %d: %v", address.ContactID, err)
				}
			} else {
				addressSuccessCount++
			}
		}
		log.Printf("Successfully seeded %d out of %d contact addresses", addressSuccessCount, len(addresses))
	} else {
		log.Printf("Seed contacts already exist (%d records), skipping seed", seedContactExists)
	}
}
