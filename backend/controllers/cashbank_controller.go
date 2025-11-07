package controllers

import (
	"app-sistem-akuntansi/services"
	"log"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
)

type CashBankController struct {
	cashBankService *services.CashBankService
	accountService  services.AccountService
}

func NewCashBankController(cashBankService *services.CashBankService, accountService services.AccountService) *CashBankController {
	return &CashBankController{
		cashBankService: cashBankService,
		accountService:  accountService,
	}
}

// GetAccounts godoc
// @Summary Get cash and bank accounts
// @Description Get all cash and bank accounts
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} models.CashBank
// @Router /api/cashbank/accounts [get]
func (c *CashBankController) GetAccounts(ctx *gin.Context) {
	accounts, err := c.cashBankService.GetCashBankAccounts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve accounts",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, accounts)
}

// GetAccountByID godoc
// @Summary Get account by ID
// @Description Get single cash/bank account details
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Success 200 {object} models.CashBank
// @Router /api/cashbank/accounts/{id} [get]
func (c *CashBankController) GetAccountByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}
	
	account, err := c.cashBankService.GetCashBankByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}
	
	ctx.JSON(http.StatusOK, account)
}

// CreateAccount godoc
// @Summary Create cash/bank account
// @Description Create new cash or bank account
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param account body services.CashBankCreateRequest true "Account data"
// @Success 201 {object} models.CashBank
// @Router /api/cashbank/accounts [post]
func (c *CashBankController) CreateAccount(ctx *gin.Context) {
	var request services.CashBankCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	
	account, err := c.cashBankService.CreateCashBankAccount(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create account",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusCreated, account)
}

// UpdateAccount godoc
// @Summary Update cash/bank account
// @Description Update cash or bank account details
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Param account body services.CashBankUpdateRequest true "Account update data"
// @Success 200 {object} models.CashBank
// @Router /api/cashbank/accounts/{id} [put]
func (c *CashBankController) UpdateAccount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}
	
	var request services.CashBankUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	// Debug log
	log.Printf("[CASHBANK UPDATE] ID: %d, Request: Name=%s, BankName=%s, AccountNo=%s, AccountHolderName=%s, Branch=%s", 
		id, request.Name, request.BankName, request.AccountNo, request.AccountHolderName, request.Branch)
	
	account, err := c.cashBankService.UpdateCashBankAccount(uint(id), request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update account",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, account)
}


// ProcessTransfer godoc
// @Summary Process transfer between accounts
// @Description Transfer money between cash/bank accounts
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param transfer body services.TransferRequest true "Transfer data"
// @Success 201 {object} services.CashBankTransfer
// @Router /api/cashbank/transfer [post]
func (c *CashBankController) ProcessTransfer(ctx *gin.Context) {
	var request services.TransferRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	
	transfer, err := c.cashBankService.ProcessTransfer(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process transfer",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusCreated, transfer)
}

// ProcessDeposit godoc
// @Summary Process deposit
// @Description Add money to cash/bank account
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param deposit body services.DepositRequest true "Deposit data"
// @Success 201 {object} models.CashBankTransaction
// @Router /api/cashbank/deposit [post]
func (c *CashBankController) ProcessDeposit(ctx *gin.Context) {
	var request services.DepositRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	userID := ctx.GetUint("user_id")
	
	transaction, err := c.cashBankService.ProcessDeposit(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process deposit",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusCreated, transaction)
}

// ProcessWithdrawal godoc
// @Summary Process withdrawal
// @Description Withdraw money from cash/bank account
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param withdrawal body services.WithdrawalRequest true "Withdrawal data"
// @Success 201 {object} models.CashBankTransaction
// @Router /api/cashbank/withdrawal [post]
func (c *CashBankController) ProcessWithdrawal(ctx *gin.Context) {
	var request services.WithdrawalRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	userID := ctx.GetUint("user_id")
	
	transaction, err := c.cashBankService.ProcessWithdrawal(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process withdrawal",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusCreated, transaction)
}

// GetTransactions godoc
// @Summary Get account transactions
// @Description Get transactions for a specific account
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Account ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} services.TransactionResult
// @Router /api/cashbank/accounts/{id}/transactions [get]
func (c *CashBankController) GetTransactions(ctx *gin.Context) {
	accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	filter := services.TransactionFilter{
		Page:  page,
		Limit: limit,
	}
	
	// Parse dates if provided
	if startDate := ctx.Query("start_date"); startDate != "" {
		if parsedStartDate, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = parsedStartDate
		}
	}
	if endDate := ctx.Query("end_date"); endDate != "" {
		if parsedEndDate, err := time.Parse("2006-01-02", endDate); err == nil {
			// make end_date inclusive by setting to end of day
			filter.EndDate = parsedEndDate.Add(24*time.Hour - time.Nanosecond)
		}
	}
	
	result, err := c.cashBankService.GetTransactions(uint(accountID), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve transactions",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, result)
}

// GetBalanceSummary godoc
// @Summary Get balance summary
// @Description Get summary of all account balances
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} services.BalanceSummary
// @Router /api/cashbank/balance-summary [get]
func (c *CashBankController) GetBalanceSummary(ctx *gin.Context) {
	summary, err := c.cashBankService.GetBalanceSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve balance summary",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, summary)
}

// GetRevenueAccounts godoc
// @Summary Get revenue accounts
// @Description Get active revenue accounts for deposit source selection
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Account
// @Router /api/cashbank/revenue-accounts [get]
func (c *CashBankController) GetRevenueAccounts(ctx *gin.Context) {
	accounts, err := c.accountService.GetRevenueAccounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve revenue accounts",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, accounts)
}

// GetDepositSourceAccounts godoc
// @Summary Get deposit source accounts
// @Description Get revenue and equity accounts for deposit source selection
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /api/cashbank/deposit-source-accounts [get]
func (c *CashBankController) GetDepositSourceAccounts(ctx *gin.Context) {
	// Get revenue accounts
revenueAccounts, err := c.accountService.GetRevenueAccounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve revenue accounts",
			"details": err.Error(),
		})
		return
	}
	
	// Get equity accounts
equityAccounts, err := c.accountService.GetEquityAccounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve equity accounts",
			"details": err.Error(),
		})
		return
	}
	
	// Return both revenue and equity accounts in the expected format
	response := gin.H{
		"data": gin.H{
			"revenue": revenueAccounts,
			"equity":  equityAccounts,
		},
	}
	
	ctx.JSON(http.StatusOK, response)
}

// GetPaymentAccounts godoc
// @Summary Get payment accounts
// @Description Get active cash and bank accounts for payment processing
// @Tags CashBank
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/cashbank/payment-accounts [get]
func (c *CashBankController) GetPaymentAccounts(ctx *gin.Context) {
	accounts, err := c.cashBankService.GetPaymentAccounts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve payment accounts",
			"details": err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    accounts,
	})
}

