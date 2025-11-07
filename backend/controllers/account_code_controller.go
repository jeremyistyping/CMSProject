package controllers

import (
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
)

// AccountCodeController handles account code generation and validation
type AccountCodeController struct {
	codeGenerator *services.AccountCodeGenerator
}

// NewAccountCodeController creates a new account code controller
func NewAccountCodeController(codeGenerator *services.AccountCodeGenerator) *AccountCodeController {
	return &AccountCodeController{
		codeGenerator: codeGenerator,
	}
}

// GenerateCodeRequest represents the request for generating account codes
type GenerateCodeRequest struct {
	AccountType string `json:"account_type" binding:"required"`
	ParentCode  string `json:"parent_code"`
	AccountName string `json:"account_name"`
}

// GenerateCodeResponse represents the response for generated codes
type GenerateCodeResponse struct {
	RecommendedCode string                           `json:"recommended_code"`
	AlternativeCodes []string                        `json:"alternative_codes,omitempty"`
	ValidationResult *AccountCodeValidationResponse  `json:"validation,omitempty"`
}

// AccountCodeValidationResponse represents validation results
type AccountCodeValidationResponse struct {
	IsValid bool   `json:"is_valid"`
	Message string `json:"message"`
}

// GenerateAccountCode generates a new account code
// @Summary Generate account code
// @Description Generate a sequential account code based on account type and parent
// @Tags Account Code
// @Accept json
// @Produce json
// @Param request body GenerateCodeRequest true "Code generation parameters"
// @Success 200 {object} GenerateCodeResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/generate [post]
func (c *AccountCodeController) GenerateAccountCode(ctx *gin.Context) {
	var req GenerateCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(ctx, "Invalid request", utils.FormatValidationError(err))
		return
	}

	// Generate recommended code
	recommendedCode, err := c.codeGenerator.GetRecommendedCode(req.AccountType, req.ParentCode, req.AccountName)
	if err != nil {
		utils.SendInternalError(ctx, "Failed to generate account code", err.Error())
		return
	}

	// Generate alternative codes
	alternativeCodes := c.generateAlternatives(req.AccountType, req.ParentCode, recommendedCode)

	// Validate the recommended code
	validationErr := c.codeGenerator.ValidateAccountCode(recommendedCode, req.AccountType, req.ParentCode)
	validation := &AccountCodeValidationResponse{
		IsValid: validationErr == nil,
		Message: "Code is valid and available",
	}
	if validationErr != nil {
		validation.Message = validationErr.Error()
	}

	response := GenerateCodeResponse{
		RecommendedCode:  recommendedCode,
		AlternativeCodes: alternativeCodes,
		ValidationResult: validation,
	}

	utils.SendSuccess(ctx, "Account code generated successfully", response)
}

// ValidateAccountCode validates an account code
// @Summary Validate account code
// @Description Validate if an account code follows proper hierarchy and is available
// @Tags Account Code
// @Accept json
// @Produce json
// @Param code query string true "Account code to validate"
// @Param account_type query string true "Account type"
// @Param parent_code query string false "Parent account code"
// @Success 200 {object} AccountCodeValidationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/validate [get]
func (c *AccountCodeController) ValidateAccountCode(ctx *gin.Context) {
	code := ctx.Query("code")
	accountType := ctx.Query("account_type")
	parentCode := ctx.Query("parent_code")

	if code == "" || accountType == "" {
		utils.SendValidationError(ctx, "Missing required parameters", map[string]string{
			"code":         "Account code is required",
			"account_type": "Account type is required",
		})
		return
	}

	err := c.codeGenerator.ValidateAccountCode(code, accountType, parentCode)
	
	response := AccountCodeValidationResponse{
		IsValid: err == nil,
		Message: "Code is valid and available",
	}
	
	if err != nil {
		response.Message = err.Error()
	}

	utils.SendSuccess(ctx, "Code validation completed", response)
}

// GetNonStandardCodes identifies non-standard account codes
// @Summary Get non-standard codes
// @Description Get a list of account codes that don't follow standard patterns with suggested fixes
// @Tags Account Code
// @Accept json
// @Produce json
// @Success 200 {array} services.AccountCodeFix
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/non-standard [get]
func (c *AccountCodeController) GetNonStandardCodes(ctx *gin.Context) {
	fixes, err := c.codeGenerator.FixNonStandardCodes()
	if err != nil {
		utils.SendInternalError(ctx, "Failed to identify non-standard codes", err.Error())
		return
	}

	utils.SendSuccess(ctx, "Non-standard codes identified", fixes)
}

// GetNextSequentialCode gets the next available code in sequence
// @Summary Get next sequential code
// @Description Get the next available sequential code for an account type
// @Tags Account Code
// @Accept json
// @Produce json
// @Param account_type query string true "Account type (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)"
// @Param parent_code query string false "Parent account code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/next [get]
func (c *AccountCodeController) GetNextSequentialCode(ctx *gin.Context) {
	accountType := ctx.Query("account_type")
	parentCode := ctx.Query("parent_code")

	if accountType == "" {
		utils.SendValidationError(ctx, "Missing account_type parameter", map[string]string{
			"account_type": "Account type is required",
		})
		return
	}

	nextCode, err := c.codeGenerator.GenerateNextCode(accountType, parentCode)
	if err != nil {
		utils.SendInternalError(ctx, "Failed to generate next code", err.Error())
		return
	}

	response := map[string]string{
		"next_code":    nextCode,
		"account_type": accountType,
		"parent_code":  parentCode,
	}

	utils.SendSuccess(ctx, "Next sequential code generated", response)
}

