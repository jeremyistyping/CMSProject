# Weekly Reports - Week Picker Update

## Perubahan yang Dilakukan

Mengubah input form Weekly Report dari 2 field terpisah (Week Number + Year) menjadi 1 field Week Picker yang lebih praktis dan user-friendly.

## Sebelum (2 Fields)

```tsx
<FormControl isRequired>
  <FormLabel>Report Week</FormLabel>
  <Input type="number" name="week" min={1} max={53} />
</FormControl>

<FormControl isRequired>
  <FormLabel>Year</FormLabel>
  <Input type="number" name="year" min={2000} />
</FormControl>
```

## Sesudah (1 Field - Week Picker)

```tsx
<FormControl isRequired>
  <FormLabel>Report Week</FormLabel>
  <Input type="week" value={selectedWeek} onChange={handleWeekChange} />
</FormControl>
```

## Cara Kerja

### 1. State Management
```typescript
// Menyimpan format ISO week: "2025-W46"
const [selectedWeek, setSelectedWeek] = useState(getCurrentWeekString());

// Helper function untuk mendapat week saat ini
const getCurrentWeekString = () => {
  const now = new Date();
  const year = now.getFullYear();
  const week = weeklyReportService.getCurrentWeek();
  return `${year}-W${week.toString().padStart(2, '0')}`;
};
```

### 2. Week Change Handler
```typescript
const handleWeekChange = (e: React.ChangeEvent<HTMLInputElement>) => {
  const weekString = e.target.value; // Format: YYYY-Www (e.g., 2025-W46)
  setSelectedWeek(weekString);

  if (weekString) {
    // Parse the week string
    const [yearStr, weekStr] = weekString.split('-W');
    const year = parseInt(yearStr);
    const week = parseInt(weekStr);

    // Update form data dengan week dan year yang sudah di-extract
    setFormData((prev) => ({
      ...prev,
      week,
      year,
    }));
  }
};
```

## Fitur Week Picker

### Visual Interface
- ✅ Calendar view dengan highlight per minggu
- ✅ Navigasi bulan dengan arrow buttons
- ✅ Display format: "Week 46, 2025" atau serupa
- ✅ Visual highlighting untuk minggu yang dipilih
- ✅ Auto-select current week sebagai default

### Keuntungan
1. **User Experience**: Lebih intuitif dengan visual calendar
2. **Praktis**: 1 field menggantikan 2 field
3. **Error Prevention**: Tidak bisa salah input kombinasi week/year
4. **Standard HTML5**: Menggunakan input type="week" native
5. **Mobile Friendly**: Browser mobile memiliki picker native yang bagus

## Format Data

### Input Format (User sees):
```
Week 46, 2025
```

### Internal Format (Stored in state):
```
selectedWeek: "2025-W46"  // ISO 8601 week format
```

### Backend Format (API payload):
```json
{
  "week": 46,
  "year": 2025
}
```

## Browser Compatibility

Week picker didukung oleh:
- ✅ Chrome/Edge (Full support)
- ✅ Safari (Full support)
- ✅ Firefox (Full support)
- ✅ Mobile browsers (Native picker)

## Testing Checklist

- [ ] Pilih week dari calendar picker
- [ ] Verify week number dan year ter-extract dengan benar
- [ ] Submit form dan cek payload ke backend
- [ ] Test dengan minggu di tahun yang berbeda
- [ ] Test dengan minggu terakhir tahun (W52/W53)
- [ ] Test dengan minggu pertama tahun (W01)
- [ ] Verify default value adalah current week

## Screenshots

Tampilan Week Picker akan serupa dengan contoh yang diberikan:
- Calendar grid dengan weeks highlighted
- Week numbers di sidebar kiri
- Navigation arrows untuk ganti bulan/tahun
- Clear visual indication untuk week yang dipilih

## File Modified

```
frontend/src/components/projects/WeeklyReportsTab.tsx
```

## Next Steps

Form sekarang siap digunakan dengan week picker yang lebih praktis. User tinggal klik pada calendar dan memilih minggu yang diinginkan.

