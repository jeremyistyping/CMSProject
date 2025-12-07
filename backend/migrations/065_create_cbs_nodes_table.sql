-- Create CBS Nodes Table
-- This table stores the hierarchical Cost Breakdown Structure for each project
-- Each node can have a parent node, creating a tree structure

CREATE TABLE cbs_nodes (
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

-- Create index for faster tree queries
CREATE INDEX idx_cbs_nodes_project_id ON cbs_nodes(project_id);
CREATE INDEX idx_cbs_nodes_parent_id ON cbs_nodes(parent_id);
CREATE INDEX idx_cbs_nodes_code ON cbs_nodes(code);

-- Add comment to table
COMMENT ON TABLE cbs_nodes IS 'Cost Breakdown Structure nodes for project cost management';
COMMENT ON COLUMN cbs_nodes.code IS 'Unique cost code within project (e.g., 1.1.1, 1.1.2)';
COMMENT ON COLUMN cbs_nodes.budget_amount IS 'Budget allocated to this node in cents/smallest currency unit';
