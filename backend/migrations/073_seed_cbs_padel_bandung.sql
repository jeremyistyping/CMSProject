-- ================================================================
-- CBS (Cost Breakdown Structure) Seed Data for Padel Bandung
-- Project ID: 6
-- ================================================================

-- Delete existing CBS nodes for Padel Bandung (if any)
DELETE FROM pr_cbs_mappings WHERE cbs_node_id IN (SELECT id FROM cbs_nodes WHERE project_id = 6);
DELETE FROM cbs_nodes WHERE project_id = 6;

-- ================================================================
-- Level 1: Main Categories
-- ================================================================

-- 1. SITE PREPARATION (Persiapan Lahan)
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '1000', 'Site Preparation', 'Persiapan lahan dan pembersihan area', 150000000, true, NOW(), NOW());

-- 2. FOUNDATION & STRUCTURE (Pondasi & Struktur)
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '2000', 'Foundation & Structure', 'Pekerjaan pondasi dan struktur bangunan', 800000000, true, NOW(), NOW());

-- 3. COURT CONSTRUCTION (Konstruksi Lapangan)
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '3000', 'Court Construction', 'Pembangunan lapangan padel', 1200000000, true, NOW(), NOW());

-- 4. BUILDING WORKS (Pekerjaan Bangunan)
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '4000', 'Building Works', 'Clubhouse, locker room, dan fasilitas pendukung', 600000000, true, NOW(), NOW());

-- 5. MEP (Mechanical, Electrical, Plumbing)
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '5000', 'MEP Systems', 'Sistem mekanikal, elektrikal, dan plumbing', 400000000, true, NOW(), NOW());

-- 6. FINISHING & INTERIOR
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '6000', 'Finishing & Interior', 'Pekerjaan finishing dan interior', 350000000, true, NOW(), NOW());

-- 7. LANDSCAPE & EXTERIOR
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '7000', 'Landscape & Exterior', 'Landscaping dan area eksterior', 200000000, true, NOW(), NOW());

-- 8. EQUIPMENT & FURNITURE
INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES (6, NULL, '8000', 'Equipment & Furniture', 'Peralatan dan furniture', 300000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Site Preparation (1000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '1000'), '1100', 'Land Clearing', 'Pembersihan lahan dan pembongkaran', 50000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '1000'), '1200', 'Excavation & Grading', 'Penggalian dan perataan tanah', 60000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '1000'), '1300', 'Temporary Facilities', 'Fasilitas sementara (gudang, kantor lapangan)', 40000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Foundation & Structure (2000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '2000'), '2100', 'Foundation Works', 'Pekerjaan pondasi', 300000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '2000'), '2200', 'Structural Steel', 'Struktur baja untuk atap dan dinding', 350000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '2000'), '2300', 'Concrete Works', 'Pekerjaan beton (kolom, balok, plat)', 150000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Court Construction (3000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3000'), '3100', 'Court Base & Drainage', 'Base lapangan dan sistem drainase', 200000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3000'), '3200', 'Court Surface', 'Permukaan lapangan (artificial turf)', 400000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3000'), '3300', 'Glass Walls', 'Dinding kaca tempered', 350000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3000'), '3400', 'Court Lighting', 'Sistem pencahayaan lapangan', 150000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3000'), '3500', 'Court Fencing & Nets', 'Pagar dan net lapangan', 100000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Building Works (4000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '4000'), '4100', 'Clubhouse', 'Bangunan clubhouse utama', 250000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '4000'), '4200', 'Locker Rooms', 'Ruang ganti dan shower', 150000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '4000'), '4300', 'Pro Shop', 'Toko peralatan padel', 100000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '4000'), '4400', 'Cafe/Restaurant', 'Area cafe dan restoran', 100000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for MEP Systems (5000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '5000'), '5100', 'Electrical System', 'Sistem kelistrikan', 150000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '5000'), '5200', 'HVAC System', 'Sistem AC dan ventilasi', 120000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '5000'), '5300', 'Plumbing & Sanitary', 'Sistem plumbing dan sanitasi', 80000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '5000'), '5400', 'Fire Protection', 'Sistem proteksi kebakaran', 50000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Finishing & Interior (6000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '6000'), '6100', 'Floor Finishing', 'Finishing lantai (keramik, vinyl)', 100000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '6000'), '6200', 'Wall Finishing', 'Finishing dinding (cat, wallpaper)', 80000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '6000'), '6300', 'Ceiling Works', 'Pekerjaan plafon', 70000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '6000'), '6400', 'Doors & Windows', 'Pintu dan jendela', 100000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Landscape & Exterior (7000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '7000'), '7100', 'Parking Area', 'Area parkir', 80000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '7000'), '7200', 'Garden & Landscaping', 'Taman dan landscaping', 60000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '7000'), '7300', 'Outdoor Lighting', 'Pencahayaan eksterior', 40000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '7000'), '7400', 'Signage', 'Papan nama dan signage', 20000000, true, NOW(), NOW());

