package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"strings"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentValidationMiddleware struct {
	db                   *gorm.DB
	contactRepo          repositories.ContactRepository
	cashBankRepo         *repositories.CashBankRepository
	salesRepo            *repositories.SalesRepository
	purchaseRepo         *repositories.PurchaseRepository
	paymentRepo          *repositories.PaymentRepository
	maxPaymentAmount     float64
	allowNegativeBalance bool
	requireReference     bool
}

type ValidationConfig struct {
	MaxPaymentAmount     float64 `json:"max_payment_amount"`
	AllowNegativeBalance bool    `json:"allow_negative_balance"`
	RequireReference     bool    `json:"require_reference"`
}

func NewPaymentValidationMiddleware(
	db *gorm.DB,
	contactRepo repositories.ContactRepository,
	cashBankRepo *repositories.CashBankRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	paymentRepo *repositories.PaymentRepository,
	config ValidationConfig,
) *PaymentValidationMiddleware {
	// Set defaults if not provided
	if config.MaxPaymentAmount == 0 {
		config.MaxPaymentAmount = 1000000000 // 1 billion default limit
	}

	return &PaymentValidationMiddleware{
		db:                   db,
		contactRepo:          contactRepo,
		cashBankRepo:         cashBankRepo,
		salesRepo:            salesRepo,
		purchaseRepo:         purchaseRepo,
		paymentRepo:          paymentRepo,
		maxPaymentAmount:     config.MaxPaymentAmount,
		allowNegativeBalance: config.AllowNegativeBalance,
		requireReference:     config.RequireReference,
	}
}

// 🛡️ MAIN VALIDATION MIDDLEWARE
func (m *PaymentValidationMiddleware) ValidatePaymentRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JSON body
		var requestData map[string]interface{}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			m.respondWithError(c, "INVALID_JSON", "Request body is not valid JSON", err)
			return
		}

		// Re-bind the body for the next handler
		c.Set("request_data", requestData)

		// Run comprehensive validation
		if err := m.runComprehensiveValidation(c, requestData); err != nil {
			return
		}

		c.Next()
	}
}

