#!/usr/bin/env python3
"""
Script untuk mendiagnosis mengapa Assets dan Settings tidak muncul di Swagger UI
meskipun sudah ada di swagger.json
"""

import json
import os
import subprocess
import urllib.request
import socket

def check_file_integrity():
    """Check swagger.json file integrity and structure"""
    print("🔍 CHECKING FILE INTEGRITY")
    print("=" * 30)
    
    base_dir = os.path.dirname(os.path.abspath(__file__))
    swagger_file = os.path.join(base_dir, 'swagger.json')
    
    try:
        # Check file size
        file_size = os.path.getsize(swagger_file)
        print(f"📄 File: {swagger_file}")
        print(f"📊 Size: {file_size:,} bytes")
        
        # Try to parse JSON
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = json.load(f)
        
        # Check structure
        required_keys = ['swagger', 'info', 'paths', 'tags']
        for key in required_keys:
            if key in swagger_data:
                print(f"✅ {key}: Found")
            else:
                print(f"❌ {key}: Missing!")
                
        # Check Assets and Settings specifically
        paths = swagger_data.get('paths', {})
        tags = swagger_data.get('tags', [])
        
        assets_count = len([p for p in paths.keys() if '/assets' in p])
        settings_count = len([p for p in paths.keys() if '/settings' in p])
        
        tag_names = [tag.get('name') for tag in tags if isinstance(tag, dict)]
        
        print(f"📂 Assets endpoints: {assets_count}")
        print(f"⚙️  Settings endpoints: {settings_count}")
        print(f"🏷️  Assets tag: {'✅' if 'Assets' in tag_names else '❌'}")
        print(f"🏷️  Settings tag: {'✅' if 'Settings' in tag_names else '❌'}")
        
        return True, swagger_data
        
    except json.JSONDecodeError as e:
        print(f"❌ JSON Parse Error: {e}")
        return False, None
    except Exception as e:
        print(f"❌ File Error: {e}")
        return False, None

def check_server_response():
    """Check if backend server is serving the updated swagger.json"""
    print("\n🌐 CHECKING SERVER RESPONSE")
    print("=" * 30)
    
    swagger_urls = [
        'http://localhost:8080/swagger.json',
        'http://localhost:8080/docs/swagger.json',
        'http://localhost:8080/api/swagger.json',
        'http://localhost:3000/swagger.json',  # Alternative port
    ]
    
    for url in swagger_urls:
        try:
            print(f"🔗 Testing: {url}")
            
            # Check if port is open
            host, port = url.split('://')[1].split('/')[0].split(':')
            port = int(port)
            
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(2)
            result = sock.connect_ex((host, port))
            sock.close()
            
            if result != 0:
                print(f"   ❌ Port {port} not accessible")
                continue
                
            # Try to fetch swagger.json from server
            with urllib.request.urlopen(url, timeout=5) as response:
                server_data = json.loads(response.read().decode('utf-8'))
                
                # Check if server version matches local file
                server_paths = server_data.get('paths', {})
                server_assets = len([p for p in server_paths.keys() if '/assets' in p])
                server_settings = len([p for p in server_paths.keys() if '/settings' in p])
                
                print(f"   ✅ Server responding")
                print(f"   📂 Server Assets endpoints: {server_assets}")
                print(f"   ⚙️  Server Settings endpoints: {server_settings}")
                
                return True, server_data
                
        except urllib.error.URLError:
            print(f"   ❌ Cannot connect to {url}")
        except socket.timeout:
            print(f"   ❌ Timeout connecting to {url}")
        except Exception as e:
            print(f"   ❌ Error: {e}")
    
    print("❌ No server found serving swagger.json")
    return False, None

