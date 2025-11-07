#!/usr/bin/env python3
"""
Script untuk memverifikasi Assets & Settings muncul setelah perbaikan whitelist
"""

import json
import urllib.request
import urllib.error
import time

def test_swagger_endpoints():
    """Test both swagger endpoints to verify Assets and Settings are now included"""
    print("🧪 TESTING SWAGGER ENDPOINTS AFTER WHITELIST FIX")
    print("=" * 55)
    
    endpoints = {
        "Enhanced (Filtered)": "http://localhost:8080/openapi/enhanced-doc.json",
        "Complete (Unfiltered)": "http://localhost:8080/openapi/doc.json"
    }
    
    results = {}
    
    for name, url in endpoints.items():
        try:
            print(f"\n🔗 Testing {name}: {url}")
            
            with urllib.request.urlopen(url, timeout=10) as response:
                data = json.loads(response.read().decode('utf-8'))
                
                paths = data.get('paths', {})
                total_paths = len(paths)
                
                # Count Assets endpoints
                assets_paths = [p for p in paths.keys() if '/assets' in p]
                assets_count = len(assets_paths)
                
                # Count Settings endpoints
                settings_paths = [p for p in paths.keys() if '/settings' in p]
                settings_count = len(settings_paths)
                
                results[name] = {
                    'total': total_paths,
                    'assets': assets_count,
                    'settings': settings_count,
                    'assets_paths': assets_paths,
                    'settings_paths': settings_paths
                }
                
                print(f"   ✅ SUCCESS")
                print(f"   📊 Total endpoints: {total_paths}")
                print(f"   📂 Assets endpoints: {assets_count}")
                print(f"   ⚙️  Settings endpoints: {settings_count}")
                
                if assets_count > 0:
                    print(f"   📂 Assets paths found:")
                    for path in sorted(assets_paths):
                        print(f"      - {path}")
                
                if settings_count > 0:
                    print(f"   ⚙️  Settings paths found:")
                    for path in sorted(settings_paths):
                        print(f"      - {path}")
                
        except urllib.error.URLError as e:
            print(f"   ❌ Connection error: {e.reason}")
            results[name] = None
        except Exception as e:
            print(f"   ❌ Error: {e}")
            results[name] = None
    
    return results

def analyze_results(results):
    """Analyze the test results and provide conclusions"""
    print(f"\n📊 ANALYSIS & RESULTS")
    print("=" * 25)
    
    enhanced = results.get("Enhanced (Filtered)")
    complete = results.get("Complete (Unfiltered)")
    
    if not enhanced or not complete:
        print("❌ Cannot analyze - some endpoints failed to respond")
        return False
    
    print(f"Complete spec: {complete['total']} endpoints")
    print(f"Enhanced spec: {enhanced['total']} endpoints")
    
    # Check if Assets are now included in enhanced spec
    if enhanced['assets'] > 0:
        print(f"✅ Assets in Enhanced spec: {enhanced['assets']} endpoints")
        if enhanced['assets'] == complete['assets']:
            print("   ✅ All Assets endpoints included!")
        else:
            print(f"   ⚠️  Missing some Assets endpoints ({complete['assets'] - enhanced['assets']} missing)")
    else:
        print(f"❌ Assets STILL MISSING from Enhanced spec")
        return False
    
    # Check if Settings are now included in enhanced spec
    if enhanced['settings'] > 0:
        print(f"✅ Settings in Enhanced spec: {enhanced['settings']} endpoints")
        if enhanced['settings'] == complete['settings']:
            print("   ✅ All Settings endpoints included!")
        else:
            print(f"   ⚠️  Missing some Settings endpoints ({complete['settings'] - enhanced['settings']} missing)")
    else:
        print(f"❌ Settings STILL MISSING from Enhanced spec")
        return False
    
    print(f"\n🎉 SUCCESS! Both Assets and Settings are now included in Swagger UI!")
    print(f"   📂 Assets: {enhanced['assets']} endpoints")
    print(f"   ⚙️  Settings: {enhanced['settings']} endpoints")
    
    return True

def wait_for_server_restart():
    """Wait for server to restart and be ready"""
    print("\n⏳ Waiting for server to restart...")
    
    max_attempts = 30
    attempt = 0
    
    while attempt < max_attempts:
        try:
            with urllib.request.urlopen("http://localhost:8080/api/v1/health", timeout=2):
                print("✅ Server is ready!")
                return True
        except:
            pass
        
        attempt += 1
        time.sleep(1)
        print(f"   Attempt {attempt}/{max_attempts}...")
    
    print("❌ Server did not respond after 30 seconds")
    return False

def main():
    """Main test function"""
    print("🔬 ASSETS & SETTINGS WHITELIST FIX TESTER")
    print("=" * 45)
    
    print("\n💡 Instructions:")
    print("1. Restart your Go backend server to reload the whitelist changes")
    print("2. This script will test both endpoints")
    print("3. Assets and Settings should now appear in BOTH endpoints")
    print("\nPress Enter when server is restarted...")
    input()
    
    # Test endpoints
    results = test_swagger_endpoints()
    
    # Analyze results
    success = analyze_results(results)
    
    if success:
        print(f"\n🌐 Next steps:")
        print(f"1. Open Swagger UI: http://localhost:8080/swagger")
        print(f"2. Clear browser cache (Ctrl+Shift+Delete)")
        print(f"3. You should now see Assets and Settings sections!")
        return 0
    else:
        print(f"\n❌ Assets and/or Settings are still missing.")
        print(f"   Server might need restart or there could be other issues.")
        return 1

if __name__ == "__main__":
    exit(main())