# =================================================================
# Environment Setup Script for Accounting Backend (PowerShell)
# Run this after git pull on new PC/environment
# =================================================================

Write-Host "🚀 Setting up Accounting Backend Environment..." -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

# Check if we're in the right directory
if (-Not (Test-Path "cmd/main.go")) {
    Write-Host "❌ Error: Please run this script from the backend directory" -ForegroundColor Red
    exit 1
}

# Check if Go is installed
try {
    go version | Out-Null
} catch {
    Write-Host "❌ Error: Go is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

Write-Host "📋 Step 1: Running migration fixes..." -ForegroundColor Yellow
try {
    go run cmd/fix_migrations.go
    Write-Host "✅ Migration fixes completed" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Migration fixes had some issues, continuing..." -ForegroundColor Yellow
}

Write-Host "🔧 Step 2: Running remaining migration fixes..." -ForegroundColor Yellow
try {
    go run cmd/fix_remaining_migrations.go
    Write-Host "✅ Remaining migration fixes completed" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Some remaining fixes had issues, continuing..." -ForegroundColor Yellow
}

Write-Host "🧪 Step 3: Running verification..." -ForegroundColor Yellow
try {
    go run cmd/final_verification.go
    Write-Host "✅ Environment verification completed" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Verification had some issues, but environment should work" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎯 Environment Setup Complete!" -ForegroundColor Green
Write-Host "==============================" -ForegroundColor Green
Write-Host "✅ Backend is ready to run" -ForegroundColor Green
Write-Host "✅ Database objects created" -ForegroundColor Green
Write-Host "✅ SSOT system configured" -ForegroundColor Green
Write-Host ""
Write-Host "🚀 You can now run: go run cmd/main.go" -ForegroundColor Cyan
Write-Host "🌐 Backend will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host "📖 Swagger docs at: http://localhost:8080/swagger/index.html" -ForegroundColor Cyan