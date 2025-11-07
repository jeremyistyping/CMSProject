package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type EmployeeApprovalHandler struct {
	approvalService *services.ApprovalService
	employeeDashboardService *services.EmployeeDashboardService
}

func NewEmployeeApprovalHandler(approvalService *services.ApprovalService, employeeDashboardService *services.EmployeeDashboardService) *EmployeeApprovalHandler {
	return &EmployeeApprovalHandler{
		approvalService: approvalService,
		employeeDashboardService: employeeDashboardService,
	}
}

// GetMyApprovalRequests returns approval requests relevant to the employee
// GET /api/employee/approvals/requests
func (h *EmployeeApprovalHandler) GetMyApprovalRequests(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	// Get filter parameters
	status := c.Query("status")
	module := c.Query("module")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	// Get approval requests for this user
	requests, total, err := h.approvalService.GetApprovalRequests(userID, userRole, status, module, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approval requests",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Approval requests retrieved successfully",
		"data": gin.H{
			"requests": requests,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_pages": (int(total) + limit - 1) / limit,
			},
		},
	})
}

// GetPendingApprovalsForMe returns approvals pending for this employee's role
// GET /api/employee/approvals/pending
func (h *EmployeeApprovalHandler) GetPendingApprovalsForMe(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	// Get pending approvals for this user's role
	pendingApprovals, err := h.approvalService.GetPendingApprovals(userID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending approvals",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Pending approvals retrieved successfully",
		"data": gin.H{
			"pending_approvals": pendingApprovals,
			"total":             len(pendingApprovals),
			"user_role":         userRole,
		},
	})
}

// ProcessApproval processes an approval action (approve/reject)
// POST /api/employee/approvals/:id/process
func (h *EmployeeApprovalHandler) ProcessApproval(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	_, err = utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval request ID"})
		return
	}
	
	var request models.ApprovalActionDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate action
	if request.Action != "APPROVE" && request.Action != "REJECT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be 'APPROVE' or 'REJECT'"})
		return
	}
	
	// If rejecting, comments are required
	if request.Action == "REJECT" && strings.TrimSpace(request.Comments) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comments are required when rejecting"})
		return
	}
	
	// Process the approval action
	if err := h.approvalService.ProcessApprovalAction(uint(requestID), userID, request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process approval",
			"details": err.Error(),
		})
		return
	}
	
	// Get updated approval request to return current status
	updatedRequest, err := h.approvalService.GetApprovalRequest(uint(requestID))
	if err != nil {
		// Still return success since the approval was processed
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Approval request %s successfully", strings.ToLower(request.Action)),
			"action":  request.Action,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Approval request %s successfully", strings.ToLower(request.Action)),
		"action":  request.Action,
		"data": gin.H{
			"approval_request": updatedRequest,
			"current_status":   updatedRequest.Status,
		},
	})
}

// GetApprovalHistory returns the history of an approval request
// GET /api/employee/approvals/:id/history
func (h *EmployeeApprovalHandler) GetApprovalHistory(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval request ID"})
		return
	}
	
	// First check if user has access to this approval request
	approvalRequest, err := h.approvalService.GetApprovalRequest(uint(requestID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval request not found"})
		return
	}
	
	// Check if user can access this request (they submitted it, can approve it, or are admin/director)
	userRoleNorm := strings.ToLower(strings.TrimSpace(userRole))
	canAccess := false
	
	// User submitted this request
	if approvalRequest.RequesterID == userID {
		canAccess = true
	}
	
	// Admin or director can see all
	if userRoleNorm == "admin" || userRoleNorm == "director" {
		canAccess = true
	}
	
	// User can approve this request
	if !canAccess {
		for _, step := range approvalRequest.ApprovalSteps {
			if strings.ToLower(step.Step.ApproverRole) == userRoleNorm {
				canAccess = true
				break
			}
		}
	}
	
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this approval request"})
		return
	}
	
	// Get approval history
	history, err := h.approvalService.GetApprovalHistory(uint(requestID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approval history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Approval history retrieved successfully",
		"data": gin.H{
			"approval_request": approvalRequest,
			"history":         history,
		},
	})
}

// GetMySubmittedRequests returns approval requests submitted by this employee
// GET /api/employee/approvals/my-requests
func (h *EmployeeApprovalHandler) GetMySubmittedRequests(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	// Get filter parameters
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	// Get approval requests submitted by this user
	requests, _, err := h.approvalService.GetApprovalRequests(userID, "requester", status, "", page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch submitted requests",
			"details": err.Error(),
		})
		return
	}
	
	// Filter to only show requests where this user is the requester
	var myRequests []models.ApprovalRequest
	for _, request := range requests {
		if request.RequesterID == userID {
			myRequests = append(myRequests, request)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Submitted requests retrieved successfully",
		"data": gin.H{
			"requests": myRequests,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       len(myRequests),
				"total_pages": (len(myRequests) + limit - 1) / limit,
			},
		},
	})
}

// GetApprovalWorkflowsForEmployee returns workflows where employee has a role
// GET /api/employee/approvals/workflows
func (h *EmployeeApprovalHandler) GetApprovalWorkflowsForEmployee(c *gin.Context) {
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	// Get workflows relevant to this employee
	workflows, err := h.employeeDashboardService.GetEmployeeApprovalWorkflows(userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approval workflows",
			"details": err.Error(),
		})
		return
	}
	
	// Add summary information about employee's role in each workflow
	type WorkflowSummary struct {
		models.ApprovalWorkflow
		EmployeeRole  string   `json:"employee_role"`
		CanSubmit     bool     `json:"can_submit"`
		CanApprove    bool     `json:"can_approve"`
		ApproverSteps []string `json:"approver_steps"`
	}
	
	userRoleNorm := strings.ToLower(strings.TrimSpace(userRole))
	var workflowSummaries []WorkflowSummary
	
	for _, workflow := range workflows {
		summary := WorkflowSummary{
			ApprovalWorkflow: workflow,
			EmployeeRole:     userRole,
			CanSubmit:        false,
			CanApprove:       false,
			ApproverSteps:    []string{},
		}
		
		// Check if employee can submit requests for this workflow
		if userRoleNorm == "employee" && workflow.Module == "PURCHASE" {
			summary.CanSubmit = true
		}
		
		// Check which steps employee can approve
		for _, step := range workflow.Steps {
			if strings.ToLower(step.ApproverRole) == userRoleNorm {
				summary.CanApprove = true
				summary.ApproverSteps = append(summary.ApproverSteps, step.StepName)
			}
		}
		
		workflowSummaries = append(workflowSummaries, summary)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Employee approval workflows retrieved successfully",
		"data": gin.H{
			"workflows":  workflowSummaries,
			"total":      len(workflowSummaries),
			"user_role":  userRole,
		},
	})
}

// GetApprovalStatistics returns approval statistics for the employee
// GET /api/employee/approvals/statistics
func (h *EmployeeApprovalHandler) GetApprovalStatistics(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user role"})
		return
	}
	
	// Get statistics from the dashboard service
	stats, err := h.employeeDashboardService.GetEmployeeDashboardData(userID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch approval statistics",
			"details": err.Error(),
		})
		return
	}
	
	// Extract just the statistics part
	quickStats, exists := stats["quick_stats"]
	if !exists {
		quickStats = map[string]interface{}{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Approval statistics retrieved successfully",
		"data": gin.H{
			"statistics": quickStats,
			"user_role":  userRole,
		},
	})
}