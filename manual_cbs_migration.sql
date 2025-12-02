-- Manual CBS Migration Script
-- Run this script directly in your PostgreSQL database

-- 1. Create CBS Nodes Table
CREATE TABLE IF NOT EXISTS cbs_nodes (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES cbs_nodes(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    coa_account_id INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
    budget_amount BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(project_id, code)
);

CREATE INDEX IF NOT EXISTS idx_cbs_nodes_project_id ON cbs_nodes(project_id);
CREATE INDEX IF NOT EXISTS idx_cbs_nodes_parent_id ON cbs_nodes(parent_id);
CREATE INDEX IF NOT EXISTS idx_cbs_nodes_code ON cbs_nodes(code);

-- 2. Create PR CBS Mappings Table
CREATE TABLE IF NOT EXISTS pr_cbs_mappings (
    id SERIAL PRIMARY KEY,
    purchase_request_id INTEGER NOT NULL REFERENCES purchase_requests(id) ON DELETE CASCADE,
    cbs_node_id INTEGER NOT NULL REFERENCES cbs_nodes(id) ON DELETE RESTRICT,
    pr_item_id INTEGER REFERENCES purchase_request_items(id) ON DELETE CASCADE,
    allocated_amount BIGINT NOT NULL,
    notes TEXT,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT positive_allocated_amount CHECK (allocated_amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_pr_cbs_mappings_pr_id ON pr_cbs_mappings(purchase_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_cbs_mappings_cbs_node_id ON pr_cbs_mappings(cbs_node_id);
CREATE INDEX IF NOT EXISTS idx_pr_cbs_mappings_pr_item_id ON pr_cbs_mappings(pr_item_id);

-- 3. Add verification fields to purchase_requests table
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='purchase_requests' AND column_name='verified_by') THEN
        ALTER TABLE purchase_requests
        ADD COLUMN verified_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
        ADD COLUMN verified_at TIMESTAMP,
        ADD COLUMN verification_notes TEXT;
        
        CREATE INDEX idx_purchase_requests_verified_by ON purchase_requests(verified_by);
    END IF;
END $$;

-- Verify the tables were created
SELECT 'cbs_nodes table created' as status WHERE EXISTS (
    SELECT 1 FROM information_schema.tables WHERE table_name = 'cbs_nodes'
);

SELECT 'pr_cbs_mappings table created' as status WHERE EXISTS (
    SELECT 1 FROM information_schema.tables WHERE table_name = 'pr_cbs_mappings'
);

SELECT 'verification fields added' as status WHERE EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'purchase_requests' AND column_name = 'verified_by'
);
