-- Remove legacy next_number columns from settings table
ALTER TABLE settings DROP COLUMN IF EXISTS invoice_next_number;
ALTER TABLE settings DROP COLUMN IF EXISTS quote_next_number;
ALTER TABLE settings DROP COLUMN IF EXISTS purchase_next_number;