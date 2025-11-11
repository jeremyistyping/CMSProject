package services

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ProjectReportService struct {
	db *gorm.DB
}

func NewProjectReportService(db *gorm.DB) *ProjectReportService {
	return &ProjectReportService{db: db}
}

// GenerateBudgetVsActualReport - Generate Budget vs Actual Report
func (s *ProjectReportService) GenerateBudgetVsActualReport(params models.ProjectReportParams) (*models.BudgetVsActualReport, error) {
	report := &models.BudgetVsActualReport{
		ReportDate: time.Now(),
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		ProjectID:  params.ProjectID,
		COAGroups:  []models.BudgetVsActualCOAGroup{},
	}

	// Get project info if project_id specified
	if params.ProjectID != nil {
		var project models.Project
		if err := s.db.First(&project, *params.ProjectID).Error; err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		report.ProjectName = project.ProjectName
	}

	// Query untuk get budget dan actual per COA
	query := `
		SELECT 
			a.code as coa_code,
			a.name as coa_name,
			a.type as coa_type,
			COALESCE(budget.estimated_amount, 0) as budget,
			COALESCE(actual.total_amount, 0) as actual
		FROM accounts a
		LEFT JOIN (
			SELECT 
				account_id,
				SUM(estimated_amount) as estimated_amount
			FROM project_budgets
			WHERE deleted_at IS NULL
	`

	budgetParams := []interface{}{}

	if params.ProjectID != nil {
		query += " AND project_id = ?"
		budgetParams = append(budgetParams, *params.ProjectID)
	}

	query += `
			GROUP BY account_id
		) budget ON budget.account_id = a.id
		LEFT JOIN (
			SELECT 
				ujl.account_id,
				SUM(ABS(ujl.amount)) as total_amount
			FROM unified_journal_ledger ujl
			WHERE ujl.status = 'POSTED'
			AND ujl.entry_date BETWEEN ? AND ?
	`

	actualParams := append(budgetParams, params.StartDate, params.EndDate)

	if params.ProjectID != nil {
		query += " AND ujl.project_id = ?"
		actualParams = append(actualParams, *params.ProjectID)
	}

	query += `
			GROUP BY ujl.account_id
		) actual ON actual.account_id = a.id
		WHERE a.deleted_at IS NULL
		AND (budget.estimated_amount > 0 OR actual.total_amount > 0)
		ORDER BY a.code
	`

	rows, err := s.db.Raw(query, actualParams...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query budget vs actual: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var group models.BudgetVsActualCOAGroup
		if err := rows.Scan(&group.COACode, &group.COAName, &group.COAType, &group.Budget, &group.Actual); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Calculate variance
		group.Variance = group.Actual - group.Budget

		// Calculate variance rate
		if group.Budget > 0 {
			group.VarianceRate = (group.Variance / group.Budget) * 100
		}

		// Determine status
		if group.Variance > group.Budget*0.1 {
			group.Status = "OVER_BUDGET"
		} else if group.Variance < -group.Budget*0.1 {
			group.Status = "UNDER_BUDGET"
		} else {
			group.Status = "ON_TARGET"
		}

		report.COAGroups = append(report.COAGroups, group)
		report.TotalBudget += group.Budget
		report.TotalActual += group.Actual
	}

	// Calculate total variance
	report.TotalVariance = report.TotalActual - report.TotalBudget
	if report.TotalBudget > 0 {
		report.VarianceRate = (report.TotalVariance / report.TotalBudget) * 100
	}

	return report, nil
}

