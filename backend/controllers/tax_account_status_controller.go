package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// TaxAccountStatusController provides enhanced status and validation endpoints
type TaxAccountStatusController struct {
	taxAccountService *services.TaxAccountService
	accountService    *services.COAService
}

// NewTaxAccountStatusController creates a new controller instance
func NewTaxAccountStatusController(
	taxAccountService *services.TaxAccountService,
	accountService *services.COAService,
) *TaxAccountStatusController {
	return &TaxAccountStatusController{
		taxAccountService: taxAccountService,
		accountService:    accountService,
	}
}

// AccountStatus represents the status of an account configuration
type AccountStatus struct {
	IsConfigured    bool                    `json:"is_configured"`
	Account         *models.AccountResponse `json:"account,omitempty"`
	Warnings        []string                `json:"warnings,omitempty"`
	Recommendations []string                `json:"recommendations,omitempty"`
}

// TaxAccountStatusResponse provides comprehensive status information
type TaxAccountStatusResponse struct {
	IsFullyConfigured bool                   `json:"is_fully_configured"`
	ConfigurationID   *uint                  `json:"configuration_id,omitempty"`
	LastUpdated       *string                `json:"last_updated,omitempty"`
	UpdatedBy         *models.UserResponse   `json:"updated_by,omitempty"`
	
	// Sales accounts status
	SalesReceivable   AccountStatus `json:"sales_receivable"`
	SalesCash         AccountStatus `json:"sales_cash"`
	SalesBank         AccountStatus `json:"sales_bank"`
	SalesRevenue      AccountStatus `json:"sales_revenue"`
	SalesOutputVAT    AccountStatus `json:"sales_output_vat"`
	
	// Purchase accounts status
	PurchasePayable   AccountStatus `json:"purchase_payable"`
	PurchaseCash      AccountStatus `json:"purchase_cash"`
	PurchaseBank      AccountStatus `json:"purchase_bank"`
	PurchaseInputVAT  AccountStatus `json:"purchase_input_vat"`
	PurchaseExpense   AccountStatus `json:"purchase_expense"`
	
	// Optional accounts
	WithholdingTax21  AccountStatus `json:"withholding_tax21"`
	WithholdingTax23  AccountStatus `json:"withholding_tax23"`
	WithholdingTax25  AccountStatus `json:"withholding_tax25"`
	TaxPayable        AccountStatus `json:"tax_payable"`
	Inventory         AccountStatus `json:"inventory"`
	COGS              AccountStatus `json:"cogs"`
	
	// Overall status
	MissingAccounts   []string `json:"missing_accounts"`
	SystemWarnings    []string `json:"system_warnings"`
	HealthScore       int      `json:"health_score"` // 0-100
}

