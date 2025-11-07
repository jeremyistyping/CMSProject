-- Safe SSOT Journal Migration Fix
-- Version: 1.0
-- Date: 2024-12-25
-- Description: Safely handles existing SSOT tables and completes the journal migration

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Function to safely check if a table exists and has the expected structure
CREATE OR REPLACE FUNCTION check_table_structure(table_name TEXT) 
RETURNS BOOLEAN AS $$
DECLARE
    column_count INTEGER;
    expected_columns TEXT[];
    actual_count INTEGER;
BEGIN
    -- Initialize expected columns array
    expected_columns := ARRAY['id', 'entry_number', 'source_type', 'source_id', 'entry_date', 'description', 'total_debit', 'total_credit', 'status', 'created_by'];
    
    -- Count expected columns that exist
    SELECT COUNT(*)
    INTO actual_count
    FROM information_schema.columns 
    WHERE table_name = check_table_structure.table_name 
    AND column_name = ANY(expected_columns);
    
    -- Return true if most expected columns exist (flexible check)
    RETURN actual_count >= 8;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 1. SAFE UNIFIED JOURNAL LEDGER HANDLING
-- =====================================================

-- Check if unified_journal_ledger exists and has proper structure
DO $$
DECLARE
    table_exists BOOLEAN;
    has_proper_structure BOOLEAN;
BEGIN
    -- Check if table exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'unified_journal_ledger'
    ) INTO table_exists;
    
    IF table_exists THEN
        RAISE NOTICE 'Table unified_journal_ledger exists, checking structure...';
        
        -- Check if it has proper structure
        SELECT check_table_structure('unified_journal_ledger') INTO has_proper_structure;
        
        IF NOT has_proper_structure THEN
            RAISE NOTICE 'Table exists but has incomplete structure, adding missing columns...';
            
            -- Add missing columns safely
            ALTER TABLE unified_journal_ledger 
                ADD COLUMN IF NOT EXISTS transaction_uuid UUID DEFAULT uuid_generate_v4(),
                ADD COLUMN IF NOT EXISTS reference VARCHAR(200),
                ADD COLUMN IF NOT EXISTS notes TEXT,
                ADD COLUMN IF NOT EXISTS is_balanced BOOLEAN NOT NULL DEFAULT TRUE,
                ADD COLUMN IF NOT EXISTS is_auto_generated BOOLEAN NOT NULL DEFAULT FALSE,
                ADD COLUMN IF NOT EXISTS posted_at TIMESTAMPTZ,
                ADD COLUMN IF NOT EXISTS posted_by BIGINT,
                ADD COLUMN IF NOT EXISTS reversed_by BIGINT,
                ADD COLUMN IF NOT EXISTS reversed_from BIGINT,
                ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
                ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
                
            RAISE NOTICE '✅ Added missing columns to unified_journal_ledger';
        ELSE
            RAISE NOTICE '✅ Table unified_journal_ledger has proper structure';
        END IF;
        
        -- Ensure unique constraints exist
        BEGIN
            CREATE UNIQUE INDEX IF NOT EXISTS idx_unified_journal_ledger_entry_number 
            ON unified_journal_ledger(entry_number);
            
            CREATE UNIQUE INDEX IF NOT EXISTS idx_unified_journal_ledger_transaction_uuid 
            ON unified_journal_ledger(transaction_uuid) WHERE transaction_uuid IS NOT NULL;
            
            RAISE NOTICE '✅ Ensured unique constraints on unified_journal_ledger';
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Warning: Could not create unique constraints: %', SQLERRM;
        END;
        
    ELSE
        RAISE NOTICE 'Creating unified_journal_ledger table...';
        
        -- Create the table with full structure
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
            posted_by BIGINT,
            
            -- Reversal Information
            reversed_by BIGINT, -- Points to reversing entry ID
            reversed_from BIGINT, -- Points to original entry ID that was reversed
            reversal_reason TEXT,
            
            -- Audit Fields
            created_by BIGINT NOT NULL,
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
        
        RAISE NOTICE '✅ Created unified_journal_ledger table';
    END IF;
END $$;

-- =====================================================
-- 2. SAFE UNIFIED JOURNAL LINES HANDLING
-- =====================================================

DO $$
DECLARE
    table_exists BOOLEAN;
BEGIN
    -- Check if table exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'unified_journal_lines'
    ) INTO table_exists;
    
    IF table_exists THEN
        RAISE NOTICE 'Table unified_journal_lines exists, ensuring structure...';
        
        -- Add missing columns safely
        ALTER TABLE unified_journal_lines 
            ADD COLUMN IF NOT EXISTS quantity DECIMAL(15,4),
            ADD COLUMN IF NOT EXISTS unit_price DECIMAL(15,4),
            ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
            
        RAISE NOTICE '✅ Ensured unified_journal_lines structure';
    ELSE
        RAISE NOTICE 'Creating unified_journal_lines table...';
        
        CREATE TABLE unified_journal_lines (
            id BIGSERIAL PRIMARY KEY,
            
            -- Parent Journal Reference
            journal_id BIGINT NOT NULL REFERENCES unified_journal_ledger(id) ON DELETE CASCADE,
            
            -- Account Information
            account_id BIGINT NOT NULL,
            
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
        
        RAISE NOTICE '✅ Created unified_journal_lines table';
    END IF;
END $$;

-- =====================================================
-- 3. SAFE JOURNAL EVENT LOG HANDLING
-- =====================================================

DO $$
DECLARE
    table_exists BOOLEAN;
BEGIN
    -- Check if table exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'journal_event_log'
    ) INTO table_exists;
    
    IF table_exists THEN
        RAISE NOTICE 'Table journal_event_log exists, ensuring structure...';
        
        -- Add missing columns safely
        ALTER TABLE journal_event_log 
            ADD COLUMN IF NOT EXISTS source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
            ADD COLUMN IF NOT EXISTS correlation_id UUID,
            ADD COLUMN IF NOT EXISTS metadata JSONB;
            
        RAISE NOTICE '✅ Ensured journal_event_log structure';
    ELSE
        RAISE NOTICE 'Creating journal_event_log table...';
        
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
            user_id BIGINT,
            user_role VARCHAR(50),
            ip_address INET,
            user_agent TEXT,
            
            -- System Context
            source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
            correlation_id UUID, -- For tracing related events
            
            -- Additional metadata
            metadata JSONB
        );
        
        RAISE NOTICE '✅ Created journal_event_log table';
    END IF;
