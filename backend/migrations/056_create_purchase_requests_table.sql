-- Create purchase_requests table (PostgreSQL)
CREATE TABLE IF NOT EXISTS purchase_requests (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(20) NOT NULL UNIQUE,
  project_id BIGINT NOT NULL,
  request_date TIMESTAMP NOT NULL,
  required_date TIMESTAMP,
  vendor_id BIGINT,
  notes TEXT,
  status VARCHAR(20) DEFAULT 'PENDING',
  total_amount DECIMAL(15,2) DEFAULT 0.00,
  approved_by BIGINT,
  approved_at TIMESTAMP,
  rejection_reason TEXT,
  created_by BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- Create indexes for purchase_requests
CREATE INDEX IF NOT EXISTS idx_purchase_requests_project_id ON purchase_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_purchase_requests_vendor_id ON purchase_requests(vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchase_requests_approved_by ON purchase_requests(approved_by);
CREATE INDEX IF NOT EXISTS idx_purchase_requests_created_by ON purchase_requests(created_by);
CREATE INDEX IF NOT EXISTS idx_purchase_requests_deleted_at ON purchase_requests(deleted_at);

-- Add foreign key constraints for purchase_requests (ignore if already exists)
DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_requests_project') THEN
    ALTER TABLE purchase_requests ADD CONSTRAINT fk_purchase_requests_project 
      FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
  END IF;
END $$;

DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_requests_vendor') THEN
    ALTER TABLE purchase_requests ADD CONSTRAINT fk_purchase_requests_vendor 
      FOREIGN KEY (vendor_id) REFERENCES contacts(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_requests_approved_by') THEN
    ALTER TABLE purchase_requests ADD CONSTRAINT fk_purchase_requests_approved_by 
      FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_requests_created_by') THEN
    ALTER TABLE purchase_requests ADD CONSTRAINT fk_purchase_requests_created_by 
      FOREIGN KEY (created_by) REFERENCES users(id);
  END IF;
END $$;

-- Create purchase_request_items table (PostgreSQL)
CREATE TABLE IF NOT EXISTS purchase_request_items (
  id BIGSERIAL PRIMARY KEY,
  purchase_request_id BIGINT NOT NULL,
  item_name VARCHAR(255) NOT NULL,
  product_id BIGINT,
  quantity DECIMAL(10,2) NOT NULL,
  unit VARCHAR(50),
  estimated_price DECIMAL(15,2) DEFAULT 0.00,
  total_price DECIMAL(15,2) DEFAULT 0.00,
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- Create indexes for purchase_request_items
CREATE INDEX IF NOT EXISTS idx_purchase_request_items_pr_id ON purchase_request_items(purchase_request_id);
CREATE INDEX IF NOT EXISTS idx_purchase_request_items_product_id ON purchase_request_items(product_id);
CREATE INDEX IF NOT EXISTS idx_purchase_request_items_deleted_at ON purchase_request_items(deleted_at);

-- Add foreign key constraints for purchase_request_items (ignore if already exists)
DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_request_items_pr') THEN
    ALTER TABLE purchase_request_items ADD CONSTRAINT fk_purchase_request_items_pr 
      FOREIGN KEY (purchase_request_id) REFERENCES purchase_requests(id) ON DELETE CASCADE;
  END IF;
END $$;

DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_request_items_product') THEN
    ALTER TABLE purchase_request_items ADD CONSTRAINT fk_purchase_request_items_product 
      FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;
  END IF;
END $$;
