-- Migration: Create Unified Journal SSOT (Single Source of Truth)
-- Version: 1.0
-- Date: 2024-01-18
-- Description: Creates the unified journal system to replace fragmented journal tables

BEGIN;

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- =====================================================
-- 1. UNIFIED JOURNAL LEDGER (Main Table)
-- =====================================================

CREATE TABLE unified_journal_ledger (
    id BIGSERIAL PRIMARY KEY,
    
    -- Transaction Identity
    transaction_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    entry_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Source Transaction Reference
    source_type VARCHAR(30) NOT NULL CHECK (source_type IN (
        'SALE', 'PURCHASE', 'PAYMENT', 'CASH_BANK', 'ASSET', 'MANUAL', 
        'OPENING', 'CLOSING', 'ADJUSTMENT', 'TRANSFER', 'DEPRECIATION'
    )),
    source_id BIGINT, -- Reference ke tabel sumber (nullable untuk manual entries)
    source_code VARCHAR(100), -- Code dari transaksi sumber
    
    -- Journal Entry Details
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(200),
    notes TEXT,
    
    -- Financial Amounts (Always Balanced)
    total_debit DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (total_debit >= 0),
    total_credit DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (total_credit >= 0),
    
    -- Status & Control Fields
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN (
        'DRAFT', 'POSTED', 'REVERSED', 'CANCELLED'
    )),
    is_balanced BOOLEAN NOT NULL DEFAULT TRUE,
    is_auto_generated BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Posting Information
    posted_at TIMESTAMPTZ,
    posted_by BIGINT REFERENCES users(id),
    
    -- Reversal Information
    reversed_by BIGINT, -- Points to reversing entry ID
    reversed_from BIGINT, -- Points to original entry ID that was reversed
    reversal_reason TEXT,
    
    -- Audit Fields
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Financial Constraint: Must be balanced
    CONSTRAINT chk_balanced CHECK (total_debit = total_credit),
    
    -- Reversal Constraints
    CONSTRAINT chk_reversal_logic CHECK (
        (reversed_by IS NULL AND reversed_from IS NULL) OR
        (reversed_by IS NOT NULL AND reversed_from IS NULL) OR
        (reversed_by IS NULL AND reversed_from IS NOT NULL)
    )
);

-- Add comments for documentation
COMMENT ON TABLE unified_journal_ledger IS 'Single source of truth for all journal transactions';
COMMENT ON COLUMN unified_journal_ledger.transaction_uuid IS 'Unique identifier for transaction tracing';
COMMENT ON COLUMN unified_journal_ledger.source_type IS 'Type of source transaction that generated this journal';
COMMENT ON COLUMN unified_journal_ledger.is_balanced IS 'Auto-calculated: true when total_debit equals total_credit';
COMMENT ON COLUMN unified_journal_ledger.reversed_by IS 'ID of the journal entry that reverses this entry';
COMMENT ON COLUMN unified_journal_ledger.reversed_from IS 'ID of the original entry that this entry reverses';

-- =====================================================
-- 2. UNIFIED JOURNAL LINES (Detail Table)
-- =====================================================

CREATE TABLE unified_journal_lines (
    id BIGSERIAL PRIMARY KEY,
    
    -- Parent Journal Reference
    journal_id BIGINT NOT NULL REFERENCES unified_journal_ledger(id) ON DELETE CASCADE,
    
    -- Account Information
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    
    -- Line Details
    line_number SMALLINT NOT NULL CHECK (line_number > 0),
    description TEXT,
    
    -- Financial Amounts (Mutually Exclusive)
    debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
    credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    
    -- Additional Information for Inventory/Asset tracking
    quantity DECIMAL(15,4),
    unit_price DECIMAL(15,4),
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Business Constraints
    CONSTRAINT chk_amounts_not_both CHECK (
        NOT (debit_amount > 0 AND credit_amount > 0)
    ),
    CONSTRAINT chk_amounts_not_zero CHECK (
        debit_amount > 0 OR credit_amount > 0
    ),
    
    -- Unique line numbering per journal
    UNIQUE(journal_id, line_number)
);

