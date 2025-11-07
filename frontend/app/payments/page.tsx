'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { DataTable } from '@/components/common/DataTable';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Badge,
  Text,
  HStack,
  VStack,
  Select,
  Input,
  InputGroup,
  InputLeftElement,
  FormControl,
  FormLabel,
  IconButton,
  Tooltip,
  Spinner,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  Card,
  CardHeader,
  CardBody,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  SimpleGrid,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  useToast,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiPlus, 
  FiEye, 
  FiEdit, 
  FiTrash2, 
  FiFilter, 
  FiRefreshCw, 
  FiDownload, 
  FiDollarSign, 
  FiFilePlus,
  FiSearch,
  FiMoreVertical,
  FiFileText,
  FiChevronDown,
  FiArrowDown,
  FiArrowRight
} from 'react-icons/fi';
import paymentService, { Payment, PaymentFilters, PaymentResult, PaymentCreateRequest } from '@/services/paymentService';
import AdvancedPaymentForm from '@/components/payments/AdvancedPaymentForm';
import PaymentDetailModal from '@/components/payments/PaymentDetailModal';
import PPNPaymentModal from '@/components/payments/PPNPaymentModal';
import { exportPaymentsToPDF, exportPaymentDetailToPDF, PDFExportOptions } from '@/utils/pdfExport';
import ExportButton from '@/components/common/ExportButton';

// Status type for filtering
type PaymentStatusType = 'ALL' | 'PENDING' | 'COMPLETED' | 'FAILED';

// Payment method type for filtering
type PaymentMethodType = 'ALL' | 'CASH' | 'BANK_TRANSFER' | 'CHECK' | 'CREDIT_CARD' | 'DEBIT_CARD' | 'OTHER';

// Pagination settings
const ITEMS_PER_PAGE = 10;

// Date formatter for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID');
};

// Currency formatter - fixed to match sales format
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};

