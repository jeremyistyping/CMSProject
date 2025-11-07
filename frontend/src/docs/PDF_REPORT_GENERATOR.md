# PDF Report Generator Documentation

Sistem PDF Report Generator ini memungkinkan Anda untuk membuat laporan PDF profesional dengan logo dan informasi perusahaan yang diambil secara otomatis dari database settings.

## Features

- ✅ **Integrasi dengan Settings Database**: Logo dan informasi perusahaan diambil otomatis dari tabel settings
- ✅ **Layout Profesional**: Mirip dengan template invoice yang sudah ada  
- ✅ **Multi-Language Support**: Mendukung Bahasa Indonesia dan Inggris
- ✅ **Auto Report Number**: Generate nomor laporan otomatis berdasarkan prefix settings
- ✅ **Currency Formatting**: Format mata uang sesuai dengan settings sistem
- ✅ **Tax Calculation**: Perhitungan pajak otomatis dengan rate yang dapat dikonfigurasi
- ✅ **Logo Support**: Mendukung logo dari upload dengan fallback placeholder
- ✅ **Responsive Design**: Layout yang menyesuaikan dengan orientasi portrait/landscape

## Installation & Setup

### Prerequisites
- jsPDF dan jspdf-autotable sudah terinstall (✅ sudah ada di package.json)
- API settings endpoint berfungsi dengan baik
- Upload logo sistem berfungsi dengan baik

### Import
```typescript
import {
  PDFReportGenerator,
  ReportData,
  TableColumn,
  ReportConfig,
  generateReportNumber,
  fetchCompanySettings,
  settingsToCompanyInfo
} from '@/utils/pdfReportGenerator';
```

## Usage Examples

### 1. Simple Usage (Recommended)
Cara termudah menggunakan static method `generateFromSettings`:

```typescript
import { PDFReportGenerator, ReportData, TableColumn } from '@/utils/pdfReportGenerator';

// Prepare your data
const reportData: ReportData = {
  columns: [
    { header: 'No.', dataKey: 'no', width: 20 },
    { header: 'Description', dataKey: 'description', width: 100 },
    { header: 'Qty', dataKey: 'qty', width: 30 },
    { header: 'Unit Price', dataKey: 'unitPrice', width: 40 },
    { header: 'Total', dataKey: 'total', width: 40 }
  ],
  data: [
    {
      no: 1,
      description: 'Laptop Dell XPS 13',
      qty: 1,
      unitPrice: 20000000,
      total: 20000000
    }
  ],
  summary: {
    subtotal: 20000000,
    tax: 2200000,
    taxRate: 11,
    total: 22200000
  }
};

// Generate PDF
const generateInvoice = async () => {
  try {
    const doc = await PDFReportGenerator.generateFromSettings(
      'INVOICE',
      reportData,
      {
        reportNumber: 'INV/2025/09/0002', // or use generateReportNumber()
        date: '25/09/2025',
        subtitle: 'Sales Invoice Document'
      }
    );
    
    // Download PDF
    doc.save('invoice-INV-2025-09-0002.pdf');
    
    // Or get as blob for other purposes
    const blob = doc.output('blob');
    
  } catch (error) {
    console.error('Error generating PDF:', error);
  }
};
```

### 2. Advanced Usage with Custom Configuration

```typescript
// Create instance
const generator = new PDFReportGenerator('portrait');

// Load settings manually if needed
await generator.loadSettings();

// Create custom config (override settings data)
const config: ReportConfig = {
  title: 'CUSTOM INVOICE',
  subtitle: 'Special Pricing Document',
  date: new Date().toLocaleDateString('id-ID'),
  reportNumber: await generateReportNumber('invoice'),
  companyInfo: {
    name: 'Custom Company Name',
    address: 'Custom Address',
    phone: '+62-xxx-xxx',
    email: 'custom@email.com',
    website: 'www.custom.com',
    tax_number: 'NPWP-123456',
    logo: 'data:image/png;base64,...' // base64 logo
  }
};

const doc = await generator.generateReport(config, reportData);
doc.save('custom-invoice.pdf');
```

### 3. Auto Report Number Generation

