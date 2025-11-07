'use client';

import React, { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import UnifiedLayout from '@/components/layout/UnifiedLayout';
import salesService, { Sale } from '@/services/salesService';
import PaymentForm from '@/components/sales/PaymentForm';
import {
  Box,
  Heading,
  Text,
  Button,
  Flex,
  HStack,
  VStack,
  Card,
  CardHeader,
  CardBody,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Divider,
  Grid,
  GridItem,
  Spinner,
  Alert,
  AlertIcon,
  useToast,
  useDisclosure,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem
} from '@chakra-ui/react';
import {
  FiArrowLeft,
  FiEdit,
  FiDollarSign,
  FiDownload,
  FiMoreVertical,
  FiCheck,
  FiX,
  FiFileText
} from 'react-icons/fi';

const SaleDetailPage: React.FC = () => {
  const params = useParams();
  const router = useRouter();
  const toast = useToast();
  const { isOpen: isPaymentOpen, onOpen: onPaymentOpen, onClose: onPaymentClose } = useDisclosure();

  const [sale, setSale] = useState<Sale | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const saleId = params?.id as string;

  // Handle back navigation - simplified and more reliable
  const handleGoBack = () => {
    console.log('Back button clicked');
    // Direct navigation to sales page - most reliable
    router.push('/sales');
  };

  // Load sale data
  const loadSale = async () => {
    if (!saleId) return;

    try {
      setLoading(true);
      setError(null);
      const saleData = await salesService.getSale(parseInt(saleId));
      setSale(saleData);
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to load sale details');
      toast({
        title: 'Error loading sale',
        description: error.response?.data?.message || 'Failed to load sale details',
        status: 'error',
        duration: 3000
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSale();
  }, [saleId]);

  // Handle status actions
  const handleStatusAction = async (action: 'confirm' | 'cancel') => {
    if (!sale) return;

    try {
      setActionLoading(true);
      
      if (action === 'confirm') {
        await salesService.confirmSale(sale.id);
        toast({
          title: 'Sale Confirmed',
          description: 'Sale has been confirmed successfully',
          status: 'success',
          duration: 3000
        });
      } else if (action === 'cancel') {
        const reason = window.prompt('Please provide a reason for cancellation:');
        if (reason) {
          await salesService.cancelSale(sale.id, reason);
          toast({
            title: 'Sale Cancelled',
            description: 'Sale has been cancelled successfully',
            status: 'success',
            duration: 3000
          });
        } else {
          return;
        }
      }

      loadSale(); // Reload to get updated status
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || `Failed to ${action} sale`,
        status: 'error',
        duration: 3000
      });
    } finally {
      setActionLoading(false);
    }
  };

  // Handle create invoice
  const handleCreateInvoice = async () => {
    if (!sale) return;

    try {
      setActionLoading(true);
      await salesService.createInvoiceFromSale(sale.id);
      toast({
        title: 'Invoice Created',
        description: 'Invoice has been created successfully',
        status: 'success',
        duration: 3000
      });
      loadSale(); // Reload to get updated data
    } catch (error: any) {
      toast({
        title: 'Error creating invoice',
        description: error.response?.data?.message || 'Failed to create invoice',
        status: 'error',
        duration: 3000
      });
    } finally {
      setActionLoading(false);
    }
  };

  // Handle payment form save
  const handlePaymentSave = () => {
    onPaymentClose();
    loadSale(); // Reload to get updated payment status
  };

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'paid': return 'green';
      case 'invoiced': return 'blue';
      case 'confirmed': return 'purple';
      case 'overdue': return 'red';
      case 'draft': return 'gray';
      case 'cancelled': return 'red';
      default: return 'gray';
    }
  };

  // Show loading state
  if (loading) {
    return (
      <UnifiedLayout>
        <Flex justify="center" align="center" minH="400px">
          <Spinner size="xl" />
        </Flex>
      </UnifiedLayout>
    );
  }

  // Show error state
  if (error || !sale) {
    return (
      <UnifiedLayout>
        <Alert status="error">
          <AlertIcon />
          {error || 'Sale not found'}
        </Alert>
      </UnifiedLayout>
    );
  }

  return (
    <UnifiedLayout>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Flex justify="space-between" align="center">
          <HStack spacing={4}>
            <IconButton
              icon={<FiArrowLeft />}
              variant="outline"
              onClick={handleGoBack}
              aria-label="Go back to Sales"
              _hover={{ 
                bg: 'var(--bg-tertiary)', 
                transform: 'translateX(-2px)',
                borderColor: 'var(--accent-color)'
              }}
              size="md"
              title="Go back to Sales"
              borderColor="var(--border-color)"
              color="var(--text-primary)"
              bg="var(--bg-secondary)"
              cursor="pointer"
            />
            <VStack align="start" spacing={1}>
              <Heading as="h1" size="xl">Sale Detail</Heading>
              <HStack spacing={2}>
                <Text color="gray.600">Code: {sale.code}</Text>
                <Badge colorScheme={getStatusColor(sale.status)} variant="subtle">
                  {salesService.getStatusLabel(sale.status)}
                </Badge>
              </HStack>
            </VStack>
          </HStack>
          
          <HStack spacing={3}>
            <Button
              leftIcon={<FiDownload />}
              variant="outline"
              onClick={() => salesService.downloadInvoicePDF(sale.id, sale.invoice_number || sale.code)}
            >
              Download PDF
            </Button>
            
            <Menu>
              <MenuButton as={IconButton} icon={<FiMoreVertical />} variant="outline" />
              <MenuList>
                {sale.status === 'DRAFT' && (
                  <MenuItem 
                    icon={<FiCheck />} 
                    onClick={() => handleStatusAction('confirm')}
                    isDisabled={actionLoading}
                  >
                    Confirm Sale
                  </MenuItem>
                )}
                {sale.status === 'CONFIRMED' && (
                  <MenuItem 
                    icon={<FiFileText />} 
                    onClick={handleCreateInvoice}
                    isDisabled={actionLoading}
                  >
                    Create Invoice
                  </MenuItem>
                )}
                {sale.outstanding_amount > 0 && (
                  <MenuItem 
                    icon={<FiDollarSign />} 
                    onClick={onPaymentOpen}
                  >
                    Record Payment
                  </MenuItem>
                )}
                <MenuItem 
                  icon={<FiEdit />} 
                  onClick={() => router.push(`/sales?edit=${sale.id}`)}
                >
                  Edit Sale
                </MenuItem>
                <Divider />
                {sale.status !== 'CANCELLED' && (
                  <MenuItem 
                    icon={<FiX />} 
                    color="red.500"
                    onClick={() => handleStatusAction('cancel')}
                    isDisabled={actionLoading}
                  >
                    Cancel Sale
                  </MenuItem>
                )}
              </MenuList>
            </Menu>
          </HStack>
        </Flex>


        {/* Basic Information */}
        <Card>
          <CardHeader>
            <Heading size="md">Sale Information</Heading>
          </CardHeader>
          <CardBody>
            <Grid templateColumns="repeat(3, 1fr)" gap={6}>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Customer</Text>
                  <Text fontWeight="medium">{sale.customer ? sale.customer.name : 'N/A'}</Text>
                </VStack>
              </GridItem>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Invoice Number</Text>
                  <Text fontWeight="medium">{sale.invoice_number ? sale.invoice_number : '-'}</Text>
                </VStack>
              </GridItem>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Sales Person</Text>
                  <Text fontWeight="medium">{sale.sales_person ? sale.sales_person.name : 'N/A'}</Text>
                </VStack>
              </GridItem>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Date</Text>
                  <Text fontWeight="medium">{salesService.formatDate(sale.date)}</Text>
                </VStack>
              </GridItem>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Due Date</Text>
                  <Text fontWeight="medium">
                    {sale.due_date && sale.due_date !== '0001-01-01T00:00:00Z' ? salesService.formatDate(sale.due_date) : '-'}
                  </Text>
                </VStack>
              </GridItem>
              <GridItem>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">Payment Terms</Text>
                  <Text fontWeight="medium">{sale.payment_terms ? sale.payment_terms : 'N/A'}</Text>
                </VStack>
              </GridItem>
            </Grid>
          </CardBody>
        </Card>

        {/* Items */}
        <Card>
          <CardHeader>
            <Heading size="md">Sale Items</Heading>
          </CardHeader>
          <CardBody>
            <TableContainer>
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th>Product</Th>
                    <Th>Description</Th>
                    <Th isNumeric>Quantity</Th>
                    <Th isNumeric>Unit Price</Th>
                    <Th isNumeric>Discount</Th>
                    <Th isNumeric>Total</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {sale.sale_items?.map((item, index) => (
                    <Tr key={index}>
                      <Td fontWeight="medium">
                        {item.product?.name || 'N/A'}
                      </Td>
                      <Td>{item.description || '-'}</Td>
                      <Td isNumeric>{item.quantity || 0}</Td>
                      <Td isNumeric>{salesService.formatCurrency(item.unit_price || 0)}</Td>
                      <Td isNumeric>
                        {item.discount_percent ? `${item.discount_percent}%` : '-'}
                      </Td>
                      <Td isNumeric fontWeight="medium">
                        {salesService.formatCurrency(item.line_total || item.total_price)}
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </TableContainer>
          </CardBody>
        </Card>

        {/* Financial Summary */}
        <Card>
          <CardHeader>
            <Heading size="md">Financial Summary</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Flex justify="space-between">
                <Text>Subtotal:</Text>
                <Text fontWeight="medium">
                  {salesService.formatCurrency(sale.subtotal || 0)}
                </Text>
              </Flex>
              <Flex justify="space-between">
                <Text>Discount ({sale.discount_percent || 0}%):</Text>
                <Text fontWeight="medium">
                  {salesService.formatCurrency(sale.discount_amount || 0)}
                </Text>
              </Flex>
              <Flex justify="space-between">
                <Text>Shipping Cost:</Text>
                <Text fontWeight="medium">
                  {salesService.formatCurrency(sale.shipping_cost || 0)}
                </Text>
              </Flex>
              <Flex justify="space-between">
                <Text>PPN ({sale.ppn_percent || 0}%):</Text>
                <Text fontWeight="medium">
                  {salesService.formatCurrency(sale.ppn_amount || sale.ppn || 0)}
                </Text>
              </Flex>
              <Divider />
              <Flex justify="space-between" fontSize="lg">
                <Text fontWeight="bold">Total Amount:</Text>
                <Text fontWeight="bold" color="blue.600">
                  {salesService.formatCurrency(sale.total_amount || 0)}
                </Text>
              </Flex>
              <Flex justify="space-between">
                <Text color="green.600">Paid Amount:</Text>
                <Text fontWeight="medium" color="green.600">
                  {salesService.formatCurrency(sale.paid_amount || 0)}
                </Text>
              </Flex>
              <Flex justify="space-between" fontSize="lg">
                <Text fontWeight="bold" color="orange.600">Outstanding:</Text>
                <Text fontWeight="bold" color="orange.600">
                  {salesService.formatCurrency(sale.outstanding_amount || 0)}
                </Text>
              </Flex>
            </VStack>
          </CardBody>
        </Card>

        {/* Notes */}
        {(sale.notes || sale.internal_notes) && (
          <Card>
            <CardHeader>
              <Heading size="md">Notes</Heading>
            </CardHeader>
            <CardBody>
              <VStack spacing={4} align="stretch">
                {sale.notes && (
                  <Box>
                    <Text fontSize="sm" color="gray.600" mb={2}>Customer Notes:</Text>
                    <Text>{sale.notes}</Text>
                  </Box>
                )}
                {sale.internal_notes && (
                  <Box>
                    <Text fontSize="sm" color="gray.600" mb={2}>Internal Notes:</Text>
                    <Text>{sale.internal_notes}</Text>
                  </Box>
                )}
              </VStack>
            </CardBody>
          </Card>
        )}
      </VStack>

      {/* Payment Form Modal */}
      <PaymentForm
        isOpen={isPaymentOpen}
        onClose={onPaymentClose}
        onSave={handlePaymentSave}
        sale={sale}
      />
    </UnifiedLayout>
  );
};

export default SaleDetailPage;
