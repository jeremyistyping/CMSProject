-- Create PR CBS Mappings Table
-- This table maps purchase request items to CBS nodes
-- Allows cost control to allocate PR costs to specific CBS nodes

CREATE TABLE pr_cbs_mappings (
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

-- Create indexes for faster queries
CREATE INDEX idx_pr_cbs_mappings_pr_id ON pr_cbs_mappings(purchase_request_id);
CREATE INDEX idx_pr_cbs_mappings_cbs_node_id ON pr_cbs_mappings(cbs_node_id);
CREATE INDEX idx_pr_cbs_mappings_pr_item_id ON pr_cbs_mappings(pr_item_id);

-- Add comments
COMMENT ON TABLE pr_cbs_mappings IS 'Maps purchase requests to CBS nodes for cost tracking';
COMMENT ON COLUMN pr_cbs_mappings.allocated_amount IS 'Amount allocated to this CBS node in cents/smallest currency unit';
COMMENT ON COLUMN pr_cbs_mappings.pr_item_id IS 'Optional reference to specific PR item, null if mapping entire PR';