-- Add comments for documentation
COMMENT ON TABLE unified_journal_lines IS 'Detail lines for each journal entry with debit/credit amounts';
COMMENT ON COLUMN unified_journal_lines.line_number IS 'Sequential line number within each journal entry';
COMMENT ON COLUMN unified_journal_lines.quantity IS 'Optional quantity for inventory-related transactions';
COMMENT ON COLUMN unified_journal_lines.unit_price IS 'Optional unit price for inventory-related transactions';

-- =====================================================
-- 3. JOURNAL EVENT LOG (Audit Trail)
-- =====================================================

CREATE TABLE journal_event_log (
    id BIGSERIAL PRIMARY KEY,
    
    -- Event Identity
    event_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    
    -- Related Journal (nullable for system events)
    journal_id BIGINT REFERENCES unified_journal_ledger(id),
    
    -- Event Details
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'CREATED', 'POSTED', 'REVERSED', 'UPDATED', 'DELETED', 
        'BALANCED', 'VALIDATED', 'MIGRATED', 'SYSTEM_ACTION'
    )),
    event_data JSONB NOT NULL,
    event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- User Context
    user_id BIGINT REFERENCES users(id),
    user_role VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    
    -- System Context
    source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
    correlation_id UUID, -- For tracing related events
    
    -- Additional metadata
    metadata JSONB
);

-- Add comments for documentation
COMMENT ON TABLE journal_event_log IS 'Complete audit trail for all journal-related events';
COMMENT ON COLUMN journal_event_log.event_data IS 'Full JSON payload of the event for complete auditability';
COMMENT ON COLUMN journal_event_log.correlation_id IS 'UUID to group related events together';

-- =====================================================
-- 4. ACCOUNT BALANCES MATERIALIZED VIEW
-- =====================================================

CREATE MATERIALIZED VIEW account_balances AS
WITH journal_totals AS (
    SELECT 
        jl.account_id,
        SUM(jl.debit_amount) as total_debits,
        SUM(jl.credit_amount) as total_credits,
        COUNT(*) as transaction_count,
        MAX(jd.posted_at) as last_transaction_date
    FROM unified_journal_lines jl
    JOIN unified_journal_ledger jd ON jl.journal_id = jd.id
    WHERE jd.status = 'POSTED' 
      AND jd.deleted_at IS NULL
    GROUP BY jl.account_id
)
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    a.category as account_category,
    
    -- Get normal balance from account type
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
        WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
        ELSE 'DEBIT'
    END as normal_balance,
    
    -- Journal totals
    COALESCE(jt.total_debits, 0) as total_debits,
    COALESCE(jt.total_credits, 0) as total_credits,
    COALESCE(jt.transaction_count, 0) as transaction_count,
    jt.last_transaction_date,
    
    -- Calculate current balance based on normal balance
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
            COALESCE(jt.total_debits, 0) - COALESCE(jt.total_credits, 0)
        WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
            COALESCE(jt.total_credits, 0) - COALESCE(jt.total_debits, 0)
        ELSE 0
    END as current_balance,
    
    -- Metadata
    NOW() as last_updated,
    a.is_active,
    a.is_header
FROM accounts a
LEFT JOIN journal_totals jt ON a.id = jt.account_id
WHERE a.deleted_at IS NULL;

-- Add comments for materialized view
COMMENT ON MATERIALIZED VIEW account_balances IS 'Real-time account balances calculated from posted journal entries';

-- =====================================================
-- 5. PERFORMANCE INDEXES
-- =====================================================

-- Journal Ledger Indexes
CREATE INDEX CONCURRENTLY idx_journal_source ON unified_journal_ledger(source_type, source_id) 
WHERE source_id IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_journal_date_status ON unified_journal_ledger(entry_date, status) 
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_journal_posted ON unified_journal_ledger(posted_at) 
WHERE status = 'POSTED' AND posted_at IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_journal_user_date ON unified_journal_ledger(created_by, entry_date DESC);

CREATE INDEX CONCURRENTLY idx_journal_uuid ON unified_journal_ledger(transaction_uuid);

CREATE INDEX CONCURRENTLY idx_journal_entry_number ON unified_journal_ledger(entry_number);

-- Journal Lines Indexes
CREATE INDEX CONCURRENTLY idx_journal_lines_journal ON unified_journal_lines(journal_id);

CREATE INDEX CONCURRENTLY idx_journal_lines_account ON unified_journal_lines(account_id, journal_id);

