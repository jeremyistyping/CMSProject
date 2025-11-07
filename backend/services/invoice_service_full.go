package services

import (
	"errors"
	"fmt"
	"time"
	"math"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

type InvoiceServiceFull struct {
	db              *gorm.DB
	settingsService *SettingsService
	contactRepo     repositories.ContactRepository
	productRepo     *repositories.ProductRepository
}

// InvoiceCodeSequence and QuoteCodeSequence are used to generate monthly sequential codes
// Table will be auto-migrated if not exists
// Composite key: year + month
// LastNumber stores the last used running number for that period
// Exported so other services can reuse
// Note: mirrors the approach used in Purchase code generation

type InvoiceCodeSequence struct {
	Year      int `gorm:"primaryKey"`
	Month     int `gorm:"primaryKey"`
	LastNumber int `gorm:"default:0"`
}

type QuoteCodeSequence struct {
	Year      int `gorm:"primaryKey"`
	Month     int `gorm:"primaryKey"`
	LastNumber int `gorm:"default:0"`
}

type InvoiceResult struct {
	Data       []models.Invoice `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

func NewInvoiceServiceFull(
	db *gorm.DB,
	contactRepo repositories.ContactRepository,
	productRepo *repositories.ProductRepository,
) *InvoiceServiceFull {
	settingsService := NewSettingsService(db)
	return &InvoiceServiceFull{
		db:              db,
		settingsService: settingsService,
		contactRepo:     contactRepo,
		productRepo:     productRepo,
	}
}

// GenerateInvoiceCode generates next invoice code using monthly sequence like Purchases
func (s *InvoiceServiceFull) GenerateInvoiceCode() (string, error) {
	// Ensure settings available (for prefix)
	_, err := s.settingsService.GetSettings()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %v", err)
	}

	var invoiceCode string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock settings for reading prefix and to keep transaction stable
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil {
			return err
		}

		// Monthly sequence using a dedicated table similar to Purchases
		year := time.Now().Year()
		month := int(time.Now().Month())

		// Auto-migrate sequence table
		if err := tx.AutoMigrate(&InvoiceCodeSequence{}); err != nil {
			return err
		}

		var seq InvoiceCodeSequence
		res := tx.Set("gorm:query_option", "FOR UPDATE").Where("year = ? AND month = ?", year, month).First(&seq)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				seq = InvoiceCodeSequence{Year: year, Month: month, LastNumber: 0}
				if err := tx.Create(&seq).Error; err != nil { return err }
			} else {
				return res.Error
			}
		}

		seq.LastNumber++
		invoiceCode = fmt.Sprintf("%s/%04d/%02d/%04d", settingsForUpdate.InvoicePrefix, year, month, seq.LastNumber)

		// Ensure uniqueness
		var existingCount int64
		tx.Model(&models.Invoice{}).Where("code = ?", invoiceCode).Count(&existingCount)
		if existingCount > 0 {
			return fmt.Errorf("generated invoice code %s already exists", invoiceCode)
		}

		// Save updated sequence
		if err := tx.Save(&seq).Error; err != nil { return err }
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate invoice code: %v", err)
	}

	fmt.Printf("✅ Generated invoice code: %s\n", invoiceCode)
	return invoiceCode, nil
}

// CreateInvoice creates a new invoice with code generation
func (s *InvoiceServiceFull) CreateInvoice(request models.InvoiceCreateRequest, userID uint) (*models.Invoice, error) {
	// Validate customer exists
	customer, err := s.contactRepo.GetByID(request.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found (ID: %d): %v", request.CustomerID, err)
	}
	fmt.Printf("✅ Customer validation passed: %s (ID: %d)\n", customer.Name, customer.ID)

	// Generate invoice code using settings
	code, err := s.GenerateInvoiceCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice code: %v", err)
	}

	// Get settings for tax calculations
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for calculations: %v", err)
	}

	// Create invoice entity
	invoice := &models.Invoice{
		Code:              code,
		CustomerID:        request.CustomerID,
		UserID:            userID,
		Date:              request.Date,
		DueDate:           request.DueDate,
		Discount:          request.Discount,
		PaymentMethod:     getInvoicePaymentMethod(request.PaymentMethod),
		PaymentReference:  request.PaymentReference,
		BankAccountID:     request.BankAccountID,
		PPNRate:           request.PPNRate,
		PPh21Rate:         request.PPh21Rate,
		PPh23Rate:         request.PPh23Rate,
		OtherTaxAdditions: request.OtherTaxAdditions,
		OtherTaxDeductions: request.OtherTaxDeductions,
		Status:            models.InvoiceStatusDraft,
		Notes:             request.Notes,
		PaidAmount:        0,
		OutstandingAmount: 0, // Will be set after total calculation
	}

	// Calculate totals and create invoice items
	fmt.Printf("ℹ Calculating invoice totals for %d items\n", len(request.Items))
	err = s.calculateInvoiceTotals(invoice, request.Items, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate invoice totals: %v", err)
	}
	fmt.Printf("✅ Invoice totals calculated: Subtotal=%.2f, Tax=%.2f, Total=%.2f\n", 
		invoice.SubtotalBeforeDiscount, invoice.TaxAmount, invoice.TotalAmount)

	// Save invoice
	fmt.Printf("ℹ Saving invoice to database\n")
	if err := s.db.Create(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to save invoice to database: %v", err)
	}
	fmt.Printf("✅ Invoice %d saved successfully with code %s\n", invoice.ID, invoice.Code)

	return s.GetInvoiceByID(invoice.ID)
}

// GetInvoices returns paginated list of invoices with filters
func (s *InvoiceServiceFull) GetInvoices(filter models.InvoiceFilter) (*InvoiceResult, error) {
	var invoices []models.Invoice
	var total int64

	query := s.db.Model(&models.Invoice{}).Preload("Customer").Preload("User").Preload("InvoiceItems.Product")

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CustomerID != "" {
		query = query.Where("customer_id = ?", filter.CustomerID)
	}
	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		query = query.Where("code ILIKE ? OR notes ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count invoices: %v", err)
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve invoices: %v", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &InvoiceResult{
		Data:       invoices,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetInvoiceByID returns a single invoice by ID
func (s *InvoiceServiceFull) GetInvoiceByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := s.db.Preload("Customer").Preload("User").Preload("InvoiceItems.Product").First(&invoice, id).Error; err != nil {
		return nil, fmt.Errorf("invoice not found (ID: %d): %v", id, err)
	}
	return &invoice, nil
}

// UpdateInvoice updates an existing invoice
func (s *InvoiceServiceFull) UpdateInvoice(id uint, request models.InvoiceUpdateRequest, userID uint) (*models.Invoice, error) {
	invoice, err := s.GetInvoiceByID(id)
	if err != nil {
		return nil, err
	}

	// Check if invoice can be updated
	if invoice.Status != models.InvoiceStatusDraft {
		return nil, fmt.Errorf("invoice cannot be updated in current status: %s", invoice.Status)
	}

	// Get settings for tax calculations
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for calculations: %v", err)
	}

	// Update fields if provided
	if request.CustomerID != nil {
		invoice.CustomerID = *request.CustomerID
	}
	if request.Date != nil {
		invoice.Date = *request.Date
	}
	if request.DueDate != nil {
		invoice.DueDate = *request.DueDate
	}
	if request.Discount != nil {
		invoice.Discount = *request.Discount
	}
	if request.PaymentMethod != nil {
		invoice.PaymentMethod = getInvoicePaymentMethod(*request.PaymentMethod)
	}
	if request.Notes != nil {
		invoice.Notes = *request.Notes
	}

	// Update items if provided
	if len(request.Items) > 0 {
		// Delete existing items
		s.db.Where("invoice_id = ?", invoice.ID).Delete(&models.InvoiceItem{})
		
		// Recalculate with new items
		itemRequests := make([]models.InvoiceItemCreateRequest, len(request.Items))
		copy(itemRequests, request.Items)
		
		err = s.calculateInvoiceTotals(invoice, itemRequests, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to recalculate invoice totals: %v", err)
		}
	}

	// Save updated invoice
	if err := s.db.Save(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to update invoice: %v", err)
	}

	return s.GetInvoiceByID(invoice.ID)
}

// DeleteInvoice deletes an invoice
func (s *InvoiceServiceFull) DeleteInvoice(id uint) error {
	invoice, err := s.GetInvoiceByID(id)
	if err != nil {
		return err
	}

	// Check if invoice can be deleted
	if invoice.Status == models.InvoiceStatusPaid || invoice.Status == models.InvoiceStatusPartially {
		return fmt.Errorf("cannot delete invoice with status: %s", invoice.Status)
	}

	// Delete invoice (cascade will delete items)
	if err := s.db.Delete(invoice).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %v", err)
	}

	return nil
}

// FormatCurrency formats amount according to system settings
func (s *InvoiceServiceFull) FormatCurrency(amount float64) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	// Format with decimal places from settings
	formatStr := fmt.Sprintf("%%.%df", settings.DecimalPlaces)
	formatted := fmt.Sprintf(formatStr, amount)
	
	return fmt.Sprintf("%s %s", settings.Currency, formatted), nil
}

// FormatDate formats date according to system settings
func (s *InvoiceServiceFull) FormatDate(date time.Time) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	switch settings.DateFormat {
	case "DD/MM/YYYY":
		return date.Format("02/01/2006"), nil
	case "MM/DD/YYYY":
		return date.Format("01/02/2006"), nil
	case "DD-MM-YYYY":
		return date.Format("02-01-2006"), nil
	case "YYYY-MM-DD":
		return date.Format("2006-01-02"), nil
	default:
		return date.Format("02/01/2006"), nil // Default to DD/MM/YYYY
	}
}

// Private helper methods

func (s *InvoiceServiceFull) calculateInvoiceTotals(invoice *models.Invoice, items []models.InvoiceItemCreateRequest, settings *models.Settings) error {
	var subtotal float64 = 0
	
	// Clear existing items
	invoice.InvoiceItems = []models.InvoiceItem{}
	
	for _, itemReq := range items {
		// Validate product exists
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product not found (ID: %d): %v", itemReq.ProductID, err)
		}
		
		totalPrice := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotal += totalPrice
		
		// Create invoice item
		item := models.InvoiceItem{
			ProductID:   itemReq.ProductID,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
			TotalPrice:  totalPrice,
			Description: itemReq.Description,
		}
		
		invoice.InvoiceItems = append(invoice.InvoiceItems, item)
		fmt.Printf("📝 Added item: %s (Qty: %d, Price: %.2f, Total: %.2f)\n", product.Name, itemReq.Quantity, itemReq.UnitPrice, totalPrice)
	}
	
	// Calculate amounts
	invoice.SubtotalBeforeDiscount = subtotal
	invoice.SubtotalAfterDiscount = subtotal - invoice.Discount
	
	// Calculate taxes using settings default rate if not specified
	var taxAmount float64 = 0
	
	if invoice.PPNRate != nil {
		taxAmount += invoice.SubtotalAfterDiscount * (*invoice.PPNRate / 100)
	} else {
		// Use default tax rate from settings
		taxAmount += invoice.SubtotalAfterDiscount * (settings.DefaultTaxRate / 100)
	}
	
	if invoice.OtherTaxAdditions != nil {
		taxAmount += *invoice.OtherTaxAdditions
	}
	
	if invoice.PPh21Rate != nil {
		taxAmount -= invoice.SubtotalAfterDiscount * (*invoice.PPh21Rate / 100)
	}
	
	if invoice.PPh23Rate != nil {
		taxAmount -= invoice.SubtotalAfterDiscount * (*invoice.PPh23Rate / 100)
	}
	
	if invoice.OtherTaxDeductions != nil {
		taxAmount -= *invoice.OtherTaxDeductions
	}
	
	invoice.TaxAmount = taxAmount
	invoice.TotalAmount = invoice.SubtotalAfterDiscount + taxAmount
	invoice.OutstandingAmount = invoice.TotalAmount // Initially all outstanding
	
	return nil
}

func getInvoicePaymentMethod(method string) string {
	if method == "" {
		return models.InvoicePaymentCredit
	}
	return method
}