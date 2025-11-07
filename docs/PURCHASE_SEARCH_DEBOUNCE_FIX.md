# Purchase Search Client-Side Filtering

## Masalah
Search transaction di halaman Purchase Management melakukan reload/fetch data setiap kali user mengetik huruf satu per satu, yang menyebabkan:
- Terlalu banyak request ke server
- User experience yang buruk (loading berulang-ulang)
- Pemborosan bandwidth dan resource server
- User tidak bisa mengetik dengan bebas

## Akar Masalah
Pada implementasi sebelumnya, search input langsung memanggil `handleFilterChange()` yang akan trigger `fetchPurchases()` setiap kali ada perubahan nilai input:

```typescript
// SEBELUM (Masalah)
<Input
  value={filters.search || ''}
  onChange={(e) => handleFilterChange({ search: e.target.value })}
/>

const handleFilterChange = (newFilters: Partial<PurchaseFilterParams>) => {
  const updatedFilters = { ...filters, ...newFilters, page: 1 };
  setFilters(updatedFilters);
  fetchPurchases(updatedFilters); // ❌ Dipanggil setiap keystroke
};
```

## Solusi
Implementasi **client-side filtering** yang melakukan pencarian langsung di data yang sudah dimuat tanpa hit API sama sekali.

### Perubahan Kode

#### 1. State tambahan untuk menyimpan semua data
```typescript
const [purchases, setPurchases] = useState<Purchase[]>([]);
const [allPurchases, setAllPurchases] = useState<Purchase[]>([]); // Store all purchases for client-side filtering

// Local search state (client-side, no debouncing needed)
const [searchInput, setSearchInput] = useState('');
```

#### 2. Simpan data ke allPurchases saat fetch
```typescript
const fetchPurchases = async (filterParams: PurchaseFilterParams = filters) => {
  // ... fetch logic ...
  
  const purchaseData = Array.isArray(response?.data) ? response.data : [];
  
  // Store all purchases for client-side filtering
  setAllPurchases(purchaseData);
  setPurchases(purchaseData);
};
```

#### 3. Handler search dengan client-side filtering
```typescript
// Client-side search handler (instant, no API call)
const handleSearchChange = (value: string) => {
  setSearchInput(value);
  
  // Client-side filtering - no API call
  if (!value.trim()) {
    // If search is empty, show all purchases
    setPurchases(allPurchases);
    return;
  }
  
  // Filter purchases based on search term
  const searchTerm = value.toLowerCase();
  const filtered = allPurchases.filter(purchase => {
    // Search in purchase code
    if (purchase.code?.toLowerCase().includes(searchTerm)) return true;
    
    // Search in vendor name
    if (purchase.vendor?.name?.toLowerCase().includes(searchTerm)) return true;
    
    // Search in notes
    if (purchase.notes?.toLowerCase().includes(searchTerm)) return true;
    
    // Search in payment reference
    if (purchase.payment_reference?.toLowerCase().includes(searchTerm)) return true;
    
    return false;
  });
  
  setPurchases(filtered);
};
```

#### 4. Auto-apply search saat data berubah
```typescript
// Apply client-side search when allPurchases changes
useEffect(() => {
  if (searchInput) {
    handleSearchChange(searchInput);
  }
}, [allPurchases]);
```

#### 5. Update search input component
```typescript
// SESUDAH (Solusi)
<Input
  placeholder={t('purchases.searchPlaceholder')}
  value={searchInput}
  onChange={(e) => handleSearchChange(e.target.value)}
  bg={cardBg}
/>
```

#### 6. Update Clear Filters button
```typescript
<Button
  onClick={() => {
    // Clear search input state
    setSearchInput('');
    // Reset purchases to show all
    setPurchases(allPurchases);
    // Reset filters
    setFilters({ 
      page: 1, 
      limit: 10, 
      status: '', 
      vendor_id: '', 
      approval_status: '', 
      search: '', 
      start_date: '', 
      end_date: '' 
    });
    fetchPurchases({ page: 1, limit: 10 });
  }}
>
  {t('purchases.clearFilters')}
</Button>
```

## Hasil
### Sebelum
- ❌ Fetch data setiap keystroke (misal: "PO" = 2 request, "PO/2" = 4 request)
- ❌ Loading indicator berkedip-kedip
- ❌ Pemborosan bandwidth dan resource
- ❌ User tidak bisa mengetik bebas

### Sesudah
- ✅ **TIDAK ADA fetch/reload sama sekali** saat mengetik
- ✅ Search **instant** tanpa delay (client-side filtering)
- ✅ User bisa mengetik dengan **bebas** tanpa batasan
- ✅ Efisien bandwidth - 0 request saat search
- ✅ Smooth user experience tanpa loading
- ✅ Filter lain (vendor, status, date) tetap hit API untuk data update

## Testing
1. Buka halaman Purchase Management
2. Ketik di search box dengan bebas (misal: "PO/2025")
3. Verifikasi bahwa **TIDAK ADA loading/reload** sama sekali saat mengetik
4. Hasil search muncul **instant** tanpa delay
5. Test Clear Filters untuk memastikan search input ter-reset dan semua data muncul
6. Test filter lain (vendor, status) untuk memastikan masih hit API untuk update data
7. Coba ketik 1 huruf saja (misal: "P") - harusnya tetap tidak reload

## File yang Dimodifikasi
- `frontend/app/purchases/page.tsx`
  - State: Tambah `allPurchases` untuk menyimpan semua data
  - State: Tambah `searchInput` untuk search lokal
  - Handler: Ubah `handleSearchChange` menjadi client-side filtering
  - Effect: Auto-apply search saat data berubah
  - Component: Update Input dan Clear Filters button

## Catatan Implementasi
- **Client-side filtering** tidak hit API sama sekali untuk search
- Search bekerja pada data yang sudah dimuat (allPurchases)
- Pencarian meliputi: purchase code, vendor name, notes, payment reference
- Filter non-search (vendor, status, date) tetap hit API untuk update data
- Search **instant** tanpa delay atau batasan karakter
- Data di-refresh ulang saat user menggunakan filter lain atau refresh button

## Keuntungan Client-Side Filtering
1. ✅ **Instant search** - hasil langsung tanpa delay
2. ✅ **Bebas mengetik** - tidak ada batasan atau reload
3. ✅ **Bandwidth efficient** - 0 API call untuk search
4. ✅ **Better UX** - smooth tanpa loading indicator
5. ✅ **Simple logic** - tidak perlu debouncing atau timeout management

## Keterbatasan
1. ⚠️ Search hanya pada data yang sudah dimuat di halaman saat ini
2. ⚠️ Jika data banyak (pagination), search tidak mencakup data di page lain
3. ⚠️ Untuk search di seluruh database, user perlu gunakan filter lain dulu

## Potensi Perbaikan Lanjutan
1. **Highlight matching text**: Highlight kata yang match di hasil search
2. **Search field expansion**: Tambah field lain untuk dicari (status, amount range, dll)
3. **Advanced search mode**: Toggle antara client-side dan server-side search
4. **Search stats**: Tampilkan "Showing X of Y results"
5. **Case-sensitive option**: Toggle case-sensitive/insensitive search

## Referensi
- React useState: https://react.dev/reference/react/useState
- React useEffect: https://react.dev/reference/react/useEffect
- Array.filter(): https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter
- Client-side filtering: Best practice untuk instant search pada dataset kecil-medium

---
**Tanggal**: 2025-11-06
**Author**: AI Assistant
**Status**: ✅ Implemented & Tested
