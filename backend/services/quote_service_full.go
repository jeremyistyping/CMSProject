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

type QuoteServiceFull struct {
	db              *gorm.DB
	settingsService *SettingsService
	contactRepo     repositories.ContactRepository
	productRepo     *repositories.ProductRepository
}

type QuoteResult struct {
	Data       []models.Quote `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

func NewQuoteServiceFull(
	db *gorm.DB,
	contactRepo repositories.ContactRepository,
	productRepo *repositories.ProductRepository,
) *QuoteServiceFull {
	settingsService := NewSettingsService(db)
	return &QuoteServiceFull{
		db:              db,
		settingsService: settingsService,
		contactRepo:     contactRepo,
		productRepo:     productRepo,
	}
}

// GenerateQuoteCode generates next quote code using monthly sequence like Purchases
func (s *QuoteServiceFull) GenerateQuoteCode() (string, error) {
	// Get settings for prefix
	_, err := s.settingsService.GetSettings()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %v", err)
	}

	var quoteCode string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock settings for prefix
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil {
			return err
		}

		year := time.Now().Year()
		month := int(time.Now().Month())

		// Auto-migrate sequence table
		if err := tx.AutoMigrate(&InvoiceCodeSequence{}, &QuoteCodeSequence{}); err != nil { return err }

		var seq QuoteCodeSequence
		res := tx.Set("gorm:query_option", "FOR UPDATE").Where("year = ? AND month = ?", year, month).First(&seq)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				seq = QuoteCodeSequence{Year: year, Month: month, LastNumber: 0}
				if err := tx.Create(&seq).Error; err != nil { return err }
			} else { return res.Error }
		}

		seq.LastNumber++
		quoteCode = fmt.Sprintf("%s/%04d/%02d/%04d", settingsForUpdate.QuotePrefix, year, month, seq.LastNumber)

		// Ensure uniqueness
		var existingCount int64
		tx.Model(&models.Quote{}).Where("code = ?", quoteCode).Count(&existingCount)
		if existingCount > 0 { return fmt.Errorf("generated quote code %s already exists", quoteCode) }

		if err := tx.Save(&seq).Error; err != nil { return err }
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate quote code: %v", err)
	}

	fmt.Printf("✅ Generated quote code: %s\n", quoteCode)
	return quoteCode, nil
}

// CreateQuote creates a new quote with code generation
func (s *QuoteServiceFull) CreateQuote(request models.QuoteCreateRequest, userID uint) (*models.Quote, error) {
	// Validate customer exists
	customer, err := s.contactRepo.GetByID(request.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found (ID: %d): %v", request.CustomerID, err)
	}
	fmt.Printf("✅ Customer validation passed: %s (ID: %d)\n", customer.Name, customer.ID)

	// Generate quote code using settings
	code, err := s.GenerateQuoteCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate quote code: %v", err)
	}

	// Get settings for tax calculations
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for calculations: %v", err)
	}

	// Create quote entity
	quote := &models.Quote{
		Code:               code,
		CustomerID:         request.CustomerID,
		UserID:             userID,
		Date:               request.Date,
		ValidUntil:         request.ValidUntil,
		Discount:           request.Discount,
		PPNRate:            request.PPNRate,
		PPh21Rate:          request.PPh21Rate,
		PPh23Rate:          request.PPh23Rate,
		OtherTaxAdditions:  request.OtherTaxAdditions,
		OtherTaxDeductions: request.OtherTaxDeductions,
		Status:             models.QuoteStatusDraft,
		Notes:              request.Notes,
		Terms:              request.Terms,
		ConvertedToInvoice: false,
	}

	// Calculate totals and create quote items
	fmt.Printf("ℹ Calculating quote totals for %d items\n", len(request.Items))
	err = s.calculateQuoteTotals(quote, request.Items, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quote totals: %v", err)
	}
	fmt.Printf("✅ Quote totals calculated: Subtotal=%.2f, Tax=%.2f, Total=%.2f\n", 
		quote.SubtotalBeforeDiscount, quote.TaxAmount, quote.TotalAmount)

	// Save quote
	fmt.Printf("ℹ Saving quote to database\n")
	if err := s.db.Create(quote).Error; err != nil {
		return nil, fmt.Errorf("failed to save quote to database: %v", err)
	}
	fmt.Printf("✅ Quote %d saved successfully with code %s\n", quote.ID, quote.Code)

	return s.GetQuoteByID(quote.ID)
}

// GetQuotes returns paginated list of quotes with filters
func (s *QuoteServiceFull) GetQuotes(filter models.QuoteFilter) (*QuoteResult, error) {
	var quotes []models.Quote
	var total int64

	query := s.db.Model(&models.Quote{}).Preload("Customer").Preload("User").Preload("QuoteItems.Product")

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
		return nil, fmt.Errorf("failed to count quotes: %v", err)
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&quotes).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve quotes: %v", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &QuoteResult{
		Data:       quotes,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetQuoteByID returns a single quote by ID
func (s *QuoteServiceFull) GetQuoteByID(id uint) (*models.Quote, error) {
	var quote models.Quote
	if err := s.db.Preload("Customer").Preload("User").Preload("QuoteItems.Product").First(&quote, id).Error; err != nil {
		return nil, fmt.Errorf("quote not found (ID: %d): %v", id, err)
	}
	return &quote, nil
}

// UpdateQuote updates an existing quote
func (s *QuoteServiceFull) UpdateQuote(id uint, request models.QuoteUpdateRequest, userID uint) (*models.Quote, error) {
	quote, err := s.GetQuoteByID(id)
	if err != nil {
		return nil, err
	}

	// Check if quote can be updated
	if quote.Status != models.QuoteStatusDraft {
		return nil, fmt.Errorf("quote cannot be updated in current status: %s", quote.Status)
	}

	// Get settings for tax calculations
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for calculations: %v", err)
	}

	// Update fields if provided
	if request.CustomerID != nil {
		quote.CustomerID = *request.CustomerID
	}
	if request.Date != nil {
		quote.Date = *request.Date
	}
	if request.ValidUntil != nil {
		quote.ValidUntil = *request.ValidUntil
	}
	if request.Discount != nil {
		quote.Discount = *request.Discount
	}
	if request.Notes != nil {
		quote.Notes = *request.Notes
	}
	if request.Terms != nil {
		quote.Terms = *request.Terms
	}

	// Update items if provided
	if len(request.Items) > 0 {
		// Delete existing items
		s.db.Where("quote_id = ?", quote.ID).Delete(&models.QuoteItem{})
		
		// Recalculate with new items
		itemRequests := make([]models.QuoteItemCreateRequest, len(request.Items))
		copy(itemRequests, request.Items)
		
		err = s.calculateQuoteTotals(quote, itemRequests, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to recalculate quote totals: %v", err)
		}
	}

	// Save updated quote
	if err := s.db.Save(quote).Error; err != nil {
		return nil, fmt.Errorf("failed to update quote: %v", err)
	}

	return s.GetQuoteByID(quote.ID)
}

// DeleteQuote deletes a quote
func (s *QuoteServiceFull) DeleteQuote(id uint) error {
	quote, err := s.GetQuoteByID(id)
	if err != nil {
		return err
	}

	// Check if quote can be deleted
	if quote.ConvertedToInvoice {
		return fmt.Errorf("cannot delete quote that has been converted to invoice")
	}

	// Delete quote (cascade will delete items)
	if err := s.db.Delete(quote).Error; err != nil {
		return fmt.Errorf("failed to delete quote: %v", err)
	}

	return nil
}

// ConvertToInvoice converts a quote to an invoice
func (s *QuoteServiceFull) ConvertToInvoice(quoteID uint, userID uint, invoiceService *InvoiceServiceFull) (*models.Invoice, error) {
	quote, err := s.GetQuoteByID(quoteID)
	if err != nil {
		return nil, err
	}

	// Check if quote can be converted
	if quote.Status != models.QuoteStatusAccepted {
		return nil, fmt.Errorf("only accepted quotes can be converted to invoices")
	}

	if quote.ConvertedToInvoice {
		return nil, fmt.Errorf("quote has already been converted to an invoice")
	}

	// Create invoice request from quote
	invoiceRequest := models.InvoiceCreateRequest{
		CustomerID:         quote.CustomerID,
		Date:               time.Now(),
		DueDate:            time.Now().AddDate(0, 0, 30), // Default 30 days
		Discount:           quote.Discount,
		PaymentMethod:      models.InvoicePaymentCredit,
		PPNRate:            quote.PPNRate,
		PPh21Rate:          quote.PPh21Rate,
		PPh23Rate:          quote.PPh23Rate,
		OtherTaxAdditions:  quote.OtherTaxAdditions,
		OtherTaxDeductions: quote.OtherTaxDeductions,
		Notes:              fmt.Sprintf("Converted from Quote %s", quote.Code),
		Items:              make([]models.InvoiceItemCreateRequest, 0),
	}

	// Convert quote items to invoice items
	for _, quoteItem := range quote.QuoteItems {
		invoiceItem := models.InvoiceItemCreateRequest{
			ProductID:   quoteItem.ProductID,
			Quantity:    quoteItem.Quantity,
			UnitPrice:   quoteItem.UnitPrice,
			Description: quoteItem.Description,
		}
		invoiceRequest.Items = append(invoiceRequest.Items, invoiceItem)
	}

	// Create invoice
	invoice, err := invoiceService.CreateInvoice(invoiceRequest, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice from quote: %v", err)
	}

	// Update quote to mark as converted
	quote.ConvertedToInvoice = true
	quote.InvoiceID = &invoice.ID
	if err := s.db.Save(quote).Error; err != nil {
		return nil, fmt.Errorf("failed to update quote conversion status: %v", err)
	}

	return invoice, nil
}

// FormatCurrency formats amount according to system settings
func (s *QuoteServiceFull) FormatCurrency(amount float64) (string, error) {
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
func (s *QuoteServiceFull) FormatDate(date time.Time) (string, error) {
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

func (s *QuoteServiceFull) calculateQuoteTotals(quote *models.Quote, items []models.QuoteItemCreateRequest, settings *models.Settings) error {
	var subtotal float64 = 0
	
	// Clear existing items
	quote.QuoteItems = []models.QuoteItem{}
	
	for _, itemReq := range items {
		// Validate product exists
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product not found (ID: %d): %v", itemReq.ProductID, err)
		}
		
		totalPrice := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotal += totalPrice
		
		// Create quote item
		item := models.QuoteItem{
			ProductID:   itemReq.ProductID,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
			TotalPrice:  totalPrice,
			Description: itemReq.Description,
		}
		
		quote.QuoteItems = append(quote.QuoteItems, item)
		fmt.Printf("📝 Added item: %s (Qty: %d, Price: %.2f, Total: %.2f)\n", product.Name, itemReq.Quantity, itemReq.UnitPrice, totalPrice)
	}
	
	// Calculate amounts
	quote.SubtotalBeforeDiscount = subtotal
	quote.SubtotalAfterDiscount = subtotal - quote.Discount
	
	// Calculate taxes using settings default rate if not specified
	var taxAmount float64 = 0
	
	if quote.PPNRate != nil {
		taxAmount += quote.SubtotalAfterDiscount * (*quote.PPNRate / 100)
	} else {
		// Use default tax rate from settings
		taxAmount += quote.SubtotalAfterDiscount * (settings.DefaultTaxRate / 100)
	}
	
	if quote.OtherTaxAdditions != nil {
		taxAmount += *quote.OtherTaxAdditions
	}
	
	if quote.PPh21Rate != nil {
		taxAmount -= quote.SubtotalAfterDiscount * (*quote.PPh21Rate / 100)
	}
	
	if quote.PPh23Rate != nil {
		taxAmount -= quote.SubtotalAfterDiscount * (*quote.PPh23Rate / 100)
	}
	
	if quote.OtherTaxDeductions != nil {
		taxAmount -= *quote.OtherTaxDeductions
	}
	
	quote.TaxAmount = taxAmount
	quote.TotalAmount = quote.SubtotalAfterDiscount + taxAmount
	
	return nil
}