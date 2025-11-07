#!/usr/bin/env python3
"""
Script untuk mengecek endpoint Swagger yang benar dan memastikan Assets & Settings muncul
"""

import json
import os
import urllib.request
import urllib.error

def check_server_swagger_endpoint():
    """Check which endpoint the server is actually serving for swagger spec"""
    print("🔍 CHECKING SWAGGER ENDPOINT")
    print("=" * 30)
    
    # Test endpoints that the server might be serving
    test_urls = [
        'http://localhost:8080/openapi/enhanced-doc.json',  # From server logs
        'http://localhost:8080/openapi/doc.json', 
        'http://localhost:8080/swagger.json',
        'http://localhost:8080/docs/swagger.json',
    ]
    
    working_endpoint = None
    server_data = None
    
    for url in test_urls:
        try:
            print(f"🔗 Testing: {url}")
            with urllib.request.urlopen(url, timeout=5) as response:
                data = json.loads(response.read().decode('utf-8'))
                
                # Check if this has our data
                paths = data.get('paths', {})
                assets_count = len([p for p in paths.keys() if '/assets' in p])
                settings_count = len([p for p in paths.keys() if '/settings' in p])
                total_paths = len(paths)
                
                print(f"   ✅ SUCCESS - {total_paths} endpoints found")
                print(f"   📂 Assets: {assets_count} endpoints")
                print(f"   ⚙️  Settings: {settings_count} endpoints")
                
                if assets_count == 0 or settings_count == 0:
                    print(f"   ❌ Missing Assets ({assets_count}) or Settings ({settings_count}) endpoints")
                    working_endpoint = url
                    server_data = data
                else:
                    print(f"   ✅ Both Assets and Settings found!")
                    return url, data
                
        except urllib.error.HTTPError as e:
            print(f"   ❌ HTTP {e.code}: {e.reason}")
        except urllib.error.URLError as e:
            print(f"   ❌ Connection error: {e.reason}")
        except json.JSONDecodeError:
            print(f"   ❌ Invalid JSON response")
        except Exception as e:
            print(f"   ❌ Error: {e}")
    
    return working_endpoint, server_data

def find_local_swagger_files():
    """Find all possible swagger/openapi files"""
    print("\n📂 FINDING LOCAL SWAGGER FILES")
    print("=" * 30)
    
    base_dir = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend"
    swagger_files = []
    
    # Look for swagger-related files
    for root, dirs, files in os.walk(base_dir):
        for file in files:
            if any(keyword in file.lower() for keyword in ['swagger', 'openapi', 'doc']):
                if file.endswith(('.json', '.yaml', '.yml')):
                    full_path = os.path.join(root, file)
                    swagger_files.append(full_path)
    
    print("Found swagger-related files:")
    for file in swagger_files:
        try:
            size = os.path.getsize(file)
            print(f"   📄 {file} ({size:,} bytes)")
        except:
            print(f"   📄 {file} (error reading)")
    
    return swagger_files

