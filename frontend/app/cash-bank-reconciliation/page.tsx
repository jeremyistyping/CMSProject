'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Text,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Card,
  CardBody,
  Badge,
  useToast,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  VStack,
  HStack,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  useColorModeValue,
  Spinner,
  IconButton,
  Tooltip,
} from '@chakra-ui/react';
import { FiRefreshCw, FiAlertTriangle, FiCheckCircle, FiSettings, FiEye } from 'react-icons/fi';

// Types for Finance Dashboard and Bank Reconciliation
interface FinanceDashboardData {
  invoices_pending_payment: number;
  invoices_not_paid: number;
  journals_need_posting: number;
  bank_reconciliation: BankReconciliation;
  outstanding_receivables: number;
  outstanding_payables: number;
  cash_bank_balance: number;
}

interface BankReconciliation {
  last_reconciled?: string;
  days_ago: number;
  status: 'up_to_date' | 'recent' | 'needs_attention' | 'never_reconciled';
}

interface CashBankAccount {
  id: number;
  name: string;
  code: string;
  type: 'CASH' | 'BANK';
  balance: number;
  ssot_balance: number;
  variance: number;
  reconciliation_status: string;
  last_transaction_date?: string;
  is_active: boolean;
}

interface IntegratedSummary {
  summary: {
    total_cash: number;
    total_bank: number;
    total_balance: number;
    total_ssot_balance: number;
    balance_variance: number;
    variance_count: number;
  };
  sync_status: {
    total_accounts: number;
    synced_accounts: number;
    variance_accounts: number;
    last_sync_at?: string;
    sync_status: string;
  };
  accounts: CashBankAccount[];
}

interface ReconciliationRequest {
  strategy: 'to_ssot' | 'to_transactions' | 'to_coa';
  dry_run: boolean;
  account_ids?: number[];
}

