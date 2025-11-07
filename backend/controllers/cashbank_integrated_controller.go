package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// CashBankIntegratedController handles integrated endpoints between CashBank and SSOT Journal
type CashBankIntegratedController struct {
	integratedService *services.CashBankIntegratedService
}

// NewCashBankIntegratedController creates a new integrated controller
func NewCashBankIntegratedController(integratedService *services.CashBankIntegratedService) *CashBankIntegratedController {
	return &CashBankIntegratedController{
		integratedService: integratedService,
	}
}

// GetIntegratedAccountDetails gets integrated account details with SSOT data
// @Summary Get integrated account details
// @Description Get cash/bank account details integrated with SSOT journal data
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Success 200 {object} services.IntegratedAccountResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cashbank/integrated/accounts/{id} [get]
func (c *CashBankIntegratedController) GetIntegratedAccountDetails(ctx *gin.Context) {
	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid account ID",
		})
		return
	}

	result, err := c.integratedService.GetIntegratedAccountDetails(uint(accountID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get integrated account details",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetIntegratedSummary gets integrated summary of all cash/bank accounts
// @Summary Get integrated summary
// @Description Get summary of all cash/bank accounts integrated with SSOT journal data
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} services.IntegratedSummaryResponse
// @Failure 500 {object} map[string]string
// @Router /api/cashbank/integrated/summary [get]
func (c *CashBankIntegratedController) GetIntegratedSummary(ctx *gin.Context) {
	result, err := c.integratedService.GetIntegratedSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get integrated summary",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetAccountReconciliation gets balance reconciliation for specific account
// @Summary Get account reconciliation
// @Description Get balance reconciliation between CashBank and SSOT for specific account
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Success 200 {object} services.ReconciliationData
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cashbank/integrated/accounts/{id}/reconciliation [get]
func (c *CashBankIntegratedController) GetAccountReconciliation(ctx *gin.Context) {
	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid account ID",
		})
		return
	}

	// Get reconciliation data using dedicated service method
	result, err := c.integratedService.GetAccountReconciliation(uint(accountID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get account reconciliation",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetAccountJournalEntries gets journal entries for specific cash/bank account
// @Summary Get account journal entries
// @Description Get journal entries related to specific cash/bank account with pagination
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cashbank/integrated/accounts/{id}/journal-entries [get]
func (c *CashBankIntegratedController) GetAccountJournalEntries(ctx *gin.Context) {
	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid account ID",
		})
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse date filters (for future implementation)
	// Currently not used by the service but available for filtering enhancement
	// startDate and endDate would be passed to service method when date filtering is implemented
	_ = ctx.Query("start_date") // Acknowledged but not yet used
	_ = ctx.Query("end_date")   // Acknowledged but not yet used

	// Get account details first to get GL account ID
	accountDetails, err := c.integratedService.GetIntegratedAccountDetails(uint(accountID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get account details",
			"error":   err.Error(),
		})
		return
	}

	// Get journal entries using paginated service method
	journalEntriesResponse, err := c.integratedService.GetJournalEntriesPaginated(
		accountDetails.Account.AccountID,
		page,
		limit,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get journal entries",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   journalEntriesResponse,
	})
}

// GetAccountTransactionHistory gets integrated transaction history
// @Summary Get account transaction history
// @Description Get transaction history with journal entry references
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cashbank/integrated/accounts/{id}/transactions [get]
func (c *CashBankIntegratedController) GetAccountTransactionHistory(ctx *gin.Context) {
	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid account ID",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get paginated transaction history using service method
	transactionHistoryResponse, err := c.integratedService.GetTransactionHistoryPaginated(
		uint(accountID),
		page,
		limit,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get transaction history",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   transactionHistoryResponse,
	})
}

// Reconcile cash/bank balances vs COA/SSOT
// @Summary Reconcile Cash&Bank vs COA
// @Description Reconcile balances between CashBank and COA using selected strategy: to_ssot | to_transactions | to_coa
// @Tags CashBank Integration
// @Accept json
// @Produce json
// @Security Bearer
// @Param reconcile body services.ReconcileRequest true "Reconcile options"
// @Success 200 {object} services.ReconcileResult
// @Router /api/cashbank/integrated/reconcile [post]
func (c *CashBankIntegratedController) Reconcile(ctx *gin.Context) {
	var req services.ReconcileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status":"error","message":"invalid request","error":err.Error()})
		return
	}
	res, err := c.integratedService.ReconcileBalances(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status":"error","message":"reconcile failed","error":err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status":"success","data":res})
}

// Helper methods - removed obsolete methods as service layer now handles pagination and reconciliation
