#!/usr/bin/env python3
"""
Script untuk menambahkan Invoice Types management endpoints ke Settings section
"""

import json
import os
from datetime import datetime

def add_invoice_types_endpoints():
    """Add invoice types endpoints to swagger.json under Settings tag"""
    print("📄 ADDING INVOICE TYPES ENDPOINTS TO SETTINGS")
    print("=" * 50)
    
    swagger_file = "D:\\Project\\clone_app_akuntansi\\accounting_proj\\backend\\docs\\swagger.json"
    
    # Load swagger.json
    try:
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = json.load(f)
        print(f"✅ Loaded swagger.json")
    except Exception as e:
        print(f"❌ Error loading swagger.json: {e}")
        return False
    
    # Invoice Types endpoints to add
    invoice_types_endpoints = {
        "/api/v1/invoice-types": {
            "get": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Get all invoice types",
                "description": "Retrieve list of all invoice types with their current counters",
                "parameters": [
                    {
                        "name": "active_only",
                        "in": "query",
                        "description": "Filter only active invoice types",
                        "required": False,
                        "type": "boolean"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice types retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/models.InvoiceTypeWithCounter"
                                    }
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "post": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Create new invoice type",
                "description": "Create a new invoice type with unique code and automatic counter initialization",
                "parameters": [
                    {
                        "name": "invoice_type",
                        "in": "body",
                        "description": "Invoice type data",
                        "required": True,
                        "schema": {
                            "$ref": "#/definitions/models.InvoiceTypeCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Invoice type created successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceTypeResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Invalid input data"
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/active": {
            "get": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Get active invoice types",
                "description": "Get only active invoice types for dropdowns and forms",
                "responses": {
                    "200": {
                        "description": "Active invoice types retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/models.InvoiceTypeResponse"
                                    }
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/{id}": {
            "get": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Get invoice type by ID",
                "description": "Retrieve specific invoice type with its counter details",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice type retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceTypeWithCounter"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Invoice type not found"
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "put": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Update invoice type",
                "description": "Update invoice type details (name, description, status)",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    },
                    {
                        "name": "invoice_type",
                        "in": "body",
                        "description": "Updated invoice type data",
                        "required": True,
                        "schema": {
                            "$ref": "#/definitions/models.InvoiceTypeUpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice type updated successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceTypeResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Invoice type not found"
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "delete": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Delete invoice type",
                "description": "Delete invoice type (admin only). Warning: This will affect historical data.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice type deleted successfully"
                    },
                    "404": {
                        "description": "Invoice type not found"
                    },
                    "403": {
                        "description": "Admin access required"
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/{id}/toggle": {
            "post": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Toggle invoice type status",
                "description": "Toggle active/inactive status of invoice type",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice type status toggled successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceTypeResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Invoice type not found"
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/{id}/counter": {
            "get": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Get invoice counter for type",
                "description": "Get current counter value for specific invoice type and year",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    },
                    {
                        "name": "year",
                        "in": "query",
                        "description": "Year for counter (default: current year)",
                        "required": False,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Counter retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceCounterResponse"
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            },
            "put": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Update invoice counter",
                "description": "Update counter value for specific invoice type and year (admin only)",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    },
                    {
                        "name": "counter_update",
                        "in": "body",
                        "description": "Counter update data",
                        "required": True,
                        "schema": {
                            "$ref": "#/definitions/models.InvoiceCounterUpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Counter updated successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceCounterResponse"
                                }
                            }
                        }
                    },
                    "403": {
                        "description": "Admin access required"
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/{id}/counter-history": {
            "get": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Get counter history",
                "description": "Get historical counter values for specific invoice type across years",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Invoice type ID",
                        "required": True,
                        "type": "integer"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Counter history retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/models.InvoiceCounterResponse"
                                    }
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        },
        "/api/v1/invoice-types/preview-number": {
            "post": {
                "tags": ["Settings", "Invoice Types"],
                "summary": "Preview next invoice number",
                "description": "Preview what the next invoice number would be for a given type and date",
                "parameters": [
                    {
                        "name": "preview_request",
                        "in": "body",
                        "description": "Preview request data",
                        "required": True,
                        "schema": {
                            "$ref": "#/definitions/models.InvoiceNumberRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Invoice number preview generated successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "status": {"type": "string"},
                                "message": {"type": "string"},
                                "data": {
                                    "$ref": "#/definitions/models.InvoiceNumberResponse"
                                }
                            }
                        }
                    }
                },
                "security": [{"BearerAuth": []}]
            }
        }
    }
    
    # Definitions to add
    new_definitions = {
        "models.InvoiceTypeCreateRequest": {
            "type": "object",
            "required": ["name", "code"],
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Invoice type name (e.g., 'Corporate Sales')",
                    "maxLength": 100
                },
                "code": {
                    "type": "string", 
                    "description": "Invoice type code (e.g., 'STA-C')",
                    "maxLength": 20
                },
                "description": {
                    "type": "string",
                    "description": "Optional description",
                    "maxLength": 500
                }
            }
        },
        "models.InvoiceTypeUpdateRequest": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Invoice type name",
                    "maxLength": 100
                },
                "code": {
                    "type": "string",
                    "description": "Invoice type code", 
                    "maxLength": 20
                },
                "description": {
                    "type": "string",
                    "description": "Description",
                    "maxLength": 500
                },
                "is_active": {
                    "type": "boolean",
                    "description": "Active status"
                }
            }
        },
        "models.InvoiceTypeResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "description": "Invoice type ID"
                },
                "name": {
                    "type": "string",
                    "description": "Invoice type name"
                },
                "code": {
                    "type": "string",
                    "description": "Invoice type code"
                },
                "description": {
                    "type": "string", 
                    "description": "Description"
                },
                "is_active": {
                    "type": "boolean",
                    "description": "Active status"
                },
                "created_by": {
                    "type": "integer",
                    "description": "Creator user ID"
                },
                "created_at": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Creation timestamp"
                },
                "updated_at": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Last update timestamp"
                }
            }
        },
        "models.InvoiceTypeWithCounter": {
            "type": "object",
            "allOf": [
                {"$ref": "#/definitions/models.InvoiceTypeResponse"},
                {
                    "type": "object",
                    "properties": {
                        "current_counter": {
                            "type": "integer",
                            "description": "Current counter value for this year"
                        },
                        "current_year": {
                            "type": "integer",
                            "description": "Current year"
                        },
                        "last_invoice_number": {
                            "type": "string",
                            "description": "Last generated invoice number"
                        }
                    }
                }
            ]
        },
        "models.InvoiceCounterResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "description": "Counter ID"
                },
                "invoice_type_id": {
                    "type": "integer",
                    "description": "Invoice type ID"
                },
                "year": {
                    "type": "integer",
                    "description": "Year"
                },
                "counter": {
                    "type": "integer",
                    "description": "Counter value"
                },
                "invoice_type": {
                    "$ref": "#/definitions/models.InvoiceTypeResponse"
                }
            }
        },
        "models.InvoiceCounterUpdateRequest": {
            "type": "object",
            "required": ["year", "counter"],
            "properties": {
                "year": {
                    "type": "integer",
                    "description": "Year for the counter",
                    "minimum": 2020,
                    "maximum": 2100
                },
                "counter": {
                    "type": "integer",
                    "description": "New counter value",
                    "minimum": 0
                }
            }
        },
        "models.InvoiceNumberRequest": {
            "type": "object",
            "required": ["invoice_type_id", "date"],
            "properties": {
                "invoice_type_id": {
                    "type": "integer",
                    "description": "Invoice type ID"
                },
                "date": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Date for the invoice number"
                }
            }
        },
        "models.InvoiceNumberResponse": {
            "type": "object",
            "properties": {
                "invoice_number": {
                    "type": "string",
                    "description": "Generated invoice number (e.g., 'STA-C/001/I/2025')"
                },
                "counter": {
                    "type": "integer",
                    "description": "Counter value used"
                },
                "year": {
                    "type": "integer",
                    "description": "Year"
                },
                "month_roman": {
                    "type": "string",
                    "description": "Month in Roman numeral"
                },
                "type_code": {
                    "type": "string",
                    "description": "Invoice type code"
                }
            }
        }
    }
    
    # Create backup
    backup_file = swagger_file + f'.backup_invoice_types_{datetime.now().strftime("%Y%m%d_%H%M%S")}'
    with open(backup_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    print(f"💾 Backup created: {backup_file}")
    
    # Add endpoints
    paths = swagger_data.get('paths', {})
    for path, methods in invoice_types_endpoints.items():
        if path in paths:
            print(f"⚠️  Path {path} already exists, merging...")
            for method, spec in methods.items():
                paths[path][method] = spec
        else:
            paths[path] = methods
            print(f"✅ Added endpoint: {path}")
    
    # Add definitions
    definitions = swagger_data.get('definitions', {})
    for name, definition in new_definitions.items():
        definitions[name] = definition
        print(f"✅ Added definition: {name}")
    
    # Add Invoice Types tag if not exists
    tags = swagger_data.get('tags', [])
    invoice_types_tag = {
        "name": "Invoice Types",
        "description": "Invoice types and numbering management"
    }
    
    # Check if tag exists
    tag_exists = any(tag.get('name') == 'Invoice Types' for tag in tags)
    if not tag_exists:
        tags.append(invoice_types_tag)
        print("✅ Added Invoice Types tag")
    
    # Save updated swagger.json
    with open(swagger_file, 'w', encoding='utf-8') as f:
        json.dump(swagger_data, f, indent=2)
    
    endpoints_added = len(invoice_types_endpoints)
    definitions_added = len(new_definitions)
    
    print(f"✅ Updated swagger.json saved")
    print(f"📄 Added {endpoints_added} invoice types endpoints")
    print(f"📝 Added {definitions_added} new definitions")
    
    return True

def main():
    """Main function"""
    print("📄 INVOICE TYPES SWAGGER INTEGRATION")
    print("=" * 40)
    
    success = add_invoice_types_endpoints()
    
    if success:
        print(f"\n🎉 INVOICE TYPES SUCCESSFULLY ADDED TO SETTINGS!")
        print(f"📝 Next steps:")
        print(f"1. Restart your Go backend server")
        print(f"2. Clear browser cache")
        print(f"3. Open Swagger UI: http://localhost:8080/swagger")
        print(f"4. Look for 'Invoice Types' section in Settings")
        print(f"\n🔧 Available endpoints:")
        print(f"   • GET /api/v1/invoice-types (list all types with counters)")
        print(f"   • POST /api/v1/invoice-types (create new type)")
        print(f"   • GET /api/v1/invoice-types/{{id}} (get specific type)")
        print(f"   • PUT /api/v1/invoice-types/{{id}} (update type)")
        print(f"   • PUT /api/v1/invoice-types/{{id}}/counter (update counter)")
        print(f"   • GET /api/v1/invoice-types/{{id}}/counter-history (view history)")
        print(f"   • POST /api/v1/invoice-types/preview-number (preview next number)")
        return 0
    else:
        print(f"\n❌ Failed to add invoice types endpoints")
        return 1

if __name__ == "__main__":
    exit(main())