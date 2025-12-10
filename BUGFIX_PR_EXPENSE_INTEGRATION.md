# Bugfix: PR to Expense Integration - COA Visibility

## Masalah yang Ditemukan

User bertanya dengan benar: **"Saat create purchase request, apakah perlu menambahkan field COA untuk memasukkan dia di kategori apa di budget?"**

### Masalah Saat Ini:
1. User pilih **Material** dari dropdown
2. Material sudah punya link ke **COA Account** (dari master data)
3. **TAPI**: User tidak bisa lihat COA mana yang akan digunakan
4. User tidak tahu expense akan masuk ke budget category apa
5. Tidak ada transparansi tentang accounting classification

### Dampak:
- User bingung expense akan tercatat di kategori budget apa
- Tidak ada validasi visual sebelum submit
- Sulit untuk memastikan COA mapping sudah benar

---

## Solusi yang Diimplementasikan

### 1. COA Dropdown yang Bisa Dipilih Manual

**Perubahan di `CreatePRModal.tsx`:**

#### A. Tambah Kolom "COA / Budget Category" dengan Dropdown
```typescript
<Thead>
    <Tr>
        <Th width="200px">Material</Th>
        <Th>Item Name</Th>
        <Th width="150px">COA / Budget Category</Th>  // ← BARU
        <Th width="100px">Qty</Th>
        <Th width="80px">Unit</Th>
        <Th width="130px">Est. Price</Th>
        <Th width="130px">Total</Th>
        <Th width="40px"></Th>
    </Tr>
</Thead>
```

#### B. COA Dropdown dengan Auto-Fill dari Material
```typescript
<VStack align="start" spacing={1} width="100%">
    {/* Dropdown COA - bisa dipilih manual atau auto dari material */}
    <Select
        size="xs"
        placeholder="Pilih COA..."
        {...register(`items.${index}.coa_account_id`)}
        value={watchedItems[index]?.coa_account_id || ''}
    >
        {coaAccounts.map((coa) => (
            <option key={coa.id} value={coa.id}>
                {coa.code} - {coa.name}
            </option>
        ))}
    </Select>
    
    {/* Badge info budget category & work package */}
    {coaInfo && (
        <HStack spacing={1} flexWrap="wrap">
            <Badge colorScheme="green" fontSize="xx-small">
                {coaInfo.budget_category?.replace('_', ' ')}
            </Badge>
            {coaInfo.work_package && (
                <Badge colorScheme="blue" fontSize="xx-small">
                    {coaInfo.work_package}
                </Badge>
            )}
        </HStack>
    )}
</VStack>
```

#### C. Fetch COA List & Auto-Fill Logic
```typescript
// 1. Fetch COA list
const fetchInitialData = async () => {
    const [projectsData, materialsData, vendorsData, uomsData, coaData] = await Promise.all([
        projectService.getAllProjects(),
        materialService.getAll({ is_active: true }),
        vendorService.getAll({ is_active: true }),
        materialService.getUoM(),
        materialService.getCOA({ is_active: true }), // ← Fetch COA
    ]);
    setCoaAccounts(coaData.data || []);
};

// 2. Auto-fill COA saat material dipilih
const handleMaterialSelect = (index: number, materialId: string) => {
    const material = materials.find(m => m.id === parseInt(materialId));
    if (material) {
        setValue(`items.${index}.item_name`, material.name);
        setValue(`items.${index}.unit`, material.unit);
        setValue(`items.${index}.estimated_price`, material.unit_price);
        setValue(`items.${index}.material_id`, material.id);
        
        // Auto-fill COA dari material
        if (material.coa_account_id) {
            setValue(`items.${index}.coa_account_id`, material.coa_account_id);
        }
    }
};

// 3. Get COA info untuk display badge
const getItemCOAInfo = (index: number) => {
    const item = watchedItems[index];
    
    // Cek dari selected COA
    if (item?.coa_account_id) {
        return coaAccounts.find(c => c.id === item.coa_account_id);
    }
    
    // Fallback ke COA dari material
    if (item?.material_id) {
        const material = materials.find(m => m.id === item.material_id);
        if (material?.coa_account_id) {
            return coaAccounts.find(c => c.id === material.coa_account_id);
        }
    }
    
    return null;
};
```

