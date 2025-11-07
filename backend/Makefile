# SSOT Migration Makefile
# =====================
# Commands for migrating to Single Source of Truth journal system

.PHONY: help migrate-ssot cleanup-models update-routes check-status full-migration

# Default target
help:
	@echo "🚀 SSOT Migration Commands"
	@echo "========================="
	@echo ""
	@echo "Available commands:"
	@echo "  make migrate-ssot    - Migrate journal data to SSOT"
	@echo "  make cleanup-models  - Archive old models and migrations"
	@echo "  make update-routes   - Update routes to use SSOT only"
	@echo "  make check-status    - Check SSOT database status"
	@echo "  make full-migration  - Run complete migration (all above)"
	@echo ""
	@echo "🔧 Development commands:"
	@echo "  make build-scripts   - Build all migration scripts"
	@echo "  make test-ssot       - Test SSOT API endpoints"
	@echo "  make rollback        - Show rollback instructions"
	@echo ""

# Build all migration scripts
build-scripts:
	@echo "🔨 Building migration scripts..."
	@go build -o bin/migrate_to_full_ssot cmd/scripts/migrate_to_full_ssot.go
	@go build -o bin/cleanup_old_models cmd/scripts/cleanup_old_models.go  
	@go build -o bin/update_routes_to_ssot cmd/scripts/update_routes_to_ssot.go
	@echo "✅ Scripts built in ./bin/"

# Migrate journal data to SSOT
migrate-ssot: build-scripts
	@echo "🚀 Starting SSOT migration..."
	@echo "This will migrate all journal data to SSOT structure"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@./bin/migrate_to_full_ssot

# Cleanup old models and migrations
cleanup-models: build-scripts
	@echo "🧹 Cleaning up old models..."
	@echo "This will archive old journal models and migrations"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@./bin/cleanup_old_models

# Update routes to SSOT only
update-routes: build-scripts
	@echo "🔄 Updating routes to SSOT..."
	@echo "This will remove old journal routes and keep SSOT only"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@./bin/update_routes_to_ssot

# Check SSOT database status  
check-status:
	@echo "📊 Checking SSOT status..."
	@go run cmd/scripts/check_ssot_status.go

# Full migration (run all steps)
full-migration:
	@echo "🎯 Starting Full SSOT Migration"
	@echo "==============================="
	@echo "This will:"
	@echo "1. Migrate all journal data to SSOT"
	@echo "2. Archive old models and migrations"
	@echo "3. Update routes to use SSOT only"
	@echo "4. Clean up old references"
	@echo ""
	@echo "⚠️  Make sure you have:"
	@echo "• Database backup"
	@echo "• Code backup (git commit)"
	@echo "• Stopped running servers"
	@echo ""
	@read -p "Ready to proceed with full migration? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@make migrate-ssot
	@make cleanup-models
	@make update-routes
	@echo ""
	@echo "🎉 Full SSOT Migration Completed!"
	@echo "================================"
	@make check-status

# Test SSOT API endpoints
test-ssot:
	@echo "🧪 Testing SSOT API endpoints..."
	@echo "Starting backend server for testing..."
	@go run main.go &
	@SERVER_PID=$$!; \
	sleep 5; \
	echo "Testing endpoints..."; \
	curl -s -X GET http://localhost:8080/api/v1/journals | head -c 100; \
	echo ""; \
	echo "✅ Basic test completed"; \
	kill $$SERVER_PID

# Show rollback instructions
rollback:
	@echo "🔄 SSOT Rollback Instructions"
	@echo "============================"
	@echo ""
	@echo "If you need to rollback to the old journal system:"
	@echo ""
	@echo "1. Database rollback:"
	@echo "   - Rename archived tables back:"
	@echo "     ALTER TABLE journal_entries_archived_YYYYMMDD_HHMMSS RENAME TO journal_entries;"
	@echo "     ALTER TABLE journals_archived_YYYYMMDD_HHMMSS RENAME TO journals;"
	@echo ""
	@echo "2. Code rollback:"
	@echo "   - Restore routes: mv routes/api.go.backup routes/api.go"
	@echo "   - Restore main.go: mv main.go.backup main.go"
	@echo "   - Restore controllers: mv controllers_backup/* controllers/"
	@echo "   - Restore models: mv models_backup/* models/"
	@echo "   - Restore migrations: mv migrations_archived/* migrations/"
	@echo ""
	@echo "3. Restart application with old system"
	@echo ""
	@echo "💡 Tip: All backups are timestamped and safe to restore"

# Clean up build artifacts
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@echo "✅ Cleanup completed"

# Development helpers
dev-setup:
	@echo "🔧 Setting up development environment..."
	@mkdir -p bin
	@mkdir -p logs
	@echo "✅ Development setup completed"

# Show current git status for backup verification
backup-check:
	@echo "📂 Current Git Status"
	@echo "===================="
	@git status --short
	@echo ""
	@echo "💡 Make sure all changes are committed before migration!"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies installed"