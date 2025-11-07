# Asset Master - Cash & Bank Integration

## Overview
Implementasi integrasi Asset Master dengan Cash & Bank accounts yang sudah terintegrasi dengan Chart of Accounts (COA).

## Problem Statement
Sebelumnya, Asset Master mengambil account bank langsung dari COA table (`/accounts?category=CURRENT_ASSET&types=ASSET`). 
Hal ini menyebabkan:
- Tidak ada informasi bank seperti bank name dan account number
- Balance tidak sync dengan transaksi Cash & Bank
- Duplikasi data antara Cash & Bank module dan COA

## Solution Implemented

### 1. Modified AssetService.getBankAccounts() 
**File**: `frontend/src/services/assetService.ts`

Mengubah endpoint dari:
```typescript
// BEFORE: Direct COA query
const response = await api.get('/accounts?category=CURRENT_ASSET&types=ASSET');

// AFTER: Use integrated Cash & Bank accounts
const response = await api.get('/cashbank/payment-accounts');
```

### 2. Enhanced Data Transformation
Data yang diterima dari Cash & Bank module ditransformasi untuk kompatibilitas dengan Asset Form:

```typescript
data: response.data.data.map((account: any) => ({
  id: account.id,
  code: account.code,
  name: account.name,
  type: account.type, // CASH or BANK
  balance: account.balance,
  currency: account.currency || 'IDR',
  bank_name: account.bank_name,
  account_no: account.account_no,
  account_id: account.account_id, // Link to COA
  // Enhanced display info
  display_name: account.bank_name 
    ? `${account.name} - ${account.bank_name} (${account.account_no})`
    : account.name,
  balance_formatted: new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: account.currency || 'IDR',
    minimumFractionDigits: 0
  }).format(account.balance)
}))
```

### 3. Updated Asset Form Dropdown
**File**: `frontend/app/assets/page.tsx`

- Filter berdasarkan `account.type` (`CASH` atau `BANK`) bukan lagi berdasarkan nama
- Menampilkan informasi yang lebih lengkap: bank name, account number, dan balance
- Menambah help text yang menjelaskan integrasi dengan COA
- Warning message jika tidak ada Cash & Bank accounts

```typescript
// Filter by actual type instead of name matching
.filter(account => 
  formData.paymentMethod === 'CASH' 
    ? account.type === 'CASH'
    : account.type === 'BANK'
)

// Enhanced display with bank info and balance
<option key={account.id} value={account.id}>
  {account.code} - {account.display_name || account.name} 
  {account.balance_formatted && ` (${account.balance_formatted})`}
</option>
```

## Benefits

1. **Accurate Data**: Balance dan informasi bank selalu accurate karena langsung dari Cash & Bank module
2. **Better UX**: User melihat informasi bank yang lengkap (bank name, account number, balance)
3. **Data Consistency**: Tidak ada duplikasi data antara Cash & Bank dan COA
4. **Real-time Balance**: Balance yang ditampilkan real-time dari transaksi Cash & Bank
5. **Type Safety**: Filter berdasarkan type yang benar (CASH/BANK) bukan string matching

## Integration Flow

```
Asset Master Form
       ↓
  assetService.getBankAccounts()
       ↓
  /cashbank/payment-accounts endpoint
       ↓
  Cash Banks Table (with COA integration)
       ↓
  Returns: Cash & Bank accounts linked to COA
       ↓
  Asset Form displays integrated accounts
```

## Testing Checklist

- [x] Asset form loads Cash & Bank accounts instead of direct COA
- [x] Dropdown shows bank information (name, account number)
- [x] Balance displayed correctly from Cash & Bank transactions  
- [x] Filter works correctly (CASH vs BANK type)
- [x] Warning shown when no accounts found
- [x] Form submission works with selected Cash & Bank account

## API Dependencies

- `/cashbank/payment-accounts` endpoint must be available
- Cash & Bank accounts must have `account_id` linked to COA
- Proper balance sync between Cash & Bank transactions and COA

## Notes

- Existing Asset records will continue to work
- New Asset records will use integrated Cash & Bank accounts
- Backend processing remains unchanged (still receives account IDs)
