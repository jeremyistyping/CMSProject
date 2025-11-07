'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Card,
  CardBody,
  useDisclosure,
  useToast,
  FormControl,
  FormLabel,
  Input,
  Grid,
  GridItem,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Badge,
  Divider,
  Tooltip,
  IconButton,
  Switch,
  Flex,
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
  Progress,
  Spinner,
} from '@chakra-ui/react';
import { FiTrendingUp, FiDownload, FiEye, FiDatabase, FiActivity, FiRefreshCw, FiAlertTriangle, FiCheckCircle } from 'react-icons/fi';
import { formatCurrency } from '@/utils/formatters';
import { reportService } from '@/services/reportService';
import { enhancedPLService } from '@/services/enhancedPLService';
import { ssotProfitLossService } from '@/services/ssotProfitLossService';
import { ssotJournalService } from '@/services/ssotJournalService';
import { cogsService } from '@/services/cogsService';
import { BalanceWebSocketClient } from '@/services/balanceWebSocketService';
import { useAuth } from '@/contexts/AuthContext';
import EnhancedProfitLossModal from './EnhancedProfitLossModal';
import { JournalDrilldownModal } from './JournalDrilldownModal';

interface EnhancedPLData {
  title: string;
  period: string;
  company: any;
  enhanced: boolean;
  sections: any[];
  financialMetrics: {
    grossProfit: number;
    grossProfitMargin: number;
    operatingIncome: number;
    operatingMargin: number;
    ebitda: number;
    ebitdaMargin: number;
    netIncome: number;
    netIncomeMargin: number;
  };
}

interface JournalDrilldownRequest {
  account_codes?: string[];
  account_ids?: number[];
  start_date: string;
  end_date: string;
  report_type?: string;
  line_item_name?: string;
  min_amount?: number;
  max_amount?: number;
  transaction_types?: string[];
  page: number;
  limit: number;
}

