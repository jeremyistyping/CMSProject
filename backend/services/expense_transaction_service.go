package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"time"
)

type ExpenseTransactionService interface {
	GetAll(filter map[string]interface{}) ([]models.ExpenseTransaction, error)
	GetByID(id uint) (*models.ExpenseTransaction, error)
	GetByProject(projectID uint, filter map[string]interface{}) ([]models.ExpenseTransaction, error)
	GetBudgetReport(projectID uint, startDate, endDate time.Time) (*models.BudgetReportResponse, error)
	Create(dto *models.CreateExpenseTransactionDTO, userID uint) (*models.ExpenseTransaction, error)
	Update(id uint, dto *models.UpdateExpenseTransactionDTO) (*models.ExpenseTransaction, error)
	Delete(id uint) error
	BatchCreate(dtos []models.CreateExpenseTransactionDTO, userID uint) ([]models.ExpenseTransaction, error)
}

type expenseTransactionService struct {
	repo        repositories.ExpenseTransactionRepository
	projectRepo repositories.ProjectRepository
	coaRepo     repositories.COARepository
}

func NewExpenseTransactionService(
	repo repositories.ExpenseTransactionRepository,
	projectRepo repositories.ProjectRepository,
	coaRepo repositories.COARepository,
) ExpenseTransactionService {
	return &expenseTransactionService{
		repo:        repo,
		projectRepo: projectRepo,
		coaRepo:     coaRepo,
	}
}

func (s *expenseTransactionService) GetAll(filter map[string]interface{}) ([]models.ExpenseTransaction, error) {
	return s.repo.GetAll(filter)
}

func (s *expenseTransactionService) GetByID(id uint) (*models.ExpenseTransaction, error) {
	return s.repo.GetByID(id)
}

func (s *expenseTransactionService) GetByProject(projectID uint, filter map[string]interface{}) ([]models.ExpenseTransaction, error) {
	return s.repo.GetByProject(projectID, filter)
}