const PaymentsPage: React.FC = () => {
  const { token, user } = useAuth();
  const toast = useToast();
  
  // Theme colors for dark mode support
  const bgColor = useColorModeValue('gray.50', 'gray.900');
  const cardBg = useColorModeValue('white', 'gray.800');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textSecondary = useColorModeValue('gray.600', 'gray.400');
  const textPrimary = useColorModeValue('gray.700', 'gray.200');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const inputBg = useColorModeValue('white', 'gray.700');
  const searchIconColor = useColorModeValue('gray.300', 'gray.500');
  const alertBg = useColorModeValue('red.50', 'red.900');
  const alertTextColor = useColorModeValue('red.600', 'red.300');
  const statColors = {
    green: useColorModeValue('green.500', 'green.400'),
    blue: useColorModeValue('blue.600', 'blue.400'),
    purple: useColorModeValue('purple.600', 'purple.400'),
    orange: useColorModeValue('orange.600', 'orange.400')
  };

  // Tooltip descriptions for payment page
  const tooltips = {
    search: 'Cari pembayaran berdasarkan kode, nama contact, atau nomor referensi',
    contact: 'Pilih customer atau vendor yang terkait dengan pembayaran',
    paymentMethod: 'Metode pembayaran: Cash (tunai), Bank Transfer (transfer bank), Check (cek), Credit Card (kartu kredit), dll',
    amount: 'Jumlah nominal pembayaran yang dilakukan',
    date: 'Tanggal pembayaran dilakukan',
    reference: 'Nomor referensi pembayaran (contoh: nomor transfer, nomor cek)',
    status: 'Status pembayaran: Pending (menunggu), Completed (selesai), Failed (gagal)',
    bankAccount: 'Akun kas/bank yang digunakan untuk pembayaran',
    notes: 'Catatan atau keterangan tambahan untuk pembayaran ini',
    allocations: 'Alokasi pembayaran ke invoice atau purchase yang terkait',
    attachments: 'Lampiran bukti pembayaran (foto transfer, receipt, dll)',
  };

  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchInput, setSearchInput] = useState(''); // Local search state for client-side filtering
  const [showFilters, setShowFilters] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    total: 1,
    totalItems: 0
  });
  const [summary, setSummary] = useState<any>(null);
  
  // Filter states
  const [filters, setFilters] = useState<PaymentFilters>({
    page: 1,
    limit: ITEMS_PER_PAGE
  });
  const [statusFilter, setStatusFilter] = useState<PaymentStatusType>('ALL');
  const [methodFilter, setMethodFilter] = useState<PaymentMethodType>('ALL');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  
  // State for modals
  const [showPaymentForm, setShowPaymentForm] = useState(false);
  const [paymentType, setPaymentType] = useState<'receivable' | 'payable'>('receivable');
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null);
  const [showConfirmDelete, setShowConfirmDelete] = useState(false);
  const [showPaymentDetail, setShowPaymentDetail] = useState(false);
  const [formLoading, setFormLoading] = useState(false);
  const [showPPNPayment, setShowPPNPayment] = useState(false);
  const [ppnPaymentType, setPPNPaymentType] = useState<'INPUT' | 'OUTPUT'>('OUTPUT');

  // Permission checks - Normalize role comparison for case-insensitive check
  const userRole = user?.role?.toLowerCase();
  const canCreate = userRole === 'finance' || userRole === 'director';
  const canEdit = userRole === 'admin' || userRole === 'finance' || userRole === 'director';
  const canDelete = userRole === 'admin';
  const canExport = userRole === 'admin' || userRole === 'finance' || userRole === 'director';

  // New Payment handler
  const handleNewPayment = (type: 'receivable' | 'payable' = 'receivable') => {
    setPaymentType(type);
    setSelectedPayment(null);
    setShowPaymentForm(true);
  };
  
  // PPN Payment handler
  const handlePPNPayment = () => {
    setPPNPaymentType('OUTPUT');
    setShowPPNPayment(true);
  };
  
  // Edit Payment handler
  const handleEditPayment = (payment: Payment) => {
    // Check if PPN payment - cannot be edited
    if (payment.payment_type?.startsWith('TAX_PPN') || payment.code?.startsWith('SETOR-PPN')) {
      toast({
        title: 'Cannot Edit PPN Payment',
        description: 'PPN payments cannot be edited. Please delete and create new if needed.',
        status: 'warning',
        duration: 5000,
      });
      return;
    }
    
    // Detect payment type from contact type or code
    let type: 'receivable' | 'payable' = 'receivable';
    
    // Check contact type first (most reliable)
    if (payment.contact?.type === 'VENDOR') {
      type = 'payable';
    } else if (payment.contact?.type === 'CUSTOMER') {
      type = 'receivable';
    } 
    // Fallback to code prefix
    else if (payment.code?.startsWith('PAY')) {
      type = 'payable';
    } else if (payment.code?.startsWith('RCV')) {
      type = 'receivable';
    }
    
    setPaymentType(type);
    setSelectedPayment(payment);
    setShowPaymentForm(true);
  };
  
  // View Payment handler
  const handleViewPayment = (payment: Payment) => {
    setSelectedPayment(payment);
    setShowPaymentDetail(true);
  };
  
  // Delete Payment handler
  const handleDeletePayment = (payment: Payment) => {
    setSelectedPayment(payment);
    setShowConfirmDelete(true);
  };

  const columns = [
  {
    header: 'Code',
    accessor: (row: Payment) => (
      <Text fontWeight="medium" color={statColors.blue}>{row.code}</Text>
    )
  },
  { 
    header: 'Contact',
    accessor: (row: Payment) => {
      // For PPN tax payments, show "Negara" instead of contact name
      if (row.payment_type === 'TAX_PPN' || 
          row.payment_type === 'TAX_PPN_INPUT' || 
          row.payment_type === 'TAX_PPN_OUTPUT' ||
          row.code?.startsWith('SETOR-PPN')) {
        return 'Negara';
      }
      return row.contact?.name || '-';
    }
  },
  {
    header: 'Date',
    accessor: (row: Payment) => formatDate(row.date)
  },
  {
    header: 'Amount',
    accessor: (row: Payment) => (
      <Text fontWeight="medium">{formatCurrency(row.amount)}</Text>
    )
  },
  {
    header: 'Method',
    accessor: (row: Payment) => paymentService.getMethodDisplayName(row.method)
  },
  {
    header: 'Status',
    accessor: (row: Payment) => (
      <Badge colorScheme={paymentService.getStatusColorScheme(row.status)} variant="subtle">
        {row.status}
      </Badge>
    )
  },
  {
    header: 'Actions',
    accessor: (row: Payment) => (
      <Menu>
        <MenuButton as={IconButton} icon={<FiMoreVertical />} variant="ghost" size="sm" />
        <MenuList>
          <MenuItem icon={<FiEye />} onClick={() => handleViewPayment(row)}>
            View Details
          </MenuItem>
          {canEdit && (
            <MenuItem icon={<FiEdit />} onClick={() => handleEditPayment(row)}>
              Edit
            </MenuItem>
          )}
          <MenuItem 
            icon={<FiFilePlus />} 
            onClick={() => paymentService.downloadPaymentDetailPDF(row.id, row.code)}
          >
            Export PDF
          </MenuItem>
          {canDelete && (
            <>
              <MenuDivider />
              <MenuItem icon={<FiTrash2 />} color="red.500" onClick={() => handleDeletePayment(row)}>
                Delete
              </MenuItem>
            </>
          )}
        </MenuList>
      </Menu>
    )
  }
];

