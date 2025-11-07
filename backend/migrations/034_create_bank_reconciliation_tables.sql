-- Migration: 034_create_bank_reconciliation_tables
-- Description: Create tables for internal bank reconciliation system
-- Author: System
-- Date: 2025-01-29

-- =====================================================
-- 1. Bank Reconciliation Snapshots
-- =====================================================
CREATE TABLE IF NOT EXISTS bank_reconciliation_snapshots (
    id SERIAL PRIMARY KEY,
    cash_bank_id INTEGER NOT NULL REFERENCES cash_banks(id) ON DELETE CASCADE,
    period VARCHAR(7) NOT NULL, -- Format: YYYY-MM
    snapshot_date TIMESTAMP NOT NULL,
    generated_by INTEGER NOT NULL REFERENCES users(id),
    
    -- Balance Information
    opening_balance DECIMAL(20,2) DEFAULT 0,
    closing_balance DECIMAL(20,2) DEFAULT 0,
    total_debit DECIMAL(20,2) DEFAULT 0,
    total_credit DECIMAL(20,2) DEFAULT 0,
    transaction_count INTEGER DEFAULT 0,
    
    -- Integrity & Security
    data_hash VARCHAR(64) NOT NULL, -- SHA-256 hash for data integrity
    is_locked BOOLEAN DEFAULT FALSE,
    locked_at TIMESTAMP,
    locked_by INTEGER REFERENCES users(id),
    
    -- Metadata
    notes TEXT,
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, ARCHIVED, SUPERSEDED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT unique_snapshot_per_period UNIQUE(cash_bank_id, period, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_snapshots_cash_bank ON bank_reconciliation_snapshots(cash_bank_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_period ON bank_reconciliation_snapshots(period);
CREATE INDEX IF NOT EXISTS idx_snapshots_date ON bank_reconciliation_snapshots(snapshot_date);
CREATE INDEX IF NOT EXISTS idx_snapshots_status ON bank_reconciliation_snapshots(status);
CREATE INDEX IF NOT EXISTS idx_snapshots_deleted ON bank_reconciliation_snapshots(deleted_at);

COMMENT ON TABLE bank_reconciliation_snapshots IS 'Frozen snapshots of bank account transactions for reconciliation purposes';
COMMENT ON COLUMN bank_reconciliation_snapshots.data_hash IS 'SHA-256 hash of transaction data for integrity verification';
COMMENT ON COLUMN bank_reconciliation_snapshots.is_locked IS 'Prevents modifications after period close';

-- =====================================================
-- 2. Reconciliation Transaction Snapshots
-- =====================================================
CREATE TABLE IF NOT EXISTS reconciliation_transaction_snapshots (
    id SERIAL PRIMARY KEY,
    snapshot_id INTEGER NOT NULL REFERENCES bank_reconciliation_snapshots(id) ON DELETE CASCADE,
    transaction_id INTEGER NOT NULL, -- Reference to original transaction
    
    -- Transaction Data (frozen at snapshot time)
    transaction_date TIMESTAMP NOT NULL,
    reference_type VARCHAR(50),
    reference_id INTEGER,
    reference_number VARCHAR(100),
    amount DECIMAL(20,2),
    debit_amount DECIMAL(20,2) DEFAULT 0,
    credit_amount DECIMAL(20,2) DEFAULT 0,
    balance_after DECIMAL(20,2),
    description TEXT,
    notes TEXT,
    
    -- Audit Info
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_snapshot_transactions_snapshot ON reconciliation_transaction_snapshots(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_transactions_original ON reconciliation_transaction_snapshots(transaction_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_transactions_date ON reconciliation_transaction_snapshots(transaction_date);
CREATE INDEX IF NOT EXISTS idx_snapshot_transactions_ref ON reconciliation_transaction_snapshots(reference_type, reference_id);

COMMENT ON TABLE reconciliation_transaction_snapshots IS 'Immutable record of transactions at snapshot time';

-- =====================================================
-- 3. Bank Reconciliations
-- =====================================================
CREATE TABLE IF NOT EXISTS bank_reconciliations (
    id SERIAL PRIMARY KEY,
    reconciliation_number VARCHAR(50) UNIQUE NOT NULL,
    cash_bank_id INTEGER NOT NULL REFERENCES cash_banks(id) ON DELETE CASCADE,
    period VARCHAR(7) NOT NULL, -- Format: YYYY-MM
    
    -- Snapshot References
    base_snapshot_id INTEGER NOT NULL REFERENCES bank_reconciliation_snapshots(id),
    comparison_snapshot_id INTEGER REFERENCES bank_reconciliation_snapshots(id),
    
    -- Reconciliation Info
    reconciliation_date TIMESTAMP NOT NULL,
    reconciliation_by INTEGER NOT NULL REFERENCES users(id),
    
    -- Balance Comparison
    base_balance DECIMAL(20,2),
    current_balance DECIMAL(20,2),
    variance DECIMAL(20,2),
    
    -- Transaction Comparison
    base_transaction_count INTEGER,
    current_transaction_count INTEGER,
    transaction_variance INTEGER,
    
    -- Differences Found
    missing_transactions INTEGER DEFAULT 0,
    added_transactions INTEGER DEFAULT 0,
    modified_transactions INTEGER DEFAULT 0,
    
    -- Status & Approval
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED, NEEDS_REVIEW
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    
    -- Result
    is_balanced BOOLEAN DEFAULT FALSE,
    balance_confirmed BOOLEAN DEFAULT FALSE,
    
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Add period column if not exists (table may exist from earlier migration)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'bank_reconciliations' AND column_name = 'period') THEN
        ALTER TABLE bank_reconciliations ADD COLUMN period VARCHAR(7);
    END IF;
END $$;

-- Create indexes after table is fully created
DO $$
BEGIN
    -- Create all indexes only if reconciliation_number column exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'bank_reconciliations' AND column_name = 'reconciliation_number') THEN
        CREATE INDEX IF NOT EXISTS idx_reconciliations_cash_bank ON bank_reconciliations(cash_bank_id);
        CREATE INDEX IF NOT EXISTS idx_reconciliations_status ON bank_reconciliations(status);
        CREATE INDEX IF NOT EXISTS idx_reconciliations_number ON bank_reconciliations(reconciliation_number);
        CREATE INDEX IF NOT EXISTS idx_reconciliations_date ON bank_reconciliations(reconciliation_date);
        CREATE INDEX IF NOT EXISTS idx_reconciliations_deleted ON bank_reconciliations(deleted_at);
        
        -- Create index on period only if column exists
        IF EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'bank_reconciliations' AND column_name = 'period') THEN
            CREATE INDEX IF NOT EXISTS idx_reconciliations_period ON bank_reconciliations(period);
        END IF;
    END IF;
END $$;

COMMENT ON TABLE bank_reconciliations IS 'Records of bank reconciliation processes comparing snapshots';

-- =====================================================
-- 4. Reconciliation Differences
-- =====================================================
CREATE TABLE IF NOT EXISTS reconciliation_differences (
    id SERIAL PRIMARY KEY,
    reconciliation_id INTEGER NOT NULL REFERENCES bank_reconciliations(id) ON DELETE CASCADE,
    
    difference_type VARCHAR(50) NOT NULL, -- MISSING, ADDED, MODIFIED, AMOUNT_CHANGE, DATE_CHANGE
    severity VARCHAR(20) DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    
    -- Transaction References
    base_transaction_id INTEGER,
    current_transaction_id INTEGER,
    
    -- Difference Details
    field VARCHAR(50), -- amount, date, description, etc
    old_value TEXT,
    new_value TEXT,
    amount_difference DECIMAL(20,2),
    
    -- Resolution
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, RESOLVED, IGNORED, ESCALATED
    resolution_notes TEXT,
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_differences_reconciliation ON reconciliation_differences(reconciliation_id);
CREATE INDEX IF NOT EXISTS idx_differences_type ON reconciliation_differences(difference_type);
CREATE INDEX IF NOT EXISTS idx_differences_severity ON reconciliation_differences(severity);
CREATE INDEX IF NOT EXISTS idx_differences_status ON reconciliation_differences(status);

COMMENT ON TABLE reconciliation_differences IS 'Detailed log of differences found during reconciliation';

-- =====================================================
-- 5. Cash Bank Audit Trail
-- =====================================================
CREATE TABLE IF NOT EXISTS cash_bank_audit_trail (
    id SERIAL PRIMARY KEY,
    cash_bank_id INTEGER NOT NULL REFERENCES cash_banks(id) ON DELETE CASCADE,
    transaction_id INTEGER,
    
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, VOID, RESTORE
    entity_type VARCHAR(50) NOT NULL, -- CASH_BANK, TRANSACTION, TRANSFER, DEPOSIT, WITHDRAWAL
    entity_id INTEGER NOT NULL,
    
    -- Change Details
    field_changed VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    
    -- Context
    reason TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Approval (for backdated or sensitive changes)
    requires_approval BOOLEAN DEFAULT FALSE,
    approved_by INTEGER REFERENCES users(id),
    approved_at TIMESTAMP,
    approval_status VARCHAR(20), -- PENDING, APPROVED, REJECTED
    
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_cash_bank ON cash_bank_audit_trail(cash_bank_id);
CREATE INDEX IF NOT EXISTS idx_audit_transaction ON cash_bank_audit_trail(transaction_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON cash_bank_audit_trail(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON cash_bank_audit_trail(action);
CREATE INDEX IF NOT EXISTS idx_audit_user ON cash_bank_audit_trail(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_date ON cash_bank_audit_trail(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_approval ON cash_bank_audit_trail(requires_approval, approval_status);

COMMENT ON TABLE cash_bank_audit_trail IS 'Complete audit trail of all cash and bank changes';
COMMENT ON COLUMN cash_bank_audit_trail.requires_approval IS 'Flags changes that need manager approval (e.g., backdated transactions)';

-- =====================================================
-- 6. Functions & Triggers
-- =====================================================

-- Function to generate reconciliation number
CREATE OR REPLACE FUNCTION generate_reconciliation_number()
RETURNS TEXT AS $$
DECLARE
    new_number TEXT;
    seq_num INTEGER;
BEGIN
    SELECT COUNT(*) + 1 INTO seq_num FROM bank_reconciliations WHERE DATE_TRUNC('month', reconciliation_date) = DATE_TRUNC('month', CURRENT_DATE);
    new_number := 'RECON-' || TO_CHAR(CURRENT_DATE, 'YYYYMM') || '-' || LPAD(seq_num::TEXT, 4, '0');
    RETURN new_number;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-generate reconciliation number
CREATE OR REPLACE FUNCTION set_reconciliation_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.reconciliation_number IS NULL OR NEW.reconciliation_number = '' THEN
        NEW.reconciliation_number := generate_reconciliation_number();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_set_reconciliation_number ON bank_reconciliations;
CREATE TRIGGER trigger_set_reconciliation_number
BEFORE INSERT ON bank_reconciliations
FOR EACH ROW
EXECUTE FUNCTION set_reconciliation_number();

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_snapshot_timestamp ON bank_reconciliation_snapshots;
CREATE TRIGGER trigger_update_snapshot_timestamp
BEFORE UPDATE ON bank_reconciliation_snapshots
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_update_reconciliation_timestamp ON bank_reconciliations;
CREATE TRIGGER trigger_update_reconciliation_timestamp
BEFORE UPDATE ON bank_reconciliations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_update_difference_timestamp ON reconciliation_differences;
CREATE TRIGGER trigger_update_difference_timestamp
BEFORE UPDATE ON reconciliation_differences
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 7. Views for Quick Access
-- =====================================================

-- View: Latest snapshots per account
CREATE OR REPLACE VIEW v_latest_snapshots AS
SELECT DISTINCT ON (cash_bank_id, period)
    s.*,
    cb.name as cash_bank_name,
    cb.type as cash_bank_type,
    cb.code as cash_bank_code,
    u.username as generated_by_username
FROM bank_reconciliation_snapshots s
JOIN cash_banks cb ON s.cash_bank_id = cb.id
JOIN users u ON s.generated_by = u.id
WHERE s.deleted_at IS NULL
ORDER BY cash_bank_id, period, snapshot_date DESC;

-- View: Reconciliation summary
-- Only create if reconciliation_by column exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'bank_reconciliations' AND column_name = 'reconciliation_by') THEN
        EXECUTE '
            CREATE OR REPLACE VIEW v_reconciliation_summary AS
            SELECT 
                r.*,
                cb.name as cash_bank_name,
                cb.code as cash_bank_code,
                u.username as reconciliation_by_username,
                ru.username as reviewed_by_username,
                (SELECT COUNT(*) FROM reconciliation_differences WHERE reconciliation_id = r.id AND status = ''PENDING'') as pending_differences
            FROM bank_reconciliations r
            JOIN cash_banks cb ON r.cash_bank_id = cb.id
            JOIN users u ON r.reconciliation_by = u.id
            LEFT JOIN users ru ON r.reviewed_by = ru.id
            WHERE r.deleted_at IS NULL
        ';
        EXECUTE 'COMMENT ON VIEW v_reconciliation_summary IS ''Summary of all reconciliations with related information''';
    END IF;
END $$;

-- View: Audit trail summary
CREATE OR REPLACE VIEW v_audit_trail_summary AS
SELECT 
    a.*,
    cb.name as cash_bank_name,
    cb.code as cash_bank_code,
    u.username as user_username,
    au.username as approved_by_username
FROM cash_bank_audit_trail a
JOIN cash_banks cb ON a.cash_bank_id = cb.id
JOIN users u ON a.user_id = u.id
LEFT JOIN users au ON a.approved_by = au.id;

-- =====================================================
-- 8. Sample Data & Comments
-- =====================================================

COMMENT ON VIEW v_latest_snapshots IS 'Shows the most recent snapshot for each account-period combination';
COMMENT ON VIEW v_audit_trail_summary IS 'Complete audit trail with user and account details';

-- Grant permissions (adjust as needed)
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO accounting_user;
-- GRANT SELECT ON ALL VIEWS IN SCHEMA public TO accounting_user;

-- Migration complete
SELECT 'Migration 034: Bank reconciliation tables created successfully' AS status;
