package repositories

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
	"app-sistem-akuntansi/utils"
)

// BaseRepository defines common database operations
type BaseRepository interface {
	// Transaction methods
	BeginTx(ctx context.Context) (*gorm.DB, error)
	WithTx(tx *gorm.DB) BaseRepository
	
	// Health check
	Health(ctx context.Context) error
}

// BaseRepo implements BaseRepository
type BaseRepo struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) BaseRepository {
	return &BaseRepo{DB: db}
}

// BeginTx starts a new transaction
func (r *BaseRepo) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := r.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, utils.NewDatabaseError("begin transaction", tx.Error)
	}
	return tx, nil
}

// WithTx returns a new repository instance with the given transaction
func (r *BaseRepo) WithTx(tx *gorm.DB) BaseRepository {
	return &BaseRepo{DB: tx}
}

// Health checks database connectivity
func (r *BaseRepo) Health(ctx context.Context) error {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return utils.NewDatabaseError("get sql db", err)
	}
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return utils.NewDatabaseError("ping database", err)
	}
	
	return nil
}

// Common query options
type QueryOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string // ASC or DESC
	Preload []string
}

// ApplyQueryOptions applies common query options to GORM DB
func ApplyQueryOptions(db *gorm.DB, opts *QueryOptions) *gorm.DB {
	if opts == nil {
		return db
	}
	
	// Apply preloads
	for _, preload := range opts.Preload {
		db = db.Preload(preload)
	}
	
	// Apply sorting
	if opts.Sort != "" {
		order := "ASC"
		if opts.Order == "DESC" {
			order = "DESC"
		}
		db = db.Order(opts.Sort + " " + order)
	}
	
	// Apply pagination
	if opts.Limit > 0 {
		db = db.Limit(opts.Limit)
	}
	
	if opts.Offset > 0 {
		db = db.Offset(opts.Offset)
	}
	
	return db
}

// PaginationResult represents paginated results
type PaginationResult struct {
	Total       int64 `json:"total"`
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(total int64, page, perPage int) *PaginationResult {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	
	return &PaginationResult{
		Total:       total,
		CurrentPage: page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}
}

// FilterOptions represents common filtering options
type FilterOptions struct {
	Search    string
	StartDate *string
	EndDate   *string
	Status    *string
	UserID    *uint
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	TotalConnections int `json:"total_connections"`
	ActiveConnections int `json:"active_connections"`
	IdleConnections  int `json:"idle_connections"`
}

// GetDatabaseStats returns database connection statistics
func (r *BaseRepo) GetDatabaseStats(ctx context.Context) (*DatabaseStats, error) {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return nil, utils.NewDatabaseError("get sql db", err)
	}
	
	stats := sqlDB.Stats()
	
	return &DatabaseStats{
		TotalConnections:  stats.OpenConnections,
		ActiveConnections: stats.InUse,
		IdleConnections:   stats.Idle,
	}, nil
}

// ExecRaw executes raw SQL query
func (r *BaseRepo) ExecRaw(ctx context.Context, query string, args ...interface{}) error {
	result := r.DB.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return utils.NewDatabaseError("execute raw query", result.Error)
	}
	return nil
}

// QueryRaw executes raw SQL query and returns rows
func (r *BaseRepo) QueryRaw(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return nil, utils.NewDatabaseError("get sql db", err)
	}
	
	rows, err := sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.NewDatabaseError("query raw", err)
	}
	
	return rows, nil
}