```typescript
import { generateReportNumber } from '@/utils/pdfReportGenerator';

// Generate report numbers based on system settings
const invoiceNumber = await generateReportNumber('invoice'); // Uses invoice_prefix from settings
const quoteNumber = await generateReportNumber('quote');     // Uses quote_prefix from settings  
const purchaseNumber = await generateReportNumber('purchase'); // Uses purchase_prefix from settings

console.log(invoiceNumber); // Output: "INV/2025/09/0001"
console.log(quoteNumber);   // Output: "QUO/2025/09/0001"  
console.log(purchaseNumber); // Output: "PUR/2025/09/0001"
```

### 4. Working with React Component

```tsx
import React, { useState } from 'react';
import { Button, useToast } from '@chakra-ui/react';
import { PDFReportGenerator } from '@/utils/pdfReportGenerator';

const InvoiceGenerator: React.FC = () => {
  const [isGenerating, setIsGenerating] = useState(false);
  const toast = useToast();

  const handleGeneratePDF = async () => {
    setIsGenerating(true);
    try {
      const doc = await PDFReportGenerator.generateFromSettings(
        'INVOICE',
        yourReportData,
        {
          reportNumber: await generateReportNumber('invoice'),
          date: new Date().toLocaleDateString('id-ID')
        }
      );
      
      doc.save('invoice.pdf');
      
      toast({
        title: 'PDF Generated',
        description: 'Invoice has been downloaded successfully',
        status: 'success',
        duration: 3000,
      });
      
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to generate PDF',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <Button 
      onClick={handleGeneratePDF} 
      isLoading={isGenerating}
      loadingText="Generating..."
    >
      Generate Invoice PDF
    </Button>
  );
};
```

## Data Structures

### ReportData Interface
```typescript
interface ReportData {
  columns: TableColumn[];
  data: any[];
  summary?: {
    subtotal?: number;
    tax?: number;
    taxRate?: number;
    total?: number;
  };
}
```

### TableColumn Interface
```typescript
interface TableColumn {
  header: string;    // Column header text
  dataKey: string;   // Property key in data objects
  width?: number;    // Optional column width
}
```

### ReportConfig Interface
```typescript
interface ReportConfig {
  title: string;
  subtitle?: string;
  date?: string;
  reportNumber?: string;
  companyInfo: CompanyInfo;
}
```

## Integration with System Settings

### Company Info Mapping
PDF generator otomatis mengambil data dari tabel settings dengan mapping:

| Settings Field | Company Info Field | Description |
|---|---|---|
| `company_name` | `name` | Nama perusahaan |
| `company_address` | `address` | Alamat perusahaan |
| `company_phone` | `phone` | Nomor telepon |
| `company_email` | `email` | Email perusahaan |
| `company_website` | `website` | Website perusahaan |
| `tax_number` | `tax_number` | Nomor NPWP |
| `company_logo` | `logo` | Path logo perusahaan |

### Currency & Formatting
- Currency format: `settings.currency` (default: 'IDR')
- Decimal places: `settings.decimal_places` (default: 0)
- Language: `settings.language` (untuk format locale)
- Tax rate: `settings.default_tax_rate` (untuk perhitungan pajak)

### Report Number Generation
- Invoice: `${invoice_prefix}/${year}/${month}/${invoice_next_number}`
- Quote: `${quote_prefix}/${year}/${month}/${quote_next_number}`
- Purchase: `${purchase_prefix}/${year}/${month}/${purchase_next_number}`

## Logo Integration

### Logo Sources
1. **Database Upload**: Logo yang diupload melalui settings page
2. **Base64 String**: Logo dalam format base64 data URL
3. **Placeholder**: Fallback icon `</>` jika logo tidak tersedia

### Logo Processing
```typescript
// Automatic logo conversion from settings
const settings = await fetchCompanySettings();
const logoUrl = getImageUrl(settings.company_logo); // Convert path to full URL
const logoBase64 = await imageToBase64(logoUrl);     // Convert to base64 for PDF
```

## Error Handling

### Common Errors & Solutions

1. **Settings Not Found**
   ```typescript
   // Error: Could not load company settings
   // Solution: Check API endpoint /settings
   ```

