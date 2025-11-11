-- Migration: Add Cost Control approval fields to purchases table
-- Date: 2025-11-11
-- Purpose: Support Cost Control → GM approval flow for purchases

-- Add Cost Control approval fields
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS cost_control_approved_by INTEGER REFERENCES users(id),
ADD COLUMN IF NOT EXISTS cost_control_approved_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS cost_control_comments TEXT;

-- Add GM approval fields  
ALTER TABLE purchases
ADD COLUMN IF NOT EXISTS gm_approved_by INTEGER REFERENCES users(id),
ADD COLUMN IF NOT EXISTS gm_approved_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS gm_comments TEXT;

-- Add current approval step tracker
ALTER TABLE purchases
ADD COLUMN IF NOT EXISTS current_approval_step VARCHAR(20) DEFAULT 'NONE';

-- Update approval_status column length to accommodate new statuses
ALTER TABLE purchases 
ALTER COLUMN approval_status TYPE VARCHAR(30);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_purchases_cost_control_approved_by ON purchases(cost_control_approved_by);
CREATE INDEX IF NOT EXISTS idx_purchases_gm_approved_by ON purchases(gm_approved_by);
CREATE INDEX IF NOT EXISTS idx_purchases_current_approval_step ON purchases(current_approval_step);
CREATE INDEX IF NOT EXISTS idx_purchases_approval_status ON purchases(approval_status);

-- Add comments for documentation
COMMENT ON COLUMN purchases.cost_control_approved_by IS 'User ID who approved at Cost Control step';
COMMENT ON COLUMN purchases.cost_control_approved_at IS 'Timestamp when Cost Control approved';
COMMENT ON COLUMN purchases.cost_control_comments IS 'Comments from Cost Control during approval';
COMMENT ON COLUMN purchases.gm_approved_by IS 'User ID who approved at GM step';
COMMENT ON COLUMN purchases.gm_approved_at IS 'Timestamp when GM approved';
COMMENT ON COLUMN purchases.gm_comments IS 'Comments from GM during approval';
COMMENT ON COLUMN purchases.current_approval_step IS 'Current approval step: NONE, COST_CONTROL, GM, COMPLETED';

-- Update existing purchases to set default current_approval_step
UPDATE purchases 
SET current_approval_step = CASE
    WHEN approval_status = 'APPROVED' THEN 'COMPLETED'
    WHEN approval_status IN ('PENDING', 'PENDING_APPROVAL') THEN 'COST_CONTROL'
    ELSE 'NONE'
END
WHERE current_approval_step IS NULL OR current_approval_step = '';
