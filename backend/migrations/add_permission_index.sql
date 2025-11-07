-- Add composite index untuk optimize permission check query
-- Ini akan mempercepat query dari 300-400ms menjadi <10ms

CREATE INDEX IF NOT EXISTS idx_module_permission_user_module 
ON module_permission_records(user_id, module);

-- Jika table menggunakan deleted_at (soft delete), buat partial index
CREATE INDEX IF NOT EXISTS idx_module_permission_user_module_active 
ON module_permission_records(user_id, module) 
WHERE deleted_at IS NULL;