// GenerateProfitabilityReport - Generate Profitability Report per Project
func (s *ProjectReportService) GenerateProfitabilityReport(params models.ProjectReportParams) (*models.ProfitabilityReport, error) {
	report := &models.ProfitabilityReport{
		ReportDate: time.Now(),
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Projects:   []models.ProjectProfitability{},
	}

	// Query projects
	projectQuery := s.db.Model(&models.Project{}).Where("deleted_at IS NULL")
	if params.ProjectID != nil {
		projectQuery = projectQuery.Where("id = ?", *params.ProjectID)
	}

	var projects []models.Project
	if err := projectQuery.Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	for _, project := range projects {
		profitability := models.ProjectProfitability{
			ProjectID:     project.ID,
			ProjectCode:   "", // Project model doesn't have Code field
			ProjectName:   project.ProjectName,
			ProjectStatus: project.Status,
		}

		// Calculate Revenue (from REVENUE accounts)
		var revenue float64
		s.db.Raw(`
			SELECT COALESCE(SUM(ABS(ujl.amount)), 0)
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.project_id = ?
			AND ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND a.type = 'REVENUE'
			AND a.deleted_at IS NULL
		`, project.ID, params.StartDate, params.EndDate).Scan(&revenue)
		profitability.Revenue = revenue

		// Calculate Direct Cost (Material, Labour, Equipment)
		var directCost float64
		s.db.Raw(`
			SELECT COALESCE(SUM(ABS(ujl.amount)), 0)
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.project_id = ?
			AND ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND a.type = 'EXPENSE'
			AND a.category IN ('MATERIAL', 'LABOUR', 'EQUIPMENT', 'SUBCONTRACTOR')
			AND a.deleted_at IS NULL
		`, project.ID, params.StartDate, params.EndDate).Scan(&directCost)
		profitability.DirectCost = directCost

		// Calculate Operational Cost (Overhead, Admin)
		var operationalCost float64
		s.db.Raw(`
			SELECT COALESCE(SUM(ABS(ujl.amount)), 0)
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.project_id = ?
			AND ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND a.type = 'EXPENSE'
			AND a.category IN ('OVERHEAD', 'OPERATIONAL', 'ADMINISTRATIVE')
			AND a.deleted_at IS NULL
		`, project.ID, params.StartDate, params.EndDate).Scan(&operationalCost)
		profitability.OperationalCost = operationalCost

		// Calculate metrics
		profitability.TotalCost = directCost + operationalCost
		profitability.GrossProfit = revenue - directCost
		profitability.NetProfit = revenue - profitability.TotalCost

		if revenue > 0 {
			profitability.GrossProfitMargin = (profitability.GrossProfit / revenue) * 100
			profitability.NetProfitMargin = (profitability.NetProfit / revenue) * 100
		}

		report.Projects = append(report.Projects, profitability)
		report.TotalRevenue += profitability.Revenue
		report.TotalDirectCost += profitability.DirectCost
		report.TotalOperational += profitability.OperationalCost
		report.TotalProfit += profitability.NetProfit
	}

	// Calculate overall margin
	if report.TotalRevenue > 0 {
		report.OverallMargin = (report.TotalProfit / report.TotalRevenue) * 100
	}

	return report, nil
}

// GenerateCashFlowReport - Generate Cash Flow Report per Project
func (s *ProjectReportService) GenerateCashFlowReport(params models.ProjectReportParams) (*models.CashFlowReport, error) {
	report := &models.CashFlowReport{
		ReportDate: time.Now(),
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Projects:   []models.ProjectCashFlow{},
	}

	// Get beginning balance (Cash/Bank accounts before start date)
	s.db.Raw(`
		SELECT COALESCE(SUM(ujl.amount), 0)
		FROM unified_journal_ledger ujl
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'ASSET'
		AND a.category IN ('CASH', 'BANK')
		AND ujl.entry_date < ?
		AND ujl.status = 'POSTED'
		AND a.deleted_at IS NULL
	`, params.StartDate).Scan(&report.BeginningBalance)

	// Query projects
	projectQuery := s.db.Model(&models.Project{}).Where("deleted_at IS NULL")
	if params.ProjectID != nil {
		projectQuery = projectQuery.Where("id = ?", *params.ProjectID)
	}

	var projects []models.Project
	if err := projectQuery.Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	for _, project := range projects {
		cashFlow := models.ProjectCashFlow{
			ProjectID:   project.ID,
			ProjectCode: "", // Project model doesn't have Code field
			ProjectName: project.ProjectName,
		}

		// Get cash in (deposits, revenue received)
		var cashInDetails []models.CashFlowDetail
		rows, _ := s.db.Raw(`
			SELECT 
				ujl.entry_date,
				ujl.description,
				a.code,
				a.name,
				ABS(ujl.amount),
				ujl.source_type
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.project_id = ?
			AND ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND ujl.entry_type = 'DEBIT'
			AND a.type = 'ASSET'
			AND a.category IN ('CASH', 'BANK')
			AND a.deleted_at IS NULL
			ORDER BY ujl.entry_date
		`, project.ID, params.StartDate, params.EndDate).Rows()

		for rows.Next() {
			var detail models.CashFlowDetail
			rows.Scan(&detail.Date, &detail.Description, &detail.COACode, &detail.COAName, &detail.Amount, &detail.Source)
			cashInDetails = append(cashInDetails, detail)
			cashFlow.CashIn += detail.Amount
		}
		rows.Close()
		cashFlow.CashInDetails = cashInDetails

		// Get cash out (expenses, payments)
		var cashOutDetails []models.CashFlowDetail
		rows2, _ := s.db.Raw(`
			SELECT 
				ujl.entry_date,
				ujl.description,
				a.code,
				a.name,
				ABS(ujl.amount),
				ujl.source_type
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.project_id = ?
			AND ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND ujl.entry_type = 'CREDIT'
			AND a.type = 'ASSET'
			AND a.category IN ('CASH', 'BANK')
			AND a.deleted_at IS NULL
			ORDER BY ujl.entry_date
		`, project.ID, params.StartDate, params.EndDate).Rows()

		for rows2.Next() {
			var detail models.CashFlowDetail
			rows2.Scan(&detail.Date, &detail.Description, &detail.COACode, &detail.COAName, &detail.Amount, &detail.Source)
			cashOutDetails = append(cashOutDetails, detail)
			cashFlow.CashOut += detail.Amount
		}
		rows2.Close()
		cashFlow.CashOutDetails = cashOutDetails

		cashFlow.NetCashFlow = cashFlow.CashIn - cashFlow.CashOut

		report.Projects = append(report.Projects, cashFlow)
		report.TotalCashIn += cashFlow.CashIn
		report.TotalCashOut += cashFlow.CashOut
	}

	report.NetCashFlow = report.TotalCashIn - report.TotalCashOut
	report.EndingBalance = report.BeginningBalance + report.NetCashFlow

	return report, nil
}