// 🔍 Run comprehensive validation checks
func (m *PaymentValidationMiddleware) runComprehensiveValidation(c *gin.Context, data map[string]interface{}) error {
	validationResults := &ValidationResults{
		Checks:   make([]ValidationCheck, 0),
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}

	// 1. Basic field validation
	if err := m.validateBasicFields(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 2. Contact validation
	if err := m.validateContact(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 3. Cash/Bank validation
	if err := m.validateCashBank(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 4. Payment method consistency validation
	if err := m.validatePaymentMethodConsistency(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 5. Target allocation validation
	if err := m.validateTargetAllocation(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 6. Balance and limit validation
	if err := m.validateBalanceAndLimits(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 7. Business logic validation
	if err := m.validateBusinessLogic(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// 8. Duplicate payment prevention
	if err := m.validateDuplicatePayment(data, validationResults); err != nil {
		m.respondWithValidationResults(c, validationResults)
		return err
	}

	// Set validation results in context for later use
	c.Set("validation_results", validationResults)

	// If we have warnings but no errors, include them in response header
	if len(validationResults.Warnings) > 0 && len(validationResults.Errors) == 0 {
		c.Header("X-Payment-Validation-Warnings", strings.Join(validationResults.Warnings, "; "))
	}

	return nil
}

// 📋 Validation result structures
type ValidationResults struct {
	Checks   []ValidationCheck `json:"checks"`
	Warnings []string          `json:"warnings"`
	Errors   []string          `json:"errors"`
	Passed   bool              `json:"passed"`
}

type ValidationCheck struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"` // PASS, WARN, FAIL
	Message     string `json:"message,omitempty"`
}

// 🔧 INDIVIDUAL VALIDATION FUNCTIONS

// 1. Basic field validation
func (m *PaymentValidationMiddleware) validateBasicFields(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "basic_fields",
		Description: "Validate basic required fields and data types",
	}

	// Required fields
	requiredFields := []string{"contact_id", "date", "amount"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Missing required field: %s", field)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate amount
	amount, ok := data["amount"].(float64)
	if !ok {
		// Try to convert from number or string
		if amountInt, ok := data["amount"].(int); ok {
			amount = float64(amountInt)
		} else if amountStr, ok := data["amount"].(string); ok {
			if parsed, err := strconv.ParseFloat(amountStr, 64); err == nil {
				amount = parsed
			} else {
				check.Status = "FAIL"
				check.Message = "Amount must be a valid number"
				results.Checks = append(results.Checks, check)
				results.Errors = append(results.Errors, check.Message)
				return fmt.Errorf("invalid amount format")
			}
		} else {
			check.Status = "FAIL"
			check.Message = "Amount must be a valid number"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid amount format")
		}
	}

	if amount <= 0 {
		check.Status = "FAIL"
		check.Message = "Amount must be greater than 0"
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("invalid amount value")
	}

	if amount > m.maxPaymentAmount {
		check.Status = "FAIL"
		check.Message = fmt.Sprintf("Amount exceeds maximum limit of %.2f", m.maxPaymentAmount)
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("amount exceeds limit")
	}

	// Validate date
	if dateStr, ok := data["date"].(string); ok {
		if _, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err != nil {
			if _, err := time.Parse("2006-01-02", dateStr); err != nil {
				check.Status = "FAIL"
				check.Message = "Date must be in valid format (YYYY-MM-DD or RFC3339)"
				results.Checks = append(results.Checks, check)
				results.Errors = append(results.Errors, check.Message)
				return fmt.Errorf("invalid date format")
			}
		}
	}

	// Validate reference if required
	if m.requireReference {
		if ref, exists := data["reference"]; !exists || ref == "" {
			check.Status = "FAIL"
			check.Message = "Reference is required for all payments"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("missing reference")
		}
	}

	check.Status = "PASS"
	check.Message = "All basic fields are valid"
	results.Checks = append(results.Checks, check)
	return nil
}

// 2. Contact validation
func (m *PaymentValidationMiddleware) validateContact(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "contact_validation",
		Description: "Validate contact exists and is active",
	}

	contactID, ok := data["contact_id"].(float64)
	if !ok {
		if id, ok := data["contact_id"].(int); ok {
			contactID = float64(id)
		} else {
			check.Status = "FAIL"
			check.Message = "Contact ID must be a valid number"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid contact ID")
		}
	}

	contact, err := m.contactRepo.GetByID(uint(contactID))
	if err != nil {
		check.Status = "FAIL"
		check.Message = fmt.Sprintf("Contact not found: %v", err)
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("contact not found")
	}

	if !contact.IsActive {
		check.Status = "FAIL"
		check.Message = fmt.Sprintf("Contact '%s' is inactive", contact.Name)
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("inactive contact")
	}

	// Store contact in context for later use
	data["_contact"] = contact

	check.Status = "PASS"
	check.Message = fmt.Sprintf("Contact '%s' is valid and active", contact.Name)
	results.Checks = append(results.Checks, check)
	return nil
}

// 3. Cash/Bank validation
func (m *PaymentValidationMiddleware) validateCashBank(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "cashbank_validation",
		Description: "Validate cash/bank account if specified",
	}

	// Cash/bank ID is optional, will be auto-selected if not provided
	cashBankIDInterface, exists := data["cash_bank_id"]
	if !exists || cashBankIDInterface == nil || cashBankIDInterface == 0 {
		check.Status = "PASS"
		check.Message = "No cash/bank specified - will be auto-selected"
		results.Checks = append(results.Checks, check)
		results.Warnings = append(results.Warnings, "Cash/bank account will be auto-selected")
		return nil
	}

	cashBankID, ok := cashBankIDInterface.(float64)
	if !ok {
		if id, ok := cashBankIDInterface.(int); ok {
			cashBankID = float64(id)
		} else {
			check.Status = "FAIL"
			check.Message = "Cash/bank ID must be a valid number"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid cash/bank ID")
		}
	}

	cashBank, err := m.cashBankRepo.FindByID(uint(cashBankID))
	if err != nil {
		check.Status = "FAIL"
		check.Message = fmt.Sprintf("Cash/bank account not found: %v", err)
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("cash/bank not found")
	}

	if !cashBank.IsActive {
		check.Status = "FAIL"
		check.Message = fmt.Sprintf("Cash/bank account '%s' is inactive", cashBank.Name)
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("inactive cash/bank account")
	}

	// Store cash/bank in context
	data["_cashbank"] = cashBank

	check.Status = "PASS"
	check.Message = fmt.Sprintf("Cash/bank account '%s' is valid", cashBank.Name)
	results.Checks = append(results.Checks, check)
	return nil
}

// 4. Payment method consistency validation
func (m *PaymentValidationMiddleware) validatePaymentMethodConsistency(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "method_consistency",
		Description: "Validate payment method is consistent with contact type",
	}

	contact, exists := data["_contact"]
	if !exists {
		check.Status = "FAIL"
		check.Message = "Contact not found in context"
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("contact not found")
	}

	contactModel := contact.(models.Contact)
	method, methodExists := data["method"].(string)

	// If method not specified, auto-detect
	if !methodExists || method == "" {
		if contactModel.Type == "CUSTOMER" {
			data["method"] = "RECEIVABLE"
			check.Status = "PASS"
			check.Message = "Auto-detected method: RECEIVABLE for customer"
			results.Warnings = append(results.Warnings, "Payment method auto-detected as RECEIVABLE")
		} else if contactModel.Type == "VENDOR" {
			data["method"] = "PAYABLE"
			check.Status = "PASS"
			check.Message = "Auto-detected method: PAYABLE for vendor"
			results.Warnings = append(results.Warnings, "Payment method auto-detected as PAYABLE")
		} else {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Cannot auto-detect method for contact type: %s", contactModel.Type)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("unsupported contact type")
		}
	} else {
		// Validate explicit method
		if contactModel.Type == "CUSTOMER" && method != "RECEIVABLE" {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Customer payments must use RECEIVABLE method, not %s", method)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid method for customer")
		}

		if contactModel.Type == "VENDOR" && method != "PAYABLE" {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Vendor payments must use PAYABLE method, not %s", method)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid method for vendor")
		}

		check.Status = "PASS"
		check.Message = fmt.Sprintf("Method %s is consistent with contact type %s", method, contactModel.Type)
	}

	results.Checks = append(results.Checks, check)
	return nil
}

// 5. Target allocation validation
func (m *PaymentValidationMiddleware) validateTargetAllocation(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "target_allocation",
		Description: "Validate target invoice/bill allocation if specified",
	}

	targetInvoiceID, hasInvoice := data["target_invoice_id"]
	targetBillID, hasBill := data["target_bill_id"]

	// Cannot specify both
	if hasInvoice && hasBill && targetInvoiceID != nil && targetBillID != nil {
		check.Status = "FAIL"
		check.Message = "Cannot specify both target_invoice_id and target_bill_id"
		results.Checks = append(results.Checks, check)
		results.Errors = append(results.Errors, check.Message)
		return fmt.Errorf("conflicting target allocations")
	}

	contact := data["_contact"].(models.Contact)

	// Validate invoice target
	if hasInvoice && targetInvoiceID != nil {
		if contact.Type != "CUSTOMER" {
			check.Status = "FAIL"
			check.Message = "target_invoice_id can only be used for customer payments"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid invoice target")
		}

		invoiceID := uint(targetInvoiceID.(float64))
		sale, err := m.salesRepo.FindByID(invoiceID)
		if err != nil {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Target invoice not found: %v", err)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invoice not found")
		}

		if sale.CustomerID != contact.ID {
			check.Status = "FAIL"
			check.Message = "Target invoice does not belong to selected customer"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invoice ownership mismatch")
		}

		if sale.OutstandingAmount <= 0 {
			check.Status = "FAIL"
			check.Message = "Target invoice is already fully paid"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invoice already paid")
		}

		amount := data["amount"].(float64)
		if amount > sale.OutstandingAmount {
			results.Warnings = append(results.Warnings, 
				fmt.Sprintf("Payment amount (%.2f) exceeds invoice outstanding (%.2f)", amount, sale.OutstandingAmount))
		}

		data["_target_invoice"] = sale
	}

	// Validate bill target
	if hasBill && targetBillID != nil {
		if contact.Type != "VENDOR" {
			check.Status = "FAIL"
			check.Message = "target_bill_id can only be used for vendor payments"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("invalid bill target")
		}

		billID := uint(targetBillID.(float64))
		purchase, err := m.purchaseRepo.FindByID(billID)
		if err != nil {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Target bill not found: %v", err)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("bill not found")
		}

		if purchase.VendorID != contact.ID {
			check.Status = "FAIL"
			check.Message = "Target bill does not belong to selected vendor"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("bill ownership mismatch")
		}

		if purchase.OutstandingAmount <= 0 {
			check.Status = "FAIL"
			check.Message = "Target bill is already fully paid"
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("bill already paid")
		}

		amount := data["amount"].(float64)
		if amount > purchase.OutstandingAmount {
			results.Warnings = append(results.Warnings, 
				fmt.Sprintf("Payment amount (%.2f) exceeds bill outstanding (%.2f)", amount, purchase.OutstandingAmount))
		}

		data["_target_bill"] = purchase
	}

	check.Status = "PASS"
	check.Message = "Target allocation validation passed"
	results.Checks = append(results.Checks, check)
	return nil
}