// GetStatus provides comprehensive tax account configuration status
// @Summary Get tax account configuration status
// @Description Get detailed status information about tax account configuration including validation and recommendations
// @Tags Tax Account Status
// @Accept json
// @Produce json
// @Success 200 {object} TaxAccountStatusResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/settings/tax-accounts/status [get]
func (c *TaxAccountStatusController) GetStatus(ctx *gin.Context) {
	settings, err := c.taxAccountService.GetSettings()
	
	response := TaxAccountStatusResponse{
		IsFullyConfigured: false,
		MissingAccounts:   []string{},
		SystemWarnings:    []string{},
		HealthScore:       0,
	}
	
	if err != nil {
		// No settings configured yet
		response.MissingAccounts = []string{
			"sales_receivable", "sales_cash", "sales_bank", "sales_revenue", "sales_output_vat",
			"purchase_payable", "purchase_cash", "purchase_bank", "purchase_input_vat", "purchase_expense",
		}
		response.SystemWarnings = append(response.SystemWarnings, "No tax account settings configured. Please configure accounts to start using the system.")
		
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    response,
		})
		return
	}
	
	// Settings exist, populate status
	response.ConfigurationID = &settings.ID
	if settings.UpdatedAt.IsZero() == false {
		updatedStr := settings.UpdatedAt.Format("2006-01-02 15:04:05")
		response.LastUpdated = &updatedStr
	}
	
	// Convert user info if available
	if settings.UpdatedByUser.ID > 0 {
		userResponse := models.UserResponse{
			ID:       settings.UpdatedByUser.ID,
			Name:     settings.UpdatedByUser.GetDisplayName(),
			Username: settings.UpdatedByUser.Username,
		}
		response.UpdatedBy = &userResponse
	}
	
	// Check each account configuration
	requiredAccounts := 0
	configuredAccounts := 0
	
	// Sales accounts
	response.SalesReceivable = c.checkAccountStatus(&settings.SalesReceivableAccount, "Sales Receivable", "ASSET", true)
	response.SalesCash = c.checkAccountStatus(&settings.SalesCashAccount, "Sales Cash", "ASSET", true)
	response.SalesBank = c.checkAccountStatus(&settings.SalesBankAccount, "Sales Bank", "ASSET", true)
	response.SalesRevenue = c.checkAccountStatus(&settings.SalesRevenueAccount, "Sales Revenue", "REVENUE", true)
	response.SalesOutputVAT = c.checkAccountStatus(&settings.SalesOutputVATAccount, "Sales Output VAT", "LIABILITY", true)
	
	// Purchase accounts
	response.PurchasePayable = c.checkAccountStatus(&settings.PurchasePayableAccount, "Purchase Payable", "LIABILITY", true)
	response.PurchaseCash = c.checkAccountStatus(&settings.PurchaseCashAccount, "Purchase Cash", "ASSET", true)
	response.PurchaseBank = c.checkAccountStatus(&settings.PurchaseBankAccount, "Purchase Bank", "ASSET", true)
	response.PurchaseInputVAT = c.checkAccountStatus(&settings.PurchaseInputVATAccount, "Purchase Input VAT", "ASSET", true)
	response.PurchaseExpense = c.checkAccountStatus(&settings.PurchaseExpenseAccount, "Purchase Expense", "EXPENSE", true)
	
	// Count required accounts
	requiredAccounts = 10 // 5 sales + 5 purchase
	requiredStatuses := []AccountStatus{
		response.SalesReceivable, response.SalesCash, response.SalesBank, response.SalesRevenue, response.SalesOutputVAT,
		response.PurchasePayable, response.PurchaseCash, response.PurchaseBank, response.PurchaseInputVAT, response.PurchaseExpense,
	}
	
	for _, status := range requiredStatuses {
		if status.IsConfigured {
			configuredAccounts++
		} else {
			// Add to missing accounts list
			response.MissingAccounts = append(response.MissingAccounts, "Required account not configured")
		}
	}
	
	// Optional accounts
	response.WithholdingTax21 = c.checkOptionalAccount(settings.WithholdingTax21Account, "Withholding Tax 21", "LIABILITY")
	response.WithholdingTax23 = c.checkOptionalAccount(settings.WithholdingTax23Account, "Withholding Tax 23", "LIABILITY")
	response.WithholdingTax25 = c.checkOptionalAccount(settings.WithholdingTax25Account, "Withholding Tax 25", "LIABILITY")
	response.TaxPayable = c.checkOptionalAccount(settings.TaxPayableAccount, "Tax Payable", "LIABILITY")
	response.Inventory = c.checkOptionalAccount(settings.InventoryAccount, "Inventory", "ASSET")
	response.COGS = c.checkOptionalAccount(settings.COGSAccount, "Cost of Goods Sold", "EXPENSE")
	
	// Calculate health score
	if requiredAccounts > 0 {
		response.HealthScore = (configuredAccounts * 100) / requiredAccounts
	}
	
	response.IsFullyConfigured = (configuredAccounts == requiredAccounts)
	
	// Add system warnings based on analysis
	if response.HealthScore < 100 {
		response.SystemWarnings = append(response.SystemWarnings, "Some required accounts are not configured")
	}
	if response.HealthScore < 50 {
		response.SystemWarnings = append(response.SystemWarnings, "Critical: Most accounts are missing. System may not function properly.")
	}
	
	// Check for account type mismatches (add to warnings)
	c.validateAccountTypes(&response, settings)
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ValidateAccountSelection validates if a specific account is suitable for a role
// @Summary Validate account selection
// @Description Validate if a specific account is suitable for a particular tax account role
// @Tags Tax Account Status
// @Accept json
// @Produce json
// @Param account_id query int true "Account ID to validate"
// @Param role query string true "Role to validate for (sales_receivable, sales_cash, etc.)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/settings/tax-accounts/validate [get]
func (c *TaxAccountStatusController) ValidateAccountSelection(ctx *gin.Context) {
	accountIDParam := ctx.Query("account_id")
	role := ctx.Query("role")
	
	if accountIDParam == "" || role == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "account_id and role parameters are required",
		})
		return
	}
	
	accountID, err := strconv.ParseUint(accountIDParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account_id",
		})
		return
	}
	
	// Get account details
	accounts, err := c.accountService.GetAccountsWithFilter(map[string]interface{}{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch accounts",
			"details": err.Error(),
		})
		return
	}
	
	var selectedAccount *models.Account
	for _, account := range accounts {
		if account.ID == uint(accountID) {
			selectedAccount = &account
			break
		}
	}
	
	if selectedAccount == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Account not found",
		})
		return
	}
	
	validation := c.validateAccountForRole(selectedAccount, role)
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    validation,
	})
}

