package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	fmt.Println("🔧 Starting Approval Workflow Bug Fix...")
	
	// 1. Find purchases with inconsistent approval states
	err := fixInconsistentApprovalStates(db)
	if err != nil {
		log.Printf("Error fixing inconsistent approval states: %v", err)
	}

	// 2. Fix approval actions that are not properly activated
	err = fixInactiveApprovalSteps(db)
	if err != nil {
		log.Printf("Error fixing inactive approval steps: %v", err)
	}

	// 3. Ensure proper notification setup for directors
	err = ensureDirectorNotifications(db)
	if err != nil {
		log.Printf("Error ensuring director notifications: %v", err)
	}

	// 4. Fix director step escalations
	err = fixDirectorStepEscalations(db)
	if err != nil {
		log.Printf("Error fixing director step escalations: %v", err)
	}

	fmt.Println("✅ Approval Workflow Bug Fix Completed!")
}

// fixInconsistentApprovalStates fixes purchases where approval status is inconsistent
func fixInconsistentApprovalStates(db *gorm.DB) error {
	fmt.Println("1. Fixing inconsistent approval states...")

	// Find purchases that are PENDING but have all steps approved
	var purchases []models.Purchase
	err := db.Preload("ApprovalRequest.ApprovalSteps.Step").
		Where("approval_status = ? AND status = ?", models.PurchaseApprovalPending, models.PurchaseStatusPending).
		Find(&purchases).Error
	if err != nil {
		return fmt.Errorf("failed to query purchases: %v", err)
	}

	for _, purchase := range purchases {
		if purchase.ApprovalRequest == nil {
			continue
		}

		// Check if all approval steps are completed
		allApproved := true
		hasActiveStep := false
		
		for _, step := range purchase.ApprovalRequest.ApprovalSteps {
			if step.Status == models.ApprovalStatusPending {
				if step.IsActive {
					hasActiveStep = true
				}
				allApproved = false
			}
		}

		if allApproved {
			// All steps approved, update purchase status
			now := time.Now()
			err = db.Model(&purchase).Updates(map[string]interface{}{
				"status":          models.PurchaseStatusApproved,
				"approval_status": models.PurchaseApprovalApproved,
				"approved_at":     now,
				"updated_at":      now,
			}).Error
			if err != nil {
				log.Printf("Failed to update purchase %s: %v", purchase.Code, err)
				continue
			}

			// Update approval request
			err = db.Model(&purchase.ApprovalRequest).Updates(map[string]interface{}{
				"status":       models.ApprovalStatusApproved,
				"completed_at": now,
				"updated_at":   now,
			}).Error
			if err != nil {
				log.Printf("Failed to update approval request for purchase %s: %v", purchase.Code, err)
			}

			fmt.Printf("✓ Fixed purchase %s - marked as approved\n", purchase.Code)
		} else if !hasActiveStep {
			// No active step but still pending - activate next pending step
			err = activateNextPendingStep(db, &purchase)
			if err != nil {
				log.Printf("Failed to activate next step for purchase %s: %v", purchase.Code, err)
			} else {
				fmt.Printf("✓ Activated next step for purchase %s\n", purchase.Code)
			}
		}
	}

	return nil
}

// fixInactiveApprovalSteps fixes approval steps that should be active but aren't
func fixInactiveApprovalSteps(db *gorm.DB) error {
	fmt.Println("2. Fixing inactive approval steps...")

	// Find approval requests that are PENDING but have no active steps
	var requests []models.ApprovalRequest
	err := db.Preload("ApprovalSteps.Step").
		Where("status = ?", models.ApprovalStatusPending).
		Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to query approval requests: %v", err)
	}

	for _, request := range requests {
		hasActiveStep := false
		for _, step := range request.ApprovalSteps {
			if step.IsActive && step.Status == models.ApprovalStatusPending {
				hasActiveStep = true
				break
			}
		}

		if !hasActiveStep {
			// Find the next step that should be active
			err = activateNextStepForRequest(db, &request)
			if err != nil {
				log.Printf("Failed to activate step for request %s: %v", request.RequestCode, err)
			} else {
				fmt.Printf("✓ Activated step for request %s\n", request.RequestCode)
			}
		}
	}

	return nil
}