-- ================================================================
-- Level 2: Sub-Categories for Equipment & Furniture (8000)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '8000'), '8100', 'Sports Equipment', 'Peralatan olahraga (raket, bola)', 80000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '8000'), '8200', 'Furniture', 'Furniture (meja, kursi, locker)', 120000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '8000'), '8300', 'Audio Visual', 'Sistem audio dan display', 60000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '8000'), '8400', 'IT Equipment', 'Komputer, POS system, WiFi', 40000000, true, NOW(), NOW());

-- ================================================================
-- Level 3: Detailed Items for Court Surface (3200)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3200'), '3210', 'Artificial Turf Material', 'Material rumput sintetis premium', 250000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3200'), '3220', 'Sand Infill', 'Pasir silika untuk infill', 80000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3200'), '3230', 'Installation Labor', 'Tenaga kerja instalasi', 70000000, true, NOW(), NOW());

-- ================================================================
-- Level 3: Detailed Items for Glass Walls (3300)
-- ================================================================

INSERT INTO cbs_nodes (project_id, parent_id, code, name, description, budget_amount, is_active, created_at, updated_at)
VALUES 
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3300'), '3310', 'Tempered Glass Panels', 'Panel kaca tempered 12mm', 280000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3300'), '3320', 'Glass Frame & Support', 'Frame dan support system', 50000000, true, NOW(), NOW()),
(6, (SELECT id FROM cbs_nodes WHERE project_id = 6 AND code = '3300'), '3330', 'Installation & Safety', 'Instalasi dan safety equipment', 20000000, true, NOW(), NOW());

-- ================================================================
-- Summary Report
-- ================================================================

-- Display CBS Tree Summary
SELECT 
    CASE 
        WHEN parent_id IS NULL THEN '📁 ' || code
        WHEN code LIKE '%10' OR code LIKE '%20' OR code LIKE '%30' OR code LIKE '%40' THEN '  📂 ' || code
        ELSE '    📄 ' || code
    END as "CBS Code",
    name as "Description",
    TO_CHAR(budget_amount, 'Rp 999,999,999,999') as "Budget",
    CASE 
        WHEN parent_id IS NULL THEN 'Level 1'
        WHEN code LIKE '%10' OR code LIKE '%20' OR code LIKE '%30' OR code LIKE '%40' THEN 'Level 2'
        ELSE 'Level 3'
    END as "Level"
FROM cbs_nodes 
WHERE project_id = 6 
ORDER BY code;

-- Display Total Budget by Level 1
SELECT 
    parent.code as "Category Code",
    parent.name as "Category",
    COUNT(child.id) as "Sub-Items",
    TO_CHAR(SUM(COALESCE(child.budget_amount, 0)) + parent.budget_amount, 'Rp 999,999,999,999') as "Total Budget"
FROM cbs_nodes parent
LEFT JOIN cbs_nodes child ON child.parent_id = parent.id
WHERE parent.project_id = 6 AND parent.parent_id IS NULL
GROUP BY parent.id, parent.code, parent.name, parent.budget_amount
ORDER BY parent.code;

-- Display Grand Total
SELECT 
    'TOTAL PROJECT BUDGET' as "Description",
    TO_CHAR(SUM(budget_amount), 'Rp 999,999,999,999') as "Amount"
FROM cbs_nodes 
WHERE project_id = 6 AND parent_id IS NULL;
