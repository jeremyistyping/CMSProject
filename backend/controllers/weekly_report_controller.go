package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"fmt"
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

type WeeklyReportController struct {
	service services.WeeklyReportService
}

// NewWeeklyReportController creates a new weekly report controller
func NewWeeklyReportController(service services.WeeklyReportService) *WeeklyReportController {
	return &WeeklyReportController{service: service}
}

// GetWeeklyReports godoc
// @Summary Get all weekly reports for a project
// @Description Get list of all weekly reports for a specific project
// @Tags weekly-reports
// @Produce json
// @Param projectId path int true "Project ID"
// @Param year query int false "Filter by year"
// @Success 200 {array} models.WeeklyReportDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports [get]
func (wc *WeeklyReportController) GetWeeklyReports(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid project ID",
			"project_id": projectIDStr,
			"details":    err.Error(),
		})
		return
	}
	
	// Check for year filter
	yearStr := c.Query("year")
	var reports []models.WeeklyReportDTO
	
	if yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid year parameter",
				"details": err.Error(),
			})
			return
		}
		reports, err = wc.service.GetWeeklyReportsByYear(uint(projectID), year)
	} else {
		reports, err = wc.service.GetWeeklyReportsByProject(uint(projectID))
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   reports,
	})
}

// GetWeeklyReport godoc
// @Summary Get weekly report by ID
// @Description Get a single weekly report by ID
// @Tags weekly-reports
// @Produce json
// @Param projectId path int true "Project ID"
// @Param reportId path int true "Weekly Report ID"
// @Success 200 {object} models.WeeklyReportDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports/{reportId} [get]
func (wc *WeeklyReportController) GetWeeklyReport(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	reportID, err := strconv.ParseUint(c.Param("reportId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekly report ID"})
		return
	}
	
	report, err := wc.service.GetWeeklyReportByID(uint(projectID), uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// CreateWeeklyReport godoc
// @Summary Create a new weekly report
// @Description Create a new weekly report for a project
// @Tags weekly-reports
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param weeklyReport body models.WeeklyReportCreateRequest true "Weekly Report object"
// @Success 201 {object} models.WeeklyReportDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports [post]
func (wc *WeeklyReportController) CreateWeeklyReport(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid project ID",
			"project_id": projectIDStr,
			"details":    err.Error(),
		})
		return
	}
	
	var request models.WeeklyReportCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Override project ID from URL param
	request.ProjectID = uint(projectID)
	
	// Set created_by from user context if available
	if userID, exists := c.Get("userID"); exists {
		if userIDStr, ok := userID.(string); ok {
			request.CreatedBy = userIDStr
		}
	}
	
	report, err := wc.service.CreateWeeklyReport(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to create weekly report",
			"details":    err.Error(),
			"project_id": projectID,
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Weekly report created successfully",
		"data":    report,
	})
}

// UpdateWeeklyReport godoc
// @Summary Update a weekly report
// @Description Update an existing weekly report
// @Tags weekly-reports
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param reportId path int true "Weekly Report ID"
// @Param weeklyReport body models.WeeklyReportUpdateRequest true "Weekly Report object"
// @Success 200 {object} models.WeeklyReportDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports/{reportId} [put]
func (wc *WeeklyReportController) UpdateWeeklyReport(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	reportID, err := strconv.ParseUint(c.Param("reportId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekly report ID"})
		return
	}
	
	var request models.WeeklyReportUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Verify the report belongs to the project
	existing, err := wc.service.GetWeeklyReportByID(uint(projectID), uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	if existing.ProjectID != uint(projectID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weekly report does not belong to this project"})
		return
	}
	
	report, err := wc.service.UpdateWeeklyReport(uint(reportID), &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Weekly report updated successfully",
		"data":    report,
	})
}

// DeleteWeeklyReport godoc
// @Summary Delete a weekly report
// @Description Delete a weekly report by ID
// @Tags weekly-reports
// @Produce json
// @Param projectId path int true "Project ID"
// @Param reportId path int true "Weekly Report ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports/{reportId} [delete]
func (wc *WeeklyReportController) DeleteWeeklyReport(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	reportID, err := strconv.ParseUint(c.Param("reportId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekly report ID"})
		return
	}
	
	if err := wc.service.DeleteWeeklyReport(uint(projectID), uint(reportID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Weekly report deleted successfully",
	})
}

// GeneratePDF godoc
// @Summary Generate PDF for weekly report
// @Description Generate and download PDF for a weekly report
// @Tags weekly-reports
// @Produce application/pdf
// @Param projectId path int true "Project ID"
// @Param reportId path int true "Weekly Report ID"
// @Success 200 {file} pdf
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/weekly-reports/{reportId}/pdf [get]
func (wc *WeeklyReportController) GeneratePDF(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	reportID, err := strconv.ParseUint(c.Param("reportId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekly report ID"})
		return
	}
	
	report, err := wc.service.GetWeeklyReportByID(uint(projectID), uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	// Generate PDF
	pdf := wc.generateWeeklyReportPDF(report)
	
	// Set response headers
	filename := fmt.Sprintf("weekly_report_%s_week%d_%d.pdf", 
		sanitizeFilename(report.ProjectName), report.Week, report.Year)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	
	// Output PDF
	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

// generateWeeklyReportPDF creates a PDF document for the weekly report
func (wc *WeeklyReportController) generateWeeklyReportPDF(report *models.WeeklyReportDTO) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(0, 10, "PROJECT WEEKLY REPORT")
	pdf.Ln(12)
	
	// Project Details
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 255)
	pdf.Cell(40, 6, "Customer:")
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, report.ProjectName)
	pdf.Ln(6)
	
	pdf.SetTextColor(0, 0, 255)
	pdf.Cell(40, 6, "Week:")
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, fmt.Sprintf("%d-W%d", report.Year, report.Week))
	pdf.Ln(6)
	
	pdf.SetTextColor(0, 0, 255)
	pdf.Cell(40, 6, "Project Manager:")
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, report.ProjectManager)
	pdf.Ln(6)
	
	pdf.SetTextColor(0, 0, 255)
	pdf.Cell(40, 6, "Generated:")
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, report.GeneratedDate.Format("01/02/2006"))
	pdf.Ln(12)
	
	// Summary Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "SUMMARY")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(60, 6, fmt.Sprintf("Total Work Days: %d", report.TotalWorkDays))
	pdf.Ln(6)
	pdf.Cell(60, 6, fmt.Sprintf("Weather Delays: %d", report.WeatherDelays))
	pdf.Ln(6)
	pdf.Cell(60, 6, fmt.Sprintf("Team Size: %d", report.TeamSize))
	pdf.Ln(12)
	
	// Accomplishments
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "ACCOMPLISHMENTS")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, report.Accomplishments, "", "", false)
	pdf.Ln(8)
	
	// Challenges
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "CHALLENGES")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	if report.Challenges != "" {
		pdf.MultiCell(0, 5, report.Challenges, "", "", false)
	} else {
		pdf.Cell(0, 5, "No challenges recorded")
	}
	pdf.Ln(8)
	
	// Next Week Priorities
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "NEXT WEEK PRIORITIES")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	if report.NextWeekPriorities != "" {
		pdf.MultiCell(0, 5, report.NextWeekPriorities, "", "", false)
	} else {
		pdf.Cell(0, 5, "No priorities listed")
	}
	
	return pdf
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(s string) string {
	// Replace spaces and special characters
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

