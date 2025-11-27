-- Add approval fields to daily_updates table
ALTER TABLE daily_updates
ADD COLUMN status VARCHAR(20) DEFAULT 'pending',
ADD COLUMN approved_by VARCHAR(100),
ADD COLUMN approved_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN rejection_reason TEXT;

-- Index for faster filtering by status
CREATE INDEX idx_daily_updates_status ON daily_updates(status);
