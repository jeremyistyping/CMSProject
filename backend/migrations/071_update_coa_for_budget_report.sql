-- Migration: Update COA for Budget Report Structure
-- Version: 071
-- Description: Add budget_category and work_package fields to COA

-- Add new columns if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'coa_accounts' AND column_name = 'budget_category') THEN
        ALTER TABLE coa_accounts ADD COLUMN budget_category VARCHAR(50);
        CREATE INDEX IF NOT EXISTS idx_coa_accounts_budget_category ON coa_accounts(budget_category);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'coa_accounts' AND column_name = 'work_package') THEN
        ALTER TABLE coa_accounts ADD COLUMN work_package VARCHAR(100);
        CREATE INDEX IF NOT EXISTS idx_coa_accounts_work_package ON coa_accounts(work_package);
    END IF;
END $$;

-- Insert seed data for COA structure based on construction budget report
-- Only insert if not exists

-- 5000 - BIAYA PROYEK (Header)
INSERT INTO coa_accounts (code, name, description, type, category, is_header, level, is_active)
VALUES ('5000', 'BIAYA PROYEK', 'Biaya Proyek Konstruksi', 'EXPENSE', 'Other', true, 1, true)
ON CONFLICT (code) DO NOTHING;

-- 5100 - LABOUR BUDGET (Header)
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, parent_id, is_header, level, is_active)
VALUES ('5100', 'LABOUR BUDGET', 'Budget Tenaga Kerja', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
        (SELECT id FROM coa_accounts WHERE code = '5000'), true, 2, true)
ON CONFLICT (code) DO NOTHING;

-- Labour Budget Details
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, parent_id, is_header, level, is_active)
VALUES 
    ('5101', 'Mandor Civil & MEP', 'Biaya Mandor Sipil dan MEP', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true),
    ('5102', 'Tukang Bangunan', 'Biaya Tukang Bangunan', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true),
    ('5103', 'Pekerja Harian', 'Biaya Pekerja Harian Lepas', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true),
    ('5104', 'Fee Marketing', 'Fee Marketing dan Komisi', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true),
    ('5105', 'PPH Tenaga Kerja', 'PPH untuk Tenaga Kerja', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true),
    ('5106', 'Kompensasi & Kasbon', 'Kompensasi dan Kasbon Pekerja', 'EXPENSE', 'Labor', 'LABOUR_BUDGET', 
     (SELECT id FROM coa_accounts WHERE code = '5100'), false, 3, true)
ON CONFLICT (code) DO NOTHING;

-- 5200 - OPERASIONAL BUDGET (Header)
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, parent_id, is_header, level, is_active)
VALUES ('5200', 'OPERASIONAL BUDGET', 'Budget Operasional Pekerjaan', 'EXPENSE', 'Other', 'OPERASIONAL_BUDGET', 
        (SELECT id FROM coa_accounts WHERE code = '5000'), true, 2, true)
ON CONFLICT (code) DO NOTHING;

-- Operasional Budget - Work Packages
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, work_package, parent_id, is_header, level, is_active)
VALUES 
    ('5201', 'Pekerjaan Persiapan', 'Biaya Pekerjaan Persiapan (Prelim)', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN PERSIAPAN',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5202', 'Pekerjaan Tanah dan Pasir', 'Biaya Pekerjaan Tanah dan Pasir', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN TANAH DAN PASIR',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5203', 'Pasangan dan Plesteran', 'Biaya Pasangan dan Plesteran', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PASANGAN DAN PLESTERAN',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5204', 'Pekerjaan Beton', 'Biaya Pekerjaan Beton', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN BETON',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5205', 'Pekerjaan Alumunium, Baja dan Kaca', 'Biaya Pekerjaan Alumunium, Baja dan Kaca', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN ALUMUNIUM,BAJA DAN KACA',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5206', 'Pekerjaan Atap dan Plafon', 'Biaya Pekerjaan Atap dan Plafon', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN ATAP DAN PLAFON',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5207', 'Pekerjaan Lantai', 'Biaya Pekerjaan Lantai', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN LANTAI',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5208', 'Pekerjaan Cat-catan', 'Biaya Pekerjaan Cat-catan', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN CAT CATAN',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5209', 'Pekerjaan Penggantung Pengunci', 'Biaya Pekerjaan Penggantung Pengunci', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN PENGGANTUNG PENGUNCI',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5210', 'Pekerjaan Sanitasi dan Sanitair', 'Biaya Pekerjaan Sanitasi dan Sanitair', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN SANITASI DAN SANITAIR',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5211', 'Pekerjaan Elektrikal', 'Biaya Pekerjaan Elektrikal', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN ELECTRIKAL',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5212', 'Pekerjaan Tata Udara', 'Biaya Pekerjaan Tata Udara', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN TATA UDARA',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5213', 'Pekerjaan Finishing', 'Biaya Pekerjaan Finishing', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN FINISHING',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5214', 'Pekerjaan Beton Landasan Lapangan', 'Biaya Pekerjaan Beton Landasan Lapangan Padel', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN BETON LANDASAN LAPANGAN PADEL',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5215', 'Pekerjaan Instalasi Listrik dan Drainasi', 'Biaya Pekerjaan Instalasi Listrik dan Drainasi', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN INSTALASI LISTRIK DAN DRAINASI',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5216', 'Pekerjaan Pagar', 'Biaya Pekerjaan Pagar', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN PAGAR',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5217', 'Pekerjaan Air Hujan', 'Biaya Pekerjaan Air Hujan', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN AIR HUJAN',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true),
    ('5218', 'Pekerjaan Lapangan', 'Biaya Pekerjaan Lapangan', 'EXPENSE', 'Material', 'OPERASIONAL_BUDGET', 'PEKERJAAN LAPANGAN',
     (SELECT id FROM coa_accounts WHERE code = '5200'), false, 3, true)
ON CONFLICT (code) DO NOTHING;

-- 5300 - BIAYA OPERASIONAL LAINNYA (Header)
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, parent_id, is_header, level, is_active)
VALUES ('5300', 'BIAYA OPERASIONAL LAINNYA', 'Biaya Operasional Lainnya', 'EXPENSE', 'Overhead', 'OTHER', 
        (SELECT id FROM coa_accounts WHERE code = '5000'), true, 2, true)
ON CONFLICT (code) DO NOTHING;

-- Biaya Operasional Lainnya Details
INSERT INTO coa_accounts (code, name, description, type, category, budget_category, parent_id, is_header, level, is_active)
VALUES 
    ('5301', 'Transportasi', 'Biaya Transportasi (Bensin, Solar, Tol, Tiket)', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5302', 'Akomodasi', 'Biaya Akomodasi (Kosan, Penginapan)', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5303', 'Konsumsi', 'Biaya Konsumsi (Makan, Minum)', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5304', 'Utilitas', 'Biaya Utilitas (Listrik, Air)', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5305', 'Keamanan', 'Biaya Keamanan dan Security', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5306', 'Koordinasi & Administrasi', 'Biaya Koordinasi dan Administrasi', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5307', 'Material Kecil & ATK', 'Biaya Material Kecil dan ATK', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5308', 'Peralatan & APD', 'Biaya Peralatan dan APD', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true),
    ('5309', 'Lain-lain', 'Biaya Operasional Lain-lain', 'EXPENSE', 'Overhead', 'OTHER',
     (SELECT id FROM coa_accounts WHERE code = '5300'), false, 3, true)
ON CONFLICT (code) DO NOTHING;
