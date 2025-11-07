// Fallback data for forms when API services are unavailable
export const fallbackCustomers = [
  { 
    id: 1, 
    code: 'CUST001', 
    name: 'Default Customer',
    email: 'customer@example.com',
    phone: '021-123456',
    company_name: 'Default Company',
    type: 'CUSTOMER'
  },
  { 
    id: 2, 
    code: 'CUST002', 
    name: 'Sample Customer',
    email: 'sample@example.com',
    phone: '021-789012',
    company_name: 'Sample Corp',
    type: 'CUSTOMER'
  }
];

export const fallbackProducts = [
  { 
    id: 1, 
    code: 'PROD001', 
    name: 'Default Product',
    price: 100000,
    cost: 75000,
    category: 'General',
    unit: 'pcs',
    description: 'Default product for sales'
  },
  { 
    id: 2, 
    code: 'PROD002', 
    name: 'Sample Service',
    price: 500000,
    cost: 300000,
    category: 'Services',
    unit: 'hour',
    description: 'Sample service product'
  }
];

export const fallbackSalesPersons = [
  { 
    id: 1, 
    name: 'System Sales', 
    email: 'system@company.com',
    code: 'SYS001',
    phone: '021-555000',
    isDefault: true
  },
  { 
    id: 2, 
    name: 'Demo Sales Rep', 
    email: 'sales@company.com',
    code: 'SALES001',
    phone: '021-555001',
    isDefault: false
  }
];

export const fallbackAccounts = [
  { 
    id: 1, 
    code: 'REV001', 
    name: 'Sales Revenue',
    type: 'REVENUE',
    category: 'Income',
    description: 'Revenue from sales'
  },
  { 
    id: 2, 
    code: 'REV002', 
    name: 'Service Revenue',
    type: 'REVENUE',
    category: 'Income',
    description: 'Revenue from services'
  }
];

export const getFallbackData = () => ({
  customers: fallbackCustomers,
  products: fallbackProducts,
  salesPersons: fallbackSalesPersons,
  accounts: fallbackAccounts
});
