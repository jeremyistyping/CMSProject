package services

import (
	"fmt"

	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

type DashboardService struct {
	DB *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{DB: db}
}

// AnalyticsData represents the complete analytics data for project management
type AnalyticsData struct {
	TotalProjects        int64   `json:"totalProjects"`
	ActiveProjects       int64   `json:"activeProjects"`
	CompletedProjects    int64   `json:"completedProjects"`
	TotalPurchaseRequests int64  `json:"totalPurchaseRequests"`
	PendingApprovals     int64   `json:"pendingApprovals"`
	TotalBudget          float64 `json:"totalBudget"`
	TotalSpent           float64 `json:"totalSpent"`
	
	// Monthly data
	MonthlyProjects      []MonthlyData           `json:"monthlyProjects"`
	MonthlyPRs           []MonthlyData           `json:"monthlyPRs"`
	RecentProjects       []ProjectSummary        `json:"recentProjects"`
	RecentPurchaseRequests []PurchaseRequestSummary `json:"recentPurchaseRequests"`
}

type MonthlyData struct {
	Month string  `json:"month"`
	Value float64 `json:"value"`
}

type DashboardCashFlowData struct {
	Month   string  `json:"month"`
	Inflow  float64 `json:"inflow"`
	Outflow float64 `json:"outflow"`
	Balance float64 `json:"balance"`
}

type AccountData struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
	Type    string  `json:"type"`
}

type TransactionData struct {
	ID            uint    `json:"id"`
	TransactionID string  `json:"transaction_id"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
	Type          string  `json:"type"`
	AccountName   string  `json:"account_name"`
	ContactName   *string `json:"contact_name"`
	Status        string  `json:"status"`
}

type ProjectSummary struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	Budget    float64 `json:"budget"`
	CreatedAt string  `json:"created_at"`
}

type PurchaseRequestSummary struct {
	ID          uint    `json:"id"`
	PRNumber    string  `json:"pr_number"`
	ProjectName string  `json:"project_name"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// GetDashboardAnalytics returns comprehensive dashboard analytics
func (ds *DashboardService) GetDashboardAnalytics() (*AnalyticsData, error) {
	return ds.GetDashboardAnalyticsForRole("")
}

// GetDashboardAnalyticsForRole returns dashboard analytics filtered by user role
func (ds *DashboardService) GetDashboardAnalyticsForRole(role string) (*AnalyticsData, error) {
	analytics := &AnalyticsData{}
	
	// Total projects
	ds.DB.Model(&models.Project{}).Count(&analytics.TotalProjects)
	
	// Active projects (case-insensitive check for various status values)
	ds.DB.Model(&models.Project{}).Where("LOWER(status) IN ?", []string{"active", "in_progress", "ongoing"}).Count(&analytics.ActiveProjects)
	
	// Completed projects (case-insensitive)
	ds.DB.Model(&models.Project{}).Where("LOWER(status) = ?", "completed").Count(&analytics.CompletedProjects)
	
	// Total purchase requests
	ds.DB.Model(&models.PurchaseRequest{}).Count(&analytics.TotalPurchaseRequests)
	
	// Pending approvals
	ds.DB.Model(&models.ApprovalRequest{}).Where("status = ?", "PENDING").Count(&analytics.PendingApprovals)
	
	// Total budget from projects
	ds.DB.Model(&models.Project{}).Select("COALESCE(SUM(budget), 0)").Scan(&analytics.TotalBudget)
	
	// Total spent from approved purchase requests
	ds.DB.Model(&models.PurchaseRequest{}).
		Where("status = ?", "APPROVED").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&analytics.TotalSpent)
	
	// Get monthly projects data
	var err error
	analytics.MonthlyProjects, err = ds.getMonthlyProjectsData()
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly projects data: %v", err)
	}
	
	// Get monthly purchase requests data
	analytics.MonthlyPRs, err = ds.getMonthlyPRsData()
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly PRs data: %v", err)
	}
	
	// Get recent projects
	analytics.RecentProjects, err = ds.getRecentProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent projects: %v", err)
	}
	
	// Get recent purchase requests
	analytics.RecentPurchaseRequests, err = ds.getRecentPurchaseRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent purchase requests: %v", err)
	}
	
	return analytics, nil
}

