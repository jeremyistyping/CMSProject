package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ExpenseTransactionRepository interface {
	GetAll(filter map[string]interface{}) ([]models.ExpenseTransaction, error)
	GetByID(id uint) (*models.ExpenseTransaction, error)
	GetByProject(projectID uint, filter map[string]interface{}) ([]models.ExpenseTransaction, error)
	GetBudgetVsActualSummary(projectID uint, startDate, endDate time.Time) ([]models.BudgetVsActualSummary, error)
	GetByDateRange(projectID uint, startDate, endDate time.Time) ([]models.ExpenseTransaction, error)
	GetByBudgetCategory(projectID uint, budgetCategory string, startDate, endDate time.Time) ([]models.ExpenseTransaction, error)
	Create(expense *models.ExpenseTransaction) error
	Update(expense *models.ExpenseTransaction) error
	Delete(id uint) error
	GetTotalByProject(projectID uint) (float64, error)
	GetTotalByCOA(projectID uint, coaAccountID uint) (float64, error)
}

type expenseTransactionRepository struct {
	db *gorm.DB
}

func NewExpenseTransactionRepository(db *gorm.DB) ExpenseTransactionRepository {
	return &expenseTransactionRepository{db: db}
}

func (r *expenseTransactionRepository) GetAll(filter map[string]interface{}) ([]models.ExpenseTransaction, error) {
	var expenses []models.ExpenseTransaction
	query := r.db.Where("deleted_at IS NULL")

	if projectID, ok := filter["project_id"]; ok {
		query = query.Where("project_id = ?", projectID)
	}
	if coaAccountID, ok := filter["coa_account_id"]; ok {
		query = query.Where("coa_account_id = ?", coaAccountID)
	}
	if transactionType, ok := filter["transaction_type"]; ok {
		query = query.Where("transaction_type = ?", transactionType)
	}
	if startDate, ok := filter["start_date"]; ok {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate, ok := filter["end_date"]; ok {
		query = query.Where("transaction_date <= ?", endDate)
	}
	if search, ok := filter["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Where("description ILIKE ? OR reference_no ILIKE ?", searchStr, searchStr)
	}

	err := query.
		Preload("Project").
		Preload("COAAccount").
		Preload("Creator").
		Order("transaction_date DESC, id DESC").
		Find(&expenses).Error

	return expenses, err
}

func (r *expenseTransactionRepository) GetByID(id uint) (*models.ExpenseTransaction, error) {
	var expense models.ExpenseTransaction
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Project").
		Preload("COAAccount").
		Preload("Creator").
		First(&expense).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense transaction not found")
		}
		return nil, err
	}
	return &expense, nil
}

func (r *expenseTransactionRepository) GetByProject(projectID uint, filter map[string]interface{}) ([]models.ExpenseTransaction, error) {
	filter["project_id"] = projectID
	return r.GetAll(filter)
}

func (r *expenseTransactionRepository) GetBudgetVsActualSummary(projectID uint, startDate, endDate time.Time) ([]models.BudgetVsActualSummary, error) {
	var summaries []models.BudgetVsActualSummary

	// Combined query: Get budget from project_budgets AND actual from expense_transactions
	// Also include COAs that have transactions but no budget
	query := `
		WITH budget_data AS (
			SELECT 
				pb.project_id,
				pb.account_id as coa_account_id,
				pb.estimated_amount as budget_estimation
			FROM project_budgets pb
			WHERE pb.project_id = ? 
				AND pb.deleted_at IS NULL
		),
		actual_data AS (
			SELECT 
				et.project_id,
				et.coa_account_id,
				SUM(et.amount) as actual_amount
			FROM expense_transactions et
			WHERE et.project_id = ?
				AND et.deleted_at IS NULL
				AND et.transaction_date BETWEEN ? AND ?
			GROUP BY et.project_id, et.coa_account_id
		),
		all_coa AS (
			SELECT DISTINCT coa_account_id, project_id FROM budget_data
			UNION
			SELECT DISTINCT coa_account_id, project_id FROM actual_data
		)
		SELECT 
			ac.project_id,
			ac.coa_account_id,
			coa.code as coa_code,
			coa.name as coa_name,
			coa.budget_category,
			coa.work_package,
			COALESCE(bd.budget_estimation, 0) as budget_estimation,
			COALESCE(ad.actual_amount, 0) as actual_amount,
			COALESCE(bd.budget_estimation, 0) - COALESCE(ad.actual_amount, 0) as variance,
			CASE 
				WHEN COALESCE(bd.budget_estimation, 0) > 0 THEN 
					((COALESCE(bd.budget_estimation, 0) - COALESCE(ad.actual_amount, 0)) / bd.budget_estimation * 100)
				ELSE 0 
			END as variance_percentage
		FROM all_coa ac
		JOIN coa_accounts coa ON ac.coa_account_id = coa.id
		LEFT JOIN budget_data bd ON bd.coa_account_id = ac.coa_account_id
		LEFT JOIN actual_data ad ON ad.coa_account_id = ac.coa_account_id
		WHERE coa.deleted_at IS NULL
		ORDER BY coa.budget_category, coa.code
	`

	err := r.db.Raw(query, projectID, projectID, startDate, endDate).Scan(&summaries).Error
	return summaries, err
}

func (r *expenseTransactionRepository) GetByDateRange(projectID uint, startDate, endDate time.Time) ([]models.ExpenseTransaction, error) {
	var expenses []models.ExpenseTransaction
	err := r.db.Where("project_id = ? AND transaction_date BETWEEN ? AND ? AND deleted_at IS NULL", 
		projectID, startDate, endDate).
		Preload("Project").
		Preload("COAAccount").
		Preload("Creator").
		Order("transaction_date ASC, id ASC").
		Find(&expenses).Error

	return expenses, err
}

func (r *expenseTransactionRepository) GetByBudgetCategory(projectID uint, budgetCategory string, startDate, endDate time.Time) ([]models.ExpenseTransaction, error) {
	var expenses []models.ExpenseTransaction
	err := r.db.
		Joins("JOIN coa_accounts ON coa_accounts.id = expense_transactions.coa_account_id").
		Where("expense_transactions.project_id = ?", projectID).
		Where("coa_accounts.budget_category = ?", budgetCategory).
		Where("expense_transactions.transaction_date BETWEEN ? AND ?", startDate, endDate).
		Where("expense_transactions.deleted_at IS NULL").
		Preload("Project").
		Preload("COAAccount").
		Preload("Creator").
		Order("expense_transactions.transaction_date ASC, expense_transactions.id ASC").
		Find(&expenses).Error

	return expenses, err
}

func (r *expenseTransactionRepository) Create(expense *models.ExpenseTransaction) error {
	return r.db.Create(expense).Error
}

func (r *expenseTransactionRepository) Update(expense *models.ExpenseTransaction) error {
	return r.db.Save(expense).Error
}

func (r *expenseTransactionRepository) Delete(id uint) error {
	return r.db.Model(&models.ExpenseTransaction{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *expenseTransactionRepository) GetTotalByProject(projectID uint) (float64, error) {
	var total float64
	err := r.db.Model(&models.ExpenseTransaction{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *expenseTransactionRepository) GetTotalByCOA(projectID uint, coaAccountID uint) (float64, error) {
	var total float64
	err := r.db.Model(&models.ExpenseTransaction{}).
		Where("project_id = ? AND coa_account_id = ? AND deleted_at IS NULL", projectID, coaAccountID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