2. **Logo Not Loading**
   ```typescript
   // Error: Could not add logo
   // Solution: Check image path and server accessibility
   // Fallback: Placeholder logo will be shown automatically
   ```

3. **Invalid Data Format**
   ```typescript
   // Error: Invalid table data
   // Solution: Ensure data matches TableColumn dataKey properties
   ```

## Advanced Configuration

### Custom Styling
```typescript
// Override default styles in PDFReportGenerator class
const generator = new PDFReportGenerator();

// Modify table styles
const tableConfig = {
  headStyles: {
    fillColor: [66, 139, 202], // Blue header
    textColor: 255,
    fontSize: 11,
    fontStyle: 'bold'
  },
  alternateRowStyles: {
    fillColor: [245, 245, 245] // Light gray alternating rows
  }
};
```

### Multi-page Support
jsPDF akan otomatis menangani multiple pages jika konten melebihi tinggi halaman.

### Landscape Mode
```typescript
const generator = new PDFReportGenerator('landscape'); // untuk laporan yang lebar
```

## Testing

Gunakan component `PDFReportExample` untuk testing:

```tsx
import PDFReportExample from '@/components/reports/PDFReportExample';

// Di dalam page atau component
<PDFReportExample />
```

Component ini menyediakan:
- Interface untuk testing berbagai tipe laporan
- Preview dan download functionality
- Sample data untuk testing
- Custom data input (JSON)

## Troubleshooting

### PDF Tidak Ter-generate
1. Check console untuk error messages
2. Verify settings API endpoint berfungsi
3. Check image URLs accessibility
4. Verify jsPDF dependencies

### Logo Tidak Muncul  
1. Check company_logo path di settings
2. Verify image server accessibility
3. Try base64 logo instead of path
4. Check browser network tab untuk failed requests

### Format Currency Salah
1. Check settings.currency value
2. Verify settings.decimal_places
3. Check settings.language for locale

### Report Number Tidak Generate
1. Check settings prefix values (invoice_prefix, etc.)
2. Verify settings next_number values
3. Check date formatting

## Best Practices

1. **Always use try-catch** untuk error handling
2. **Load settings once** dan cache jika memungkinkan  
3. **Use static method** `generateFromSettings` untuk kemudahan
4. **Validate data** sebelum passing ke generator
5. **Show loading states** saat generate PDF
6. **Handle errors gracefully** dengan user feedback
7. **Test dengan data real** dari production database

## Example Integration dalam Sales/Invoice Module

```typescript
// Di sales invoice page
const generateSalesInvoice = async (saleId: number) => {
  try {
    // Fetch sale data from API
    const saleResponse = await api.get(`/sales/${saleId}`);
    const sale = saleResponse.data;
    
    // Convert to PDF format
    const reportData: ReportData = {
      columns: [
        { header: 'No.', dataKey: 'no', width: 20 },
        { header: 'Product', dataKey: 'product_name', width: 80 },
        { header: 'Qty', dataKey: 'quantity', width: 30 },
        { header: 'Price', dataKey: 'unit_price', width: 40 },
        { header: 'Total', dataKey: 'total_price', width: 40 }
      ],
      data: sale.items.map((item, index) => ({
        no: index + 1,
        product_name: item.product.name,
        quantity: item.quantity,
        unit_price: item.unit_price,
        total_price: item.total_price
      })),
      summary: {
        subtotal: sale.subtotal,
        tax: sale.tax_amount,
        taxRate: sale.tax_rate,
        total: sale.total_amount
      }
    };
    
    // Generate PDF
    const doc = await PDFReportGenerator.generateFromSettings(
      'INVOICE',
      reportData,
      {
        reportNumber: sale.invoice_number,
        date: new Date(sale.sale_date).toLocaleDateString('id-ID'),
        subtitle: `Bill To: ${sale.customer.name}`
      }
    );
    
    doc.save(`invoice-${sale.invoice_number.replace(/\//g, '-')}.pdf`);
    
  } catch (error) {
    console.error('Error generating invoice PDF:', error);
  }
};
```

---

Sistem PDF Report Generator ini sudah terintegrasi penuh dengan sistem settings yang ada dan siap digunakan untuk membuat berbagai jenis laporan profesional.