// Helper methods

func (c *TaxAccountStatusController) checkAccountStatus(account *models.Account, name, expectedType string, required bool) AccountStatus {
	status := AccountStatus{
		IsConfigured:    false,
		Warnings:        []string{},
		Recommendations: []string{},
	}
	
	if account != nil && account.ID > 0 {
		status.IsConfigured = true
		status.Account = &models.AccountResponse{
			ID:       account.ID,
			Code:     account.Code,
			Name:     account.Name,
			Type:     account.Type,
			Category: account.Category,
			IsActive: account.IsActive,
		}
		
		// Validate account type
		if account.Type != expectedType {
			status.Warnings = append(status.Warnings, 
				"Account type mismatch. Expected: "+expectedType+", Got: "+account.Type)
		}
		
		// Check if account is active
		if !account.IsActive {
			status.Warnings = append(status.Warnings, "Account is inactive")
		}
	} else if required {
		status.Recommendations = append(status.Recommendations, 
			"Please configure "+name+" account of type "+expectedType)
	}
	
	return status
}

func (c *TaxAccountStatusController) checkOptionalAccount(account *models.Account, name, expectedType string) AccountStatus {
	status := AccountStatus{
		IsConfigured: false,
		Warnings:     []string{},
		Recommendations: []string{},
	}
	
	if account != nil && account.ID > 0 {
		status.IsConfigured = true
		status.Account = &models.AccountResponse{
			ID:       account.ID,
			Code:     account.Code,
			Name:     account.Name,
			Type:     account.Type,
			Category: account.Category,
			IsActive: account.IsActive,
		}
		
		if account.Type != expectedType {
			status.Warnings = append(status.Warnings, 
				"Account type mismatch. Expected: "+expectedType+", Got: "+account.Type)
		}
		
		if !account.IsActive {
			status.Warnings = append(status.Warnings, "Account is inactive")
		}
	} else {
		status.Recommendations = append(status.Recommendations, 
			"Optional: Configure "+name+" account for enhanced tax management")
	}
	
	return status
}

func (c *TaxAccountStatusController) validateAccountTypes(response *TaxAccountStatusResponse, settings *models.TaxAccountSettings) {
	// Add specific validation logic for account type compatibility
	// This can be expanded based on business rules
}

func (c *TaxAccountStatusController) validateAccountForRole(account *models.Account, role string) map[string]interface{} {
	validation := map[string]interface{}{
		"is_valid":        true,
		"warnings":        []string{},
		"recommendations": []string{},
		"account":         map[string]interface{}{
			"id":       account.ID,
			"code":     account.Code,
			"name":     account.Name,
			"type":     account.Type,
			"category": account.Category,
			"is_active": account.IsActive,
		},
	}
	
	// Define expected types for each role
	expectedTypes := map[string]string{
		"sales_receivable":       "ASSET",
		"sales_cash":            "ASSET", 
		"sales_bank":            "ASSET",
		"sales_revenue":         "REVENUE",
		"sales_output_vat":      "LIABILITY",
		"purchase_payable":      "LIABILITY",
		"purchase_cash":         "ASSET",
		"purchase_bank":         "ASSET",
		"purchase_input_vat":    "ASSET",
		"purchase_expense":      "EXPENSE",
		"withholding_tax21":     "LIABILITY",
		"withholding_tax23":     "LIABILITY", 
		"withholding_tax25":     "LIABILITY",
		"tax_payable":           "LIABILITY",
		"inventory":             "ASSET",
		"cogs":                  "EXPENSE",
	}
	
	expectedType, exists := expectedTypes[role]
	if !exists {
		validation["is_valid"] = false
		validation["warnings"] = append(validation["warnings"].([]string), "Unknown role: "+role)
		return validation
	}
	
	// Check account type
	if account.Type != expectedType {
		validation["is_valid"] = false
		validation["warnings"] = append(validation["warnings"].([]string), 
			"Account type mismatch. Expected: "+expectedType+", Got: "+account.Type)
	}
	
	// Check if account is active
	if !account.IsActive {
		validation["warnings"] = append(validation["warnings"].([]string), 
			"Account is inactive. Consider using an active account.")
	}
	
	// Add role-specific validations
	switch role {
	case "sales_receivable":
		if account.Code[0] != '1' {
			validation["recommendations"] = append(validation["recommendations"].([]string), 
				"Receivable accounts typically start with '1' (Assets)")
		}
	case "sales_revenue":
		if account.Code[0] != '4' {
			validation["recommendations"] = append(validation["recommendations"].([]string), 
				"Revenue accounts typically start with '4'")
		}
	}
	
	return validation
}