func (s *expenseTransactionService) GetBudgetReport(projectID uint, startDate, endDate time.Time) (*models.BudgetReportResponse, error) {
	// Get project info
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, err
	}

	// Get budget vs actual summary
	summaries, err := s.repo.GetBudgetVsActualSummary(projectID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get all transactions for the period
	transactions, err := s.repo.GetByDateRange(projectID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Build report response
	report := &models.BudgetReportResponse{
		ProjectID:   projectID,
		ProjectName: project.ProjectName,
		ReportDate:  time.Now(),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Group by budget category
	labourBudget := &models.BudgetCategoryReport{
		Transactions:  []models.ExpenseTransactionDetail{},
		ByWorkPackage: []models.WorkPackageSummary{},
	}
	operationalBudget := &models.BudgetCategoryReport{
		Transactions:  []models.ExpenseTransactionDetail{},
		ByWorkPackage: []models.WorkPackageSummary{},
	}
	otherBudget := &models.BudgetCategoryReport{
		Transactions:  []models.ExpenseTransactionDetail{},
		ByWorkPackage: []models.WorkPackageSummary{},
	}

	// Calculate totals from summaries
	for _, summary := range summaries {
		switch summary.BudgetCategory {
		case models.COABudgetCategoryLabour:
			labourBudget.BudgetEstimation += summary.BudgetEstimation
			labourBudget.Actual += summary.ActualAmount
		case models.COABudgetCategoryOperasional:
			operationalBudget.BudgetEstimation += summary.BudgetEstimation
			operationalBudget.Actual += summary.ActualAmount
		case models.COABudgetCategoryOther:
			otherBudget.BudgetEstimation += summary.BudgetEstimation
			otherBudget.Actual += summary.ActualAmount
		}
	}

	labourBudget.Variance = labourBudget.BudgetEstimation - labourBudget.Actual
	operationalBudget.Variance = operationalBudget.BudgetEstimation - operationalBudget.Actual
	otherBudget.Variance = otherBudget.BudgetEstimation - otherBudget.Actual

	// Group transactions by budget category
	workPackageMap := make(map[string]*models.WorkPackageSummary)

	for _, tx := range transactions {
		detail := models.ExpenseTransactionDetail{
			Date:        tx.TransactionDate,
			Description: tx.Description,
			Unit:        tx.Unit,
			Quantity:    tx.Quantity,
			TotalPrice:  tx.Amount,
			COACode:     tx.COAAccount.Code,
			COAName:     tx.COAAccount.Name,
			WorkPackage: tx.COAAccount.WorkPackage,
			ReferenceNo: tx.ReferenceNo,
		}

		switch tx.COAAccount.BudgetCategory {
		case models.COABudgetCategoryLabour:
			labourBudget.Transactions = append(labourBudget.Transactions, detail)

		case models.COABudgetCategoryOperasional:
			operationalBudget.Transactions = append(operationalBudget.Transactions, detail)

			// Group by work package
			if tx.COAAccount.WorkPackage != "" {
				key := tx.COAAccount.WorkPackage
				if _, exists := workPackageMap[key]; !exists {
					workPackageMap[key] = &models.WorkPackageSummary{
						WorkPackage:  tx.COAAccount.WorkPackage,
						Transactions: []models.ExpenseTransactionDetail{},
					}
				}
				workPackageMap[key].Actual += tx.Amount
				workPackageMap[key].Transactions = append(workPackageMap[key].Transactions, detail)
			}

		case models.COABudgetCategoryOther:
			otherBudget.Transactions = append(otherBudget.Transactions, detail)
		}
	}

	// Add budget estimation to work packages
	for _, summary := range summaries {
		if summary.BudgetCategory == models.COABudgetCategoryOperasional && summary.WorkPackage != "" {
			if wp, exists := workPackageMap[summary.WorkPackage]; exists {
				wp.BudgetEstimation += summary.BudgetEstimation
				wp.Variance = wp.BudgetEstimation - wp.Actual
			}
		}
	}

	// Convert map to slice
	for _, wp := range workPackageMap {
		operationalBudget.ByWorkPackage = append(operationalBudget.ByWorkPackage, *wp)
	}

	report.LabourBudget = labourBudget
	report.OperationalBudget = operationalBudget
	report.OtherBudget = otherBudget

	return report, nil
}

func (s *expenseTransactionService) Create(dto *models.CreateExpenseTransactionDTO, userID uint) (*models.ExpenseTransaction, error) {
	// Parse transaction date
	transactionDate, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		return nil, errors.New("invalid transaction_date format, use YYYY-MM-DD")
	}

	// Validate project exists
	_, err = s.projectRepo.GetByID(dto.ProjectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// Validate COA exists
	coa, err := s.coaRepo.GetByID(dto.COAAccountID)
	if err != nil {
		return nil, errors.New("COA account not found")
	}

	// Validate COA is not a header account
	if coa.IsHeader {
		return nil, errors.New("cannot create transaction for header account")
	}

	// Set defaults
	if dto.Unit == "" {
		dto.Unit = "ls"
	}
	if dto.Quantity == 0 {
		dto.Quantity = 1
	}
	if dto.TransactionType == "" {
		// Auto-detect transaction type from COA budget category
		switch coa.BudgetCategory {
		case models.COABudgetCategoryLabour:
			dto.TransactionType = models.ExpenseTypeLabour
		case models.COABudgetCategoryOperasional:
			dto.TransactionType = models.ExpenseTypeMaterial
		default:
			dto.TransactionType = models.ExpenseTypeOperational
		}
	}
	if dto.ReferenceType == "" {
		dto.ReferenceType = models.ExpenseRefTypeManual
	}

	expense := &models.ExpenseTransaction{
		ProjectID:       dto.ProjectID,
		TransactionDate: transactionDate,
		COAAccountID:    dto.COAAccountID,
		Description:     dto.Description,
		Amount:          dto.Amount,
		Unit:            dto.Unit,
		Quantity:        dto.Quantity,
		TransactionType: dto.TransactionType,
		ReferenceType:   dto.ReferenceType,
		ReferenceID:     dto.ReferenceID,
		ReferenceNo:     dto.ReferenceNo,
		Notes:           dto.Notes,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(expense); err != nil {
		return nil, err
	}

	return s.repo.GetByID(expense.ID)
}

func (s *expenseTransactionService) Update(id uint, dto *models.UpdateExpenseTransactionDTO) (*models.ExpenseTransaction, error) {
	expense, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if dto.TransactionDate != nil {
		expense.TransactionDate = *dto.TransactionDate
	}
	if dto.COAAccountID != nil {
		// Validate COA exists
		coa, err := s.coaRepo.GetByID(*dto.COAAccountID)
		if err != nil {
			return nil, errors.New("COA account not found")
		}
		if coa.IsHeader {
			return nil, errors.New("cannot use header account")
		}
		expense.COAAccountID = *dto.COAAccountID
	}
	if dto.Description != "" {
		expense.Description = dto.Description
	}
	if dto.Amount != nil {
		expense.Amount = *dto.Amount
	}
	if dto.Unit != "" {
		expense.Unit = dto.Unit
	}
	if dto.Quantity != nil {
		expense.Quantity = *dto.Quantity
	}
	if dto.TransactionType != "" {
		expense.TransactionType = dto.TransactionType
	}
	if dto.ReferenceType != "" {
		expense.ReferenceType = dto.ReferenceType
	}
	if dto.ReferenceID != nil {
		expense.ReferenceID = dto.ReferenceID
	}
	if dto.ReferenceNo != "" {
		expense.ReferenceNo = dto.ReferenceNo
	}
	if dto.Notes != "" {
		expense.Notes = dto.Notes
	}

	expense.UpdatedAt = time.Now()

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return s.repo.GetByID(expense.ID)
}

func (s *expenseTransactionService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *expenseTransactionService) BatchCreate(dtos []models.CreateExpenseTransactionDTO, userID uint) ([]models.ExpenseTransaction, error) {
	var expenses []models.ExpenseTransaction

	for _, dto := range dtos {
		expense, err := s.Create(&dto, userID)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, *expense)
	}

	return expenses, nil
}
