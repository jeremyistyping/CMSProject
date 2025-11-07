#!/usr/bin/env python3
"""
Script to merge additional Swagger API endpoints into the existing swagger.json file.
This script adds the missing CRUD endpoints for core modules: Accounts, Products, Assets, Contacts, Users, and Settings.
"""

import json
import os
import shutil
from datetime import datetime

def backup_file(file_path):
    """Create a backup of the original file with timestamp"""
    if os.path.exists(file_path):
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = f"{file_path}.backup_{timestamp}"
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

def merge_swagger_endpoints():
    """Main function to merge additional endpoints into swagger.json"""
    
    # File paths
    base_dir = os.path.dirname(os.path.abspath(__file__))
    swagger_file = os.path.join(base_dir, 'swagger.json')
    additional_file = os.path.join(base_dir, 'swagger_additional_endpoints.json')
    
    print("🔄 Starting Swagger endpoints merge process...")
    print(f"📂 Base directory: {base_dir}")
    print(f"📄 Main swagger file: {swagger_file}")
    print(f"📄 Additional endpoints file: {additional_file}")
    
    # Create backup of existing swagger.json
    backup_path = backup_file(swagger_file)
    
    # Load existing swagger.json
    swagger_data = load_json_file(swagger_file)
    if not swagger_data:
        print("❌ Failed to load existing swagger.json")
        return False
    
    # Load additional endpoints
    additional_data = load_json_file(additional_file)
    if not additional_data:
        print("❌ Failed to load additional endpoints file")
        return False
    
    # Validate structure
    if 'paths' not in swagger_data:
        swagger_data['paths'] = {}
    
    if 'paths' not in additional_data:
        print("❌ Additional endpoints file missing 'paths' section")
        return False
    
    # Count existing endpoints
    original_count = len(swagger_data['paths'])
    additional_count = len(additional_data['paths'])
    
    print(f"📊 Original endpoints: {original_count}")
    print(f"📊 Additional endpoints to merge: {additional_count}")
    
    # Merge additional paths
    conflicts = []
    merged_count = 0
    
    for path, methods in additional_data['paths'].items():
        if path in swagger_data['paths']:
            # Check for method conflicts
            existing_methods = set(swagger_data['paths'][path].keys())
            new_methods = set(methods.keys())
            conflicting_methods = existing_methods.intersection(new_methods)
            
            if conflicting_methods:
                conflicts.append(f"{path}: {', '.join(conflicting_methods)}")
                print(f"⚠️  Conflict detected for {path}: methods {', '.join(conflicting_methods)} already exist")
                # Merge non-conflicting methods only
                for method, method_data in methods.items():
                    if method not in existing_methods:
                        swagger_data['paths'][path][method] = method_data
                        merged_count += 1
            else:
                # No conflicts, merge all methods
                for method, method_data in methods.items():
                    swagger_data['paths'][path][method] = method_data
                    merged_count += 1
        else:
            # New path, add it completely
            swagger_data['paths'][path] = methods
            merged_count += len(methods)
    
    # Update swagger info if needed
    if 'info' in swagger_data:
        if 'description' in swagger_data['info']:
            if 'dengan fitur lengkap CRUD' not in swagger_data['info']['description']:
                swagger_data['info']['description'] = (
                    swagger_data['info']['description'] + 
                    " Termasuk endpoint CRUD lengkap untuk akun, produk, aset, kontak, user, dan pengaturan sistem."
                )
    
    # Add tags if not present
    if 'tags' not in swagger_data:
        swagger_data['tags'] = []
    
    # Define core module tags
    core_tags = [
        {
            "name": "Accounts",
            "description": "Chart of accounts management - Manajemen bagan akun (COA)"
        },
        {
            "name": "Products", 
            "description": "Product and inventory management - Manajemen produk dan inventori"
        },
        {
            "name": "Assets",
            "description": "Fixed assets and depreciation management - Manajemen aset tetap dan depresiasi"
        },
        {
            "name": "Contacts",
            "description": "Customer, vendor, and employee contact management - Manajemen kontak pelanggan, vendor, dan karyawan"
        },
        {
            "name": "Users",
            "description": "User account and role management - Manajemen pengguna dan peran"
        },
        {
            "name": "Settings",
            "description": "System settings and configuration - Pengaturan sistem dan konfigurasi"
        }
    ]
    
    # Add missing tags
    existing_tag_names = {tag.get('name') for tag in swagger_data['tags'] if isinstance(tag, dict)}
    for tag in core_tags:
        if tag['name'] not in existing_tag_names:
            swagger_data['tags'].append(tag)
    
    # Save the merged swagger.json
    if save_json_file(swagger_data, swagger_file):
        final_count = len(swagger_data['paths'])
        print(f"✅ Merge completed successfully!")
        print(f"📊 Final endpoint count: {final_count}")
        print(f"📊 Endpoints merged: {merged_count}")
        
        if conflicts:
            print(f"⚠️  {len(conflicts)} conflicts detected:")
            for conflict in conflicts:
                print(f"   - {conflict}")
        
        # Create summary report
        report_file = os.path.join(base_dir, 'swagger_merge_report.txt')
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(f"Swagger Endpoints Merge Report\n")
            f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
            f.write(f"Original endpoints: {original_count}\n")
            f.write(f"Additional endpoints: {additional_count}\n")
            f.write(f"Successfully merged: {merged_count}\n")
            f.write(f"Final total: {final_count}\n")
            f.write(f"Backup created: {backup_path}\n\n")
            
            if conflicts:
                f.write(f"Conflicts detected ({len(conflicts)}):\n")
                for conflict in conflicts:
                    f.write(f"  - {conflict}\n")
            else:
                f.write("No conflicts detected.\n")
            
            f.write("\nNew endpoints added:\n")
            for path in additional_data['paths']:
                f.write(f"  - {path}\n")
        
        print(f"📝 Merge report saved: {report_file}")
        return True
    else:
        print("❌ Failed to save merged swagger.json")
        return False

def main():
    """Main entry point"""
    print("🚀 Swagger API Documentation Merger")
    print("=" * 50)
    
    try:
        success = merge_swagger_endpoints()
        if success:
            print("\n🎉 All done! Your Swagger documentation now includes:")
            print("   ✓ Accounts (Akun) - Complete CRUD operations")
            print("   ✓ Products (Produk) - Complete CRUD operations") 
            print("   ✓ Assets (Aset) - Complete CRUD operations")
            print("   ✓ Contacts (Kontak) - Complete CRUD operations")
            print("   ✓ Users - Complete CRUD operations")
            print("   ✓ Settings - System configuration")
            print("\n💡 Next steps:")
            print("   1. Test the Swagger UI at: http://localhost:8080/swagger")
            print("   2. Verify all endpoints are working correctly")
            print("   3. Update any controller annotations if needed")
        else:
            print("\n❌ Merge process failed. Check the error messages above.")
            return 1
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())