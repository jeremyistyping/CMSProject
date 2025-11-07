-- Migration 026: Purchase Balance Minimal (Account Creation Only)
-- Purpose: Create accounts payable account and mark migration as complete
-- Date: 2025-09-27
-- Compatibility: Ultra minimal - no functions, just account setup

-- Create Accounts Payable account if not exists
INSERT INTO accounts (code, name, type, balance, created_at, updated_at)
VALUES ('2101', 'Hutang Usaha', 'LIABILITY', 0.00, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Log migration completion
INSERT INTO migration_logs (migration_name, executed_at, description, status)
VALUES (
    '026_purchase_balance_minimal',
    NOW(),
    'Purchase Balance Account created (minimal version for Go compatibility)',
    'COMPLETED'
)
ON CONFLICT (migration_name) DO UPDATE SET 
    executed_at = NOW(),
    status = 'COMPLETED',
    description = EXCLUDED.description;