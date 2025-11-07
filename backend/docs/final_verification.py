#!/usr/bin/env python3
"""
Script final untuk memverifikasi Assets & Settings sudah benar di Swagger UI
"""

import json
import urllib.request
import urllib.error
import time

def check_all_swagger_endpoints():
    """Check all swagger endpoints to verify the fixes"""
    print("🔍 FINAL VERIFICATION - CHECKING ALL SWAGGER ENDPOINTS")
    print("=" * 60)
    
    endpoints = {
        "Enhanced (Filtered - Swagger UI uses this)": "http://localhost:8080/openapi/enhanced-doc.json",
        "Complete (Unfiltered)": "http://localhost:8080/openapi/doc.json"
    }
    
    results = {}
    
    for name, url in endpoints.items():
        try:
            print(f"\n🔗 Testing {name}:")
            print(f"   URL: {url}")
            
            with urllib.request.urlopen(url, timeout=10) as response:
                data = json.loads(response.read().decode('utf-8'))
                
                paths = data.get('paths', {})
                total_paths = len(paths)
                
                # Count Assets endpoints
                assets_paths = [p for p in paths.keys() if '/assets' in p]
                assets_count = len(assets_paths)
                
                # Count Settings endpoints (both /settings and /tax-accounts)
                settings_paths = [p for p in paths.keys() if '/settings' in p or '/tax-accounts' in p]
                settings_count = len(settings_paths)
                
                # Check for the problematic endpoint
                has_accounting_endpoint = "/api/v1/settings/accounting" in paths
                
                results[name] = {
                    'total': total_paths,
                    'assets': assets_count,
                    'settings': settings_count,
                    'assets_paths': assets_paths,
                    'settings_paths': settings_paths,
                    'has_invalid_accounting': has_accounting_endpoint
                }
                
                print(f"   ✅ SUCCESS")
                print(f"   📊 Total endpoints: {total_paths}")
                print(f"   📂 Assets endpoints: {assets_count}")
                print(f"   ⚙️  Settings endpoints: {settings_count}")
                print(f"   🗑️  Invalid /accounting endpoint: {'❌ STILL PRESENT' if has_accounting_endpoint else '✅ REMOVED'}")
                
        except urllib.error.URLError as e:
            print(f"   ❌ Connection error: {e.reason}")
            results[name] = None
        except Exception as e:
            print(f"   ❌ Error: {e}")
            results[name] = None
    
    return results

def analyze_final_results(results):
    """Analyze final results and provide summary"""
    print(f"\n📊 FINAL ANALYSIS & SUMMARY")
    print("=" * 30)
    
    enhanced = results.get("Enhanced (Filtered - Swagger UI uses this)")
    complete = results.get("Complete (Unfiltered)")
    
    if not enhanced or not complete:
        print("❌ Cannot analyze - some endpoints failed to respond")
        return False
    
    issues = []
    successes = []
    
    # Check Assets
    if enhanced['assets'] > 0:
        successes.append(f"✅ Assets: {enhanced['assets']} endpoints in Swagger UI")
        if enhanced['assets'] != complete['assets']:
            issues.append(f"⚠️  Assets: Missing {complete['assets'] - enhanced['assets']} endpoints from complete spec")
    else:
        issues.append("❌ Assets: Still missing from Swagger UI")
    
    # Check Settings
    if enhanced['settings'] > 0:
        successes.append(f"✅ Settings: {enhanced['settings']} endpoints in Swagger UI")
        if enhanced['settings'] != complete['settings']:
            issues.append(f"⚠️  Settings: Missing {complete['settings'] - enhanced['settings']} endpoints from complete spec")
    else:
        issues.append("❌ Settings: Still missing from Swagger UI")
    
    # Check invalid endpoint removal
    if enhanced['has_invalid_accounting']:
        issues.append("❌ Invalid /settings/accounting endpoint still present")
    else:
        successes.append("✅ Invalid /settings/accounting endpoint removed")
    
    print("🎉 SUCCESSES:")
    for success in successes:
        print(f"   {success}")
    
    if issues:
        print("\n⚠️  REMAINING ISSUES:")
        for issue in issues:
            print(f"   {issue}")
    
    # Overall status
    critical_success = (
        enhanced['assets'] > 0 and 
        enhanced['settings'] > 0 and 
        not enhanced['has_invalid_accounting']
    )
    
    print(f"\n{'🎉 OVERALL: SUCCESS!' if critical_success else '❌ OVERALL: ISSUES REMAIN'}")
    
    return critical_success

def show_swagger_ui_instructions():
    """Show instructions for accessing Swagger UI"""
    print(f"\n🌐 HOW TO ACCESS SWAGGER UI")
    print("=" * 25)
    print("1. Make sure your Go backend server is running")
    print("2. Clear your browser cache (Ctrl+Shift+Delete)")
    print("3. Open one of these URLs:")
    print("   • http://localhost:8080/swagger")
    print("   • http://localhost:8080/docs")
    print("   • http://localhost:8080/swagger/index.html")
    print("\n📂 You should now see these sections in Swagger UI:")
    print("   • Assets - Fixed assets and depreciation management")
    print("   • Settings - System settings and configuration") 
    print("   • Tax Accounts - Tax account configuration")
    print("   • And all other modules (Accounts, Products, Contacts, Users)")

def main():
    """Main verification function"""
    print("🔬 FINAL SWAGGER VERIFICATION TOOL")
    print("=" * 35)
    
    print("\n💡 This tool verifies all fixes:")
    print("   1. Assets endpoints are in Swagger UI")
    print("   2. Settings endpoints are in Swagger UI") 
    print("   3. Invalid endpoints have been removed")
    print("   4. Whitelist has been updated correctly")
    
    print("\n⚠️  Make sure you have RESTARTED your Go backend server first!")
    input("Press Enter when ready to test...")
    
    # Check all endpoints
    results = check_all_swagger_endpoints()
    
    # Analyze results
    success = analyze_final_results(results)
    
    # Show instructions
    show_swagger_ui_instructions()
    
    if success:
        print(f"\n🎉 ALL FIXES VERIFIED SUCCESSFULLY!")
        print(f"   Assets and Settings should now appear in Swagger UI")
        print(f"   The 404 error for /settings/accounting is fixed")
        return 0
    else:
        print(f"\n❌ Some issues remain - check analysis above")
        return 1

if __name__ == "__main__":
    exit(main())