import React, { useState, useEffect } from 'react';
import {
  Box,
  FormControl,
  FormLabel,
  Input,
  Select,
  Badge,
  FormHelperText,
  Alert,
  AlertIcon,
  Text,
  VStack,
  HStack
} from '@chakra-ui/react';
import { formatDateWithIndonesianMonth } from '../../utils/dataFormatters';

interface SalesFormProps {
  formData: any;
  setFormData: (data: any) => void;
}

export const EnhancedSalesForm: React.FC<SalesFormProps> = ({ formData, setFormData }) => {
  const [calculatedDueDate, setCalculatedDueDate] = useState<string>('');
  const [paymentTermsExplanation, setPaymentTermsExplanation] = useState<string>('');

  // Format date untuk display Indonesia dengan nama bulan (DD Month YYYY)
  const formatDateForDisplay = (dateString: string): string => {
    if (!dateString) return '';
    return formatDateWithIndonesianMonth(dateString);
  };

  // Format date untuk input HTML (YYYY-MM-DD)
  const formatDateForInput = (dateString: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toISOString().split('T')[0];
  };

  // Hitung due date berdasarkan payment terms
  const calculateDueDate = (invoiceDate: string, paymentTerms: string): string => {
    if (!invoiceDate) return '';
    
    const baseDate = new Date(invoiceDate);
    let daysToAdd = 0;

    switch (paymentTerms) {
      case 'COD':
        daysToAdd = 0;
        break;
      case 'NET15':
        daysToAdd = 15;
        break;
      case 'NET30':
        daysToAdd = 30;
        break;
      case 'NET45':
        daysToAdd = 45;
        break;
      case 'NET60':
        daysToAdd = 60;
        break;
      case 'NET90':
        daysToAdd = 90;
        break;
      default:
        daysToAdd = 30; // Default NET30
    }

    const dueDate = new Date(baseDate);
    dueDate.setDate(dueDate.getDate() + daysToAdd);
    
    return dueDate.toISOString().split('T')[0];
  };

  // Update payment terms explanation
  const getPaymentTermsExplanation = (terms: string, invoiceDate: string): string => {
    if (!invoiceDate || !terms) return '';
    
    const invoiceDisplayDate = formatDateForDisplay(invoiceDate);
    const dueDate = calculateDueDate(invoiceDate, terms);
    const dueDisplayDate = formatDateForDisplay(dueDate);

    switch (terms) {
      case 'COD':
        return `Pembayaran tunai pada saat pengiriman (${invoiceDisplayDate})`;
      case 'NET15':
        return `Pembayaran dalam 15 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
      case 'NET30':
        return `Pembayaran dalam 30 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
      case 'NET45':
        return `Pembayaran dalam 45 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
      case 'NET60':
        return `Pembayaran dalam 60 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
      case 'NET90':
        return `Pembayaran dalam 90 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
      default:
        return `Pembayaran dalam 30 hari dari tanggal invoice (${invoiceDisplayDate} → ${dueDisplayDate})`;
    }
  };

  // Effect untuk auto-calculate due date
  useEffect(() => {
    if (formData.date && formData.payment_terms) {
      const newDueDate = calculateDueDate(formData.date, formData.payment_terms);
      setCalculatedDueDate(newDueDate);
      setPaymentTermsExplanation(getPaymentTermsExplanation(formData.payment_terms, formData.date));
      
      // Update form data dengan due date yang dihitung
      setFormData({
        ...formData,
        due_date: newDueDate
      });
    }
  }, [formData.date, formData.payment_terms]);

  const handleDateChange = (field: string, value: string) => {
    setFormData({
      ...formData,
      [field]: value
    });
  };

  const handlePaymentTermsChange = (value: string) => {
    setFormData({
      ...formData,
      payment_terms: value
    });
  };

  return (
    <VStack spacing={4} align="stretch">
      {/* Alert untuk format tanggal */}
      <Alert status="info" size="sm">
        <AlertIcon />
        <Text fontSize="sm">
          📅 Tanggal ditampilkan dalam format Indonesia dengan nama bulan (DD Bulan YYYY)
        </Text>
      </Alert>

      <HStack spacing={4}>
        {/* Invoice Date */}
        <FormControl isRequired>
          <FormLabel>
            Tanggal Invoice <Text as="span" color="red.500">*</Text>
          </FormLabel>
          <Input
            type="date"
            value={formatDateForInput(formData.date)}
            onChange={(e) => handleDateChange('date', e.target.value)}
            bg="white"
          />
          <FormHelperText>
            Kapan transaksi ini terjadi
            {formData.date && (
              <Text as="span" fontWeight="medium" color="blue.600">
                {' → '}{formatDateForDisplay(formData.date)}
              </Text>
            )}
          </FormHelperText>
        </FormControl>

        {/* Payment Terms */}
        <FormControl>
          <FormLabel>Syarat Pembayaran</FormLabel>
          <Select
            value={formData.payment_terms || 'NET30'}
            onChange={(e) => handlePaymentTermsChange(e.target.value)}
            bg="white"
          >
            <option value="COD">COD - Cash on Delivery</option>
            <option value="NET15">NET 15 - 15 Hari</option>
            <option value="NET30">NET 30 - 30 Hari</option>
            <option value="NET45">NET 45 - 45 Hari</option>
            <option value="NET60">NET 60 - 60 Hari</option>
            <option value="NET90">NET 90 - 90 Hari</option>
          </Select>
          {paymentTermsExplanation && (
            <FormHelperText color="blue.600" fontWeight="medium">
              {paymentTermsExplanation}
            </FormHelperText>
          )}
        </FormControl>
      </HStack>

      {/* Due Date - Auto calculated */}
      <FormControl>
        <FormLabel>
          Tanggal Jatuh Tempo 
          <Badge ml={2} colorScheme="green" size="sm">Auto-calculated</Badge>
        </FormLabel>
        <Input
          type="date"
          value={calculatedDueDate}
          isReadOnly
          bg="gray.50"
          cursor="not-allowed"
        />
        <FormHelperText color="green.600">
          Otomatis dihitung dari tanggal invoice + syarat pembayaran
          {calculatedDueDate && (
            <Text as="span" fontWeight="bold">
              {' → '}{formatDateForDisplay(calculatedDueDate)}
            </Text>
          )}
        </FormHelperText>
      </FormControl>

      {/* Validation Alert */}
      {formData.date && formData.payment_terms && (
        <Alert status="success" size="sm">
          <AlertIcon />
          <Box>
            <Text fontWeight="medium">Konfirmasi Perhitungan:</Text>
            <Text fontSize="sm">
              Invoice: <strong>{formatDateForDisplay(formData.date)}</strong> + 
              Payment Terms: <strong>{formData.payment_terms}</strong> = 
              Due Date: <strong>{formatDateForDisplay(calculatedDueDate)}</strong>
            </Text>
          </Box>
        </Alert>
      )}
    </VStack>
  );
};

export default EnhancedSalesForm;

'use client';

import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Select,
  Textarea,
  VStack,
  HStack,
  Box,
  Divider,
  Text,
  IconButton,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  useToast,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Switch,
  Badge,
  Flex,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Icon,
  useColorModeValue,
  Tooltip,
  ButtonGroup,
  RadioGroup,
  Radio,
  Stack
} from '@chakra-ui/react';
import CurrencyInput from '@/components/common/CurrencyInput';
import { useForm, useFieldArray } from 'react-hook-form';
import { 
  FiPlus, 
  FiTrash2, 
  FiSave, 
  FiX, 
  FiDollarSign, 
  FiShoppingCart, 
  FiFileText,
  FiCalendar,
  FiClock,
  FiCheck,
  FiAlertCircle,
  FiRefreshCw
} from 'react-icons/fi';
import salesService, { 
  Sale, 
  SaleCreateRequest, 
  SaleUpdateRequest, 
  SaleItemRequest,
  SaleItemUpdateRequest 
} from '@/services/salesService';
import ErrorHandler from '@/utils/errorHandler';
import { useAuth } from '@/contexts/AuthContext';

// Enhanced interfaces
interface DocumentTypeConfig {
  type: 'QUOTATION' | 'INVOICE';
  validUntil?: Date;
  dueDate?: Date;
  paymentTerms?: string;
  date: Date;
}

interface PaymentTerm {
  code: string;
  label: string;
  days: number | null;
  discount?: { rate: number; days: number };
}

interface EnhancedFormData {
  // Document type control
  document_type: 'QUOTATION' | 'INVOICE';
  
  // Basic info
  customer_id: number;
  sales_person_id?: number;
  type: string;
  
  // Smart date management
  date: string;
  due_date?: string;
  due_date_manual_override?: boolean;
  valid_until?: string;
  
  // Financial
  currency: string;
  exchange_rate: number;
  discount_percent: number;
  ppn_percent: number;
  pph_percent: number;
  pph_type?: string;
  payment_terms: string;
  payment_method?: string;
  shipping_method?: string;
  shipping_cost: number;
  shipping_taxable?: boolean;
  
  // Address & notes
  billing_address?: string;
  shipping_address?: string;
  notes?: string;
  internal_notes?: string;
  reference?: string;
  
  // Items
  items: Array<{
    id?: number;
    product_id: number;
    description: string;
    quantity: number;
    unit_price: number;
    discount_percent: number;
    taxable: boolean;
    revenue_account_id?: number;
    delete?: boolean;
  }>;
}

interface EnhancedSalesFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
  sale?: Sale | null;
}

// Enhanced Payment Terms with smart calculation - aligned with backend
const PAYMENT_TERMS: PaymentTerm[] = [
  { code: 'COD', label: 'Cash on Delivery (COD)', days: 0 },
  { code: 'NET15', label: 'NET 15', days: 15 },
  { code: 'NET30', label: 'NET 30', days: 30 },
  { code: 'NET45', label: 'NET 45', days: 45 },
  { code: 'NET60', label: 'NET 60', days: 60 },
  { code: 'NET90', label: 'NET 90', days: 90 },
  { code: 'EOM', label: 'End of Month', days: null },
  { code: '2_10_NET_30', label: '2/10, Net 30', days: 30, discount: { rate: 2, days: 10 }}
];

const EnhancedSalesForm: React.FC<EnhancedSalesFormProps> = ({
  isOpen,
  onClose,
  onSave,
  sale
}) => {
  const [loading, setLoading] = useState(false);
  const [customers, setCustomers] = useState<any[]>([]);
  const [products, setProducts] = useState<any[]>([]);
  const [salesPersons, setSalesPersons] = useState<any[]>([]);
  const [accounts, setAccounts] = useState<any[]>([]);
  const [loadingData, setLoadingData] = useState(true);
  const toast = useToast();
  const { user, token } = useAuth();
  const modalBodyRef = useRef<HTMLDivElement>(null);
  
  // Color mode values
  const modalBg = useColorModeValue('white', 'gray.800');
  const headerBg = useColorModeValue('blue.50', 'gray.700');
  const headingColor = useColorModeValue('blue.700', 'blue.300');
  const subHeadingColor = useColorModeValue('gray.600', 'gray.300');
  const textColor = useColorModeValue('gray.600', 'gray.400');
  const inputBg = useColorModeValue('gray.50', 'gray.600');
  const inputFocusBg = useColorModeValue('white', 'gray.500');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const alertBg = useColorModeValue('blue.50', 'blue.900');
  const alertBorderColor = useColorModeValue('blue.200', 'blue.700');
  
  // Permission check
  const userRole = user?.role?.toLowerCase();
  const canCreateSales = userRole === 'finance' || userRole === 'director' || userRole === 'admin';
  const canEditSales = userRole === 'admin' || userRole === 'finance' || userRole === 'director';
  const hasPermission = sale ? canEditSales : canCreateSales;
  
  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    control,
    formState: { errors }
  } = useForm<EnhancedFormData>({
    defaultValues: {
      document_type: 'INVOICE',
      type: 'INVOICE',
      currency: 'IDR',
      exchange_rate: 1,
      discount_percent: 0,
      ppn_percent: 11,
      pph_percent: 0,
      payment_terms: 'NET30',
      shipping_cost: 0,
      shipping_taxable: true,
      due_date_manual_override: false,
      items: [
        {
          product_id: 0,
          description: '',
          quantity: 1,
          unit_price: 0,
          discount_percent: 0,
          taxable: true
        }
      ]
    }
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items'
  });

  // Watch important fields for calculations and auto-updates
  const watchDocumentType = watch('document_type');
  const watchDate = watch('date');
  const watchPaymentTerms = watch('payment_terms');
  const watchDueDateManualOverride = watch('due_date_manual_override');
  const watchItems = watch('items');
  const watchDiscountPercent = watch('discount_percent');
  const watchPPNPercent = watch('ppn_percent');
  const watchShippingCost = watch('shipping_cost');
  const watchShippingTaxable = watch('shipping_taxable');

  // Smart Due Date Calculation
  const calculateDueDate = (invoiceDate: string, terms: string): string => {
    if (!invoiceDate || watchDocumentType === 'QUOTATION') return '';
    
    const date = new Date(invoiceDate);
    const term = PAYMENT_TERMS.find(t => t.code === terms);
    
    if (!term) return invoiceDate;
    
    if (term.code === 'EOM') {
      // End of month
      const nextMonth = new Date(date.getFullYear(), date.getMonth() + 1, 0);
      return nextMonth.toISOString().split('T')[0];
    } else if (term.days !== null) {
      // Add days
      const dueDate = new Date(date);
      dueDate.setDate(dueDate.getDate() + term.days);
      return dueDate.toISOString().split('T')[0];
    }
    
    return invoiceDate;
  };

  // Auto-update due date when date or terms change (unless manually overridden)
  useEffect(() => {
    if (watchDate && !watchDueDateManualOverride && watchDocumentType === 'INVOICE') {
      const newDueDate = calculateDueDate(watchDate, watchPaymentTerms);
      setValue('due_date', newDueDate);
    }
  }, [watchDate, watchPaymentTerms, watchDueDateManualOverride, watchDocumentType, setValue]);

  // Enhanced Tax Calculation
  const calculateTotals = useMemo(() => {
    if (!watchItems?.length) {
      return {
        subtotal: 0,
        afterGlobalDiscount: 0,
        taxableAmount: 0,
        ppn: 0,
        grandTotal: 0
      };
    }

    // 1. Calculate subtotal from all items
    const subtotal = watchItems.reduce((sum, item) => {
      if (!item.quantity || !item.unit_price) return sum;
      const lineTotal = item.quantity * item.unit_price;
      const discountAmount = lineTotal * (item.discount_percent || 0) / 100;
      return sum + (lineTotal - discountAmount);
    }, 0);

    // 2. Apply global discount
    const globalDiscountAmount = subtotal * (watchDiscountPercent || 0) / 100;
    const afterGlobalDiscount = subtotal - globalDiscountAmount;

    // 3. Calculate taxable amount (includes shipping if taxable)
    const shippingForTax = (watchShippingTaxable && watchShippingCost) ? watchShippingCost : 0;
    const taxableAmount = afterGlobalDiscount + shippingForTax;

    // 4. Calculate PPN
    const ppn = taxableAmount * (watchPPNPercent || 0) / 100;

    // 5. Grand total
    const grandTotal = afterGlobalDiscount + ppn + (watchShippingCost || 0);

    return {
      subtotal,
      afterGlobalDiscount,
      taxableAmount,
      ppn,
      grandTotal
    };
  }, [watchItems, watchDiscountPercent, watchPPNPercent, watchShippingCost, watchShippingTaxable]);

  // Permission check effect
  useEffect(() => {
    if (isOpen && user && !hasPermission) {
      const action = sale ? 'edit' : 'create';
      toast({
        title: 'Access Denied',
        description: `You do not have permission to ${action} sales. Contact your administrator for access.`,
        status: 'error',
        duration: 5000,
      });
      onClose();
    }
  }, [isOpen, user, hasPermission, sale, toast, onClose]);

  // Load form data effect
  useEffect(() => {
    if (isOpen) {
      loadFormData();
      if (sale) {
        populateFormWithSale(sale);
      } else {
        resetForm();
      }
    }
  }, [isOpen, sale]);

  const loadFormData = async () => {
    if (!token) {
      toast({
        title: 'Authentication Required',
        description: 'Please login to access this feature.',
        status: 'error',
        duration: 5000,
      });
      return;
    }

    setLoadingData(true);
    
    try {
      // Load all necessary data (same as original, but cleaned up)
      const [customersResult, productsResult, salesPersonsResult, accountsResult] = await Promise.allSettled([
        // Load customers
        (async () => {
          const contactService = await import('@/services/contactService');
          return await contactService.default.getContacts(token, 'CUSTOMER');
        })(),
        
        // Load products with fallback
        (async () => {
          try {
            const productService = await import('@/services/productService');
            const result = await productService.default.getProducts({}, token);
            return result;
          } catch (error: any) {
            console.warn('Failed to load products:', error?.message);
            return { data: [] };
          }
        })(),
        
        // Load sales persons
        (async () => {
          const contactService = await import('@/services/contactService');
          return await contactService.default.getContacts(token, 'EMPLOYEE');
        })(),
        
        // Load revenue accounts
        (async () => {
          const accountService = await import('@/services/accountService');
          return await accountService.default.getAccounts(token, 'REVENUE');
        })()
      ]);

      // Process results (same logic as original)
      if (customersResult.status === 'fulfilled' && Array.isArray(customersResult.value)) {
        setCustomers(customersResult.value);
      } else {
        setCustomers([]);
      }

      if (productsResult.status === 'fulfilled' && productsResult.value?.data && Array.isArray(productsResult.value.data)) {
        setProducts(productsResult.value.data);
      } else {
        setProducts([]);
      }

      if (salesPersonsResult.status === 'fulfilled' && Array.isArray(salesPersonsResult.value)) {
        const salesPersonsData = salesPersonsResult.value.map(contact => ({
          ...contact,
          name: contact.name || contact.company_name || 'Unknown Employee'
        }));
        setSalesPersons(salesPersonsData);
      } else {
        setSalesPersons([]);
      }

      if (accountsResult.status === 'fulfilled' && Array.isArray(accountsResult.value)) {
        setAccounts(accountsResult.value);
      } else {
        setAccounts([]);
      }

    } catch (error: any) {
      console.error('Error loading form data:', error);
      toast({
        title: 'Loading Error',
        description: 'Failed to load form data. Please try again.',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setLoadingData(false);
    }
  };

  const populateFormWithSale = (saleData: Sale) => {
    reset({
      document_type: 'INVOICE',
      customer_id: saleData.customer_id,
      sales_person_id: saleData.sales_person_id,
      type: 'INVOICE',
      date: saleData.date.split('T')[0],
      due_date: saleData.due_date ? saleData.due_date.split('T')[0] : undefined,
      valid_until: saleData.valid_until ? saleData.valid_until.split('T')[0] : undefined,
      currency: saleData.currency,
      exchange_rate: saleData.exchange_rate,
      discount_percent: saleData.discount_percent,
      ppn_percent: saleData.ppn_percent,
      pph_percent: saleData.pph_percent,
      pph_type: saleData.pph_type,
      payment_terms: saleData.payment_terms,
      payment_method: saleData.payment_method,
      shipping_method: saleData.shipping_method,
      shipping_cost: saleData.shipping_cost,
      shipping_taxable: true, // Default assumption
      billing_address: saleData.billing_address,
      shipping_address: saleData.shipping_address,
      notes: saleData.notes,
      internal_notes: saleData.internal_notes,
      reference: saleData.reference,
      due_date_manual_override: !!saleData.due_date, // If due_date exists, assume it was set manually
      items: saleData.sale_items?.map(item => ({
        id: item.id,
        product_id: item.product_id,
        description: item.description || '',
        quantity: item.quantity,
        unit_price: item.unit_price,
        discount_percent: item.discount_percent,
        taxable: item.taxable,
        revenue_account_id: item.revenue_account_id
      })) || []
    });
  };

  const resetForm = () => {
    const today = new Date().toISOString().split('T')[0];
    reset({
      document_type: 'INVOICE',
      type: 'INVOICE',
      date: today,
      currency: 'IDR',
      exchange_rate: 1,
      discount_percent: 0,
      ppn_percent: 11,
      pph_percent: 0,
      payment_terms: 'NET30',
      shipping_cost: 0,
      shipping_taxable: true,
      due_date_manual_override: false,
      items: [
        {
          product_id: 0,
          description: '',
          quantity: 1,
          unit_price: 0,
          discount_percent: 0,
          taxable: true
        }
      ]
    });
  };

  // Handle document type change
  const handleDocumentTypeChange = (newType: 'QUOTATION' | 'INVOICE') => {
    setValue('document_type', newType);
    setValue('type', newType);
    
    if (newType === 'QUOTATION') {
      // For quotation, set valid_until to 30 days from date
      if (watchDate) {
        const validDate = new Date(watchDate);
        validDate.setDate(validDate.getDate() + 30);
        setValue('valid_until', validDate.toISOString().split('T')[0]);
      }
      setValue('due_date', '');
    } else {
      // For invoice, calculate due date
      setValue('valid_until', '');
      if (watchDate && !watchDueDateManualOverride) {
        const newDueDate = calculateDueDate(watchDate, watchPaymentTerms);
        setValue('due_date', newDueDate);
      }
    }
  };

  // Handle due date manual override
  const handleDueDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setValue('due_date', e.target.value);
    setValue('due_date_manual_override', true);
  };

  // Reset due date to auto-calculated
  const resetDueDateToAuto = () => {
    setValue('due_date_manual_override', false);
    if (watchDate) {
      const newDueDate = calculateDueDate(watchDate, watchPaymentTerms);
      setValue('due_date', newDueDate);
    }
  };

  const handleProductChange = (index: number, productId: number) => {
    const product = products.find(p => p.id === parseInt(productId.toString()));
    if (product) {
      setValue(`items.${index}.product_id`, product.id);
      setValue(`items.${index}.description`, product.name);
      setValue(`items.${index}.unit_price`, product.price || 0);
    }
  };

  const addItem = () => {
    append({
      product_id: 0,
      description: '',
      quantity: 1,
      unit_price: 0,
      discount_percent: 0,
      taxable: true
    });
  };

  const removeItem = (index: number) => {
    if (fields.length > 1) {
      remove(index);
    }
  };

  const validateForm = (data: EnhancedFormData): string[] => {
    const errors: string[] = [];

    // Basic validation
    if (!data.customer_id) errors.push('Customer is required');
    if (!data.date) errors.push('Date is required');

    // Document type specific validation
    // Invoice-only validation
    if (data.due_date && new Date(data.due_date) < new Date(data.date)) {
      errors.push('Due Date cannot be earlier than invoice date');
    }

    // Items validation
    const validItems = data.items.filter(item => 
      (products.length > 0 ? item.product_id > 0 : item.description.trim() !== '') && 
      item.unit_price > 0
    );

    if (validItems.length === 0) {
      const errorMsg = products.length > 0 
        ? 'At least one item with a selected product is required'
        : 'At least one item with description and price is required';
      errors.push(errorMsg);
    }

    return errors;
  };

  const onSubmit = async (data: EnhancedFormData) => {
    try {
      setLoading(true);

      // Validation
      const validationErrors = validateForm(data);
      if (validationErrors.length > 0) {
        ErrorHandler.handleValidationError(validationErrors, toast, 'sales form');
        return;
      }

      // Filter valid items
      const validItems = data.items.filter(item => {
        if (products.length > 0) {
          return item.product_id > 0;
        }
        return item.description && item.description.trim() !== '' && item.unit_price > 0;
      });

      if (sale) {
        // Update existing sale
        const updateData: SaleUpdateRequest = {
          customer_id: data.customer_id,
          sales_person_id: data.sales_person_id,
          date: data.date ? `${data.date}T00:00:00Z` : undefined,
          due_date: data.due_date ? `${data.due_date}T00:00:00Z` : undefined,
          valid_until: data.valid_until ? `${data.valid_until}T00:00:00Z` : undefined,
          discount_percent: data.discount_percent,
          ppn_percent: data.ppn_percent,
          pph_percent: data.pph_percent,
          pph_type: data.pph_type,
          payment_terms: data.payment_terms,
          payment_method: data.payment_method,
          shipping_method: data.shipping_method,
          shipping_cost: data.shipping_cost,
          billing_address: data.billing_address,
          shipping_address: data.shipping_address,
          notes: data.notes,
          internal_notes: data.internal_notes,
          reference: data.reference,
          items: validItems.map(item => ({
            id: item.id,
            product_id: item.product_id,
            description: item.description,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount: item.discount_percent || 0,
            tax: 0,
            revenue_account_id: item.revenue_account_id,
            delete: item.delete || false
          }))
        };

        await salesService.updateSale(sale.id, updateData);
        ErrorHandler.handleSuccess('Sale has been updated successfully', toast, 'update sale');
      } else {
        // Create new sale
        const createData: SaleCreateRequest = {
          customer_id: data.customer_id,
          sales_person_id: data.sales_person_id,
          type: 'INVOICE',
          date: `${data.date}T00:00:00Z`,
          due_date: data.due_date ? `${data.due_date}T00:00:00Z` : undefined,
          valid_until: data.valid_until ? `${data.valid_until}T00:00:00Z` : undefined,
          currency: data.currency,
          exchange_rate: data.exchange_rate,
          discount_percent: data.discount_percent,
          ppn_percent: data.ppn_percent,
          pph_percent: data.pph_percent,
          pph_type: data.pph_type,
          payment_terms: data.payment_terms,
          payment_method: data.payment_method,
          shipping_method: data.shipping_method,
          shipping_cost: data.shipping_cost,
          billing_address: data.billing_address,
          shipping_address: data.shipping_address,
          notes: data.notes,
          internal_notes: data.internal_notes,
          reference: data.reference,
          items: validItems.map(item => ({
            product_id: item.product_id,
            description: item.description || '',
            quantity: Math.max(1, Math.floor(item.quantity || 1)),
            unit_price: Math.min(999999999999.99, Math.max(0, item.unit_price || 0)),
            discount: Math.min(999999.99, Math.max(0, item.discount_percent || 0)),
            discount_percent: Math.min(100, Math.max(0, item.discount_percent || 0)),
            tax: 0,
            taxable: item.taxable !== false,
            revenue_account_id: item.revenue_account_id || 0
          }))
        };

        await salesService.createSale(createData);
        ErrorHandler.handleSuccess(`${data.document_type.toLowerCase()} has been created successfully`, toast, 'create sale');
      }

      onSave();
      onClose();
    } catch (error: any) {
      ErrorHandler.handleSaveError('sale', error, toast, !!sale);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  // Don't render if no permission
  if (!hasPermission && user) {
    return null;
  }

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={handleClose} 
      size="6xl" 
      isCentered
      closeOnOverlayClick={false}
      scrollBehavior="inside"
      motionPreset="slideInBottom"
    >
      <ModalOverlay bg="blackAlpha.700" backdropFilter="blur(4px)" />
      <ModalContent 
        maxH="95vh" 
        minH="80vh"
        mx={4} 
        my={2} 
        borderRadius="xl"
        bg={modalBg}
        shadow="2xl"
        overflow="hidden"
        display="flex"
        flexDirection="column"
        w="full"
        maxW="6xl"
      >
        {/* Enhanced Header with Document Type Selector */}
        <ModalHeader 
          bg={headerBg} 
          borderBottomWidth={1} 
          borderColor={borderColor}
          pb={4}
          pt={6}
        >
          <VStack spacing={4} align="stretch">
            <HStack justify="space-between" align="center">
              <Box>
                <Heading size="lg" color={headingColor}>
                  {sale ? 'Edit Sale Transaction' : 'Create New Sale'}
                </Heading>
                <Text color={textColor} fontSize="sm" mt={1}>
                  {sale ? 'Modify existing sale details and items' : 'Create a new sales transaction with items and pricing'}
                </Text>
              </Box>
              <Badge colorScheme="blue" variant="solid" px={3} py={1} borderRadius="md">
                <Icon as={FiShoppingCart} mr={1} />
                Sale Form
              </Badge>
            </HStack>
            
          </VStack>
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)} style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
          {/* Hidden fields */}
          <input type="hidden" {...register('type')} />
          
          <ModalBody 
            ref={modalBodyRef}
            flex="1" 
            overflowY="auto" 
            px={6} 
            py={4}
            pb={8}
            maxH="calc(95vh - 200px)"
            minH="400px"
          >
            <VStack spacing={6} align="stretch">
              {/* Basic Information with Smart Date Logic */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  📋 Basic Information
                </Heading>
                <VStack spacing={4}>
                  <HStack w="full" spacing={4}>
                    <FormControl isRequired isInvalid={!!errors.customer_id} flex={2}>
                      <FormLabel>Customer</FormLabel>
                      <Select
                        {...register('customer_id', {
                          required: 'Customer is required',
                          setValueAs: value => parseInt(value) || 0
                        })}
                        bg={inputBg}
                        _focus={{ bg: inputFocusBg }}
                        isDisabled={loadingData}
                      >
                        <option value="">
                          {loadingData ? 'Loading customers...' : 
                           customers.length === 0 ? 'No customers available' : 'Select customer'}
                        </option>
                        {customers.map(customer => (
                          <option key={customer.id} value={customer.id}>
                            {customer.code} - {customer.name}
                          </option>
                        ))}
                      </Select>
                      <FormErrorMessage>{errors.customer_id?.message}</FormErrorMessage>
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>Sales Person</FormLabel>
                      <Select
                        {...register('sales_person_id', {
                          setValueAs: value => value ? parseInt(value) : undefined
                        })}
                        bg={inputBg}
                        _focus={{ bg: inputFocusBg }}
                        isDisabled={loadingData}
                      >
                        <option value="">
                          {loadingData ? 'Loading sales persons...' : 
                           salesPersons.length === 0 ? 'No sales persons available' : 'Select sales person'}
                        </option>
                        {salesPersons.map(person => (
                          <option key={person.id} value={person.id}>
                            {person.name}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                  </HStack>
                  
                  {/* Smart Date Row */}
                  <HStack w="full" spacing={4}>
                    <FormControl isRequired isInvalid={!!errors.date} flex={1}>
                      <FormLabel>
                        <HStack>
                          <Icon as={FiCalendar} />
                          <Text>Date</Text>
                          <Text fontSize="xs" color={textColor}>(DD/MM/YYYY)</Text>
                        </HStack>
                      </FormLabel>
                      <Input
                        type="date"
                        {...register('date', {
                          required: 'Date is required'
                        })}
                        bg={inputBg}
                        _focus={{ bg: inputFocusBg }}
                      />
                      <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                    </FormControl>

                    {/* Payment Terms - Only for Invoice */}
                    {watchDocumentType === 'INVOICE' && (
                      <FormControl flex={1}>
                        <FormLabel>Payment Terms</FormLabel>
                        <Select 
                          {...register('payment_terms')}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        >
                          {PAYMENT_TERMS.map(term => (
                            <option key={term.code} value={term.code}>
                              {term.label}
                            </option>
                          ))}
                        </Select>
                      </FormControl>
                    )}

                    {/* Due Date - Only for Invoice */}
                    {watchDocumentType === 'INVOICE' && (
                      <FormControl flex={1}>
                        <FormLabel>
                          <HStack>
                            <Text>Due Date</Text>
                            <Text fontSize="xs" color={textColor}>(DD/MM/YYYY)</Text>
                            {watchDueDateManualOverride ? (
                              <Tooltip label="Reset to auto-calculated">
                                <IconButton
                                  size="xs"
                                  icon={<FiRefreshCw />}
                                  aria-label="Reset due date"
                                  onClick={resetDueDateToAuto}
                                  colorScheme="blue"
                                  variant="ghost"
                                />
                              </Tooltip>
                            ) : (
                              <Badge size="sm" colorScheme="green" variant="subtle">
                                <Icon as={FiCheck} mr={1} />
                                Auto
                              </Badge>
                            )}
                          </HStack>
                        </FormLabel>
                        <Input
                          type="date"
                          {...register('due_date')}
                          onChange={handleDueDateChange}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                        {/* Due Date Calculation Information */}
                        {watchDate && watchPaymentTerms && watchDocumentType === 'INVOICE' && (
                          <Text fontSize="xs" color={textColor} mt={1}>
                            {(() => {
                              const term = PAYMENT_TERMS.find(t => t.code === watchPaymentTerms);
                              const invoiceDate = new Date(watchDate).toLocaleDateString('en-GB', {
                                day: '2-digit',
                                month: '2-digit', 
                                year: 'numeric'
                              });
                              const dueDate = watch('due_date');
                              const dueDateFormatted = dueDate ? new Date(dueDate).toLocaleDateString('en-GB', {
                                day: '2-digit',
                                month: '2-digit',
                                year: 'numeric'
                              }) : '';
                              
                              if (!term) return '';
                              
                              if (term.code === 'COD') {
                                return `💰 Cash on Delivery - Payment due immediately`;
                              } else if (term.code === 'EOM') {
                                return `📅 End of Month - Payment due: ${dueDateFormatted}`;
                              } else if (term.days) {
                                return `📅 ${term.label} - Invoice: ${invoiceDate} + ${term.days} days = ${dueDateFormatted}`;
                              }
                              return '';
                            })()
                          }
                          </Text>
                        )}
                      </FormControl>
                    )}


                    <FormControl flex={1}>
                      <FormLabel>Reference</FormLabel>
                      <Input
                        {...register('reference')}
                        placeholder="External reference number"
                        bg={inputBg}
                        _focus={{ bg: inputFocusBg }}
                      />
                    </FormControl>
                  </HStack>
                </VStack>
              </Box>

              <Divider />

              {/* Items Section - Same as original but with live calculation */}
              <Box>
                <Flex justify="space-between" align="center" mb={4}>
                  <Heading size="md" color={subHeadingColor}>
                    🛍️ Sale Items
                  </Heading>
                  <Button
                    size="sm"
                    colorScheme="blue"
                    leftIcon={<FiPlus />}
                    onClick={addItem}
                  >
                    Add Item
                  </Button>
                </Flex>

                {/* Products warning if empty */}
                {!loadingData && products.length === 0 && (
                  <Alert status="warning" mb={4} borderRadius="md" bg={alertBg} borderColor={alertBorderColor}>
                    <AlertIcon />
                    <AlertDescription fontSize="sm">
                      Products are not available. You can still create sales by manually entering product information in the description field.
                    </AlertDescription>
                  </Alert>
                )}

                <TableContainer
                  border="1px" 
                  borderColor={borderColor} 
                  borderRadius="md"
                  bg="white"
                  shadow="sm"
                >
                  <Table variant="simple" size="sm">
                    <Thead>
                      <Tr>
                        <Th>Product</Th>
                        <Th>Description</Th>
                        <Th>Qty</Th>
                        <Th>Unit Price</Th>
                        <Th>Discount %</Th>
                        <Th>Taxable</Th>
                        <Th>Total</Th>
                        <Th>Action</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {fields.map((field, index) => {
                        const item = watchItems[index] || {};
                        const lineTotal = (item.quantity || 0) * (item.unit_price || 0);
                        const discountAmount = lineTotal * (item.discount_percent || 0) / 100;
                        const finalAmount = lineTotal - discountAmount;

                        return (
                          <Tr key={field.id}>
                            <Td>
                              <Select
                                size="sm"
                                {...register(`items.${index}.product_id`, {
                                  setValueAs: value => parseInt(value) || 0
                                })}
                                onChange={(e) => handleProductChange(index, parseInt(e.target.value))}
                                bg={inputBg}
                                _focus={{ bg: inputFocusBg }}
                              >
                                <option value="">Select product</option>
                                {products.map(product => (
                                  <option key={product.id} value={product.id}>
                                    {product.code} - {product.name}
                                  </option>
                                ))}
                              </Select>
                            </Td>
                            <Td>
                              <Input
                                size="sm"
                                {...register(`items.${index}.description`)}
                                placeholder="Item description"
                                bg={inputBg}
                                _focus={{ bg: inputFocusBg }}
                              />
                            </Td>
                            <Td>
                              <NumberInput size="sm" min={1} w="80px">
                                <NumberInputField
                                  {...register(`items.${index}.quantity`, {
                                    setValueAs: value => parseInt(value) || 1
                                  })}
                                />
                                <NumberInputStepper>
                                  <NumberIncrementStepper />
                                  <NumberDecrementStepper />
                                </NumberInputStepper>
                              </NumberInput>
                            </Td>
                            <Td>
                              <CurrencyInput
                                value={item.unit_price || 0}
                                onChange={(value) => setValue(`items.${index}.unit_price`, value)}
                                placeholder="Rp 10.000"
                                size="sm"
                                min={0}
                                showLabel={false}
                              />
                            </Td>
                            <Td>
                              <NumberInput size="sm" min={0} max={100} w="80px">
                                <NumberInputField
                                  {...register(`items.${index}.discount_percent`, {
                                    setValueAs: value => parseFloat(value) || 0
                                  })}
                                />
                              </NumberInput>
                            </Td>
                            <Td>
                              <Switch
                                size="sm"
                                {...register(`items.${index}.taxable`)}
                              />
                            </Td>
                            <Td>
                              <Text fontSize="sm" fontWeight="medium">
                                {salesService.formatCurrency(finalAmount)}
                              </Text>
                            </Td>
                            <Td>
                              <IconButton
                                size="sm"
                                colorScheme="red"
                                variant="ghost"
                                icon={<FiTrash2 />}
                                onClick={() => removeItem(index)}
                                isDisabled={fields.length === 1}
                                aria-label="Remove item"
                              />
                            </Td>
                          </Tr>
                        );
                      })}
                    </Tbody>
                  </Table>
                </TableContainer>
              </Box>

              <Divider />

              {/* Enhanced Pricing & Taxes Section */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  💰 Pricing & Taxes
                </Heading>
                <VStack spacing={4}>
                  <HStack w="full" spacing={4}>
                    <FormControl>
                      <FormLabel>Global Discount (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('discount_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </NumberInput>
                    </FormControl>

                    <FormControl>
                      <FormLabel>PPN (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('ppn_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </NumberInput>
                    </FormControl>

                    <FormControl>
                      <FormLabel>
                        <HStack>
                          <Text>Shipping Cost</Text>
                          <Switch
                            size="sm"
                            {...register('shipping_taxable')}
                          />
                          <Text fontSize="xs" color={textColor}>Taxable</Text>
                        </HStack>
                      </FormLabel>
                      <CurrencyInput
                        value={watchShippingCost || 0}
                        onChange={(value) => setValue('shipping_cost', value)}
                        placeholder="Rp 0"
                        min={0}
                        showLabel={false}
                        size="md"
                      />
                    </FormControl>
                  </HStack>

                  {/* Live Total Calculation Display */}
                  {calculateTotals.subtotal > 0 && (
                    <Alert status="info" borderRadius="lg" bg={alertBg} borderColor={alertBorderColor}>
                      <AlertIcon color="blue.500" />
                      <Box w="full">
                        <AlertTitle fontSize="sm" mb={2}>Calculation Summary</AlertTitle>
                        <VStack align="stretch" spacing={2}>
                          <HStack justify="space-between">
                            <Text fontSize="sm" color={textColor}>Subtotal:</Text>
                            <Text fontSize="sm" fontWeight="medium">
                              {salesService.formatCurrency(calculateTotals.subtotal)}
                            </Text>
                          </HStack>
                          {watchDiscountPercent > 0 && (
                            <HStack justify="space-between">
                              <Text fontSize="sm" color={textColor}>
                                Global Discount ({watchDiscountPercent}%):
                              </Text>
                              <Text fontSize="sm" color="red.500" fontWeight="medium">
                                -{salesService.formatCurrency(calculateTotals.subtotal - calculateTotals.afterGlobalDiscount)}
                              </Text>
                            </HStack>
                          )}
                          <HStack justify="space-between">
                            <Text fontSize="sm" color={textColor}>After Discount:</Text>
                            <Text fontSize="sm" fontWeight="medium">
                              {salesService.formatCurrency(calculateTotals.afterGlobalDiscount)}
                            </Text>
                          </HStack>
                          {watchShippingCost > 0 && (
                            <HStack justify="space-between">
                              <Text fontSize="sm" color={textColor}>
                                Shipping {watchShippingTaxable ? '(Taxable)' : '(Non-taxable)'}:
                              </Text>
                              <Text fontSize="sm" fontWeight="medium">
                                {salesService.formatCurrency(watchShippingCost)}
                              </Text>
                            </HStack>
                          )}
                          {watchPPNPercent > 0 && (
                            <HStack justify="space-between">
                              <Text fontSize="sm" color={textColor}>
                                PPN ({watchPPNPercent}%):
                              </Text>
                              <Text fontSize="sm" fontWeight="medium">
                                {salesService.formatCurrency(calculateTotals.ppn)}
                              </Text>
                            </HStack>
                          )}
                          <Divider />
                          <HStack justify="space-between">
                            <Text fontSize="md" fontWeight="bold" color="blue.600">Grand Total:</Text>
                            <Text fontSize="lg" fontWeight="bold" color="blue.600">
                              {salesService.formatCurrency(calculateTotals.grandTotal)}
                            </Text>
                          </HStack>
                        </VStack>
                      </Box>
                    </Alert>
                  )}
                </VStack>
              </Box>

              <Divider />

              {/* Additional Information */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  📝 Additional Information
                </Heading>
                <VStack spacing={4}>
                  <FormControl>
                    <FormLabel>Notes</FormLabel>
                    <Textarea
                      {...register('notes')}
                      placeholder="Customer-visible notes"
                      rows={3}
                      bg={inputBg}
                      _focus={{ bg: inputFocusBg }}
                    />
                  </FormControl>

                  <FormControl>
                    <FormLabel>Internal Notes</FormLabel>
                    <Textarea
                      {...register('internal_notes')}
                      placeholder="Internal notes (not visible to customer)"
                      rows={2}
                      bg={inputBg}
                      _focus={{ bg: inputFocusBg }}
                    />
                  </FormControl>
                </VStack>
              </Box>
            </VStack>
          </ModalBody>

          {/* Enhanced Footer */}
          <ModalFooter 
            borderTopWidth={1} 
            borderColor={borderColor}
            bg="white"
            px={6}
            py={4}
          >
            <HStack justify="space-between" w="full">
              {/* Left - Status Info */}
              <VStack align="start" spacing={1}>
                <HStack spacing={2}>
                  <Badge colorScheme={'green'}>
                    {'INVOICE'}
                  </Badge>
                  <Text fontSize="sm" color={textColor}>
                    {fields.length} item{fields.length !== 1 ? 's' : ''}
                  </Text>
                  {calculateTotals.grandTotal > 0 && (
                    <Text fontSize="sm" fontWeight="medium" color="blue.600">
                      Total: {salesService.formatCurrency(calculateTotals.grandTotal)}
                    </Text>
                  )}
                </HStack>
                {watchDocumentType === 'INVOICE' && watchDueDateManualOverride && (
                  <HStack spacing={1}>
                    <Icon as={FiAlertCircle} color="orange.500" size="12px" />
                    <Text fontSize="xs" color="orange.500">Due date manually overridden</Text>
                  </HStack>
                )}
              </VStack>
              
              {/* Right - Actions */}
              <HStack spacing={3}>
                <Button
                  leftIcon={<FiX />}
                  onClick={handleClose}
                  variant="outline"
                  size="lg"
                  isDisabled={loading}
                  colorScheme="gray"
                >
                  Cancel
                </Button>
                <Button
                  leftIcon={loading ? undefined : <FiSave />}
                  type="submit"
                  colorScheme="blue"
                  size="lg"
                  isLoading={loading}
                  loadingText={sale ? "Updating..." : 'Creating Invoice...'}
                  shadow="md"
                >
                  {sale ? 'Update Sale' : 'Create Invoice'}
                </Button>
              </HStack>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default EnhancedSalesForm;
