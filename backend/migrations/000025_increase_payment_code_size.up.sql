-- Migration: Increase payment code size from VARCHAR(20) to VARCHAR(30)
-- Reason: SETOR-PPN code format (SETOR-PPN-YYMM-NNNN) can reach 19 chars
-- Need buffer for future code formats and sequences

ALTER TABLE payments ALTER COLUMN code TYPE VARCHAR(30);