// Load payments data
const loadPayments = async (newFilters?: Partial<PaymentFilters>) => {
  try {
    setLoading(true);
    setError(null);
    
    const currentFilters = newFilters ? { ...filters, ...newFilters } : filters;
    
    // Prepare filters for API request
    const apiFilters: PaymentFilters = {
      page: currentFilters.page,
      limit: currentFilters.limit
    };
    
    // Add search filter if provided (use searchInput instead of currentFilters.search)
    if (searchInput && searchInput.trim()) {
      apiFilters.search = searchInput.trim();
    }
    
    // Add status filter if selected
    if (statusFilter !== 'ALL') {
      apiFilters.status = statusFilter;
    }
    
    // Add method filter if selected
    if (methodFilter !== 'ALL') {
      apiFilters.method = methodFilter;
    }
    
    // Add date filters if selected
    if (startDate) {
      apiFilters.start_date = startDate;
    }
    
    if (endDate) {
      apiFilters.end_date = endDate;
    }
    
    // Make API call
    const result = await paymentService.getPayments(apiFilters);
    
    // Update state with results (no need for allPayments anymore since filtering is on server-side)
    const paymentData = result?.data || [];
    setPayments(paymentData);
    setFilters({ ...currentFilters, page: result?.page || currentFilters.page });
    setPagination({
      current: result?.page || 1,
      total: result?.total_pages || 1,
      totalItems: result?.total || 0
    });
    
  } catch (err: any) {
    console.error('Error fetching payments:', err);
    setError(err.message || 'An error occurred while fetching payments');
    setPayments([]);
    
    toast({
      title: 'Error loading payments',
      description: err.message || 'Failed to load payments data',
      status: 'error',
      duration: 3000
    });
  } finally {
    setLoading(false);
  }
};

// Load payment summary
const loadPaymentSummary = async () => {
  try {
    const totalAmount = payments.reduce((sum, payment) => sum + payment.amount, 0);
    const completedPayments = payments.filter(p => p.status === 'COMPLETED');
    const pendingPayments = payments.filter(p => p.status === 'PENDING');
    const completedAmount = completedPayments.reduce((sum, payment) => sum + payment.amount, 0);
    const pendingAmount = pendingPayments.reduce((sum, payment) => sum + payment.amount, 0);
    
    setSummary({
      total_payments: payments.length,
      total_amount: totalAmount,
      completed_amount: completedAmount,
      pending_amount: pendingAmount,
      completed_count: completedPayments.length,
      pending_count: pendingPayments.length,
      avg_payment_value: payments.length > 0 ? totalAmount / payments.length : 0
    });
  } catch (error) {
    console.error('Failed to calculate payment summary:', error);
  }
};

// Initial load
useEffect(() => {
  if (token) {
    loadPayments();
  }
}, [token]);

// Update summary when payments change
useEffect(() => {
  if (payments.length > 0) {
    loadPaymentSummary();
  }
}, [payments]);

// Handle filters change - trigger API call
useEffect(() => {
  if (token) {
    loadPayments({ page: 1 });
  }
}, [statusFilter, methodFilter, startDate, endDate]);

// Handle search with debounce - trigger API call after 500ms of inactivity
useEffect(() => {
  if (!token) return;
  
  // Debounce search to avoid too many API calls
  const debounceTimer = setTimeout(() => {
    loadPayments({ page: 1 });
  }, 500);
  
  return () => clearTimeout(debounceTimer);
}, [searchInput]);

// Handle page change
const handlePageChange = (page: number) => {
  setFilters(prev => ({
    ...prev,
    page: page
  }));
};

// Server-side search handler with API call (debounced via useEffect)
const handleSearch = (value: string) => {
  setSearchInput(value);
  // API call will be triggered by useEffect with 500ms debounce
};

// Handle filter change
const handleFilterChange = (key: string, value: string) => {
  switch(key) {
    case 'status':
      setStatusFilter(value as PaymentStatusType);
      break;
    case 'method':
      setMethodFilter(value as PaymentMethodType);
      break;
    case 'start_date':
      setStartDate(value);
      break;
    case 'end_date':
      setEndDate(value);
      break;
  }
};

