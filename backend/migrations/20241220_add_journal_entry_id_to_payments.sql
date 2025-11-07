-- Migration: Add journal_entry_id field to payments table
-- This migration adds support for linking payments to SSOT journal entries

-- Add journal_entry_id column to payments table
ALTER TABLE payments 
ADD COLUMN journal_entry_id bigint UNSIGNED NULL,
ADD INDEX idx_payments_journal_entry_id (journal_entry_id);

-- Add foreign key constraint if unified_journal_ledger table exists
-- Note: This constraint can be added manually after ensuring the table exists
-- ALTER TABLE payments 
-- ADD CONSTRAINT fk_payments_journal_entry 
-- FOREIGN KEY (journal_entry_id) REFERENCES unified_journal_ledger(id);