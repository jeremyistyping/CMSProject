-- Add verification fields to purchase_requests table
-- This enables the new PR workflow: PENDING_VERIFICATION -> VERIFIED -> APPROVED

ALTER TABLE purchase_requests
ADD COLUMN verified_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN verified_at TIMESTAMP,
ADD COLUMN verification_notes TEXT;

-- Add indexes
CREATE INDEX idx_purchase_requests_verified_by ON purchase_requests(verified_by);

-- Add comments
COMMENT ON COLUMN purchase_requests.verified_by IS 'User ID of cost control who verified this PR';
COMMENT ON COLUMN purchase_requests.verified_at IS 'Timestamp when PR was verified and mapped to CBS';
COMMENT ON COLUMN purchase_requests.verification_notes IS 'Notes from cost control during verification';
