'use client';

import React, { useEffect, useState } from 'react';
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
  Progress,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiFolder,
  FiCheckSquare,
  FiFileText,
  FiArrowRight,
  FiClock,
  FiTrendingUp,
  FiAlertCircle,
  FiUsers,
  FiDollarSign,
} from 'react-icons/fi';
import { useRouter } from 'next/navigation';
import projectService from '@/services/projectService';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';

interface GMDashboardProps {
  analytics: DashboardAnalytics | null;
}

export const GMDashboard: React.FC<GMDashboardProps> = ({ analytics }) => {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [pendingDailyReports, setPendingDailyReports] = useState(0);

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      
      // Get pending daily reports for approval
      const pendingReports = await projectService.getPendingDailyUpdates();
      setPendingDailyReports(pendingReports?.length || 0);

    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

  if (loading && !analytics) {
    return (
      <Flex justify="center" align="center" minH="60vh">
        <Spinner size="xl" color="blue.500" thickness="4px" />
      </Flex>
    );
  }

  const budgetUtilization = analytics?.totalBudget && analytics.totalBudget > 0
    ? ((analytics.totalSpent / analytics.totalBudget) * 100).toFixed(1)
    : '0';

  const getStatusColor = (status: string) => {
    const s = status?.toLowerCase();
    if (s === 'active' || s === 'ongoing' || s === 'in_progress') return 'green';
    if (s === 'completed') return 'blue';
    if (s === 'on-hold') return 'yellow';
    return 'gray';
  };

  return (
    <Box>
      <Heading as="h2" size="lg" color={textColor} mb={6}>
        Dashboard GM
      </Heading>

      {/* Stats Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Proyek</StatLabel>
              <StatNumber color="blue.500">{analytics?.totalProjects || 0}</StatNumber>
              <StatHelpText>{analytics?.activeProjects || 0} aktif</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Daily Report Pending</StatLabel>
              <StatNumber color="orange.500">{pendingDailyReports}</StatNumber>
              <StatHelpText>
                <Icon as={FiClock} mr={1} />
                Menunggu approval
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>PR Pending</StatLabel>
              <StatNumber color="yellow.500">{analytics?.pendingApprovals || 0}</StatNumber>
              <StatHelpText>
                <Icon as={FiCheckSquare} mr={1} />
                Perlu approval
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Budget Utilization</StatLabel>
              <StatNumber color="teal.500">{budgetUtilization}%</StatNumber>
              <StatHelpText>
                <Icon as={FiDollarSign} mr={1} />
                {formatCurrency(analytics?.totalSpent || 0)}
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} mb={6}>
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
                Lihat Semua
              </Button>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {analytics?.recentProjects && analytics.recentProjects.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>Nama Proyek</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Progress</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {analytics.recentProjects.slice(0, 5).map((project) => (
                    <Tr
                      key={project.id}
                      cursor="pointer"
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push(`/projects/${project.id}`)}
                    >
                      <Td>
                        <Text fontWeight="medium" noOfLines={1}>{project.name}</Text>
                      </Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(project.status)} size="sm">
                          {project.status}
                        </Badge>
                      </Td>
                      <Td isNumeric>
                        <HStack justify="flex-end">
                          <Progress 
                            value={project.progress} 
                            size="sm" 
                            w="60px"
                            colorScheme={project.progress >= 100 ? 'green' : 'blue'}
                            borderRadius="full"
                          />
                          <Text fontSize="sm">{project.progress}%</Text>
                        </HStack>
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

        {/* Quick Actions */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiUsers} mr={2} color="green.500" />
              Akses Cepat
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Button
                leftIcon={<FiFileText />}
                colorScheme="orange"
                size="lg"
                onClick={() => router.push('/daily-report-approval')}
              >
                Approval Daily Report ({pendingDailyReports})
              </Button>
              <Button
                leftIcon={<FiCheckSquare />}
                colorScheme="blue"
                variant="outline"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Approval Purchase Request
              </Button>
              <Button
                leftIcon={<FiFolder />}
                colorScheme="teal"
                variant="outline"
                onClick={() => router.push('/projects')}
              >
                Kelola Proyek
              </Button>
              <Button
                leftIcon={<FiTrendingUp />}
                colorScheme="purple"
                variant="outline"
                onClick={() => router.push('/cost-control/budget-vs-actual')}
              >
                Budget vs Actual
              </Button>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Pending Approvals Alert */}
      {(pendingDailyReports > 0 || (analytics?.pendingApprovals || 0) > 0) && (
        <Card bg="orange.50" borderColor="orange.200" borderWidth={1}>
          <CardBody>
            <HStack>
              <Icon as={FiClock} color="orange.500" boxSize={6} />
              <Box>
                <Text fontWeight="bold" color="orange.700">Item Menunggu Approval</Text>
                <Text color="orange.600" fontSize="sm">
                  {pendingDailyReports > 0 && `${pendingDailyReports} daily report`}
                  {pendingDailyReports > 0 && (analytics?.pendingApprovals || 0) > 0 && ' dan '}
                  {(analytics?.pendingApprovals || 0) > 0 && `${analytics?.pendingApprovals} purchase request`}
                  {' '}menunggu persetujuan Anda
                </Text>
              </Box>
            </HStack>
          </CardBody>
        </Card>
      )}
    </Box>
  );
};
