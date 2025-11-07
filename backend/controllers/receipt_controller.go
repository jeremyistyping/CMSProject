package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReceiptController handles receipt-related operations
type ReceiptController struct {
	DB *gorm.DB
}

// NewReceiptController creates a new receipt controller
func NewReceiptController(db *gorm.DB) *ReceiptController {
	return &ReceiptController{
		DB: db,
	}
}

// Receipt represents a receipt record
type Receipt struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ReceiptNumber string    `json:"receipt_number" gorm:"uniqueIndex"`
	PaymentID     uint      `json:"payment_id"`
	Date          time.Time `json:"date"`
	Amount        float64   `json:"amount"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetReceipts handles GET /receipts
func (rc *ReceiptController) GetReceipts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var receipts []Receipt
	var total int64

	// Count total records
	if err := rc.DB.Model(&Receipt{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to count receipts",
			"error":   err.Error(),
		})
		return
	}

	// Get paginated receipts
	if err := rc.DB.Offset(offset).Limit(limit).Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch receipts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   receipts,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetReceipt handles GET /receipts/:id
func (rc *ReceiptController) GetReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid receipt ID",
		})
		return
	}

	var receipt Receipt
	if err := rc.DB.First(&receipt, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Receipt not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch receipt",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   receipt,
	})
}

// CreateReceipt handles POST /receipts
func (rc *ReceiptController) CreateReceipt(c *gin.Context) {
	var receipt Receipt
	if err := c.ShouldBindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
		return
	}

	// Generate receipt number if not provided
	if receipt.ReceiptNumber == "" {
		receipt.ReceiptNumber = generateReceiptNumber()
	}

	if err := rc.DB.Create(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create receipt",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Receipt created successfully",
		"data":    receipt,
	})
}

// UpdateReceipt handles PUT /receipts/:id
func (rc *ReceiptController) UpdateReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid receipt ID",
		})
		return
	}

	var receipt Receipt
	if err := rc.DB.First(&receipt, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Receipt not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch receipt",
			"error":   err.Error(),
		})
		return
	}

	var updateData Receipt
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
		return
	}

	if err := rc.DB.Model(&receipt).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update receipt",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Receipt updated successfully",
		"data":    receipt,
	})
}

// DeleteReceipt handles DELETE /receipts/:id
func (rc *ReceiptController) DeleteReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid receipt ID",
		})
		return
	}

	if err := rc.DB.Delete(&Receipt{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete receipt",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Receipt deleted successfully",
	})
}

// PrintReceipt handles POST /receipts/:id/print
func (rc *ReceiptController) PrintReceipt(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Receipt print job queued",
	})
}

// GenerateReceiptPDF handles GET /receipts/:id/pdf
func (rc *ReceiptController) GenerateReceiptPDF(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "PDF generation not yet implemented",
	})
}

// ExportReceipts handles GET /receipts/export
func (rc *ReceiptController) ExportReceipts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Export functionality not yet implemented",
	})
}

// BulkPrintReceipts handles POST /receipts/bulk-print
func (rc *ReceiptController) BulkPrintReceipts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Bulk print functionality not yet implemented",
	})
}

// SearchReceipts handles GET /receipts/search
func (rc *ReceiptController) SearchReceipts(c *gin.Context) {
	query := c.Query("q")
	
	var receipts []Receipt
	if err := rc.DB.Where("receipt_number LIKE ? OR notes LIKE ?", "%"+query+"%", "%"+query+"%").Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Search failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   receipts,
	})
}

// GetReceiptsByPayment handles GET /receipts/by-payment/:payment_id
func (rc *ReceiptController) GetReceiptsByPayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("payment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid payment ID",
		})
		return
	}

	var receipts []Receipt
	if err := rc.DB.Where("payment_id = ?", paymentID).Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch receipts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   receipts,
	})
}

// GetReceiptsByDateRange handles GET /receipts/by-date-range
func (rc *ReceiptController) GetReceiptsByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var receipts []Receipt
	query := rc.DB.Model(&Receipt{})

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch receipts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   receipts,
	})
}

// generateReceiptNumber generates a unique receipt number
func generateReceiptNumber() string {
	return "RCP-" + time.Now().Format("20060102-150405")
}