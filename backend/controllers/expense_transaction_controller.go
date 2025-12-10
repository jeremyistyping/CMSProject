package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExpenseTransactionController struct {
	service services.ExpenseTransactionService
}

func NewExpenseTransactionController(service services.ExpenseTransactionService) *ExpenseTransactionController {
	return &ExpenseTransactionController{service: service}
}

// GetAll retrieves all expense transactions with filters
// GET /api/v1/expense-transactions
func (c *ExpenseTransactionController) GetAll(ctx *gin.Context) {
	filter := make(map[string]interface{})

	if projectID := ctx.Query("project_id"); projectID != "" {
		if id, err := strconv.ParseUint(projectID, 10, 32); err == nil {
			filter["project_id"] = uint(id)
		}
	}
	if coaAccountID := ctx.Query("coa_account_id"); coaAccountID != "" {
		if id, err := strconv.ParseUint(coaAccountID, 10, 32); err == nil {
			filter["coa_account_id"] = uint(id)
		}
	}
	if transactionType := ctx.Query("transaction_type"); transactionType != "" {
		filter["transaction_type"] = transactionType
	}
	if startDate := ctx.Query("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filter["start_date"] = date
		}
	}
	if endDate := ctx.Query("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filter["end_date"] = date
		}
	}
	if search := ctx.Query("search"); search != "" {
		filter["search"] = search
	}

	expenses, err := c.service.GetAll(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": expenses, "total": len(expenses)})
}

// GetByID retrieves an expense transaction by ID
// GET /api/v1/expense-transactions/:id
func (c *ExpenseTransactionController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense transaction ID"})
		return
	}

	expense, err := c.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Expense transaction not found"})
		return
	}

	ctx.JSON(http.StatusOK, expense)
}

// GetByProject retrieves expense transactions by project
// GET /api/v1/projects/:id/expenses
func (c *ExpenseTransactionController) GetByProject(ctx *gin.Context) {
	// Debug logging
	idParam := ctx.Param("id")
	log.Printf("[EXPENSE DEBUG] Received project ID param: '%s'", idParam)
	log.Printf("[EXPENSE DEBUG] All params: %+v", ctx.Params)
	
	projectID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.Printf("[EXPENSE ERROR] Failed to parse project ID '%s': %v", idParam, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID", "received": idParam})
		return
	}

	filter := make(map[string]interface{})
	if startDate := ctx.Query("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filter["start_date"] = date
		}
	}
	if endDate := ctx.Query("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filter["end_date"] = date
		}
	}
	if transactionType := ctx.Query("transaction_type"); transactionType != "" {
		filter["transaction_type"] = transactionType
	}

	expenses, err := c.service.GetByProject(uint(projectID), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": expenses, "total": len(expenses)})
}

// GetBudgetReport generates budget vs actual report
// GET /api/v1/projects/:id/reports/budget-vs-actual
func (c *ExpenseTransactionController) GetBudgetReport(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Parse date range
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	report, err := c.service.GetBudgetReport(uint(projectID), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// ExportBudgetReportPDF exports budget vs actual report as PDF
// GET /api/v1/projects/:id/reports/budget-vs-actual/pdf
func (c *ExpenseTransactionController) ExportBudgetReportPDF(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Parse date range
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Get report data
	report, err := c.service.GetBudgetReport(uint(projectID), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate PDF
	pdfService := services.NewBudgetReportPDFService()
	pdfBytes, err := pdfService.GenerateBudgetReportPDF(report)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("Budget_Report_%s_%s.pdf", report.ProjectName, time.Now().Format("20060102"))
	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	ctx.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// Create creates a new expense transaction
// POST /api/v1/projects/:id/expenses
func (c *ExpenseTransactionController) Create(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var dto models.CreateExpenseTransactionDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override project ID from URL
	dto.ProjectID = uint(projectID)

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	expense, err := c.service.Create(&dto, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, expense)
}

// BatchCreate creates multiple expense transactions
// POST /api/v1/projects/:id/expenses/batch
func (c *ExpenseTransactionController) BatchCreate(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var dtos []models.CreateExpenseTransactionDTO
	if err := ctx.ShouldBindJSON(&dtos); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override project ID from URL for all items
	for i := range dtos {
		dtos[i].ProjectID = uint(projectID)
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	expenses, err := c.service.BatchCreate(dtos, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": expenses, "total": len(expenses)})
}

// Update updates an existing expense transaction
// PUT /api/v1/expense-transactions/:id
func (c *ExpenseTransactionController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense transaction ID"})
		return
	}

	var dto models.UpdateExpenseTransactionDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := c.service.Update(uint(id), &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, expense)
}

// Delete deletes an expense transaction
// DELETE /api/v1/expense-transactions/:id
func (c *ExpenseTransactionController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense transaction ID"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Expense transaction deleted successfully"})
}
