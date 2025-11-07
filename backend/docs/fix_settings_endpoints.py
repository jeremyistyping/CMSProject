#!/usr/bin/env python3
"""
Script untuk memperbaiki endpoint settings yang tidak sesuai dengan routes backend
"""

import json
import os
from datetime import datetime

def fix_settings_endpoints():
    """Fix settings endpoints to match actual backend routes"""
    print("🔧 FIXING SETTINGS ENDPOINTS TO MATCH BACKEND ROUTES")
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
    
    # Define what endpoints should exist based on backend routes
    # settings_routes.go analysis:
    correct_settings_endpoints = {
        "/api/v1/settings": ["GET", "PUT"],  # lines 30-31
        "/api/v1/settings/company": ["PUT"],  # line 34 - NO GET!
        "/api/v1/settings/system": ["PUT"],   # line 35 - NO GET!
        "/api/v1/settings/company/logo": ["POST"],  # line 36
        "/api/v1/settings/reset": ["POST"],   # line 39
        "/api/v1/settings/validation-rules": ["GET"],  # line 40
        "/api/v1/settings/history": ["GET"],  # line 41
    }
    
    incorrect_endpoints = []
    fixed_endpoints = []
    
    # Check each settings endpoint
    for path, path_data in paths.items():
        if path.startswith("/api/v1/settings"):
            if path in correct_settings_endpoints:
                correct_methods = correct_settings_endpoints[path]
                existing_methods = list(path_data.keys())
                
                for method in existing_methods:
                    if method.upper() not in correct_methods:
                        incorrect_endpoints.append(f"{method.upper()} {path}")
                        print(f"   ❌ Found incorrect: {method.upper()} {path}")
            else:
                incorrect_endpoints.append(f"UNKNOWN {path}")
                print(f"   ❌ Found unknown endpoint: {path}")
    
    if not incorrect_endpoints:
        print("✅ All settings endpoints are correct")
        return True
    
    # Create backup
    backup_file = swagger_file + f'.backup_settings_fix_{datetime.now().strftime("%Y%m%d_%H%M%S")}'
    with open(backup_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    print(f"💾 Backup created: {backup_file}")
    
    # Fix endpoints by removing incorrect methods
    for path, path_data in paths.items():
        if path.startswith("/api/v1/settings"):
            if path in correct_settings_endpoints:
                correct_methods = correct_settings_endpoints[path]
                existing_methods = list(path_data.keys())
                
                for method in existing_methods:
                    if method.upper() not in correct_methods:
                        print(f"   🗑️  Removing: {method.upper()} {path}")
                        del path_data[method]
                        fixed_endpoints.append(f"{method.upper()} {path}")
    
    # Save fixed swagger.json
    with open(swagger_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    
    print(f"✅ Fixed swagger.json saved")
    print(f"🔧 Fixed {len(fixed_endpoints)} incorrect endpoints")
    
    return True

def show_correct_settings_routes():
    """Show what the correct settings endpoints should be"""
    print(f"\n📋 CORRECT SETTINGS ENDPOINTS (from settings_routes.go)")
    print("=" * 60)
    
    correct_endpoints = [
        "GET    /api/v1/settings                   (GetSettings)",
        "PUT    /api/v1/settings                   (UpdateSettings)",
        "PUT    /api/v1/settings/company           (UpdateCompanyInfo)",
        "PUT    /api/v1/settings/system            (UpdateSystemConfig)",
        "POST   /api/v1/settings/company/logo      (UploadCompanyLogo)",
        "POST   /api/v1/settings/reset             (ResetToDefaults)",
        "GET    /api/v1/settings/validation-rules  (GetValidationRules)",
        "GET    /api/v1/settings/history           (GetSettingsHistory)",
    ]
    
    print("✅ These endpoints SHOULD exist:")
    for endpoint in correct_endpoints:
        print(f"   {endpoint}")
    
    print(f"\n❌ These endpoints should NOT exist:")
    wrong_endpoints = [
        "GET    /api/v1/settings/company           (NOT DEFINED IN ROUTES)",
        "GET    /api/v1/settings/system            (NOT DEFINED IN ROUTES)",
        "GET    /api/v1/settings/accounting        (NOT DEFINED IN ROUTES)",
    ]
    
    for endpoint in wrong_endpoints:
        print(f"   {endpoint}")

def main():
    """Main fix function"""
    print("🔧 SETTINGS ENDPOINT ROUTES FIXER")
    print("=" * 35)
    
    # Show correct vs wrong endpoints
    show_correct_settings_routes()
    
    # Fix the endpoints
    success = fix_settings_endpoints()
    
    if success:
        print(f"\n🎉 SETTINGS ENDPOINTS FIXED!")
        print(f"📝 Next steps:")
        print(f"1. Restart your Go backend server")
        print(f"2. Clear browser cache")
        print(f"3. Test settings endpoints - 404s should be resolved!")
        print(f"4. Only valid endpoints should remain in Swagger UI")
        return 0
    else:
        print(f"\n❌ Fix failed")
        return 1

if __name__ == "__main__":
    exit(main())