// 6. Balance and limits validation
func (m *PaymentValidationMiddleware) validateBalanceAndLimits(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "balance_limits",
		Description: "Validate balance requirements and payment limits",
	}

	method, _ := data["method"].(string)
	amount := data["amount"].(float64)

	// For outgoing payments (PAYABLE), check if we have sufficient balance
	if method == "PAYABLE" {
		cashBank, exists := data["_cashbank"]
		if exists && cashBank != nil {
			cashBankModel := cashBank.(models.CashBank)
			
			if !m.allowNegativeBalance && cashBankModel.Balance < amount {
				check.Status = "FAIL"
				check.Message = fmt.Sprintf("Insufficient balance in %s (%.2f) for payment amount (%.2f)", 
					cashBankModel.Name, cashBankModel.Balance, amount)
				results.Checks = append(results.Checks, check)
				results.Errors = append(results.Errors, check.Message)
				return fmt.Errorf("insufficient balance")
			}

			if cashBankModel.Balance < amount {
				results.Warnings = append(results.Warnings, 
					fmt.Sprintf("Payment will result in negative balance for %s", cashBankModel.Name))
			}
		} else {
			// Will auto-select account - check if any account has sufficient balance
			var cashBanks []models.CashBank
			m.db.Where("is_active = ? AND balance >= ?", true, amount).Find(&cashBanks)
			
			if len(cashBanks) == 0 && !m.allowNegativeBalance {
				check.Status = "FAIL"
				check.Message = fmt.Sprintf("No cash/bank account has sufficient balance (%.2f) for this payment", amount)
				results.Checks = append(results.Checks, check)
				results.Errors = append(results.Errors, check.Message)
				return fmt.Errorf("insufficient funds")
			}
		}
	}

	check.Status = "PASS"
	check.Message = "Balance and limits validation passed"
	results.Checks = append(results.Checks, check)
	return nil
}

