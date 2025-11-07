'use client';

import React, { useState } from 'react';
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Alert,
  AlertIcon,
  Select,
  Input,
  useToast,
  Card,
  CardHeader,
  CardBody,
  Heading,
  FormControl,
  FormLabel,
  Divider
} from '@chakra-ui/react';
import { FiDownload, FiEye } from 'react-icons/fi';
import {
  PDFReportGenerator,
  ReportData,
  TableColumn,
  generateReportNumber
} from '@/utils/pdfReportGenerator';

interface PDFReportExampleProps {
  className?: string;
}

const PDFReportExample: React.FC<PDFReportExampleProps> = ({ className }) => {
  const [reportType, setReportType] = useState<'invoice' | 'quote' | 'purchase'>('invoice');
  const [reportTitle, setReportTitle] = useState('INVOICE');
  const [customData, setCustomData] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const toast = useToast();

  // Sample data similar to your invoice
  const getSampleData = (): ReportData => {
    const baseColumns: TableColumn[] = [
      { header: 'No.', dataKey: 'no', width: 20 },
      { header: 'Description', dataKey: 'description', width: 100 },
      { header: 'Qty', dataKey: 'qty', width: 30 },
      { header: 'Unit Price', dataKey: 'unitPrice', width: 40 },
      { header: 'Total', dataKey: 'total', width: 40 }
    ];

    const sampleItems = [
      {
        no: 1,
        description: 'Laptop Dell XPS 13',
        qty: 1,
        unitPrice: 20000000,
        total: 20000000
      },
      {
        no: 2,
        description: 'Microsoft Office 365 License',
        qty: 1,
        unitPrice: 1500000,
        total: 1500000
      },
      {
        no: 3,
        description: 'Wireless Mouse Logitech',
        qty: 2,
        unitPrice: 350000,
        total: 700000
      }
    ];

    // Try to parse custom data if provided
    let items = sampleItems;
    if (customData.trim()) {
      try {
        const parsed = JSON.parse(customData);
        if (Array.isArray(parsed)) {
          items = parsed.map((item, index) => ({
            no: index + 1,
            ...item
          }));
        }
      } catch (error) {
        console.warn('Invalid custom data, using sample data');
      }
    }

    const subtotal = items.reduce((sum, item) => sum + item.total, 0);
    const taxRate = 11;
    const tax = subtotal * (taxRate / 100);
    const total = subtotal + tax;

    return {
      columns: baseColumns,
      data: items,
      summary: {
        subtotal,
        tax,
        taxRate,
        total
      }
    };
  };

  const generateReport = async (action: 'download' | 'preview' = 'download') => {
    setIsGenerating(true);
    try {
      // Generate report number
      const reportNumber = await generateReportNumber(reportType);
      
      // Get sample data
      const reportData = getSampleData();
      
      // Generate PDF using static method
      const doc = await PDFReportGenerator.generateFromSettings(
        reportTitle,
        reportData,
        {
          reportNumber,
          date: new Date().toLocaleDateString('id-ID'),
          subtitle: reportType === 'invoice' ? 'Sales Invoice Document' : 
                   reportType === 'quote' ? 'Price Quotation Document' :
                   'Purchase Order Document'
        }
      );

      if (action === 'download') {
        // Download the PDF
        doc.save(`${reportType}-${reportNumber.replace(/\//g, '-')}.pdf`);
        
        toast({
          title: 'PDF Generated Successfully',
          description: `${reportTitle} has been downloaded`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        // Open in new tab for preview
        const pdfBlob = doc.output('blob');
        const url = URL.createObjectURL(pdfBlob);
        window.open(url, '_blank');
        
        // Clean up the URL after a short delay
        setTimeout(() => URL.revokeObjectURL(url), 1000);
        
        toast({
          title: 'PDF Preview Opened',
          description: 'Check the new tab for your PDF preview',
          status: 'info',
          duration: 3000,
          isClosable: true,
        });
      }
    } catch (error: any) {
      console.error('Error generating PDF:', error);
      toast({
        title: 'Error Generating PDF',
        description: error.message || 'An unexpected error occurred',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsGenerating(false);
    }
  };

  const handleReportTypeChange = (type: 'invoice' | 'quote' | 'purchase') => {
    setReportType(type);
    switch (type) {
      case 'invoice':
        setReportTitle('INVOICE');
        break;
      case 'quote':
        setReportTitle('QUOTATION');
        break;
      case 'purchase':
        setReportTitle('PURCHASE ORDER');
        break;
    }
  };

  return (
    <Box className={className}>
      <Card>
        <CardHeader>
          <Heading size="md">PDF Report Generator</Heading>
          <Text fontSize="sm" color="gray.600">
            Generate professional PDF reports using company settings from database
          </Text>
        </CardHeader>
        <CardBody>
          <VStack spacing={4} align="stretch">
            <Alert status="info" borderRadius="md">
              <AlertIcon />
              <Box>
                <Text fontWeight="bold">Company Integration</Text>
                <Text fontSize="sm">
                  Logo and company information will be automatically loaded from Settings
                </Text>
              </Box>
            </Alert>

            <HStack spacing={4} align="end">
              <FormControl>
                <FormLabel>Report Type</FormLabel>
                <Select 
                  value={reportType} 
                  onChange={(e) => handleReportTypeChange(e.target.value as any)}
                >
                  <option value="invoice">Invoice</option>
                  <option value="quote">Quotation</option>
                  <option value="purchase">Purchase Order</option>
                </Select>
              </FormControl>

              <FormControl>
                <FormLabel>Report Title</FormLabel>
                <Input
                  value={reportTitle}
                  onChange={(e) => setReportTitle(e.target.value)}
                  placeholder="Enter custom title"
                />
              </FormControl>
            </HStack>

            <FormControl>
              <FormLabel>Custom Data (JSON)</FormLabel>
              <Input
                as="textarea"
                rows={4}
                value={customData}
                onChange={(e) => setCustomData(e.target.value)}
                placeholder={`Optional: Enter custom items data in JSON format:\n[\n  {\n    "description": "Product Name",\n    "qty": 1,\n    "unitPrice": 100000,\n    "total": 100000\n  }\n]`}
                fontSize="sm"
                fontFamily="mono"
              />
              <Text fontSize="xs" color="gray.500" mt={1}>
                Leave empty to use sample data (Laptop Dell XPS 13, etc.)
              </Text>
            </FormControl>

            <Divider />

            <HStack justify="space-between">
              <Text fontSize="sm" color="gray.600">
                Features: Auto company header, logo integration, tax calculation, professional layout
              </Text>
              
              <HStack>
                <Button
                  leftIcon={<FiEye />}
                  onClick={() => generateReport('preview')}
                  isLoading={isGenerating}
                  variant="outline"
                  size="sm"
                >
                  Preview
                </Button>
                <Button
                  leftIcon={<FiDownload />}
                  onClick={() => generateReport('download')}
                  colorScheme="blue"
                  isLoading={isGenerating}
                  loadingText="Generating..."
                  size="sm"
                >
                  Generate PDF
                </Button>
              </HStack>
            </HStack>

            <Box bg="gray.50" p={4} borderRadius="md">
              <Text fontSize="sm" fontWeight="bold" mb={2}>Sample Output Features:</Text>
              <VStack align="start" spacing={1}>
                <Text fontSize="xs">✓ Company logo from settings (with fallback placeholder)</Text>
                <Text fontSize="xs">✓ Company name, address, phone, email from database</Text>
                <Text fontSize="xs">✓ Auto-generated report numbers with proper formatting</Text>
                <Text fontSize="xs">✓ Professional invoice layout similar to your template</Text>
                <Text fontSize="xs">✓ Tax calculation (PPN) with configurable rates</Text>
                <Text fontSize="xs">✓ Currency formatting based on system settings</Text>
                <Text fontSize="xs">✓ Multi-language support (Indonesian/English)</Text>
              </VStack>
            </Box>
          </VStack>
        </CardBody>
      </Card>
    </Box>
  );
};

export default PDFReportExample;