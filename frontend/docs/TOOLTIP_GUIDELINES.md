# Tooltip Guidelines untuk Aplikasi Akuntansi

## Tujuan
Tooltip digunakan untuk memberikan informasi tambahan kepada user tentang field-field dalam form, terutama untuk field yang berhubungan dengan akuntansi, pajak, dan transaksi keuangan.

## Kapan Menggunakan Tooltip

### ✅ WAJIB menggunakan tooltip untuk:
1. **Field teknis akuntansi** (contoh: COA, GL Account, Depreciation Method)
2. **Field pajak** (contoh: PPN, PPh 21, PPh 23)
3. **Field perhitungan** (contoh: Exchange Rate, Discount, Tax Rate)
4. **Field status** dengan multiple options (contoh: Approval Status, Payment Status)
5. **Field yang mempengaruhi jurnal** (contoh: Revenue Account, Expense Account)

### ⚠️ TIDAK perlu tooltip untuk:
1. Field yang sangat jelas (contoh: Name, Email)
2. Field dengan label yang sudah deskriptif
3. Field yang sudah memiliki FormHelperText yang cukup jelas

## Cara Implementasi

### Metode 1: Menggunakan FormFieldWithTooltip Component

```tsx
import FormFieldWithTooltip from '@/components/common/FormFieldWithTooltip';

<FormFieldWithTooltip
  label="PPN Rate"
  tooltip="Tarif PPN/Pajak Pertambahan Nilai (default: 11%)"
  isRequired
  error={errors.ppn_rate?.message}
>
  <Input
    type="number"
    {...register('ppn_rate')}
  />
</FormFieldWithTooltip>
```

### Metode 2: Menggunakan Inline Tooltip (untuk custom layouts)

```tsx
import { FormControl, FormLabel, Tooltip, Icon, HStack } from '@chakra-ui/react';
import { FiInfo } from 'react-icons/fi';

<FormControl>
  <HStack spacing={1} mb={1}>
    <FormLabel mb={0}>Payment Method</FormLabel>
    <Tooltip 
      label="Metode pembayaran: Cash (tunai), Bank (transfer bank), Credit (kredit)"
      fontSize="sm"
      placement="top"
      hasArrow
      bg="gray.700"
      color="white"
    >
      <Box display="inline-flex" cursor="help">
        <Icon as={FiInfo} color="blue.500" boxSize={4} />
      </Box>
    </Tooltip>
  </HStack>
  <Select {...register('payment_method')}>
    <option value="CASH">Cash</option>
    <option value="BANK">Bank Transfer</option>
    <option value="CREDIT">Credit</option>
  </Select>
</FormControl>
```

## Best Practices untuk Menulis Tooltip Text

### ✅ DO:
- **Gunakan bahasa Indonesia** yang jelas dan mudah dipahami
- **Berikan konteks praktis** (contoh: "Tarif PPN yang dikenakan (default: 11%)")
- **Sertakan contoh** jika diperlukan (contoh: "Nomor referensi (contoh: TRF-2024-001)")
- **Jelaskan impact** jika field mempengaruhi perhitungan (contoh: "Akan mempengaruhi jurnal otomatis")
- **List options** untuk field dengan pilihan terbatas

### ❌ DON'T:
- Jangan terlalu panjang (maksimal 2-3 kalimat)
- Jangan mengulang label (label: "Tax Rate", tooltip: ❌ "Tax Rate")
- Jangan gunakan jargon tanpa penjelasan
- Jangan bertele-tele

## Contoh Tooltip Text yang Baik

### Sales/Purchase Forms:
```javascript
const tooltips = {
  customer: 'Pilih customer/pelanggan untuk transaksi ini',
  vendor: 'Pilih vendor/supplier untuk transaksi pembelian',
  paymentMethod: 'Metode pembayaran: Cash (tunai), Bank (transfer bank), Credit (kredit/hutang)',
  ppnRate: 'Tarif PPN/Pajak Pertambahan Nilai (default: 11%)',
  pph23Rate: 'Tarif PPh 23 untuk pemotongan pajak jasa',
  discount: 'Diskon global untuk seluruh transaksi (dalam persen)',
  dueDate: 'Tanggal jatuh tempo pembayaran (untuk transaksi kredit)',
};
```

### Asset Forms:
```javascript
const tooltips = {
  depreciationMethod: 'Metode penyusutan: Straight Line (garis lurus), Declining Balance (saldo menurun)',
  usefulLife: 'Masa manfaat aset dalam tahun (untuk perhitungan depresiasi)',
  salvageValue: 'Nilai sisa/residu aset setelah masa manfaat habis',
  purchasePrice: 'Harga perolehan aset (cost basis)',
};
```

### Payment Forms:
```javascript
const tooltips = {
  amount: 'Jumlah nominal pembayaran yang dilakukan',
  reference: 'Nomor referensi pembayaran (contoh: nomor transfer, nomor cek)',
  allocation: 'Alokasi pembayaran ke invoice atau purchase yang terkait',
  bankAccount: 'Akun kas/bank yang digunakan untuk pembayaran',
};
```

## Styling Guidelines

### Warna & Ukuran:
- Icon: `color="blue.500"` dengan `boxSize={4}`
- Tooltip background: `bg="gray.700"` (dark mode friendly)
- Font size: `fontSize="sm"` (untuk readability)
- Padding: `px={3} py={2}` (untuk spacing yang nyaman)

### Positioning:
- Default placement: `placement="top"`
- Gunakan `placement="right"` untuk form dengan banyak field vertikal
- Gunakan `placement="left"` untuk field di sisi kanan form

### Icon Hover Effect:
```tsx
<Icon 
  as={FiInfo} 
  color="blue.500" 
  boxSize={4}
  _hover={{ color: 'blue.600' }}
  cursor="help"
/>
```

## Checklist Implementasi

Saat menambahkan tooltip ke halaman baru, pastikan:

- [ ] Semua field akuntansi memiliki tooltip
- [ ] Semua field pajak memiliki tooltip
- [ ] Tooltip text menggunakan bahasa Indonesia
- [ ] Tooltip text tidak lebih dari 2-3 kalimat
- [ ] Icon info muncul di samping label
- [ ] Tooltip berfungsi saat di-hover
- [ ] Styling konsisten dengan existing tooltips
- [ ] Tooltip responsive di mobile

## Contoh Halaman dengan Tooltip Lengkap

Lihat implementasi di:
- `frontend/app/sales/page.tsx` - Sales page dengan tooltip
- `frontend/src/components/sales/SalesForm.tsx` - Sales form modal dengan tooltip
- `frontend/app/purchases/page.tsx` - Purchase page dengan tooltip
- `frontend/app/payments/page.tsx` - Payment page dengan tooltip
- `frontend/app/assets/page.tsx` - Asset page dengan tooltip
- `frontend/src/components/products/ProductCatalog.tsx` - Product page dengan tooltip
- `frontend/app/cash-bank/page.tsx` - Cash/Bank page dengan tooltip

## Testing Tooltip

Pastikan untuk test:
1. Hover interaction berfungsi
2. Text tooltip terbaca dengan jelas
3. Tidak overlap dengan elemen lain
4. Responsive di berbagai screen size
5. Dark mode compatibility
6. Touch interaction di mobile (long press)

## Maintenance

- Review tooltip text setiap 3-6 bulan
- Update jika ada perubahan fitur atau business logic
- Kumpulkan feedback dari user tentang clarity
- Tambahkan tooltip baru untuk fitur baru
