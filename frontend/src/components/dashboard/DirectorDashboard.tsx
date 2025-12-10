'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
  Flex, 
  Heading, 
  Text, 
  Button, 
  Card,
  CardHeader,
  CardBody,
  HStack,
  Icon,
  Badge,
  SimpleGrid,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Progress,
  Stat,
  StatLabel,
  StatHelpText,
  useColorModeValue,
  useColorMode,
} from '@chakra-ui/react';
import {
  FiFolder,
  FiShoppingCart,
  FiDollarSign,
  FiPlus,
  FiBarChart2,
  FiActivity,
  FiCheckCircle,
  FiClock,
  FiTrendingUp,
  FiAlertCircle,
  FiPackage,
} from 'react-icons/fi';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import api from '@/services/api';
import { API_ENDPOINTS } from '@/config/api';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';
import AutoFitText from '@/components/common/AutoFitText';

interface DirectorDashboardProps {
  analytics: DashboardAnalytics | null;
}

const StatCard = ({ icon, title, stat, subtitle, colorScheme = 'blue' }: {
  icon: any;
  title: string;
  stat: string | number;
  subtitle?: string;
  colorScheme?: string;
}) => {
  const labelColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const numberColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const iconBgColor = useColorModeValue(`${colorScheme}.100`, `${colorScheme}.900`);
  const iconColor = useColorModeValue(`${colorScheme}.500`, `${colorScheme}.300`);
  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');

  return (
    <Card bg={cardBg} p={5} minH="130px">
      <CardHeader display="flex" flexDirection="row" alignItems="center" justifyContent="space-between" pb={2} p={0}>
        <Stat flex="1" minW={0} overflow="hidden">
          <StatLabel color={labelColor} noOfLines={1} title={title}>{title}</StatLabel>
          <Box mt={1}>
            <AutoFitText 
              value={String(stat)}
              maxFontSize={24}
              minFontSize={14}
              fontWeight={700}
              color={numberColor as string}
              title={String(stat)}
              style={{ lineHeight: 1.1 }}
            />
          </Box>
          {subtitle && (
            <StatHelpText noOfLines={1} mb={0}>
              {subtitle}
            </StatHelpText>
          )}
        </Stat>
        <Flex
          w={10}
          h={10}
          align="center"
          justify="center"
          borderRadius="full"
          bg={iconBgColor}
          transition="all 0.3s ease"
          ml={4}
          flexShrink={0}
        >
          <Icon as={icon} color={iconColor} w={5} h={5} />
        </Flex>
      </CardHeader>
    </Card>
  );
};

