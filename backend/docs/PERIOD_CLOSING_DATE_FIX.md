# Period Closing Date Fix - Fiscal Year Start Issue

## Problem
Pada UI Settings, field "Dari Tanggal (Start Date)" di section Period Closing menampilkan tanggal dari "Fiscal Year Start" (misalnya 01/01/2025) alih-alih mengikuti tanggal terakhir tutup buku + 1 hari.

## Root Cause
Backend API `GET /api/v1/period-closing/last-info` sudah benar dan mengembalikan:
- `has_previous_closing`: boolean
- `last_closing_date`: tanggal tutup buku terakhir
- `next_start_date`: tanggal awal periode berikutnya (last_closing_date + 1 hari)
- `period_start_date`: tanggal transaksi paling awal (untuk first closing)

Namun frontend tidak selalu menggunakan data ini dengan benar, terutama saat initial render.

## Solution

### 1. Backend (Already Correct)
File: `services/period_closing_service.go`

Method `GetLastClosingInfo` sudah benar:
```go
// Line 49-51
// Next period starts the day after last closing
nextStart := lastPeriod.EndDate.AddDate(0, 0, 1)
info.NextStartDate = &nextStart
```

### 2. Frontend Fix
File: `frontend/app/settings/page.tsx`

#### Changes Made:

**A. Improved `fetchLastClosingInfo` function (lines 248-276)**
- Menambahkan console logging untuk debugging
- Memastikan `periodStartDate` diisi dari `next_start_date` (bukan fiscal_year_start)
- Format tanggal yang benar: `split('T')[0]` untuk mengambil YYYY-MM-DD

```typescript
if (info.has_previous_closing && info.next_start_date) {
  // Has previous closing - MUST use next start date (day after last closing)
  const nextStart = info.next_start_date.split('T')[0];
  setPeriodStartDate(nextStart);
  console.log('✅ Period start date set from last closing:', nextStart);
}
```

**B. Enhanced UI Visual Feedback (lines 867-892)**
- Field "Start Date" readonly jika ada previous closing
- Background biru (`blue.50`) untuk menandakan field locked
- Label tambahan "\ud83d\udd12 Locked - Auto-filled"
- Helper text lebih jelas: menampilkan tanggal tutup buku terakhir
- Cursor `not-allowed` untuk memberikan feedback visual

```tsx
<Input
  type="date"
  value={periodStartDate}
  bg={lastClosingInfo?.has_previous_closing ? 'blue.50' : undefined}
  isReadOnly={lastClosingInfo?.has_previous_closing === true}
  cursor={lastClosingInfo?.has_previous_closing ? 'not-allowed' : 'text'}
/>
```

## Behavior After Fix

### Scenario 1: Has Previous Closing
- Field "Dari Tanggal" **otomatis terisi** dengan `next_start_date` dari API
- Field **locked** (readonly, background biru)
- Helper text: "\ud83d\udd12 Otomatis diisi dari tutup buku terakhir (DD/MM/YYYY). Tanggal mulai harus 1 hari setelahnya."
- User **tidak bisa** mengubah tanggal ini

### Scenario 2: First Time Closing (No Previous Closing)
- Field "Dari Tanggal" terisi dengan `period_start_date` (tanggal transaksi pertama)
- Field **editable**
- Helper text: "\ud83d\udcc5 Pilih tanggal awal periode yang akan ditutup (first closing - gunakan tanggal transaksi pertama)"
- User **bisa** mengubah tanggal sesuai kebutuhan

### Scenario 3: No Data Available
- Field kosong
- User harus input manual
- Console log: "\u26a0\ufe0f No closing info available, user must input start date manually"

## API Endpoints Reference

### GET /api/v1/period-closing/last-info
Returns:
```json
{
  "success": true,
  "data": {
    "has_previous_closing": true,
    "last_closing_date": "2024-12-31T00:00:00Z",
    "next_start_date": "2025-01-01T00:00:00Z",
    "period_start_date": null
  }
}
```

### GET /api/v1/period-closing/preview
Query params: `start_date`, `end_date`

Validates:
- Tanggal continuity (harus lanjut dari last closing)
- Overlap dengan closed periods
- Unbalanced/draft entries
- Retained earnings account availability

## Testing Checklist

- [x] Backend API mengembalikan `next_start_date` dengan benar
- [x] Frontend `fetchLastClosingInfo` dipanggil saat page load
- [x] Field "Start Date" otomatis terisi dari API
- [x] Field "Start Date" readonly jika ada previous closing
- [x] Visual feedback (blue background, locked icon) ditampilkan
- [x] Helper text menampilkan informasi yang jelas
- [x] Console logging untuk debugging
- [x] User tidak bisa edit tanggal jika sudah ada previous closing

## Related Files

### Backend
- `models/accounting_period.go` - Data models
- `services/period_closing_service.go` - Business logic
- `controllers/period_closing_controller.go` - API endpoints

### Frontend
- `app/settings/page.tsx` - UI dan logic

## Notes

- **Fiscal Year Start** (`fiscal_year_start` di settings) **OTOMATIS DI-UPDATE** setelah period closing berhasil
- Setelah tutup buku, `fiscal_year_start` akan diset ke tanggal 1 hari setelah end date period closing
- Contoh: Tutup buku 31/12/2024 → fiscal_year_start otomatis jadi "January 1" (01/01/2025)
- Period closing **harus berkesinambungan** (continuous), tidak boleh ada gap
- Backend sudah menghandle validasi untuk mencegah overlap atau gap
- Frontend sekarang lebih jelas menunjukkan constraint ini kepada user

## Prevention for Future

- Jangan gunakan `fiscal_year_start` untuk auto-fill period start date
- Selalu ambil dari `last-info` API endpoint
- Pastikan field locked jika ada previous closing
- Tambahkan visual feedback yang jelas untuk auto-filled fields

---

**Created**: 2025-01-07
**Author**: AI Assistant
**Status**: Fixed & Documented