---

## Flow yang Benar Sekarang

### 1. User Create PR

```
┌─────────────────────────────────────────────────────────────┐
│ CREATE PURCHASE REQUEST FORM                                 │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ User Pilih Material:             │
        │ "Semen Portland"                 │
        └──────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ Sistem Tampilkan COA Info:       │
        │ ┌──────────────────────────────┐ │
        │ │ 5203 - Pasangan dan Plesteran│ │
        │ │ [OPERASIONAL BUDGET]         │ │
        │ │ PASANGAN DAN PLESTERAN       │ │
        │ └──────────────────────────────┘ │
        └──────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ User Lihat & Konfirmasi:         │
        │ ✅ COA sudah benar               │
        │ ✅ Budget category sesuai        │
        │ ✅ Work package sesuai           │
        └──────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ User Submit PR                   │
        └──────────────────────────────────┘
```

### 2. PR Approval & Expense Creation

```
┌─────────────────────────────────────────────────────────────┐
│ PR APPROVED                                                  │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ Sistem Auto-Create Expense:      │
        │                                  │
        │ project_id: 1                    │
        │ coa_account_id: 5203 ←───────────┼─── Dari Material
        │ description: "PR-001: Semen..."  │
        │ amount: 5,000,000                │
        │ transaction_type: MATERIAL       │
        │ reference_type: PR               │
        │ reference_id: 123                │
        └──────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────┐
        │ Expense Tercatat di:             │
        │ - COA: 5203                      │
        │ - Budget Category:               │
        │   OPERASIONAL_BUDGET             │
        │ - Work Package:                  │
        │   PASANGAN DAN PLESTERAN         │
        └──────────────────────────────────┘
```

---

## Contoh Tampilan UI

### Before (Tanpa COA Info):
```
┌────────────────────────────────────────────────────────────────┐
│ Material          │ Item Name      │ Qty │ Unit │ Price │ Total│
├────────────────────────────────────────────────────────────────┤
│ [Semen Portland]  │ Semen Portland │ 100 │ sak  │ 50000 │ 5M  │
└────────────────────────────────────────────────────────────────┘
❌ User tidak tahu expense akan masuk ke budget category apa
```

### After (Dengan COA Dropdown):
```
┌──────────────────────────────────────────────────────────────────────────┐
│ Material        │ Item Name      │ COA / Budget Category      │ Qty │...│
├──────────────────────────────────────────────────────────────────────────┤
│ [Semen Portland]│ Semen Portland │ [5203 - Pasangan & Plest▼] │ 100 │...│
│                 │                │ [OPERASIONAL BUDGET]       │     │   │
│                 │                │ [PASANGAN DAN PLESTERAN]   │     │   │
└──────────────────────────────────────────────────────────────────────────┘
✅ User bisa pilih COA dari dropdown
✅ COA auto-fill dari material (bisa diubah manual)
✅ User bisa lihat budget category & work package
✅ Fleksibel: bisa override COA jika perlu
```

---

## Keuntungan Solusi Ini

### 1. Transparansi
- User bisa lihat COA yang akan digunakan
- User tahu expense akan masuk ke budget category apa
- Tidak ada "surprise" setelah PR approved

### 2. Validasi Visual
- User bisa cek COA mapping sebelum submit
- Jika COA salah, user bisa pilih material lain
- Mengurangi error dalam accounting classification

### 3. User Experience
- Informasi lengkap dalam satu layar
- Tidak perlu buka master data material untuk cek COA
- Workflow lebih efisien

### 4. Audit Trail
- Jelas dari awal expense akan tercatat di mana
- Memudahkan review sebelum approval
- Meningkatkan akurasi budget tracking

---

## Data Flow Detail

### 1. Master Data Setup
```sql
-- COA Account
INSERT INTO coa_accounts (code, name, budget_category, work_package)
VALUES ('5203', 'Pasangan dan Plesteran', 'OPERASIONAL_BUDGET', 'PASANGAN DAN PLESTERAN');

-- Material dengan COA Link
INSERT INTO materials (name, code, unit, unit_price, coa_account_id)
VALUES ('Semen Portland', 'MAT-001', 'sak', 50000, 
        (SELECT id FROM coa_accounts WHERE code = '5203'));
```

