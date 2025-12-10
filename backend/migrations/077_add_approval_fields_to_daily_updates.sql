-- Migration: Add approval fields to daily_updates table
-- Purpose: Add status, approved_by, approved_at, rejection_reason for approval workflow
-- Author: System
-- Date: 2025-12-08

DO $$
BEGIN
    -- Add status column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'status') THEN
        ALTER TABLE daily_updates ADD COLUMN status VARCHAR(20) DEFAULT 'pending';
    END IF;

    -- Add approved_by column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'approved_by') THEN
        ALTER TABLE daily_updates ADD COLUMN approved_by VARCHAR(100);
    END IF;

    -- Add approved_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'approved_at') THEN
        ALTER TABLE daily_updates ADD COLUMN approved_at TIMESTAMP;
    END IF;

    -- Add rejection_reason column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'rejection_reason') THEN
        ALTER TABLE daily_updates ADD COLUMN rejection_reason TEXT;
    END IF;
END $$;

-- Add comments
COMMENT ON COLUMN daily_updates.status IS 'Approval status: pending, approved, rejected';
COMMENT ON COLUMN daily_updates.approved_by IS 'Username of the approver';
COMMENT ON COLUMN daily_updates.approved_at IS 'Timestamp when approved';
COMMENT ON COLUMN daily_updates.rejection_reason IS 'Reason for rejection if rejected';

-- Create index for status queries
CREATE INDEX IF NOT EXISTS idx_daily_updates_status ON daily_updates(status);
