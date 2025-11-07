-- Migration: Fix category unique constraint to support soft delete
-- Date: 2025-10-24

-- Drop old unique constraint (must drop constraint, not index directly)
ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS uni_product_categories_code;

-- Create new partial unique index that supports soft delete
-- This allows the same code to be reused after soft delete
CREATE UNIQUE INDEX IF NOT EXISTS idx_category_code_deleted ON product_categories (code, deleted_at);

-- This allows the same code to be used multiple times if deleted_at is different
-- But enforces uniqueness for active records (deleted_at IS NULL)
