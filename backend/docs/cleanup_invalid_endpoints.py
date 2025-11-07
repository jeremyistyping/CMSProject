#!/usr/bin/env python3
"""
Script untuk membersihkan endpoint yang tidak ada dari swagger.json
"""

import json
import os
from datetime import datetime

def cleanup_swagger():
    """Remove invalid endpoints from swagger.json"""
    print("🧹 CLEANING UP INVALID ENDPOINTS FROM SWAGGER.JSON")
    print("=" * 55)
    
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
    
    # List of invalid endpoints to remove
    invalid_endpoints = [
        "/api/v1/settings/accounting",  # This endpoint doesn't exist in backend
        # Add more invalid endpoints here if found
    ]
    
    removed_endpoints = []
    
    # Remove invalid endpoints
    for endpoint in invalid_endpoints:
        if endpoint in paths:
            del paths[endpoint]
            removed_endpoints.append(endpoint)
            print(f"🗑️  Removed: {endpoint}")
        else:
            print(f"ℹ️  Not found: {endpoint}")
    
    if removed_endpoints:
        # Create backup
        backup_file = swagger_file + f'.backup_cleanup_{datetime.now().strftime("%Y%m%d_%H%M%S")}'
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
    else:
        print("✅ No invalid endpoints found - swagger.json is clean")
        return True

def verify_real_settings_endpoints():
    """Show what settings endpoints actually exist based on server logs"""
    print(f"\n📋 ACTUAL SETTINGS ENDPOINTS (from server logs)")
    print("=" * 50)
    
    actual_endpoints = [
        "GET    /api/v1/settings",
        "PUT    /api/v1/settings", 
        "PUT    /api/v1/settings/company",
        "PUT    /api/v1/settings/system",
        "POST   /api/v1/settings/company/logo",
        "POST   /api/v1/settings/reset",
        "GET    /api/v1/settings/validation-rules", 
        "GET    /api/v1/settings/history",
        "GET    /api/v1/tax-accounts/current",
        "GET    /api/v1/tax-accounts",
        "POST   /api/v1/tax-accounts",
        "PUT    /api/v1/tax-accounts/:id",
        "POST   /api/v1/tax-accounts/:id/activate",
        "GET    /api/v1/tax-accounts/accounts",
        "POST   /api/v1/tax-accounts/validate",
        "POST   /api/v1/tax-accounts/refresh-cache",
        "GET    /api/v1/tax-accounts/suggestions",
        "GET    /api/v1/tax-accounts/status",
        "GET    /api/v1/tax-accounts/validate",
    ]
    
    print("✅ These endpoints ACTUALLY exist in backend:")
    for endpoint in actual_endpoints:
        print(f"   {endpoint}")
    
    print(f"\n❌ These endpoints were incorrectly added to swagger.json:")
    print(f"   GET    /api/v1/settings/accounting  (DOES NOT EXIST)")

def main():
    """Main cleanup function"""
    print("🔧 SWAGGER ENDPOINT CLEANUP TOOL")
    print("=" * 35)
    
    # Show what endpoints actually exist
    verify_real_settings_endpoints()
    
    # Clean up swagger.json
    success = cleanup_swagger()
    
    if success:
        print(f"\n🎉 CLEANUP COMPLETED!")
        print(f"📝 Next steps:")
        print(f"1. Restart your Go backend server")
        print(f"2. Test the corrected endpoints")
        print(f"3. /api/v1/settings/accounting should no longer appear (404 fixed)")
        return 0
    else:
        print(f"\n❌ Cleanup failed")
        return 1

if __name__ == "__main__":
    exit(main())