// Apply filters
const applyFilters = () => {
  loadPayments({ page: 1 });
};

// Reset filters
const resetFilters = () => {
  setStatusFilter('ALL');
  setMethodFilter('ALL');
  setStartDate('');
  setEndDate('');
  setSearchInput(''); // Clear search input
  setFilters({
    page: 1,
    limit: ITEMS_PER_PAGE
  });
  // loadPayments will be triggered by useEffect when searchInput changes
};

  // Handle delete payment
  const handleDelete = async (payment: Payment) => {
    if (!window.confirm('Are you sure you want to delete this payment?')) return;
    
    try {
      await paymentService.deletePayment(payment.id);
      toast({
        title: 'Success',
        description: 'Payment has been deleted successfully',
        status: 'success',
        duration: 3000
      });
      loadPayments();
    } catch (error: any) {
      toast({
        title: 'Error deleting payment',
        description: error.message || 'Failed to delete payment',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Confirm delete
  const confirmDeletePayment = async () => {
    if (!selectedPayment) return;
    
    try {
      await paymentService.deletePayment(selectedPayment.id);
      toast({
        title: 'Success',
        description: 'Payment has been deleted successfully',
        status: 'success',
        duration: 3000
      });
      loadPayments();
      setShowConfirmDelete(false);
      setSelectedPayment(null);
    } catch (error: any) {
      console.error('Error deleting payment:', error);
      toast({
        title: 'Error deleting payment',
        description: error.message || 'Failed to delete payment',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Export payments to Excel handler
  const handleExportPayments = async () => {
    try {
      // Prepare current filters for export
      const exportFilters: PaymentFilters = { ...filters };
      if (statusFilter !== 'ALL') exportFilters.status = statusFilter;
      if (methodFilter !== 'ALL') exportFilters.method = methodFilter;
      if (startDate) exportFilters.start_date = startDate;
      if (endDate) exportFilters.end_date = endDate;
      
      // Call export API
      const blob = await paymentService.exportPayments(exportFilters);
      
      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `payments-export-${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Error exporting payments:', error);
      setError(error.message || 'Failed to export payments');
    }
  };

  // Export payments to PDF handler
  const handleExportPaymentsPDF = async () => {
    try {
      setError(null);
      
      // Prepare PDF export options
      const pdfOptions: PDFExportOptions = {
        title: 'Laporan Pembayaran',
        subtitle: 'Daftar Transaksi Pembayaran',
        companyName: 'PT. Sistem Akuntansi',
        companyAddress: 'Jakarta, Indonesia',
        includeFilters: true,
        statusFilter: statusFilter,
        methodFilter: methodFilter,
        startDate: startDate,
        endDate: endDate
      };
      
      // Export PDF with current data
      exportPaymentsToPDF(payments, pdfOptions);
      
    } catch (error: any) {
      console.error('Error exporting payments to PDF:', error);
      setError(error.message || 'Failed to export payments to PDF');
    }
  };
  
  // Handle bulk export (like sales page)
  const handleBulkExport = async () => {
    try {
      // Prepare filters for export
      const exportStatus = statusFilter !== 'ALL' ? statusFilter : undefined;
      const exportMethod = methodFilter !== 'ALL' ? methodFilter : undefined;
      
      // Use backend PDF export
      await paymentService.downloadPaymentReportPDF(
        startDate || undefined,
        endDate || undefined,
        exportStatus,
        exportMethod
      );
      
      toast({
        title: 'Success',
        description: 'Payment report has been downloaded',
        status: 'success',
        duration: 3000
      });
    } catch (error: any) {
      console.error('Error exporting payments to PDF:', error);
      toast({
        title: 'Export failed',
        description: 'Failed to export payment report',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Refresh data handler
  const handleRefreshData = () => {
    loadPayments();
    loadPaymentSummary();
  };

  // Handle form save
  const handleFormSave = () => {
    setShowPaymentForm(false);
    setSelectedPayment(null);
    loadPayments();
  };

  // Handle form cancel
  const handleFormCancel = () => {
    setShowPaymentForm(false);
    setSelectedPayment(null);
    setError(null);
  };
  
  if (loading && (!payments || payments.length === 0)) {
    return (
      <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'employee', 'inventory_manager']}>
        <Box display="flex" justifyContent="center" alignItems="center" height="400px">
          <Spinner size="xl" thickness="4px" speed="0.65s" color="brand.500" />
          <Text ml={4} fontSize="lg">Loading payments...</Text>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'employee', 'inventory_manager']}>
      <Box bg={bgColor} minH="100vh" p={6}>
        <VStack spacing={6} align="stretch">
        {/* Summary Cards */}
        {summary && (
          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
            <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel>Total Payments</StatLabel>
                  <StatNumber>{summary.total_payments}</StatNumber>
                  <StatHelpText>This period</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel>Total Amount</StatLabel>
                  <StatNumber>{formatCurrency(summary.total_amount)}</StatNumber>
                  <StatHelpText>Gross amount</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel>Completed</StatLabel>
                  <StatNumber color={statColors.green}>
                    {formatCurrency(summary.completed_amount)}
                  </StatNumber>
                  <StatHelpText>{summary.completed_count} payments</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel>Avg Payment Value</StatLabel>
                  <StatNumber>{formatCurrency(summary.avg_payment_value)}</StatNumber>
                  <StatHelpText>Per transaction</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </SimpleGrid>
        )}

        {/* Header */}
        <Flex justify="space-between" align="center">
          <Box>
            <Heading as="h1" size="xl" mb={2} color={headingColor}>Payment Management</Heading>
            <Text color={textSecondary}>Manage your payment transactions</Text>
          </Box>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              variant="ghost"
              size="md"
              onClick={handleRefreshData}
              isLoading={loading}
            >
              Refresh
            </Button>
            {canExport && (
              <Menu>
                <MenuButton 
                  as={Button} 
                  leftIcon={<FiDownload />}
                  colorScheme="green"
                  variant="outline"
                  size="md"
                >
                  Export Report
                </MenuButton>
                <MenuList>
                  <MenuItem icon={<FiFileText />} onClick={handleBulkExport}>
                    Export PDF Report
                  </MenuItem>
                  <MenuItem icon={<FiDownload />} onClick={async () => {
                    try {
                      // Prepare current filters for export
                      const exportFilters: any = {};
                      if (statusFilter !== 'ALL') exportFilters.status = statusFilter;
                      if (methodFilter !== 'ALL') exportFilters.method = methodFilter;
                      if (startDate) exportFilters.start_date = startDate;
                      if (endDate) exportFilters.end_date = endDate;
                      
                      // Call Excel export API
                      await paymentService.downloadPaymentReportExcel(
                        startDate || undefined,
                        endDate || undefined,
                        exportFilters.status,
                        exportFilters.method
                      );
                      
                      toast({
                        title: 'Success',
                        description: 'Payment report has been downloaded as Excel',
                        status: 'success',
                        duration: 3000
                      });
                    } catch (error: any) {
                      console.error('Error exporting payments to Excel:', error);
                      toast({
                        title: 'Export failed',
                        description: 'Failed to export payment report as Excel',
                        status: 'error',
                        duration: 3000
                      });
                    }
                  }}>
                    Export Excel Report
                  </MenuItem>
                </MenuList>
              </Menu>
            )}
            {canCreate && (
              <Menu>
                <MenuButton 
                  as={Button} 
                  leftIcon={<FiPlus />}
                  rightIcon={<FiChevronDown />}
                  colorScheme="blue"
                  size="md"
                >
                  Create Payment
                </MenuButton>
                <MenuList>
                  <MenuItem 
                    icon={<FiArrowDown />} 
                    onClick={() => handleNewPayment('receivable')}
                  >
                    Receivable Payment (from Customer)
                  </MenuItem>
                  <MenuItem 
                    icon={<FiArrowRight />} 
                    onClick={() => handleNewPayment('payable')}
                  >
                    Payable Payment (to Vendor)
                  </MenuItem>
                  <MenuDivider />
                  <MenuItem 
                    icon={<FiDollarSign />} 
                    onClick={() => handlePPNPayment()}
                  >
                    Setor PPN (Tax Remittance)
                  </MenuItem>
                </MenuList>
              </Menu>
            )}
          </HStack>
        </Flex>

        {/* Search and Filters */}
        <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
          <CardBody>
            <HStack spacing={4} wrap="wrap">
              <InputGroup maxW="400px">
                <InputLeftElement pointerEvents="none">
                  <FiSearch color={searchIconColor} />
                </InputLeftElement>
                <Input 
                  placeholder="Search by payment code or contact..."
                  value={searchInput}
                  onChange={(e) => handleSearch(e.target.value)}
                  bg={inputBg}
                  borderColor={borderColor}
                />
              </InputGroup>
              
              <Select 
                maxW="200px" 
                placeholder="All Status"
                value={statusFilter}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                bg={inputBg}
                borderColor={borderColor}
              >
                <option value="PENDING">Pending</option>
                <option value="COMPLETED">Completed</option>
                <option value="FAILED">Failed</option>
              </Select>
              
              <Select
                maxW="200px"
                placeholder="All Methods"
                value={methodFilter}
                onChange={(e) => handleFilterChange('method', e.target.value)}
                bg={inputBg}
                borderColor={borderColor}
              >
                <option value="CASH">Cash</option>
                <option value="BANK_TRANSFER">Bank Transfer</option>
                <option value="CHECK">Check</option>
                <option value="CREDIT_CARD">Credit Card</option>
                <option value="DEBIT_CARD">Debit Card</option>
              </Select>
              
              <Input
                type="date"
                maxW="200px"
                placeholder="Start Date"
                value={startDate}
                onChange={(e) => handleFilterChange('start_date', e.target.value)}
                bg={inputBg}
                borderColor={borderColor}
              />
              
              <Input
                type="date"
                maxW="200px"
                placeholder="End Date"
                value={endDate}
                onChange={(e) => handleFilterChange('end_date', e.target.value)}
                bg={inputBg}
                borderColor={borderColor}
              />
              
              <Button
                leftIcon={<FiFilter />}
                variant="outline"
                onClick={resetFilters}
                colorScheme="gray"
              >
                Clear Filters
              </Button>
            </HStack>
          </CardBody>
        </Card>

        {/* Error Alert */}
        {error && (
          <Alert status="error">
            <AlertIcon />
            {error}
          </Alert>
        )}

        {/* Payments Table */}
        <Card bg={cardBg} borderWidth="1px" borderColor={borderColor}>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="md" color={headingColor}>Payment Transactions ({payments?.length || 0})</Heading>
            </Flex>
          </CardHeader>
          <CardBody>
            {loading ? (
              <Flex justify="center" py={10}>
                <Spinner size="lg" />
              </Flex>
            ) : (
              <DataTable 
                columns={columns} 
                data={payments || []} 
                keyField="id"
                searchable={false}
                pagination={true}
                pageSize={ITEMS_PER_PAGE}
              />
            )}
          </CardBody>
        </Card>
      </VStack>

      {/* Payment Form Modal */}
      <AdvancedPaymentForm
        isOpen={showPaymentForm}
        onClose={handleFormCancel}
        type={paymentType}
        onSuccess={handleFormSave}
        preSelectedContact={selectedPayment ? { 
          id: selectedPayment.contact_id,
          name: selectedPayment.contact?.name || 'Unknown'
        } : undefined}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        isOpen={showConfirmDelete}
        leastDestructiveRef={undefined}
        onClose={() => setShowConfirmDelete(false)}
      >
        <AlertDialogOverlay>
          <AlertDialogContent bg={cardBg}>
            <AlertDialogHeader fontSize="lg" fontWeight="bold" color={headingColor}>
              Delete Payment
            </AlertDialogHeader>

            <AlertDialogBody color={textPrimary}>
              Are you sure you want to delete this payment?
              {selectedPayment && (
                <Box mt={3} p={3} bg={alertBg} borderRadius="md">
                  <Text fontSize="sm" fontWeight="bold" color={textPrimary}>Payment: {selectedPayment.code}</Text>
                  <Text fontSize="sm" color={textPrimary}>Amount: {formatCurrency(selectedPayment.amount)}</Text>
                  <Text fontSize="sm" color={alertTextColor}>This action cannot be undone.</Text>
                </Box>
              )}
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button onClick={() => setShowConfirmDelete(false)}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={confirmDeletePayment} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
      {/* Payment Detail Modal */}
      <PaymentDetailModal
        payment={selectedPayment}
        isOpen={showPaymentDetail}
        onClose={() => {
          setShowPaymentDetail(false);
          setSelectedPayment(null);
        }}
      />
      
      {/* PPN Payment Modal */}
      <PPNPaymentModal
        isOpen={showPPNPayment}
        onClose={() => setShowPPNPayment(false)}
        ppnType={ppnPaymentType}
        onSuccess={() => {
          setShowPPNPayment(false);
          loadPayments();
          toast({
            title: 'Success',
            description: 'PPN remittance (Setor PPN) has been processed successfully',
            status: 'success',
            duration: 3000,
          });
        }}
      />
      </Box>
    </SimpleLayout>
  );
};

export default PaymentsPage;
