package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
)

type QuoteController struct {
	quoteService   *services.QuoteServiceFull
	invoiceService *services.InvoiceServiceFull
}

func NewQuoteController(quoteService *services.QuoteServiceFull, invoiceService *services.InvoiceServiceFull) *QuoteController {
	return &QuoteController{
		quoteService:   quoteService,
		invoiceService: invoiceService,
	}
}

// GetQuotes handles GET /quotes
func (c *QuoteController) GetQuotes(ctx *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Build filter from query parameters
	filter := models.QuoteFilter{
		Status:     ctx.Query("status"),
		CustomerID: ctx.Query("customer_id"),
		StartDate:  ctx.Query("start_date"),
		EndDate:    ctx.Query("end_date"),
		Search:     ctx.Query("search"),
		Page:       page,
		Limit:      limit,
	}

	result, err := c.quoteService.GetQuotes(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetQuote handles GET /quotes/:id
func (c *QuoteController) GetQuote(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	quote, err := c.quoteService.GetQuoteByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": quote})
}

// CreateQuote handles POST /quotes
func (c *QuoteController) CreateQuote(ctx *gin.Context) {
	var request models.QuoteCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	quote, err := c.quoteService.CreateQuote(request, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Quote created successfully",
		"data":    quote,
	})
}

// UpdateQuote handles PUT /quotes/:id
func (c *QuoteController) UpdateQuote(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	var request models.QuoteUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	quote, err := c.quoteService.UpdateQuote(uint(id), request, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Quote updated successfully",
		"data":    quote,
	})
}

// DeleteQuote handles DELETE /quotes/:id
func (c *QuoteController) DeleteQuote(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	err = c.quoteService.DeleteQuote(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Quote deleted successfully"})
}

// GenerateQuoteCode handles POST /quotes/generate-code
func (c *QuoteController) GenerateQuoteCode(ctx *gin.Context) {
	code, err := c.quoteService.GenerateQuoteCode()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": code})
}

// FormatCurrency handles POST /quotes/format-currency
func (c *QuoteController) FormatCurrency(ctx *gin.Context) {
	var request struct {
		Amount float64 `json:"amount" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	formatted := utils.FormatCurrency(request.Amount, "IDR")
	ctx.JSON(http.StatusOK, gin.H{"formatted": formatted})
}

// ConvertToInvoice handles POST /quotes/:id/convert-to-invoice
func (c *QuoteController) ConvertToInvoice(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		DueDate          *string `json:"due_date"`
		PaymentMethod    *string `json:"payment_method"`
		PaymentReference *string `json:"payment_reference"`
		BankAccountID    *uint   `json:"bank_account_id"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := c.quoteService.ConvertToInvoice(uint(id), userID.(uint), nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Quote converted to invoice successfully",
		"data":    invoice,
	})
}