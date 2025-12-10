'use client';

import React from 'react';
import {
  Box,
  Heading,
  Text,
  Card,
  CardHeader,
  CardBody,
  Button,
  VStack,
  HStack,
  Icon,
  SimpleGrid,
  Flex,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Spinner,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  useColorModeValue,
  CircularProgress,
  CircularProgressLabel,
} from '@chakra-ui/react';
import {
  FiFolder,
  FiDollarSign,
  FiTrendingUp,
  FiArrowRight,
  FiAlertCircle,
  FiShoppingCart,
  FiPieChart,
  FiActivity,
  FiTarget,
} from 'react-icons/fi';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import { useRouter } from 'next/navigation';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';

interface ManagingDirectorDashboardProps {
  analytics: DashboardAnalytics | null;
}

export const ManagingDirectorDashboard: React.FC<ManagingDirectorDashboardProps> = ({ analytics }) => {
  const router = useRouter();

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

  if (!analytics) {
    return (
      <Flex justify="center" align="center" minH="60vh">
        <Spinner size="xl" color="blue.500" thickness="4px" />
      </Flex>
    );
  }

  const budgetUtilization = analytics.totalBudget > 0
    ? ((analytics.totalSpent / analytics.totalBudget) * 100)
    : 0;

  const remainingBudget = analytics.totalBudget - analytics.totalSpent;
  const isOverBudget = remainingBudget < 0;

  // Project status distribution for pie chart
  const projectStatusData = [
    { name: 'Aktif', value: analytics.activeProjects || 0 },
    { name: 'Selesai', value: analytics.completedProjects || 0 },
    { name: 'Lainnya', value: Math.max(0, (analytics.totalProjects || 0) - (analytics.activeProjects || 0) - (analytics.completedProjects || 0)) },
  ].filter(item => item.value > 0);

  const getStatusColor = (status: string) => {
    const s = status?.toLowerCase();
    if (s === 'active' || s === 'ongoing' || s === 'in_progress') return 'green';
    if (s === 'completed') return 'blue';
    if (s === 'pending') return 'yellow';
    if (s === 'approved') return 'green';
    if (s === 'rejected') return 'red';
    return 'gray';
  };

  return (
    <Box>
      <Heading as="h2" size="lg" color={textColor} mb={6}>
        Dashboard Direktur Utama
      </Heading>

      {/* Executive Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
        <Card bg={cardBg} borderLeft="4px solid" borderLeftColor="blue.500">
          <CardBody>
            <Stat>
              <StatLabel>Portfolio Proyek</StatLabel>
              <StatNumber color="blue.500">{analytics.totalProjects}</StatNumber>
              <StatHelpText>
                <Badge colorScheme="green" mr={1}>{analytics.activeProjects}</Badge>
                aktif
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg} borderLeft="4px solid" borderLeftColor="green.500">
          <CardBody>
            <Stat>
              <StatLabel>Total Budget</StatLabel>
              <StatNumber fontSize="lg" color="green.500">
                {formatCurrency(analytics.totalBudget)}
              </StatNumber>
              <StatHelpText>Seluruh proyek</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg} borderLeft="4px solid" borderLeftColor={isOverBudget ? 'red.500' : 'teal.500'}>
          <CardBody>
            <Stat>
              <StatLabel>Total Pengeluaran</StatLabel>
              <StatNumber fontSize="lg" color={isOverBudget ? 'red.500' : 'teal.500'}>
                {formatCurrency(analytics.totalSpent)}
              </StatNumber>
              <StatHelpText>
                {budgetUtilization.toFixed(1)}% dari budget
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg} borderLeft="4px solid" borderLeftColor="orange.500">
          <CardBody>
            <Stat>
              <StatLabel>Pending Approval</StatLabel>
              <StatNumber color="orange.500">{analytics.pendingApprovals}</StatNumber>
              <StatHelpText>
                <Icon as={FiShoppingCart} mr={1} />
                Purchase requests
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 3 }} spacing={6} mb={6}>
        {/* Project Status Distribution */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiPieChart} mr={2} color="blue.500" />
              Distribusi Status Proyek
            </Heading>
          </CardHeader>
          <CardBody>
            {projectStatusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={200}>
                <PieChart>
                  <Pie
                    data={projectStatusData}
                    cx="50%"
                    cy="50%"
                    innerRadius={40}
                    outerRadius={70}
                    fill="#8884d8"
                    dataKey="value"
                    label={({ name, percent }: any) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {projectStatusData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Flex justify="center" align="center" h="200px" color="gray.500">
                <Text>Belum ada data</Text>
              </Flex>
            )}
          </CardBody>
        </Card>

        {/* Budget Health */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiTarget} mr={2} color="green.500" />
              Kesehatan Budget
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4}>
              <Box textAlign="center">
                <CircularProgress 
                  value={Math.min(budgetUtilization, 100)} 
                  size="100px" 
                  color={isOverBudget ? 'red.400' : budgetUtilization > 80 ? 'yellow.400' : 'green.400'}
                  trackColor="gray.200"
                >
                  <CircularProgressLabel fontWeight="bold" fontSize="sm">
                    {budgetUtilization.toFixed(0)}%
                  </CircularProgressLabel>
                </CircularProgress>
                <Text mt={2} fontSize="sm" fontWeight="medium">
                  {isOverBudget ? 'Over Budget!' : budgetUtilization > 80 ? 'Hampir Penuh' : 'Sehat'}
                </Text>
              </Box>
              <Box w="full" fontSize="sm">
                <HStack justify="space-between" mb={1}>
                  <Text color="gray.500">Terpakai</Text>
                  <Text fontWeight="medium">{formatCurrency(analytics.totalSpent)}</Text>
                </HStack>
                <HStack justify="space-between">
                  <Text color="gray.500">Sisa</Text>
                  <Text fontWeight="medium" color={isOverBudget ? 'red.500' : 'green.500'}>
                    {formatCurrency(remainingBudget)}
                  </Text>
                </HStack>
              </Box>
            </VStack>
          </CardBody>
        </Card>

        {/* Quick Actions */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiActivity} mr={2} color="purple.500" />
              Akses Cepat
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              <Button
                leftIcon={<FiFolder />}
                colorScheme="blue"
                onClick={() => router.push('/projects')}
              >
                Lihat Semua Proyek
              </Button>
              <Button
                leftIcon={<FiShoppingCart />}
                colorScheme="orange"
                variant="outline"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Approval PR ({analytics.pendingApprovals})
              </Button>
              <Button
                leftIcon={<FiTrendingUp />}
                colorScheme="green"
                variant="outline"
                onClick={() => router.push('/cost-control/budget-vs-actual')}
              >
                Budget vs Actual
              </Button>
              <Button
                leftIcon={<FiDollarSign />}
                colorScheme="purple"
                variant="outline"
                onClick={() => router.push('/cost-control')}
              >
                Cost Control
              </Button>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Recent Data Tables */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        {/* Recent Projects */}
        <Card bg={cardBg}>
          <CardHeader>
            <HStack justify="space-between">
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiFolder} mr={2} color="blue.500" />
                Proyek Terbaru
              </Heading>
              <Button
                size="sm"
                rightIcon={<FiArrowRight />}
                variant="ghost"
                colorScheme="blue"
                onClick={() => router.push('/projects')}
              >
                Semua
              </Button>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {analytics.recentProjects && analytics.recentProjects.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>Proyek</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Progress</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {analytics.recentProjects.slice(0, 4).map((project) => (
                    <Tr
                      key={project.id}
                      cursor="pointer"
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push(`/projects/${project.id}`)}
                    >
                      <Td>
                        <Text fontWeight="medium" noOfLines={1} fontSize="sm">{project.name}</Text>
                      </Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(project.status)} size="sm">
                          {project.status}
                        </Badge>
                      </Td>
                      <Td isNumeric>
                        <Text fontSize="sm">{project.progress}%</Text>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={6} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada proyek</Text>
              </Flex>
            )}
          </CardBody>
        </Card>

        {/* Recent Purchase Requests */}
        <Card bg={cardBg}>
          <CardHeader>
            <HStack justify="space-between">
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiShoppingCart} mr={2} color="orange.500" />
                Purchase Request Terbaru
              </Heading>
              <Button
                size="sm"
                rightIcon={<FiArrowRight />}
                variant="ghost"
                colorScheme="orange"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Semua
              </Button>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {analytics.recentPurchaseRequests && analytics.recentPurchaseRequests.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>No. PR</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Nilai</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {analytics.recentPurchaseRequests.slice(0, 4).map((pr) => (
                    <Tr
                      key={pr.id}
                      cursor="pointer"
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push('/cost-control/purchase-requests')}
                    >
                      <Td>
                        <Text fontWeight="medium" fontSize="sm">{pr.pr_number}</Text>
                      </Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(pr.status)} size="sm">
                          {pr.status}
                        </Badge>
                      </Td>
                      <Td isNumeric fontSize="sm">{formatCurrency(pr.total_amount)}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={6} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada PR</Text>
              </Flex>
            )}
          </CardBody>
        </Card>
      </SimpleGrid>
    </Box>
  );
};
