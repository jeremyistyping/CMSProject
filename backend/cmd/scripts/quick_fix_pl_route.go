package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Quick fix untuk menambahkan route profit-loss yang hilang
// File ini bisa digunakan untuk testing atau sebagai referensi implementasi

func main() {
	fmt.Println("🔧 Quick Fix: Adding Missing Profit-Loss Route")
	fmt.Println("=" + strings.Repeat("=", 59))
	
	// Create a simple Gin router for testing
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	
	// Add the missing profit-loss route
	v1 := router.Group("/api/v1")
	{
		reports := v1.Group("/reports")
		{
			// Profit & Loss endpoint that frontend is looking for
			reports.GET("/profit-loss", handleProfitLoss)
		}
	}
	
	fmt.Println("✅ Added route: GET /api/v1/reports/profit-loss")
	fmt.Println("🚀 Starting test server on :8080")
	fmt.Println("📋 Test the route:")
	fmt.Println("   curl \"http://localhost:8080/api/v1/reports/profit-loss?start_date=2025-08-31&end_date=2025-09-17&format=json\"")
	
	// Start server
	log.Fatal(router.Run(":8080"))
}

// handleProfitLoss - Mock handler untuk profit-loss endpoint
func handleProfitLoss(c *gin.Context) {
	// Parse parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")
	
	// Validate parameters
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code": "MISSING_PARAMETERS",
				"message": "start_date and end_date are required",
			},
		})
		return
	}
	
	// Parse dates to validate format
	_, err1 := time.Parse("2006-01-02", startDate)
	_, err2 := time.Parse("2006-01-02", endDate)
	
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code": "INVALID_DATE_FORMAT",
				"message": "Dates must be in YYYY-MM-DD format",
			},
		})
		return
	}
	
	// Mock enhanced profit-loss data structure
	// This matches the EnhancedProfitLossData structure from the backend
	mockData := gin.H{
		"company": gin.H{
			"name": "Mock Company Ltd",
			"address": "123 Test Street",
		},
		"start_date": startDate,
		"end_date": endDate,
		"currency": "IDR",
		"generated_at": time.Now().Format("2006-01-02T15:04:05Z"),
		
		// Revenue section with enhanced structure
		"revenue": gin.H{
			"sales_revenue": gin.H{
				"name": "Sales Revenue",
				"items": []gin.H{
					{
						"account_id": 4101,
						"code": "4101",
						"name": "Sales Revenue - Products",
						"amount": 15000000.0,
						"percentage": 0.0,
						"category": "SALES_REVENUE",
					},
				},
				"subtotal": 15000000.0,
			},
			"service_revenue": gin.H{
				"name": "Service Revenue",
				"items": []gin.H{
					{
						"account_id": 4102,
						"code": "4102", 
						"name": "Service Revenue",
						"amount": 5000000.0,
						"percentage": 0.0,
						"category": "SERVICE_REVENUE",
					},
				},
				"subtotal": 5000000.0,
			},
			"other_revenue": gin.H{
				"name": "Other Revenue",
				"items": []gin.H{},
				"subtotal": 0.0,
			},
			"total_revenue": 20000000.0,
		},
		
		// Cost of Goods Sold section
		"cost_of_goods_sold": gin.H{
			"direct_materials": gin.H{
				"name": "Direct Materials",
				"items": []gin.H{},
				"subtotal": 0.0,
			},
			"other_cogs": gin.H{
				"name": "Other COGS",
				"items": []gin.H{
					{
						"account_id": 5101,
						"code": "5101",
						"name": "Cost of Goods Sold",
						"amount": 12000000.0,
						"percentage": 0.0,
						"category": "COST_OF_GOODS_SOLD",
					},
				},
				"subtotal": 12000000.0,
			},
			"total_cogs": 12000000.0,
		},
		
		// Calculated profitability metrics
		"gross_profit": 8000000.0,
		"gross_profit_margin": 40.0,
		
		// Operating expenses
		"operating_expenses": gin.H{
			"administrative": gin.H{
				"name": "Administrative Expenses",
				"items": []gin.H{
					{
						"account_id": 6101,
						"code": "6101",
						"name": "Administrative Expenses",
						"amount": 3000000.0,
						"percentage": 0.0,
						"category": "ADMINISTRATIVE_EXPENSE",
					},
				},
				"subtotal": 3000000.0,
			},
			"selling_marketing": gin.H{
				"name": "Selling & Marketing Expenses",
				"items": []gin.H{},
				"subtotal": 0.0,
			},
			"general": gin.H{
				"name": "General Expenses", 
				"items": []gin.H{
					{
						"account_id": 6201,
						"code": "6201",
						"name": "General Expenses",
						"amount": 2000000.0,
						"percentage": 0.0,
						"category": "GENERAL_EXPENSE",
					},
				},
				"subtotal": 2000000.0,
			},
			"total_opex": 5000000.0,
		},
		
		// Operating performance
		"operating_income": 3000000.0,
		"operating_margin": 15.0,
		"ebitda": 3000000.0,
		"ebitda_margin": 15.0,
		
		// Final results
		"income_before_tax": 3000000.0,
		"tax_expense": 450000.0,
		"tax_rate": 15.0,
		"net_income": 2550000.0,
		"net_income_margin": 12.75,
		
		// Additional metrics
		"earnings_per_share": 0.0,
		"shares_outstanding": 0.0,
	}
	
	// Return data in standard response format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": mockData,
		"metadata": gin.H{
			"report_type": "profit-loss",
			"generated_at": time.Now().Format("2006-01-02T15:04:05Z"),
			"generated_by": "mock-api",
			"parameters": gin.H{
				"start_date": startDate,
				"end_date": endDate,
				"format": format,
			},
			"generation_time": "0.001s",
			"record_count": 1,
			"version": "1.0.0",
			"format": format,
		},
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
	})
	
	fmt.Printf("📊 Served P&L data: %s to %s (format: %s)\n", startDate, endDate, format)
}