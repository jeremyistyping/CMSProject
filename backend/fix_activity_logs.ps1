# Fix activity_logs user_id constraint
# This script makes user_id nullable to allow anonymous user logging

Write-Host "🔧 Fixing activity_logs user_id constraint..." -ForegroundColor Cyan

# Run the Go script to fix the constraint
go run cmd/scripts/fix_activity_logs_constraint.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Successfully fixed activity_logs constraint!" -ForegroundColor Green
    Write-Host "Anonymous users can now be logged properly." -ForegroundColor Green
} else {
    Write-Host "`n❌ Failed to fix constraint. Check the error above." -ForegroundColor Red
    exit 1
}
