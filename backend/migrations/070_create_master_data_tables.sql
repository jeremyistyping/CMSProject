-- Migration: Create Master Data Tables (COA, Materials, Vendors)
-- Version: 070
-- Description: Creates tables for Chart of Accounts, Materials, and Vendors master data

-- =====================================================
-- COA (Chart of Accounts) Table
-- =====================================================
CREATE TABLE IF NOT EXISTS coa_accounts (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(30) NOT NULL, -- ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    category VARCHAR(50), -- Material, Labor, Equipment, Overhead, Subcontractor
    budget_category VARCHAR(50), -- LABOUR_BUDGET, OPERASIONAL_BUDGET, OTHER
    work_package VARCHAR(100), -- Pekerjaan Persiapan, Pekerjaan Beton, etc
    parent_id INTEGER REFERENCES coa_accounts(id),
    level INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    is_header BOOLEAN DEFAULT FALSE,
    balance DECIMAL(20,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coa_accounts_code ON coa_accounts(code);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_type ON coa_accounts(type);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_category ON coa_accounts(category);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_budget_category ON coa_accounts(budget_category);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_work_package ON coa_accounts(work_package);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_parent_id ON coa_accounts(parent_id);
CREATE INDEX IF NOT EXISTS idx_coa_accounts_deleted_at ON coa_accounts(deleted_at);

-- =====================================================
-- Material Categories Table
-- =====================================================
CREATE TABLE IF NOT EXISTS material_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES material_categories(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_material_categories_code ON material_categories(code);
CREATE INDEX IF NOT EXISTS idx_material_categories_parent_id ON material_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_material_categories_deleted_at ON material_categories(deleted_at);


-- =====================================================
-- Unit of Measures Table
-- =====================================================
CREATE TABLE IF NOT EXISTS unit_of_measures (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10),
    category VARCHAR(30), -- Length, Area, Volume, Weight, Quantity
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_unit_of_measures_code ON unit_of_measures(code);
CREATE INDEX IF NOT EXISTS idx_unit_of_measures_category ON unit_of_measures(category);

-- =====================================================
-- Materials Table
-- =====================================================
CREATE TABLE IF NOT EXISTS materials (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES material_categories(id),
    unit VARCHAR(20) NOT NULL,
    unit_price DECIMAL(15,2) DEFAULT 0,
    min_stock DECIMAL(15,2) DEFAULT 0,
    max_stock DECIMAL(15,2) DEFAULT 0,
    current_stock DECIMAL(15,2) DEFAULT 0,
    coa_account_id INTEGER REFERENCES coa_accounts(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_materials_code ON materials(code);
CREATE INDEX IF NOT EXISTS idx_materials_category_id ON materials(category_id);
CREATE INDEX IF NOT EXISTS idx_materials_coa_account_id ON materials(coa_account_id);
CREATE INDEX IF NOT EXISTS idx_materials_is_active ON materials(is_active);
CREATE INDEX IF NOT EXISTS idx_materials_deleted_at ON materials(deleted_at);

-- =====================================================
-- Vendor Categories Table
-- =====================================================
CREATE TABLE IF NOT EXISTS vendor_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vendor_categories_code ON vendor_categories(code);
CREATE INDEX IF NOT EXISTS idx_vendor_categories_deleted_at ON vendor_categories(deleted_at);

-- =====================================================
-- Vendors Table
-- =====================================================
CREATE TABLE IF NOT EXISTS vendors (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    npwp VARCHAR(30),
    bank_name VARCHAR(100),
    bank_account VARCHAR(50),
    bank_branch VARCHAR(100),
    payment_terms INTEGER DEFAULT 30,
    category_id INTEGER REFERENCES vendor_categories(id),
    rating DECIMAL(3,2) DEFAULT 0,
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vendors_code ON vendors(code);
CREATE INDEX IF NOT EXISTS idx_vendors_category_id ON vendors(category_id);
CREATE INDEX IF NOT EXISTS idx_vendors_city ON vendors(city);
CREATE INDEX IF NOT EXISTS idx_vendors_is_active ON vendors(is_active);
CREATE INDEX IF NOT EXISTS idx_vendors_deleted_at ON vendors(deleted_at);

-- =====================================================
-- Vendor Materials (Many-to-Many relationship)
-- =====================================================
CREATE TABLE IF NOT EXISTS vendor_materials (
    id SERIAL PRIMARY KEY,
    vendor_id INTEGER NOT NULL REFERENCES vendors(id),
    material_id INTEGER NOT NULL REFERENCES materials(id),
    unit_price DECIMAL(15,2) DEFAULT 0,
    lead_time_days INTEGER DEFAULT 7,
    min_order_qty DECIMAL(15,2) DEFAULT 1,
    is_preferred BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(vendor_id, material_id)
);

CREATE INDEX IF NOT EXISTS idx_vendor_materials_vendor_id ON vendor_materials(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_materials_material_id ON vendor_materials(material_id);
CREATE INDEX IF NOT EXISTS idx_vendor_materials_is_preferred ON vendor_materials(is_preferred);

-- =====================================================
-- Add material_id to purchase_request_items
-- =====================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'purchase_request_items' AND column_name = 'material_id') THEN
        ALTER TABLE purchase_request_items ADD COLUMN material_id INTEGER REFERENCES materials(id);
        CREATE INDEX IF NOT EXISTS idx_pr_items_material_id ON purchase_request_items(material_id);
    END IF;
END $$;

-- =====================================================
-- Add coa_account_id to cbs_nodes if not exists
-- =====================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'cbs_nodes' AND column_name = 'coa_account_id') THEN
        ALTER TABLE cbs_nodes ADD COLUMN coa_account_id INTEGER REFERENCES coa_accounts(id);
        CREATE INDEX IF NOT EXISTS idx_cbs_nodes_coa_account_id ON cbs_nodes(coa_account_id);
    END IF;
END $$;
