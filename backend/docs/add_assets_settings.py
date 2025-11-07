#!/usr/bin/env python3
"""
Script khusus untuk menambahkan dokumentasi API Assets dan Settings ke swagger.json
"""

import json
import os
import shutil
from datetime import datetime

def backup_file(file_path):
    """Create a backup of the original file with timestamp"""
    if os.path.exists(file_path):
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = f"{file_path}.backup_assets_settings_{timestamp}"
        shutil.copy2(file_path, backup_path)
        print(f"✅ Backup created: {backup_path}")
        return backup_path
    return None

def load_json_file(file_path):
    """Load JSON file with error handling"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"❌ File not found: {file_path}")
        return None
    except json.JSONDecodeError as e:
        print(f"❌ JSON decode error in {file_path}: {e}")
        return None

def save_json_file(data, file_path):
    """Save JSON file with proper formatting"""
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"✅ File saved: {file_path}")
        return True
    except Exception as e:
        print(f"❌ Error saving {file_path}: {e}")
        return False

def add_assets_settings():
    """Main function to add Assets and Settings endpoints to swagger.json"""
    
    # File paths
    base_dir = os.path.dirname(os.path.abspath(__file__))
    swagger_file = os.path.join(base_dir, 'swagger.json')
    assets_settings_file = os.path.join(base_dir, 'assets_settings_endpoints.json')
    
    print("🔄 Adding Assets and Settings endpoints to Swagger...")
    print(f"📂 Base directory: {base_dir}")
    print(f"📄 Main swagger file: {swagger_file}")
    print(f"📄 Assets & Settings endpoints file: {assets_settings_file}")
    
    # Create backup of existing swagger.json
    backup_path = backup_file(swagger_file)
    
    # Load existing swagger.json
    swagger_data = load_json_file(swagger_file)
    if not swagger_data:
        print("❌ Failed to load existing swagger.json")
        return False
    
    # Load assets and settings endpoints
    assets_settings_data = load_json_file(assets_settings_file)
    if not assets_settings_data:
        print("❌ Failed to load assets and settings endpoints file")
        return False
    
    # Validate structure
    if 'paths' not in swagger_data:
        swagger_data['paths'] = {}
    
    if 'paths' not in assets_settings_data:
        print("❌ Assets and settings file missing 'paths' section")
        return False
    
    # Count existing endpoints
    original_count = len(swagger_data['paths'])
    additional_count = len(assets_settings_data['paths'])
    
    print(f"📊 Original endpoints: {original_count}")
    print(f"📊 Assets & Settings endpoints to add: {additional_count}")
    
    # Add new paths
    conflicts = []
    added_count = 0
    
    for path, methods in assets_settings_data['paths'].items():
        if path in swagger_data['paths']:
            # Check for method conflicts
            existing_methods = set(swagger_data['paths'][path].keys())
            new_methods = set(methods.keys())
            conflicting_methods = existing_methods.intersection(new_methods)
            
            if conflicting_methods:
                conflicts.append(f"{path}: {', '.join(conflicting_methods)}")
                print(f"⚠️  Conflict detected for {path}: methods {', '.join(conflicting_methods)} already exist")
                # Add non-conflicting methods only
                for method, method_data in methods.items():
                    if method not in existing_methods:
                        swagger_data['paths'][path][method] = method_data
                        added_count += 1
                        print(f"✅ Added {method.upper()} {path}")
            else:
                # No conflicts, add all methods
                for method, method_data in methods.items():
                    swagger_data['paths'][path][method] = method_data
                    added_count += 1
                    print(f"✅ Added {method.upper()} {path}")
        else:
            # New path, add completely
            swagger_data['paths'][path] = methods
            method_count = len(methods)
            added_count += method_count
            print(f"✅ Added new path: {path} ({method_count} methods)")
    
    # Add tags if not present
    if 'tags' not in swagger_data:
        swagger_data['tags'] = []
    
    # Add Assets and Settings tags
    if 'tags' in assets_settings_data:
        existing_tag_names = {tag.get('name') for tag in swagger_data['tags'] if isinstance(tag, dict)}
        for tag in assets_settings_data['tags']:
            if tag['name'] not in existing_tag_names:
                swagger_data['tags'].append(tag)
                print(f"✅ Added tag: {tag['name']}")
    
    # Save the updated swagger.json
    if save_json_file(swagger_data, swagger_file):
        final_count = len(swagger_data['paths'])
        print(f"✅ Successfully added Assets and Settings endpoints!")
        print(f"📊 Original endpoints: {original_count}")
        print(f"📊 Endpoints added: {added_count}")
        print(f"📊 Final total: {final_count}")
        
        if conflicts:
            print(f"⚠️  {len(conflicts)} conflicts detected:")
            for conflict in conflicts:
                print(f"   - {conflict}")
        
        # Create summary report
        report_file = os.path.join(base_dir, 'assets_settings_add_report.txt')
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(f"Assets & Settings Endpoints Addition Report\\n")
            f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\\n\\n")
            f.write(f"Original endpoints: {original_count}\\n")
            f.write(f"Assets & Settings endpoints: {additional_count}\\n")
            f.write(f"Successfully added: {added_count}\\n")
            f.write(f"Final total: {final_count}\\n")
            f.write(f"Backup created: {backup_path}\\n\\n")
            
            if conflicts:
                f.write(f"Conflicts detected ({len(conflicts)}):\\n")
                for conflict in conflicts:
                    f.write(f"  - {conflict}\\n")
            else:
                f.write("No conflicts detected.\\n")
            
            f.write("\\nEndpoints added:\\n")
            for path in assets_settings_data['paths']:
                f.write(f"  - {path}\\n")
        
        print(f"📝 Addition report saved: {report_file}")
        return True
    else:
        print("❌ Failed to save updated swagger.json")
        return False

def main():
    """Main entry point"""
    print("🚀 Assets & Settings API Documentation Adder")
    print("=" * 50)
    
    try:
        success = add_assets_settings()
        if success:
            print("\\n🎉 Assets and Settings endpoints successfully added to Swagger!")
            print("\\n✅ New endpoints available:")
            print("   📂 Assets (Aset):")
            print("      • GET /api/v1/assets - Get all assets")
            print("      • POST /api/v1/assets - Create new asset")  
            print("      • GET /api/v1/assets/{id} - Get asset by ID")
            print("      • PUT /api/v1/assets/{id} - Update asset")
            print("      • DELETE /api/v1/assets/{id} - Delete asset")
            print("      • GET /api/v1/assets/{id}/depreciation-schedule - Get depreciation schedule")
            print("      • POST /api/v1/assets/{id}/capitalize - Capitalize asset")
            print("      • GET /api/v1/assets/categories - Get asset categories")
            print("      • POST /api/v1/assets/categories - Create asset category")
            print("      • GET /api/v1/assets/summary - Get assets summary")
            print("\\n   ⚙️  Settings (Pengaturan):")
            print("      • GET /api/v1/settings - Get system settings")
            print("      • PUT /api/v1/settings - Update system settings")
            print("      • GET /api/v1/settings/company - Get company settings")
            print("      • PUT /api/v1/settings/company - Update company settings")
            print("      • GET /api/v1/settings/accounting - Get accounting settings")
            print("      • PUT /api/v1/settings/accounting - Update accounting settings")
            print("\\n💡 Next steps:")
            print("   1. Refresh Swagger UI at: http://localhost:8080/swagger")
            print("   2. Check the new Assets and Settings sections")
            print("   3. Test the endpoints with proper authentication")
        else:
            print("\\n❌ Failed to add endpoints. Check the error messages above.")
            return 1
    except Exception as e:
        print(f"\\n💥 Unexpected error: {e}")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())