def update_server_swagger_file(server_endpoint, local_swagger_data):
    """Update the file that server is actually serving"""
    print("\n🔧 UPDATING SERVER SWAGGER FILE")
    print("=" * 35)
    
    # Try to determine which local file corresponds to server endpoint
    base_dir = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend"
    
    possible_server_files = [
        os.path.join(base_dir, "docs", "enhanced-doc.json"),
        os.path.join(base_dir, "docs", "doc.json"),
        os.path.join(base_dir, "openapi", "enhanced-doc.json"),
        os.path.join(base_dir, "openapi", "doc.json"),
        os.path.join(base_dir, "static", "enhanced-doc.json"),
    ]
    
    # Check which file exists and might be served
    for file_path in possible_server_files:
        if os.path.exists(file_path):
            print(f"📄 Found potential server file: {file_path}")
            
            try:
                # Read current content
                with open(file_path, 'r', encoding='utf-8') as f:
                    current_data = json.load(f)
                
                current_paths = current_data.get('paths', {})
                current_assets = len([p for p in current_paths.keys() if '/assets' in p])
                current_settings = len([p for p in current_paths.keys() if '/settings' in p])
                
                print(f"   📊 Current: {len(current_paths)} endpoints")
                print(f"   📂 Current Assets: {current_assets}")
                print(f"   ⚙️  Current Settings: {current_settings}")
                
                if current_assets == 0 or current_settings == 0:
                    print(f"   🔄 Updating this file with our swagger.json data...")
                    
                    # Create backup
                    backup_path = file_path + '.backup'
                    with open(backup_path, 'w', encoding='utf-8') as f:
                        json.dump(current_data, f, indent=2)
                    print(f"   💾 Backup created: {backup_path}")
                    
                    # Update with our data
                    with open(file_path, 'w', encoding='utf-8') as f:
                        json.dump(local_swagger_data, f, indent=2)
                    
                    print(f"   ✅ Updated {file_path}")
                    return True
                else:
                    print(f"   ✅ File already has Assets and Settings")
                    
            except Exception as e:
                print(f"   ❌ Error processing {file_path}: {e}")
    
    return False

def main():
    """Main function"""
    print("🚀 SWAGGER ENDPOINT FIXER")
    print("=" * 25)
    
    # Step 1: Check what server is serving
    server_endpoint, server_data = check_server_swagger_endpoint()
    
    if not server_endpoint:
        print("\n❌ Could not find working swagger endpoint from server")
        return 1
    
    print(f"\n✅ Server is using: {server_endpoint}")
    
    # Step 2: Load our updated swagger.json
    swagger_json_path = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\docs\\swagger.json"
    
    try:
        with open(swagger_json_path, 'r', encoding='utf-8') as f:
            local_swagger_data = json.load(f)
        
        local_paths = local_swagger_data.get('paths', {})
        local_assets = len([p for p in local_paths.keys() if '/assets' in p])
        local_settings = len([p for p in local_paths.keys() if '/settings' in p])
        
        print(f"\n📄 Local swagger.json has:")
        print(f"   📊 Total endpoints: {len(local_paths)}")
        print(f"   📂 Assets endpoints: {local_assets}")
        print(f"   ⚙️  Settings endpoints: {local_settings}")
        
    except Exception as e:
        print(f"\n❌ Error loading local swagger.json: {e}")
        return 1
    
    # Step 3: Compare with server data
    if server_data:
        server_paths = server_data.get('paths', {})
        server_assets = len([p for p in server_paths.keys() if '/assets' in p])
        server_settings = len([p for p in server_paths.keys() if '/settings' in p])
        
        print(f"\n🌐 Server swagger data has:")
        print(f"   📊 Total endpoints: {len(server_paths)}")  
        print(f"   📂 Assets endpoints: {server_assets}")
        print(f"   ⚙️  Settings endpoints: {server_settings}")
        
        if server_assets == 0 or server_settings == 0:
            print(f"\n⚠️  Server swagger is missing Assets or Settings!")
            print(f"   Need to update the server's swagger file")
        else:
            print(f"\n✅ Server already has Assets and Settings!")
            print(f"   The issue might be browser cache or Swagger UI configuration")
            return 0
    
    # Step 4: Find and update server files
    find_local_swagger_files()
    
    if server_data and (server_assets == 0 or server_settings == 0):
        success = update_server_swagger_file(server_endpoint, local_swagger_data)
        
        if success:
            print(f"\n🎉 SUCCESS!")
            print(f"   Updated server swagger file")
            print(f"   🔄 Please refresh your browser (Ctrl+F5)")
            print(f"   🌐 Check Swagger UI: http://localhost:8080/swagger")
        else:
            print(f"\n❌ Could not update server swagger file")
            print(f"   Manual action needed")
    
    return 0

if __name__ == "__main__":
    exit(main())