CREATE INDEX CONCURRENTLY idx_journal_lines_amounts ON unified_journal_lines(account_id) 
WHERE debit_amount > 0 OR credit_amount > 0;

-- Composite index for balance calculations
CREATE INDEX CONCURRENTLY idx_journal_lines_balance_calc ON unified_journal_lines(account_id, debit_amount, credit_amount);

-- Event Log Indexes
CREATE INDEX CONCURRENTLY idx_event_log_journal ON journal_event_log(journal_id, event_timestamp DESC) 
WHERE journal_id IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_event_log_user ON journal_event_log(user_id, event_timestamp DESC) 
WHERE user_id IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_event_log_type_time ON journal_event_log(event_type, event_timestamp DESC);

CREATE INDEX CONCURRENTLY idx_event_log_correlation ON journal_event_log(correlation_id) 
WHERE correlation_id IS NOT NULL;

-- JSONB indexes for fast event data queries
CREATE INDEX CONCURRENTLY idx_event_log_data_gin ON journal_event_log USING GIN (event_data);

-- Materialized View Index
CREATE UNIQUE INDEX idx_account_balances_account_id ON account_balances(account_id);

-- =====================================================
-- 6. DATABASE FUNCTIONS & TRIGGERS
-- =====================================================

-- Function to refresh account balances materialized view
-- NOTE: This trigger is DISABLED by default to prevent concurrent refresh conflicts
-- Use manual refresh or scheduled jobs instead
-- Balance sync is handled by setup_automatic_balance_sync.sql triggers
CREATE OR REPLACE FUNCTION refresh_account_balances()
RETURNS TRIGGER AS $$
BEGIN
    -- DISABLED: Automatic refresh causes SQLSTATE 55000 concurrent refresh errors
    -- Balance sync is already handled by trigger_sync_account_balance() in setup_automatic_balance_sync.sql
    -- This function is kept for manual refresh calls only
    RAISE NOTICE 'Automatic materialized view refresh is disabled. Use manual refresh or scheduled jobs.';
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain balance consistency
-- DISABLED: This trigger causes concurrent refresh conflicts in production
-- Balance sync is handled by setup_automatic_balance_sync.sql
-- CREATE TRIGGER trg_refresh_account_balances
--     AFTER INSERT OR UPDATE OR DELETE ON unified_journal_lines
--     FOR EACH STATEMENT
--     EXECUTE FUNCTION refresh_account_balances();

-- Function to validate journal entry balance
CREATE OR REPLACE FUNCTION validate_journal_balance()
RETURNS TRIGGER AS $$
DECLARE
    calculated_debit DECIMAL(20,2);
    calculated_credit DECIMAL(20,2);
    line_count INTEGER;
BEGIN
    -- Calculate totals from lines for this journal
    SELECT 
        COALESCE(SUM(debit_amount), 0),
        COALESCE(SUM(credit_amount), 0),
        COUNT(*)
    INTO calculated_debit, calculated_credit, line_count
    FROM unified_journal_lines
    WHERE journal_id = NEW.id;
    
    -- Update calculated totals
    NEW.total_debit := calculated_debit;
    NEW.total_credit := calculated_credit;
    NEW.is_balanced := (calculated_debit = calculated_credit AND calculated_debit > 0);
    
    -- Prevent posting unbalanced entries
    IF NEW.status = 'POSTED' AND NOT NEW.is_balanced THEN
        RAISE EXCEPTION 'Cannot post unbalanced journal entry %. Debit: %, Credit: %', 
                       NEW.entry_number, calculated_debit, calculated_credit;
    END IF;
    
    -- Ensure minimum lines for posting
    IF NEW.status = 'POSTED' AND line_count < 2 THEN
        RAISE EXCEPTION 'Cannot post journal entry % with less than 2 lines', NEW.entry_number;
    END IF;
    
    -- Auto-update posting timestamp
    IF NEW.status = 'POSTED' AND OLD.status != 'POSTED' THEN
        NEW.posted_at := NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for journal validation
CREATE TRIGGER trg_validate_journal_balance
    BEFORE INSERT OR UPDATE ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION validate_journal_balance();

-- Function to generate entry number
CREATE OR REPLACE FUNCTION generate_entry_number()
RETURNS TRIGGER AS $$
DECLARE
    entry_prefix VARCHAR(10);
    date_part VARCHAR(10);
    sequence_num INTEGER;
    new_entry_number VARCHAR(50);