const CashBankReconciliationPage: React.FC = () => {
  const { token } = useAuth();
  const { t } = useTranslation();
  const toast = useToast();
  
  // Color mode values
  const bg = useColorModeValue('white', 'gray.800');
  const textColor = useColorModeValue('gray.800', 'white');
  const mutedTextColor = useColorModeValue('gray.500', 'gray.400');
  
  // State
  const [financeData, setFinanceData] = useState<FinanceDashboardData | null>(null);
  const [integratedData, setIntegratedData] = useState<IntegratedSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reconciling, setReconciling] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<CashBankAccount | null>(null);
  
  const { isOpen, onOpen, onClose } = useDisclosure();
  
  // Fetch finance dashboard data
  const fetchFinanceData = async () => {
    try {
      const response = await fetch('/api/v1/dashboard/finance', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) throw new Error('Failed to fetch finance data');
      
      const result = await response.json();
      setFinanceData(result.data);
    } catch (err: any) {
      console.error('Error fetching finance data:', err);
      toast({
        title: 'Error',
        description: err.message || 'Failed to fetch finance dashboard data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  // Fetch integrated cashbank data
  const fetchIntegratedData = async () => {
    try {
      const response = await fetch('/api/v1/cashbank/integrated/summary', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) throw new Error('Failed to fetch integrated data');
      
      const result = await response.json();
      setIntegratedData(result.data);
    } catch (err: any) {
      console.error('Error fetching integrated data:', err);
      toast({
        title: 'Error',
        description: err.message || 'Failed to fetch integrated cashbank data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  // Fetch all data
  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      await Promise.all([fetchFinanceData(), fetchIntegratedData()]);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  };
  
  useEffect(() => {
    if (token) {
      fetchData();
    }
  }, [token]);
  
  // Handle reconciliation
  const handleReconcile = async (strategy: 'to_ssot' | 'to_transactions' | 'to_coa' = 'to_ssot', dryRun = false) => {
    try {
      setReconciling(true);
      
      const request: ReconciliationRequest = {
        strategy,
        dry_run: dryRun,
        account_ids: selectedAccount ? [selectedAccount.id] : undefined,
      };
      
      const response = await fetch('/api/v1/cashbank/integrated/reconcile', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });
      
      if (!response.ok) throw new Error('Failed to perform reconciliation');
      
      const result = await response.json();
      
      toast({
        title: dryRun ? 'Simulation Complete' : 'Reconciliation Complete',
        description: dryRun 
          ? 'Reconciliation simulation completed successfully' 
          : 'All accounts have been reconciled successfully',
        status: 'success',
        duration: 5000,
        isClosable: true,
      });
      
      // Refresh data
      await fetchData();
      onClose();
      
    } catch (err: any) {
      toast({
        title: 'Reconciliation Failed',
        description: err.message || 'Failed to perform reconciliation',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setReconciling(false);
    }
  };
  
  // Get reconciliation status info
  const getReconciliationStatusInfo = (status: string) => {
    switch (status) {
      case 'never_reconciled':
        return {
          status: 'error' as const,
          color: 'red',
          icon: FiAlertTriangle,
          label: 'Never Reconciled',
          description: 'Bank reconciliation has never been performed',
        };
      case 'needs_attention':
        return {
          status: 'warning' as const,
          color: 'orange',
          icon: FiAlertTriangle,
          label: 'Needs Attention',
          description: 'Reconciliation is overdue (more than 7 days ago)',
        };
      case 'recent':
        return {
          status: 'info' as const,
          color: 'yellow',
          icon: FiSettings,
          label: 'Recent',
          description: 'Reconciled within the last 7 days',
        };
      case 'up_to_date':
        return {
          status: 'success' as const,
          color: 'green',
          icon: FiCheckCircle,
          label: 'Up to Date',
          description: 'Reconciled within the last 24 hours',
        };
      default:
        return {
          status: 'info' as const,
          color: 'gray',
          icon: FiSettings,
          label: 'Unknown',
          description: 'Reconciliation status is unknown',
        };
    }
  };
  
  if (loading) {
    return (
      <SimpleLayout allowedRoles={['admin', 'finance', 'director']}>
        <Flex justify="center" align="center" h="200px">
          <VStack spacing={4}>
            <Spinner size="xl" color="blue.500" />
            <Text>{t('common.loading')}</Text>
          </VStack>
        </Flex>
      </SimpleLayout>
    );
  }
  
  if (error) {
    return (
      <SimpleLayout allowedRoles={['admin', 'finance', 'director']}>
        <Alert status="error" mb={6}>
          <AlertIcon />
          <AlertTitle mr={2}>Error!</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </SimpleLayout>
    );
  }
  
  const statusInfo = financeData?.bank_reconciliation 
    ? getReconciliationStatusInfo(financeData.bank_reconciliation.status)
    : null;
  
  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <div>
            <Heading size="lg" mb={2}>Cash & Bank Reconciliation</Heading>
            <Text color={mutedTextColor}>Kelola dan rekonsiliasi saldo kas dan bank dengan sistem akuntansi</Text>
          </div>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              onClick={fetchData}
              isLoading={loading}
              variant="outline"
            >
              Refresh
            </Button>
            <Button
              colorScheme="blue"
              leftIcon={statusInfo?.icon ? React.createElement(statusInfo.icon) : undefined}
              onClick={onOpen}
              isDisabled={!integratedData || integratedData.summary.variance_count === 0}
            >
              Reconcile All
            </Button>
          </HStack>
        </Flex>
        
        {/* Reconciliation Status Alert */}
        {financeData?.bank_reconciliation && statusInfo && (
          <Alert 
            status={statusInfo.status}
            mb={6}
            borderRadius="lg"
          >
            <AlertIcon as={statusInfo.icon} />
            <Box>
              <AlertTitle>Bank Reconciliation Status: {statusInfo.label}</AlertTitle>
              <AlertDescription>
                {statusInfo.description}
                {financeData.bank_reconciliation.last_reconciled && (
                  <Text fontSize="sm" mt={1}>
                    Last reconciled: {new Date(financeData.bank_reconciliation.last_reconciled).toLocaleDateString('id-ID')}
                    {financeData.bank_reconciliation.days_ago > 0 && (
                      <> ({financeData.bank_reconciliation.days_ago} days ago)</>
                    )}
                  </Text>
                )}
              </AlertDescription>
            </Box>
          </Alert>
        )}
        
        <Tabs variant="enclosed" colorScheme="blue">
          <TabList>
            <Tab>Dashboard Overview</Tab>
            <Tab>Variance Analysis</Tab>
            <Tab>Account Details</Tab>
          </TabList>
          
          <TabPanels>
            {/* Dashboard Overview Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                {/* Summary Cards */}
                <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6}>
                  <Card>
                    <CardBody>
                      <Stat>
                        <StatLabel>Total Cash Balance</StatLabel>
                        <StatNumber color="green.500">
                          IDR {integratedData?.summary.total_cash?.toLocaleString() || '0'}
                        </StatNumber>
                        <StatHelpText>Physical cash accounts</StatHelpText>
                      </Stat>
                    </CardBody>
                  </Card>
                  
                  <Card>
                    <CardBody>
                      <Stat>
                        <StatLabel>Total Bank Balance</StatLabel>
                        <StatNumber color="blue.500">
                          IDR {integratedData?.summary.total_bank?.toLocaleString() || '0'}
                        </StatNumber>
                        <StatHelpText>Bank accounts</StatHelpText>
                      </Stat>
                    </CardBody>
                  </Card>
                  
                  <Card>
                    <CardBody>
                      <Stat>
                        <StatLabel>SSOT Balance</StatLabel>
                        <StatNumber color="purple.500">
                          IDR {integratedData?.summary.total_ssot_balance?.toLocaleString() || '0'}
                        </StatNumber>
                        <StatHelpText>Journal system balance</StatHelpText>
                      </Stat>
                    </CardBody>
                  </Card>
                  
                  <Card>
                    <CardBody>
                      <Stat>
                        <StatLabel>Variance</StatLabel>
                        <StatNumber color={integratedData?.summary.variance_count ? 'red.500' : 'green.500'}>
                          {integratedData?.summary.variance_count || 0} Accounts
                        </StatNumber>
                        <StatHelpText>
                          IDR {Math.abs(integratedData?.summary.balance_variance || 0).toLocaleString()}
                        </StatHelpText>
                      </Stat>
                    </CardBody>
                  </Card>
                </SimpleGrid>
                
                {/* Sync Status */}
                {integratedData?.sync_status && (
                  <Card>
                    <CardBody>
                      <Heading size="md" mb={4}>Synchronization Status</Heading>
                      <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
                        <Box textAlign="center">
                          <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                            {integratedData.sync_status.total_accounts}
                          </Text>
                          <Text fontSize="sm" color={mutedTextColor}>Total Accounts</Text>
                        </Box>
                        <Box textAlign="center">
                          <Text fontSize="2xl" fontWeight="bold" color="green.600">
                            {integratedData.sync_status.synced_accounts}
                          </Text>
                          <Text fontSize="sm" color={mutedTextColor}>Synchronized</Text>
                        </Box>
                        <Box textAlign="center">
                          <Text fontSize="2xl" fontWeight="bold" color="red.600">
                            {integratedData.sync_status.variance_accounts}
                          </Text>
                          <Text fontSize="sm" color={mutedTextColor}>With Variance</Text>
                        </Box>
                      </SimpleGrid>
                      {integratedData.sync_status.last_sync_at && (
                        <Text fontSize="sm" color={mutedTextColor} textAlign="center" mt={3}>
                          Last synchronized: {new Date(integratedData.sync_status.last_sync_at).toLocaleString('id-ID')}
                        </Text>
                      )}
                    </CardBody>
                  </Card>
                )}
              </VStack>
            </TabPanel>
            
            {/* Variance Analysis Tab */}
            <TabPanel>
              {integratedData?.accounts && (
                <Card>
                  <CardBody>
                    <Heading size="md" mb={4}>Account Variance Analysis</Heading>
                    <VStack spacing={4} align="stretch">
                      {integratedData.accounts
                        .filter(account => Math.abs(account.variance) > 0.01)
                        .map(account => (
                        <Box key={account.id} p={4} borderRadius="md" border="1px" borderColor="red.200" bg="red.50">
                          <HStack justify="space-between">
                            <VStack align="start" spacing={1}>
                              <Text fontWeight="bold">{account.name}</Text>
                              <Text fontSize="sm" color={mutedTextColor}>
                                {account.code} • {account.type}
                              </Text>
                            </VStack>
                            <VStack align="end" spacing={1}>
                              <Text fontWeight="bold" color="red.600">
                                Variance: IDR {Math.abs(account.variance).toLocaleString()}
                              </Text>
                              <HStack spacing={4}>
                                <Text fontSize="sm">
                                  Balance: IDR {account.balance.toLocaleString()}
                                </Text>
                                <Text fontSize="sm">
                                  SSOT: IDR {account.ssot_balance.toLocaleString()}
                                </Text>
                              </HStack>
                              <Button 
                                size="xs" 
                                colorScheme="blue" 
                                onClick={() => { setSelectedAccount(account); onOpen(); }}
                              >
                                Reconcile
                              </Button>
                            </VStack>
                          </HStack>
                        </Box>
                      ))}
                      {integratedData.accounts.filter(account => Math.abs(account.variance) > 0.01).length === 0 && (
                        <Box textAlign="center" py={8}>
                          <FiCheckCircle size={48} color="green" style={{ margin: '0 auto 16px' }} />
                          <Text color="green.600" fontWeight="bold">No Variance Detected</Text>
                          <Text fontSize="sm" color={mutedTextColor}>All accounts are properly reconciled</Text>
                        </Box>
                      )}
                    </VStack>
                  </CardBody>
                </Card>
              )}
            </TabPanel>
            
            {/* Account Details Tab */}
            <TabPanel>
              {integratedData?.accounts && (
                <Card>
                  <CardBody>
                    <Heading size="md" mb={4}>All Cash & Bank Accounts</Heading>
                    <VStack spacing={3} align="stretch">
                      {integratedData.accounts.map(account => {
                        const hasVariance = Math.abs(account.variance) > 0.01;
                        return (
                          <Box 
                            key={account.id} 
                            p={4} 
                            borderRadius="md" 
                            border="1px" 
                            borderColor={hasVariance ? 'red.200' : 'green.200'}
                            bg={hasVariance ? 'red.50' : 'green.50'}
                          >
                            <HStack justify="space-between">
                              <HStack spacing={4}>
                                <Text fontSize="2xl">{account.type === 'CASH' ? '💵' : '🏦'}</Text>
                                <VStack align="start" spacing={1}>
                                  <Text fontWeight="bold">{account.name}</Text>
                                  <Text fontSize="sm" color={mutedTextColor}>
                                    {account.code} • {account.type}
                                  </Text>
                                  <Badge 
                                    colorScheme={account.is_active ? 'green' : 'red'}
                                    size="sm"
                                  >
                                    {account.is_active ? 'Active' : 'Inactive'}
                                  </Badge>
                                </VStack>
                              </HStack>
                              <VStack align="end" spacing={1}>
                                <Text fontWeight="bold">
                                  IDR {account.balance.toLocaleString()}
                                </Text>
                                <Text fontSize="sm" color={mutedTextColor}>
                                  SSOT: IDR {account.ssot_balance.toLocaleString()}
                                </Text>
                                {hasVariance && (
                                  <Text fontSize="sm" color="red.600" fontWeight="bold">
                                    Variance: IDR {Math.abs(account.variance).toLocaleString()}
                                  </Text>
                                )}
                                <HStack spacing={2}>
                                  <Tooltip label="View Details">
                                    <IconButton
                                      aria-label="View details"
                                      icon={<FiEye />}
                                      size="xs"
                                      variant="outline"
                                    />
                                  </Tooltip>
                                  {hasVariance && (
                                    <Button 
                                      size="xs" 
                                      colorScheme="blue"
                                      onClick={() => { setSelectedAccount(account); onOpen(); }}
                                    >
                                      Reconcile
                                    </Button>
                                  )}
                                </HStack>
                              </VStack>
                            </HStack>
                          </Box>
                        );
                      })}
                    </VStack>
                  </CardBody>
                </Card>
              )}
            </TabPanel>
          </TabPanels>
        </Tabs>
        
        {/* Reconciliation Modal */}
        <Modal isOpen={isOpen} onClose={onClose} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              {selectedAccount ? `Reconcile ${selectedAccount.name}` : 'Reconcile All Accounts'}
            </ModalHeader>
            <ModalCloseButton />
            
            <ModalBody>
              <VStack spacing={4} align="stretch">
                <Alert status="info">
                  <AlertIcon />
                  <Box>
                    <AlertTitle fontSize="sm">Reconciliation Process</AlertTitle>
                    <AlertDescription fontSize="xs">
                      This will synchronize balances between Cash/Bank accounts and the journal system.
                    </AlertDescription>
                  </Box>
                </Alert>
                
                {selectedAccount && (
                  <Box p={4} bg="gray.50" borderRadius="md">
                    <Text fontWeight="bold" mb={2}>{selectedAccount.name}</Text>
                    <HStack justify="space-between" mb={1}>
                      <Text fontSize="sm">Current Balance:</Text>
                      <Text fontSize="sm" fontWeight="bold">
                        IDR {selectedAccount.balance.toLocaleString()}
                      </Text>
                    </HStack>
                    <HStack justify="space-between" mb={1}>
                      <Text fontSize="sm">SSOT Balance:</Text>
                      <Text fontSize="sm" fontWeight="bold">
                        IDR {selectedAccount.ssot_balance.toLocaleString()}
                      </Text>
                    </HStack>
                    <HStack justify="space-between">
                      <Text fontSize="sm">Variance:</Text>
                      <Text fontSize="sm" fontWeight="bold" color="red.600">
                        IDR {Math.abs(selectedAccount.variance).toLocaleString()}
                      </Text>
                    </HStack>
                  </Box>
                )}
                
                <Text fontSize="sm" color={mutedTextColor}>
                  Choose reconciliation strategy:
                </Text>
                
                <VStack spacing={2} align="stretch">
                  <Button 
                    variant="outline" 
                    onClick={() => handleReconcile('to_ssot')}
                    isLoading={reconciling}
                  >
                    Reconcile to SSOT Balance (Recommended)
                  </Button>
                  <Button 
                    variant="outline" 
                    onClick={() => handleReconcile('to_transactions')}
                    isLoading={reconciling}
                  >
                    Reconcile to Transaction Balance
                  </Button>
                  <Button 
                    variant="outline" 
                    onClick={() => handleReconcile('to_coa')}
                    isLoading={reconciling}
                  >
                    Reconcile to COA Balance
                  </Button>
                </VStack>
              </VStack>
            </ModalBody>
            
            <ModalFooter>
              <HStack spacing={3}>
                <Button variant="outline" onClick={onClose}>
                  Cancel
                </Button>
                <Button 
                  colorScheme="yellow" 
                  onClick={() => handleReconcile('to_ssot', true)}
                  isLoading={reconciling}
                >
                  Simulate First
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </Box>
    </SimpleLayout>
  );
};

export default CashBankReconciliationPage;