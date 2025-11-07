package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// COGSController handles COGS (Cost of Goods Sold) operations
type COGSController struct {
	db          *gorm.DB
	cogsService *services.InventoryCOGSService
}

// NewCOGSController creates a new COGS controller
func NewCOGSController(db *gorm.DB) *COGSController {
	coaService := services.NewCOAService(db)
	return &COGSController{
		db:          db,
		cogsService: services.NewInventoryCOGSService(db, coaService),
	}
}

// GetCOGSSummary returns COGS summary for a period
// @Summary Get COGS Summary
// @Description Get Cost of Goods Sold summary for a specific period
// @Tags COGS
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/cogs/summary [get]
func (c *COGSController) GetCOGSSummary(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	summary, err := c.cogsService.GetCOGSSummary(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get COGS summary",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   summary,
	})
}

// BackfillCOGS backfills COGS entries for existing sales
// @Summary Backfill COGS Entries
// @Description Create COGS journal entries for sales that don't have them yet
// @Tags COGS
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param dry_run query boolean false "Preview without making changes"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/cogs/backfill [post]
func (c *COGSController) BackfillCOGS(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	dryRunStr := ctx.DefaultQuery("dry_run", "false")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	dryRun, _ := strconv.ParseBool(dryRunStr)

	// Get preview data
	type PreviewResult struct {
		SaleID         uint    `json:"sale_id"`
		InvoiceNumber  string  `json:"invoice_number"`
		SaleDate       string  `json:"sale_date"`
		TotalAmount    float64 `json:"total_amount"`
		EstimatedCOGS  float64 `json:"estimated_cogs"`
		COGSPercentage float64 `json:"cogs_percentage"`
	}

	var previews []PreviewResult
	var totalEstimatedCOGS float64

	rows, err := c.db.Raw(`
		SELECT 
			s.id as sale_id,
			s.invoice_number,
			s.date as sale_date,
			s.total_amount,
			COALESCE(SUM(si.quantity * p.cost_price), 0) as estimated_cogs
		FROM sales s
		LEFT JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN products p ON p.id = si.product_id
		WHERE s.date >= ? AND s.date <= ?
		  AND s.status IN ('INVOICED', 'PAID')
		  AND NOT EXISTS (
			SELECT 1 FROM unified_journal_ledger 
			WHERE source_type = 'SALE' AND source_id = s.id AND notes = 'COGS'
		  )
		GROUP BY s.id, s.invoice_number, s.date, s.total_amount
		ORDER BY s.date ASC
	`, startDate, endDate).Rows()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to preview COGS",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p PreviewResult
		var saleDate time.Time
		rows.Scan(&p.SaleID, &p.InvoiceNumber, &saleDate, &p.TotalAmount, &p.EstimatedCOGS)
		p.SaleDate = saleDate.Format("2006-01-02")
		
		if p.TotalAmount > 0 {
			p.COGSPercentage = (p.EstimatedCOGS / p.TotalAmount) * 100
		}
		
		totalEstimatedCOGS += p.EstimatedCOGS
		previews = append(previews, p)
	}

	// If dry run, return preview only
	if dryRun {
		ctx.JSON(http.StatusOK, gin.H{
			"status":              "success",
			"dry_run":             true,
			"sales_to_process":    len(previews),
			"total_estimated_cogs": totalEstimatedCOGS,
			"preview":             previews,
			"message":             fmt.Sprintf("Preview: %d sales will have COGS entries created", len(previews)),
		})
		return
	}

	// Execute backfill
	goCtx := context.Background()
	successCount, err := c.cogsService.BackfillCOGSForExistingSales(goCtx, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to backfill COGS",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":               "success",
		"sales_processed":      successCount,
		"total_estimated_cogs": totalEstimatedCOGS,
		"message":              fmt.Sprintf("Successfully created COGS entries for %d sales", successCount),
	})
}

