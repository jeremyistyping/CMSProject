# 🔧 Fix Hydration Mismatch Error - Next.js 15

## 📋 Problem

Aplikasi mengalami **Hydration Mismatch Error** yang disebabkan oleh:

1. **Inline styles yang berbeda** antara server-side render (SSR) dan client-side render
2. **localStorage access** sebelum component mounted
3. **Content yang berbeda** antara server dan client saat initial render

Error yang muncul:
```
Hydration failed because the server rendered text didn't match the client.
```

---

## ✅ Solusi yang Diterapkan

### 1. **Fix `app/page.tsx`** - Menghindari localStorage access di SSR

**Sebelum:**
```tsx
// ❌ Masalah: localStorage di-access langsung
const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
```

**Sesudah:**
```tsx
// ✅ Solusi: Wait for client-side mount
const [mounted, setMounted] = React.useState(false);

React.useEffect(() => {
  setMounted(true);
}, []);

React.useEffect(() => {
  if (!mounted) return;
  const token = localStorage.getItem('token');
  // ... redirect logic
}, [mounted]);
```

---

### 2. **Fix `app/ClientProviders.tsx`** - Menghapus inline styles kompleks

**Sebelum:**
```tsx
// ❌ Masalah: Inline styles dengan banyak properties
if (!mounted) {
  return (
    <div style={{
      fontSize: '1.25rem',
      fontWeight: '600',
      marginBottom: '0.5rem'
    }}>
      Loading...
    </div>
  );
}
```

**Sesudah:**
```tsx
// ✅ Solusi: Return null saat SSR
if (!mounted) {
  return null;
}
```

---

### 3. **Update `app/globals.css`** - Prevent Flash of Unstyled Content (FOUC)

Ditambahkan CSS untuk menyembunyikan content sebelum hydration complete:

```css
/* Prevent Flash of Unstyled Content (FOUC) during hydration */
html:not(.hydrated) {
  visibility: hidden;
}

html.hydrated {
  visibility: visible;
}
```

---

### 4. **Update `app/layout.tsx`** - Menandai hydration complete

Ditambahkan script untuk menambahkan class `hydrated` setelah DOM ready:

```javascript
// Mark as hydrated to prevent FOUC
window.addEventListener('DOMContentLoaded', function() {
  document.documentElement.classList.add('hydrated');
});

// Fallback if DOMContentLoaded already fired
if (document.readyState === 'complete' || document.readyState === 'interactive') {
  document.documentElement.classList.add('hydrated');
}
```

---

## 🎯 Best Practices untuk Menghindari Hydration Mismatch

### 1. **Gunakan `mounted` state untuk client-only code**

```tsx
const [mounted, setMounted] = useState(false);

useEffect(() => {
  setMounted(true);
}, []);

if (!mounted) return null; // atau loading skeleton
```

### 2. **Hindari inline styles yang kompleks di SSR**

❌ **Jangan:**
```tsx
<div style={{ fontSize: '1.25rem', fontWeight: '600' }}>Text</div>
```

✅ **Lakukan:**
```tsx
<div className="text-xl font-semibold">Text</div>
```

### 3. **Hindari akses `window`, `document`, `localStorage` di render awal**

❌ **Jangan:**
```tsx
const theme = localStorage.getItem('theme'); // Error di SSR
```

✅ **Lakukan:**
```tsx
useEffect(() => {
  const theme = localStorage.getItem('theme');
  // use theme
}, []);
```

### 4. **Gunakan `suppressHydrationWarning` untuk content yang memang berbeda**

```tsx
<html suppressHydrationWarning>
  {/* Theme class yang diatur via script akan berbeda */}
</html>
```

### 5. **Gunakan `ClientOnly` wrapper untuk component yang hanya berjalan di client**

```tsx
'use client';

export default function ClientOnly({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;

  return <>{children}</>;
}
```

---

## 🧪 Testing

Setelah fix diterapkan, pastikan:

1. ✅ Tidak ada error hydration di console
2. ✅ Tidak ada "flash" atau perubahan visual saat page load
3. ✅ Theme (dark/light) terdeteksi dengan benar
4. ✅ Redirect ke `/dashboard` atau `/login` berjalan lancar
5. ✅ Tidak ada warning di console terkait hydration

---

## 📚 Reference

- [Next.js Hydration Errors](https://nextjs.org/docs/messages/react-hydration-error)
- [React Hydration](https://react.dev/reference/react-dom/client/hydrateRoot)
- [suppressHydrationWarning](https://react.dev/reference/react-dom/client/hydrateRoot#suppressing-unavoidable-hydration-mismatch-errors)

---

## 🔍 Common Hydration Error Causes

1. **Browser extensions** yang memodifikasi HTML
2. **Date/Time formatting** yang berbeda antara server dan client
3. **Random values** atau `Math.random()`
4. **localStorage/sessionStorage** access di render
5. **Window size** (`window.innerWidth`) di render
6. **User agent detection** yang berbeda

---

**Status:** ✅ **FIXED**

**Date:** November 10, 2025

**Next.js Version:** 15.5.4

