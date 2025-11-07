-- Payment Performance Optimization Migration
-- Created for optimizing payment creation and query performance

-- Add indexes for payment-related tables to improve query performance

-- Payments table indexes
CREATE INDEX IF NOT EXISTS idx_payments_contact_id ON payments(contact_id);
CREATE INDEX IF NOT EXISTS idx_payments_date ON payments(date);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(method);
CREATE INDEX IF NOT EXISTS idx_payments_code ON payments(code);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);

-- Payment allocations indexes
CREATE INDEX IF NOT EXISTS idx_payment_allocations_payment_id ON payment_allocations(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_allocations_invoice_id ON payment_allocations(invoice_id) WHERE invoice_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_allocations_bill_id ON payment_allocations(bill_id) WHERE bill_id IS NOT NULL;

-- Cash bank transactions indexes
CREATE INDEX IF NOT EXISTS idx_cashbank_transactions_cashbank_id ON cash_bank_transactions(cash_bank_id);
CREATE INDEX IF NOT EXISTS idx_cashbank_transactions_reference ON cash_bank_transactions(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_cashbank_transactions_date ON cash_bank_transactions(transaction_date);

-- Journal entries indexes for payment operations
CREATE INDEX IF NOT EXISTS idx_journal_entries_reference ON journal_entries(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_date ON journal_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_journal_entries_status ON journal_entries(status);

-- Journal lines indexes
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal_entry_id ON journal_lines(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id ON journal_lines(account_id);

-- Accounts table indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_accounts_code ON accounts(code);
CREATE INDEX IF NOT EXISTS idx_accounts_name_lower ON accounts(LOWER(name));

-- Cash bank table indexes
CREATE INDEX IF NOT EXISTS idx_cash_bank_account_id ON cash_banks(account_id);

-- Contacts table indexes for payment operations
CREATE INDEX IF NOT EXISTS idx_contacts_type ON contacts(type);

-- Purchase table indexes for payment integration
CREATE INDEX IF NOT EXISTS idx_purchases_vendor_id ON purchases(vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchases_status ON purchases(status);
CREATE INDEX IF NOT EXISTS idx_purchases_payment_method ON purchases(payment_method);

-- Sales table indexes for payment integration
CREATE INDEX IF NOT EXISTS idx_sales_customer_id ON sales(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);
CREATE INDEX IF NOT EXISTS idx_sales_outstanding_amount ON sales(outstanding_amount) WHERE outstanding_amount > 0;

-- Purchase payments table indexes (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'purchase_payments') THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_purchase_payments_purchase_id ON purchase_payments(purchase_id)';
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_purchase_payments_payment_id ON purchase_payments(payment_id) WHERE payment_id IS NOT NULL';
    END IF;
END $$;

-- Sale payments table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_sale_payments_sale_id ON sale_payments(sale_id) WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'sale_payments');
CREATE INDEX IF NOT EXISTS idx_sale_payments_payment_id ON sale_payments(payment_id) WHERE payment_id IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'sale_payments');

-- Payment code sequence table indexes
CREATE INDEX IF NOT EXISTS idx_payment_code_sequence_prefix ON payment_code_sequences(prefix, year, month);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_payments_contact_status_date ON payments(contact_id, status, date);
CREATE INDEX IF NOT EXISTS idx_payments_date_method ON payments(date, method);

-- Add partial indexes for active/pending records only
CREATE INDEX IF NOT EXISTS idx_payments_active ON payments(id) WHERE status IN ('PENDING', 'COMPLETED');
CREATE INDEX IF NOT EXISTS idx_purchases_payable ON purchases(id) WHERE status = 'APPROVED' AND payment_method = 'CREDIT' AND outstanding_amount > 0;
CREATE INDEX IF NOT EXISTS idx_sales_receivable ON sales(id) WHERE status = 'INVOICED' AND outstanding_amount > 0;

-- Optimize sequence table for better payment code generation
CREATE INDEX IF NOT EXISTS idx_payment_code_seq_unique ON payment_code_sequences(prefix, year, month) WHERE sequence_number > 0;

-- Add statistics update for PostgreSQL optimization
ANALYZE payments;
ANALYZE payment_allocations;
ANALYZE cash_bank_transactions;
ANALYZE journal_entries;
ANALYZE journal_lines;
ANALYZE accounts;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cash_banks') THEN EXECUTE 'ANALYZE cash_banks'; END IF; END $$;
ANALYZE purchases;
ANALYZE sales;

-- Add comments for documentation
COMMENT ON INDEX idx_payments_contact_id IS 'Optimizes queries filtering payments by contact';
COMMENT ON INDEX idx_payments_date IS 'Optimizes date range queries for payments';
COMMENT ON INDEX idx_payments_status IS 'Optimizes status-based filtering';
COMMENT ON INDEX idx_accounts_code IS 'Optimizes account lookups by code (e.g., 1101, 2101)';
COMMENT ON INDEX idx_purchases_payable IS 'Optimizes queries for purchases eligible for payment';
COMMENT ON INDEX idx_sales_receivable IS 'Optimizes queries for sales with outstanding amounts';

-- Performance monitoring view (optional)
CREATE OR REPLACE VIEW payment_performance_stats AS
SELECT 
    'payments' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed_count,
    COUNT(CASE WHEN status = 'PENDING' THEN 1 END) as pending_count,
    AVG(amount) as avg_amount,
    MAX(created_at) as last_payment
FROM payments
UNION ALL
SELECT 
    'journal_entries' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN status = 'POSTED' THEN 1 END) as posted_count,
    COUNT(CASE WHEN reference_type = 'PAYMENT' THEN 1 END) as payment_journals,
    AVG(total_debit) as avg_amount,
    MAX(created_at) as last_entry
FROM journal_entries;

COMMENT ON VIEW payment_performance_stats IS 'Provides performance statistics for payment-related tables';