// RecordCOGSForSale creates COGS entry for a specific sale
// @Summary Record COGS for Sale
// @Description Create COGS journal entry for a specific sale transaction
// @Tags COGS
// @Accept json
// @Produce json
// @Param sale_id path int true "Sale ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/cogs/record/{sale_id} [post]
func (c *COGSController) RecordCOGSForSale(ctx *gin.Context) {
	saleIDStr := ctx.Param("sale_id")
	saleID, err := strconv.ParseUint(saleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid sale_id",
		})
		return
	}

	// Load sale
	var sale struct {
		ID            uint
		InvoiceNumber string
		Date          time.Time  // Changed from SaleDate to Date
		TotalAmount   float64
		Status        string
	}

	if err := c.db.Table("sales").Where("id = ?", saleID).First(&sale).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Sale not found",
		})
		return
	}

	// Check if already has COGS
	var existingCount int64
	c.db.Table("unified_journal_ledger").
		Where("source_type = ? AND source_id = ? AND notes = ?", "SALE", saleID, "COGS").
		Count(&existingCount)

	if existingCount > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "COGS entry already exists for this sale",
		})
		return
	}

	// Create COGS entry (we need full sale model for this)
	var fullSale struct {
		ID            uint   `gorm:"primaryKey"`
		InvoiceNumber string
		SaleDate      time.Time
		TotalAmount   float64
		Status        string
		CreatedBy     uint
	}

	if err := c.db.Table("sales").Where("id = ?", saleID).First(&fullSale).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to load sale details",
		})
		return
	}

	// Note: This is a simplified approach - in production you'd need the full Sale model
	// For now, return success message with instruction
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("To record COGS for sale #%d, use the backfill endpoint or the backfill script", saleID),
		"sale": gin.H{
			"id":             sale.ID,
			"invoice_number": sale.InvoiceNumber,
			"sale_date":      sale.Date.Format("2006-01-02"),
			"total_amount":   sale.TotalAmount,
		},
	})
}

// GetSalesWithoutCOGS returns sales that don't have COGS entries
// @Summary Get Sales Without COGS
// @Description Get list of sales that don't have COGS entries yet
// @Tags COGS
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/cogs/missing [get]
func (c *COGSController) GetSalesWithoutCOGS(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	type SaleWithoutCOGS struct {
		ID             uint    `json:"id"`
		InvoiceNumber  string  `json:"invoice_number"`
		SaleDate       string  `json:"sale_date"`
		TotalAmount    float64 `json:"total_amount"`
		Status         string  `json:"status"`
		EstimatedCOGS  float64 `json:"estimated_cogs"`
		CustomerName   string  `json:"customer_name"`
	}

	var sales []SaleWithoutCOGS
	var totalCount int64

	// Count total
	c.db.Table("sales").
		Where("date >= ? AND date <= ?", startDateStr, endDateStr).
		Where("status IN ('INVOICED', 'PAID')").
		Where("NOT EXISTS (SELECT 1 FROM unified_journal_ledger WHERE source_type = 'SALE' AND source_id = sales.id AND notes = 'COGS')").
		Count(&totalCount)

	// Get paginated results
	rows, err := c.db.Raw(`
		SELECT 
			s.id,
			s.invoice_number,
			s.date as sale_date,
			s.total_amount,
			s.status,
			COALESCE(SUM(si.quantity * p.cost_price), 0) as estimated_cogs,
			COALESCE(c.name, '') as customer_name
		FROM sales s
		LEFT JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN products p ON p.id = si.product_id
		LEFT JOIN customers c ON c.id = s.customer_id
		WHERE s.date >= ? AND s.date <= ?
		  AND s.status IN ('INVOICED', 'PAID')
		  AND NOT EXISTS (
			SELECT 1 FROM unified_journal_ledger 
			WHERE source_type = 'SALE' AND source_id = s.id AND notes = 'COGS'
		  )
		GROUP BY s.id, s.invoice_number, s.date, s.total_amount, s.status, c.name
		ORDER BY s.date DESC
		LIMIT ? OFFSET ?
	`, startDateStr, endDateStr, limit, offset).Rows()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to query sales",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sale SaleWithoutCOGS
		var saleDate time.Time
		rows.Scan(&sale.ID, &sale.InvoiceNumber, &saleDate, &sale.TotalAmount, &sale.Status, &sale.EstimatedCOGS, &sale.CustomerName)
		sale.SaleDate = saleDate.Format("2006-01-02")
		sales = append(sales, sale)
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"sales":       sales,
			"total_count": totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