export const DirectorDashboard: React.FC<DirectorDashboardProps> = ({ analytics }) => {
  const router = useRouter();
  const { colorMode } = useColorMode();
  const [approvalStats, setApprovalStats] = useState<{ pending_approvals: number; total_amount_pending: number } | null>(null);
  const [loadingStats, setLoadingStats] = useState<boolean>(false);

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');

  useEffect(() => {
    const fetchApprovalStats = async () => {
      try {
        setLoadingStats(true);
        const res = await api.get(API_ENDPOINTS.PURCHASES_APPROVAL_STATS);
        setApprovalStats({
          pending_approvals: res.data?.pending_approvals ?? 0,
          total_amount_pending: res.data?.total_amount_pending ?? 0,
        });
      } catch (_) {
        setApprovalStats({ pending_approvals: 0, total_amount_pending: 0 });
      } finally {
        setLoadingStats(false);
      }
    };
    fetchApprovalStats();
  }, []);

  const formatCurrency = (value: number) => 
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value || 0);

  // Chart colors
  const chartColors = {
    primary: colorMode === 'dark' ? '#4dabf7' : '#2196F3',
    secondary: colorMode === 'dark' ? '#51cf66' : '#28a745',
    background: colorMode === 'dark' ? 'var(--bg-secondary)' : 'white',
    gridColor: colorMode === 'dark' ? '#495057' : '#e0e0e0',
    textColor: colorMode === 'dark' ? 'var(--text-primary)' : '#333333',
  };

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  const getStatusColor = (status: string) => {
    const statusLower = status?.toLowerCase();
    if (statusLower === 'active' || statusLower === 'in_progress' || statusLower === 'ongoing') return 'green';
    if (statusLower === 'completed') return 'blue';
    if (statusLower === 'pending') return 'yellow';
    if (statusLower === 'approved') return 'green';
    if (statusLower === 'rejected') return 'red';
    return 'gray';
  };

  if (!analytics) {
    return <Box color={textColor}>Loading analytics...</Box>;
  }

  const budgetUtilization = analytics.totalBudget > 0 
    ? ((analytics.totalSpent / analytics.totalBudget) * 100).toFixed(1)
    : '0';
  
  const remainingBudget = analytics.totalBudget - analytics.totalSpent;

  const projectStatusData = [
    { name: 'Aktif', value: analytics.activeProjects || 0 },
    { name: 'Selesai', value: analytics.completedProjects || 0 },
    { name: 'Lainnya', value: Math.max(0, (analytics.totalProjects || 0) - (analytics.activeProjects || 0) - (analytics.completedProjects || 0)) },
  ].filter(item => item.value > 0);

  const monthlyChartData = analytics.monthlyProjects?.map((item) => ({
    month: item.month,
    projects: item.value,
  })) || [];

  return (
    <Box>
      <Heading as="h2" size="lg" mb={6} color={textColor}>
        Dashboard Direktur
      </Heading>

      <SimpleGrid columns={{ base: 1, md: 2, lg: 5 }} spacing={4} mb={6}>
        <StatCard icon={FiFolder} title="Total Proyek" stat={analytics.totalProjects || 0} subtitle={`${analytics.activeProjects || 0} aktif`} colorScheme="blue" />
        <StatCard icon={FiCheckCircle} title="Proyek Selesai" stat={analytics.completedProjects || 0} colorScheme="green" />
        <StatCard icon={FiDollarSign} title="Total Budget" stat={formatCurrency(analytics.totalBudget || 0)} subtitle={`Terpakai: ${budgetUtilization}%`} colorScheme="purple" />
        <StatCard icon={FiTrendingUp} title="Total Pengeluaran" stat={formatCurrency(analytics.totalSpent || 0)} subtitle={`Sisa: ${formatCurrency(remainingBudget)}`} colorScheme="teal" />
        <StatCard icon={FiClock} title="Menunggu Approval" stat={loadingStats ? '...' : (approvalStats?.pending_approvals ?? analytics.pendingApprovals ?? 0)} subtitle={loadingStats ? '' : formatCurrency(approvalStats?.total_amount_pending || 0)} colorScheme="orange" />
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} mb={6}>
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiActivity} mr={2} color="blue.500" />
              Distribusi Status Proyek
            </Heading>
          </CardHeader>
          <CardBody>
            {projectStatusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie data={projectStatusData} cx="50%" cy="50%" labelLine={false} label={({ name, percent }: any) => `${name} ${((percent || 0) * 100).toFixed(0)}%`} outerRadius={80} fill="#8884d8" dataKey="value">
                    {projectStatusData.map((entry, index) => (<Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />))}
                  </Pie>
                  <Tooltip contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.gridColor}`, borderRadius: '8px' }} />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Flex justify="center" align="center" h="250px" color="gray.500"><Text>Belum ada data proyek</Text></Flex>
            )}
          </CardBody>
        </Card>

        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiBarChart2} mr={2} color="green.500" />
              Proyek Baru per Bulan
            </Heading>
          </CardHeader>
          <CardBody>
            {monthlyChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={monthlyChartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke={chartColors.gridColor} />
                  <XAxis dataKey="month" tick={{ fill: chartColors.textColor, fontSize: 12 }} />
                  <YAxis tick={{ fill: chartColors.textColor, fontSize: 12 }} />
                  <Tooltip contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.gridColor}`, borderRadius: '8px' }} />
                  <Legend />
                  <Bar dataKey="projects" fill={chartColors.primary} name="Proyek Baru" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <Flex justify="center" align="center" h="250px" color="gray.500"><Text>Belum ada data aktivitas</Text></Flex>
            )}
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} mb={6}>
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiFolder} mr={2} color="blue.500" />
              Proyek Terbaru
            </Heading>
          </CardHeader>
          <CardBody>
            {analytics.recentProjects && analytics.recentProjects.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead><Tr><Th>Nama Proyek</Th><Th>Status</Th><Th isNumeric>Progress</Th></Tr></Thead>
                <Tbody>
                  {analytics.recentProjects.slice(0, 5).map((project) => (
                    <Tr key={project.id} cursor="pointer" _hover={{ bg: hoverBg }} onClick={() => router.push(`/projects/${project.id}`)}>
                      <Td><Text fontWeight="medium" noOfLines={1}>{project.name}</Text><Text fontSize="xs" color="gray.500">{formatCurrency(project.budget)}</Text></Td>
                      <Td><Badge colorScheme={getStatusColor(project.status)} size="sm">{project.status}</Badge></Td>
                      <Td isNumeric><Box><Text fontSize="sm" mb={1}>{project.progress}%</Text><Progress value={project.progress} size="xs" colorScheme={project.progress >= 100 ? 'green' : 'blue'} borderRadius="full" /></Box></Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500"><Icon as={FiAlertCircle} mr={2} /><Text>Belum ada proyek</Text></Flex>
            )}
          </CardBody>
        </Card>

        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiShoppingCart} mr={2} color="orange.500" />
              Purchase Request Terbaru
            </Heading>
          </CardHeader>
          <CardBody>
            {analytics.recentPurchaseRequests && analytics.recentPurchaseRequests.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead><Tr><Th>No. PR</Th><Th>Proyek</Th><Th>Status</Th><Th isNumeric>Nilai</Th></Tr></Thead>
                <Tbody>
                  {analytics.recentPurchaseRequests.slice(0, 5).map((pr) => (
                    <Tr key={pr.id} cursor="pointer" _hover={{ bg: hoverBg }} onClick={() => router.push(`/cost-control/purchase-requests`)}>
                      <Td><Text fontWeight="medium" fontSize="sm">{pr.pr_number}</Text></Td>
                      <Td><Text fontSize="sm" noOfLines={1}>{pr.project_name || '-'}</Text></Td>
                      <Td><Badge colorScheme={getStatusColor(pr.status)} size="sm">{pr.status}</Badge></Td>
                      <Td isNumeric><Text fontSize="sm">{formatCurrency(pr.total_amount)}</Text></Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500"><Icon as={FiAlertCircle} mr={2} /><Text>Belum ada purchase request</Text></Flex>
            )}
          </CardBody>
        </Card>
      </SimpleGrid>

      <Card bg={cardBg}>
        <CardHeader>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiPlus} mr={2} color="blue.500" />
            Akses Cepat
          </Heading>
        </CardHeader>
        <CardBody>
          <HStack spacing={4} flexWrap="wrap">
            <Button leftIcon={<FiFolder />} colorScheme="blue" variant="outline" onClick={() => router.push('/projects')} size="md">Lihat Semua Proyek</Button>
            <Button leftIcon={<FiShoppingCart />} colorScheme="orange" variant="outline" onClick={() => router.push('/cost-control/purchase-requests')} size="md">Purchase Request</Button>
            <Button leftIcon={<FiBarChart2 />} colorScheme="green" variant="outline" onClick={() => router.push('/cost-control/budget-vs-actual')} size="md">Budget vs Actual</Button>
            <Button leftIcon={<FiPackage />} colorScheme="purple" variant="outline" onClick={() => router.push('/cost-control/material-tracking')} size="md">Material Tracking</Button>
          </HStack>
        </CardBody>
      </Card>
    </Box>
  );
};
