'use client';

import React, { useState } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Card,
  CardBody,
  CardHeader,
  Icon,
  Badge,
  SimpleGrid,
  useToast,
  Spinner,
  FormControl,
  FormLabel,
  Input,
  Switch,
  Divider,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiBarChart,
  FiDownload,
  FiFileText,
  FiFilePlus,
  FiRefreshCw,
  FiSettings,
  FiCheckCircle,
  FiAlertTriangle,
  FiInfo,
} from 'react-icons/fi';

// Import services and utilities
import { 
  ssotBalanceSheetReportService, 
  SSOTBalanceSheetData 
} from '../../services/ssotBalanceSheetReportService';
import { 
  exportAndDownloadCSV, 
  exportAndDownloadPDF
} from '../../utils/balanceSheetExportClient';

interface EnhancedBalanceSheetReportProps {
  onClose?: () => void;
}

export const EnhancedBalanceSheetReport: React.FC<EnhancedBalanceSheetReportProps> = ({
  onClose
}) => {
  // State management
  const [balanceSheetData, setBalanceSheetData] = useState<SSOTBalanceSheetData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [asOfDate, setAsOfDate] = useState(new Date().toISOString().split('T')[0]);
  const [includeAccountDetails, setIncludeAccountDetails] = useState(true);
  const [companyName, setCompanyName] = useState('Company Name Not Set');
  
  // Export states
  const [exportingCSV, setExportingCSV] = useState(false);
  const [exportingPDF, setExportingPDF] = useState(false);
  
  // Modal controls
  const { 
    isOpen: isExportModalOpen, 
    onOpen: onExportModalOpen, 
    onClose: onExportModalClose 
  } = useDisclosure();
  
  const toast = useToast();
  
  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headingColor = useColorModeValue('gray.700', 'white');
  const textColor = useColorModeValue('gray.800', 'white');
  const descriptionColor = useColorModeValue('gray.600', 'gray.300');

  // Format currency for display with account type consideration
  const formatCurrency = (amount: number | null | undefined, accountCode?: string, accountType?: string) => {
    if (amount === null || amount === undefined || isNaN(Number(amount))) {
      return 'Rp 0';
    }
    
    let displayAmount = Number(amount);
    
    // Convert to positive for revenue and liability accounts for user-friendly display
    if (accountType === 'REVENUE' || accountType === 'LIABILITY' || 
        accountCode?.startsWith('4') || // Revenue accounts
        accountCode?.startsWith('2103') || accountCode?.startsWith('2203')) { // PPN accounts
      displayAmount = Math.abs(displayAmount);
      console.log(`[EnhancedBalanceSheetReport] Converted ${accountCode} (${accountType}) balance ${amount} to positive ${displayAmount}`);
    }
    
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(displayAmount);
  };

  // Generate Balance Sheet Report
  const generateBalanceSheetReport = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await ssotBalanceSheetReportService.generateSSOTBalanceSheet({
        as_of_date: asOfDate,
        format: 'json'
      });
      
      setBalanceSheetData(data);
      
      toast({
        title: 'Success',
        description: 'Balance Sheet generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to generate Balance Sheet';
      setError(errorMessage);
      
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  // Export CSV function
  const handleExportCSV = async () => {
    if (!asOfDate) {
      toast({
        title: 'No Data',
        description: 'Please select a date first',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setExportingCSV(true);
    
    try {
      console.log('Exporting CSV with asOfDate:', asOfDate);
      // Use the new CSV generation method from the service
      const csvContent = await ssotBalanceSheetReportService.generateSSOTBalanceSheetCSV({
        as_of_date: asOfDate
      });
      
      console.log('Received CSV content, creating download link');
      // Create download link
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `SSOT_BalanceSheet_${asOfDate}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      toast({
        title: 'CSV Export Successful',
        description: 'Balance Sheet has been downloaded as CSV',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to export CSV';
      console.error('CSV Export Error:', error); // Log the error for debugging
      
      toast({
        title: 'CSV Export Failed',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setExportingCSV(false);
    }
  };

  // Export PDF function
  const handleExportPDF = async () => {
    if (!asOfDate) {
      toast({
        title: 'No Data',
        description: 'Please select a date first',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setExportingPDF(true);
    
    try {
      // Use the new PDF generation method from the service
      const pdfBlob = await ssotBalanceSheetReportService.generateSSOTBalanceSheetPDF({
        as_of_date: asOfDate
      });
      
      // Create download link
      const url = window.URL.createObjectURL(pdfBlob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `SSOT_BalanceSheet_${asOfDate}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      toast({
        title: 'PDF Export Successful',
        description: 'Balance Sheet has been downloaded as PDF',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to export PDF';
      toast({
        title: 'PDF Export Failed',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setExportingPDF(false);
    }
  };

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Card bg={cardBg} border="1px" borderColor={borderColor}>
          <CardHeader pb={3}>
            <HStack justify="space-between">
              <HStack>
                <Icon as={FiBarChart} color="blue.500" boxSize={6} />
                <VStack align="start" spacing={0}>
                  <Text fontSize="xl" fontWeight="bold" color={headingColor}>
                    Enhanced Balance Sheet Report
                  </Text>
                  <Text fontSize="sm" color={descriptionColor}>
                    Real-time SSOT integration with advanced export features
                  </Text>
                </VStack>
              </HStack>
              {onClose && (
                <Button variant="ghost" onClick={onClose}>
                  Close
                </Button>
              )}
            </HStack>
          </CardHeader>

          <CardBody pt={0}>
            <VStack spacing={4} align="stretch">
              {/* Configuration Section */}
              <Box bg={useColorModeValue('gray.50', 'gray.700')} p={4} borderRadius="md">
                <Text fontSize="md" fontWeight="semibold" mb={3} color={headingColor}>
                  Report Configuration
                </Text>
                <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                  <FormControl>
                    <FormLabel fontSize="sm">As of Date</FormLabel>
                    <Input 
                      type="date" 
                      value={asOfDate} 
                      onChange={(e) => setAsOfDate(e.target.value)}
                      size="sm"
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel fontSize="sm">Company Name</FormLabel>
                    <Input 
                      value={companyName} 
                      onChange={(e) => setCompanyName(e.target.value)}
                      placeholder="Company Name"
                      size="sm"
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel fontSize="sm">Include Account Details</FormLabel>
                    <Switch
                      isChecked={includeAccountDetails}
                      onChange={(e) => setIncludeAccountDetails(e.target.checked)}
                      colorScheme="blue"
                    />
                  </FormControl>
                </SimpleGrid>
              </Box>

              {/* Action Buttons */}
              <HStack spacing={3} justify="space-between">
                <HStack spacing={3}>
                  <Button
                    colorScheme="blue"
                    leftIcon={<FiRefreshCw />}
                    onClick={generateBalanceSheetReport}
                    isLoading={loading}
                    loadingText="Generating..."
                    size="md"
                  >
                    Generate Report
                  </Button>
                  
                  {balanceSheetData && (
                    <Button
                      variant="outline"
                      leftIcon={<FiSettings />}
                      onClick={onExportModalOpen}
                      size="md"
                    >
                      Export Options
                    </Button>
                  )}
                </HStack>

                {balanceSheetData && (
                  <HStack spacing={2}>
                    <Tooltip label="Export as CSV file">
                      <Button
                        colorScheme="green"
                        variant="outline"
                        leftIcon={<FiFileText />}
                        onClick={handleExportCSV}
                        isLoading={exportingCSV}
                        loadingText="Exporting..."
                        size="sm"
                      >
                        CSV
                      </Button>
                    </Tooltip>
                    
                    <Tooltip label="Export as PDF file">
                      <Button
                        colorScheme="red"
                        variant="outline"
                        leftIcon={<FiFilePlus />}
                        onClick={handleExportPDF}
                        isLoading={exportingPDF}
                        loadingText="Exporting..."
                        size="sm"
                      >
                        PDF
                      </Button>
                    </Tooltip>
                  </HStack>
                )}
              </HStack>
            </VStack>
          </CardBody>
        </Card>

        {/* Loading State */}
        {loading && (
          <Card bg={cardBg} border="1px" borderColor={borderColor}>
            <CardBody>
              <VStack spacing={4} py={8}>
                <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                <VStack spacing={2}>
                  <Text fontSize="lg" fontWeight="medium" color={headingColor}>
                    Generating Balance Sheet
                  </Text>
                  <Text fontSize="sm" color={descriptionColor}>
                    Fetching real-time data from SSOT journal system...
                  </Text>
                </VStack>
              </VStack>
            </CardBody>
          </Card>
        )}

        {/* Error State */}
        {error && (
          <Alert status="error" borderRadius="md">
            <AlertIcon />
            <VStack align="start" spacing={2} flex="1">
              <AlertTitle>Error generating Balance Sheet</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </VStack>
            <Button
              size="sm"
              colorScheme="red"
              variant="outline"
              onClick={generateBalanceSheetReport}
              ml={4}
            >
              Retry
            </Button>
          </Alert>
        )}

        {/* Balance Sheet Data Display */}
        {balanceSheetData && !loading && (
          <VStack spacing={6} align="stretch">
            {/* Report Header */}
            <Card bg={cardBg} border="1px" borderColor={borderColor}>
              <CardBody textAlign="center">
                <VStack spacing={2}>
                  <Text fontSize="xl" fontWeight="bold" color={headingColor}>
                    {balanceSheetData.company?.name || companyName}
                  </Text>
                  <Text fontSize="lg" fontWeight="semibold">
                    Balance Sheet
                  </Text>
                  <Text fontSize="md" color={descriptionColor}>
                    As of: {new Date(balanceSheetData.as_of_date).toLocaleDateString('id-ID')}
                  </Text>
                  <Badge
                    colorScheme={balanceSheetData.is_balanced ? 'green' : 'red'}
                    px={3}
                    py={1}
                    borderRadius="full"
                  >
                    {balanceSheetData.is_balanced ? (
                      <>
                        <Icon as={FiCheckCircle} mr={1} />
                        Balanced ✓
                      </>
                    ) : (
                      <>
                        <Icon as={FiAlertTriangle} mr={1} />
                        Not Balanced (Diff: {formatCurrency(balanceSheetData.balance_difference)})
                      </>
                    )}
                  </Badge>
                </VStack>
              </CardBody>
            </Card>

            {/* Financial Summary */}
            <SimpleGrid columns={[1, 2, 3]} spacing={4}>
              <Card bg="green.50" border="1px" borderColor="green.200">
                <CardBody textAlign="center">
                  <VStack spacing={1}>
                    <Text fontSize="sm" color="green.600">Total Assets</Text>
                    <Text fontSize="2xl" fontWeight="bold" color="green.700">
                      {formatCurrency(balanceSheetData.assets?.total_assets || 0)}
                    </Text>
                  </VStack>
                </CardBody>
              </Card>
              
              <Card bg="orange.50" border="1px" borderColor="orange.200">
                <CardBody textAlign="center">
                  <VStack spacing={1}>
                    <Text fontSize="sm" color="orange.600">Total Liabilities</Text>
                    <Text fontSize="2xl" fontWeight="bold" color="orange.700">
                      {formatCurrency(balanceSheetData.liabilities?.total_liabilities || 0)}
                    </Text>
                  </VStack>
                </CardBody>
              </Card>
              
              <Card bg="blue.50" border="1px" borderColor="blue.200">
                <CardBody textAlign="center">
                  <VStack spacing={1}>
                    <Text fontSize="sm" color="blue.600">Total Equity</Text>
                    <Text fontSize="2xl" fontWeight="bold" color="blue.700">
                      {formatCurrency(balanceSheetData.equity?.total_equity || 0)}
                    </Text>
                  </VStack>
                </CardBody>
              </Card>
            </SimpleGrid>

            {/* Detailed Breakdown */}
            {includeAccountDetails && (
              <VStack spacing={4} align="stretch">
                <Divider />
                <Text fontSize="lg" fontWeight="bold" color={headingColor}>
                  Account Details
                </Text>

                {/* Assets Table */}
                {balanceSheetData.assets?.current_assets?.items && (
                  <Card bg={cardBg} border="1px" borderColor={borderColor}>
                    <CardHeader pb={2}>
                      <Text fontSize="md" fontWeight="bold" color="green.600">
                        ASSETS
                      </Text>
                    </CardHeader>
                    <CardBody pt={0}>
                      <Table size="sm" variant="simple">
                        <Thead>
                          <Tr>
                            <Th>Account Code</Th>
                            <Th>Account Name</Th>
                            <Th isNumeric>Amount</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {/* Current Assets */}
                          <Tr>
                            <Td colSpan={3}>
                              <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>
                                Current Assets
                              </Text>
                            </Td>
                          </Tr>
                          {balanceSheetData.assets.current_assets.items.map((item, index) => (
                            <Tr key={index}>
                              <Td pl={6}>{item.account_code}</Td>
                              <Td pl={6}>{item.account_name}</Td>
                              <Td isNumeric>{formatCurrency(item.amount, item.account_code, 'ASSET')}</Td>
                            </Tr>
                          ))}
                          
                          {/* Non-Current Assets */}
                          {balanceSheetData.assets.non_current_assets?.items && (
                            <>
                              <Tr>
                                <Td colSpan={3}>
                                  <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>
                                    Non-Current Assets
                                  </Text>
                                </Td>
                              </Tr>
                              {balanceSheetData.assets.non_current_assets.items.map((item, index) => (
                                <Tr key={index}>
                                  <Td pl={6}>{item.account_code}</Td>
                                  <Td pl={6}>{item.account_name}</Td>
                                  <Td isNumeric>{formatCurrency(item.amount, item.account_code, 'ASSET')}</Td>
                                </Tr>
                              ))}
                            </>
                          )}
                          
                          {/* Total Assets */}
                          <Tr bg="green.50">
                            <Td fontWeight="bold">TOTAL ASSETS</Td>
                            <Td></Td>
                            <Td isNumeric fontWeight="bold">
                              {formatCurrency(balanceSheetData.assets?.total_assets || 0)}
                            </Td>
                          </Tr>
                        </Tbody>
                      </Table>
                    </CardBody>
                  </Card>
                )}

                {/* Liabilities & Equity Table */}
                <Card bg={cardBg} border="1px" borderColor={borderColor}>
                  <CardHeader pb={2}>
                    <Text fontSize="md" fontWeight="bold" color="orange.600">
                      LIABILITIES & EQUITY
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <Table size="sm" variant="simple">
                      <Thead>
                        <Tr>
                          <Th>Account Code</Th>
                          <Th>Account Name</Th>
                          <Th isNumeric>Amount</Th>
                        </Tr>
                      </Thead>
                      <Tbody>
                        {/* Liabilities */}
                        {balanceSheetData.liabilities?.current_liabilities?.items && (
                          <>
                            <Tr>
                              <Td colSpan={3}>
                                <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>
                                  Current Liabilities
                                </Text>
                              </Td>
                            </Tr>
                            {balanceSheetData.liabilities.current_liabilities.items.map((item, index) => (
                              <Tr key={index}>
                                <Td pl={6}>{item.account_code}</Td>
                                <Td pl={6}>{item.account_name}</Td>
                                <Td isNumeric>{formatCurrency(item.amount, item.account_code, 'LIABILITY')}</Td>
                              </Tr>
                            ))}
                          </>
                        )}
                        
                        {/* Equity */}
                        {balanceSheetData.equity?.items && (
                          <>
                            <Tr>
                              <Td colSpan={3}>
                                <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>
                                  Equity
                                </Text>
                              </Td>
                            </Tr>
                            {balanceSheetData.equity.items.map((item, index) => (
                              <Tr key={index}>
                                <Td pl={6}>{item.account_code}</Td>
                                <Td pl={6}>{item.account_name}</Td>
                                <Td isNumeric>{formatCurrency(item.amount, item.account_code, 'EQUITY')}</Td>
                              </Tr>
                            ))}
                          </>
                        )}
                        
                        {/* Totals */}
                        <Tr bg="orange.50">
                          <Td fontWeight="bold">TOTAL LIABILITIES</Td>
                          <Td></Td>
                          <Td isNumeric fontWeight="bold">
                            {formatCurrency(balanceSheetData.liabilities?.total_liabilities || 0)}
                          </Td>
                        </Tr>
                        <Tr bg="blue.50">
                          <Td fontWeight="bold">TOTAL EQUITY</Td>
                          <Td></Td>
                          <Td isNumeric fontWeight="bold">
                            {formatCurrency(balanceSheetData.equity?.total_equity || 0)}
                          </Td>
                        </Tr>
                        <Tr bg="gray.100">
                          <Td fontWeight="bold">TOTAL LIABILITIES + EQUITY</Td>
                          <Td></Td>
                          <Td isNumeric fontWeight="bold">
                            {formatCurrency(balanceSheetData.total_liabilities_and_equity || 0)}
                          </Td>
                        </Tr>
                      </Tbody>
                    </Table>
                  </CardBody>
                </Card>
              </VStack>
            )}

            {/* Export Information */}
            <Alert status="info" borderRadius="md">
              <AlertIcon />
              <Box>
                <AlertTitle>Export Ready!</AlertTitle>
                <AlertDescription>
                  Your Balance Sheet is ready for export. Use the CSV button for spreadsheet analysis 
                  or PDF button for professional reporting. Both formats include all account details 
                  and are optimized for printing and sharing.
                </AlertDescription>
              </Box>
            </Alert>
          </VStack>
        )}
      </VStack>

      {/* Export Options Modal */}
      <Modal isOpen={isExportModalOpen} onClose={onExportModalClose} size="md">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Export Options</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} align="stretch">
              <FormControl>
                <FormLabel>Company Name</FormLabel>
                <Input 
                  value={companyName} 
                  onChange={(e) => setCompanyName(e.target.value)}
                  placeholder="Company Name for Export"
                />
              </FormControl>
              
              <FormControl>
                <HStack justify="space-between">
                  <FormLabel mb={0}>Include Account Details</FormLabel>
                  <Switch
                    isChecked={includeAccountDetails}
                    onChange={(e) => setIncludeAccountDetails(e.target.checked)}
                    colorScheme="blue"
                  />
                </HStack>
                <Text fontSize="sm" color={descriptionColor}>
                  Include individual account breakdowns in the export
                </Text>
              </FormControl>

              <VStack spacing={2} align="stretch">
                <Text fontSize="sm" fontWeight="semibold">Export Formats:</Text>
                <HStack spacing={3}>
                  <Button
                    flex={1}
                    colorScheme="green"
                    leftIcon={<FiFileText />}
                    onClick={handleExportCSV}
                    isLoading={exportingCSV}
                    loadingText="Exporting..."
                  >
                    Export CSV
                  </Button>
                  <Button
                    flex={1}
                    colorScheme="red"
                    leftIcon={<FiFilePlus />}
                    onClick={handleExportPDF}
                    isLoading={exportingPDF}
                    loadingText="Exporting..."
                  >
                    Export PDF
                  </Button>
                </HStack>
                <Text fontSize="xs" color={descriptionColor} textAlign="center">
                  Files will be downloaded with filename: balance_sheet_{asOfDate}.{'{format}'}
                </Text>
              </VStack>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" onClick={onExportModalClose}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
};

export default EnhancedBalanceSheetReport;