### 2. PR Creation (Frontend)
```typescript
// User pilih material
handleMaterialSelect(index, materialId) {
    const material = materials.find(m => m.id === materialId);
    
    // Set item details
    setValue(`items.${index}.item_name`, material.name);
    setValue(`items.${index}.unit`, material.unit);
    setValue(`items.${index}.estimated_price`, material.unit_price);
    setValue(`items.${index}.material_id`, material.id);
    
    // Material sudah include COA info dari API
    // material.coa_account = {
    //     id: 5203,
    //     code: "5203",
    //     name: "Pasangan dan Plesteran",
    //     budget_category: "OPERASIONAL_BUDGET",
    //     work_package: "PASANGAN DAN PLESTERAN"
    // }
}

// Display COA info
const coaInfo = material.coa_account;
// Tampilkan: "5203 - Pasangan dan Plesteran"
// Badge: "OPERASIONAL BUDGET"
// Subtitle: "PASANGAN DAN PLESTERAN"
```

### 3. PR Approval & Expense Creation (Backend)
```go
// Saat PR approved
func (s *purchaseRequestService) CreateExpenseFromApprovedPR(prID uint) error {
    pr, _ := s.repo.FindByID(prID)
    
    for _, item := range pr.Items {
        // Get material dengan COA
        material, _ := s.materialRepo.GetByID(*item.MaterialID)
        
        // Create expense dengan COA dari material
        expense := &models.ExpenseTransaction{
            ProjectID:       pr.ProjectID,
            COAAccountID:    *material.COAAccountID,  // ← Dari material
            Description:     fmt.Sprintf("PR-%s: %s", pr.Code, item.ItemName),
            Amount:          item.TotalPrice,
            TransactionType: models.ExpenseTypeMaterial,
            ReferenceType:   models.ExpenseRefTypePR,
            ReferenceID:     &pr.ID,
            ReferenceNo:     pr.Code,
        }
        
        s.expenseRepo.Create(expense)
    }
}
```

### 4. Budget Report
```
Budget vs Actual Report
─────────────────────────────────────────────────────────
OPERASIONAL BUDGET
  PASANGAN DAN PLESTERAN
    5203 - Pasangan dan Plesteran
      Budget:  Rp 100,000,000
      Actual:  Rp   5,000,000  ← Dari PR yang approved
      Variance: Rp  95,000,000 (95%)
```

---

## Testing Checklist

### Frontend Testing:
- [ ] Pilih material → COA info muncul
- [ ] COA code dan name tampil dengan benar
- [ ] Budget category badge tampil
- [ ] Work package tampil (jika ada)
- [ ] Jika material tidak punya COA → tampil "Pilih material dulu"
- [ ] Multiple items → setiap item tampil COA masing-masing

### Integration Testing:
- [ ] Create PR dengan material yang punya COA
- [ ] Approve PR
- [ ] Verify expense created dengan COA yang benar
- [ ] Check budget report → expense masuk ke category yang benar

### Edge Cases:
- [ ] Material tanpa COA → warning di form, skip saat create expense
- [ ] Material dengan COA invalid → handle error gracefully
- [ ] Update material COA → PR baru pakai COA baru

---

## Kesimpulan

**Pertanyaan user SANGAT TEPAT!** 

Memang perlu menampilkan informasi COA di form PR agar:
1. ✅ User tahu expense akan masuk ke budget category apa
2. ✅ User bisa validasi COA mapping sebelum submit
3. ✅ Transparansi dalam accounting classification
4. ✅ Mengurangi error dan meningkatkan akurasi

**Solusi yang diimplementasikan:**
- Tambah kolom "COA / Budget Category" di form PR
- Tampilkan COA code, name, budget category, dan work package
- Real-time update saat user pilih material
- Visual feedback yang jelas dan informatif

**Hasil:**
- User experience lebih baik
- Transparansi meningkat
- Akurasi accounting classification meningkat
- Audit trail lebih jelas