// 7. Business logic validation
func (m *PaymentValidationMiddleware) validateBusinessLogic(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "business_logic",
		Description: "Validate business logic and constraints",
	}

	// Validate payment date is not in future
	dateStr, _ := data["date"].(string)
	paymentDate, _ := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if paymentDate.IsZero() {
		paymentDate, _ = time.Parse("2006-01-02", dateStr)
	}

	if paymentDate.After(time.Now().AddDate(0, 0, 1)) { // Allow 1 day in future
		results.Warnings = append(results.Warnings, "Payment date is in the future")
	}

	// Validate payment is not too old (more than 1 year ago)
	if paymentDate.Before(time.Now().AddDate(-1, 0, 0)) {
		results.Warnings = append(results.Warnings, "Payment date is more than 1 year old")
	}

	check.Status = "PASS"
	check.Message = "Business logic validation passed"
	results.Checks = append(results.Checks, check)
	return nil
}

// 8. Duplicate payment prevention
func (m *PaymentValidationMiddleware) validateDuplicatePayment(data map[string]interface{}, results *ValidationResults) error {
	check := ValidationCheck{
		Name:        "duplicate_prevention",
		Description: "Check for potential duplicate payments",
	}

	// Check for potential duplicates based on contact, amount, and recent date
	contactID := uint(data["contact_id"].(float64))
	amount := data["amount"].(float64)
	dateStr, _ := data["date"].(string)
	
	paymentDate, _ := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if paymentDate.IsZero() {
		paymentDate, _ = time.Parse("2006-01-02", dateStr)
	}

	// Look for payments within 24 hours with same contact and amount
	startTime := paymentDate.Add(-24 * time.Hour)
	endTime := paymentDate.Add(24 * time.Hour)

	var duplicateCount int64
	m.db.Model(&models.Payment{}).
		Where("contact_id = ? AND amount = ? AND date BETWEEN ? AND ?", 
			contactID, amount, startTime, endTime).
		Count(&duplicateCount)

	if duplicateCount > 0 {
		results.Warnings = append(results.Warnings, 
			fmt.Sprintf("Found %d similar payment(s) within 24 hours - please verify this is not a duplicate", duplicateCount))
	}

	// Check for duplicate reference if provided
	if ref, exists := data["reference"]; exists && ref != "" {
		var refCount int64
		m.db.Model(&models.Payment{}).
			Where("reference = ? AND contact_id = ?", ref, contactID).
			Count(&refCount)
			
		if refCount > 0 {
			check.Status = "FAIL"
			check.Message = fmt.Sprintf("Payment with reference '%s' already exists for this contact", ref)
			results.Checks = append(results.Checks, check)
			results.Errors = append(results.Errors, check.Message)
			return fmt.Errorf("duplicate reference")
		}
	}

	check.Status = "PASS"
	check.Message = "No duplicate payments detected"
	results.Checks = append(results.Checks, check)
	return nil
}

// 📤 Response helpers
func (m *PaymentValidationMiddleware) respondWithError(c *gin.Context, code string, message string, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": err.Error(),
		},
		"validation": gin.H{
			"passed": false,
		},
	})
	c.Abort()
}

func (m *PaymentValidationMiddleware) respondWithValidationResults(c *gin.Context, results *ValidationResults) {
	results.Passed = len(results.Errors) == 0

	statusCode := http.StatusBadRequest
	if results.Passed && len(results.Warnings) > 0 {
		statusCode = http.StatusOK // Passed with warnings
	}

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    "VALIDATION_FAILED",
			"message": "Payment validation failed",
			"details": results.Errors,
		},
		"validation": results,
	})
	c.Abort()
}

// 🛠️ Utility functions
func parseFloatFromInterface(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func parseUintFromInterface(value interface{}) (uint, bool) {
	if f, ok := parseFloatFromInterface(value); ok && f >= 0 {
		return uint(f), true
	}
	return 0, false
}