BEGIN
    -- Only generate if not provided
    IF NEW.entry_number IS NULL OR NEW.entry_number = '' THEN
        -- Determine prefix based on source type
        entry_prefix := CASE NEW.source_type
            WHEN 'MANUAL' THEN 'JE'
            WHEN 'SALE' THEN 'SJ'
            WHEN 'PURCHASE' THEN 'PJ'
            WHEN 'PAYMENT' THEN 'PY'
            WHEN 'CASH_BANK' THEN 'CB'
            WHEN 'ASSET' THEN 'AJ'
            ELSE 'JE'
        END;
        
        -- Format: JE-YYYY-MM-XXXX
        date_part := TO_CHAR(NEW.entry_date, 'YYYY-MM');
        
        -- Get next sequence number for this month
        SELECT COALESCE(MAX(
            CAST(SUBSTRING(entry_number FROM '[0-9]+$') AS INTEGER)
        ), 0) + 1
        INTO sequence_num
        FROM unified_journal_ledger
        WHERE entry_number LIKE entry_prefix || '-' || date_part || '-%'
          AND EXTRACT(YEAR FROM entry_date) = EXTRACT(YEAR FROM NEW.entry_date)
          AND EXTRACT(MONTH FROM entry_date) = EXTRACT(MONTH FROM NEW.entry_date);
        
        -- Generate final entry number
        new_entry_number := entry_prefix || '-' || date_part || '-' || LPAD(sequence_num::TEXT, 4, '0');
        NEW.entry_number := new_entry_number;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for entry number generation
CREATE TRIGGER trg_generate_entry_number
    BEFORE INSERT ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION generate_entry_number();

-- Function to log journal events
CREATE OR REPLACE FUNCTION log_journal_event()
RETURNS TRIGGER AS $$
DECLARE
    event_type_val VARCHAR(50);
    event_data_val JSONB;
    new_event_uuid UUID;