// GenerateCostSummaryReport - Generate Cost Summary Report
func (s *ProjectReportService) GenerateCostSummaryReport(params models.ProjectReportParams) (*models.CostSummaryReport, error) {
	report := &models.CostSummaryReport{
		ReportDate: time.Now(),
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		ProjectID:  params.ProjectID,
		Categories: []models.CostCategory{},
	}

	// Get project info if specified
	if params.ProjectID != nil {
		var project models.Project
		if err := s.db.First(&project, *params.ProjectID).Error; err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		report.ProjectName = project.ProjectName
	}

	// Get cost categories
	categoryQuery := `
		SELECT 
			COALESCE(a.category, 'UNCATEGORIZED') as category_code,
			COALESCE(a.category, 'Uncategorized') as category_name,
			COUNT(*) as item_count,
			SUM(ABS(ujl.amount)) as total_amount
		FROM unified_journal_ledger ujl
		JOIN accounts a ON a.id = ujl.account_id
		WHERE ujl.entry_date BETWEEN ? AND ?
		AND ujl.status = 'POSTED'
		AND a.type = 'EXPENSE'
		AND a.deleted_at IS NULL
	`

	queryParams := []interface{}{params.StartDate, params.EndDate}

	if params.ProjectID != nil {
		categoryQuery += " AND ujl.project_id = ?"
		queryParams = append(queryParams, *params.ProjectID)
	}

	categoryQuery += ` GROUP BY a.category ORDER BY total_amount DESC`

	rows, err := s.db.Raw(categoryQuery, queryParams...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query cost categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category models.CostCategory
		rows.Scan(&category.CategoryCode, &category.CategoryName, &category.ItemCount, &category.TotalAmount)

		// Get items for this category
		itemQuery := `
			SELECT 
				ujl.entry_date,
				a.code,
				a.name,
				ujl.description,
				ABS(ujl.amount),
				'',
				COALESCE(p.project_name, '')
			FROM unified_journal_ledger ujl
			JOIN accounts a ON a.id = ujl.account_id
			LEFT JOIN projects p ON p.id = ujl.project_id
			WHERE ujl.entry_date BETWEEN ? AND ?
			AND ujl.status = 'POSTED'
			AND a.type = 'EXPENSE'
			AND COALESCE(a.category, 'UNCATEGORIZED') = ?
			AND a.deleted_at IS NULL
		`

		itemParams := []interface{}{params.StartDate, params.EndDate, category.CategoryCode}

		if params.ProjectID != nil {
			itemQuery += " AND ujl.project_id = ?"
			itemParams = append(itemParams, *params.ProjectID)
		}

		itemQuery += " ORDER BY ujl.entry_date DESC LIMIT 100"

		itemRows, _ := s.db.Raw(itemQuery, itemParams...).Rows()
		for itemRows.Next() {
			var item models.CostItem
			itemRows.Scan(&item.Date, &item.COACode, &item.COAName, &item.Description, &item.Amount, &item.ProjectCode, &item.ProjectName)
			category.Items = append(category.Items, item)
		}
		itemRows.Close()

		report.Categories = append(report.Categories, category)
		report.TotalCost += category.TotalAmount

		// Track largest category
		if category.TotalAmount > report.LargestAmount {
			report.LargestAmount = category.TotalAmount
			report.LargestCategory = category.CategoryName
		}
	}

	// Calculate percentages
	if report.TotalCost > 0 {
		for i := range report.Categories {
			report.Categories[i].Percentage = (report.Categories[i].TotalAmount / report.TotalCost) * 100
		}
	}

	return report, nil
}
