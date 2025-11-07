'use client';

import React, { useEffect, useState } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  SimpleGrid,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Select,
  Button,
  Spinner,
  Alert,
  AlertIcon,
  AlertDescription,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Flex,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Tag,
  TagLabel,
  TagLeftIcon,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react';
import { 
  FiTrendingUp, 
  FiTrendingDown, 
  FiDollarSign, 
  FiClock, 
  FiRefreshCw,
  FiCalendar,
  FiActivity,
  FiBarChart3,
  FiPlus,
  FiFileText,
  FiEye,
  FiEdit,
  FiArrowDown,
  FiArrowRight
} from 'react-icons/fi';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
} from 'chart.js';
import { paymentService } from '@/services/paymentService';
import type { PaymentAnalytics, Payment, PaymentIntegrationMetrics, PaymentWithJournalInfo } from '@/services/paymentService';
import PaymentWithJournalForm from './PaymentWithJournalForm';
import PaymentDetailsView from './PaymentDetailsView';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
);

interface StatCardProps {
  title: string;
  value: string;
  change?: number;
  icon: React.ElementType;
  colorScheme: string;
  isLoading?: boolean;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  change,
  icon: Icon,
  colorScheme,
  isLoading = false
}) => {
  const getChangeColor = (change?: number) => {
    if (!change) return 'gray';
    return change >= 0 ? 'green' : 'red';
  };

  const getChangeIcon = (change?: number) => {
    if (!change) return null;
    return change >= 0 ? 'increase' : 'decrease';
  };

  return (
    <Card>
      <CardBody>
        <Stat>
          <Flex justify="space-between" align="center">
            <Box>
              <StatLabel fontSize="sm" color="gray.500">
                {title}
              </StatLabel>
              {isLoading ? (
                <Spinner size="sm" />
              ) : (
                <StatNumber fontSize="2xl" fontWeight="bold">
                  {value}
                </StatNumber>
              )}
              {change !== undefined && !isLoading && (
                <StatHelpText mb={0}>
                  <StatArrow type={getChangeIcon(change) as any} />
                  <Text as="span" color={`${getChangeColor(change)}.500`}>
                    {Math.abs(change).toFixed(1)}%
                  </Text>
                </StatHelpText>
              )}
            </Box>
            <Box p={3} borderRadius="full" bg={`${colorScheme}.100`}>
              <Icon size={24} color={`var(--chakra-colors-${colorScheme}-500)`} />
            </Box>
          </Flex>
        </Stat>
      </CardBody>
    </Card>
  );
};

interface PaymentTrendChartProps {
  data: Array<{
    date: string;
    received: number;
    paid: number;
  }>;
  isLoading?: boolean;
}

const PaymentTrendChart: React.FC<PaymentTrendChartProps> = ({ data, isLoading }) => {
  const chartData = {
    labels: data.map(item => new Date(item.date).toLocaleDateString('id-ID')),
    datasets: [
      {
        label: 'Received',
        data: data.map(item => item.received),
        borderColor: 'rgb(34, 197, 94)',
        backgroundColor: 'rgba(34, 197, 94, 0.1)',
        tension: 0.4,
      },
      {
        label: 'Paid',
        data: data.map(item => item.paid),
        borderColor: 'rgb(239, 68, 68)',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        tension: 0.4,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: true,
        text: 'Payment Trend (Last 30 Days)',
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          callback: function(value: any) {
            return new Intl.NumberFormat('id-ID', {
              style: 'currency',
              currency: 'IDR',
              notation: 'compact',
              maximumFractionDigits: 0,
            }).format(value);
          },
        },
      },
    },
  };

  return (
    <Card>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <FiActivity style={{ marginRight: '8px' }} />
          Payment Trend
        </Heading>
      </CardHeader>
      <CardBody>
        {isLoading ? (
          <Flex justify="center" align="center" height="300px">
            <Spinner size="lg" />
          </Flex>
        ) : (
          <Box height="300px">
            <Line data={chartData} options={options} />
          </Box>
        )}
      </CardBody>
    </Card>
  );
};

interface PaymentMethodChartProps {
  data: Record<string, number>;
  isLoading?: boolean;
}

