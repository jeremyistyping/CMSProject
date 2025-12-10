'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  SimpleGrid,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Text,
  Stat,
  StatLabel,
  StatHelpText,
  Flex,
  Icon,
  Button,
  HStack,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Progress,
  useColorMode,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiFolder,
  FiDollarSign,
  FiShoppingCart,
  FiActivity,
  FiBarChart2,
  FiPlus,
  FiTrendingUp,
  FiAlertCircle,
} from 'react-icons/fi';
import AutoFitText from '@/components/common/AutoFitText';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';

interface AdminDashboardProps {
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

  return (
    <Card p={5} minH="130px">
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

export const AdminDashboard: React.FC<AdminDashboardProps> = ({ analytics }) => {
  const router = useRouter();
  const { colorMode } = useColorMode();
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  
  // Dynamic colors for charts based on theme
  const chartColors = {
    primary: colorMode === 'dark' ? '#4dabf7' : '#2196F3',
    secondary: colorMode === 'dark' ? '#51cf66' : '#28a745',
    tertiary: colorMode === 'dark' ? '#ffd43b' : '#ffc107',
    quaternary: colorMode === 'dark' ? '#ff6b6b' : '#dc3545',
    background: colorMode === 'dark' ? 'var(--bg-secondary)' : 'white',
    gridColor: colorMode === 'dark' ? '#495057' : '#e0e0e0',
    textColor: colorMode === 'dark' ? 'var(--text-primary)' : '#333333',
  };

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);
  
  const formatNumber = (value: number) =>
    new Intl.NumberFormat('id-ID').format(value);
  
  if (!analytics) {
    return <Box color={colorMode === 'dark' ? 'var(--text-primary)' : 'gray.800'}>Loading analytics...</Box>;
  }

  // Calculate budget utilization
  const budgetUtilization = analytics.totalBudget > 0 
    ? ((analytics.totalSpent / analytics.totalBudget) * 100).toFixed(1)
    : '0';
  
  const remainingBudget = analytics.totalBudget - analytics.totalSpent;

  // Prepare chart data
  const projectChartData = analytics.monthlyProjects?.map((item, index) => ({
    month: item.month,
    projects: item.value,
    prs: analytics.monthlyPRs?.[index]?.value || 0,
  })) || [];

  // Status badge color helper
  const getStatusColor = (status: string) => {
    const statusLower = status?.toLowerCase();
    if (statusLower === 'active' || statusLower === 'in_progress' || statusLower === 'ongoing') return 'green';
    if (statusLower === 'completed') return 'blue';
    if (statusLower === 'pending') return 'yellow';
    if (statusLower === 'approved') return 'green';
    if (statusLower === 'rejected') return 'red';
    return 'gray';
  };

  return (
    <Box>
      {/* Statistics Cards */}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6} mb={6}>
        <StatCard
          icon={FiFolder}
          title="Total Proyek"
          stat={analytics.totalProjects || 0}
          subtitle={`${analytics.activeProjects || 0} aktif, ${analytics.completedProjects || 0} selesai`}
          colorScheme="blue"
        />
        <StatCard
          icon={FiShoppingCart}
          title="Purchase Request"
          stat={analytics.totalPurchaseRequests || 0}
          subtitle={`${analytics.pendingApprovals || 0} menunggu approval`}
          colorScheme="orange"
        />
        <StatCard
          icon={FiDollarSign}
          title="Total Budget"
          stat={formatCurrency(analytics.totalBudget || 0)}
          subtitle={`Terpakai: ${budgetUtilization}%`}
          colorScheme="green"
        />
        <StatCard
          icon={FiTrendingUp}
          title="Total Pengeluaran"
          stat={formatCurrency(analytics.totalSpent || 0)}
          subtitle={`Sisa: ${formatCurrency(remainingBudget)}`}
          colorScheme="purple"
        />
      </SimpleGrid>

      {/* Quick Access Section */}
      <Card mb={6}>
        <CardHeader>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiPlus} mr={2} color="blue.500" />
            Akses Cepat
          </Heading>
        </CardHeader>
        <CardBody>
          <HStack spacing={4} flexWrap="wrap">
            <Button
              leftIcon={<FiFolder />}
              colorScheme="blue"
              variant="outline"
              onClick={() => router.push('/projects')}
              size="md"
            >
              Kelola Proyek
            </Button>
            <Button
              leftIcon={<FiShoppingCart />}
              colorScheme="orange"
              variant="outline"
              onClick={() => router.push('/cost-control/purchase-requests')}
              size="md"
            >
              Purchase Request
            </Button>
            <Button
              leftIcon={<FiBarChart2 />}
              colorScheme="green"
              variant="outline"
              onClick={() => router.push('/cost-control/budget-vs-actual')}
              size="md"
            >
              Budget vs Actual
            </Button>
            <Button
              leftIcon={<FiActivity />}
              colorScheme="purple"
              variant="outline"
              onClick={() => router.push('/cost-control/material-tracking')}
              size="md"
            >
              Material Tracking
            </Button>
          </HStack>
        </CardBody>
      </Card>

