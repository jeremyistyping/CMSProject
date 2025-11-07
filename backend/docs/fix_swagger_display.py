#!/usr/bin/env python3
"""
Script untuk memastikan Assets dan Settings muncul di Swagger UI dengan benar
dengan menambahkan tags dan memastikan struktur yang tepat
"""

import json
import os
import shutil
from datetime import datetime

def backup_file(file_path):
    """Create a backup of the original file with timestamp"""
    if os.path.exists(file_path):
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = f"{file_path}.backup_fix_display_{timestamp}"
        shutil.copy2(file_path, backup_path)
        print(f"✅ Backup created: {backup_path}")
        return backup_path
    return None

def fix_swagger_display():
    """Fix Swagger display to show Assets and Settings properly"""
    
    base_dir = os.path.dirname(os.path.abspath(__file__))
    swagger_file = os.path.join(base_dir, 'swagger.json')
    
    print("🔄 Fixing Swagger display for Assets and Settings...")
    print(f"📄 Swagger file: {swagger_file}")
    
    # Create backup
    backup_path = backup_file(swagger_file)
    
    # Load swagger.json
    try:
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = json.load(f)
    except Exception as e:
        print(f"❌ Error loading swagger.json: {e}")
        return False
    
    print(f"📊 Total endpoints found: {len(swagger_data.get('paths', {}))}")
    
    # Check if paths exist
    paths = swagger_data.get('paths', {})
    assets_paths = [p for p in paths.keys() if '/assets' in p]
    settings_paths = [p for p in paths.keys() if '/settings' in p]
    
    print(f"🔍 Assets paths found: {len(assets_paths)}")
    for path in assets_paths[:5]:  # Show first 5
        print(f"   - {path}")
    
    print(f"🔍 Settings paths found: {len(settings_paths)}")  
    for path in settings_paths[:5]:  # Show first 5
        print(f"   - {path}")
    
    # Ensure tags section exists
    if 'tags' not in swagger_data:
        swagger_data['tags'] = []
        print("✅ Added missing 'tags' section")
    
    # Get existing tag names
    existing_tags = {tag.get('name', '') for tag in swagger_data['tags'] if isinstance(tag, dict)}
    print(f"🏷️  Existing tags: {list(existing_tags)}")
    
    # Add Assets tag if missing
    if 'Assets' not in existing_tags:
        assets_tag = {
            "name": "Assets",
            "description": "Fixed assets and depreciation management - Manajemen aset tetap dan depresiasi"
        }
        swagger_data['tags'].append(assets_tag)
        print("✅ Added 'Assets' tag")
    else:
        print("ℹ️  'Assets' tag already exists")
    
    # Add Settings tag if missing
    if 'Settings' not in existing_tags:
        settings_tag = {
            "name": "Settings", 
            "description": "System settings and configuration - Pengaturan sistem dan konfigurasi"
        }
        swagger_data['tags'].append(settings_tag)
        print("✅ Added 'Settings' tag")
    else:
        print("ℹ️  'Settings' tag already exists")
    
    # Check and fix tags in paths
    assets_endpoints_fixed = 0
    settings_endpoints_fixed = 0
    
    for path, methods in paths.items():
        if '/assets' in path:
            for method, method_data in methods.items():
                if isinstance(method_data, dict):
                    # Ensure tags exist for Assets endpoints
                    if 'tags' not in method_data:
                        method_data['tags'] = ['Assets']
                        assets_endpoints_fixed += 1
                    elif 'Assets' not in method_data.get('tags', []):
                        if not method_data['tags']:
                            method_data['tags'] = ['Assets']
                        else:
                            method_data['tags'].append('Assets')
                        assets_endpoints_fixed += 1
        
        elif '/settings' in path:
            for method, method_data in methods.items():
                if isinstance(method_data, dict):
                    # Ensure tags exist for Settings endpoints
                    if 'tags' not in method_data:
                        method_data['tags'] = ['Settings']
                        settings_endpoints_fixed += 1
                    elif 'Settings' not in method_data.get('tags', []):
                        if not method_data['tags']:
                            method_data['tags'] = ['Settings']
                        else:
                            method_data['tags'].append('Settings')
                        settings_endpoints_fixed += 1
    
    print(f"🔧 Fixed tags for {assets_endpoints_fixed} Assets endpoints")
    print(f"🔧 Fixed tags for {settings_endpoints_fixed} Settings endpoints")
    
    # Add comprehensive Assets endpoints if missing
    assets_endpoints_to_add = {
        "/api/v1/assets": {
            "get": {
                "tags": ["Assets"],
                "summary": "Get all assets", 
                "description": "Retrieve list of all assets with filtering and pagination",
                "parameters": [
                    {
                        "name": "search",
                        "in": "query",
                        "description": "Search by asset name or code",
                        "type": "string"
                    },
                    {
                        "name": "category_id", 
                        "in": "query",
                        "description": "Filter by category ID",
                        "type": "integer"
                    },
                    {
                        "name": "status",
                        "in": "query", 
                        "description": "Filter by asset status",
                        "type": "string"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string", "example": "success"},
                                "data": {
                                    "type": "array",
                                    "items": {"type": "object"}
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "post": {
                "tags": ["Assets"],
                "summary": "Create new asset",
                "description": "Create a new fixed asset",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    {
                        "name": "asset",
                        "in": "body",
                        "description": "Asset creation data",
                        "required": True,
                        "schema": {
                            "type": "object",
                            "required": ["name", "code", "acquisition_cost"],
                            "properties": {
                                "name": {"type": "string", "example": "Office Computer"},
                                "code": {"type": "string", "example": "AST-001"},
                                "acquisition_cost": {"type": "number", "example": 5000000}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Asset created successfully",
                        "schema": {"type": "object"}
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/assets/{id}": {
            "get": {
                "tags": ["Assets"],
                "summary": "Get asset by ID",
                "description": "Get asset details by ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Asset ID", 
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {"type": "object"}
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "put": {
                "tags": ["Assets"],
                "summary": "Update asset",
                "description": "Update asset details",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Asset ID",
                        "required": True, 
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Asset updated successfully",
                        "schema": {"type": "object"}
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "delete": {
                "tags": ["Assets"],
                "summary": "Delete asset", 
                "description": "Delete asset",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Asset ID",
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Asset deleted successfully",
                        "schema": {"type": "object"}
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        }
    }
    
    # Add Settings endpoints
    settings_endpoints_to_add = {
        "/api/v1/settings": {
            "get": {
                "tags": ["Settings"],
                "summary": "Get system settings",
                "description": "Get current system settings and configuration",
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string", "example": "success"},
                                "data": {"type": "object"}
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "put": {
                "tags": ["Settings"],
                "summary": "Update system settings",
                "description": "Update system settings and configuration",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    {
                        "name": "settings",
                        "in": "body",
                        "description": "Settings data",
                        "required": True,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "company_name": {"type": "string"},
                                "currency": {"type": "string"},
                                "tax_rate": {"type": "number"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Settings updated successfully",
                        "schema": {"type": "object"}
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        }
    }
    
    # Add missing endpoints
    added_count = 0
    for path, methods in assets_endpoints_to_add.items():
        if path not in paths:
            paths[path] = methods
            added_count += len(methods)
            print(f"✅ Added missing path: {path}")
        else:
            # Check individual methods
            for method, method_data in methods.items():
                if method not in paths[path]:
                    paths[path][method] = method_data
                    added_count += 1
                    print(f"✅ Added missing method: {method.upper()} {path}")
    
    for path, methods in settings_endpoints_to_add.items():
        if path not in paths:
            paths[path] = methods
            added_count += len(methods)
            print(f"✅ Added missing path: {path}")
        else:
            # Check individual methods
            for method, method_data in methods.items():
                if method not in paths[path]:
                    paths[path][method] = method_data
                    added_count += 1
                    print(f"✅ Added missing method: {method.upper()} {path}")
    
    print(f"📊 Total new methods added: {added_count}")
    
    # Save the updated swagger.json
    try:
        with open(swagger_file, 'w', encoding='utf-8') as f:
            json.dump(swagger_data, f, indent=2, ensure_ascii=False)
        print("✅ Swagger file updated successfully!")
        
        # Create summary
        print(f"\n🎉 Summary:")
        print(f"   - Total endpoints: {len(swagger_data.get('paths', {}))}")
        print(f"   - Assets endpoints: {len([p for p in swagger_data['paths'].keys() if '/assets' in p])}")
        print(f"   - Settings endpoints: {len([p for p in swagger_data['paths'].keys() if '/settings' in p])}")
        print(f"   - Total tags: {len(swagger_data.get('tags', []))}")
        
        return True
    except Exception as e:
        print(f"❌ Error saving swagger.json: {e}")
        return False

def main():
    """Main entry point"""
    print("🚀 Swagger Display Fixer for Assets & Settings")
    print("=" * 50)
    
    try:
        success = fix_swagger_display()
        if success:
            print("\n🎉 Assets and Settings should now appear in Swagger UI!")
            print("\n💡 Next steps:")
            print("   1. Restart your backend server if it's running")
            print("   2. Clear browser cache or refresh Swagger UI")
            print("   3. Navigate to: http://localhost:8080/swagger")
            print("   4. Look for 'Assets' and 'Settings' sections")
            print("\n🔍 If still not showing:")
            print("   - Check that the backend is serving from the updated swagger.json")
            print("   - Verify the swagger generation process in your Go application")
            print("   - Check console for any JavaScript errors in Swagger UI")
        else:
            print("\n❌ Failed to fix Swagger display")
            return 1
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())