END $$;

-- =====================================================
-- 4. PERFORMANCE INDEXES (SAFE)
-- =====================================================

-- Create indexes safely (ignore if they exist)
CREATE INDEX IF NOT EXISTS idx_journal_source ON unified_journal_ledger(source_type, source_id) 
WHERE source_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_journal_date_status ON unified_journal_ledger(entry_date, status) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_journal_posted ON unified_journal_ledger(posted_at) 
WHERE status = 'POSTED' AND posted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_journal_user_date ON unified_journal_ledger(created_by, entry_date DESC);

CREATE INDEX IF NOT EXISTS idx_journal_uuid ON unified_journal_ledger(transaction_uuid);

CREATE INDEX IF NOT EXISTS idx_journal_entry_number ON unified_journal_ledger(entry_number);

-- Journal Lines Indexes
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal ON unified_journal_lines(journal_id);

CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON unified_journal_lines(account_id, journal_id);

CREATE INDEX IF NOT EXISTS idx_journal_lines_amounts ON unified_journal_lines(account_id) 
WHERE debit_amount > 0 OR credit_amount > 0;

-- Composite index for balance calculations
CREATE INDEX IF NOT EXISTS idx_journal_lines_balance_calc ON unified_journal_lines(account_id, debit_amount, credit_amount);

-- Event Log Indexes
CREATE INDEX IF NOT EXISTS idx_event_log_journal ON journal_event_log(journal_id, event_timestamp DESC) 
WHERE journal_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_event_log_user ON journal_event_log(user_id, event_timestamp DESC) 
WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_event_log_type_time ON journal_event_log(event_type, event_timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_event_log_correlation ON journal_event_log(correlation_id) 
WHERE correlation_id IS NOT NULL;

-- JSONB indexes for fast event data queries
CREATE INDEX IF NOT EXISTS idx_event_log_data_gin ON journal_event_log USING GIN (event_data);

-- =====================================================
-- 5. DATABASE FUNCTIONS & TRIGGERS (SAFE)
-- =====================================================

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

-- Drop and recreate trigger to avoid conflicts
DROP TRIGGER IF EXISTS trg_generate_entry_number ON unified_journal_ledger;
CREATE TRIGGER trg_generate_entry_number
    BEFORE INSERT ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION generate_entry_number();

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
    IF NEW.status = 'POSTED' AND (OLD IS NULL OR OLD.status != 'POSTED') THEN
        NEW.posted_at := NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop and recreate trigger to avoid conflicts
DROP TRIGGER IF EXISTS trg_validate_journal_balance ON unified_journal_ledger;
CREATE TRIGGER trg_validate_journal_balance
    BEFORE INSERT OR UPDATE ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION validate_journal_balance();

-- Function to log journal events
CREATE OR REPLACE FUNCTION log_journal_event()
RETURNS TRIGGER AS $$
DECLARE
    event_type_val VARCHAR(50);
    event_data_val JSONB;
BEGIN
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
    
    -- Insert event log
    INSERT INTO journal_event_log (
        journal_id,
        event_type,
        event_data,
        user_id,
        correlation_id
    ) VALUES (
        COALESCE(NEW.id, OLD.id),
        event_type_val,
        event_data_val,
        COALESCE(NEW.created_by, OLD.created_by),
        COALESCE(NEW.transaction_uuid, OLD.transaction_uuid)
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Drop and recreate trigger to avoid conflicts
DROP TRIGGER IF EXISTS trg_log_journal_event ON unified_journal_ledger;
CREATE TRIGGER trg_log_journal_event
    AFTER INSERT OR UPDATE OR DELETE ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION log_journal_event();

-- =====================================================
-- 6. HELPER VIEWS FOR REPORTING (SAFE)
-- =====================================================

-- Drop views if they exist and recreate them
DROP VIEW IF EXISTS v_journal_entries_detailed;
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
-- 7. CLEANUP AND FINALIZATION
-- =====================================================

-- Drop the helper function
DROP FUNCTION IF EXISTS check_table_structure(TEXT);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE '===============================================';
    RAISE NOTICE 'SSOT Journal Migration Fix Complete';
    RAISE NOTICE '===============================================';
    RAISE NOTICE 'Tables handled:';
    RAISE NOTICE '  - unified_journal_ledger';
    RAISE NOTICE '  - unified_journal_lines';
    RAISE NOTICE '  - journal_event_log';
    RAISE NOTICE '';
    RAISE NOTICE 'Views Created:';
    RAISE NOTICE '  - v_journal_entries_detailed';
    RAISE NOTICE '';
    RAISE NOTICE 'Functions & Triggers:';
    RAISE NOTICE '  - Auto entry number generation';
    RAISE NOTICE '  - Balance validation';
    RAISE NOTICE '  - Event logging';
    RAISE NOTICE '';
    RAISE NOTICE 'Status: Migration completed successfully!';
    RAISE NOTICE '===============================================';
END $$;