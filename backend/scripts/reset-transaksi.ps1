# SCRIPT RESET DATA TRANSAKSI SISTEM AKUNTANSI
# Script PowerShell untuk memudahkan eksekusi reset

Write-Host "=== SISTEM AKUNTANSI - RESET DATA TRANSAKSI ===" -ForegroundColor Yellow
Write-Host ""
Write-Host "Pilih jenis reset yang ingin dilakukan:" -ForegroundColor Cyan
Write-Host "1. Reset data transaksi saja (COA dipertahankan) - DISARANKAN" -ForegroundColor Green
Write-Host "2. Reset database total (hapus semua data) - BERBAHAYA!" -ForegroundColor Red
Write-Host "3. Restore COA dari backup" -ForegroundColor Blue
Write-Host "4. Keluar" -ForegroundColor Gray
Write-Host ""

$choice = Read-Host "Masukkan pilihan (1-4)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "🔄 Menjalankan reset data transaksi..." -ForegroundColor Yellow
        Write-Host "COA dan master data akan dipertahankan" -ForegroundColor Green
        Write-Host ""
        
        # Build and run reset transaction data (versi GORM yang aman)
        go build -o reset_transaction_data.exe cmd/reset_transaction_data_gorm.go
        if ($LASTEXITCODE -eq 0) {
            .\reset_transaction_data.exe
            Remove-Item -Path "reset_transaction_data.exe" -ErrorAction SilentlyContinue
        } else {
            Write-Host "❌ Gagal build script reset" -ForegroundColor Red
        }
    }
    
    "2" {
        Write-Host ""
        Write-Host "⚠️  PERINGATAN: Anda akan menghapus SEMUA DATA!" -ForegroundColor Red
        Write-Host "Ini termasuk COA, master data, dan semua transaksi" -ForegroundColor Red
        Write-Host ""
        
        $confirm = Read-Host "Ketik 'KONFIRMASI TOTAL RESET' untuk melanjutkan"
        if ($confirm -eq "KONFIRMASI TOTAL RESET") {
            Write-Host ""
            Write-Host "🔄 Menjalankan total reset database..." -ForegroundColor Red
            
            # Build and run total reset
            go build -o reset_database_total.exe cmd/reset_database_total.go
            if ($LASTEXITCODE -eq 0) {
                .\reset_database_total.exe
                Remove-Item -Path "reset_database_total.exe" -ErrorAction SilentlyContinue
            } else {
                Write-Host "❌ Gagal build script total reset" -ForegroundColor Red
            }
        } else {
            Write-Host "Total reset dibatalkan" -ForegroundColor Yellow
        }
    }
    
    "3" {
        Write-Host ""
        Write-Host "🔄 Menjalankan restore COA dari backup..." -ForegroundColor Blue
        
        # Build and run restore COA
        go build -o restore_coa_from_backup.exe cmd/restore_coa_from_backup.go
        if ($LASTEXITCODE -eq 0) {
            .\restore_coa_from_backup.exe
            Remove-Item -Path "restore_coa_from_backup.exe" -ErrorAction SilentlyContinue
        } else {
            Write-Host "❌ Gagal build script restore" -ForegroundColor Red
        }
    }
    
    "4" {
        Write-Host "Keluar dari script" -ForegroundColor Gray
        exit 0
    }
    
    default {
        Write-Host "Pilihan tidak valid" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Script selesai dieksekusi" -ForegroundColor Green
Write-Host "Tekan Enter untuk keluar..."
Read-Host
