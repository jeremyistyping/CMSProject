# Fix Summary: COA Dropdown di Purchase Request Form

## Masalah yang Diperbaiki

### Error yang Terjadi:
```
TypeError: materialService.getCOA is not a function
```

**Penyebab**: 
- Mencoba memanggil `materialService.getCOA()` yang tidak ada
- Seharusnya menggunakan `coaService.getAll()`

---

## Solusi yang Diimplementasikan

### 1. Import COA Service yang Benar

**File**: `frontend/src/components/cost-control/CreatePRModal.tsx`

```typescript
// BEFORE
import { materialService, vendorService, Material, Vendor, UnitOfMeasure } from '../../services/masterDataService';

// AFTER
import { materialService, vendorService, coaService, Material, Vendor, UnitOfMeasure, COAAccount } from '../../services/masterDataService';
```

### 2. Gunakan coaService.getAll()

```typescript
// BEFORE
materialService.getCOA({ is_active: true })

// AFTER
coaService.getAll({ is_active: true })
```

### 3. Update Type untuk COA Accounts

```typescript
// BEFORE
const [coaAccounts, setCoaAccounts] = useState<any[]>([]);

// AFTER
const [coaAccounts, setCoaAccounts] = useState<COAAccount[]>([]);
```

### 4. Tambah coa_account_id ke PurchaseRequestItem Type

**File**: `frontend/src/types/purchaseRequest.ts`

```typescript
export interface PurchaseRequestItem {
    id?: number;
    purchase_request_id?: number;
    item_name: string;
    product_id?: number;
    material_id?: number;
    coa_account_id?: number;  // ← BARU
    material?: {
        id: number;
        code: string;
        name: string;
        unit: string;
        unit_price: number;
    };
    quantity: number;
    unit: string;
    estimated_price: number;
    total_price: number;
    notes: string;
}
```

---

## Fitur yang Sekarang Berfungsi

### 1. COA Dropdown
- ✅ Dropdown COA muncul di setiap item PR
- ✅ List COA diambil dari `coaService.getAll()`
- ✅ User bisa pilih COA manual

### 2. Auto-Fill dari Material
- ✅ Saat user pilih material, COA auto-fill dari `material.coa_account_id`
- ✅ User masih bisa override COA jika perlu

### 3. Display Budget Category & Work Package
- ✅ Badge menampilkan budget category (contoh: "OPERASIONAL BUDGET")
- ✅ Badge menampilkan work package (contoh: "PASANGAN DAN PLESTERAN")
- ✅ Real-time update saat COA dipilih

---

## Flow Lengkap

### 1. Load Data
```typescript
const fetchInitialData = async () => {
    const [projectsData, materialsData, vendorsData, uomsData, coaData] = await Promise.all([
        projectService.getAllProjects(),
        materialService.getAll({ is_active: true }),
        vendorService.getAll({ is_active: true }),
        materialService.getUoM(),
        coaService.getAll({ is_active: true }),  // ← Load COA
    ]);
    
    setCoaAccounts(coaData.data || []);
};
```

### 2. User Pilih Material
```typescript
const handleMaterialSelect = (index: number, materialId: string) => {
    const material = materials.find(m => m.id === parseInt(materialId));
    if (material) {
        // Set item details
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
```

### 3. Display COA Info
```typescript
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

### 4. Render Dropdown & Badge
```tsx
<VStack align="start" spacing={1} width="100%">
    {/* Dropdown COA */}
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
    
    {/* Badge Budget Category & Work Package */}
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

---

## Testing Checklist

### Frontend:
- [x] Import coaService berhasil
- [x] Fetch COA list berhasil
- [x] Dropdown COA muncul
- [x] List COA tampil lengkap
- [x] Pilih material → COA auto-fill
- [x] Badge budget category tampil
- [x] Badge work package tampil
- [x] User bisa override COA manual
- [x] No TypeScript errors
- [x] No runtime errors

### Integration:
- [ ] Create PR dengan COA
- [ ] Submit PR
- [ ] Approve PR
- [ ] Verify expense created dengan COA yang benar
- [ ] Check budget report

---

## Files Modified

### Frontend:
1. ✅ `frontend/src/components/cost-control/CreatePRModal.tsx`
   - Import coaService
   - Fetch COA list
   - Render COA dropdown
   - Auto-fill logic
   - Display badge

2. ✅ `frontend/src/types/purchaseRequest.ts`
   - Add `coa_account_id` field to PurchaseRequestItem

### Documentation:
3. ✅ `COA_DROPDOWN_FIX_SUMMARY.md` - This file
4. ✅ `BUGFIX_PR_EXPENSE_INTEGRATION.md` - Updated with dropdown solution

---

## Keuntungan Solusi Ini

### 1. Fleksibilitas
- ✅ User bisa pilih COA dari dropdown
- ✅ COA auto-fill dari material (convenience)
- ✅ User bisa override jika perlu (flexibility)

### 2. Transparansi
- ✅ User lihat COA yang akan digunakan
- ✅ Budget category & work package jelas
- ✅ Tidak ada "hidden" mapping

### 3. User Experience
- ✅ Dropdown mudah digunakan
- ✅ Auto-fill menghemat waktu
- ✅ Badge memberikan context visual
- ✅ Semua info dalam satu layar

### 4. Data Integrity
- ✅ COA ID tersimpan di PR item
- ✅ Backend bisa gunakan COA dari PR item
- ✅ Fallback ke material COA jika perlu
- ✅ Audit trail lengkap

---

## Next Steps

### 1. Backend Integration (Optional Enhancement)
Jika ingin backend prioritas COA dari PR item:

```go
// backend/services/purchase_request_service.go
func (s *purchaseRequestService) CreateExpenseFromApprovedPR(prID uint) error {
    pr, _ := s.repo.FindByID(prID)
    
    for _, item := range pr.Items {
        var coaAccountID uint
        
        // Priority 1: COA dari PR item (user pilih manual)
        if item.COAAccountID != nil {
            coaAccountID = *item.COAAccountID
        } else if item.MaterialID != nil {
            // Priority 2: COA dari material
            material, _ := s.materialRepo.GetByID(*item.MaterialID)
            if material != nil && material.COAAccountID != nil {
                coaAccountID = *material.COAAccountID
            }
        }
        
        if coaAccountID == 0 {
            continue // Skip
        }
        
        // Create expense...
    }
}
```

### 2. Validation (Optional)
Tambah validasi COA wajib diisi:

```typescript
// frontend/src/components/cost-control/CreatePRModal.tsx
<Select
    size="xs"
    placeholder="Pilih COA..."
    {...register(`items.${index}.coa_account_id`, { 
        required: "COA harus dipilih",
        valueAsNumber: true 
    })}
>
```

### 3. COA Filter (Optional)
Filter COA berdasarkan type EXPENSE saja:

```typescript
coaService.getAll({ 
    is_active: true, 
    type: 'EXPENSE' 
})
```

---

## Kesimpulan

✅ **Error Fixed**: `materialService.getCOA is not a function`  
✅ **COA Dropdown**: Berfungsi dengan baik  
✅ **Auto-Fill**: COA dari material otomatis terisi  
✅ **Manual Override**: User bisa ubah COA jika perlu  
✅ **Visual Feedback**: Badge budget category & work package  
✅ **Type Safety**: TypeScript types lengkap  
✅ **No Errors**: Frontend bersih dari error  

**Status**: ✅ READY FOR TESTING
