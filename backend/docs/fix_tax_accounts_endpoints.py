#!/usr/bin/env python3
"""
Script untuk memperbaiki endpoint tax-accounts yang menggunakan prefix salah
"""

import json
import os
from datetime import datetime

def fix_tax_accounts_endpoints():
    """Fix tax-accounts endpoints that have wrong prefix"""
    print("🔧 FIXING TAX-ACCOUNTS ENDPOINTS PREFIX")
    print("=" * 40)
    
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
    
    # Find endpoints with wrong prefix /api/v1/settings/tax-accounts
    wrong_prefix = "/api/v1/settings/tax-accounts"
    correct_prefix = "/api/v1/tax-accounts"
    
    endpoints_to_fix = []
    for path in paths.keys():
        if path.startswith(wrong_prefix):
            endpoints_to_fix.append(path)
    
    print(f"🔍 Found {len(endpoints_to_fix)} endpoints with wrong prefix:")
    for endpoint in endpoints_to_fix:
        print(f"   ❌ {endpoint}")
    
    if not endpoints_to_fix:
        print("✅ No endpoints found with wrong prefix")
        return True
    
    # Create backup
    backup_file = swagger_file + f'.backup_tax_accounts_fix_{datetime.now().strftime("%Y%m%d_%H%M%S")}'
    with open(backup_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    print(f"💾 Backup created: {backup_file}")
    
    # Fix endpoints
    fixed_endpoints = []
    new_paths = {}
    
    for old_path, path_data in paths.items():
        if old_path.startswith(wrong_prefix):
            # Create new path with correct prefix
            new_path = old_path.replace(wrong_prefix, correct_prefix, 1)
            new_paths[new_path] = path_data
            fixed_endpoints.append((old_path, new_path))
            print(f"   🔄 {old_path} → {new_path}")
        else:
            # Keep existing paths
            new_paths[old_path] = path_data
    
    # Update paths in swagger data
    swagger_data['paths'] = new_paths
    
    # Save fixed swagger.json
    with open(swagger_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    
    new_count = len(new_paths)
    print(f"✅ Fixed swagger.json saved")
    print(f"📊 Endpoints after fix: {new_count}")
    print(f"🔧 Fixed {len(fixed_endpoints)} endpoints")
    
    return True

def show_correct_endpoints():
    """Show what the correct tax-accounts endpoints should be"""
    print(f"\n📋 CORRECT TAX-ACCOUNTS ENDPOINTS")
    print("=" * 35)
    
    correct_endpoints = [
        "GET    /api/v1/tax-accounts/current",
        "GET    /api/v1/tax-accounts",
        "POST   /api/v1/tax-accounts",
        "PUT    /api/v1/tax-accounts/{id}",
        "POST   /api/v1/tax-accounts/{id}/activate",
        "GET    /api/v1/tax-accounts/accounts",
        "POST   /api/v1/tax-accounts/validate",
        "POST   /api/v1/tax-accounts/refresh-cache",
        "GET    /api/v1/tax-accounts/suggestions",
        "GET    /api/v1/tax-accounts/status",
    ]
    
    print("✅ These are the CORRECT endpoint paths:")
    for endpoint in correct_endpoints:
        print(f"   {endpoint}")
    
    print(f"\n❌ These were WRONG (now fixed):")
    wrong_endpoints = [
        "GET    /api/v1/settings/tax-accounts",
        "POST   /api/v1/settings/tax-accounts",
        "GET    /api/v1/settings/tax-accounts/all",
        "GET    /api/v1/settings/tax-accounts/available-accounts",
        "POST   /api/v1/settings/tax-accounts/refresh-cache",
        "GET    /api/v1/settings/tax-accounts/suggestions",
        "POST   /api/v1/settings/tax-accounts/validate",
        "PUT    /api/v1/settings/tax-accounts/{id}",
        "POST   /api/v1/settings/tax-accounts/{id}/activate",
    ]
    
    for endpoint in wrong_endpoints:
        print(f"   {endpoint} (WRONG PREFIX)")

def main():
    """Main fix function"""
    print("🔧 TAX-ACCOUNTS ENDPOINT PREFIX FIXER")
    print("=" * 40)
    
    # Show correct vs wrong endpoints
    show_correct_endpoints()
    
    # Fix the endpoints
    success = fix_tax_accounts_endpoints()
    
    if success:
        print(f"\n🎉 TAX-ACCOUNTS ENDPOINTS FIXED!")
        print(f"📝 Next steps:")
        print(f"1. Restart your Go backend server")
        print(f"2. Clear browser cache")
        print(f"3. Test tax-accounts endpoints - they should work now!")
        print(f"4. All 404 errors for tax-accounts should be resolved")
        return 0
    else:
        print(f"\n❌ Fix failed")
        return 1

if __name__ == "__main__":
    exit(main())