const PaymentMethodChart: React.FC<PaymentMethodChartProps> = ({ data, isLoading }) => {
  const methods = Object.keys(data);
  const values = Object.values(data);

  const chartData = {
    labels: methods.map(method => paymentService.getMethodDisplayName(method)),
    datasets: [
      {
        label: 'Amount',
        data: values,
        backgroundColor: [
          'rgba(59, 130, 246, 0.8)', // blue
          'rgba(34, 197, 94, 0.8)',  // green
          'rgba(251, 191, 36, 0.8)', // yellow
          'rgba(239, 68, 68, 0.8)',  // red
          'rgba(168, 85, 247, 0.8)', // purple
          'rgba(236, 72, 153, 0.8)', // pink
        ],
        borderWidth: 0,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'right' as const,
      },
      title: {
        display: true,
        text: 'Payment Methods Distribution',
      },
      tooltip: {
        callbacks: {
          label: function(context: any) {
            const value = context.parsed;
            const total = values.reduce((sum, val) => sum + val, 0);
            const percentage = ((value / total) * 100).toFixed(1);
            return `${context.label}: ${paymentService.formatCurrency(value)} (${percentage}%)`;
          },
        },
      },
    },
  };

  return (
    <Card>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <FiBarChart3 style={{ marginRight: '8px' }} />
          Payment Methods
        </Heading>
      </CardHeader>
      <CardBody>
        {isLoading ? (
          <Flex justify="center" align="center" height="300px">
            <Spinner size="lg" />
          </Flex>
        ) : values.length > 0 ? (
          <Box height="300px">
            <Doughnut data={chartData} options={options} />
          </Box>
        ) : (
          <Flex justify="center" align="center" height="300px">
            <Text color="gray.500">No data available</Text>
          </Flex>
        )}
      </CardBody>
    </Card>
  );
};

interface RecentPaymentsTableProps {
  payments: Payment[];
  isLoading?: boolean;
}