// getMonthlyProjectsData gets project creation data for the last 7 months
func (ds *DashboardService) getMonthlyProjectsData() ([]MonthlyData, error) {
	type QueryResult struct {
		Month string  `json:"month"`
		Value float64 `json:"value"`
	}
	
	var results []QueryResult
	err := ds.DB.Raw(`
		SELECT 
			TO_CHAR(created_at, 'Mon') as month,
			COUNT(*) as value
		FROM projects 
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 months'
			AND deleted_at IS NULL
		GROUP BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), TO_CHAR(created_at, 'Mon')
		ORDER BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at)
	`).Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	var data []MonthlyData
	for _, result := range results {
		data = append(data, MonthlyData{
			Month: result.Month,
			Value: result.Value,
		})
	}
	
	return data, nil
}

// getMonthlyPRsData gets purchase request data for the last 7 months
func (ds *DashboardService) getMonthlyPRsData() ([]MonthlyData, error) {
	type QueryResult struct {
		Month string  `json:"month"`
		Value float64 `json:"value"`
	}
	
	var results []QueryResult
	err := ds.DB.Raw(`
		SELECT 
			TO_CHAR(created_at, 'Mon') as month,
			COALESCE(SUM(total_amount), 0) as value
		FROM purchase_requests 
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 months'
			AND deleted_at IS NULL
		GROUP BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), TO_CHAR(created_at, 'Mon')
		ORDER BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at)
	`).Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	var data []MonthlyData
	for _, result := range results {
		data = append(data, MonthlyData{
			Month: result.Month,
			Value: result.Value,
		})
	}
	
	return data, nil
}

// getRecentProjects gets recent projects
func (ds *DashboardService) getRecentProjects() ([]ProjectSummary, error) {
	var projects []models.Project
	err := ds.DB.Order("created_at DESC").Limit(5).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	
	var summaries []ProjectSummary
	for _, p := range projects {
		summaries = append(summaries, ProjectSummary{
			ID:        p.ID,
			Name:      p.ProjectName,
			Status:    p.Status,
			Progress:  p.OverallProgress,
			Budget:    p.Budget,
			CreatedAt: p.CreatedAt.Format("2006-01-02"),
		})
	}
	
	return summaries, nil
}

// getRecentPurchaseRequests gets recent purchase requests
func (ds *DashboardService) getRecentPurchaseRequests() ([]PurchaseRequestSummary, error) {
	var prs []models.PurchaseRequest
	err := ds.DB.Preload("Project").Order("created_at DESC").Limit(5).Find(&prs).Error
	if err != nil {
		return nil, err
	}
	
	var summaries []PurchaseRequestSummary
	for _, pr := range prs {
		projectName := ""
		if pr.Project.ID > 0 {
			projectName = pr.Project.ProjectName
		}
		summaries = append(summaries, PurchaseRequestSummary{
			ID:          pr.ID,
			PRNumber:    pr.Code,
			ProjectName: projectName,
			TotalAmount: pr.TotalAmount,
			Status:      pr.Status,
			CreatedAt:   pr.CreatedAt.Format("2006-01-02"),
		})
	}
	
	return summaries, nil
}

// GetQuickStats returns quick statistics for dashboard widgets
func (ds *DashboardService) GetQuickStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total projects
	var totalProjects int64
	ds.DB.Model(&models.Project{}).Count(&totalProjects)
	stats["total_projects"] = totalProjects
	
	// Active projects (case-insensitive)
	var activeProjects int64
	ds.DB.Model(&models.Project{}).Where("LOWER(status) IN ?", []string{"active", "in_progress", "ongoing"}).Count(&activeProjects)
	stats["active_projects"] = activeProjects
	
	// Pending purchase requests (case-insensitive)
	var pendingPRs int64
	ds.DB.Model(&models.PurchaseRequest{}).Where("LOWER(status) = ?", "pending").Count(&pendingPRs)
	stats["pending_purchase_requests"] = pendingPRs
	
	// Pending approvals (case-insensitive)
	var pendingApprovals int64
	ds.DB.Model(&models.ApprovalRequest{}).Where("LOWER(status) = ?", "pending").Count(&pendingApprovals)
	stats["pending_approvals"] = pendingApprovals
	
	// Today's daily updates
	var todayUpdates int64
	ds.DB.Model(&models.DailyUpdate{}).
		Where("DATE(created_at) = CURRENT_DATE").
		Count(&todayUpdates)
	stats["today_daily_updates"] = todayUpdates
	
	return stats, nil
}
