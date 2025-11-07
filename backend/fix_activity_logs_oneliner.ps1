# One-liner fix untuk activity_logs user_id constraint
# Copy dan paste di PowerShell

Write-Host "🔧 Fixing activity_logs user_id constraint..." -ForegroundColor Yellow

$sql = @"
BEGIN;
ALTER TABLE activity_logs DROP CONSTRAINT IF EXISTS fk_activity_logs_user;
ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE activity_logs ADD CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';
COMMIT;
"@

try {
    $result = $sql | docker exec -i postgres_db psql -U postgres -d accounting_db
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Fix berhasil diterapkan!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Verifying..." -ForegroundColor Yellow
        docker exec postgres_db psql -U postgres -d accounting_db -c "\d activity_logs" | Select-String "user_id"
    } else {
        Write-Host "❌ Fix gagal!" -ForegroundColor Red
        Write-Host $result
    }
} catch {
    Write-Host "❌ Error: $_" -ForegroundColor Red
}
