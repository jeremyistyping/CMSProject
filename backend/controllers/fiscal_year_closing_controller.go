package controllers

import (
	"net/http"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// FiscalYearClosingController handles fiscal year-end closing
type FiscalYearClosingController struct {
	service *services.FiscalYearClosingService
}

// NewFiscalYearClosingController creates a new fiscal year closing controller
func NewFiscalYearClosingController(service *services.FiscalYearClosingService) *FiscalYearClosingController {
	return &FiscalYearClosingController{
		service: service,
	}
}

// PreviewClosing godoc
// @Summary Preview fiscal year-end closing
// @Description Get preview of fiscal year-end closing with detailed breakdown
// @Tags Fiscal Year Closing
// @Accept json
// @Produce json
// @Param fiscal_year_end query string true "Fiscal year end date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fiscal-closing/preview [get]
func (fycc *FiscalYearClosingController) PreviewClosing(c *gin.Context) {
	fiscalYearEndStr := c.Query("fiscal_year_end")
	if fiscalYearEndStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "fiscal_year_end parameter is required (format: YYYY-MM-DD)",
		})
		return
	}

	fiscalYearEnd, err := time.Parse("2006-01-02", fiscalYearEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid date format. Use YYYY-MM-DD",
			"details": err.Error(),
		})
		return
	}

	preview, err := fycc.service.PreviewFiscalYearClosing(c.Request.Context(), fiscalYearEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate closing preview",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    preview,
	})
}

// ExecuteClosing godoc
// @Summary Execute fiscal year-end closing
// @Description Execute fiscal year-end closing with automated closing entries
// @Tags Fiscal Year Closing
// @Accept json
// @Produce json
// @Param request body map[string]string true "Closing request with fiscal_year_end and notes"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fiscal-closing/execute [post]
func (fycc *FiscalYearClosingController) ExecuteClosing(c *gin.Context) {
	var req struct {
		FiscalYearEnd string `json:"fiscal_year_end" binding:"required"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	fiscalYearEnd, err := time.Parse("2006-01-02", req.FiscalYearEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid date format. Use YYYY-MM-DD",
			"details": err.Error(),
		})
		return
	}

	if req.Notes == "" {
		req.Notes = "Fiscal year-end closing"
	}

	// Pass Gin context instead of Request context to preserve user_id
	err = fycc.service.ExecuteFiscalYearClosing(c, fiscalYearEnd, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to execute fiscal year closing",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Fiscal year closed successfully. All revenue and expense accounts have been reset to zero and transferred to retained earnings.",
	})
}

// GetClosingHistory godoc
// @Summary Get fiscal year closing history
// @Description Get history of fiscal year closings
// @Tags Fiscal Year Closing
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fiscal-closing/history [get]
func (fycc *FiscalYearClosingController) GetClosingHistory(c *gin.Context) {
	history, err := fycc.service.GetFiscalYearClosingHistory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get closing history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}
