#!/usr/bin/env python3
"""
Script untuk memverifikasi bahwa Assets dan Settings sudah benar-benar ada di Swagger
"""

import json
import os

def verify_swagger():
    """Verify Assets and Settings are properly configured in swagger.json"""
    
    base_dir = os.path.dirname(os.path.abspath(__file__))
    swagger_file = os.path.join(base_dir, 'swagger.json')
    
    print("🔍 Verifying Swagger documentation for Assets and Settings...")
    print(f"📄 File: {swagger_file}")
    print("=" * 60)
    
    # Load swagger.json
    try:
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = json.load(f)
    except Exception as e:
        print(f"❌ Error loading swagger.json: {e}")
        return False
    
    # Check basic structure
    if 'paths' not in swagger_data:
        print("❌ No 'paths' section found in swagger.json")
        return False
        
    if 'tags' not in swagger_data:
        print("❌ No 'tags' section found in swagger.json")
        return False
    
    paths = swagger_data['paths']
    tags = swagger_data['tags']
    
    print(f"📊 Total endpoints: {len(paths)}")
    print(f"🏷️  Total tags: {len(tags)}")
    print()
    
    # Check tags
    tag_names = [tag.get('name', '') for tag in tags if isinstance(tag, dict)]
    print("🏷️  Available tags:")
    for tag_name in sorted(tag_names):
        print(f"   ✓ {tag_name}")
    print()
    
    # Check Assets
    assets_paths = [p for p in paths.keys() if '/assets' in p]
    print(f"📂 Assets endpoints found: {len(assets_paths)}")
    if assets_paths:
        for path in sorted(assets_paths):
            methods = list(paths[path].keys())
            print(f"   ✓ {path} ({', '.join(m.upper() for m in methods)})")
    else:
        print("   ❌ No Assets endpoints found!")
    print()
    
    # Check Settings  
    settings_paths = [p for p in paths.keys() if '/settings' in p]
    print(f"⚙️  Settings endpoints found: {len(settings_paths)}")
    if settings_paths:
        for path in sorted(settings_paths):
            methods = list(paths[path].keys())
            print(f"   ✓ {path} ({', '.join(m.upper() for m in methods)})")
    else:
        print("   ❌ No Settings endpoints found!")
    print()
    
    # Check if endpoints have proper tags
    print("🔍 Checking endpoint tags...")
    
    # Check Assets endpoints tags
    assets_without_tags = []
    for path in assets_paths:
        for method, method_data in paths[path].items():
            if isinstance(method_data, dict):
                method_tags = method_data.get('tags', [])
                if 'Assets' not in method_tags:
                    assets_without_tags.append(f"{method.upper()} {path}")
    
    if assets_without_tags:
        print("❌ Assets endpoints missing 'Assets' tag:")
        for endpoint in assets_without_tags:
            print(f"   - {endpoint}")
    else:
        print("✅ All Assets endpoints have proper 'Assets' tag")
    
    # Check Settings endpoints tags  
    settings_without_tags = []
    for path in settings_paths:
        for method, method_data in paths[path].items():
            if isinstance(method_data, dict):
                method_tags = method_data.get('tags', [])
                if 'Settings' not in method_tags:
                    settings_without_tags.append(f"{method.upper()} {path}")
    
    if settings_without_tags:
        print("❌ Settings endpoints missing 'Settings' tag:")
        for endpoint in settings_without_tags:
            print(f"   - {endpoint}")
    else:
        print("✅ All Settings endpoints have proper 'Settings' tag")
    
    print()
    
    # Summary
    assets_ok = 'Assets' in tag_names and len(assets_paths) > 0 and len(assets_without_tags) == 0
    settings_ok = 'Settings' in tag_names and len(settings_paths) > 0 and len(settings_without_tags) == 0
    
    print("📋 VERIFICATION SUMMARY")
    print("=" * 25)
    print(f"Assets Module:   {'✅ OK' if assets_ok else '❌ ISSUES'}")
    print(f"Settings Module: {'✅ OK' if settings_ok else '❌ ISSUES'}")
    print()
    
    if assets_ok and settings_ok:
        print("🎉 SUCCESS! Both Assets and Settings are properly configured!")
        print()
        print("📍 Expected Swagger UI sections:")
        print("   📂 Assets - Fixed assets and depreciation management")
        print("   ⚙️ Settings - System settings and configuration")
        print()
        print("🌐 Access at: http://localhost:8080/swagger")
        return True
    else:
        print("❌ ISSUES FOUND! Assets and/or Settings may not display properly.")
        return False

def main():
    """Main entry point"""
    print("🚀 Swagger Assets & Settings Verifier")
    print()
    
    success = verify_swagger()
    
    if not success:
        print("\n💡 Troubleshooting tips:")
        print("   1. Run fix_swagger_display.py again")
        print("   2. Check if backend server is using the updated swagger.json")
        print("   3. Restart backend server")  
        print("   4. Clear browser cache")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())