// CheckCodeAvailability checks if a code is available
// @Summary Check code availability
// @Description Check if an account code is already in use
// @Tags Account Code
// @Accept json
// @Produce json
// @Param code query string true "Account code to check"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/availability [get]
func (c *AccountCodeController) CheckCodeAvailability(ctx *gin.Context) {
	code := ctx.Query("code")
	
	if code == "" {
		utils.SendValidationError(ctx, "Missing code parameter", map[string]string{
			"code": "Account code is required",
		})
		return
	}

	exists, err := c.codeGenerator.CodeExists(code)
	if err != nil {
		utils.SendInternalError(ctx, "Failed to check code availability", err.Error())
		return
	}

	response := map[string]interface{}{
		"code":      code,
		"exists":    exists,
		"available": !exists,
	}

	utils.SendSuccess(ctx, "Code availability checked", response)
}

// GetAccountingStructure returns the standard accounting code structure
// @Summary Get accounting structure
// @Description Get the standard account code ranges and rules
// @Tags Account Code
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/account-codes/structure [get]
func (c *AccountCodeController) GetAccountingStructure(ctx *gin.Context) {
	structure := map[string]interface{}{
		"code_ranges": map[string]map[string]interface{}{
			"ASSET": {
				"start":       1000,
				"end":         1999,
				"description": "Asset accounts (Cash, Bank, Receivables, Inventory, Fixed Assets)",
				"examples":    []string{"1101 - Kas", "1201 - Piutang Usaha", "1301 - Persediaan"},
			},
			"LIABILITY": {
				"start":       2000,
				"end":         2999,
				"description": "Liability accounts (Payables, Taxes, Loans)",
				"examples":    []string{"2101 - Utang Usaha", "2103 - PPN Keluaran", "2201 - Utang Bank"},
			},
			"EQUITY": {
				"start":       3000,
				"end":         3999,
				"description": "Equity accounts (Capital, Retained Earnings)",
				"examples":    []string{"3101 - Modal Pemilik", "3201 - Laba Ditahan"},
			},
			"REVENUE": {
				"start":       4000,
				"end":         4999,
				"description": "Revenue accounts (Sales, Service Income)",
				"examples":    []string{"4101 - Pendapatan Penjualan", "4201 - Pendapatan Jasa"},
			},
			"EXPENSE": {
				"start":       5000,
				"end":         5999,
				"description": "Expense accounts (Operating Expenses, Cost of Goods Sold)",
				"examples":    []string{"5101 - Harga Pokok Penjualan", "5201 - Beban Gaji"},
			},
		},
		"naming_conventions": map[string]string{
			"main_accounts":  "Use 4-digit codes (e.g., 1100, 1200)",
			"sub_accounts":   "Use sequential numbering (e.g., 1101, 1102, 1103)",
			"detail_accounts": "Continue sequential numbering or use sub-categories",
		},
		"hierarchy_rules": []string{
			"First digit determines account type (1=Assets, 2=Liabilities, 3=Equity, 4=Revenue, 5=Expenses)",
			"Second digit indicates category (e.g., 11xx=Current Assets, 12xx=Receivables)",
			"Child accounts should sequentially follow parent codes",
			"Avoid using dashes in main account codes",
			"Use consistent numbering within each category",
		},
	}

	utils.SendSuccess(ctx, "Accounting structure retrieved", structure)
}

// generateAlternatives generates alternative code suggestions
func (c *AccountCodeController) generateAlternatives(accountType, parentCode, recommendedCode string) []string {
	var alternatives []string
	
	// Try to generate a few alternative codes
	for i := 1; i <= 3; i++ {
		altCode, err := c.codeGenerator.GenerateNextCode(accountType, parentCode)
		if err != nil {
			break
		}
		
		if altCode != recommendedCode {
			// Check if this alternative is available
			if exists, _ := c.codeGenerator.CodeExists(altCode); !exists {
				alternatives = append(alternatives, altCode)
			}
		}
	}
	
	return alternatives
}

// SuggestCodeBasedOnName suggests codes based on account name patterns
// @Summary Suggest code by name
// @Description Suggest account codes based on account name and type
// @Tags Account Code
// @Accept json
// @Produce json
// @Param account_name query string true "Account name"
// @Param account_type query string true "Account type"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/account-codes/suggest [get]
func (c *AccountCodeController) SuggestCodeBasedOnName(ctx *gin.Context) {
	accountName := ctx.Query("account_name")
	accountType := ctx.Query("account_type")

	if accountName == "" || accountType == "" {
		utils.SendValidationError(ctx, "Missing required parameters", map[string]string{
			"account_name": "Account name is required",
			"account_type": "Account type is required",
		})
		return
	}

	recommendedCode, err := c.codeGenerator.GetRecommendedCode(accountType, "", accountName)
	if err != nil {
		utils.SendInternalError(ctx, "Failed to suggest code", err.Error())
		return
	}

	// Get some alternatives
	alternatives := c.generateAlternatives(accountType, "", recommendedCode)

	response := map[string]interface{}{
		"account_name":     accountName,
		"account_type":     accountType,
		"recommended_code": recommendedCode,
		"alternatives":     alternatives,
	}

	utils.SendSuccess(ctx, "Code suggestions generated", response)
}