const EnhancedPLReportPage: React.FC = () => {
  const { token } = useAuth();
  const toast = useToast();
  const [loading, setLoading] = useState(false);
  const [plData, setPLData] = useState<EnhancedPLData | null>(null);
  const [reportParams, setReportParams] = useState({
    start_date: '',
    end_date: '',
  });
  const [drilldownRequest, setDrilldownRequest] = useState<JournalDrilldownRequest | null>(null);
  const [realTimeUpdates, setRealTimeUpdates] = useState(false);
  const [isConnectedToBalanceService, setIsConnectedToBalanceService] = useState(false);
  const [lastUpdateTime, setLastUpdateTime] = useState<Date | null>(null);
  const balanceClientRef = useRef<BalanceWebSocketClient | null>(null);
  
  // COGS Health Check States
  const [cogsHealth, setCogsHealth] = useState<any>(null);
  const [checkingCOGS, setCheckingCOGS] = useState(false);
  const [backfillingCOGS, setBackfillingCOGS] = useState(false);
  const [showCOGSWarning, setShowCOGSWarning] = useState(false);
  
  const { isOpen: isPLModalOpen, onOpen: onPLModalOpen, onClose: onPLModalClose } = useDisclosure();
  const { isOpen: isDrilldownModalOpen, onOpen: onDrilldownModalOpen, onClose: onDrilldownModalClose } = useDisclosure();
  const { isOpen: isCOGSModalOpen, onOpen: onCOGSModalOpen, onClose: onCOGSModalClose } = useDisclosure();

  // Set default date range (current month)
  React.useEffect(() => {
    const today = new Date();
    const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    setReportParams({
      start_date: firstDayOfMonth.toISOString().split('T')[0],
      end_date: today.toISOString().split('T')[0],
    });
  }, []);
  
  // Real-time WebSocket connection effect
  useEffect(() => {
    if (realTimeUpdates && token) {
      initializeBalanceConnection();
    } else {
      disconnectBalanceService();
    }
    
    return () => {
      disconnectBalanceService();
    };
  }, [realTimeUpdates, token]);
  
  const initializeBalanceConnection = async () => {
    if (!token) return;
    
    try {
      balanceClientRef.current = new BalanceWebSocketClient();
      await balanceClientRef.current.connect(token);
      
      balanceClientRef.current.onBalanceUpdate((data) => {
        setLastUpdateTime(new Date());
        
        // Auto-refresh P&L data if we have current data and dates are set
        if (plData && reportParams.start_date && reportParams.end_date) {
          toast({
            title: 'Balance Updated',
            description: `Account ${data.account_code} updated. Refreshing P&L data...`,
            status: 'info',
            duration: 2000,
            isClosable: true,
            position: 'bottom-right',
            size: 'sm'
          });
          
          // Auto-regenerate P&L with updated data
          generateEnhancedPL();
        }
      });
      
      setIsConnectedToBalanceService(true);
      toast({
        title: 'Real-time Updates Enabled',
        description: 'P&L report will refresh automatically when balances change',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error) {
      console.warn('Failed to connect to balance service:', error);
      setIsConnectedToBalanceService(false);
      toast({
        title: 'Real-time Connection Failed',
        description: 'Manual refresh is still available',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
    }
  };
  
  const disconnectBalanceService = () => {
    if (balanceClientRef.current) {
      balanceClientRef.current.disconnect();
      setIsConnectedToBalanceService(false);
    }
  };

  // Check COGS Health before generating P&L
  const checkCOGSHealth = async () => {
    if (!reportParams.start_date || !reportParams.end_date) {
      return null;
    }

    try {
      setCheckingCOGS(true);
      const health = await cogsService.getCOGSHealthStatus(
        reportParams.start_date,
        reportParams.end_date
      );
      setCogsHealth(health);
      return health;
    } catch (error) {
      console.error('Error checking COGS health:', error);
      return null;
    } finally {
      setCheckingCOGS(false);
    }
  };

  // Auto-backfill COGS if needed
  const handleBackfillCOGS = async () => {
    if (!reportParams.start_date || !reportParams.end_date) {
      return;
    }

    try {
      setBackfillingCOGS(true);
      
      const result = await cogsService.backfillCOGS(
        reportParams.start_date,
        reportParams.end_date,
        false // execute, not dry run
      );

      toast({
        title: 'COGS Backfill Complete',
        description: `Successfully processed ${result.sales_processed} sales transactions. COGS entries have been created.`,
        status: 'success',
        duration: 5000,
        isClosable: true,
      });

      // Close modal and re-check health
      onCOGSModalClose();
      await checkCOGSHealth();
      
    } catch (error) {
      console.error('Error backfilling COGS:', error);
      toast({
        title: 'Backfill Failed',
        description: error instanceof Error ? error.message : 'Failed to backfill COGS',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setBackfillingCOGS(false);
    }
  };

  const generateEnhancedPL = async () => {
    if (!reportParams.start_date || !reportParams.end_date) {
      toast({
        title: 'Missing Parameters',
        description: 'Please provide start date and end date',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);
    try {
      console.log('Generating Enhanced P&L with params:', reportParams);
      
      // 🔍 STEP 1: Check COGS Health First
      const health = await checkCOGSHealth();
      
      // 🚨 STEP 2: Show warning if COGS is missing
      if (health && !health.healthy && health.sales_without_cogs > 0) {
        setShowCOGSWarning(true);
        onCOGSModalOpen();
        setLoading(false);
        return; // Stop here, let user decide
      }
      
      // ✅ STEP 3: Generate SSOT P&L (includes COGS automatically)
      const ssotData = await ssotProfitLossService.generateSSOTProfitLoss({
        start_date: reportParams.start_date,
        end_date: reportParams.end_date,
        format: 'json'
      });

      console.log('SSOT P&L data received:', ssotData);

      // Convert SSOT data to the format expected by EnhancedProfitLossModal
      const formattedData: EnhancedPLData = {
        title: ssotData.title || 'Enhanced Profit and Loss Statement',
        period: ssotData.period || `${new Date(reportParams.start_date).toLocaleDateString()} - ${new Date(reportParams.end_date).toLocaleDateString()}`,
        company: ssotData.company || { name: 'Company Name Not Set' },
        enhanced: ssotData.enhanced || true,
        sections: ssotData.sections || [],
        financialMetrics: ssotData.financialMetrics || {
          grossProfit: 0,
          grossProfitMargin: 0,
          operatingIncome: 0,
          operatingMargin: 0,
          ebitda: 0,
          ebitdaMargin: 0,
          netIncome: 0,
          netIncomeMargin: 0,
        },
      };

      setPLData(formattedData);
      setShowCOGSWarning(false); // Clear warning
      onPLModalOpen();

    } catch (error) {
      console.error('Error generating enhanced P&L:', error);
      toast({
        title: 'Generation Failed',
        description: error instanceof Error ? error.message : 'Failed to generate enhanced P&L report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const formatPLSections = (data: any) => {
    const sections = [];

    // Revenue section
    if (data.revenue) {
      const revenueSection = {
        name: 'REVENUE',
        items: [],
        total: data.revenue.total_revenue || 0,
        subsections: []
      };

      // Add subsections for different revenue types
      if (data.revenue.sales_revenue?.items?.length > 0) {
        revenueSection.subsections.push({
          name: 'Sales Revenue',
          items: data.revenue.sales_revenue.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.revenue.sales_revenue.subtotal
        });
      }

      if (data.revenue.service_revenue?.items?.length > 0) {
        revenueSection.subsections.push({
          name: 'Service Revenue',
          items: data.revenue.service_revenue.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.revenue.service_revenue.subtotal
        });
      }

      sections.push(revenueSection);
    }

    // Cost of Goods Sold section
    if (data.cost_of_goods_sold) {
      const cogsSection = {
        name: 'COST OF GOODS SOLD',
        items: [],
        total: data.cost_of_goods_sold.total_cogs || 0,
        subsections: []
      };

      if (data.cost_of_goods_sold.direct_materials?.items?.length > 0) {
        cogsSection.subsections.push({
          name: 'Direct Materials',
          items: data.cost_of_goods_sold.direct_materials.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.cost_of_goods_sold.direct_materials.subtotal
        });
      }

      sections.push(cogsSection);
    }

    // Operating Expenses section
    if (data.operating_expenses) {
      const opexSection = {
        name: 'OPERATING EXPENSES',
        items: [],
        total: data.operating_expenses.total_opex || 0,
        subsections: []
      };

      if (data.operating_expenses.administrative?.items?.length > 0) {
        opexSection.subsections.push({
          name: 'Administrative Expenses',
          items: data.operating_expenses.administrative.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.operating_expenses.administrative.subtotal
        });
      }

      sections.push(opexSection);
    }

    // Calculated sections
    if (data.gross_profit !== undefined) {
      sections.push({
        name: 'GROSS PROFIT',
        items: [
          { name: 'Gross Profit', amount: data.gross_profit },
          { name: 'Gross Profit Margin', amount: data.gross_profit_margin, isPercentage: true }
        ],
        total: data.gross_profit,
        isCalculated: true
      });
    }

    sections.push({
      name: 'NET INCOME',
      items: [
        { name: 'Operating Income', amount: data.operating_income || 0 },
        { name: 'EBITDA', amount: data.ebitda || 0 },
        { name: 'Net Income', amount: data.net_income || 0 },
        { name: 'Net Income Margin', amount: data.net_income_margin || 0, isPercentage: true }
      ],
      total: data.net_income || 0,
      isCalculated: true
    });

    return sections;
  };

  // Handle journal drilldown from P&L modal
  const handleJournalDrilldown = useCallback((itemName: string, accountCode?: string, amount?: number) => {
    console.log('Journal drilldown requested:', { itemName, accountCode, amount });
    
    if (!reportParams.start_date || !reportParams.end_date) {
      toast({
        title: 'Invalid Date Range',
        description: 'Please ensure the P&L report has valid start and end dates',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const drilldownReq: JournalDrilldownRequest = {
      start_date: reportParams.start_date,
      end_date: reportParams.end_date,
      report_type: 'PROFIT_LOSS',
      line_item_name: itemName,
      page: 1,
      limit: 50,
    };

    // Add account filter if available
    if (accountCode) {
      drilldownReq.account_codes = [accountCode];
    }

    // Add amount filter if available
    if (amount !== undefined && amount > 0) {
      drilldownReq.min_amount = Math.max(0, amount * 0.01); // Allow 1% variance
      drilldownReq.max_amount = amount * 1.01;
    }

    setDrilldownRequest(drilldownReq);
    onDrilldownModalOpen();
  }, [reportParams, toast, onDrilldownModalOpen]);

  // Handle P&L export
  const handlePLExport = useCallback(async (format: 'pdf' | 'excel') => {
    if (!reportParams.start_date || !reportParams.end_date) return;

    try {
      const result = await reportService.generateProfessionalReport('profit-loss', {
        ...reportParams,
        format: format === 'excel' ? 'csv' : 'pdf'
      });

      if (result instanceof Blob) {
        const fileName = `profit-loss-${new Date().toISOString().split('T')[0]}.${format === 'excel' ? 'csv' : 'pdf'}`;
        await reportService.downloadReport(result, fileName);
        
        toast({
          title: 'Export Successful',
          description: `P&L report exported as ${format.toUpperCase()}`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
    } catch (error) {
      toast({
        title: 'Export Failed',
        description: error instanceof Error ? error.message : 'Failed to export report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  }, [reportParams, toast]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setReportParams(prev => ({ ...prev, [name]: value }));
  };

  return (
    <Box p={8}>
      <VStack spacing={8} align="stretch">
        {/* Header */}
        <Box>
          <HStack spacing={4} align="center">
            <Box p={3} bg="blue.100" borderRadius="md">
              <FiTrendingUp size="24px" color="blue.600" />
            </Box>
            <VStack align="start" spacing={1}>
              <Text fontSize="2xl" fontWeight="bold" color="gray.700">
                Enhanced Profit & Loss Report
              </Text>
              <Text fontSize="md" color="gray.600">
                Generate comprehensive P&L statements with journal entry drilldown capabilities
              </Text>
            </VStack>
          </HStack>
        </Box>

        {/* Parameters Card */}
        <Card>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Flex justify="space-between" align="center">
                <Text fontSize="lg" fontWeight="semibold">Report Parameters</Text>
                
                {/* Real-time Updates Control */}
                <HStack spacing={3}>
                  <Text fontSize="sm" color="gray.600">Real-time Updates</Text>
                  <Switch
                    isChecked={realTimeUpdates}
                    onChange={(e) => setRealTimeUpdates(e.target.checked)}
                    colorScheme="green"
                  />
                  {isConnectedToBalanceService && (
                    <Tooltip label="Connected to real-time balance service" fontSize="xs">
                      <Badge
                        colorScheme="green"
                        variant="subtle"
                        fontSize="xs"
                        display="flex"
                        alignItems="center"
                        gap={1}
                      >
                        <FiActivity size={10} />
                        LIVE
                      </Badge>
                    </Tooltip>
                  )}
                  {lastUpdateTime && (
                    <Text fontSize="xs" color="gray.500">
                      Last update: {lastUpdateTime.toLocaleTimeString()}
                    </Text>
                  )}
                </HStack>
              </Flex>
              
              <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>Start Date</FormLabel>
                    <Input
                      type="date"
                      name="start_date"
                      value={reportParams.start_date}
                      onChange={handleInputChange}
                    />
                  </FormControl>
                </GridItem>
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>End Date</FormLabel>
                    <Input
                      type="date"
                      name="end_date"
                      value={reportParams.end_date}
                      onChange={handleInputChange}
                    />
                  </FormControl>
                </GridItem>
              </Grid>
            </VStack>
          </CardBody>
        </Card>

        {/* Action Buttons */}
        <HStack spacing={4}>
          <Button
            colorScheme="blue"
            size="lg"
            leftIcon={<FiEye />}
            onClick={generateEnhancedPL}
            isLoading={loading}
            loadingText="Generating..."
          >
            Generate Enhanced P&L
          </Button>
          
          {/* Manual Refresh Button - visible when we have existing data */}
          {plData && (
            <Tooltip label="Refresh with latest data" fontSize="xs">
              <IconButton
                aria-label="Refresh P&L data"
                icon={<FiRefreshCw />}
                size="lg"
                variant="outline"
                colorScheme="gray"
                onClick={generateEnhancedPL}
                isLoading={loading}
                isDisabled={loading}
              />
            </Tooltip>
          )}
          
          <Button
            variant="outline"
            size="lg"
            leftIcon={<FiDownload />}
            onClick={() => handlePLExport('pdf')}
            isDisabled={loading}
          >
            Export PDF
          </Button>
          <Button
            variant="outline"
            size="lg"
            leftIcon={<FiDatabase />}
            onClick={() => handlePLExport('excel')}
            isDisabled={loading}
          >
            Export CSV
          </Button>
        </HStack>

        {/* Quick Stats Preview */}
        {plData?.financialMetrics && (
          <Grid templateColumns="repeat(4, 1fr)" gap={6}>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Gross Profit</StatLabel>
                    <StatNumber color={plData.financialMetrics.grossProfit >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.grossProfit)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.grossProfitMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.grossProfitMargin.toFixed(1)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Operating Income</StatLabel>
                    <StatNumber color={plData.financialMetrics.operatingIncome >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.operatingIncome)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.operatingMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.operatingMargin.toFixed(1)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>EBITDA</StatLabel>
                    <StatNumber color={plData.financialMetrics.ebitda >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.ebitda)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.ebitdaMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.ebitdaMargin.toFixed(1)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Net Income</StatLabel>
                    <StatNumber color={plData.financialMetrics.netIncome >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.netIncome)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.netIncomeMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.netIncomeMargin.toFixed(1)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
          </Grid>
        )}

        {/* Features */}
        <Card>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Text fontSize="lg" fontWeight="semibold">Features</Text>
              <VStack spacing={3} align="start">
                <HStack>
                  <Badge colorScheme="green">Enhanced</Badge>
                  <Text fontSize="sm">Journal entry-based calculations for accurate financial metrics</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="blue">Interactive</Badge>
                  <Text fontSize="sm">Click any line item to drill down to supporting journal entries</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="purple">Comprehensive</Badge>
                  <Text fontSize="sm">Includes gross profit, operating income, EBITDA, and net income with margins</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="orange">Export Ready</Badge>
                  <Text fontSize="sm">Export to PDF or CSV with detailed formatting</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="teal">Real-time</Badge>
                  <Text fontSize="sm">Enable auto-refresh when account balances change via WebSocket</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="yellow">Live Data</Badge>
                  <Text fontSize="sm">Get instant notifications when underlying journal entries are updated</Text>
                </HStack>
              </VStack>
            </VStack>
          </CardBody>
        </Card>
      </VStack>

      {/* Enhanced P&L Modal */}
      {plData && (
        <EnhancedProfitLossModal
          isOpen={isPLModalOpen}
          onClose={onPLModalClose}
          data={plData}
          onJournalDrilldown={handleJournalDrilldown}
          onExport={handlePLExport}
        />
      )}

      {/* Journal Drilldown Modal */}
      {drilldownRequest && (
        <JournalDrilldownModal
          isOpen={isDrilldownModalOpen}
          onClose={onDrilldownModalClose}
          drilldownRequest={drilldownRequest}
          title={`Journal Entries: ${drilldownRequest.line_item_name || 'Selected Item'}`}
        />
      )}

      {/* COGS Warning Modal */}
      <Modal isOpen={isCOGSModalOpen} onClose={onCOGSModalClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            <HStack spacing={2}>
              <FiAlertTriangle color="orange" />
              <Text>COGS Data Missing</Text>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} align="stretch">
              {cogsHealth && (
                <>
                  <Alert status="warning">
                    <AlertIcon />
                    <VStack align="start" spacing={1}>
                      <AlertTitle>Cost of Goods Sold (COGS) entries are missing</AlertTitle>
                      <AlertDescription>
                        {cogsHealth.message}
                      </AlertDescription>
                    </VStack>
                  </Alert>

                  <Card>
                    <CardBody>
                      <VStack spacing={3} align="stretch">
                        <Text fontWeight="semibold">COGS Status:</Text>
                        <HStack justify="space-between">
                          <Text fontSize="sm">Total Sales:</Text>
                          <Badge colorScheme="blue">{cogsHealth.total_sales}</Badge>
                        </HStack>
                        <HStack justify="space-between">
                          <Text fontSize="sm">Sales WITH COGS:</Text>
                          <Badge colorScheme="green">{cogsHealth.sales_with_cogs}</Badge>
                        </HStack>
                        <HStack justify="space-between">
                          <Text fontSize="sm">Sales WITHOUT COGS:</Text>
                          <Badge colorScheme="red">{cogsHealth.sales_without_cogs}</Badge>
                        </HStack>
                        <HStack justify="space-between">
                          <Text fontSize="sm">Completeness:</Text>
                          <Badge colorScheme={cogsHealth.completeness_percentage >= 95 ? 'green' : 'red'}>
                            {cogsHealth.completeness_percentage.toFixed(1)}%
                          </Badge>
                        </HStack>
                      </VStack>
                    </CardBody>
                  </Card>

                  <Alert status="info">
                    <AlertIcon />
                    <VStack align="start" spacing={1} fontSize="sm">
                      <Text fontWeight="semibold">What is COGS?</Text>
                      <Text>
                        COGS (Cost of Goods Sold) represents the direct costs of producing goods sold by a company.
                        Without COGS entries, your P&L report will show incorrect Net Profit.
                      </Text>
                    </VStack>
                  </Alert>

                  <Divider />

                  <VStack spacing={2} align="stretch">
                    <Text fontWeight="semibold" fontSize="sm">Options:</Text>
                    <Button
                      colorScheme="green"
                      leftIcon={backfillingCOGS ? <Spinner size="sm" /> : <FiCheckCircle />}
                      onClick={async () => {
                        await handleBackfillCOGS();
                        // After backfill, generate P&L again
                        await generateEnhancedPL();
                      }}
                      isLoading={backfillingCOGS}
                      loadingText="Creating COGS entries..."
                    >
                      Auto-Create COGS Entries (Recommended)
                    </Button>
                    <Text fontSize="xs" color="gray.600" px={2}>
                      This will automatically create COGS journal entries for all sales without them.
                    </Text>

                    <Divider />

                    <Button
                      variant="outline"
                      onClick={async () => {
                        onCOGSModalClose();
                        setShowCOGSWarning(false);
                        // Continue without COGS check (using SSOT endpoint)
                        setLoading(true);
                        try {
                          const ssotData = await ssotProfitLossService.generateSSOTProfitLoss({
                            start_date: reportParams.start_date,
                            end_date: reportParams.end_date,
                            format: 'json'
                          });
                          const formattedData: EnhancedPLData = {
                            title: ssotData.title || 'Enhanced Profit and Loss Statement',
                            period: ssotData.period || `${new Date(reportParams.start_date).toLocaleDateString()} - ${new Date(reportParams.end_date).toLocaleDateString()}`,
                            company: ssotData.company || { name: 'Company Name Not Set' },
                            enhanced: ssotData.enhanced || true,
                            sections: ssotData.sections || [],
                            financialMetrics: ssotData.financialMetrics || {
                              grossProfit: 0,
                              grossProfitMargin: 0,
                              operatingIncome: 0,
                              operatingMargin: 0,
                              ebitda: 0,
                              ebitdaMargin: 0,
                              netIncome: 0,
                              netIncomeMargin: 0,
                            },
                          };
                          setPLData(formattedData);
                          onPLModalOpen();
                        } catch (error) {
                          console.error('Error:', error);
                          toast({
                            title: 'Generation Failed',
                            description: 'Failed to generate P&L report',
                            status: 'error',
                            duration: 5000,
                            isClosable: true,
                          });
                        } finally {
                          setLoading(false);
                        }
                      }}
                    >
                      Continue Anyway (Not Recommended)
                    </Button>
                    <Text fontSize="xs" color="gray.600" px={2}>
                      Generate P&L without COGS. The report will show Expenses = 0 and incorrect Net Profit.
                    </Text>
                  </VStack>
                </>
              )}
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" onClick={onCOGSModalClose}>
              Cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
};

export default EnhancedPLReportPage;