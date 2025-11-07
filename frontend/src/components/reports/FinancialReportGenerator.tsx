import React, { useState, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Select,
  FormControl,
  FormLabel,
  Switch,
  Alert,
  AlertIcon,
  useToast,
  Spinner,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Divider,
  SimpleGrid,
  Badge,
  Icon,
  Input,
} from '@chakra-ui/react';
import { FiCalendar, FiDownload, FiEye, FiFileText } from 'react-icons/fi';
import financialReportService, { 
  FinancialReportRequest, 
  ReportMetadata,
  ProfitLossStatement,
  BalanceSheet,
  CashFlowStatement,
  TrialBalance,
  GeneralLedger
} from '../../services/financialReportService';

interface FinancialReportGeneratorProps {
  onReportGenerated?: (reportType: string, reportData: any) => void;
}

const FinancialReportGenerator: React.FC<FinancialReportGeneratorProps> = ({ 
  onReportGenerated 
}) => {
  const [reportTypes, setReportTypes] = useState<ReportMetadata[]>([]);
  const [isLoadingReports, setIsLoadingReports] = useState<boolean>(true);
  const [selectedReportType, setSelectedReportType] = useState('');
  const [startDate, setStartDate] = useState<Date>(new Date(new Date().getFullYear(), 0, 1));
  const [endDate, setEndDate] = useState<Date>(new Date());
  const [comparative, setComparative] = useState(false);
  const [showZero, setShowZero] = useState(false);
  const [accountId, setAccountId] = useState<number | undefined>();
  const [isLoading, setIsLoading] = useState(false);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [generatedReport, setGeneratedReport] = useState<any>(null);
  
  const toast = useToast();

  useEffect(() => {
    loadReportTypes();
  }, []);

  const loadReportTypes = async () => {
    setIsLoadingReports(true);
    try {
      const reports = await financialReportService.getReportsList();
      setReportTypes(reports || []);
    } catch (error) {
      console.error('Failed to load report types:', error);
      setReportTypes([]); // Ensure we have an empty array on error
      toast({
        title: 'Error',
        description: 'Failed to load available reports',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsLoadingReports(false);
    }
  };

  const validateRequest = async (): Promise<boolean> => {
    try {
      const request: FinancialReportRequest = {
        report_type: selectedReportType,
        start_date: new Date(startDate.toISOString().split('T')[0] + 'T00:00:00.000Z'),
        end_date: new Date(endDate.toISOString().split('T')[0] + 'T00:00:00.000Z'),
        comparative,
        show_zero: showZero,
      };

      const validation = await financialReportService.validateReportRequest(request);
      
      if (!validation.valid) {
        setValidationErrors(validation.errors);
        return false;
      }

      setValidationErrors([]);
      return true;
    } catch (error) {
      console.error('Validation failed:', error);
      return false;
    }
  };

  const generateReport = async () => {
    if (!selectedReportType) {
      toast({
        title: 'Error',
        description: 'Please select a report type',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const isValid = await validateRequest();
    if (!isValid) {
      toast({
        title: 'Validation Error',
        description: 'Please check the form inputs',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setIsLoading(true);
    try {
      let reportData: any;

      const request: FinancialReportRequest = {
        report_type: selectedReportType,
        start_date: startDate,
        end_date: endDate,
        comparative,
        show_zero: showZero,
      };

      switch (selectedReportType) {
        case 'PROFIT_LOSS':
          reportData = await financialReportService.generateProfitLossStatement(request);
          break;
        case 'BALANCE_SHEET':
          reportData = await financialReportService.generateBalanceSheet(request);
          break;
        case 'CASH_FLOW':
          reportData = await financialReportService.generateCashFlowStatement(request);
          break;
        case 'TRIAL_BALANCE':
          reportData = await financialReportService.generateTrialBalance(request);
          break;
        case 'GENERAL_LEDGER':
          if (!accountId) {
            toast({
              title: 'Error',
              description: 'Account ID is required for General Ledger',
              status: 'error',
              duration: 3000,
              isClosable: true,
            });
            return;
          }
          reportData = await financialReportService.generateGeneralLedger(
            accountId,
            startDate.toISOString().split('T')[0],
            endDate.toISOString().split('T')[0]
          );
          break;
        default:
          throw new Error('Unsupported report type');
      }

      setGeneratedReport(reportData);
      onReportGenerated?.(selectedReportType, reportData);

      toast({
        title: 'Success',
        description: 'Report generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Failed to generate report:', error);
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const getSelectedReportMetadata = (): ReportMetadata | undefined => {
    return reportTypes.find(report => report.reportType === selectedReportType);
  };

  const requiresAccountId = (): boolean => {
    const metadata = getSelectedReportMetadata();
    return metadata?.requiredParams.includes('account_id') || false;
  };

  const supportsComparative = (): boolean => {
    const metadata = getSelectedReportMetadata();
    return metadata?.supportsComparative || false;
  };

  return (
    <Card>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <Icon as={FiFileText} mr={2} />
          Generate Financial Report
        </Heading>
      </CardHeader>
      <CardBody>
        <VStack spacing={6} align="stretch">
          {/* Report Type Selection */}
          <FormControl isRequired>
            <FormLabel>Report Type</FormLabel>
            <Select 
              placeholder="Select report type"
              value={selectedReportType}
              onChange={(e) => setSelectedReportType(e.target.value)}
            >
              {reportTypes && reportTypes.length > 0 ? (
                reportTypes.map((report, index) => (
                  <option 
                    key={report.reportType || `report-${index}`} 
                    value={report.reportType}
                  >
                    {report.name || 'Unknown Report'}
                  </option>
                ))
              ) : (
                <option disabled key="no-reports">Loading reports...</option>
              )}
            </Select>
            {selectedReportType && (
              <Text fontSize="sm" color="gray.600" mt={2}>
                {getSelectedReportMetadata()?.description}
              </Text>
            )}
          </FormControl>

          {/* Date Range Selection */}
          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
            <FormControl isRequired>
              <FormLabel display="flex" alignItems="center">
                <Icon as={FiCalendar} mr={2} />
                Start Date
              </FormLabel>
              <Box>
                <Input
                  type="date"
                  value={startDate.toISOString().split('T')[0]}
                  onChange={(e) => setStartDate(new Date(e.target.value))}
                />
              </Box>
            </FormControl>

            <FormControl isRequired>
              <FormLabel display="flex" alignItems="center">
                <Icon as={FiCalendar} mr={2} />
                End Date
              </FormLabel>
              <Box>
                <Input
                  type="date"
                  value={endDate.toISOString().split('T')[0]}
                  onChange={(e) => setEndDate(new Date(e.target.value))}
                />
              </Box>
            </FormControl>
          </SimpleGrid>

          {/* Account ID for General Ledger */}
          {requiresAccountId() && (
            <FormControl isRequired>
              <FormLabel>Account ID</FormLabel>
              <input
                type="number"
                placeholder="Enter account ID"
                value={accountId || ''}
                onChange={(e) => setAccountId(Number(e.target.value) || undefined)}
                style={{
                  padding: '8px 12px',
                  border: '1px solid #CBD5E0',
                  borderRadius: '6px',
                  width: '100%',
                }}
              />
            </FormControl>
          )}

          {/* Options */}
          <VStack spacing={4} align="stretch">
            {supportsComparative() && (
              <FormControl display="flex" alignItems="center">
                <FormLabel mb="0">
                  Comparative Analysis
                </FormLabel>
                <Switch
                  isChecked={comparative}
                  onChange={(e) => setComparative(e.target.checked)}
                  colorScheme="blue"
                />
              </FormControl>
            )}

            <FormControl display="flex" alignItems="center">
              <FormLabel mb="0">
                Show Zero Balances
              </FormLabel>
              <Switch
                isChecked={showZero}
                onChange={(e) => setShowZero(e.target.checked)}
                colorScheme="blue"
              />
            </FormControl>
          </VStack>

          {/* Validation Errors */}
          {validationErrors.length > 0 && (
            <Alert status="error">
              <AlertIcon />
              <VStack align="start" spacing={1}>
                {validationErrors.map((error, index) => (
                  <Text key={`validation-error-${index}`} fontSize="sm">{error}</Text>
                ))}
              </VStack>
            </Alert>
          )}

          {/* Generate Button */}
          <Button
            colorScheme="blue"
            size="lg"
            onClick={generateReport}
            isLoading={isLoading}
            loadingText="Generating Report..."
            leftIcon={<Icon as={FiFileText} />}
          >
            Generate Report
          </Button>

          {/* Report Preview */}
          {generatedReport && (
            <Card borderColor="blue.200">
              <CardHeader>
                <HStack justify="space-between">
                  <Heading size="sm">Generated Report</Heading>
                  <HStack>
                    <Badge colorScheme="green">Ready</Badge>
                    <Button
                      size="sm"
                      leftIcon={<Icon as={FiEye} />}
                      onClick={() => onReportGenerated?.(selectedReportType, generatedReport)}
                    >
                      View
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      leftIcon={<Icon as={FiDownload} />}
                      onClick={() => {
                        // Export functionality would go here
                        toast({
                          title: 'Info',
                          description: 'Export functionality coming soon',
                          status: 'info',
                          duration: 3000,
                          isClosable: true,
                        });
                      }}
                    >
                      Export
                    </Button>
                  </HStack>
                </HStack>
              </CardHeader>
              <CardBody pt={0}>
                <VStack align="start" spacing={2}>
                  <Text fontSize="sm" color="gray.600">
                    <strong>Report Type:</strong> {getSelectedReportMetadata()?.name}
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    <strong>Period:</strong> {financialReportService.formatDateRange(startDate, endDate)}
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    <strong>Generated At:</strong> {new Date().toLocaleString('id-ID')}
                  </Text>
                </VStack>
              </CardBody>
            </Card>
          )}
        </VStack>
      </CardBody>
    </Card>
  );
};

export default FinancialReportGenerator;
