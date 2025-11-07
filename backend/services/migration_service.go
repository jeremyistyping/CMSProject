package services

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// MigrationService handles database migrations
type MigrationService struct {
	db *sql.DB
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *sql.DB) *MigrationService {
	return &MigrationService{db: db}
}

// RunMigrations runs all pending migrations automatically
// This is called during application startup
func (s *MigrationService) RunMigrations(migrationsFS embed.FS, migrationsDir string) error {
	log.Println("🔄 Starting database migrations...")

	// Create source from embedded filesystem
	sourceDriver, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to create migration source: %v", err)
	}

	// Create database driver
	dbDriver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %v", err)
	}

	// Create migrator
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %v", err)
	}

	if dirty {
		log.Printf("⚠️ Database is in dirty state at version %d, attempting to fix...", currentVersion)
		// Force version to clean state
		if err := m.Force(int(currentVersion)); err != nil {
			return fmt.Errorf("failed to force version: %v", err)
		}
	}

	if err != migrate.ErrNilVersion {
		log.Printf("📊 Current migration version: %d", currentVersion)
	} else {
		log.Println("📊 No migrations applied yet")
	}

	// Run migrations
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✅ Database is up to date - no migrations needed")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Get new version
	newVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get new version: %v", err)
	}

	log.Printf("✅ Migrations completed successfully - now at version %d", newVersion)
	return nil
}

// RollbackMigration rolls back one migration step
func (s *MigrationService) RollbackMigration(migrationsFS embed.FS, migrationsDir string) error {
	log.Println("🔄 Rolling back last migration...")

	sourceDriver, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to create migration source: %v", err)
	}

	dbDriver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}

	currentVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current version: %v", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback: %v", err)
	}

	log.Printf("✅ Rolled back from version %d", currentVersion)
	return nil
}

// GetCurrentVersion returns the current migration version
func (s *MigrationService) GetCurrentVersion(migrationsFS embed.FS, migrationsDir string) (uint, bool, error) {
	sourceDriver, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return 0, false, err
	}

	dbDriver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return 0, false, err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return 0, false, err
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	return version, dirty, err
}
