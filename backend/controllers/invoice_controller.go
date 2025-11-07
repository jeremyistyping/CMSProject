package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
)

type InvoiceController struct {
	invoiceService *services.InvoiceServiceFull
}

func NewInvoiceController(invoiceService *services.InvoiceServiceFull) *InvoiceController {
	return &InvoiceController{
		invoiceService: invoiceService,
	}
}

// GetInvoices handles GET /invoices
func (c *InvoiceController) GetInvoices(ctx *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Build filter from query parameters
	filter := models.InvoiceFilter{
		Status:     ctx.Query("status"),
		CustomerID: ctx.Query("customer_id"),
		StartDate:  ctx.Query("start_date"),
		EndDate:    ctx.Query("end_date"),
		Search:     ctx.Query("search"),
		Page:       page,
		Limit:      limit,
	}

	result, err := c.invoiceService.GetInvoices(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetInvoice handles GET /invoices/:id
func (c *InvoiceController) GetInvoice(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := c.invoiceService.GetInvoiceByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": invoice})
}

// CreateInvoice handles POST /invoices
func (c *InvoiceController) CreateInvoice(ctx *gin.Context) {
	var request models.InvoiceCreateRequest
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

	invoice, err := c.invoiceService.CreateInvoice(request, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Invoice created successfully",
		"data":    invoice,
	})
}

// UpdateInvoice handles PUT /invoices/:id
func (c *InvoiceController) UpdateInvoice(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var request models.InvoiceUpdateRequest
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

	invoice, err := c.invoiceService.UpdateInvoice(uint(id), request, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Invoice updated successfully",
		"data":    invoice,
	})
}

// DeleteInvoice handles DELETE /invoices/:id
func (c *InvoiceController) DeleteInvoice(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	err = c.invoiceService.DeleteInvoice(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

// GenerateInvoiceCode handles POST /invoices/generate-code
func (c *InvoiceController) GenerateInvoiceCode(ctx *gin.Context) {
	code, err := c.invoiceService.GenerateInvoiceCode()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": code})
}

// FormatCurrency handles POST /invoices/format-currency
func (c *InvoiceController) FormatCurrency(ctx *gin.Context) {
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