BEGIN
    -- Generate new UUID for this event
    new_event_uuid := uuid_generate_v4();
    
    -- Determine event type
    IF TG_OP = 'INSERT' THEN
        event_type_val := 'CREATED';
        event_data_val := jsonb_build_object(
            'operation', 'INSERT',
            'journal_id', NEW.id,
            'entry_number', NEW.entry_number,
            'source_type', NEW.source_type,
            'total_debit', NEW.total_debit,
            'total_credit', NEW.total_credit,
            'status', NEW.status
        );
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status != NEW.status AND NEW.status = 'POSTED' THEN
            event_type_val := 'POSTED';
        ELSIF OLD.status != NEW.status AND NEW.status = 'REVERSED' THEN
            event_type_val := 'REVERSED';
        ELSE
            event_type_val := 'UPDATED';
        END IF;
        
        event_data_val := jsonb_build_object(
            'operation', 'UPDATE',
            'journal_id', NEW.id,
            'entry_number', NEW.entry_number,
            'changes', jsonb_build_object(
                'status', jsonb_build_object('from', OLD.status, 'to', NEW.status),
                'total_debit', jsonb_build_object('from', OLD.total_debit, 'to', NEW.total_debit),
                'total_credit', jsonb_build_object('from', OLD.total_credit, 'to', NEW.total_credit)
            )
        );
    ELSIF TG_OP = 'DELETE' THEN
        event_type_val := 'DELETED';
        event_data_val := jsonb_build_object(
            'operation', 'DELETE',
            'journal_id', OLD.id,
            'entry_number', OLD.entry_number,
            'deleted_at', OLD.deleted_at
        );
    END IF;
    
    -- Insert event log with explicit event_uuid
    INSERT INTO journal_event_log (
        event_uuid,
        journal_id,
        event_type,
        event_data,
        user_id,
        correlation_id
    ) VALUES (
        new_event_uuid,
        COALESCE(NEW.id, OLD.id),
        event_type_val,
        event_data_val,
        COALESCE(NEW.created_by, OLD.created_by),
        COALESCE(NEW.transaction_uuid, OLD.transaction_uuid)
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger for event logging
CREATE TRIGGER trg_log_journal_event
    AFTER INSERT OR UPDATE OR DELETE ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION log_journal_event();

-- =====================================================
-- 7. HELPER VIEWS FOR REPORTING
-- =====================================================

-- View for balance sheet data
CREATE VIEW v_balance_sheet_data AS
SELECT 
    ab.account_code,
    ab.account_name,
    ab.account_type,
    ab.account_category,
    ab.current_balance,
    ab.normal_balance,
    -- Format balance according to normal balance
    CASE 
        WHEN ab.normal_balance = 'DEBIT' AND ab.current_balance < 0 THEN ABS(ab.current_balance)
        WHEN ab.normal_balance = 'CREDIT' AND ab.current_balance < 0 THEN ABS(ab.current_balance)
        ELSE ab.current_balance
    END as display_balance,
    ab.transaction_count,
    ab.last_transaction_date
FROM account_balances ab
WHERE ab.is_active = true 
  AND ab.account_type IN ('ASSET', 'LIABILITY', 'EQUITY')
  AND (ab.current_balance != 0 OR ab.transaction_count > 0)
ORDER BY ab.account_code;

-- View for income statement data
CREATE VIEW v_income_statement_data AS
SELECT 
    ab.account_code,
    ab.account_name,
    ab.account_type,
    ab.account_category,
    ab.current_balance,
    ab.normal_balance,
    -- Format balance for P&L (Revenue and Expenses)
    CASE 
        WHEN ab.account_type = 'REVENUE' THEN ab.current_balance
        WHEN ab.account_type = 'EXPENSE' THEN ab.current_balance
        ELSE 0
    END as display_balance,
    ab.transaction_count,
    ab.last_transaction_date
FROM account_balances ab
WHERE ab.is_active = true 
  AND ab.account_type IN ('REVENUE', 'EXPENSE')
  AND (ab.current_balance != 0 OR ab.transaction_count > 0)
ORDER BY ab.account_code;

-- View for journal entries with line details
CREATE VIEW v_journal_entries_detailed AS
SELECT 
    jd.id as journal_id,
    jd.transaction_uuid,
    jd.entry_number,
    jd.source_type,
    jd.source_id,
    jd.source_code,
    jd.entry_date,
    jd.description as journal_description,
    jd.reference,
    jd.notes,
    jd.total_debit,
    jd.total_credit,
    jd.status,
    jd.is_balanced,
    jd.is_auto_generated,
    jd.posted_at,
    jd.posted_by,
    jd.created_by,
    jd.created_at,
    
    -- Line details
    jl.id as line_id,
    jl.account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    jl.line_number,
    jl.description as line_description,
    jl.debit_amount,
    jl.credit_amount,
    jl.quantity,
    jl.unit_price
FROM unified_journal_ledger jd
LEFT JOIN unified_journal_lines jl ON jd.id = jl.journal_id
LEFT JOIN accounts a ON jl.account_id = a.id
WHERE jd.deleted_at IS NULL
ORDER BY jd.entry_date DESC, jd.entry_number, jl.line_number;

-- =====================================================
-- 8. MONITORING & HEALTH CHECK VIEWS
-- =====================================================

-- Health check for balance consistency
CREATE VIEW v_balance_health_check AS
SELECT 
    COUNT(*) FILTER (WHERE current_balance != 0) as accounts_with_balance,
    COUNT(*) FILTER (WHERE transaction_count = 0 AND current_balance != 0) as orphaned_balances,
    COUNT(*) FILTER (WHERE transaction_count > 0 AND current_balance = 0) as zero_balance_with_transactions,
    COUNT(*) FILTER (WHERE NOT is_active AND current_balance != 0) as inactive_with_balance,
    NOW() as check_timestamp
FROM account_balances;

-- Performance monitoring for journal operations
CREATE VIEW v_journal_performance AS
SELECT 
    DATE_TRUNC('hour', created_at) as hour,
    source_type,
    COUNT(*) as entries_count,
    AVG(total_debit) as avg_amount,
    COUNT(*) FILTER (WHERE status = 'POSTED') as posted_count,
    COUNT(*) FILTER (WHERE status = 'DRAFT') as draft_count,
    AVG(EXTRACT(EPOCH FROM (COALESCE(posted_at, NOW()) - created_at))) as avg_processing_time_seconds
FROM unified_journal_ledger 
WHERE created_at >= NOW() - INTERVAL '7 days'
  AND deleted_at IS NULL
GROUP BY DATE_TRUNC('hour', created_at), source_type
ORDER BY hour DESC;

-- =====================================================
-- 9. SECURITY & PERMISSIONS
-- =====================================================

-- Create role for journal read-only access
CREATE ROLE journal_reader;
GRANT SELECT ON unified_journal_ledger TO journal_reader;
GRANT SELECT ON unified_journal_lines TO journal_reader;
GRANT SELECT ON account_balances TO journal_reader;
GRANT SELECT ON v_balance_sheet_data TO journal_reader;
GRANT SELECT ON v_income_statement_data TO journal_reader;
GRANT SELECT ON v_journal_entries_detailed TO journal_reader;

-- Create role for journal write access
CREATE ROLE journal_writer;
GRANT journal_reader TO journal_writer;
GRANT INSERT, UPDATE, DELETE ON unified_journal_ledger TO journal_writer;
GRANT INSERT, UPDATE, DELETE ON unified_journal_lines TO journal_writer;
GRANT INSERT ON journal_event_log TO journal_writer;
GRANT USAGE ON SEQUENCE unified_journal_ledger_id_seq TO journal_writer;
GRANT USAGE ON SEQUENCE unified_journal_lines_id_seq TO journal_writer;
GRANT USAGE ON SEQUENCE journal_event_log_id_seq TO journal_writer;

-- Create role for journal admin (can refresh materialized views)
CREATE ROLE journal_admin;
GRANT journal_writer TO journal_admin;
GRANT SELECT ON journal_event_log TO journal_admin;

-- =====================================================
-- 10. FINAL STEPS
-- =====================================================

-- Initial refresh of materialized view
REFRESH MATERIALIZED VIEW account_balances;

-- Add table constraints validation
ALTER TABLE unified_journal_ledger VALIDATE CONSTRAINT chk_balanced;
ALTER TABLE unified_journal_lines VALIDATE CONSTRAINT chk_amounts_not_both;
ALTER TABLE unified_journal_lines VALIDATE CONSTRAINT chk_amounts_not_zero;

-- Create indexes for foreign key constraints to improve performance
CREATE INDEX CONCURRENTLY idx_journal_ledger_created_by ON unified_journal_ledger(created_by);
CREATE INDEX CONCURRENTLY idx_journal_ledger_posted_by ON unified_journal_ledger(posted_by) WHERE posted_by IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_journal_lines_account_id ON unified_journal_lines(account_id);
CREATE INDEX CONCURRENTLY idx_event_log_user_id ON journal_event_log(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_event_log_journal_id ON journal_event_log(journal_id) WHERE journal_id IS NOT NULL;

COMMIT;

-- =====================================================
-- MIGRATION COMPLETE
-- =====================================================

-- Display summary information
DO $$
BEGIN
    RAISE NOTICE '===============================================';
    RAISE NOTICE 'Unified Journal SSOT Migration Complete';
    RAISE NOTICE '===============================================';
    RAISE NOTICE 'Tables Created:';
    RAISE NOTICE '  - unified_journal_ledger (main journal table)';
    RAISE NOTICE '  - unified_journal_lines (journal line details)';
    RAISE NOTICE '  - journal_event_log (audit trail)';
    RAISE NOTICE '';
    RAISE NOTICE 'Views Created:';
    RAISE NOTICE '  - account_balances (materialized view)';
    RAISE NOTICE '  - v_balance_sheet_data';
    RAISE NOTICE '  - v_income_statement_data';
    RAISE NOTICE '  - v_journal_entries_detailed';
    RAISE NOTICE '  - v_balance_health_check';
    RAISE NOTICE '  - v_journal_performance';
    RAISE NOTICE '';
    RAISE NOTICE 'Functions & Triggers:';
    RAISE NOTICE '  - Auto balance validation';
    RAISE NOTICE '  - Entry number generation';
    RAISE NOTICE '  - Event logging';
    RAISE NOTICE '  - Balance refresh triggers';
    RAISE NOTICE '';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '  1. Test schema in development environment';
    RAISE NOTICE '  2. Create data migration scripts';
    RAISE NOTICE '  3. Update application services';
    RAISE NOTICE '  4. Performance test with sample data';
    RAISE NOTICE '===============================================';
END $$;