def check_swagger_ui_config():
    """Check common Swagger UI configuration issues"""
    print("\n⚙️  SWAGGER UI CONFIGURATION CHECK")
    print("=" * 35)
    
    # Look for common Swagger UI files
    possible_locations = [
        "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\static",
        "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\public",
        "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\assets",
        "D:\\Project\\clone_app_akuntansi\\accounting_proj\\frontend\\public",
        "D:\\Project\\clone_app_akuntansi\\accounting_proj\\docs",
    ]
    
    ui_files_found = []
    for location in possible_locations:
        if os.path.exists(location):
            for root, dirs, files in os.walk(location):
                for file in files:
                    if 'swagger' in file.lower():
                        ui_files_found.append(os.path.join(root, file))
    
    if ui_files_found:
        print("🔍 Swagger UI related files found:")
        for file in ui_files_found:
            print(f"   📄 {file}")
    else:
        print("❌ No Swagger UI files found in common locations")
    
    # Check for Go files that might configure Swagger
    backend_dir = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend"
    go_files = []
    
    if os.path.exists(backend_dir):
        for root, dirs, files in os.walk(backend_dir):
            for file in files:
                if file.endswith('.go') and ('swagger' in file.lower() or 'docs' in file.lower()):
                    go_files.append(os.path.join(root, file))
    
    if go_files:
        print("\n🔍 Go files that might configure Swagger:")
        for file in go_files:
            print(f"   📄 {file}")
    
    return ui_files_found, go_files

def generate_troubleshooting_steps():
    """Generate specific troubleshooting steps based on findings"""
    print("\n🛠️  TROUBLESHOOTING STEPS")
    print("=" * 25)
    
    print("1. 🔄 RESTART BACKEND SERVER")
    print("   - Stop your Go backend server")
    print("   - Start it again to reload swagger.json")
    print("   - Command: go run main.go (or your server command)")
    
    print("\n2. 🧹 CLEAR BROWSER CACHE")
    print("   - Press Ctrl+Shift+Delete in your browser")
    print("   - Clear all browsing data")
    print("   - Or try opening Swagger in incognito/private mode")
    
    print("\n3. 🌐 VERIFY SWAGGER URL")
    print("   - Try these URLs:")
    print("     • http://localhost:8080/swagger")
    print("     • http://localhost:8080/docs")
    print("     • http://localhost:8080/swagger-ui")
    print("     • http://localhost:8080/api-docs")
    
    print("\n4. 🔍 CHECK BROWSER CONSOLE")
    print("   - Open F12 Developer Tools")
    print("   - Check Console tab for errors")
    print("   - Check Network tab for failed requests")
    
    print("\n5. 📄 VERIFY SWAGGER.JSON URL")
    print("   - Try accessing swagger.json directly:")
    print("     • http://localhost:8080/swagger.json")
    print("     • http://localhost:8080/docs/swagger.json")
    
    print("\n6. 🔧 CHECK BACKEND CONFIGURATION")
    print("   - Ensure your Go app serves swagger.json from docs/swagger.json")
    print("   - Check if Swagger middleware is properly configured")
    print("   - Verify swagger-ui static files are accessible")

def main():
    """Main diagnostic function"""
    print("🩺 SWAGGER UI DIAGNOSTIC TOOL")
    print("=" * 30)
    print()
    
    # Step 1: Check local file
    file_ok, local_data = check_file_integrity()
    
    if not file_ok:
        print("\n❌ CRITICAL: swagger.json file has issues!")
        print("Fix the file before proceeding.")
        return 1
    
    # Step 2: Check server
    server_ok, server_data = check_server_response()
    
    # Step 3: Check UI config
    ui_files, go_files = check_swagger_ui_config()
    
    # Step 4: Generate troubleshooting
    generate_troubleshooting_steps()
    
    print("\n" + "=" * 60)
    print("📊 DIAGNOSTIC SUMMARY")
    print("=" * 60)
    
    print(f"Local swagger.json: {'✅ OK' if file_ok else '❌ ISSUES'}")
    print(f"Server response:    {'✅ OK' if server_ok else '❌ NO SERVER'}")
    print(f"UI files found:     {'✅ YES' if ui_files else '❌ NONE'}")
    
    if file_ok and not server_ok:
        print("\n💡 LIKELY ISSUE: Backend server not running or not serving swagger.json")
        print("   → Start/restart your backend server")
        
    elif file_ok and server_ok:
        print("\n💡 LIKELY ISSUE: Browser cache or Swagger UI configuration")
        print("   → Clear browser cache and restart server")
        
    else:
        print("\n💡 MULTIPLE ISSUES: Check all troubleshooting steps above")
    
    return 0

if __name__ == "__main__":
    exit(main())