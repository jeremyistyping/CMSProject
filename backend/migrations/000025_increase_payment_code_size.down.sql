-- Rollback: Revert payment code size from VARCHAR(30) back to VARCHAR(20)
-- Warning: This will fail if any existing codes exceed 20 characters

ALTER TABLE payments ALTER COLUMN code TYPE VARCHAR(20);