const RecentPaymentsTable: React.FC<RecentPaymentsTableProps> = ({ payments, isLoading }) => {
  return (
    <Card>
      <CardHeader>
        <Heading size="md">Recent Payments</Heading>
      </CardHeader>
      <CardBody>
        {isLoading ? (
          <Flex justify="center" py={8}>
            <Spinner size="lg" />
          </Flex>
        ) : (
          <TableContainer>
            <Table size="sm">
              <Thead>
                <Tr>
                  <Th>Payment #</Th>
                  <Th>Contact</Th>
                  <Th>Date</Th>
                  <Th>Amount</Th>
                  <Th>Status</Th>
                </Tr>
              </Thead>
              <Tbody>
                {payments.slice(0, 5).map((payment) => (
                  <Tr key={payment.id}>
                    <Td>{payment.code}</Td>
                    <Td>{payment.contact?.name || '-'}</Td>
                    <Td>{paymentService.formatDate(payment.date)}</Td>
                    <Td>{paymentService.formatCurrency(payment.amount)}</Td>
                    <Td>
                      <Badge 
                        colorScheme={paymentService.getStatusColorScheme(payment.status)}
                        variant="subtle"
                      >
                        {payment.status}
                      </Badge>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
            {payments.length === 0 && (
              <Text textAlign="center" py={4} color="gray.500">
                No recent payments found
              </Text>
            )}
          </TableContainer>
        )}
      </CardBody>
    </Card>
  );
};

type DateRangeType = 'today' | 'week' | 'month' | 'quarter' | 'year';

const PaymentDashboard: React.FC = () => {
  const [analytics, setAnalytics] = useState<PaymentAnalytics | null>(null);
  const [integrationMetrics, setIntegrationMetrics] = useState<PaymentIntegrationMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dateRange, setDateRange] = useState<DateRangeType>('month');
  const [viewMode, setViewMode] = useState<'dashboard' | 'create' | 'details'>('dashboard');
  const [selectedPaymentId, setSelectedPaymentId] = useState<string | null>(null);
  
  const {
    isOpen: isCreateModalOpen,
    onOpen: onCreateModalOpen,
    onClose: onCreateModalClose
  } = useDisclosure();

  const getDateRange = (range: DateRangeType) => {
    const now = new Date();
    let startDate: Date;
    let endDate = now;

    switch (range) {
      case 'today':
        startDate = new Date(now.setHours(0, 0, 0, 0));
        break;
      case 'week':
        startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case 'month':
        startDate = new Date(now.getFullYear(), now.getMonth(), 1);
        break;
      case 'quarter':
        const quarter = Math.floor(now.getMonth() / 3);
        startDate = new Date(now.getFullYear(), quarter * 3, 1);
        break;
      case 'year':
        startDate = new Date(now.getFullYear(), 0, 1);
        break;
      default:
        startDate = new Date(now.getFullYear(), now.getMonth(), 1);
    }

    return {
      start: startDate.toISOString().split('T')[0],
      end: endDate.toISOString().split('T')[0]
    };
  };

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      setError(null);

      const { start, end } = getDateRange(dateRange);
      const analyticsData = await paymentService.getPaymentAnalytics(start, end);
      setAnalytics(analyticsData);
      
      // Also fetch integration metrics
      try {
        const metricsData = await paymentService.getIntegrationMetrics();
        setIntegrationMetrics(metricsData);
      } catch (metricsError) {
        console.warn('Failed to load integration metrics:', metricsError);
      }
    } catch (err: any) {
      console.error('Error fetching payment analytics:', err);
      setError(err.message || 'Failed to load payment analytics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAnalytics();
  }, [dateRange]);

  const handleRefresh = () => {
    fetchAnalytics();
  };

  const handleDateRangeChange = (newRange: string) => {
    setDateRange(newRange as DateRangeType);
  };

  const handleCreatePayment = (paymentId: string) => {
    onCreateModalClose();
    fetchAnalytics(); // Refresh data
  };

  const handleViewPayment = (paymentId: string) => {
    setSelectedPaymentId(paymentId);
    setViewMode('details');
  };

  const handleBackToDashboard = () => {
    setViewMode('dashboard');
    setSelectedPaymentId(null);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  // Handle different view modes
  if (viewMode === 'create') {
    return (
      <Box>
        <Button
          mb={4}
          variant="ghost"
          onClick={handleBackToDashboard}
          leftIcon={<FiArrowRight style={{ transform: 'rotate(180deg)' }} />}
        >
          Back to Dashboard
        </Button>
        <PaymentWithJournalForm
          onSubmit={handleCreatePayment}
          onCancel={handleBackToDashboard}
        />
      </Box>
    );
  }

  if (viewMode === 'details' && selectedPaymentId) {
    return (
      <Box>
        <Button
          mb={4}
          variant="ghost"
          onClick={handleBackToDashboard}
          leftIcon={<FiArrowRight style={{ transform: 'rotate(180deg)' }} />}
        >
          Back to Dashboard
        </Button>
        <PaymentDetailsView
          paymentId={selectedPaymentId}
          onEdit={(id) => console.log('Edit payment:', id)}
          onReverse={() => fetchAnalytics()}
        />
      </Box>
    );
  }

  return (
    <VStack spacing={6} align="stretch">
      {/* Header */}
      <Flex justify="space-between" align="center" wrap="wrap" gap={4}>
        <Heading size="lg" display="flex" alignItems="center">
          <FiCalendar style={{ marginRight: '12px' }} />
          Payment Management Dashboard
        </Heading>
        
        <HStack spacing={3}>
          <Button
            leftIcon={<FiPlus />}
            colorScheme="blue"
            onClick={onCreateModalOpen}
          >
            Create Payment
          </Button>
          
          <Select
            size="sm"
            value={dateRange}
            onChange={(e) => handleDateRangeChange(e.target.value)}
            minW="120px"
          >
            <option value="today">Today</option>
            <option value="week">This Week</option>
            <option value="month">This Month</option>
            <option value="quarter">This Quarter</option>
            <option value="year">This Year</option>
          </Select>
          
          <Button
            size="sm"
            leftIcon={<FiRefreshCw />}
            onClick={handleRefresh}
            isLoading={loading}
            variant="outline"
          >
            Refresh
          </Button>
        </HStack>
      </Flex>

      {/* Tabs for Dashboard and Payments */}
      <Tabs colorScheme="blue">
        <TabList>
          <Tab>Analytics Dashboard</Tab>
          <Tab>Payment Management</Tab>
        </TabList>

        <TabPanels>
          {/* Analytics Dashboard Tab */}
          <TabPanel px={0}>
            <VStack spacing={6} align="stretch">

      {/* Error Alert */}
      {error && (
        <Alert status="error">
          <AlertIcon />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6}>
        <StatCard
          title="Total Received"
          value={analytics ? paymentService.formatCurrency(analytics.total_received) : 'Rp 0'}
          change={analytics?.received_growth}
          icon={FiTrendingUp}
          colorScheme="green"
          isLoading={loading}
        />
        <StatCard
          title="Total Paid"
          value={analytics ? paymentService.formatCurrency(analytics.total_paid) : 'Rp 0'}
          change={analytics?.paid_growth}
          icon={FiTrendingDown}
          colorScheme="red"
          isLoading={loading}
        />
        <StatCard
          title="Net Cash Flow"
          value={analytics ? paymentService.formatCurrency(analytics.net_flow) : 'Rp 0'}
          change={analytics?.flow_growth}
          icon={FiDollarSign}
          colorScheme="blue"
          isLoading={loading}
        />
        <StatCard
          title="Outstanding"
          value={analytics ? paymentService.formatCurrency(analytics.total_outstanding) : 'Rp 0'}
          icon={FiClock}
          colorScheme="orange"
          isLoading={loading}
        />
      </SimpleGrid>

      {/* Charts */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        <PaymentTrendChart 
          data={analytics?.daily_trend || []} 
          isLoading={loading}
        />
        <PaymentMethodChart 
          data={analytics?.by_method || {}} 
          isLoading={loading}
        />
      </SimpleGrid>

      {/* Recent Payments Table */}
      <RecentPaymentsTable 
        payments={analytics?.recent_payments || []}
        isLoading={loading}
      />

      {/* Performance Metrics */}
      {analytics && (
        <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Average Processing Time</StatLabel>
                <StatNumber>{analytics.avg_payment_time}s</StatNumber>
                <StatHelpText>Per transaction</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Success Rate</StatLabel>
                <StatNumber>{analytics.success_rate.toFixed(1)}%</StatNumber>
                <StatHelpText>Of all transactions</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </SimpleGrid>
      )}
              </VStack>
            </TabPanel>

            {/* Payment Management Tab */}
            <TabPanel px={0}>
              <VStack spacing={6} align="stretch">
                {/* Journal Integration Metrics */}
                {integrationMetrics && (
                  <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6}>
                    <StatCard
                      title="Total Payments"
                      value={integrationMetrics.total_payments.toLocaleString()}
                      icon={FiDollarSign}
                      colorScheme="blue"
                      isLoading={loading}
                    />
                    <StatCard
                      title="With Journal Entries"
                      value={integrationMetrics.payments_with_journal.toLocaleString()}
                      icon={FiFileText}
                      colorScheme="green"
                      isLoading={loading}
                    />
                    <StatCard
                      title="Total Amount"
                      value={formatCurrency(integrationMetrics.total_amount)}
                      icon={FiTrendingUp}
                      colorScheme="purple"
                      isLoading={loading}
                    />
                    <StatCard
                      title="Success Rate"
                      value={`${(integrationMetrics.success_rate * 100).toFixed(1)}%`}
                      icon={FiActivity}
                      colorScheme="orange"
                      isLoading={loading}
                    />
                  </SimpleGrid>
                )}

                {/* Quick Actions */}
                <Card>
                  <CardHeader>
                    <Heading size="md">Quick Actions</Heading>
                  </CardHeader>
                  <CardBody>
                    <HStack spacing={4} wrap="wrap">
                      <Button
                        leftIcon={<FiPlus />}
                        colorScheme="blue"
                        onClick={() => setViewMode('create')}
                      >
                        Create Payment with Journal
                      </Button>
                      <Button
                        leftIcon={<FiRefreshCw />}
                        variant="outline"
                        onClick={handleRefresh}
                        isLoading={loading}
                      >
                        Refresh Data
                      </Button>
                    </HStack>
                  </CardBody>
                </Card>

                {/* Coming Soon - Payment List */}
                <Card>
                  <CardHeader>
                    <Heading size="md">Payment Management</Heading>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={4} py={8}>
                      <FiDollarSign size={48} color="gray.400" />
                      <Text color="gray.500" fontSize="lg">
                        Payment list and management features
                      </Text>
                      <Text color="gray.400" fontSize="sm" textAlign="center">
                        Use the "Create Payment with Journal" button to create new payments with automatic journal entries.
                        Individual payment details can be viewed from the analytics dashboard.
                      </Text>
                      <Button
                        colorScheme="blue"
                        variant="outline"
                        onClick={() => setViewMode('create')}
                        leftIcon={<FiPlus />}
                      >
                        Create Your First Payment
                      </Button>
                    </VStack>
                  </CardBody>
                </Card>
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>

        {/* Create Payment Modal */}
        <Modal isOpen={isCreateModalOpen} onClose={onCreateModalClose} size="6xl">
          <ModalOverlay />
          <ModalContent maxW="90vw">
            <ModalHeader>Create Payment with Journal Entry</ModalHeader>
            <ModalCloseButton />
            <ModalBody p={0}>
              <PaymentWithJournalForm
                onSubmit={handleCreatePayment}
                onCancel={onCreateModalClose}
              />
            </ModalBody>
          </ModalContent>
        </Modal>
      </VStack>
    );
  };

export default PaymentDashboard;