// ensureDirectorNotifications ensures directors get notifications for escalated requests
func ensureDirectorNotifications(db *gorm.DB) error {
	fmt.Println("3. Ensuring director notifications...")

	// Find escalated requests that might not have director notifications
	var histories []models.ApprovalHistory
	err := db.Preload("Request").
		Where("action = ? AND created_at > ?", "ESCALATED_TO_DIRECTOR", time.Now().Add(-7*24*time.Hour)).
		Find(&histories).Error
	if err != nil {
		return fmt.Errorf("failed to query escalation history: %v", err)
	}

	for _, history := range histories {
		// Create notification for directors
		err = createDirectorNotifications(db, &history.Request, history.Comments)
		if err != nil {
			log.Printf("Failed to create director notification for request %s: %v", history.Request.RequestCode, err)
		} else {
			fmt.Printf("✓ Created director notification for request %s\n", history.Request.RequestCode)
		}
	}

	return nil
}

// fixDirectorStepEscalations fixes director step escalations that might be broken
func fixDirectorStepEscalations(db *gorm.DB) error {
	fmt.Println("4. Fixing director step escalations...")

	// Find requests with director steps that are not properly activated
	var requests []models.ApprovalRequest
	err := db.Preload("ApprovalSteps.Step").
		Where("status = ? AND priority = ?", models.ApprovalStatusPending, models.ApprovalPriorityHigh).
		Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to query high priority requests: %v", err)
	}

	for _, request := range requests {
		// Check if there's a director step that should be active
		var directorStep *models.ApprovalAction
		for i := range request.ApprovalSteps {
			step := &request.ApprovalSteps[i]
			if strings.ToLower(step.Step.ApproverRole) == "director" {
				directorStep = step
				break
			}
		}

		if directorStep != nil && !directorStep.IsActive && directorStep.Status == models.ApprovalStatusPending {
			// Deactivate all other steps
			err = db.Model(&models.ApprovalAction{}).
				Where("request_id = ? AND id != ?", request.ID, directorStep.ID).
				Update("is_active", false).Error
			if err != nil {
				log.Printf("Failed to deactivate other steps for request %s: %v", request.RequestCode, err)
				continue
			}

			// Activate director step
			err = db.Model(directorStep).Update("is_active", true).Error
			if err != nil {
				log.Printf("Failed to activate director step for request %s: %v", request.RequestCode, err)
				continue
			}

			// Create notification for directors
			err = createDirectorNotifications(db, &request, "Escalated request requiring director approval")
			if err != nil {
				log.Printf("Failed to create director notification for request %s: %v", request.RequestCode, err)
			}

			fmt.Printf("✓ Fixed director step for request %s\n", request.RequestCode)
		}
	}

	return nil
}

// activateNextPendingStep activates the next pending step in workflow
func activateNextPendingStep(db *gorm.DB, purchase *models.Purchase) error {
	if purchase.ApprovalRequest == nil {
		return fmt.Errorf("no approval request found")
	}

	// Find the next step that should be active (lowest step_order with PENDING status)
	var nextStep *models.ApprovalAction
	minOrder := 999999

	for i := range purchase.ApprovalRequest.ApprovalSteps {
		step := &purchase.ApprovalRequest.ApprovalSteps[i]
		if step.Status == models.ApprovalStatusPending && step.Step.StepOrder < minOrder {
			nextStep = step
			minOrder = step.Step.StepOrder
		}
	}

	if nextStep != nil {
		// Deactivate all steps first
		err := db.Model(&models.ApprovalAction{}).
			Where("request_id = ?", purchase.ApprovalRequest.ID).
			Update("is_active", false).Error
		if err != nil {
			return err
		}

		// Activate the next step
		err = db.Model(nextStep).Update("is_active", true).Error
		if err != nil {
			return err
		}

		// Create notification for appropriate approvers
		err = createApprovalNotifications(db, purchase.ApprovalRequest, nextStep)
		if err != nil {
			log.Printf("Failed to create notifications: %v", err)
		}
	}

	return nil
}

