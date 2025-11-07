-- Add sample sales journal entries for P&L demonstration
-- This will create some sales transactions that will show up in Enhanced P&L report

-- First, let's add a few sales journal entries
INSERT INTO journal_entries (
    code, description, reference, reference_type, entry_date, 
    user_id, status, total_debit, total_credit, is_balanced,
    created_at, updated_at
) VALUES 
-- Sales Entry 1: Cash Sale
('JE-2025-09-18-SALES-001', 'Cash Sales - Product A', 'SALES-001', 'SALE', '2025-09-15', 
 1, 'POSTED', 11000000, 11000000, true,
 NOW(), NOW()),

-- Sales Entry 2: Credit Sale  
('JE-2025-09-18-SALES-002', 'Credit Sales - Product B', 'SALES-002', 'SALE', '2025-09-16',
 1, 'POSTED', 5500000, 5500000, true,
 NOW(), NOW()),

-- Sales Entry 3: Service Revenue
('JE-2025-09-18-SALES-003', 'Consulting Service Revenue', 'SERVICE-001', 'SALE', '2025-09-17',
 1, 'POSTED', 2750000, 2750000, true,
 NOW(), NOW()),

-- COGS Entry 1: Cost of goods sold for Product A
('JE-2025-09-18-COGS-001', 'COGS - Product A', 'COGS-001', 'MANUAL', '2025-09-15',
 1, 'POSTED', 6600000, 6600000, true,
 NOW(), NOW()),

-- COGS Entry 2: Cost of goods sold for Product B  
('JE-2025-09-18-COGS-002', 'COGS - Product B', 'COGS-002', 'MANUAL', '2025-09-16',
 1, 'POSTED', 3300000, 3300000, true,
 NOW(), NOW()),

-- Operating Expense Entry 1: Administrative Expenses
('JE-2025-09-18-OPEX-001', 'Monthly Office Rent', 'RENT-001', 'MANUAL', '2025-09-15',
 1, 'POSTED', 1500000, 1500000, true,
 NOW(), NOW()),

-- Operating Expense Entry 2: Marketing Expenses
('JE-2025-09-18-OPEX-002', 'Marketing Campaign Costs', 'MARKETING-001', 'MANUAL', '2025-09-16',
 1, 'POSTED', 800000, 800000, true,
 NOW(), NOW()),

-- Operating Expense Entry 3: Utilities
('JE-2025-09-18-OPEX-003', 'Monthly Electricity Bill', 'UTILITY-001', 'MANUAL', '2025-09-17',
 1, 'POSTED', 450000, 450000, true,
 NOW(), NOW());

-- Note: The above entries simulate the journal entry structure but don't include specific account linking.
-- In a real system, these would be created through proper sales transactions that automatically generate
-- the appropriate journal entries with correct account allocations.

-- For the Enhanced P&L to work properly, we would need to ensure these entries are linked to appropriate accounts:
-- - Revenue entries should impact account codes 4xxx (Revenue accounts)
-- - COGS entries should impact account codes 5101-5199 (Cost of Goods Sold accounts)  
-- - Operating expense entries should impact account codes 5200+ (Operating Expense accounts)

-- The Enhanced P&L service will analyze these entries and categorize them based on description patterns
-- and reference types to build the P&L statement.