      {/* Charts Section */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} mb={6}>
        {/* Monthly Projects & PRs Chart */}
        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiActivity} mr={2} color="blue.500" />
              Aktivitas Bulanan
            </Heading>
          </CardHeader>
          <CardBody>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={projectChartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartColors.gridColor} />
                <XAxis 
                  dataKey="month" 
                  tick={{ fill: chartColors.textColor, fontSize: 12 }}
                  axisLine={{ stroke: chartColors.gridColor }}
                />
                <YAxis 
                  tick={{ fill: chartColors.textColor, fontSize: 12 }}
                  axisLine={{ stroke: chartColors.gridColor }}
                  allowDecimals={false}
                  domain={[0, (dataMax: number) => Math.max(5, Math.ceil(dataMax))]}
                  tickFormatter={(value) => Math.floor(value).toString()}
                />
                <Tooltip 
                  contentStyle={{
                    backgroundColor: chartColors.background,
                    border: `1px solid ${chartColors.gridColor}`,
                    borderRadius: '8px',
                    color: chartColors.textColor,
                  }}
                  formatter={(value: number) => [formatNumber(value), 'Proyek Baru']}
                />
                <Legend wrapperStyle={{ color: chartColors.textColor }} />
                <Bar dataKey="projects" fill={chartColors.primary} name="Proyek Baru" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardBody>
        </Card>

        {/* Monthly PR Value Chart */}
        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiDollarSign} mr={2} color="green.500" />
              Nilai Purchase Request Bulanan
            </Heading>
          </CardHeader>
          <CardBody>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={projectChartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartColors.gridColor} />
                <XAxis 
                  dataKey="month" 
                  tick={{ fill: chartColors.textColor, fontSize: 12 }}
                  axisLine={{ stroke: chartColors.gridColor }}
                />
                <YAxis 
                  tick={{ fill: chartColors.textColor, fontSize: 12 }}
                  axisLine={{ stroke: chartColors.gridColor }}
                  tickFormatter={(value) => formatNumber(value)}
                />
                <Tooltip 
                  contentStyle={{
                    backgroundColor: chartColors.background,
                    border: `1px solid ${chartColors.gridColor}`,
                    borderRadius: '8px',
                    color: chartColors.textColor,
                  }}
                  formatter={(value: number) => [formatCurrency(value), 'Nilai PR']}
                />
                <Legend wrapperStyle={{ color: chartColors.textColor }} />
                <Line 
                  type="monotone" 
                  dataKey="prs" 
                  stroke={chartColors.secondary} 
                  strokeWidth={2}
                  dot={{ fill: chartColors.secondary }}
                  name="Nilai PR"
                />
              </LineChart>
            </ResponsiveContainer>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Recent Data Tables */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        {/* Recent Projects */}
        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiFolder} mr={2} color="blue.500" />
              Proyek Terbaru
            </Heading>
          </CardHeader>
          <CardBody>
            {analytics.recentProjects && analytics.recentProjects.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>Nama Proyek</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Progress</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {analytics.recentProjects.map((project) => (
                    <Tr 
                      key={project.id} 
                      cursor="pointer" 
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push(`/projects/${project.id}`)}
                    >
                      <Td>
                        <Text fontWeight="medium" noOfLines={1}>{project.name}</Text>
                        <Text fontSize="xs" color="gray.500">
                          {formatCurrency(project.budget)}
                        </Text>
                      </Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(project.status)} size="sm">
                          {project.status}
                        </Badge>
                      </Td>
                      <Td isNumeric>
                        <Box>
                          <Text fontSize="sm" mb={1}>{project.progress}%</Text>
                          <Progress 
                            value={project.progress} 
                            size="xs" 
                            colorScheme={project.progress >= 100 ? 'green' : 'blue'}
                            borderRadius="full"
                          />
                        </Box>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada proyek</Text>
              </Flex>
            )}
          </CardBody>
        </Card>

        {/* Recent Purchase Requests */}
        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiShoppingCart} mr={2} color="orange.500" />
              Purchase Request Terbaru
            </Heading>
          </CardHeader>
          <CardBody>
            {analytics.recentPurchaseRequests && analytics.recentPurchaseRequests.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>No. PR</Th>
                    <Th>Proyek</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Nilai</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {analytics.recentPurchaseRequests.map((pr) => (
                    <Tr 
                      key={pr.id}
                      cursor="pointer"
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push(`/cost-control/purchase-requests`)}
                    >
                      <Td>
                        <Text fontWeight="medium" fontSize="sm">{pr.pr_number}</Text>
                      </Td>
                      <Td>
                        <Text fontSize="sm" noOfLines={1}>{pr.project_name || '-'}</Text>
                      </Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(pr.status)} size="sm">
                          {pr.status}
                        </Badge>
                      </Td>
                      <Td isNumeric>
                        <Text fontSize="sm">{formatCurrency(pr.total_amount)}</Text>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada purchase request</Text>
              </Flex>
            )}
          </CardBody>
        </Card>
      </SimpleGrid>
    </Box>
  );
};
