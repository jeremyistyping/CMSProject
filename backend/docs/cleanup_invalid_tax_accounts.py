#!/usr/bin/env python3
"""
Script untuk menghapus endpoint tax-accounts yang tidak ada di backend routes
"""

import json
import os
from datetime import datetime

def cleanup_invalid_tax_accounts():
    """Remove tax-accounts endpoints that don't exist in backend routes"""
    print("🧹 CLEANING INVALID TAX-ACCOUNTS ENDPOINTS")
    print("=" * 45)
    
    swagger_file = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\docs\\swagger.json"
    
    # Load swagger.json
    try:
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = json.load(f)
        print(f"✅ Loaded swagger.json")
    except Exception as e:
        print(f"❌ Error loading swagger.json: {e}")
        return False
    
    paths = swagger_data.get('paths', {})
    original_count = len(paths)
    print(f"📊 Original endpoints: {original_count}")
    
    # Based on settings_routes.go analysis, these endpoints do NOT exist:
    invalid_tax_accounts_endpoints = [
        "/api/v1/tax-accounts/all",                    # NOT DEFINED in routes
        "/api/v1/tax-accounts/available-accounts",     # NOT DEFINED in routes
    ]
    
    # These endpoints DO exist (for comparison):
    valid_tax_accounts_endpoints = [
        "/api/v1/tax-accounts/current",           # line 49 - GetCurrentSettings
        "/api/v1/tax-accounts",                   # line 52 - GetAllSettings (admin)
        "/api/v1/tax-accounts/{id}",              # line 58 - UpdateSettings
        "/api/v1/tax-accounts/{id}/activate",     # line 61 - ActivateSettings
        "/api/v1/tax-accounts/accounts",          # line 64 - GetAvailableAccounts
        "/api/v1/tax-accounts/validate",          # line 67 - ValidateAccountConfiguration
        "/api/v1/tax-accounts/refresh-cache",     # line 70 - RefreshCache
        "/api/v1/tax-accounts/suggestions",       # line 73 - GetAccountSuggestions
        "/api/v1/tax-accounts/status",            # line 76 - GetStatus
    ]
    
    print(f"\n❌ These endpoints should NOT exist (will be removed):")
    for endpoint in invalid_tax_accounts_endpoints:
        if endpoint in paths:
            print(f"   🗑️  Found: {endpoint}")
        else:
            print(f"   ℹ️  Not found: {endpoint}")
    
    print(f"\n✅ These endpoints SHOULD exist (will be kept):")
    for endpoint in valid_tax_accounts_endpoints:
        if endpoint in paths:
            print(f"   ✓ Found: {endpoint}")
        else:
            print(f"   ⚠️  Missing: {endpoint}")
    
    # Remove invalid endpoints
    removed_endpoints = []
    for endpoint in invalid_tax_accounts_endpoints:
        if endpoint in paths:
            del paths[endpoint]
            removed_endpoints.append(endpoint)
            print(f"\n🗑️  Removed: {endpoint}")
    
    if not removed_endpoints:
        print("\n✅ No invalid tax-accounts endpoints found")
        return True
    
    # Create backup
    backup_file = swagger_file + f'.backup_cleanup_tax_accounts_{datetime.now().strftime("%Y%m%d_%H%M%S")}'
    with open(backup_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    print(f"💾 Backup created: {backup_file}")
    
    # Save cleaned swagger.json
    with open(swagger_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    
    new_count = len(paths)
    print(f"✅ Cleaned swagger.json saved")
    print(f"📊 Endpoints after cleanup: {new_count}")
    print(f"🗑️  Removed {len(removed_endpoints)} invalid endpoints")
    
    return True

def show_route_mapping():
    """Show correct mapping between swagger endpoints and backend routes"""
    print(f"\n📋 CORRECT TAX-ACCOUNTS ENDPOINTS MAPPING")
    print("=" * 50)
    print("(Based on settings_routes.go analysis)")
    
    mappings = [
        ("GET    /api/v1/tax-accounts/current", "line 49 - GetCurrentSettings"),
        ("GET    /api/v1/tax-accounts", "line 52 - GetAllSettings (admin only)"),
        ("POST   /api/v1/tax-accounts", "line 55 - CreateSettings"),
        ("PUT    /api/v1/tax-accounts/{id}", "line 58 - UpdateSettings"),
        ("POST   /api/v1/tax-accounts/{id}/activate", "line 61 - ActivateSettings"),
        ("GET    /api/v1/tax-accounts/accounts", "line 64 - GetAvailableAccounts"),
        ("POST   /api/v1/tax-accounts/validate", "line 67 - ValidateAccountConfiguration"),
        ("POST   /api/v1/tax-accounts/refresh-cache", "line 70 - RefreshCache (admin)"),
        ("GET    /api/v1/tax-accounts/suggestions", "line 73 - GetAccountSuggestions"),
        ("GET    /api/v1/tax-accounts/status", "line 76 - GetStatus"),
        ("GET    /api/v1/tax-accounts/validate", "line 77 - ValidateAccountSelection"),
    ]
    
    print("✅ VALID endpoints:")
    for endpoint, route in mappings:
        print(f"   {endpoint:45} → {route}")
    
    print(f"\n❌ INVALID endpoints (not in routes):")
    invalid = [
        "GET    /api/v1/tax-accounts/all",
        "GET    /api/v1/tax-accounts/available-accounts",
    ]
    for endpoint in invalid:
        print(f"   {endpoint:45} → NOT DEFINED IN BACKEND")

def main():
    """Main cleanup function"""
    print("🔧 TAX-ACCOUNTS ENDPOINT CLEANUP TOOL")
    print("=" * 40)
    
    # Show correct endpoint mapping
    show_route_mapping()
    
    # Clean up invalid endpoints
    success = cleanup_invalid_tax_accounts()
    
    if success:
        print(f"\n🎉 TAX-ACCOUNTS CLEANUP COMPLETED!")
        print(f"📝 Next steps:")
        print(f"1. Restart your Go backend server")
        print(f"2. Clear browser cache")
        print(f"3. Test tax-accounts endpoints:")
        print(f"   • /api/v1/tax-accounts/accounts (should work)")
        print(f"   • /api/v1/tax-accounts/current (should work)")
        print(f"4. The 404 errors should be resolved!")
        return 0
    else:
        print(f"\n❌ Cleanup failed")
        return 1

if __name__ == "__main__":
    exit(main())