// activateNextStepForRequest activates next step for approval request
func activateNextStepForRequest(db *gorm.DB, request *models.ApprovalRequest) error {
	// Find the next step that should be active
	var nextStep *models.ApprovalAction
	minOrder := 999999

	for i := range request.ApprovalSteps {
		step := &request.ApprovalSteps[i]
		if step.Status == models.ApprovalStatusPending && step.Step.StepOrder < minOrder {
			nextStep = step
			minOrder = step.Step.StepOrder
		}
	}

	if nextStep != nil {
		// Deactivate all steps first
		err := db.Model(&models.ApprovalAction{}).
			Where("request_id = ?", request.ID).
			Update("is_active", false).Error
		if err != nil {
			return err
		}

		// Activate the next step
		err = db.Model(nextStep).Update("is_active", true).Error
		if err != nil {
			return err
		}

		// Create notification
		err = createApprovalNotifications(db, request, nextStep)
		if err != nil {
			log.Printf("Failed to create notifications: %v", err)
		}
	}

	return nil
}

// createDirectorNotifications creates notifications for directors
func createDirectorNotifications(db *gorm.DB, request *models.ApprovalRequest, reason string) error {
	// Get all director users
	var directors []models.User
	err := db.Where("LOWER(role) = ? AND is_active = ?", "director", true).Find(&directors).Error
	if err != nil {
		return err
	}

	for _, director := range directors {
		notification := models.Notification{
			UserID:   director.ID,
			Type:     models.NotificationTypeApprovalPending,
			Title:    fmt.Sprintf("URGENT: Director Approval Required - %s", request.RequestTitle),
			Message:  fmt.Sprintf("Finance has escalated this request: %s. Reason: %s (Amount: %.2f)", request.RequestTitle, reason, request.Amount),
			Priority: models.ApprovalPriorityUrgent,
			Data:     fmt.Sprintf(`{"request_id":%d,"entity_type":"%s","entity_id":%d,"amount":%.2f}`, request.ID, request.EntityType, request.EntityID, request.Amount),
		}

		err = db.Create(&notification).Error
		if err != nil {
			log.Printf("Failed to create notification for director %s: %v", director.Username, err)
		}
	}

	return nil
}

// createApprovalNotifications creates notifications for approvers
func createApprovalNotifications(db *gorm.DB, request *models.ApprovalRequest, step *models.ApprovalAction) error {
	// Get users with the required role
	var users []models.User
	err := db.Where("LOWER(role) = LOWER(?) AND is_active = ?", step.Step.ApproverRole, true).Find(&users).Error
	if err != nil {
		return err
	}

	for _, user := range users {
		notification := models.Notification{
			UserID:   user.ID,
			Type:     models.NotificationTypeApprovalPending,
			Title:    fmt.Sprintf("Approval Required: %s", request.RequestTitle),
			Message:  fmt.Sprintf("You have a pending approval request for %s (Amount: %.2f)", request.RequestTitle, request.Amount),
			Priority: request.Priority,
			Data:     fmt.Sprintf(`{"request_id":%d,"entity_type":"%s","entity_id":%d,"amount":%.2f}`, request.ID, request.EntityType, request.EntityID, request.Amount),
		}

		err = db.Create(&notification).Error
		if err != nil {
			log.Printf("Failed to create notification for user %s: %v", user.Username, err)
		}
	}

	return nil
}
