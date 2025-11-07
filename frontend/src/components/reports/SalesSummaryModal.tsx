import React, { useState } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Box,
  Text,
  VStack,
  HStack,
  Button,
  SimpleGrid,
  Badge,
  Flex,
  useColorModeValue,
  Grid,
  GridItem,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  useToast,
  Spinner,
  Icon
} from '@chakra-ui/react';
import { FiDownload, FiShoppingCart, FiDollarSign, FiPieChart, FiUsers, FiTrendingUp, FiFilePlus, FiFileText } from 'react-icons/fi';
import { 
  FormControl,
  FormLabel,
  Input
} from '@chakra-ui/react';
import { formatCurrency } from '../../utils/formatters';
import { SSOTSalesSummaryData, CustomerSalesData, ProductSalesData } from '../../services/ssotSalesSummaryService';

interface SalesSummaryModalProps {
  isOpen: boolean;
  onClose: () => void;
  data: SSOTSalesSummaryData | null;
  isLoading: boolean;
  error: string | null;
  startDate: string;
  endDate: string;
  onDateChange?: (startDate: string, endDate: string) => void;
  onFetch?: () => void;
  onExport?: (format: 'pdf' | 'excel') => void;
}

const SalesSummaryModal: React.FC<SalesSummaryModalProps> = ({
  isOpen,
  onClose,
  data,
  isLoading,
  error,
  startDate,
  endDate,
  onDateChange,
  onFetch,
  onExport
}) => {
  const [activeTab, setActiveTab] = useState<'summary'>('summary');
  const toast = useToast();
  
  // Color mode values
  const modalBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const sectionBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');
  const secondaryTextColor = useColorModeValue('gray.600', 'gray.300');
  const loadingTextColor = useColorModeValue('gray.700', 'gray.300');
  const previewPeriodTextColor = useColorModeValue('gray.500', 'gray.400');
  
  const handleExport = (format: 'pdf' | 'excel') => {
    if (onExport) {
      onExport(format);
    } else {
      if (data) {
        // Fallback: download as JSON
        const reportData = {
          reportType: 'Sales Summary',
          period: `${startDate} to ${endDate}`,
          generatedOn: new Date().toISOString(),
          data: data
        };
        const dataStr = JSON.stringify(reportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `sales-summary-${startDate}-to-${endDate}.json`;
        link.click();
        URL.revokeObjectURL(url);
      }
      
      toast({
        title: 'Export Feature',
        description: `${format.toUpperCase()} export will be implemented soon`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  const renderSummaryMetrics = () => {
    if (!data) return null;
    
    // Calculate total customers from sales_by_customer array
    const totalCustomers = data.sales_by_customer ? data.sales_by_customer.length : 0;
    
    // Calculate total orders from transaction counts
    const totalOrders = data.sales_by_customer 
      ? data.sales_by_customer.reduce((sum, customer) => sum + (customer.transaction_count || 0), 0)
      : 0;
    
    // Calculate average order value
    const averageOrderValue = totalOrders > 0 ? (data.total_revenue || data.total_sales || 0) / totalOrders : 0;
    
    return (
      <Grid templateColumns="repeat(auto-fit, minmax(240px, 1fr))" gap={4} mb={6}>
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Revenue</StatLabel>
                <StatNumber color="green.600">
                  {formatCurrency(data.total_revenue || data.total_sales || 0)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type="increase" />
                  Sales for the period
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Customers</StatLabel>
                <StatNumber color="blue.600">
                  {totalCustomers}
                </StatNumber>
                <StatHelpText>
                  <Icon as={FiUsers} />
                  Active customers
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        {totalOrders > 0 && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Total Orders</StatLabel>
                  <StatNumber color="purple.600">
                    {totalOrders}
                  </StatNumber>
                  <StatHelpText>
                    <Icon as={FiShoppingCart} />
                    Orders processed
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        )}
        
        {averageOrderValue > 0 && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Average Order Value</StatLabel>
                  <StatNumber color="orange.600">
                    {formatCurrency(averageOrderValue)}
                  </StatNumber>
                  <StatHelpText>
                    <Icon as={FiTrendingUp} />
                    Per order average
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        )}
      </Grid>
    );
  };

  // Note: This function is currently unused as the Customers tab has been removed
  const renderCustomersTab = () => {
    if (!data) return null;

    const customers = data.sales_by_customer || [];
    // For top customers, we'll sort the sales_by_customer by total_sales
    const topCustomers = [...customers]
      .sort((a, b) => (b.total_sales || 0) - (a.total_sales || 0))
      .slice(0, 6);

    return (
      <VStack spacing={6} align="stretch">
        {customers.length > 0 && (
          <Card>
            <CardHeader>
              <Heading size="sm">Sales by Customer</Heading>
            </CardHeader>
            <CardBody>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>Customer</Th>
                    <Th isNumeric>Total Sales</Th>
                    <Th isNumeric>Orders</Th>
                    <Th isNumeric>Avg Order</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {customers.map((customer: CustomerSalesData, index: number) => (
                    <Tr key={index}>
                      <Td>
                        <VStack align="start" spacing={1}>
                          <Text fontWeight="medium">
                            {customer.customer_name || 'Unnamed Customer'}
                          </Text>
                        </VStack>
                      </Td>
                      <Td isNumeric>
                        <Text fontWeight="bold" color="green.600">
                          {formatCurrency(customer.total_sales || 0)}
                        </Text>
                      </Td>
                      <Td isNumeric>
                        <Text color="purple.600">
                          {customer.transaction_count || 0}
                        </Text>
                      </Td>
                      <Td isNumeric>
                        <Text color="orange.600">
                          {formatCurrency(
                            customer.average_transaction ||
                            (customer.total_sales && customer.transaction_count > 0 
                              ? customer.total_sales / customer.transaction_count 
                              : 0)
                          )}
                        </Text>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>
        )}

        {topCustomers.length > 0 && (
          <Card>
            <CardHeader>
              <Heading size="sm">Top Performing Customers</Heading>
            </CardHeader>
            <CardBody>
              <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                {topCustomers.map((customer: CustomerSalesData, index: number) => (
                  <Box key={index} border="1px" borderColor={borderColor} borderRadius="md" p={4}>
                    <VStack spacing={3}>
                      <Badge colorScheme="gold" size="lg" variant="solid">
                        #{index + 1}
                      </Badge>
                      <Text fontWeight="bold" fontSize="md" textAlign="center">
                        {customer.customer_name}
                      </Text>
                      <Text fontSize="lg" fontWeight="bold" color="green.600">
                        {formatCurrency(customer.total_sales)}
                      </Text>
                      <Text fontSize="sm" color="gray.500">
                        {((customer.total_sales / (data.total_revenue || data.total_sales || 1)) * 100).toFixed(1)}% of total
                      </Text>
                      {customer.transaction_count && (
                        <Text fontSize="xs" color="purple.500">
                          {customer.transaction_count} orders
                        </Text>
                      )}
                    </VStack>
                  </Box>
                ))}
              </SimpleGrid>
            </CardBody>
          </Card>
        )}

        {customers.length === 0 && topCustomers.length === 0 && (
          <Box textAlign="center" py={8}>
            <Text color={secondaryTextColor}>
              No customer data available for this period
            </Text>
          </Box>
        )}
      </VStack>
    );
  };

  // Note: This function is currently unused as the Analysis tab has been removed
  const renderAnalysisTab = () => {
    if (!data) return null;

    // Calculate some basic trends from the data
    const totalCustomers = data.sales_by_customer ? data.sales_by_customer.length : 0;
    const totalOrders = data.sales_by_customer 
      ? data.sales_by_customer.reduce((sum, customer) => sum + (customer.transaction_count || 0), 0)
      : 0;
    
    // Calculate growth rate (simplified)
    const growthRate = totalOrders > 0 ? ((data.total_revenue || data.total_sales || 0) / totalOrders) : 0;
    
    return (
      <VStack spacing={6} align="stretch">
        <Card>
          <CardHeader>
            <Heading size="sm">Sales Performance Analysis</Heading>
          </CardHeader>
          <CardBody>
            <SimpleGrid columns={[1, 2, 4]} spacing={4}>
              <Box p={4} bg="green.50" borderRadius="md" textAlign="center">
                <Text fontSize="2xl" fontWeight="bold" color="green.600">
                  {growthRate.toFixed(1)}%
                </Text>
                <Text fontSize="sm" color="green.800">
                  Growth Rate
                </Text>
              </Box>
              
              <Box p={4} bg="blue.50" borderRadius="md" textAlign="center">
                <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                  {totalCustomers}
                </Text>
                <Text fontSize="sm" color="blue.800">
                  Total Customers
                </Text>
              </Box>
              
              <Box p={4} bg="purple.50" borderRadius="md" textAlign="center">
                <Text fontSize="2xl" fontWeight="bold" color="purple.600">
                  {totalOrders}
                </Text>
                <Text fontSize="sm" color="purple.800">
                  Total Orders
                </Text>
              </Box>
              
              <Box p={4} bg="teal.50" borderRadius="md" textAlign="center">
                <Text fontSize="2xl" fontWeight="bold" color="teal.600">
                  {data.sales_by_customer && data.sales_by_customer.length > 0 
                    ? Math.round(data.sales_by_customer.reduce((sum, customer) => sum + (customer.transaction_count || 0), 0) / data.sales_by_customer.length)
                    : 0}
                </Text>
                <Text fontSize="sm" color="teal.800">
                  Avg Orders/Customer
                </Text>
              </Box>
            </SimpleGrid>
          </CardBody>
        </Card>
        
        <Card>
          <CardHeader>
            <Heading size="sm">Key Insights</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              <Text fontSize="sm">
                📈 <strong>Sales Performance:</strong> {growthRate > 0 ? 'Your sales are growing positively' : 'Sales performance shows room for improvement'}
              </Text>
              <Text fontSize="sm">
                👥 <strong>Customer Base:</strong> {totalCustomers} active customers indicate {totalCustomers > 10 ? 'strong' : 'developing'} customer loyalty
              </Text>
              <Text fontSize="sm">
                🎯 <strong>Market Expansion:</strong> {totalOrders} total orders shows business activity level
              </Text>
            </VStack>
          </CardBody>
        </Card>
      </VStack>
    );
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent bg={modalBg} maxW="90vw">
        <ModalHeader>
          <HStack>
            <Icon as={FiShoppingCart} color="blue.500" />
            <VStack align="start" spacing={0}>
              <Text fontSize="lg" fontWeight="bold">
                Sales Summary Report (SSOT)
              </Text>
              <Text fontSize="sm" color={previewPeriodTextColor}>
                {startDate} - {endDate} | SSOT Journal Integration
              </Text>
            </VStack>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody pb={6} px={8}>
          {/* Date Range Controls - Moved to top like other modals */}
          <Box mb={4}>
            <HStack spacing={4} mb={4} flexWrap="wrap">
              <FormControl>
                <FormLabel>Start Date</FormLabel>
                <Input 
                  type="date" 
                  value={startDate} 
                  onChange={(e) => onDateChange && onDateChange(e.target.value, endDate)} 
                />
              </FormControl>
              <FormControl>
                <FormLabel>End Date</FormLabel>
                <Input 
                  type="date" 
                  value={endDate} 
                  onChange={(e) => onDateChange && onDateChange(startDate, e.target.value)} 
                />
              </FormControl>
              <Button
                colorScheme="blue"
                onClick={onFetch}
                isLoading={isLoading}
                leftIcon={<FiShoppingCart />}
                size="md"
                mt={8}
                whiteSpace="nowrap"
              >
                Generate Report
              </Button>
            </HStack>
          </Box>

          {isLoading && (
            <Box textAlign="center" py={8}>
              <VStack spacing={4}>
                <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                <VStack spacing={2}>
                  <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                    Generating Sales Summary Report
                  </Text>
                  <Text fontSize="sm" color={secondaryTextColor}>
                    Analyzing sales transactions from SSOT journal system...
                  </Text>
                </VStack>
              </VStack>
            </Box>
          )}

          {error && (
            <Box bg="red.50" p={4} borderRadius="md" mb={4}>
              <Text color="red.600" fontWeight="medium">Error: {error}</Text>
              <Button
                mt={2}
                size="sm"
                colorScheme="red"
                variant="outline"
                onClick={onFetch}
              >
                Retry
              </Button>
            </Box>
          )}

          {data && !isLoading && (
            <VStack spacing={6} align="stretch">
              {/* Company Header */}
              {data.company && (
                <Box bg="blue.50" p={4} borderRadius="md">
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color="blue.800">
                        {data.company.name || 'Company Name Not Set'}
                      </Text>
                      <Text fontSize="sm" color="blue.600">
                        {data.company.address ? (
                          data.company.city ? `${data.company.address}, ${data.company.city}` : data.company.address
                        ) : 'Address not available'}
                      </Text>
                      {data.company.phone && (
                        <Text fontSize="sm" color="blue.600">
                          {data.company.phone} | {data.company.email}
                        </Text>
                      )}
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color="blue.600">
                        Currency: {data.currency || 'IDR'}
                      </Text>
                      <Text fontSize="xs" color="blue.500">
                        Generated: {data.generated_at ? new Date(data.generated_at).toLocaleString('id-ID') : new Date().toLocaleString('id-ID')}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>
              )}

              {/* Report Header */}
              <Box textAlign="center" bg={sectionBg} p={4} borderRadius="md">
                <Heading size="md" color={textColor}>
                  Sales Summary Report
                </Heading>
                <Text fontSize="sm" color={secondaryTextColor}>
                  Period: {startDate} - {endDate}
                </Text>
                <Text fontSize="xs" color={secondaryTextColor} mt={1}>
                  Generated: {new Date().toLocaleDateString('id-ID')} at {new Date().toLocaleTimeString('id-ID')}
                </Text>
              </Box>

              {renderSummaryMetrics()}
              
              {/* Period Summary */}
              <Card>
                <CardBody>
                  <Flex justify="space-between" align="center" mb={3}>
                    <Heading size="md" color={textColor}>
                      Sales Performance
                    </Heading>
                    <Text fontWeight="bold" fontSize="lg" color="green.600">
                      {formatCurrency(data.total_revenue || data.total_sales || 0)}
                    </Text>
                  </Flex>
                  
                  <SimpleGrid columns={[1, 3]} spacing={4}>
                    <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                      <Text fontSize="sm" color={secondaryTextColor}>Period</Text>
                      <Text fontWeight="medium">{startDate} to {endDate}</Text>
                    </Box>
                    <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                      <Text fontSize="sm" color={secondaryTextColor}>Report Type</Text>
                      <Text fontWeight="medium">SSOT Integration</Text>
                    </Box>
                    <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                      <Text fontSize="sm" color={secondaryTextColor}>Status</Text>
                      <Badge colorScheme="green">Active</Badge>
                    </Box>
                  </SimpleGrid>
                </CardBody>
              </Card>

              {/* Top Customers Section */}
              {data.sales_by_customer && data.sales_by_customer.length > 0 && (
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Top Performing Customers ({Math.min(data.sales_by_customer.length, 6)} customers)
                  </Heading>
                  <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                    {[...data.sales_by_customer]
                      .sort((a, b) => (b.total_sales || 0) - (a.total_sales || 0))
                      .slice(0, 6)
                      .map((customer: CustomerSalesData, index: number) => (
                        <Box key={index} border="1px" borderColor={borderColor} borderRadius="md" p={4} bg="white">
                          <VStack spacing={3}>
                            <Badge colorScheme="blue" size="lg" variant="solid">
                              #{index + 1}
                            </Badge>
                            <Text fontWeight="bold" fontSize="md" textAlign="center" color="gray.800">
                              {customer.customer_name}
                            </Text>
                            <Text fontSize="lg" fontWeight="bold" color="green.600">
                              {formatCurrency(customer.total_sales)}
                            </Text>
                            <Text fontSize="sm" color="gray.600">
                              {((customer.total_sales / (data.total_revenue || data.total_sales || 1)) * 100).toFixed(1)}% of total
                            </Text>
                            {customer.transaction_count && (
                              <Text fontSize="xs" color="purple.600">
                                {customer.transaction_count} orders
                              </Text>
                            )}
                          </VStack>
                        </Box>
                      ))}
                  </SimpleGrid>
                </Box>
              )}

              {/* Sales by Customer Table */}
              {data.sales_by_customer && data.sales_by_customer.length > 0 && (
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Sales by Customer ({data.sales_by_customer.length} customers)
                  </Heading>
                  
                  {/* Customer Table Header */}
                  <Box bg="blue.50" p={3} borderRadius="md" mb={2} border="1px solid" borderColor="blue.200">
                    <SimpleGrid columns={[1, 3]} spacing={2} fontSize="sm" fontWeight="bold" color="blue.800">
                      <Text>Customer</Text>
                      <Text textAlign="right">Total Sales</Text>
                      <Text textAlign="right">Orders</Text>
                    </SimpleGrid>
                  </Box>
                  
                  {/* Customer Rows */}
                  <VStack spacing={2} align="stretch" maxH="400px" overflow="auto">
                    {data.sales_by_customer.map((customer: CustomerSalesData, index: number) => (
                      <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white" _hover={{ bg: 'gray.50' }}>
                        <SimpleGrid columns={[1, 3]} spacing={2} fontSize="sm">
                          <VStack align="start" spacing={1}>
                            <Text fontWeight="bold" fontSize="md" color="gray.800">
                              {customer.customer_name || 'Unnamed Customer'}
                            </Text>
                          </VStack>
                          <Text textAlign="right" fontSize="sm" fontWeight="bold" color="green.600">
                            {formatCurrency(customer.total_sales || 0)}
                          </Text>
                          <Text textAlign="right" fontSize="sm" fontWeight="medium" color="purple.600">
                            {customer.transaction_count || 0}
                          </Text>
                        </SimpleGrid>
                      </Box>
                    ))}
                  </VStack>
                </Box>
              )}
            </VStack>
          )}
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            {data && !isLoading && (
              <>
                <Button
                  colorScheme="red"
                  variant="outline"
                  size="sm"
                  leftIcon={<FiFilePlus />}
                  onClick={() => handleExport('pdf')}
                >
                  Export PDF
                </Button>
                <Button
                  colorScheme="green"
                  variant="outline"
                  size="sm"
                  leftIcon={<FiFileText />}
                  onClick={() => handleExport('excel')}
                >
                  Export CSV
                </Button>
              </>
            )}
          </HStack>
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default SalesSummaryModal;