package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PurchaseRequestController struct {
	service services.PurchaseRequestService
}

func NewPurchaseRequestController(service services.PurchaseRequestService) *PurchaseRequestController {
	return &PurchaseRequestController{service}
}

func (c *PurchaseRequestController) Create(ctx *gin.Context) {
	var pr models.PurchaseRequest
	if err := ctx.ShouldBindJSON(&pr); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set CreatedBy from context (assuming auth middleware sets "user_id")
	userID, exists := ctx.Get("user_id")
	if exists {
		pr.CreatedBy = userID.(uint)
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.service.CreatePR(&pr); err != nil {
		fmt.Printf("Error creating PR: %v\n", err) // Add logging
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, pr)
}

func (c *PurchaseRequestController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var pr models.PurchaseRequest
	if err := ctx.ShouldBindJSON(&pr); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdatePR(uint(id), &pr); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Purchase Request updated successfully"})
}

func (c *PurchaseRequestController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.service.DeletePR(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Purchase Request deleted successfully"})
}

func (c *PurchaseRequestController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	pr, err := c.service.GetPRByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Purchase Request not found"})
		return
	}

	ctx.JSON(http.StatusOK, pr)
}

func (c *PurchaseRequestController) GetAll(ctx *gin.Context) {
	filter := make(map[string]interface{})
	if projectID := ctx.Query("project_id"); projectID != "" {
		filter["project_id"] = projectID
	}
	if status := ctx.Query("status"); status != "" {
		filter["status"] = status
	}

	prs, err := c.service.GetAllPRs(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, prs)
}

func (c *PurchaseRequestController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if err := c.service.UpdateStatus(uint(id), req.Status, userID.(uint), req.Reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// GetMaterialImpact returns the estimated material impact of a purchase request
func (c *PurchaseRequestController) GetMaterialImpact(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	impacts, err := c.service.GetEstimatedMaterialImpact(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, impacts)
}
